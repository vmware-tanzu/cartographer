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

// +versionName=v1alpha1
// +groupName=carto.run
// +kubebuilder:object:generate=true

package v1alpha1

import (
	"fmt"
	"reflect"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	SupplyChainReady          = "Ready"
	SupplyChainTemplatesReady = "TemplatesReady"
)

const (
	ReadyTemplatesReadyReason    = "Ready"
	NotFoundTemplatesReadyReason = "TemplatesNotFound"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

type ClusterSupplyChain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	// Spec describes the suppply chain.
	// More info: https://cartographer.sh/docs/latest/reference/workload/#clustersupplychain
	Spec SupplyChainSpec `json:"spec"`

	// Status conforms to the Kubernetes conventions:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
	Status SupplyChainStatus `json:"status,omitempty"`
}

func (c *ClusterSupplyChain) validateNewState() error {
	names := make(map[string]bool)

	if err := c.validateParams(); err != nil {
		return err
	}

	for _, resource := range c.Spec.Resources {
		if _, ok := names[resource.Name]; ok {
			return fmt.Errorf("duplicate resource name [%s] found", resource.Name)
		}
		names[resource.Name] = true
	}

	for _, resource := range c.Spec.Resources {
		optionNames := make(map[string]bool)
		for _, option := range resource.TemplateRef.Options {
			if _, ok := optionNames[option.Name]; ok {
				return fmt.Errorf(
					"duplicate template name [%s] found in options for resource [%s]",
					option.Name,
					resource.Name,
				)
			}
			optionNames[option.Name] = true
		}
	}

	for _, resource := range c.Spec.Resources {
		if err := validateResourceTemplateRef(resource.TemplateRef); err != nil {
			return fmt.Errorf("error validating resource [%s]: %w", resource.Name, err)
		}
	}

	for _, resource := range c.Spec.Resources {
		if err := c.validateResourceRefs(resource.Sources, "ClusterSourceTemplate"); err != nil {
			return fmt.Errorf(
				"invalid sources for resource [%s]: %w",
				resource.Name,
				err,
			)
		}

		if err := c.validateResourceRefs(resource.Images, "ClusterImageTemplate"); err != nil {
			return fmt.Errorf(
				"invalid images for resource [%s]: %w",
				resource.Name,
				err,
			)
		}

		if err := c.validateResourceRefs(resource.Configs, "ClusterConfigTemplate"); err != nil {
			return fmt.Errorf(
				"invalid configs for resource [%s]: %w",
				resource.Name,
				err,
			)
		}
	}

	return nil
}

func (c *ClusterSupplyChain) validateParams() error {
	for _, param := range c.Spec.Params {
		err := param.validate()
		if err != nil {
			return err
		}
	}

	for _, resource := range c.Spec.Resources {
		for _, param := range resource.Params {
			err := param.validate()
			if err != nil {
				return fmt.Errorf("resource [%s] is invalid: %w", resource.Name, err)
			}
		}
	}

	return nil
}

func (c *ClusterSupplyChain) validateResourceRefs(references []ResourceReference, targetKind string) error {
	for _, ref := range references {
		referencedResource := c.getResourceByName(ref.Resource)
		if referencedResource == nil {
			return fmt.Errorf(
				"[%s] is provided by unknown resource [%s]",
				ref.Name,
				ref.Resource,
			)
		}
		if referencedResource.TemplateRef.Kind != targetKind {
			return fmt.Errorf(
				"resource [%s] providing [%s] must reference a %s",
				referencedResource.Name,
				ref.Name,
				targetKind,
			)
		}
	}
	return nil
}

func (c *ClusterSupplyChain) getResourceByName(name string) *SupplyChainResource {
	for _, resource := range c.Spec.Resources {
		if resource.Name == name {
			return &resource
		}
	}

	return nil
}

func validateResourceTemplateRef(ref SupplyChainTemplateReference) error {
	if ref.Name != "" && len(ref.Options) > 0 {
		return fmt.Errorf("exactly one of templateRef.Name or templateRef.Options must be specified, found both")
	}

	if ref.Name == "" && len(ref.Options) < 2 {
		if len(ref.Options) == 1 {
			return fmt.Errorf("templateRef.Options must have more than one option")
		}
		return fmt.Errorf("exactly one of templateRef.Name or templateRef.Options must be specified, found neither")
	}

	if err := validateResourceOptions(ref.Options); err != nil {
		return err
	}
	return nil
}

func validateResourceOptions(options []TemplateOption) error {
	for _, option := range options {
		if err := validateFieldSelectorRequirements(option.Selector.MatchFields); err != nil {
			return fmt.Errorf("error validating option [%s]: %w", option.Name, err)
		}
	}

	for _, option1 := range options {
		for _, option2 := range options {
			if option1.Name != option2.Name && reflect.DeepEqual(option1.Selector, option2.Selector) {
				return fmt.Errorf(
					"duplicate selector found in options [%s, %s]",
					option1.Name,
					option2.Name,
				)
			}
		}
	}

	return nil
}

func validateFieldSelectorRequirements(reqs []FieldSelectorRequirement) error {
	for _, req := range reqs {
		switch req.Operator {
		case FieldSelectorOpExists, FieldSelectorOpDoesNotExist:
			if len(req.Values) != 0 {
				return fmt.Errorf("cannot specify values with operator [%s]", req.Operator)
			}
		case FieldSelectorOpIn, FieldSelectorOpNotIn:
			if len(req.Values) == 0 {
				return fmt.Errorf("must specify values with operator [%s]", req.Operator)
			}
		default:
			return fmt.Errorf("operator [%s] is invalid", req.Operator)
		}

		if !validRequirementKey(req.Key) {
			return fmt.Errorf("requirement key [%s] is not a valid workload path", req.Key)
		}
	}
	return nil
}

