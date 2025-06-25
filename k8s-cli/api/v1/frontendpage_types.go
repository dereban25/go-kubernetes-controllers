package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FrontendPageSpec defines the desired state of FrontendPage
type FrontendPageSpec struct {
	// Title of the frontend page
	Title string `json:"title"`

	// Description of the frontend page
	Description string `json:"description"`

	// URL path for the frontend page
	Path string `json:"path"`

	// Template to use for rendering
	Template string `json:"template,omitempty"`

	// Configuration for the frontend page
	Config map[string]string `json:"config,omitempty"`

	// Replicas for the frontend deployment
	Replicas int32 `json:"replicas,omitempty"`

	// Image for the frontend container
	Image string `json:"image,omitempty"`
}

// FrontendPageStatus defines the observed state of FrontendPage
type FrontendPageStatus struct {
	// Phase represents the current phase of the FrontendPage
	Phase string `json:"phase,omitempty"`

	// Ready indicates if the frontend page is ready
	Ready bool `json:"ready,omitempty"`

	// URL where the frontend page is accessible
	URL string `json:"url,omitempty"`

	// Conditions represent the latest available observations
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// DeploymentName is the name of the created deployment
	DeploymentName string `json:"deploymentName,omitempty"`

	// ServiceName is the name of the created service
	ServiceName string `json:"serviceName,omitempty"`

	// LastUpdated timestamp
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Title",type="string",JSONPath=".spec.title"
//+kubebuilder:printcolumn:name="Path",type="string",JSONPath=".spec.path"
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
//+kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// FrontendPage is the Schema for the frontendpages API
type FrontendPage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FrontendPageSpec   `json:"spec,omitempty"`
	Status FrontendPageStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FrontendPageList contains a list of FrontendPage
type FrontendPageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FrontendPage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FrontendPage{}, &FrontendPageList{})
}
