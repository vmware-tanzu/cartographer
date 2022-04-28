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

package workload

import (
	"reflect"

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
			name: previousResourceStatus.Name,
			previous: &v1alpha1.ResourceStatus{
				RealizedResource: previousResourceStatus.RealizedResource,
			},
			current: nil,
		})
	}

	return &resourceStatuses{
		statuses: statuses,
	}
}

type resourceStatus struct {
	name     string
	previous *v1alpha1.ResourceStatus
	current  *v1alpha1.ResourceStatus
	changed  bool
}

type resourceStatuses struct {
	statuses []*resourceStatus
}

func (r *resourceStatuses) IsChanged() bool {
	for _, status := range r.statuses {
		if status.current == nil {
			continue
		}
		if status.changed {
			return true
		}
		if status.previous == nil {
			return true
		}
		if !reflect.DeepEqual(status.current.RealizedResource, status.previous.RealizedResource) {
			return true
		}
	}
	return false
}

func (r *resourceStatuses) GetCurrent() []v1alpha1.ResourceStatus {
	var currentStatuses []v1alpha1.ResourceStatus

	// TODO: if current is nil, we need to use previous?
	// ...In the case of reconciler failing before we get to the realizer
	for _, status := range r.statuses {
		if status != nil && status.current != nil {
			currentStatuses = append(currentStatuses, *status.current)
		}
	}

	return currentStatuses
}

func (r *resourceStatuses) GetPreviousRealizedResource(name string) *v1alpha1.RealizedResource {
	for _, status := range r.statuses {
		if status.name == name {
			return &status.previous.RealizedResource
		}
	}

	return nil
}

func (r *resourceStatuses) Add(realizedResource *v1alpha1.RealizedResource, err error) {
	name := realizedResource.Name

	var existingStatus *resourceStatus
	for _, status := range r.statuses {
		if status.name == name {
			existingStatus = status
			break
		}
	}

	if existingStatus == nil {
		existingStatus = &resourceStatus{
			name: name,
		}
		r.statuses = append(r.statuses, existingStatus)
	}

	existingStatus.current = &v1alpha1.ResourceStatus{
		RealizedResource: *realizedResource,
		Conditions:       r.createConditions(name, err),
	}
}

func (r *resourceStatuses) createConditions(name string, err error) []metav1.Condition {
	var existingStatus *resourceStatus
	for _, status := range r.statuses {
		if status.name == name {
			existingStatus = status
			break
		}
	}

	var previousConditions []metav1.Condition
	if existingStatus.previous != nil {
		previousConditions = existingStatus.previous.Conditions
	}

	conditionManager := conditions.NewConditionManager(v1alpha1.ResourceReady, previousConditions)
	if err != nil {
		conditions.AddConditionForResourceSubmittedWorkload(&conditionManager, false, err)
	} else {
		conditionManager.AddPositive(conditions.ResourceSubmittedCondition())
	}

	resourceConditions, changed := conditionManager.Finalize()
	existingStatus.changed = changed
	return resourceConditions
}
