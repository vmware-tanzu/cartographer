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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	WorkloadReady             = "Ready"
	WorkloadSupplyChainReady  = "SupplyChainReady"
	WorkloadResourceSubmitted = "ResourcesSubmitted"
)

const (
	ReadySupplyChainReason                               = "Ready"
	WorkloadLabelsMissingSupplyChainReason               = "WorkloadLabelsMissing"
	NotFoundSupplyChainReadyReason                       = "SupplyChainNotFound"
	MultipleMatchesSupplyChainReadyReason                = "MultipleSupplyChainMatches"
	ServiceAccountSecretErrorResourcesSubmittedReason    = "ServiceAccountSecretError"
	ResourceRealizerBuilderErrorResourcesSubmittedReason = "ResourceRealizerBuilderError"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type Workload struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              WorkloadSpec   `json:"spec"`
	Status            WorkloadStatus `json:"status,omitempty"`
}

type WorkloadServiceClaim struct {
	Name string                         `json:"name"`
	Ref  *WorkloadServiceClaimReference `json:"ref,omitempty"`
}

type WorkloadServiceClaimReference struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}

type WorkloadSpec struct {
	Params []Param         `json:"params,omitempty"`
	Source *Source         `json:"source,omitempty"`
	Build  WorkloadBuild   `json:"build,omitempty"`
	Env    []corev1.EnvVar `json:"env,omitempty"`
	// Image is a pre-built image in a registry. It is an alternative to defining source
	// code.
	Image              *string                      `json:"image,omitempty"`
	Resources          *corev1.ResourceRequirements `json:"resources,omitempty"`
	ServiceAccountName string                       `json:"serviceAccountName,omitempty"`
	ServiceClaims      []WorkloadServiceClaim       `json:"serviceClaims,omitempty"`
}

type WorkloadBuild struct {
	Env []corev1.EnvVar `json:"env,omitempty"`
}

type WorkloadStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
	SupplyChainRef     ObjectReference    `json:"supplyChainRef,omitempty"`
}

// +kubebuilder:object:root=true

type WorkloadList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workload `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&Workload{},
		&WorkloadList{},
	)
}
