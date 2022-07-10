package v2alpha1

// +versionName=v2alpha1
// +groupName=carto.run
// +kubebuilder:object:generate=true

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Todo: 100% open to better ideas than "imprint"
// Todo: Put the status structre in here
// 			Personally, I don't want to revert to the overcommunication we have in workload. But I have no idea how to pare that down?

// ImprintStatus represents the overall status of the objects maintained for
// an owner object.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=imprintstatuses,scope=Namespaced,shortName=is
type ImprintStatus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

// +kubebuilder:object:root=true

type ImprintStatusList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ImprintStatus `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&ImprintStatus{},
		&ImprintStatusList{},
	)
}
