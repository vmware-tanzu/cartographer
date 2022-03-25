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
// +kubebuilder:resource:path=clusterdeploymenttemplates,scope=Cluster,shortName=cdt

type ClusterDeploymentTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	// Spec describes the deployment template.
	// More info: https://cartographer.sh/docs/latest/reference/template/#clusterdeploymenttemplate
	Spec DeploymentSpec `json:"spec"`
}

type DeploymentSpec struct {
	TemplateSpec `json:",inline"`

	// ObservedMatches describe the criteria for determining that the templated object
	// completed configuration of environment.
	// These criteria assert completion when an output (usually a field in .status)
	// matches an input (usually a field in .spec)
	// Cannot specify both ObservedMatches and ObservedCompletion.
	ObservedMatches []ObservedMatch `json:"observedMatches,omitempty"`

	// ObservedCompletion describe the criteria for determining that the templated object
	// completed configuration of environment.
	// These criteria assert completion when metadata.Generation and status.ObservedGeneration
	// match, AND success or failure criteria match.
	// Cannot specify both ObservedMatches and ObservedCompletion.
	ObservedCompletion *ObservedCompletion `json:"observedCompletion,omitempty"`
}

type ObservedMatch struct {
	// Input is a jsonPath to a value that is fulfilled before the templated object is reconciled.
	// Usually a value in the .spec of the object
	Input string `json:"input"`
	// Output is a jsonPath to a value that is fulfilled after the templated object is reconciled.
	// Usually a value in the .status of the object
	Output string `json:"output"`
}

type ObservedCompletion struct {
	// SucceededCondition, when matched, indicates that the input was successfully deployed.
	SucceededCondition Condition `json:"succeeded"`

	// FailedCondition, when matched, indicates that the input did not deploy successfully.
	FailedCondition *Condition `json:"failed,omitempty"`
}

type Condition struct {
	// Key is a jsonPath expression pointing to the field to inspect on the templated
	// object, eg: 'status.conditions[?(@.type=="Succeeded")].status'
	Key string `json:"key"`

	// Value is the expected value that, when matching the key's actual value,
	// makes this condition true.
	Value string `json:"value"`
}

// +kubebuilder:object:root=true

type ClusterDeploymentTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterDeploymentTemplate `json:"items"`
}

// +kubebuilder:webhook:path=/validate-carto-run-v1alpha1-clusterdeploymenttemplate,mutating=false,failurePolicy=fail,sideEffects=none,admissionReviewVersions=v1beta1;v1,groups=carto.run,resources=clusterdeploymenttemplates,verbs=create;update,versions=v1alpha1,name=deployment-template-validator.cartographer.com

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

	if c.bothConditionsSet() || c.neitherConditionSet() {
		return fmt.Errorf("invalid spec: must set exactly one of spec.ObservedMatches and spec.ObservedCompletion")
	}

	return nil
}

func (c *ClusterDeploymentTemplate) bothConditionsSet() bool {
	return c.Spec.ObservedMatches != nil && c.Spec.ObservedCompletion != nil
}

func (c *ClusterDeploymentTemplate) neitherConditionSet() bool {
	return c.Spec.ObservedMatches == nil && c.Spec.ObservedCompletion == nil
}

func init() {
	SchemeBuilder.Register(
		&ClusterDeploymentTemplate{},
		&ClusterDeploymentTemplateList{},
	)
}
