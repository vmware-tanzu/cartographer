package v1alpha1

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterBlueprintTypeSpec defines the desired state of ClusterBlueprintType
type ClusterBlueprintTypeSpec struct {
	// Qualifier is provided to avoid name collisions when blueprint authors
	// start creating new types on a platform.
	// There is a validation rule that metadata.name must be of
	// the form: <qualifier>.<name> or <qualifier>-<name>.
	// If the qualifier is omitted, then just <name> will suffice.
	// Note: For TAP, this should be "tap" to avoid collisions with blueprint
	// authors. We recommend other platforms follow this pattern also.
	Qualifier string `json:"qualifier,omitempty"`

	// Schema a JSON schema that is a valid representation of a type.
	// Due to a limitation in k8s CRD definitions, this field accepts
	// any valid JSON, however the validation will fail if it's not
	// JSONSchema as per apiextensions.JSONSchemaProps
	// (see: https://pkg.go.dev/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions@v0.24.2#JSONSchemaProps)
	// Todo: explain the problem with the absence of schema here, and semantic error checking
	Schema *apiextensionsv1.JSON `json:"schema"`

	// Description describes this output to provide documentation to consumers.
	Description string `json:"description,omitempty"`
}

// ClusterBlueprintTypeStatus defines the observed state of ClusterBlueprintType
type ClusterBlueprintTypeStatus struct {
	// Conditions follow k8s sig-arch guidelines: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
	//
	// # Possible Sub-Conditions are:
	// ## SchemaValid
	// Describes the validity of the spec.schema field. A reason "InvalidJSONAPISchema" means the
	// Schema could not be parsed, and the Message will explain the error. You must provide valid [OpenAPI v3 Schema](https://swagger.io/specification/#schema-object)
	Conditions []metav1.Condition `json:"conditions"`
}

// ClusterBlueprintType defines a valid output/input between Components
// +kubebuilder:object:root=true
// +kubebuilder:subresource:spec
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=clusterblueprinttypes,scope=Cluster,shortName=cbt
type ClusterBlueprintType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterBlueprintTypeSpec   `json:"spec,omitempty"`
	Status ClusterBlueprintTypeStatus `json:"status,omitempty"`
}

// ClusterBlueprintTypeList contains a list of ClusterBlueprintType
//+kubebuilder:object:root=true
type ClusterBlueprintTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterBlueprintType `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterBlueprintType{}, &ClusterBlueprintTypeList{})
}
