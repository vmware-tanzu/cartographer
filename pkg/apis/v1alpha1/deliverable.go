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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DeliverableReady              = "Ready"
	DeliverableDeliveryReady      = "DeliveryReady"
	DeliverableResourcesSubmitted = "ResourcesSubmitted"
)

const (
	ReadyDeliveryReason                    = "Ready"
	DeliverableLabelsMissingDeliveryReason = "DeliverableLabelsMissing"
	NotFoundDeliveryReadyReason            = "DeliveryNotFound"
	MultipleMatchesDeliveryReadyReason     = "MultipleDeliveryMatches"
)

const (
	CompleteResourcesSubmittedReason                       = "ResourceSubmissionComplete"
	TemplateObjectRetrievalFailureResourcesSubmittedReason = "TemplateObjectRetrievalFailure"
	MissingValueAtPathResourcesSubmittedReason             = "MissingValueAtPath"
	TemplateStampFailureResourcesSubmittedReason           = "TemplateStampFailure"
	TemplateRejectedByAPIServerResourcesSubmittedReason    = "TemplateRejectedByAPIServer"
	UnknownErrorResourcesSubmittedReason                   = "UnknownError"
	DeploymentConditionNotMetResourcesSubmittedReason      = "ConditionNotMet"
	DeploymentFailedConditionMetResourcesSubmittedReason   = "FailedConditionMet"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=all
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Source",type="string",JSONPath=`.spec.source['git.url','image']`
// +kubebuilder:printcolumn:name="Delivery",type="string",JSONPath=".status.deliveryRef.name"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=='Ready')].status`
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=`.status.conditions[?(@.type=='Ready')].reason`

type Deliverable struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	// Spec describes the deliverable.
	// More info: https://cartographer.sh/docs/latest/reference/workload/#deliverable
	Spec DeliverableSpec `json:"spec"`

	// Status conforms to the Kubernetes conventions:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
	Status DeliverableStatus `json:"status,omitempty"`
}

type DeliverableSpec struct {
	// Additional parameters.
	// +optional
	Params []OwnerParam `json:"params,omitempty"`

	// The location of the source configuration for the deliverable. Specify
	// one of `spec.source` or `spec.image`
	// +optional
	Source *Source `json:"source,omitempty"`

	// ServiceAccountName refers to the Service account with permissions to create resources
	// submitted by the supply chain.
	//
	// If not set, Cartographer will use serviceAccountName from supply chain.
	//
	// If that is also not set, Cartographer will use the default service account in the
	// workload's namespace.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}

type DeliverableStatus struct {
	OwnerStatus `json:",inline"`

	// DeliveryRef is the Delivery resource that was used when this status was set.
	DeliveryRef ObjectReference `json:"deliveryRef,omitempty"`
}

// +kubebuilder:object:root=true

type DeliverableList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Deliverable `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&Deliverable{},
		&DeliverableList{},
	)
}
