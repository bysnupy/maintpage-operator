# Maintpage Operator

Maintpage Operator automates the task to show a maintenance page.

## Description

Maintpage Operator help you to switch mode from production to maintenance mode.

![maintpage-operator work process](https://github.com/bysnupy/maintpage-operator/blob/master/maintpage-operator-process.png)

### Example MaintPage CR
~~~
apiVersion: maintpage.example.com/v1alpha1
kind: MaintPage
metadata:
  name: maintpage
spec:
  maintpageconfig:
    maintpagetoggle: false
    maintpageimage: quay.io/daein/maintpage:latest
  appconfig:
  appname: httpd
  appimage: quay.io/daein/prodpage:latest
~~~

Spec name|Description
-|-
maintpagetoggle| Toggle to change the maintenance page
maintpageimage| Container image to run as maintenance page pod
appname| Application resource name
appimage| Container image to run as application pod

### Installation

* Install Maintpage Operator as follows.
~~~
# oc new-project maintpage-operator
# oc create -f deploy/service_account.yaml
# oc create -f deploy/role.yaml
# oc create -f deploy/role_binding.yaml
# oc create -f deploy/operator.yaml
~~~

* Check installed operator pod
~~~
# oc get pod 
NAME                                  READY     STATUS    RESTARTS   AGE
maintpage-operator-6df9d9c85c-s7xmw   1/1       Running   0          20s
~~~

* Define MaintPage CR for deploying application and maintenance page pod
~~~
# oc create -f - <<EOF
apiVersion: maintpage.example.com/v1alpha1
kind: MaintPage
metadata:
  name: example
spec:
  maintpageconfig:
    maintpagetoggle: false
    maintpageimage: quay.io/daein/maintpage:latest
  appconfig:  
    appname: httpd
    appimage: quay.io/daein/prodpage:latest
EOF
~~~

* Check delpoyed application pod(named as httpd), application service and maintenance page pod(<cr name>-maintpage-pod format)
~~~
# oc get pod
NAME                                  READY     STATUS    RESTARTS   AGE
example-maintpage-pod                 1/1       Running   0          20s
httpd-784b46459b-spfzg                1/1       Running   0          19s
maintpage-operator-6df9d9c85c-s7xmw   1/1       Running   0          40s

# oc get svc httpd
NAME      TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
httpd     ClusterIP   172.30.131.106   <none>        8080/TCP   21m
~~~

### Test to change the mode from production to maintenance
* Check created MaintPage CR, look `Status.Maintpublishstatus` is `Not Published`
~~~
# oc describe maintpage example
Name:         example
Namespace:    maintpage-operator
Labels:       <none>
Annotations:  <none>
API Version:  maintpage.example.com/v1alpha1
Kind:         MaintPage
...
Spec:
  Appconfig:
    Appimage:  quay.io/daein/prodpage:latest
    Appname:   httpd
  Maintpageconfig:
    Maintpageimage:   quay.io/daein/maintpage:latest
    Maintpagetoggle:  false
Status:
  Maintpublishstatus:  Not Published
Events:                <none>
~~~

* Monitor the showing page during this test as following command, 172.30.131.106 is httpd service IP, it would be different by installed cluster.
~~~
# while :; do echo "$(date '+%H:%M:%S'): $(curl -s http://172.30.131.106:8080)" ;sleep 1; done
12:34:53: Production Page !
12:34:54: Production Page !
12:34:55: Production Page !
...
~~~

* Update `maintpageconfig.maintpagetoggle: true` to switch with maintenance mode.
~~~
# oc edit maintpage example
...
spec:
  ...
  maintpageconfig:
    ...
    maintpagetoggle: true
~~~

* Verify the showing page is changed as maintenance one
~~~
# while :; do echo "$(date '+%H:%M:%S'): $(curl -s http://172.30.131.106:8080)" ;sleep 1; done
12:34:53: Production Page !
12:34:54: Production Page !
12:34:55: Production Page !
...
12:37:49: Maintenance Page !
12:37:50: Maintenance Page !
~~~

* Verify the status of MaintPage CR after changes, you can see `Status.Maintpublishstatus: Published`
~~~
# oc describe maintpage example
Name:         example
Namespace:    maintpage-operator
Labels:       <none>
Annotations:  <none>
API Version:  maintpage.example.com/v1alpha1
Kind:         MaintPage
...
Spec:
  Appconfig:
    Appimage:  quay.io/daein/prodpage:latest
    Appname:   httpd
  Maintpageconfig:
    Maintpageimage:   quay.io/daein/maintpage:latest
    Maintpagetoggle:  true
Status:
  Maintpublishstatus:  Published
Events:                <none>
~~~
