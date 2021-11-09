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
	DeliveryReady          = "Ready"
	DeliveryTemplatesReady = "TemplatesReady"
)

const (
	ReadyDeliveryTemplatesReadyReason    = "Ready"
	NotFoundDeliveryTemplatesReadyReason = "TemplatesNotFound"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

type ClusterDelivery struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ClusterDeliverySpec   `json:"spec"`
	Status            ClusterDeliveryStatus `json:"status,omitempty"`
}

type ClusterDeliverySpec struct {
	Resources []ClusterDeliveryResource `json:"resources"`
	Selector  map[string]string         `json:"selector"`
}

type ClusterDeliveryStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}

type ClusterDeliveryResource struct {
	Name        string                           `json:"name"`
	TemplateRef DeliveryClusterTemplateReference `json:"templateRef"`
	Params      []Param                          `json:"params,omitempty"`
	Sources     []ResourceReference              `json:"sources,omitempty"`
	Deployment  *DeploymentReference             `json:"deployment,omitempty"`
	Configs     []ResourceReference              `json:"configs,omitempty"`
}

type DeploymentReference struct {
	Resource string `json:"resource"`
}

var ValidDeliveryTemplates = []client.Object{
	&ClusterSourceTemplate{},
	&ClusterDeploymentTemplate{},
	&ClusterTemplate{},
}

type DeliveryClusterTemplateReference struct {
	// +kubebuilder:validation:Enum=ClusterSourceTemplate;ClusterDeploymentTemplate;ClusterTemplate
	Kind string `json:"kind"`
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
}

// +kubebuilder:object:root=true

type ClusterDeliveryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterDelivery `json:"items"`
}

func (c *ClusterDelivery) ValidateCreate() error {
	return c.validateNewState()
}

func (c *ClusterDelivery) ValidateUpdate(_ runtime.Object) error {
	return c.validateNewState()
}

func (c *ClusterDelivery) ValidateDelete() error {
	return nil
}

func (c *ClusterDelivery) validateNewState() error {
	if err := c.validateResourceNamesUnique(); err != nil {
		return err
	}

	if err := c.validateDeploymentPassedToProperReceivers(); err != nil {
		return err
	}

	return c.validateDeploymentTemplateDidNotReceiveConfig()
}

func (c *ClusterDelivery) validateDeploymentPassedToProperReceivers() error {
	for _, resource := range c.Spec.Resources {
		if resource.TemplateRef.Kind == "ClusterDeploymentTemplate" && resource.Deployment == nil {
			return fmt.Errorf("spec.resources['%s'] is a ClusterDeploymentTemplate and must receive a deployment", resource.Name)
		}

		if resource.Deployment != nil && resource.TemplateRef.Kind != "ClusterDeploymentTemplate" {
			return fmt.Errorf("spec.resources['%s'] receives a deployment but is not a ClusterDeploymentTemplate", resource.Name)
		}
	}
	return nil
}

func (c *ClusterDelivery) validateResourceNamesUnique() error {
	names := map[string]bool{}

	for idx, resource := range c.Spec.Resources {
		if names[resource.Name] {
			return fmt.Errorf("spec.resources[%d].name \"%s\" cannot appear twice", idx, resource.Name)
		}
		names[resource.Name] = true
	}
	return nil
}

func (c *ClusterDelivery) validateDeploymentTemplateDidNotReceiveConfig() error {
	for _, resource := range c.Spec.Resources {
		if resource.TemplateRef.Kind == "ClusterDeploymentTemplate" && resource.Configs != nil {
			return fmt.Errorf("spec.resources['%s'] is a ClusterDeploymentTemplate and must not receive config", resource.Name)
		}
	}
	return nil
}

func init() {
	SchemeBuilder.Register(
		&ClusterDelivery{},
		&ClusterDeliveryList{},
	)
}
