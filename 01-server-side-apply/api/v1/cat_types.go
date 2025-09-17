package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Cat is a custom resource with Spec and Status
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
type Cat struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of Cat
	Spec CatSpec `json:"spec,omitempty"`

	// Status represents the status of the Cat resource
	Status CatStatus `json:"status,omitempty"`
}

// CatSpec defines the desired state of Cat
type CatSpec struct {
	// Breed is the breed of the cat
	// +optional
	Breed string `json:"breed,omitempty"`

	// Color is the color of the cat
	// +optional
	Color string `json:"color,omitempty"`

	// Age is the age of the cat in years
	// +optional
	Age int32 `json:"age,omitempty"`
}

// CatStatus defines the observed state of Cat
type CatStatus struct {
	// Conditions represent the latest available observations of Cat's state
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=32
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// CatList contains a list of Cat objects
// +kubebuilder:object:root=true
type CatList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cat `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cat{}, &CatList{})
}
