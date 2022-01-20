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
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

type ClusterConfigTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	// Spec describes the config template.
	// More info: https://cartographer.sh/docs/latest/reference/template/#clusterconfigtemplate
	Spec ConfigTemplateSpec `json:"spec"`
}

type ConfigTemplateSpec struct {
	TemplateSpec `json:",inline"`

	// ConfigPath is a path into the templated object's
	// data that contains valid yaml. This
	// is typically the information that will configure the
	// components of the deployable image.
	// ConfigPath is specified in jsonpath format, eg: .data
	ConfigPath string `json:"configPath"`
}

var _ webhook.Validator = &ClusterConfigTemplate{}

func (c *ClusterConfigTemplate) ValidateCreate() error {
	return c.Spec.TemplateSpec.validate()
}

func (c *ClusterConfigTemplate) ValidateUpdate(_ runtime.Object) error {
	return c.Spec.TemplateSpec.validate()
}

func (c *ClusterConfigTemplate) ValidateDelete() error {
	return nil
}

// +kubebuilder:object:root=true

type ClusterConfigTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterConfigTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&ClusterConfigTemplate{},
		&ClusterConfigTemplateList{},
	)
}
