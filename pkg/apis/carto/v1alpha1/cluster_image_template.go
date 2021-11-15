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

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

type ClusterImageTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ImageTemplateSpec   `json:"spec"`
	Status            ImageTemplateStatus `json:"status,omitempty"`
}
type ImageTemplateSpec struct {
	TemplateSpec `json:",inline"`
	ImagePath    string `json:"imagePath"`
}

type ImageTemplateStatus struct {
}

var _ webhook.Validator = &ClusterImageTemplate{}

func (c *ClusterImageTemplate) ValidateCreate() error {
	return c.Spec.TemplateSpec.validate()
}

func (c *ClusterImageTemplate) ValidateUpdate(_ runtime.Object) error {
	return c.Spec.TemplateSpec.validate()
}

func (c *ClusterImageTemplate) ValidateDelete() error {
	return nil
}

// +kubebuilder:object:root=true

type ClusterImageTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterImageTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&ClusterImageTemplate{},
		&ClusterImageTemplateList{},
	)
}
