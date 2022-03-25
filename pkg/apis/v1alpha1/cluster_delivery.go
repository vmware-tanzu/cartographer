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
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const (
	DeliveryReady          = "Ready"
	DeliveryTemplatesReady = "TemplatesReady"
)

const (
	ReadyDeliveryTemplatesReadyReason    = "Ready"
	NotFoundDeliveryTemplatesReadyReason = "TemplatesNotFound"
)

var ValidDeliveryTemplates = []client.Object{
	&ClusterSourceTemplate{},
	&ClusterDeploymentTemplate{},
	&ClusterTemplate{},
	&ClusterConfigTemplate{},
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=clusterdeliveries,scope=Cluster,shortName=cd
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=='Ready')].status`
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=`.status.conditions[?(@.type=='Ready')].reason`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`

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

type DeliverySpec struct {
	LegacySelector `json:",inline"`

	// Resources that are responsible for deploying and validating
	// the deliverable
	Resources []DeliveryResource `json:"resources"`

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

type DeliveryTemplateReference struct {
	// Kind of the template to apply
	// +kubebuilder:validation:Enum=ClusterSourceTemplate;ClusterDeploymentTemplate;ClusterTemplate;ClusterConfigTemplate
	Kind string `json:"kind"`
	// Name of the template to apply
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name,omitempty"`

	// Options is a list of template names and Selector. The templates must all be of type Kind.
	// A template will be selected if the deliverable matches the specified selector.
	// Only one template can be selected.
	// Only one of Name and Options can be specified.
	// +kubebuilder:validation:MinItems=2
	Options []TemplateOption `json:"options,omitempty"`
}

type DeploymentReference struct {
	Resource string `json:"resource"`
}

// +kubebuilder:object:root=true

type ClusterDeliveryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterDelivery `json:"items"`
}

// +kubebuilder:webhook:path=/validate-carto-run-v1alpha1-clusterdelivery,mutating=false,failurePolicy=fail,sideEffects=none,admissionReviewVersions=v1beta1;v1,groups=carto.run,resources=clusterdeliveries,verbs=create;update,versions=v1alpha1,name=delivery-validator.cartographer.com

var _ webhook.Validator = &ClusterDelivery{}

func (c *ClusterDelivery) ValidateCreate() error {
	err := c.validateNewState()
	if err != nil {
		return fmt.Errorf("error validating clusterdelivery [%s]: %w", c.Name, err)
	}
	return nil
}

func (c *ClusterDelivery) ValidateUpdate(_ runtime.Object) error {
	err := c.validateNewState()
	if err != nil {
		return fmt.Errorf("error validating clusterdelivery [%s]: %w", c.Name, err)
	}
	return nil
}

func (c *ClusterDelivery) ValidateDelete() error {
	return nil
}

func (c *ClusterDelivery) GetSelectors() LegacySelector {
	return c.Spec.LegacySelector
}

func init() {
	SchemeBuilder.Register(
		&ClusterDelivery{},
		&ClusterDeliveryList{},
	)
}
