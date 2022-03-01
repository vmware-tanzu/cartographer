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
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	WorkloadReady             = "Ready"
	WorkloadSupplyChainReady  = "SupplyChainReady"
	WorkloadResourceSubmitted = "ResourcesSubmitted"
)

const (
	ReadySupplyChainReason                               = "Ready"
	WorkloadLabelsMissingSupplyChainReason               = "WorkloadLabelsMissing"
	NotFoundSupplyChainReadyReason                       = "SupplyChainNotFound"
	MultipleMatchesSupplyChainReadyReason                = "MultipleSupplyChainMatches"
	ServiceAccountSecretErrorResourcesSubmittedReason    = "ServiceAccountSecretError"
	ResourceRealizerBuilderErrorResourcesSubmittedReason = "ResourceRealizerBuilderError"
	ResolveTemplateOptionsErrorResourcesSubmittedReason  = "ResolveTemplateOptionsError"
	TemplateOptionsMatchErrorResourcesSubmittedReason    = "TemplateOptionsMatchError"
)

// ValidWorkloadPaths Note: this needs to be updated anytime the spec changes
var ValidWorkloadPaths = map[string]bool{
	"workload.spec.source":                true,
	"workload.spec.source.git":            true,
	"workload.spec.source.git.url":        true,
	"workload.spec.source.git.ref":        true,
	"workload.spec.source.git.ref.branch": true,
	"workload.spec.source.git.ref.tag":    true,
	"workload.spec.source.git.ref.commit": true,
	"workload.spec.source.image":          true,
	"workload.spec.source.subPath":        true,
	"workload.spec.build":                 true,
	"workload.spec.image":                 true,
	"workload.spec.serviceAccountName":    true,
}

// ValidWorkloadPrefixes Note: this needs to be updated anytime the spec changes
var ValidWorkloadPrefixes = []string{
	"workload.spec.params",
	"workload.spec.build.env",
	"workload.spec.env",
	"workload.spec.resources",
	"workload.spec.serviceClaims",
	"workload.metadata",
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories="all"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Source",type="string",JSONPath=`.spec.source['git.url','image']`
// +kubebuilder:printcolumn:name="SupplyChain",type="string",JSONPath=".status.supplyChainRef.name"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=='Ready')].status`
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=`.status.conditions[?(@.type=='Ready')].reason`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`

type Workload struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	// Spec describes the workload.
	// More info: https://cartographer.sh/docs/latest/reference/workload/#workload
	Spec WorkloadSpec `json:"spec"`

	// Status conforms to the Kubernetes conventions:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
	Status WorkloadStatus `json:"status,omitempty"`
}

type WorkloadServiceClaim struct {
	Name string                         `json:"name"`
	Ref  *WorkloadServiceClaimReference `json:"ref,omitempty"`
}

type WorkloadServiceClaimReference struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}

type WorkloadSpec struct {
	// Additional parameters.
	// See: https://cartographer.sh/docs/latest/architecture/#parameter-hierarchy
	// +optional
	Params []OwnerParam `json:"params,omitempty"`

	// The location of the source code for the workload. Specify
	// one of `spec.source` or `spec.image`
	// +optional
	Source *Source `json:"source,omitempty"`

	// Build configuration, for the build resources in the supply chain
	// +optional
	Build WorkloadBuild `json:"build,omitempty"`

	// Environment variables to be passed to the main container
	// running the application.
	// See https://kubernetes.io/docs/tasks/inject-data-application/environment-variable-expose-pod-information/
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Image refers to a pre-built image in a registry. It is an alternative
	// to specifying the location of source code for the workload. Specify
	// one of `spec.source` or `spec.image`.
	// +optional
	Image *string `json:"image,omitempty"`

	// Resource constraints for the application. See https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// ServiceAccountName refers to the Service account with permissions to create resources
	// submitted by the supply chain.
	//
	// If not set, Cartographer will use serviceAccountName from supply chain.
	//
	// If that is also not set, Cartographer will use the default service account in the
	// workload's namespace.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// ServiceClaims to be bound through ServiceBindings.
	// +optional
	ServiceClaims []WorkloadServiceClaim `json:"serviceClaims,omitempty"`
}

type WorkloadBuild struct {
	// Env is an array of environment variables to propagate to build resources in the
	// supply chain.
	// See https://kubernetes.io/docs/tasks/inject-data-application/environment-variable-expose-pod-information/
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`
}

type WorkloadStatus struct {
	OwnerStatus `json:",inline"`

	// SupplyChainRef is the Supply Chain resource that was used when this status was set.
	SupplyChainRef ObjectReference `json:"supplyChainRef,omitempty"`

	// Resources contain references to the objects created by the Supply Chain and the templates used to create them.
	// It also contains Inputs and Outputs that were passed between the templates as the Supply Chain was processed.
	Resources []RealizedResource `json:"resources,omitempty"`
}

// +kubebuilder:object:root=true

type WorkloadList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workload `json:"items"`
}

func validWorkloadPath(path string) bool {
	if ValidWorkloadPaths[path] {
		return true
	}

	for _, prefix := range ValidWorkloadPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

func init() {
	SchemeBuilder.Register(
		&Workload{},
		&WorkloadList{},
	)
}
