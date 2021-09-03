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
// +groupName=kontinue.io
// +kubebuilder:object:generate=true

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


const (
	PipelineReady = "Ready"
	RunTemplateReady = "RunTemplateReady"
)

const (
	ReadyRunTemplateReason    = "Ready"
	NotFoundRunTemplateReason = "RunTemplateNotFound"
)


// +kubebuilder:object:root=true

type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              PipelineSpec   `json:"spec"`
	Status            PipelineStatus `json:"status"`
}

type PipelineStatus struct {
	//ObservedGeneration int64                        `json:"observedGeneration,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

type PipelineSpec struct {
	// +kubebuilder:validation:Required
	RunTemplateName string `json:"runTemplateName"`
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
