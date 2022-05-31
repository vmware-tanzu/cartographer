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
// +groupName=test.run
// +kubebuilder:object:generate=true

package resources

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// TestObj A crd for our tests.
// this is a basic one, to avoid coupling, you can also build for-purpose objects in this directory
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type TestObj struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              TestSpec   `json:"spec"`
	Status            TestStatus `json:"status,omitempty"`
}

type TestStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}

type TestSpec struct {
	Foo string `json:"foo,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Value runtime.RawExtension `json:"value,omitempty"`
}

// TestObjList is a list of our TestObj CRD
// +kubebuilder:object:root=true
type TestObjList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TestObj `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&TestObj{},
		&TestObjList{},
	)
}
