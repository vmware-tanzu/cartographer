// +versionName=v2alpha1
// +groupName=carto.run
// +kubebuilder:object:generate=true

package v2alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO:
//	 * Outputs!
//		* discuss: for a single output with structured content (fields), the current design says it must match one JSONPath in the
//		  stamped object. Is that sufficient?
//   * Inputs!
//	 * Implement health rules for blueprints - also think about how this impacts our new status object
//	 * What to do with Field matchers... can we nuke them?
//   * Nested Params
//   * Discuss Duck Typing vs OutputType
//		* Duck typing requires no extra CRDS
//		* Duck typing makes it easier to proliferate useless contracts
//	 * Discuss Param behavior, especially collisions
//   * Discuss Template's being one variation of a Blueprint vs being their own CRD
//   * Discuss Resource->Component
//   * Discuss lack of oneOf validation today (for spec.Components vs spec.Template)

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

	// OutputTypeRef refers to an object describing the contract this blueprint can fulfill
	// This is optional, however without an output, this Blueprint cannot be the cause of
	// a reconciliation of sibling components in a parent blueprint.
	OutputTypeRef OutputTypeRef `json:"outputTypeRef,omitempty"`

	// Params for templates and overrides for sub-blueprints.
	Params []Param `json:"params,omitempty"`

	// Components are a list of sub-blueprints and templates which this blueprint
	// creates and maintains during the lifetime of the OwnerObject.
	// If OutputTypeRef is specified, the last item in this list must emit that type.
	// If it doesn't, the condition (todo: document condition here)
	// One of Components or Template can be specified exclusively.
	Components []Component `json:"components,omitempty"`

	// Template is a definition of a resource this component stamps onto the cluster
	// One of Components or Template can be specified exclusively.
	// Todo: opinions about template.template? resource.template instead?

	Template TemplateSpec `json:"template,omitempty"`
}

type OutputTypeRef struct {
	// Name of the ClusterOutputType that defines the output type of this blueprint.
	Name string `json:"name"`
}

type BlueprintStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the Generation of this resource's spec that reconciled the contents of this status.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Calculated list of input params based on templates and overrides
	Calculated []CalculatedParam `json:"calculated,omitempty"`
}

// Component to a subcomponent
// Note: There are no params specified at this level. See BlueprintSpec.Params and TemplateSpec.Params
type Component struct {
	// Name of the component. Used as a reference for inputs.
	// Template components are identified by this name in the BlueprintStatus
	Name string `json:"name"`

	// BlueprintRef identifies the template used to produce this resource
	BlueprintRef BlueprintRef `json:"blueprintRef"`

	// Options is a list of template names and Selector.
	// A template will be selected if the workload matches the specified selector.
	// Only one template can be selected.
	// Only one of BlueprintRef and Options can be specified.
	// Minimum number of items in list is two.
	// +kubebuilder:validation:MinItems=2
	Options []TemplateOption `json:"options,omitempty"`
}

type BlueprintRef struct {
	// Name of the blueprint
	// Only one of Name and Options can be specified.  // todo: options
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name,omitempty"`

	// +kubebuilder:validation:Enum=ClusterBlueprint;ClusterTemplate
	Kind string ``
}

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
