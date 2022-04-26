package controllers

import "github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"

type ResourceStatuses interface {
	GetPrevious(name string) *v1alpha1.ResourceStatus
	Add(status *v1alpha1.ResourceStatus)
}

func NewResourceStatuses() *resourceStatuses {
	return &resourceStatuses{
		statuses: map[string]*resourceStatus{},
	}
}

type resourceStatus struct {
	Previous *v1alpha1.ResourceStatus
	Current  *v1alpha1.ResourceStatus
}

type resourceStatuses struct {
	statuses map[string]*resourceStatus
}

func (r resourceStatuses) GetPrevious(name string) *v1alpha1.ResourceStatus {
	return r.statuses[name].Previous
}

func (r resourceStatuses) Add(status *v1alpha1.ResourceStatus) {
	if r.statuses[status.Name] == nil {
		r.statuses[status.Name] = &resourceStatus{}
	}
	r.statuses[status.Name].Current = status
}
