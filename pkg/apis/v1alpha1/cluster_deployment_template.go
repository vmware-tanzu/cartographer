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
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

type ClusterDeploymentTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              DeploymentSpec `json:"spec"`
	Status            TemplateStatus `json:"status,omitempty"`
}

type DeploymentSpec struct {
	ObservedCompletion *ObservedCompletion `json:"observedCompletion,omitempty"`
	ObservedMatches    []ObservedMatch     `json:"observedMatches,omitempty"`
	TemplateSpec       `json:",inline"`
}

type ObservedMatch struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

type ObservedCompletion struct {
	SucceededCondition Condition  `json:"succeeded"`
	FailedCondition    *Condition `json:"failed,omitempty"`
}

type Condition struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// +kubebuilder:object:root=true

type ClusterDeploymentTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterDeploymentTemplate `json:"items"`
}

var _ webhook.Validator = &ClusterSourceTemplate{}

func (c *ClusterDeploymentTemplate) ValidateCreate() error {
	return c.validate()
}

func (c *ClusterDeploymentTemplate) ValidateUpdate(_ runtime.Object) error {
	return c.validate()
}

func (c *ClusterDeploymentTemplate) ValidateDelete() error {
	return nil
}

func (c *ClusterDeploymentTemplate) validate() error {
	err := c.Spec.TemplateSpec.validate()
	if err != nil {
		return err
	}

	if bothSet(c.Spec) || neitherSet(c.Spec) {
		return fmt.Errorf("invalid spec: must set exactly one of spec.ObservedMatches and spec.ObservedCompletion")
	}

	return nil
}

func bothSet(spec DeploymentSpec) bool {
	return spec.ObservedMatches != nil && spec.ObservedCompletion != nil
}

func neitherSet(spec DeploymentSpec) bool {
	return spec.ObservedMatches == nil && spec.ObservedCompletion == nil
}

func init() {
	SchemeBuilder.Register(
		&ClusterDeploymentTemplate{},
		&ClusterDeploymentTemplateList{},
	)
}
