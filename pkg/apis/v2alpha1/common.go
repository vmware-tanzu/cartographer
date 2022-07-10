package v2alpha1

import apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

type ServiceAccountRef struct {
	// Name of the service account being referred to
	Name string `json:"name"`
	// Namespace of the service account being referred to
	// if omitted, the Owner's namespace is used.
	Namespace string `json:"namespace,omitempty"`
}

// CalculatedParam is one of the available parameters exposed by the template or sub-blueprints within
// this blueprint.
// ClusterSelectors further allow these parameters to be configured or mapped to OwnerResource fields.
type CalculatedParam struct {
	// Name of the parameter
	Name string `json:"name"`

	// Value of the parameter. If set, cannot be overridden by the ClusterSelectorMapping
	// Value is mutually exclusive with Default
	Value *apiextensionsv1.JSON `json:"value,omitempty"`

	// Default value of the parameter. If set, can be overridden by the ClusterSelectorMapping
	Default *apiextensionsv1.JSON `json:"default,omitempty"`

	// Description(s) of this parameter.
	// if multiple children exist, then they are joined with newlines
	Description string `json:"description,omitempty"`
}
