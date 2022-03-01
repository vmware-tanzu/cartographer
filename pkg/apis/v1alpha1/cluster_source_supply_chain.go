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
)

//var ValidSupplyChainTemplates = []client.Object{
//	&ClusterSourceTemplate{},
//	&ClusterImageTemplate{},
//	&ClusterConfigTemplate{},
//	&ClusterTemplate{},
//}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=='Ready')].status`
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=`.status.conditions[?(@.type=='Ready')].reason`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`

type ClusterSourceSupplyChain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	// Spec describes the suppply chain.
	// More info: https://cartographer.sh/docs/latest/reference/workload/#clustersupplychain
	Spec SourceSupplyChainSpec `json:"spec"`

	// Status conforms to the Kubernetes conventions:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
	Status SupplyChainStatus `json:"status,omitempty"`
}

type SourceSupplyChainSpec struct {
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

	// Resources that are responsible for bringing the application to a
	// deliverable state.
	Resources []SupplyChainResource `json:"resources"`

	// Specifies the label key-value pairs used to select workloads
	// See: https://cartographer.sh/docs/v0.1.0/architecture/#selectors
	Selector map[string]string `json:"selector,omitempty"`

	// Additional parameters.
	// See: https://cartographer.sh/docs/latest/architecture/#parameter-hierarchy
	// +optional
	Params []BlueprintParam `json:"params,omitempty"`

	// ServiceAccountName refers to the Service account with permissions to create resources
	// submitted by the supply chain.
	//
	// If not set, Cartographer will use serviceAccountName from supply chain.
	//
	// If that is also not set, Cartographer will use the default service account in the
	// workload's namespace.
	// +optional
	ServiceAccountRef ServiceAccountRef `json:"serviceAccountRef,omitempty"`
}

// +kubebuilder:object:root=true

type ClusterSourceSupplyChainList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterSourceSupplyChain `json:"items"`
}

//func (c *ClusterSourceSupplyChain) ValidateCreate() error {
//	err := c.validateNewState()
//	if err != nil {
//		return fmt.Errorf("error validating clustersupplychain [%s]: %w", c.Name, err)
//	}
//	return nil
//}
//
//func (c *ClusterSourceSupplyChain) ValidateUpdate(_ runtime.Object) error {
//	err := c.validateNewState()
//	if err != nil {
//		return fmt.Errorf("error validating clustersupplychain [%s]: %w", c.Name, err)
//	}
//	return nil
//}
//
//func (c *ClusterSourceSupplyChain) ValidateDelete() error {
//	return nil
//}

func (c *ClusterSourceSupplyChain) GetSelector() map[string]string {
	return c.Spec.Selector
}

//func GetSelectorsFromObject(o client.Object) []string {
//	var res []string
//	res = []string{}
//
//	sc, ok := o.(*ClusterSupplyChain)
//	if ok {
//		for key, value := range sc.Spec.Selector {
//			res = append(res, fmt.Sprintf("%s: %s", key, value))
//		}
//	}
//
//	return res
//}

func init() {
	SchemeBuilder.Register(
		&ClusterSourceSupplyChain{},
		&ClusterSourceSupplyChainList{},
	)
}
