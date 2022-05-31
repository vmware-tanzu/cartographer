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

// ClusterRunTemplate defines how to build a runnable's stamped object
// See https://cartographer.sh/docs/latest/runnable/architecture/#clusterruntemplate
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=clusterruntemplates, scope=Cluster, shortName=crt
type ClusterRunTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	// Spec describes the run template.
	// More info: https://cartographer.sh/docs/latest/reference/runnable/#clusterruntemplate
	Spec RunTemplateSpec `json:"spec"`
}

type RunTemplateSpec struct {
	// Template defines a resource template for a Kubernetes Resource or
	// Custom Resource which is applied to the server each time
	// the blueprint is applied. Templates support simple value
	// interpolation using the $()$ marker format. For more
	// information, see: https://cartographer.sh/docs/latest/templating/
	// You should not define the namespace for the resource - it will automatically
	// be created in the owner namespace. If the namespace is specified and is not
	// the owner namespace, the resource will fail to be created.
	// +kubebuilder:pruning:PreserveUnknownFields
	Template runtime.RawExtension `json:"template"`

	// Outputs are a named list of jsonPaths that are used to gather results
	// from the last successful object stamped by the template.
	// E.g: 	my-output: .status.results[?(@.name=="IMAGE-DIGEST")].value
	// Note: outputs are only filled on the runnable when the templated object
	// has a Succeeded condition with a Status of True
	// E.g:     status.conditions[?(@.type=="Succeeded")].status == True
	// a runnable creating an object without a Succeeded condition (like a Job or ConfigMap)
	// will never display an output
	// +optional
	Outputs map[string]string `json:"outputs,omitempty"`
}

// +kubebuilder:object:root=true

type ClusterRunTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterRunTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&ClusterRunTemplate{},
		&ClusterRunTemplateList{},
	)
}
