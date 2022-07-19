package v2alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// Selector is the collection of selection fields used to specify
// an owner object
type Selector struct {
	metav1.LabelSelector `json:",inline"`

	// MatchParams is a list of param selector requirements. The requirements are ANDed.
	// +optional
	MatchParams []FieldSelectorRequirement `json:"matchFields,omitempty"`
}

type FieldSelectorRequirement struct {
	// Name is the parameter's name
	// A parameter with this name must be specified in BlueprintSpec.Params
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Operator represents a parameter's relationship to a set of values.
	// Valid operators are In, NotIn, Exists and DoesNotExist.
	// +kubebuilder:validation:Enum=In;NotIn;Exists;DoesNotExist
	Operator FieldSelectorOperator `json:"operator"`

	// Values is an array of string values. If the operator is In or NotIn,
	// the values array must be non-empty. If the operator is Exists or DoesNotExist,
	// the values array must be empty.
	Values []string `json:"values,omitempty"`
}

type FieldSelectorOperator string
