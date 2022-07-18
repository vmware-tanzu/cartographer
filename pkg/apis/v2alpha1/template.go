package v2alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type TemplateSpec struct {
	Templateable `json:",inline"`

	// HealthRule specifies rubric for determining the health of a resource
	// stamped by this template
	// +optional
	HealthRule *HealthRule `json:"healthRule,omitempty"`

	// An output mapping connects fields in the stamped resource with the
	// structure of the ClusterBlueprintType specified for this Component.
	// With only one entry in the mapping, it's possible to map a simple value
	// onto a simple type, or a complex, value onto a complex type (a one to one mapping).
	// When the resource's results do not match the exact shape of the ClusterBlueprintType,
	// you can use multiple mappings to coerce the correct shape.
	// Todo: examples in docs and a link.
	OutputMapping OutputMapping `json:"outputMapping,omitempty"`
}

type OutputMapping []OutputReference

type OutputReference struct {
	// Path is a JSONPath that represents the field in the OutputType that is fulfilled by Path
	Path string `json:"path"`
	// ResourcePath	is a JSONPath that represents where to find the value in the stamped resource.
	// ResourcePath can refer to a simple or complex type,
	ResourcePath string `json:"resourcePath"`
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
