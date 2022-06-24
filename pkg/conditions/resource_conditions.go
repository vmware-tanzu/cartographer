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

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

// -- Resource.Conditions - ResourcesHealthy - True

func OutputAvailableResourcesHealthyCondition() metav1.Condition {
	return metav1.Condition{
		Type:   v1alpha1.ResourceHealthy,
		Status: metav1.ConditionTrue,
		Reason: v1alpha1.OutputAvailableResourcesHealthyReason,
	}
}

func AlwaysHealthyResourcesHealthyCondition() metav1.Condition {
	return metav1.Condition{
		Type:   v1alpha1.ResourceHealthy,
		Status: metav1.ConditionTrue,
		Reason: v1alpha1.AlwaysHealthyResourcesHealthyReason,
	}
}

func SingleConditionMatchCondition(status metav1.ConditionStatus, conditionName, message string) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.ResourceHealthy,
		Status:  status,
		Reason:  fmt.Sprintf("%sCondition", conditionName),
		Message: message,
	}
}

// -- Resource.Conditions - ResourcesHealthy - Unknown

func NoStampedObjectResourcesHealthyCondition() metav1.Condition {
	return metav1.Condition{
		Type:   v1alpha1.ResourceHealthy,
		Status: metav1.ConditionUnknown,
		Reason: v1alpha1.NoStampedObjectHealthyReason,
	}
}

func UnknownResourcesHealthyCondition() metav1.Condition {
	return metav1.Condition{
		Type:   v1alpha1.ResourceHealthy,
		Status: metav1.ConditionUnknown,
		Reason: "Unknown",
	}
}

func NoResourceResourcesHealthyCondition() metav1.Condition {
	return metav1.Condition{
		Type:   v1alpha1.ResourceHealthy,
		Status: metav1.ConditionUnknown,
		Reason: v1alpha1.NoResourceResourcesHealthyReason,
	}
}

func OutputNotAvailableResourcesHealthyCondition() metav1.Condition {
	return metav1.Condition{
		Type:   v1alpha1.ResourceHealthy,
		Status: metav1.ConditionUnknown,
		Reason: v1alpha1.OutputNotAvailableResourcesHealthyReason,
	}
}

func MultiMatchNoMatchesCondition() metav1.Condition {
	return metav1.Condition{
		Type:   v1alpha1.ResourceHealthy,
		Status: metav1.ConditionUnknown,
		Reason: v1alpha1.NoMatchesFulfilledReason,
	}
}

// -- Resource.Conditions - ResourcesHealthy - MultiMatch

func MultiMatchResourcesHealthyCondition(status metav1.ConditionStatus, reason, message string) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.ResourceHealthy,
		Status:  status,
		Reason:  reason,
		Message: message,
	}
}
