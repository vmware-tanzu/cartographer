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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ValidDeliverablePaths = map[string]bool{
	"spec.source":                true,
	"spec.source.git":            true,
	"spec.source.git.url":        true,
	"spec.source.git.ref":        true,
	"spec.source.git.ref.branch": true,
	"spec.source.git.ref.tag":    true,
	"spec.source.git.ref.commit": true,
	"spec.source.image":          true,
	"spec.source.subPath":        true,
	"spec.serviceAccountName":    true,
}

var ValidDeliverablePrefixes = []string{
	"spec.params",
	"metadata",
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=deliverables,categories=all,shortName=dlv
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Source",type="string",JSONPath=`.spec.source['git.url','image']`
// +kubebuilder:printcolumn:name="Delivery",type="string",JSONPath=".status.deliveryRef.name"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=='Ready')].status`
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=`.status.conditions[?(@.type=='Ready')].reason`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`

type Deliverable struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	// Spec describes the deliverable.
	// More info: https://cartographer.sh/docs/latest/reference/workload/#deliverable
	Spec DeliverableSpec `json:"spec"`

	// Status conforms to the Kubernetes conventions:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
	Status DeliverableStatus `json:"status,omitempty"`
}

type DeliverableSpec struct {
	// Additional parameters.
	// See: https://cartographer.sh/docs/latest/architecture/#parameter-hierarchy
	// +optional
	Params []OwnerParam `json:"params,omitempty"`

	// The location of the source code for the workload. Specify
	// one of `spec.source` or `spec.image`
	// +optional
	Source *Source `json:"source,omitempty"`

	// ServiceAccountName refers to the Service account with permissions to create resources
	// submitted by the supply chain.
	//
	// If not set, Cartographer will use serviceAccountName from delivery.
	//
	// If that is also not set, Cartographer will use the default service account in the
	// deliverable's namespace.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}

type DeliverableStatus struct {
	OwnerStatus `json:",inline"`

	// DeliveryRef is the Delivery resource that was used when this status was set.
	DeliveryRef ObjectReference `json:"deliveryRef,omitempty"`

	// Resources contain references to the objects created by the Delivery and the templates used to create them.
	// It also contains Inputs and Outputs that were passed between the templates as the Delivery was processed.
	Resources []ResourceStatus `json:"resources,omitempty"`
}

// +kubebuilder:object:root=true

type DeliverableList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Deliverable `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&Deliverable{},
		&DeliverableList{},
	)
}
