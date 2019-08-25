package maintpage

import (
	"context"

	maintpagev1alpha1 "github.com/bysnupy/maintpage-operator/pkg/apis/maintpage/v1alpha1"
    
        appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_maintpage")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new MaintPage Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMaintPage{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("maintpage-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource MaintPage
	err = c.Watch(&source.Kind{Type: &maintpagev1alpha1.MaintPage{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner MaintPage
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &maintpagev1alpha1.MaintPage{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
                IsController: true,
                OwnerType:    &maintpagev1alpha1.MaintPage{},
        })
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileMaintPage implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMaintPage{}

// ReconcileMaintPage reconciles a MaintPage object
type ReconcileMaintPage struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a MaintPage object and makes changes based on the state read
// and what is in the MaintPage.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMaintPage) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling MaintPage")

	// Fetch the MaintPage instance
	maintpage := &maintpagev1alpha1.MaintPage{}
	err := r.client.Get(context.TODO(), request.NamespacedName, maintpage)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new Pod object
	pod := newPodForMaintPage(maintpage)

	// Set MaintPage instance as the owner and controller
	if err := controllerutil.SetControllerReference(maintpage, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	podfound := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: maintpage.Name + "-maintpage-pod", Namespace: pod.Namespace}, podfound)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Pod already exists - don't requeue
	reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", podfound.Namespace, "Pod.Name", podfound.Name)

	// Check if the deployment already exists, if not create a new one
	depfound := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: maintpage.Spec.AppConfig.AppName, Namespace: maintpage.Namespace}, depfound)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
	        reqLogger.Info("App Name: " + maintpage.Spec.AppConfig.AppName + ", App Image: " + maintpage.Spec.AppConfig.AppImage)

		dep := r.deploymentForApp(maintpage)
		reqLogger.Info("Creating a App Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return reconcile.Result{}, err
		}

		// Deployment created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Deployment")
		return reconcile.Result{}, err
	}

	// Check if the service already exists, if not create a new one
	svcfound := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: maintpage.Spec.AppConfig.AppName, Namespace: maintpage.Namespace}, svcfound)
	if err != nil && errors.IsNotFound(err) {
		// Define a new service
		svc := r.serviceForApp(maintpage)

		reqLogger.Info("Creating a App Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
		err = r.client.Create(context.TODO(), svc)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Service")
		return reconcile.Result{}, err
	}

        // Check if MaintPage flag is enabled
        if maintpage.Spec.MaintPageConfig.MaintPageToggle {
                svcfound.Spec.Selector["app"] = maintpage.Name
                err := r.client.Update(context.TODO(), svcfound) 
                if err != nil {
                        reqLogger.Error(err, "Failed to Update Service", svcfound.Name)
                        return reconcile.Result{}, err
                }
                reqLogger.Info("Changed service selector: " + svcfound.Spec.Selector["app"])

                // Update Status
                statusErr := r.client.Status().Update(context.TODO(), updateMaintStatus(maintpage, "Published"))
                if statusErr != nil {
                        reqLogger.Error(statusErr, "Failed to Update MaintPage status")
                        return reconcile.Result{}, statusErr
                }
        } else {
                if svcfound.Spec.Selector["app"] != maintpage.Spec.AppConfig.AppName {
                        svcfound.Spec.Selector["app"] = maintpage.Spec.AppConfig.AppName
                        err := r.client.Update(context.TODO(), svcfound)
                        if err != nil {
                                reqLogger.Error(err, "Failed to Update Service to App Name", svcfound.Name)
                                return reconcile.Result{}, err
                        }
                        reqLogger.Info("Changed service selector: " + svcfound.Spec.Selector["app"])            
           
                }
                // Update Status
                statusErr := r.client.Status().Update(context.TODO(), updateMaintStatus(maintpage, "Not Published"))
                if statusErr != nil {
                       reqLogger.Error(statusErr, "Failed to Update MaintPage status")
                       return reconcile.Result{}, statusErr
                }
        }

        // Revert if current image does not match with defined one
        if depfound.Spec.Template.Spec.Containers[0].Image != maintpage.Spec.AppConfig.AppImage {
                depfound.Spec.Template.Spec.Containers[0].Image = maintpage.Spec.AppConfig.AppImage
                err := r.client.Update(context.TODO(), depfound)
                if err != nil {
                        reqLogger.Error(err, "Failed to Update Deployment App Image", depfound.Name)
                        return reconcile.Result{}, err
                }
                reqLogger.Info("Reverted Deployment Image as App Image")
        }     
 
	return reconcile.Result{}, nil
}

// newPodForMaintPage returns a maintpage pod with the same name/namespace as the cr
func newPodForMaintPage(m *maintpagev1alpha1.MaintPage) *corev1.Pod {
        maintpageimage := m.Spec.MaintPageConfig.MaintPageImage
	labels := map[string]string{
		"app": m.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-maintpage-pod",
			Namespace: m.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "maintpage",
					Image:   maintpageimage,
				},
			},
		},
	}
}

// deploymentForApp returns a App Deployment object
func (r *ReconcileMaintPage) deploymentForApp(m *maintpagev1alpha1.MaintPage) *appsv1.Deployment {
	appname  := m.Spec.AppConfig.AppName
	appimage := m.Spec.AppConfig.AppImage
        replicas := int32(1)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appname,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
                        Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": appname},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": appname},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:   appimage,
						Name:    appname,
					}},
				},
			},
		},
	}
	// Set MaintPage instance as the owner and controller
	controllerutil.SetControllerReference(m, dep, r.scheme)
	return dep
}

func (r *ReconcileMaintPage) serviceForApp(m *maintpagev1alpha1.MaintPage) *corev1.Service {
	appname  := m.Spec.AppConfig.AppName

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appname,
			Namespace: m.Namespace,
		},
		Spec: corev1.ServiceSpec{
		        Ports: []corev1.ServicePort{{
                                Name:       "8080-tcp",
                                Protocol:   "TCP",
                                Port:       8080,
                                TargetPort: intstr.FromInt(8080),
                        }},
			Selector: map[string]string{"app": appname},
	        },
	}
	// Set MaintPage instance as the owner and controller
	controllerutil.SetControllerReference(m, svc, r.scheme)
	return svc
}

func updateMaintStatus(m *maintpagev1alpha1.MaintPage, status string) *maintpagev1alpha1.MaintPage {
        m.Status.MaintPublishStatus = status
        return m
}
