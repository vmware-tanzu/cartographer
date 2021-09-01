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

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

	for _, component := range c.Spec.Components {
		if _, ok := names[component.Name]; ok {
			return fmt.Errorf(
				"duplicate component name '%s' found in clustersupplychain '%s'",
				component.Name,
				c.Name,
			)
		}
		names[component.Name] = true
	}

	for _, component := range c.Spec.Components {
		err := c.validateComponentRef(component.Sources, component, "ClusterSourceTemplate", "Source")
		if err != nil {
			return err
		}

		err = c.validateComponentRef(component.Images, component, "ClusterImageTemplate", "Image")
		if err != nil {
			return err
		}

		err = c.validateComponentRef(component.Configs, component, "ClusterConfigTemplate", "Config")
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *ClusterSupplyChain) validateComponentRef(references []ComponentReference, component SupplyChainComponent, targetKind string, sourceField string) error {
	for _, ref := range references {
		referencedComponent := c.getComponentByName(ref.Component)
		if referencedComponent == nil {
			continue
		}
		if referencedComponent.TemplateRef.Kind != targetKind {
			return fmt.Errorf(
				"%s '%s' for '%s' component must reference a %s",
				sourceField,
				referencedComponent.Name,
				component.Name,
				targetKind,
			)
		}
	}
	return nil
}

func (c *ClusterSupplyChain) getComponentByName(name string) *SupplyChainComponent {
	for _, component := range c.Spec.Components {
		if component.Name == name {
			return &component
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

type SupplyChainSpec struct {
	Components []SupplyChainComponent `json:"components"`
	Selector   map[string]string      `json:"selector"`
}

type SupplyChainParam struct {
	Name  string               `json:"name"`
	Value apiextensionsv1.JSON `json:"value"`
}

type SupplyChainComponent struct {
	Name        string               `json:"name"`
	TemplateRef TemplateReference    `json:"templateRef"`
	Params      []SupplyChainParam   `json:"params,omitempty"`
	Sources     []ComponentReference `json:"sources,omitempty"`
	Images      []ComponentReference `json:"images,omitempty"`
	Configs     []ComponentReference `json:"configs,omitempty"`
}

type TemplateReference struct {
	// +kubebuilder:validation:Enum=ClusterSourceTemplate;ClusterImageTemplate;ClusterTemplate;ClusterConfigTemplate
	Kind string `json:"kind"`
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
}

type ComponentReference struct {
	Name      string `json:"name"`
	Component string `json:"component"`
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
