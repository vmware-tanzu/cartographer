// +versionName=v2alpha1
// +groupName=carto.run
// +kubebuilder:object:generate=true

package v2alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=clusteroutputtypes,scope=Cluster,shortName=cb

// FIXME: just Type?
type ClusterOutputType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	// Default qualifier is "default"
	// Validation rule, metadata.name must be on the form: <qualifier>-<name> (or <qualifier>.<name>)
	Qualifier string `json:"qualifier,omitempty"`

	// Schema a JSON schema that is a valid representation of a type.
	// Although json schema is deep and complex, it's usually best to keep to a simple type definition
	// for consumers of your component
	//
	Schema SimpleJSONSchema `json:"schema"`

	// Description describes this output to ease consumption by others.
	Description string `json:"description,omitempty"`
}

// SimpleJSONSchema is meant to represent something like a json schema
// Todo: I'm not all that happy with this - definitely incomplete.
// Need a validation hook because nesting explodes in k8s crds
type SimpleJSONSchema struct {
	// Type is the kind of object expected
	// +kubebuilder:validation:Enum=Object;String;Number;integer;array;boolean
	Type string `json:"type"`

	// Properties of a complex kind
	Properties *SimpleJSONSchemaProps `json:"properties,omitempty"`
}

type SimpleJSONSchemaProps struct {
	Name   string           `json:"name"`
	Schema SimpleJSONSchema `json:"schema"`


// ClusterOutputTypeList is a collection of ClusterOutputType
// +kubebuilder:object:root=true
type ClusterOutputTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterOutputType `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&ClusterOutputType{},
		&ClusterOutputTypeList{},
	)
}
