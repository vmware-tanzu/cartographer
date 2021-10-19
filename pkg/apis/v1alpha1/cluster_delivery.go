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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	Configs     []ResourceReference              `json:"configs,omitempty"`
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

func init() {
	SchemeBuilder.Register(
		&ClusterDelivery{},
		&ClusterDeliveryList{},
	)
}
