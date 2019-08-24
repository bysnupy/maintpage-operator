package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MaintPageSpec defines the desired state of MaintPage
// +k8s:openapi-gen=true
type MaintPageSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
        AppName  string `json:"appname"`
        AppImage string `json:"appimage"`
        MaintPage bool  `json:"maintpage"`
}

// MaintPageStatus defines the observed state of MaintPage
// +k8s:openapi-gen=true
type MaintPageStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
        MaintPublishStatus string `json:"maintpublishstatus"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MaintPage is the Schema for the maintpages API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type MaintPage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MaintPageSpec   `json:"spec,omitempty"`
	Status MaintPageStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MaintPageList contains a list of MaintPage
type MaintPageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MaintPage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MaintPage{}, &MaintPageList{})
}
