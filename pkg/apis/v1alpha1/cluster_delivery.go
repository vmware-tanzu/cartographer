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

	// Spec describes the delivery.
	// More info: https://cartographer.sh/docs/latest/reference/deliverable/#clusterdelivery
	Spec DeliverySpec `json:"spec"`

	// Status conforms to the Kubernetes conventions:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
	Status DeliveryStatus `json:"status,omitempty"`
}

func (c *ClusterDelivery) GetSelector() map[string]string {
	return c.Spec.Selector
}

type DeliverySpec struct {
	// Resources that are responsible for deploying and validating
	// the deliverable
	Resources []DeliveryResource `json:"resources"`

	// Specifies the label key-value pairs used to select deliverables
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

type DeliveryStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}

type DeliveryResource struct {
	// Name of the resource. Used as a reference for inputs, as well as being
	// the name presented in deliverable statuses to identify this resource.
	Name string `json:"name"`

	// TemplateRef identifies the template used to produce this resource
	TemplateRef DeliveryTemplateReference `json:"templateRef"`

	// Params are a list of parameters to provide to the template in TemplateRef
	// Template params do not have to be specified here, unless you want to
	// force a particular value, or add a default value.
	//
	// Parameters are consumed in a template with the syntax:
	//   $(params.<name>)$
	Params []BlueprintParam `json:"params,omitempty"`

	// Sources is a list of references to other 'source' resources in this list.
	// A source resource has the kind ClusterSourceTemplate or ClusterDeploymentTemplate
	//
	// In a template, sources can be consumed as:
	//    $(sources.<name>.url)$ and $(sources.<name>.revision)$
	//
	// If there is only one source, it can be consumed as:
	//    $(source.url)$ and $(source.revision)$
	Sources []ResourceReference `json:"sources,omitempty"`

	// Deployment is a reference to a 'deployment' resource.
	// A deployment resource has the kind ClusterDeploymentTemplate
	//
	// In a template, the deployment can be consumed as:
	//    $(deployment.url)$ and $(deployment.revision)$
	Deployment *DeploymentReference `json:"deployment,omitempty"`

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

type DeploymentReference struct {
	Resource string `json:"resource"`
}

var ValidDeliveryTemplates = []client.Object{
	&ClusterSourceTemplate{},
	&ClusterDeploymentTemplate{},
	&ClusterTemplate{},
}

type DeliveryTemplateReference struct {
	// Kind of the template to apply
	// +kubebuilder:validation:Enum=ClusterSourceTemplate;ClusterDeploymentTemplate;ClusterTemplate
	Kind string `json:"kind"`
	// Name of the template to apply
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name,omitempty"`

	// Options is a list of template names and Selectors. The templates must all be of type Kind.
	// A template will be selected if the workload matches the specified Selector.
	// Only one template can be selected.
	// Only one of Name and Options can be specified.
	// +kubebuilder:validation:MinItems=2
	Options []TemplateOption `json:"options,omitempty"`
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
	if err := c.validateParams(); err != nil {
		return err
	}

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
			return fmt.Errorf("spec.resources[%s] is a ClusterDeploymentTemplate and must receive a deployment", resource.Name)
		}

		if resource.Deployment != nil && resource.TemplateRef.Kind != "ClusterDeploymentTemplate" {
			return fmt.Errorf("spec.resources[%s] receives a deployment but is not a ClusterDeploymentTemplate", resource.Name)
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
			return fmt.Errorf("spec.resources[%s] is a ClusterDeploymentTemplate and must not receive config", resource.Name)
		}
	}
	return nil
}

func (c *ClusterDelivery) validateParams() error {
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

func init() {
	SchemeBuilder.Register(
		&ClusterDelivery{},
		&ClusterDeliveryList{},
	)
}
