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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	PipelineReady    = "Ready"
	RunTemplateReady = "RunTemplateReady"
)

const (
	ReadyRunTemplateReason                            = "Ready"
	NotFoundRunTemplateReason                         = "RunTemplateNotFound"
	StampedObjectRejectedByAPIServerRunTemplateReason = "StampedObjectRejectedByAPIServer"
	OutputPathNotSatisfiedRunTemplateReason           = "OutputPathNotSatisfied"
	TemplateStampFailureRunTemplateReason             = "TemplateStampFailure"
	FailedToListCreatedObjectsReason                  = "FailedToListCreatedObjects"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              PipelineSpec   `json:"spec"`
	Status            PipelineStatus `json:"status,omitempty"`
}

type PipelineStatus struct {
	ObservedGeneration int64                           `json:"observedGeneration,omitempty"`
	Conditions         []metav1.Condition              `json:"conditions,omitempty"`
	Outputs            map[string]apiextensionsv1.JSON `json:"outputs,omitempty"`
}

type PipelineSpec struct {
	// +kubebuilder:validation:Required
	RunTemplateRef TemplateReference               `json:"runTemplateRef"`
	Selector       *ResourceSelector               `json:"selector,omitempty"`
	Inputs         map[string]apiextensionsv1.JSON `json:"inputs,omitempty"`
}

type ResourceSelector struct {
	Resource       ResourceType      `json:"resource"`
	MatchingLabels map[string]string `json:"matchingLabels"`
}

type ResourceType struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`
}

type TemplateReference struct {
	Kind string `json:"kind,omitempty"`
	// +kubebuilder:validation:MinLength=1
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// +kubebuilder:object:root=true

type PipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pipeline `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&Pipeline{},
		&PipelineList{},
	)
}
