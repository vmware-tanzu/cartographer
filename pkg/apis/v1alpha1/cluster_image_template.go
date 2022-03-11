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
// +kubebuilder:resource:path=clusterimagetemplates,scope=Cluster,shortName=cit

type ClusterImageTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	// Spec describes the image template.
	// More info: https://cartographer.sh/docs/latest/reference/template/#clusterimagetemplate
	Spec ImageTemplateSpec `json:"spec"`
}
type ImageTemplateSpec struct {
	TemplateSpec `json:",inline"`

	// ImagePath is a path into the templated object's
	// data that contains a valid image digest. This
	// might be a URL or in some cases just a repository path and digest.
	// The final spec for this field may change as we implement
	// RFC-0016 https://github.com/vmware-tanzu/cartographer/blob/main/rfc/rfc-0016-validate-template-outputs.md
	// ImagePath is specified in jsonpath format, eg: .status.artifact.image_digest
	ImagePath string `json:"imagePath"`
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
