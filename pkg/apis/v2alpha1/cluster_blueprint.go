// +versionName=v2alpha1
// +groupName=carto.run
// +kubebuilder:object:generate=true

package v2alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO:
//   * Inputs!
//   * Options can't select on Owner, only parameters
//	 * Match the thris param design: https://gist.github.com/squeedee/7a5bce7f52147afc5c9ba37a061685d6#file-2-all-components-define-their-params-yaml
//	 * Document how the last resource is the source of a Healthrule and the output type

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

	// OutputTypeRef refers to an object describing the contract this blueprint can fulfill
	// This is optional, however without an output, this Blueprint cannot be the cause of
	// a reconciliation of sibling components in a parent blueprint.
	OutputTypeRef OutputTypeRef `json:"outputTypeRef,omitempty"`

	// Components are a list of child blueprints which this blueprint
	// creates and maintains during the lifetime of the OwnerObject.
	// If OutputTypeRef is specified, the last item in this list must emit that type.
	// If it doesn't, the condition (todo: document condition here)
	// One of Components or Template can be specified exclusively.
	Components Components `json:"components,omitempty"`

	// Template is a definition of a resource this component stamps onto the cluster
	// One of Components or Template can be specified exclusively.
	// Todo: explain the problem with the absence of oneOf and semantic error checking
	// Todo: opinions about template.template? resource.template instead?
	Template TemplateSpec `json:"template,omitempty"`
}

type OutputTypeRef struct {
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
