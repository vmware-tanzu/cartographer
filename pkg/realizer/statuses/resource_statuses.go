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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

type ResourceStatuses interface {
	ChangedConditionTypes(realizedResourceName string) []string
	GetPreviousResourceStatus(realizedResourceName string) *v1alpha1.ResourceStatus
	Add(status *v1alpha1.RealizedResource, resourceIndex int, err error, isPassThrough bool, furtherConditions ...metav1.Condition)
	GetCurrent() ResourceStatusList
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

type ResourceStatusList []v1alpha1.ResourceStatus

func (rsl ResourceStatusList) ConditionsForResourceNamed(resourceName string) utils.ConditionList {
	for _, status := range rsl {
		if status.Name == resourceName {
			return status.Conditions
		}
	}
	return nil
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

func (r *resourceStatuses) GetCurrent() ResourceStatusList {
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

func (r *resourceStatuses) Add(realizedResource *v1alpha1.RealizedResource, resourceIndex int, err error, isPassThrough bool, furtherConditions ...metav1.Condition) {
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
		r.statuses = append(r.statuses, &resourceStatus{})
		copy(r.statuses[resourceIndex+1:], r.statuses[resourceIndex:len(r.statuses)-1])
		r.statuses[resourceIndex] = existingStatus
	}

	existingStatus.current = &v1alpha1.ResourceStatus{
		RealizedResource: *realizedResource,
		Conditions:       r.createConditions(name, err, isPassThrough, furtherConditions...),
	}
}

func (r *resourceStatuses) ChangedConditionTypes(realizedResourceName string) []string {
	var changed []string
	for _, status := range r.statuses {
		if status.name == realizedResourceName {
			if status.current != nil {
				for _, condition := range status.current.Conditions {
					if conditionChanged(status.previous, condition) {
						changed = append(changed, condition.Type)
					}
				}
			}
		}
	}
	return changed
}

func conditionChanged(previousResourceStatus *v1alpha1.ResourceStatus, newCondition metav1.Condition) bool {
	if previousResourceStatus != nil {
		for _, previousCondition := range previousResourceStatus.Conditions {
			if previousCondition.Type == newCondition.Type {
				return previousCondition.Status != newCondition.Status
			}
		}
	}
	return newCondition.Status != "Unknown"
}

func (r *resourceStatuses) createConditions(name string, err error, isPassThrough bool, furtherConditions ...metav1.Condition) []metav1.Condition {
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
		conditionManager.AddPositive(conditions.ResourceSubmittedCondition(isPassThrough))
	}
	for _, condition := range furtherConditions {
		conditionManager.AddPositive(condition)
	}

	resourceConditions, changed := conditionManager.Finalize()
	existingStatus.conditionsChanged = changed
	return resourceConditions
}
