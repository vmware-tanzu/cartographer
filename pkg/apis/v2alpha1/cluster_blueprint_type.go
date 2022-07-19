// +versionName=v2alpha1
// +groupName=carto.run
// +kubebuilder:object:generate=true

package v2alpha1

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterBlueprintType defines a valid output/input between Components
// +kubebuilder:object:root=true
// +kubebuilder:subresource:spec
// +kubebuilder:resource:path=clusterblueprinttypes,scope=Cluster,shortName=cb
type ClusterBlueprintType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec BlueprintTypeSpec `json:"spec"`
}

type BlueprintTypeSpec struct {
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

// ClusterOutputTypeList is a collection of ClusterBlueprintType
// +kubebuilder:object:root=true
type ClusterOutputTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterBlueprintType `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&ClusterBlueprintType{},
		&ClusterOutputTypeList{},
	)
}
