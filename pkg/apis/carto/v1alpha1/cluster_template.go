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
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

type ClusterTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              TemplateSpec   `json:"spec"`
	Status            TemplateStatus `json:"status,omitempty"`
}

type TemplateSpec struct {
	// +kubebuilder:pruning:PreserveUnknownFields
	Template *runtime.RawExtension `json:"template,omitempty"`
	Ytt      string                `json:"ytt,omitempty"`
	Params   TemplateParams        `json:"params,omitempty"`
}

type TemplateStatus struct {
}

var _ webhook.Validator = &ClusterTemplate{}

func (c *ClusterTemplate) ValidateCreate() error {
	return c.Spec.validate()
}

func (c *ClusterTemplate) ValidateUpdate(_ runtime.Object) error {
	return c.Spec.validate()
}

func (c *ClusterTemplate) ValidateDelete() error {
	return nil
}

func (t *TemplateSpec) validate() error {
	if t.Template == nil && t.Ytt == "" {
		return fmt.Errorf("invalid template: must specify one of template or ytt, found neither")
	}
	if t.Template != nil && t.Ytt != "" {
		return fmt.Errorf("invalid template: must specify one of template or ytt, found both")
	}
	if t.Template != nil {
		obj := metav1.PartialObjectMetadata{}
		if err := json.Unmarshal(t.Template.Raw, &obj); err != nil {
			return fmt.Errorf("invalid template: failed to parse object metadata: %w", err)
		}
		if obj.Namespace != metav1.NamespaceNone {
			return errors.New("invalid template: template should not set metadata.namespace on the child object")
		}
	}
	return nil
}

// +kubebuilder:object:root=true

type ClusterTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&ClusterTemplate{},
		&ClusterTemplateList{},
	)
}
