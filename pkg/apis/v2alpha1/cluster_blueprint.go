// +versionName=v2alpha1
// +groupName=carto.run
// +kubebuilder:object:generate=true

package v2alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO:
// 	 * ParamRenames!
//   * Inputs!
//   * Options can't select on Owner, only parameters
//	 * What to do with healthrules for a compound blueprint.

// ClusterBlueprint represents a component within Cartographer Todo: be less asinine
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
	//A Template may reference the inputs by name and then follow the input's
	// schema, eg: $(inputs.my-source.url)$
	//
	// A blueprint can map
	// Note: If your compound blueprint does not expect any inputs, it can not be used as a
	// component to a parent blueprint.
	//Inputs Inputs `json:"inputs,omitempty"`

	// TypeRef refers to an object describing the contract this blueprint can fulfill
	// This is optional, however without an output, this Blueprint cannot be the cause of
	// a reconciliation of sibling components in a parent blueprint.
	// Templates can specify an output mapping via TemplateSpec.OutputMapping.
	// The last component in Components must match the type of TypeRef, as this is the
	// component that is used for this blueprint's output.
	TypeRef BlueTypeRef `json:"typeRef,omitempty"`

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

type BlueTypeRef struct {
	// Name of the ClusterBlueprintType that defines the output type of this blueprint.
	Name string `json:"name"`
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

	// Options is a list of template names and Selector.
	// A template will be selected if the workload matches the specified selector.
	// Only one template can be selected.
	// Only one of BlueprintRef and Options can be specified.
	// Minimum number of items in list is two.
	// +kubebuilder:validation:MinItems=2
	Options []TemplateOption `json:"options,omitempty"`
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
