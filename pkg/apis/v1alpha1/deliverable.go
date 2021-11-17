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
	NotReadyDeliveryReason                 = "DeliveryNotReady"
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
// +kubebuilder:subresource:status

type Deliverable struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              DeliverableSpec   `json:"spec"`
	Status            DeliverableStatus `json:"status,omitempty"`
}

type DeliverableSpec struct {
	Params []Param `json:"params,omitempty"`
	Source *Source `json:"source,omitempty"`
}

type DeliverableStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
	DeliveryRef        ObjectReference    `json:"deliveryRef,omitempty"`
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
