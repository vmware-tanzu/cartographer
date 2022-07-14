package v2alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// Selector is the collection of selection fields used congruously to specify
// the selection of a template Option. In a future API revision, it will also
// be used to specify selection of a target Owner.
// See: LegacySelector
type Selector struct {
	metav1.LabelSelector `json:",inline"`

	// MatchFields is a list of field selector requirements. The requirements are ANDed.
	// +optional
	MatchFields []FieldSelectorRequirement `json:"matchFields,omitempty"`
}

type FieldSelectorRequirement struct {
	// Key is the JSON path in the workload to match against.
	// e.g. for workload: "workload.spec.source.git.url",
	// e.g. for deliverable: "deliverable.spec.source.git.url"
	// +kubebuilder:validation:MinLength=1
	Key string `json:"key"`

	// Operator represents a key's relationship to a set of values.
	// Valid operators are In, NotIn, Exists and DoesNotExist.
	// +kubebuilder:validation:Enum=In;NotIn;Exists;DoesNotExist
	Operator FieldSelectorOperator `json:"operator"`

	// Values is an array of string values. If the operator is In or NotIn,
	// the values array must be non-empty. If the operator is Exists or DoesNotExist,
	// the values array must be empty.
	Values []string `json:"values,omitempty"`
}

type FieldSelectorOperator string
