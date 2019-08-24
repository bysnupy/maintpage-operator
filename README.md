# Maintpage Operator

Maintpage Operator provide automation to control over showing maintenance page.

## Define CR MaintPage as follows

~~~
apiVersion: maintpage.example.com/v1alpha1
kind: MaintPage
metadata:
  name: maintpage
spec:
  appname: httpd
  appimage: registry.redhat.io/rhscl/httpd-24-rhel7:latest
  maintpage: false
~~~

Spec name|Description
-|-
appname| Application resource name
appimage| Container image to run
maintpage| Togle to control showing the maintenance page

## 
