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

package resources

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
)

type ResourceStatuses interface {
	GetPreviousRealizedResource(name string) *v1alpha1.RealizedResource
	Add(status *v1alpha1.RealizedResource, err error)
	GetCurrent() []v1alpha1.ResourceStatus
	IsChanged() bool
}

func NewResourceStatuses(previousResourceStatuses []v1alpha1.ResourceStatus) *resourceStatuses {
	var statuses []*resourceStatus

	for _, previousResourceStatus := range previousResourceStatuses {

		statuses = append(statuses, &resourceStatus{
			Name: previousResourceStatus.Name,
			Previous: &v1alpha1.ResourceStatus{
				RealizedResource: previousResourceStatus.RealizedResource,
			},
			Current: nil,
		})
	}

	return &resourceStatuses{
		statuses: statuses,
	}
}

type resourceStatus struct {
	Name     string
	Previous *v1alpha1.ResourceStatus
	Current  *v1alpha1.ResourceStatus
}

type resourceStatuses struct {
	statuses []*resourceStatus
}

func (r *resourceStatuses) IsChanged() bool {
	// TODO
	return false
}

func (r *resourceStatuses) GetCurrent() []v1alpha1.ResourceStatus {
	var currentStatuses []v1alpha1.ResourceStatus

	// TODO: if current is nil, we need to use previous?
	// ...In the case of reconciler failing before we get to the realizer
	for _, status := range r.statuses {
		if status != nil && status.Current != nil {
			currentStatuses = append(currentStatuses, *status.Current)
		}
	}

	return currentStatuses
}

func (r *resourceStatuses) GetPreviousRealizedResource(name string) *v1alpha1.RealizedResource {
	for _, status := range r.statuses {
		if status.Name == name {
			return &status.Previous.RealizedResource
		}
	}

	return nil
}

func (r *resourceStatuses) Add(realizedResource *v1alpha1.RealizedResource, err error) {
	name := realizedResource.Name

	var existingStatus *resourceStatus
	for _, status := range r.statuses {
		if status.Name == name {
			existingStatus = status
			break
		}
	}

	if existingStatus == nil {
		existingStatus = &resourceStatus{
			Name: name,
		}
		r.statuses = append(r.statuses, existingStatus)
	}

	existingStatus.Current = &v1alpha1.ResourceStatus{
		RealizedResource: *realizedResource,
		Conditions:       r.createConditions(name, err),
	}
}

func (r *resourceStatuses) createConditions(name string, err error) []metav1.Condition {
	var existingStatus *resourceStatus
	for _, status := range r.statuses {
		if status.Name == name {
			existingStatus = status
			break
		}
	}

	var previousConditions []metav1.Condition
	if existingStatus.Previous != nil {
		previousConditions = existingStatus.Previous.Conditions
	}

	conditionManager := conditions.NewConditionManager(v1alpha1.ResourceReady, previousConditions)
	if err != nil {
		conditions.AddConditionForResourceSubmitted(&conditionManager, false, err)
	} else {
		conditionManager.AddPositive(conditions.ResourceSubmittedCondition())
	}

	resourceConditions, _ := conditionManager.Finalize()
	return resourceConditions
}
