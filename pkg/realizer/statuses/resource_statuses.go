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

package statuses

import (
	"reflect"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
)

type ResourceStatuses interface {
	GetPreviousResourceStatus(realizedResourceName string) *v1alpha1.ResourceStatus
	Add(status *v1alpha1.RealizedResource, err error, furtherConditions ...metav1.Condition)
	GetCurrent() []v1alpha1.ResourceStatus
	IsChanged() bool
}

type AddConditionsFunc func(conditionManager *conditions.ConditionManager, isOwner bool, err error)

func NewResourceStatuses(previousResourceStatuses []v1alpha1.ResourceStatus, addConditionsFunc AddConditionsFunc) *resourceStatuses {
	var statuses []*resourceStatus

	for i := range previousResourceStatuses {
		statuses = append(statuses, &resourceStatus{
			name:     previousResourceStatuses[i].Name,
			previous: &previousResourceStatuses[i],
			current:  nil,
		})
	}

	return &resourceStatuses{
		statuses:          statuses,
		addConditionsFunc: addConditionsFunc,
	}
}

type resourceStatus struct {
	name              string
	previous          *v1alpha1.ResourceStatus
	current           *v1alpha1.ResourceStatus
	conditionsChanged bool
}

type resourceStatuses struct {
	statuses          []*resourceStatus
	addConditionsFunc AddConditionsFunc
}

func (r *resourceStatuses) IsChanged() bool {
	for _, status := range r.statuses {
		if status.current == nil {
			return true
		}
		if status.previous == nil {
			return true
		}
		if status.conditionsChanged {
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

	for _, status := range r.statuses {
		if status != nil && status.current != nil {
			currentStatuses = append(currentStatuses, *status.current)
		}
	}

	return currentStatuses
}

func (r *resourceStatuses) GetPreviousResourceStatus(realizedResourceName string) *v1alpha1.ResourceStatus {
	for _, status := range r.statuses {
		if status.name == realizedResourceName {
			return status.previous
		}
	}

	return nil
}

func (r *resourceStatuses) Add(realizedResource *v1alpha1.RealizedResource, err error, furtherConditions ...metav1.Condition) {
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

	//TODO: okay this is necessary? really? sure? okay, but if really, then reason about and ensure semantic correctness with LastTransitionTime.....
	if len(furtherConditions) > 0 {
		furtherConditions[0].LastTransitionTime = metav1.Time{
			Time: time.Now(),
		}
	}

	existingStatus.current = &v1alpha1.ResourceStatus{
		RealizedResource: *realizedResource,
		Conditions:       append(r.createConditions(name, err), furtherConditions...),
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
		r.addConditionsFunc(&conditionManager, false, err)
	} else {
		conditionManager.AddPositive(conditions.ResourceSubmittedCondition())
	}

	resourceConditions, changed := conditionManager.Finalize()
	existingStatus.conditionsChanged = changed
	return resourceConditions
}
