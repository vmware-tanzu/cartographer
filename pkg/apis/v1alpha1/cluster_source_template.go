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
// +kubebuilder:resource:path=clustersourcetemplates,scope=Cluster,shortName=cst

type ClusterSourceTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	// Spec describes the source template.
	// More info: https://cartographer.sh/docs/latest/reference/template/#clustersourcetemplate
	Spec SourceTemplateSpec `json:"spec"`
}

type SourceTemplateSpec struct {
	TemplateSpec `json:",inline"`

	// URLPath is a path into the templated object's
	// data that contains a URL. The URL, along with the revision,
	// represents the output of the Template.
	// URLPath is specified in jsonpath format, eg: .status.artifact.url
	URLPath string `json:"urlPath"`

	// RevisionPath is a path into the templated object's
	// data that contains a revision. The revision, along with the URL,
	// represents the output of the Template.
	// RevisionPath is specified in jsonpath format, eg: .status.artifact.revision
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
