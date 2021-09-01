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

package conditions

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// To learn more about condition conventions:
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

// Polarity represents how a Status is represented as success.
type Polarity string

// Positive Polarity means a "True" ConditionStatus is a success
const Positive Polarity = "Positive"

// Negative Polarity means a "False" ConditionStatus is a success
const Negative Polarity = "Negative"

//counterfeiter:generate . ConditionManager

// ConditionManager supports collecting condition statuses for your controller
// It adds a complete top level condition when Finalize is called.
//
// TBD: either error or warn if the same Condition.Type is reused
// TBD2: should accept an existing []Condition slice to compare against
//		should only update LastTransitionTime if the other fields have changed
type ConditionManager interface {
	// Add a condition and associate a polarity with it.
	Add(condition metav1.Condition, positive Polarity)

	// AddPositive Adds a condition with a positive polarity
	AddPositive(condition metav1.Condition)

	// AddNegative Adds a condition with a negative polarity
	AddNegative(condition metav1.Condition)

	// Finalize	returns all conditions
	// not idempotent! subsequent finalizes will keep adding Parent conditions
	// The changed result represents whether the conditions have changed enough to warrant an update to the APIServer
	Finalize() (conditions []metav1.Condition, changed bool)

	// IsSuccessful can be called any time. Start's off true, but an
	// add of an unsuccessful condition (Positive-False or Negative-True)
	// causes this to return false
	IsSuccessful() bool
}

type conditionManager struct {
	previousConditions []metav1.Condition
	conditions         []metav1.Condition
	topLevelType       string
	status             metav1.ConditionStatus
	changed            bool
}

type ConditionManagerBuilder func(topLevelType string, previousConditions []metav1.Condition) ConditionManager

// NewConditionManager returns a ConditionManager with a top level
// Condition.Type specified in topLevelType
func NewConditionManager(topLevelType string, previousConditions []metav1.Condition) ConditionManager {
	return &conditionManager{
		previousConditions: previousConditions,
		conditions:         []metav1.Condition{},
		topLevelType:       topLevelType,
		status:             metav1.ConditionTrue,
	}
}

func (c *conditionManager) Add(condition metav1.Condition, polarity Polarity) {
	condition.LastTransitionTime = metav1.Now()

	if (condition.Status == metav1.ConditionFalse && polarity == Positive) ||
		(condition.Status == metav1.ConditionTrue && polarity == Negative) {
		c.status = metav1.ConditionFalse
	} else if condition.Status == metav1.ConditionUnknown {
		if c.status == metav1.ConditionTrue {
			c.status = metav1.ConditionUnknown
		}
	}

	isNewCondition := true

	for _, previousCondition := range c.previousConditions {
		if previousCondition.Type == condition.Type {
			isNewCondition = false
			lastTransitionTime := condition.LastTransitionTime
			condition.LastTransitionTime = previousCondition.LastTransitionTime
			if !reflect.DeepEqual(previousCondition, condition) {
				condition.LastTransitionTime = lastTransitionTime
				c.changed = true
			}
		}
	}

	if isNewCondition {
		c.changed = true
	}

	c.conditions = append(c.conditions, condition)
}

func (c *conditionManager) IsSuccessful() bool {
	return !(c.status == metav1.ConditionFalse)
}

func (c *conditionManager) AddPositive(condition metav1.Condition) {
	c.Add(condition, Positive)
}

func (c *conditionManager) AddNegative(condition metav1.Condition) {
	c.Add(condition, Negative)
}

func (c *conditionManager) Finalize() ([]metav1.Condition, bool) {
	if len(c.conditions) == 0 {
		c.status = metav1.ConditionFalse
		return []metav1.Condition{{
			Type:               c.topLevelType,
			Status:             "Unknown",
			Reason:             "Unknown",
			LastTransitionTime: metav1.Now(),
		}}, true
	}

	var status metav1.ConditionStatus
	var message, reason string

	switch c.status {
	case metav1.ConditionTrue:
		status = metav1.ConditionTrue
		message = ""
		reason = "Ready"
	case metav1.ConditionFalse:
		status = metav1.ConditionFalse
		message = "not all conditions are met"
		reason = "ConditionsUnmet"
	case metav1.ConditionUnknown:
		status = metav1.ConditionUnknown
		reason = "ConditionInUnknownState"
	}

	c.AddPositive(
		metav1.Condition{
			Type:               c.topLevelType,
			Status:             status,
			LastTransitionTime: metav1.Now(),
			Reason:             reason,
			Message:            message,
		},
	)

	return c.conditions, c.changed
}
