package v2alpha1

type TemplateOption struct {
	// Name of the template to apply
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Selector is a criteria to match against the params list.
	Selector Selector `json:"selector"`
}
