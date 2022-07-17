package v2alpha1

// +versionName=v2alpha1
// +groupName=carto.run
// +kubebuilder:object:generate=true

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterSelector represents a mechanism to bind a Blueprint to an OwnerObject
// +kubebuilder:object:root=true
// +kubebuilder:subresource:spec
// +kubebuilder:resource:path=clusterselectors,scope=Cluster,shortName=cs
type ClusterSelector struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ClusterSelectorSpec `json:"spec"`
}

type ClusterSelectorSpec struct {
	metav1.TypeMeta `json:",inline"`
	BlueprintRef    BlueprintRef       `json:"blueprintRef"`
	ParamMap        []ParameterMapping `json:"paramMap,omitempty"` // Todo Does this want to be an externally referenced CRD?

	// ServiceAccountName refers to the Service account with permissions to create resources
	// submitted by the ClusterBlueprint.
	// TODO: fixme docs.
	// If that is also not set, Cartographer will use the default service account in the
	// owner object's namespace.
	// +optional
	ServiceAccountRef ServiceAccountRef `json:"serviceAccountRef,omitempty"`

	Selector
}

type ParameterMapping struct {
	Name string `json:"name"`

	// DefaultValue makes this parameter optional
	// if already optional, overrides the default value
	DefaultValue string `json:"defaultValue,omitempty"`

	// Value set's the value. You cannot map an ownerObject value at the same time
	// Using this field lets you configure blueprints on a per "Mapping" basis.
	// This is the best place for operator configuration to live.
	// Todo: Validate this
	Value string `json:"value,omitempty"`

	// OwnerPath defines where in the Owner object this parameter is sourced from
	// Follow JSONPath Syntax
	OwnerPath string `json:"ownerPath,omitempty"`
}

// ClusterSelectorList
// +kubebuilder:object:root=true
type ClusterSelectorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterSelector `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&ClusterSelector{},
		&ClusterSelectorList{},
	)
}
