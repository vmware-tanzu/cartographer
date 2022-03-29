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
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=clustertemplates,scope=Cluster,shortName=ct

type ClusterTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	// Spec describes the template.
	// More info: https://cartographer.sh/docs/latest/reference/template/#clustertemplate
	Spec TemplateSpec `json:"spec"`
}

type TemplateSpec struct {
	// Template defines a resource template for a Kubernetes Resource or
	// Custom Resource which is applied to the server each time
	// the blueprint is applied. Templates support simple value
	// interpolation using the $()$ marker format. For more
	// information, see: https://cartographer.sh/docs/latest/templating/
	// You cannot define both Template and Ytt at the same time.
	// You should not define the namespace for the resource - it will automatically
	// be created in the owner namespace. If the namespace is specified and is not
	// the owner namespace, the resource will fail to be created.
	// +kubebuilder:pruning:PreserveUnknownFields
	Template *runtime.RawExtension `json:"template,omitempty"`

	// Ytt defines a resource template written in `ytt` for a Kubernetes Resource or
	// Custom Resource which is applied to the server each time
	// the blueprint is applied. Templates support simple value
	// interpolation using the $()$ marker format. For more
	// information, see: https://cartographer.sh/docs/latest/templating/
	// You cannot define both Template and Ytt at the same time.
	// You should not define the namespace for the resource - it will automatically
	// be created in the owner namespace. If the namespace is specified and is not
	// the owner namespace, the resource will fail to be created.
	Ytt string `json:"ytt,omitempty"`

	// Additional parameters.
	// See: https://cartographer.sh/docs/latest/architecture/#parameter-hierarchy
	// +optional
	Params TemplateParams `json:"params,omitempty"`
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
