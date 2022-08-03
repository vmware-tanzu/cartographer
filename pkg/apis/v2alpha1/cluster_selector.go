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
	// OwnerSelector is the criteria used to match an Owner to the BlueprintRef
	// todo: explain selection criteria, precedence and how version is only used for representation
	OwnerSelector `json:"ownerSelector"`

	// BlueprintRef selects a specific blueprint for the matched OwnerSelector
	BlueprintRef BlueprintRef `json:"blueprintRef"`

	// Params maps Blueprint parameters to the specific Owner specified in OwnerSelector's TypeMeta.
	Params []ParameterMapping `json:"params,omitempty"`

	// StatusMapping represents the mechanism used to record the status of the Blueprint's imprint back
	// to the Owner.
	// If omitted, the Owner is not updated by Cartographer
	OwnerStatusMapping StatusMapping `json:"statusMapping,omitempty"`

	// AdditionalStatusMappings provide a mechanism to add additional status objects per matched owner object.
	// Note: We can perhaps implement this at a later date if OwnerStatusMapping proves to be insufficient for all
	// use cases.
	AdditionalStatusMappings []AdditionalStatusMapping `json:"additionalStatusMappings,omitempty"`

	// ServiceAccountName refers to the Service account with permissions to create resources
	// submitted by the ClusterBlueprint.
	// TODO: fixme docs.
	// If that is also not set, Cartographer will use the default service account in the
	// owner object's namespace.
	// +optional
	ServiceAccountRef ServiceAccountRef `json:"serviceAccountRef,omitempty"`
}

type OwnerSelector struct {
	metav1.TypeMeta `json:",inline"`
	Selector        `json:",inline"`
}

type ParameterMapping struct {
	Name string `json:"name"`

	// Default makes this parameter optional
	// if already optional, overrides the default value
	Default string `json:"default,omitempty"`

	// Value set's the value. You cannot map an ownerObject value at the same time
	// Using this field lets you configure blueprints on a per "Mapping" basis.
	// This is the best place for operator configuration to live.
	Value string `json:"value,omitempty"`

	// Path defines where in the Owner object this parameter is sourced from
	// using JSONPath syntax.
	Path string `json:"path,omitempty"`
}

// AdditionalStatusMapping provides a mechanism to create other status objects
// as a result of
type AdditionalStatusMapping struct {
	metav1.TypeMeta `json:",inline"`

	// TODO: we can always update .status, or we can let the template decide (make it a root template)
	// if we only template .status, then the new objects metadata.name/namespace could either match the
	// ownerObject's, or also have templating to generate them.
	StatusMapping `json:"inline"`
}

// StatusMapping is the template used to create the `status` field of the owner object
// or other object.
type StatusMapping struct {
	Templateable `json:",inline"`

	// TODO: we need some kind of reverse reference for condition's lastUpdatedAt
	// or we ensure a complete object is stored and not accessible
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
