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
//

type RunTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec RunTemplateSpec `json:"spec"`
}

type RunTemplateSpec struct {
	// Completion describe rules for determining whether the invocation was
	// successfull, failed, or is still running (neither succeeded nor
	// failed).
	//
	Completion Completion `json:"completion,omitempty"`

	// Outputs lets you specify rules for capturing results from the
	// pipeline invocations.
	//
	// +kubebuilder:pruning:PreserveUnknownFields
	Outputs []RunTemplateOutput `json:"outputs,omitempty"`

	// Template is the template that should be stamped out when submitting
	// a new pipeline invocation.
	//
	// To this template are made available:
	//
	// 	- the pipeline object that is referencing this RunTemplate
	// 	- (optional) the object that matched the Pipeline's selection rules.
	//
	// For instance:
	//
	// 	kind: ConfigMap
	// 	apiVersion: v1
	// 	metadata:
	// 	  generateName: $(pipeline.metadata.name)-
	// 	data:
	// 	  foo: $(selected.metadata.name)
	//
	// +kubebuilder:pruning:PreserveUnknownFields
	Template runtime.RawExtension `json:"template"`
}

type RunTemplateOutput struct {
	// Name indicates the value that should be used as a key in the
	// `carto.run/Pipeline`'s .status.latestOutputs.
	//
	Name string `json:"name"`

	// Path denotes the query that should be performed to retrieve a value
	// from the object stamped out according to this RunTemplate template.
	//
	Path string `json:"path"`
}

// +kubebuilder:object:root=true

type RunTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RunTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&RunTemplate{},
		&RunTemplateList{},
	)
}

type Completion struct {
	// Succeeded provides the rule to evaluate to determine if the
	// invocation succeeded or failed.
	//
	Succeeded CompletionEvaluation `json:"succeeded,omitempty"`

	// Failed provides the rule to evaluate to determine if the invocation
	// succeeded or failed.
	//
	Failed CompletionEvaluation `json:"failed,omitempty"`
}

type CompletionEvaluation struct {
	// Key is the gjson query the perform to retrieve an indication of
	// success/failure.
	//
	Key string `json:"key,omitempty"`

	// Value is the expected result to indicate success/failure.
	//
	Value string `json:"value,omitempty"`
}
