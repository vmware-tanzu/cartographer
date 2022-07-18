package v2alpha1

import apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

type Param struct {
	// Name of the parameter.
	// Blueprints must specify params used by the template or child components
	Name string `json:"name"`

	// Value of the parameter.
	// If specified, parent parameters are ignored.
	// It's invalid to provide a value in a blueprint with a spec.template specified
	Value *apiextensionsv1.JSON `json:"value,omitempty"`

	// Default of the parameter.
	// Causes the parameter to be optional
	Default *apiextensionsv1.JSON `json:"default,omitempty"`

	// Description of the parameter
	Description string `json:"description,omitempty"`
}
