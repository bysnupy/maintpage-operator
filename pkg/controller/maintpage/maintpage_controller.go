package maintpage

import (
        "time"
	"context"

	maintpagev1alpha1 "github.com/bysnupy/maintpage-operator/pkg/apis/maintpage/v1alpha1"
    
        appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

	// Watch for changes to primary resource MaintPage
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForObject{})
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
	pod := newPodForCR(maintpage)

	// Set MaintPage instance as the owner and controller
	if err := controllerutil.SetControllerReference(maintpage, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	podfound := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, podfound)
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


        servicefound := &corev1.Service{}
        err = r.client.Get(context.TODO(), types.NamespacedName{Name: maintpage.Spec.TargetService, Namespace: maintpage.Namespace}, servicefound)
        if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Not Found Target Service", "Service.Namespace", maintpage.Namespace, "Service.Name", maintpage.Spec.TargetService)
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Deployment")
		return reconcile.Result{}, err
	}

	deployfound := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: maintpage.Spec.TargetDeployment, Namespace: maintpage.Namespace}, deployfound)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Not Found Target Deployment", "Deployment.Namespace", maintpage.Namespace, "Deployment.Name", maintpage.Spec.TargetDeployment)
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Deployment")
		return reconcile.Result{}, err
	} 

        currentimage := deployfound.Spec.Template.Spec.Containers[0].Image
        reqLogger.Info("Current Image: " + currentimage)
        reqLogger.Info("Target Image: " + maintpage.Spec.TargetImage)
        reqLogger.Info("Service Name: " + servicefound.Name)
        reqLogger.Info("Service Selector before changes: " + servicefound.Spec.Selector["app"])
        if currentimage != maintpage.Spec.TargetImage {
                servicefound.Spec.Selector["app"] = "maintpage"
                err := r.client.Update(context.TODO(), servicefound) 
                if err != nil {
                        reqLogger.Error(err, "Failed to Update Service", servicefound.Name)
                        return reconcile.Result{}, err
                }
                reqLogger.Info("Changed Service Selector")
                reqLogger.Info("Service Selector after changes: " + servicefound.Spec.Selector["app"])
        } else {
                reqLogger.Info("Not changed Image")
        }

	return reconcile.Result{RequeueAfter: time.Second*5}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *maintpagev1alpha1.MaintPage) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-maintpage-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "maintpage",
					Image:   "quay.io/daein/maintpage:latest",
				},
			},
		},
	}
}
