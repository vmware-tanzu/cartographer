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
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Succeeded",type=string,JSONPath=`.status.conditions[?(@.type=="Succeeded")].status`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   PipelineSpec   `json:"spec"`
	Status PipelineStatus `json:"status,omitempty"`
}

type PipelineSpec struct {
	// Inputs define the parameters that should be passed onto the run
	// template. The contents of this field are used to provide idempotency
	// for the reconciler, being used to define a digest that's compared
	// before creating a new run.
	//
	// +kubebuilder:pruning:PreserveUnknownFields
	Inputs runtime.RawExtension `json:"inputs"`

	// RunTemplateName is the name of the RunTemplate object to use to
	// dictate how to run an instance of a pipeline.
	//
	RunTemplateName string `json:"runTemplateName"`

	// Selector is used for matching against an object to be accessible for
	// templating a Run (for instance, finding out the name of a
	// developer-provided tekton/Pipeline, or a ConfigMap to be embedded
	// in).
	//
	Selector ObjectSelector `json:"selector,omitempty"`
}

type ObjectSelector struct {
	// Resource indicates a GVR to match against.
	//
	Resource ObjectSelectorResource `json:"resource"`

	// MatchingLabels indicates the label set to use when searching for
	// objects of a given gvr.
	//
	// +kubebuilder:validation:MinProperties=1
	MatchingLabels map[string]string `json:"matchingLabels"`
}

type ObjectSelectorResource struct {
	// APIVersion
	//
	APIVersion string `json:"apiVersion"`

	// Kind
	//
	Kind string `json:"kind"`
}

type PipelineStatus struct {
	// Conditions
	//
	Conditions []metav1.Condition `json:"conditions"`

	// LatestInputs represents the set of inputs that were used the last
	// time that this reconciler observed a success for a run that it
	// created.
	//
	// +kubebuilder:pruning:PreserveUnknownFields
	//
	LatestInputs runtime.RawExtension `json:"latestInputs,omitempty"`

	// LatestOutputs displays the set of outputs that were produced in the
	// last time that this reconciler observed a success for a run that it
	// created.
	//
	// +kubebuilder:pruning:PreserveUnknownFields
	//
	LatestOutputs runtime.RawExtension `json:"latestOutputs,omitempty"`
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
