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

	// TypeRef is the name of the ClusterBlueprintType that can be enforced for this parameter.
	// TypeRef is optional, but cannot be specified at the same time as Schema
	TypeRef BlueprintTypeRef `json:"typeRef,omitempty"`

	// Schema is a JSON schema that is a valid representation of this input.
	// Schema is optional, but cannot be specified at the same time as TypeRef
	// This field allows you to enforce a type without needing a shared contract in
	// a ClusterBlueprintType.
	Schema *apiextensionsv1.JSON `json:"schema"`

	// Schemas represent named subschemas, great for using with openapi $ref syntax,
	// especially in oneOf, anyOf, allOf etc expressions. See:
	// https://swagger.io/docs/specification/data-models/oneof-anyof-allof-not/
	Schemas map[string]*apiextensionsv1.JSON `json:"schemas"`

	// Description of the parameter
	Description string `json:"description,omitempty"`
}
