// +versionName=v2alpha1
// +groupName=carto.run
// +kubebuilder:object:generate=true

package v2alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO
//	 * Outputs!
//   * Add options back!
//	 * Implement health rules for blueprints - also think about how this impacts our new status object
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
	OutputTypeRef OutputTypeRef `json:"outputTypeRef"`

	// Components are a list of sub-blueprints and templates this blueprint
	// creates and maintains during the lifetime of the OwnerObject.
	// This cannot be specified alongside Template
	Components []Component `json:"components,omitempty"`

	// Todo: opinions about template.template? resource.template instead?
	Template TemplateSpec `json:"template,omitempty"`

	// Params overrides for sub-blueprints and templates.
	Params []BlueprintParam `json:"params,omitempty"`

	// ServiceAccountName refers to the Service account with permissions to create resources
	// submitted by the supply chain.
	//
	// If not set, Cartographer will use serviceAccountName from supply chain.
	//
	// If that is also not set, Cartographer will use the default service account in the
	// workload's namespace.
	// +optional
	ServiceAccountRef ServiceAccountRef `json:"serviceAccountRef,omitempty"`
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

// Component represents a subcomponent
// Note: There are no params specified at this level. See BlueprintSpec.Params
type Component struct {
	// Name of the component. Used as a reference for inputs.
	// Template components are identified by this name in the BlueprintStatus
	Name string `json:"name"`

	// BlueprintRef identifies the template used to produce this resource
	BlueprintRef BlueprintRef `json:"blueprintRef"`
}

type BlueprintRef struct {
	// Name of the blueprint
	// Only one of Name and Options can be specified.  // todo: options
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name,omitempty"`
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
