// Copyright 2021 VMware
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	FieldSelectorOpIn           FieldSelectorOperator = "In"
	FieldSelectorOpNotIn        FieldSelectorOperator = "NotIn"
	FieldSelectorOpExists       FieldSelectorOperator = "Exists"
	FieldSelectorOpDoesNotExist FieldSelectorOperator = "DoesNotExist"
)

type OwnerStatus struct {
	// ObservedGeneration refers to the metadata.Generation of the spec that resulted in
	// the current `status`.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions describing this resource's reconcile state. The top level condition is
	// of type `Ready`, and follows these Kubernetes conventions:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

type TemplateParams []TemplateParam

type TemplateParam struct {
	// Name of a parameter the template accepts from the
	// Blueprint or Owner.
	Name string `json:"name"`

	// DefaultValue of the parameter.
	// Causes the parameter to be optional; If the Owner or Template
	// does not specify this parameter, this value is used.
	DefaultValue apiextensionsv1.JSON `json:"default"`
}

type OwnerParam struct {
	// Name of the parameter.
	// Should match a blueprint or template parameter name.
	Name string `json:"name"`

	// Value of the parameter.
	Value apiextensionsv1.JSON `json:"value"`
}

type BlueprintParam struct {
	// Name of the parameter.
	// Should match a template parameter name.
	Name string `json:"name"`

	// Value of the parameter.
	// If specified, owner properties are ignored.
	Value *apiextensionsv1.JSON `json:"value,omitempty"`

	// DefaultValue of the parameter.
	// Causes the parameter to be optional; If the Owner does not specify
	// this parameter, this value is used.
	DefaultValue *apiextensionsv1.JSON `json:"default,omitempty"`
}

func (p *BlueprintParam) validate() error {
	if p.bothValuesSet() || p.neitherValueSet() {
		return fmt.Errorf("param [%s] is invalid: must set exactly one of value and default", p.Name)
	}
	return nil
}

func (p *BlueprintParam) bothValuesSet() bool {
	return p.DefaultValue != nil && p.Value != nil
}

func (p *BlueprintParam) neitherValueSet() bool {
	return p.DefaultValue == nil && p.Value == nil
}

type ResourceInputs struct {
	// Sources is a list of references to other 'source' resources in this list.
	// A source resource has the kind ClusterSourceTemplate
	//
	// In a template, sources can be consumed as:
	//    $(sources.<name>.url)$ and $(sources.<name>.revision)$
	//
	// If there is only one source, it can be consumed as:
	//    $(source.url)$ and $(source.revision)$
	Sources []ResourceReference `json:"sources,omitempty"`

	// Images is a list of references to other 'image' resources in this list.
	// An image resource has the kind ClusterImageTemplate
	//
	// In a template, images can be consumed as:
	//   $(images.<name>.image)$
	//
	// If there is only one image, it can be consumed as:
	//   $(image)$
	Images []ResourceReference `json:"images,omitempty"`

	// Configs is a list of references to other 'config' resources in this list.
	// A config resource has the kind ClusterConfigTemplate
	//
	// In a template, configs can be consumed as:
	//   $(configs.<name>.config)$
	//
	// If there is only one image, it can be consumed as:
	//   $(config)$
	Configs []ResourceReference `json:"configs,omitempty"`

}

type ResourceReference struct {
	Name     string `json:"name"`
	Resource string `json:"resource"`
}

type Source struct {
	// Source code location in a git repository.
	// +optional
	Git *GitSource `json:"git,omitempty"`

	// OCI Image in a repository, containing the source code to
	// be used throughout the supply chain.
	// +optional
	Image *string `json:"image,omitempty"`

	// Subpath inside the Git repository or Image to treat as the root
	// of the application. Defaults to the root if left empty.
	// +optional
	Subpath *string `json:"subPath,omitempty"`
}

type GitSource struct {
	URL *string `json:"url,omitempty"`
	Ref *GitRef `json:"ref,omitempty"`
}

type GitRef struct {
	Branch *string `json:"branch,omitempty"`
	Tag    *string `json:"tag,omitempty"`
	Commit *string `json:"commit,omitempty"`
}

type ObjectReference struct {
	Kind       string `json:"kind,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
	Name       string `json:"name,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
}

type ServiceAccountRef struct {
	// Name of the service account being referred to
	Name string `json:"name"`
	// Namespace of the service account being referred to
	// if omitted, the Owner's namespace is used.
	Namespace string `json:"namespace,omitempty"`
}

func GetAPITemplate(templateKind string) (client.Object, error) {
	var template client.Object

	switch templateKind {
	case "ClusterSourceTemplate":
		template = &ClusterSourceTemplate{}
	case "ClusterImageTemplate":
		template = &ClusterImageTemplate{}
	case "ClusterConfigTemplate":
		template = &ClusterConfigTemplate{}
	case "ClusterTemplate":
		template = &ClusterTemplate{}
	case "ClusterDeploymentTemplate":
		template = &ClusterDeploymentTemplate{}
	default:
		return nil, fmt.Errorf("resource does not have valid kind: %s", templateKind)
	}
	return template, nil
}

type TemplateOption struct {
	// Name of the template to apply
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	Selectable `json:",inline"`
}

type OptionSelector struct {
	// MatchFields is a list of field selector requirements. The requirements are ANDed.
	// +kubebuilder:validation:MinItems=1
	MatchFields []FieldSelectorRequirement `json:"matchFields"`
}

type FieldSelectorRequirement struct {
	// Key is the JSON path in the workload to match against.
	// e.g. for workload: "workload.spec.source.git.url",
	// e.g. for deliverable: "deliverable.spec.source.git.url"
	// +kubebuilder:validation:MinLength=1
	Key string `json:"key"`

	// Operator represents a key's relationship to a set of values.
	// Valid operators are In, NotIn, Exists and DoesNotExist.
	// +kubebuilder:validation:Enum=In;NotIn;Exists;DoesNotExist
	Operator FieldSelectorOperator `json:"operator"`

	// Values is an array of string values. If the operator is In or NotIn,
	// the values array must be non-empty. If the operator is Exists or DoesNotExist,
	// the values array must be empty.
	Values []string `json:"values,omitempty"`
}

type RealizedResource struct {
	// Name is the name of the resource in the blueprint
	Name string `json:"name"`

	// StampedRef is a reference to the object that was created by the resource
	StampedRef *corev1.ObjectReference `json:"stampedRef,omitempty"`

	// TemplateRef is a reference to the template used to create the object in StampedRef
	TemplateRef *corev1.ObjectReference `json:"templateRef,omitempty"`

	// Inputs are references to resources that were used to template the object in StampedRef
	Inputs []Input `json:"inputs,omitempty"`

	// Outputs are values from the object in StampedRef that can be consumed by other resources
	Outputs []Output `json:"outputs,omitempty"`
}

type Input struct {
	// Name is the name of the resource in the blueprint whose output the resource consumes as an input
	Name string `json:"name"`
}

type Output struct {
	// Name is the output type generated from the resource [url, revision, image or config]
	Name string `json:"name"`

	// Preview is a preview of the value of the output
	Preview string `json:"preview"`

	// Digest is a sha256 of the full value of the output
	Digest string `json:"digest"`

	// LastTransitionTime is a timestamp of the last time the value changed
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
}

// Selectors are the collection of selection fields used congruously to specify
// the selection of a target Owner, or ensure the owner matches a selection for a template Option
// TODO: when we switch to v1alpha2 we can rename this just "Selector"
type Selectors struct {
	// Specifies the label key-value pairs used to select owners
	// See: https://cartographer.sh/docs/v0.1.0/architecture/#selectors
	// +optional
	Selector map[string]string `json:"selector,omitempty"`

	// Specifies the requirements used to select owners based on their labels
	// See: https://cartographer.sh/docs/v0.1.0/architecture/#selectors
	// +optional
	SelectorMatchExpressions []metav1.LabelSelectorRequirement `json:"selectorMatchExpressions,omitempty"`

	// Specifies the requirements used to select owners based on their fields
	// See: https://cartographer.sh/docs/v0.1.0/architecture/#selectors
	// +optional
	SelectorMatchFields []FieldSelectorRequirement `json:"selectorMatchFields,omitempty"`
}
