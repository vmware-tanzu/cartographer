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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	RunnableReady    = "Ready"
	RunTemplateReady = "RunTemplateReady"
)

const (
	ReadyRunTemplateReason                            = "Ready"
	NotFoundRunTemplateReason                         = "RunTemplateNotFound"
	StampedObjectRejectedByAPIServerRunTemplateReason = "StampedObjectRejectedByAPIServer"
	OutputPathNotSatisfiedRunTemplateReason           = "OutputPathNotSatisfied"
	TemplateStampFailureRunTemplateReason             = "TemplateStampFailure"
	FailedToListCreatedObjectsReason                  = "FailedToListCreatedObjects"
	UnknownErrorReason                                = "UnknownError"
	ClientBuilderErrorResourcesSubmittedReason        = "ClientBuilderError"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type Runnable struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	// Spec describes the runnable.
	// More info: https://cartographer.sh/docs/latest/reference/runnable/#runnable
	Spec RunnableSpec `json:"spec"`

	// Status conforms to the Kubernetes conventions:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
	Status RunnableStatus `json:"status,omitempty"`
}

type RunnableStatus struct {
	ObservedGeneration int64                           `json:"observedGeneration,omitempty"`
	Conditions         []metav1.Condition              `json:"conditions,omitempty"`
	Outputs            map[string]apiextensionsv1.JSON `json:"outputs,omitempty"`
}

type RunnableSpec struct {
	// RunTemplateRef identifies the run template used to produce resources
	// for this runnable.
	// +kubebuilder:validation:Required
	RunTemplateRef TemplateReference `json:"runTemplateRef"`

	// Selector refers to an additional object that the template can refer
	// to using: $(selected)$.
	// +optional
	Selector *ResourceSelector `json:"selector,omitempty"`

	// Inputs are key/values providing inputs to the templated object created for this runnable.
	// Reference inputs in the template using the jsonPath: $(runnable.spec.inputs.<key>)$
	Inputs map[string]apiextensionsv1.JSON `json:"inputs,omitempty"`

	// ServiceAccountName refers to the Service account with permissions to create resources
	// submitted by the ClusterRunTemplate.
	//
	// If not set, Cartographer will use the default service account in the
	// runnable's namespace.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// RetentionPolicy specifies how many successful and failed runs should be retained.
	// Runs older than this (ordered by creation time) will be deleted.
	// +kubebuilder:default={maxFailedRuns: 10, maxSuccessfulRuns: 10}
	RetentionPolicy RetentionPolicy `json:"retentionPolicy,omitempty"`
}

type RetentionPolicy struct {
	// MaxFailedRuns is the number of failed runs to retain.
	// +kubebuilder:validation:Minimum:=1
	MaxFailedRuns int64 `json:"maxFailedRuns"`
	// MaxSuccessfulRuns is the number of successful runs to retain.
	// +kubebuilder:validation:Minimum:=1
	MaxSuccessfulRuns int64 `json:"maxSuccessfulRuns"`
}

type ResourceSelector struct {
	// Resource is the GVK that must match the selected object.
	Resource ResourceType `json:"resource"`

	// MatchingLabels must match on a single target object, making the object
	// available in the template as $(selected)$
	MatchingLabels map[string]string `json:"matchingLabels"`
}

type ResourceType struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`
}

type TemplateReference struct {
	Kind string `json:"kind,omitempty"`

	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
}

// +kubebuilder:object:root=true

type RunnableList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Runnable `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&Runnable{},
		&RunnableList{},
	)
}
