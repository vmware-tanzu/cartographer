package v2alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type TemplateSpec struct {
	// Template defines a resource template for a Kubernetes Resource or
	// Custom Resource which is applied to the server each time
	// the blueprint is applied. Templates support simple value
	// interpolation using the $()$ marker format. For more
	// information, see: https://cartographer.sh/docs/latest/templating/
	// You cannot define both Template and Ytt at the same time.
	// You should not define the namespace for the resource - it will automatically
	// be created in the owner namespace. If the namespace is specified and is not
	// the owner namespace, the resource will fail to be created.
	// +kubebuilder:pruning:PreserveUnknownFields
	Template *runtime.RawExtension `json:"template,omitempty"`

	// Ytt defines a resource template written in `ytt` for a Kubernetes Resource or
	// Custom Resource which is applied to the server each time
	// the blueprint is applied. Templates support simple value
	// interpolation using the $()$ marker format. For more
	// information, see: https://cartographer.sh/docs/latest/templating/
	// You cannot define both Template and Ytt at the same time.
	// You should not define the namespace for the resource - it will automatically
	// be created in the owner namespace. If the namespace is specified and is not
	// the owner namespace, the resource will fail to be created.
	Ytt string `json:"ytt,omitempty"`

	// HealthRule specifies rubric for determining the health of a resource
	// stamped by this template
	// +optional
	HealthRule *HealthRule `json:"healthRule,omitempty"`
}

// HealthRule specifies rubric for determining the health of a resource.
// One of AlwaysHealthy, SingleConditionType or MultiMatch must be specified.
type HealthRule struct {
	// AlwaysHealthy being set indicates the resource should always be considered healthy
	// +optional
	AlwaysHealthy *runtime.RawExtension `json:"alwaysHealthy,omitempty"`

	// SingleConditionType names a single condition which, when True indicates the resource
	// is healthy. When False it is unhealthy. Otherwise, healthiness is Unknown.
	// +optional
	SingleConditionType string `json:"singleConditionType,omitempty"`

	// MultiMatch specifies explicitly which conditions and/or fields should be used
	// to determine healthiness.
	// +optional
	MultiMatch *MultiMatchHealthRule `json:"multiMatch,omitempty"`
}

// MultiMatchHealthRule is a pair of HealthMatchRule defining when a resource should be considered healthy or unhealthy
type MultiMatchHealthRule struct {
	// Healthy is a HealthMatchRule which stipulates requirements, ALL of which must be met for the resource to be
	// considered healthy.
	Healthy HealthMatchRule `json:"healthy"`
	// Unhealthy is a HealthMatchRule which stipulates requirements, ANY of which, when met, indicate that the resource
	// should be considered unhealthy.
	Unhealthy HealthMatchRule `json:"unhealthy"`
}

// HealthMatchRule specifies a rule for determining the health of a resource
type HealthMatchRule struct {
	// MatchConditions are the conditions and statuses to read.
	// +optional
	MatchConditions []ConditionRequirement `json:"matchConditions,omitempty"`
	// MatchFields stipulates a FieldSelectorRequirement for this rule.
	// +optional
	MatchFields []HealthMatchFieldSelectorRequirement `json:"matchFields,omitempty"`
}

type HealthMatchFieldSelectorRequirement struct {
	FieldSelectorRequirement `json:",inline"`
	// MessagePath is specified in jsonpath format. It is evaluated against the resource to provide a message in the
	// owner's resource condition if it is the first matching requirement that determine the current ResourcesHealthy
	// condition status.
	MessagePath string `json:"messagePath,omitempty"`
}

type ConditionRequirement struct {
	// Type is the type of the condition
	Type string `json:"type"`
	// Status is the status of the condition
	Status metav1.ConditionStatus `json:"status"`
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
