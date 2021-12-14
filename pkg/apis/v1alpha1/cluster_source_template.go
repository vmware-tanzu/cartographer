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

type ClusterSourceTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SourceTemplateSpec `json:"spec"`
}

type SourceTemplateSpec struct {
	TemplateSpec `json:",inline"`
	URLPath      string `json:"urlPath"`
	RevisionPath string `json:"revisionPath"`
}

var _ webhook.Validator = &ClusterSourceTemplate{}

func (c *ClusterSourceTemplate) ValidateCreate() error {
	return c.Spec.TemplateSpec.validate()
}

func (c *ClusterSourceTemplate) ValidateUpdate(_ runtime.Object) error {
	return c.Spec.TemplateSpec.validate()
}

func (c *ClusterSourceTemplate) ValidateDelete() error {
	return nil
}

// +kubebuilder:object:root=true

type ClusterSourceTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterSourceTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&ClusterSourceTemplate{},
		&ClusterSourceTemplateList{},
	)
}
