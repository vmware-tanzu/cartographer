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
	Spec              SupplyChainSpec   `json:"spec"`
	Status            SupplyChainStatus `json:"status,omitempty"`
}

func (c *ClusterSupplyChain) validateNewState() error {
	names := make(map[string]bool)

	if err := c.validateParams(); err != nil {
		return err
	}

	for _, resource := range c.Spec.Resources {
		if _, ok := names[resource.Name]; ok {
			return fmt.Errorf(
				"duplicate resource name [%s] found in clustersupplychain [%s]",
				resource.Name,
				c.Name,
			)
		}
		names[resource.Name] = true
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

func (c *ClusterSupplyChain) ValidateCreate() error {
	return c.validateNewState()
}

func (c *ClusterSupplyChain) ValidateUpdate(_ runtime.Object) error {
	return c.validateNewState()
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
	Resources         []SupplyChainResource `json:"resources"`
	Selector          map[string]string     `json:"selector"`
	Params            []BlueprintParam      `json:"params,omitempty"`
	ServiceAccountRef ServiceAccountRef     `json:"serviceAccountRef,omitempty"`
}

type SupplyChainResource struct {
	Name        string                       `json:"name"`
	TemplateRef SupplyChainTemplateReference `json:"templateRef"`
	Params      []BlueprintParam             `json:"params,omitempty"`
	Sources     []ResourceReference          `json:"sources,omitempty"`
	Images      []ResourceReference          `json:"images,omitempty"`
	Configs     []ResourceReference          `json:"configs,omitempty"`
}

var ValidSupplyChainTemplates = []client.Object{
	&ClusterSourceTemplate{},
	&ClusterImageTemplate{},
	&ClusterConfigTemplate{},
	&ClusterTemplate{},
}

type SupplyChainTemplateReference struct {
	//+kubebuilder:validation:Enum=ClusterSourceTemplate;ClusterImageTemplate;ClusterTemplate;ClusterConfigTemplate
	Kind string `json:"kind"`
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
}

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
