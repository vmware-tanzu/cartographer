// +versionName=v2alpha1
// +groupName=carto.run
// +kubebuilder:object:generate=true

package v2alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO:
//  * explain how healthrules are nested and continue to work with this spec
//  * try adding schema to inputs

// ClusterBlueprint represents a component within Cartographe
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=clusterblueprints,scope=Cluster,shortName=cb
type ClusterBlueprint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BlueprintSpec   `json:"spec"`
	Status BlueprintStatus `json:"status,omitempty"`
}

type BlueprintSpec struct {
	// Description of the blueprint
	// If not set, this does not reflect descriptions in child blueprints or templates
	Description string `json:"description,omitempty"`

	// Params specifies accepted parameters for the template.
	// Any parameter consumed in the template MUST be specified
	// as a Param
	Params []Param `json:"params,omitempty"`

	// Inputs specifies the input types and names required by this blueprint
	//
	// A Template may reference the inputs by name and then follow the input's
	// schema, eg: $(inputs.my-source.url)$
	//
	// Note: If your compound blueprint does not expect any inputs, it can not be used as a
	// component to a parent blueprint.
	Inputs BlueprintInputs `json:"inputs,omitempty"`

	// TypeRef refers to an object describing the contract this blueprint can fulfill
	// This is optional, however without an output, this Blueprint cannot be the cause of
	// a reconciliation of sibling components in a parent blueprint.
	// Templates can specify an output mapping via TemplateSpec.OutputMapping.
	// The last component in Components must match the type of TypeRef, as this is the
	// component that is used for this blueprint's output.
	TypeRef BlueprintTypeRef `json:"typeRef,omitempty"`

	// Components are a list of child blueprints managed by this blueprint.
	// If TypeRef is specified, the last item in this list must emit that type.
	// One of Components or Template can be specified exclusively.
	// The last Component in this list is assumed to be the Output for this blueprint.
	Components Components `json:"components,omitempty"`

	// Template is a definition of a resource this component stamps onto the cluster
	// One of Components or Template can be specified exclusively.
	// Todo: explain the problem with the absence of oneOf and semantic error checking
	Template TemplateSpec `json:"template,omitempty"`
}

type BlueprintInputs []BlueprintInput

type BlueprintInput struct {
	// Name is used to reference this input in templates and component definitions
	Name string `json:"name"`

	// Ref is the name of the ClusterBlueprintType that must be provided
	Ref BlueprintTypeRef `json:"ref"`

	// Description allows authors to describe this input
	Description string `json:"description,omitempty"`
}

type BlueprintTypeRef struct {
	// Name of the ClusterBlueprintType that defines the output type of this blueprint.
	Name string `json:"name"`

	// Description allows authors to describe the output of their blueprint
	Description string `json:"description,omitempty"`
}

type BlueprintStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the Generation of this resource's spec that reconciled the contents of this status.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Calculated list of input params based on templates and overrides
	Calculated []CalculatedParam `json:"calculated,omitempty"`
}

// Components are a list of child blueprints
type Components []Component

// Component to a subcomponent
// Note: There are no params specified at this level. See BlueprintSpec.Params and TemplateSpec.Params
type Component struct {
	// Name of the component. Used as a reference for inputs.
	// Template components are identified by this name in the BlueprintStatus
	Name string `json:"name"`

	// BlueprintRef identifies the template used to produce this resource
	// Only one of BlueprintRef and Options can be specified.
	BlueprintRef BlueprintRef `json:"blueprintRef"`

	// ParamRenames are a list of parameters that need to be renamed to satisfy the child blueprint
	// Any params specified in Blueprint.Params are passed to the child, however if the name doesn't match,
	// then use a param rename.
	ParamRenames ParamRename `json:"paramRenames,omitempty"`

	// Inputs specifies the input mapping from other components and Blueprint.Inputs
	Inputs ComponentInputs `json:"inputs,omitempty"`

	// Options is a list of template names and Selector.
	// A template will be selected if the workload matches the specified selector.
	// Only one template can be selected.
	// Only one of BlueprintRef and Options can be specified.
	// Minimum number of items in list is two.
	// +kubebuilder:validation:MinItems=2
	Options []TemplateOption `json:"options,omitempty"`
}

type ComponentInputs []ComponentInput

type ComponentInput struct {
	// Name of this input. It must match with an input in the child component.
	Name string `json:"name"`

	// ValueFrom specifies which input to use.
	ValueFrom InputValueFrom `json:"valueFrom"`
}

// InputValueFrom describes the source of an input
// You can not specify both Component and Input at the same time.
type InputValueFrom struct {
	// Component specifies a sibling component's name as the input.
	Component string `json:"component,omitempty"`

	// Input names a top level ClusterBlueprintSpec.Inputs as the input.
	Input string `json:"input,omitempty"`
}

type ParamRename struct {
	// From is the name of the parameter in the current blueprint
	From string `json:"from"`

	// To is the name of the parameter in the child component
	To string `json:"to"`
}

// ClusterBlueprintList is a list of ClusterBlueprint
// +kubebuilder:object:root=true
type ClusterBlueprintList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterBlueprint `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&ClusterBlueprint{},
		&ClusterBlueprintList{},
	)
}
