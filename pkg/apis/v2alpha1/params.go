package v2alpha1

import apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

type Param struct {
	// Name of the parameter.
	// Template blueprints must specify params to use them
	// Non-Template blueprints can modify template parameters by specifying parameters
	Name string `json:"name"`

	// Value of the parameter.
	// If specified, parent properties are ignored.
	// If multiple children exist that specify Value,
	// this must be set, otherwise a "ParametersReady:False, Reason: ParameterValueCollision"
	// condition occurs.
	Value *apiextensionsv1.JSON `json:"value,omitempty"`

	// DefaultValue of the parameter.
	// Causes the parameter to be optional
	// If multiple children exist that specify DefaultValue, and Value is not set,
	// this must be set.
	// Otherwise a "ParametersReady:False, Reason: ParameterDefaultValueCollision"
	// condition occurs.
	DefaultValue *apiextensionsv1.JSON `json:"default,omitempty"`

	// Description of the parameter
	// If a children exist, will hide child descriptions
	// Otherwise, child descriptions are joined with newlines and
	// that result is used
	Description string `json:"description,omitempty"`
}