func (c *ClusterSupplyChain) ValidateCreate() error {
	err := c.validateNewState()
	if err != nil {
		return fmt.Errorf("error validating clustersupplychain [%s]: %w", c.Name, err)
	}
	return nil
}

func (c *ClusterSupplyChain) ValidateUpdate(_ runtime.Object) error {
	err := c.validateNewState()
	if err != nil {
		return fmt.Errorf("error validating clustersupplychain [%s]: %w", c.Name, err)
	}
	return nil
}

func (c *ClusterSupplyChain) ValidateDelete() error {
	return nil
}

func (c *ClusterSupplyChain) GetSelector() map[string]string {
	return c.Spec.Selector
}

func GetSelectorsFromObject(o client.Object) []string {
	var res []string
	res = []string{}

	sc, ok := o.(*ClusterSupplyChain)
	if ok {
		for key, value := range sc.Spec.Selector {
			res = append(res, fmt.Sprintf("%s: %s", key, value))
		}
	}

	return res
}

type SupplyChainSpec struct {
	// Resources that are responsible for bringing the application to a
	// deliverable state.
	Resources []SupplyChainResource `json:"resources"`

	// Specifies the label key-value pairs used to select workloads
	// See: https://cartographer.sh/docs/v0.1.0/architecture/#selectors
	Selector map[string]string `json:"selector"`

	// Additional parameters.
	// See: https://cartographer.sh/docs/latest/architecture/#parameter-hierarchy
	// +optional
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

type SupplyChainResource struct {
	// Name of the resource. Used as a reference for inputs, as well as being
	// the name presented in workload statuses to identify this resource.
	Name string `json:"name"`

	// TemplateRef identifies the template used to produce this resource
	TemplateRef SupplyChainTemplateReference `json:"templateRef"`

	// Params are a list of parameters to provide to the template in TemplateRef
	// Template params do not have to be specified here, unless you want to
	// force a particular value, or add a default value.
	//
	// Parameters are consumed in a template with the syntax:
	//   $(params.<name>)$
	Params []BlueprintParam `json:"params,omitempty"`

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

var ValidSupplyChainTemplates = []client.Object{
	&ClusterSourceTemplate{},
	&ClusterImageTemplate{},
	&ClusterConfigTemplate{},
	&ClusterTemplate{},
}

type SupplyChainTemplateReference struct {
	// Kind of the template to apply
	//+kubebuilder:validation:Enum=ClusterSourceTemplate;ClusterImageTemplate;ClusterTemplate;ClusterConfigTemplate
	Kind string `json:"kind"`

	// Name of the template to apply
	// Only one of Name and Options can be specified.
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name,omitempty"`

	// Options is a list of template names and Selectors. The templates must all be of type Kind.
	// A template will be selected if the workload matches the specified Selector.
	// Only one template can be selected.
	// Only one of Name and Options can be specified.
	// +kubebuilder:validation:MinItems=2
	Options []TemplateOption `json:"options,omitempty"`
}

type TemplateOption struct {
	// Name of the template to apply
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Selector is a field query over a workload resource.
	Selector Selector `json:"selector"`
}

type Selector struct {
	// MatchFields is a list of field selector requirements. The requirements are ANDed.
	// +kubebuilder:validation:MinItems=1
	MatchFields []FieldSelectorRequirement `json:"matchFields"`
}

type FieldSelectorRequirement struct {
	// Key is the JSON path in the workload to match against.
	// e.g. "workload.spec.source.git.url"
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

type FieldSelectorOperator string

const (
	FieldSelectorOpIn           FieldSelectorOperator = "In"
	FieldSelectorOpNotIn        FieldSelectorOperator = "NotIn"
	FieldSelectorOpExists       FieldSelectorOperator = "Exists"
	FieldSelectorOpDoesNotExist FieldSelectorOperator = "DoesNotExist"
)

type SupplyChainStatus struct {
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true

type ClusterSupplyChainList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterSupplyChain `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&ClusterSupplyChain{},
		&ClusterSupplyChainList{},
	)
}

func validRequirementKey(key string) bool {
	validWorkloadPaths := map[string]bool{
		//source
		"workload.spec.source":                true,
		"workload.spec.source.git":            true,
		"workload.spec.source.git.url":        true,
		"workload.spec.source.git.ref":        true,
		"workload.spec.source.git.ref.branch": true,
		"workload.spec.source.git.ref.tag":    true,
		"workload.spec.source.git.ref.commit": true,
		"workload.spec.source.image":          true,
		"workload.spec.source.subPath":        true,
		//build
		"workload.spec.build": true,
		//image
		"workload.spec.image": true,
		//serviceAccountName
		"workload.spec.serviceAccountName": true,
	}

	validWorkloadPrefixes := []string{
		//params
		"workload.spec.params",
		//build
		"workload.spec.build.env",
		//env
		"workload.spec.env",
		//resources
		"workload.spec.resources",
		//serviceClaims
		"workload.spec.serviceClaims",
	}

	if validWorkloadPaths[key] {
		return true
	}

	for _, prefix := range validWorkloadPrefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}

	return false
}
