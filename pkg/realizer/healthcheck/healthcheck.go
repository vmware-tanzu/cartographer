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

package healthcheck

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/eval"
	"github.com/vmware-tanzu/cartographer/pkg/selector"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

func IsClusterTemplate(reference *corev1.ObjectReference) bool {
	apiVersion := v1alpha1.SchemeGroupVersion.Group + "/" + v1alpha1.SchemeGroupVersion.Version
	if reference != nil &&
		reference.Kind == "ClusterTemplate" &&
		reference.APIVersion == apiVersion {
		return true
	}
	return false
}

func OwnerHealthCondition(resourceStatuses []v1alpha1.ResourceStatus, previousConditions []metav1.Condition) metav1.Condition {
	var previousHealthCondition []metav1.Condition
	condition := utils.ConditionList(previousConditions).ConditionWithType(v1alpha1.ResourceHealthy)
	if condition != nil {
		previousHealthCondition = append(previousHealthCondition, *condition)
	}
	healthyConditionManager := conditions.NewConditionManager(v1alpha1.ResourcesHealthy, previousHealthCondition)

	for _, resourceStatus := range resourceStatuses {
		resourceHealthyCondition := utils.ConditionList(resourceStatus.Conditions).ConditionWithType(v1alpha1.ResourceHealthy)
		if resourceHealthyCondition != nil {
			healthyConditionManager.AddPositive(*resourceHealthyCondition)
		}
	}

	healthyConditionResult, _ := healthyConditionManager.Finalize()
	healthyCondition := healthyConditionResult[len(healthyConditionResult)-1]
	healthyCondition.Reason = "HealthyConditionRule"
	return healthyCondition
}

func DetermineHealthCondition(rule *v1alpha1.HealthRule, realizedResource *v1alpha1.RealizedResource, stampedObject *unstructured.Unstructured) metav1.Condition {
	if rule == nil {
		if realizedResource == nil {
			return conditions.NoResourceResourcesHealthyCondition()
		} else if len(realizedResource.Outputs) > 0 {
			return conditions.OutputAvailableResourcesHealthyCondition()
		} else if IsClusterTemplate(realizedResource.TemplateRef) {
			if realizedResource.StampedRef != nil {
				return conditions.AlwaysHealthyResourcesHealthyCondition()
			} else {
				return conditions.NoStampedObjectResourcesHealthyCondition()
			}
		}
		return conditions.OutputNotAvailableResourcesHealthyCondition()
	} else {
		if rule.AlwaysHealthy != nil {
			return conditions.AlwaysHealthyResourcesHealthyCondition()
		}
		if stampedObject != nil {
			if rule.SingleConditionType != "" {
				return singleConditionTypeCondition(rule.SingleConditionType, stampedObject)
			}
			if rule.MultiMatch != nil {
				return multiMatchCondition(rule.MultiMatch, stampedObject)
			}
		}
	}
	return conditions.UnknownResourcesHealthyCondition()
}

func singleConditionTypeCondition(singleConditionType string, stampedObject *unstructured.Unstructured) metav1.Condition {
	singleCondition := utils.ExtractConditions(stampedObject).ConditionWithType(singleConditionType)
	if singleCondition != nil {
		if singleCondition.Status == metav1.ConditionFalse || singleCondition.Status == metav1.ConditionTrue {
			return conditions.SingleConditionMatchCondition(singleCondition.Status, singleConditionType, singleCondition.Message)
		} else {
			return conditions.SingleConditionMatchCondition(metav1.ConditionUnknown, singleConditionType, singleCondition.Message)
		}
	}
	return conditions.SingleConditionMatchCondition(metav1.ConditionUnknown, singleConditionType, fmt.Sprintf("condition with type [%s] not found on resource status", singleConditionType))
}

func multiMatchCondition(multiMatchRule *v1alpha1.MultiMatchHealthRule, stampedObject *unstructured.Unstructured) metav1.Condition {
	condition := anyUnhealthyMatchCondition(multiMatchRule.Unhealthy, stampedObject)
	if condition != nil {
		return *condition
	}
	condition = allHealthyMatchCondition(multiMatchRule.Healthy, stampedObject)
	if condition != nil {
		return *condition
	}
	return conditions.MultiMatchNoMatchesCondition()
}

func messageForMatchingFieldRequirement(requirement v1alpha1.HealthMatchFieldSelectorRequirement, stampedObject *unstructured.Unstructured) string {
	evaluator := eval.EvaluatorBuilder()
	fieldValue, fieldErr := evaluator.EvaluateJsonPath(requirement.Key, stampedObject.UnstructuredContent())
	if fieldErr != nil {
		fieldValue = "<error retrieving field value>"
	}
	messageValue, messageErr := evaluator.EvaluateJsonPath(requirement.MessagePath, stampedObject.UnstructuredContent())
	if messageErr != nil {
		messageValue = fmt.Sprintf("unknown, error retrieving message path [%s]", requirement.MessagePath)
	}

	return fmt.Sprintf("field value: %v, message: %v", fieldValue, messageValue)
}

func anyUnhealthyMatchCondition(rule v1alpha1.HealthMatchRule, stampedObject *unstructured.Unstructured) *metav1.Condition {
	for _, conditionRule := range rule.MatchConditions {
		singleCondition := utils.ExtractConditions(stampedObject).ConditionWithType(conditionRule.Type)
		if singleCondition != nil && singleCondition.Status == conditionRule.Status {
			condition := conditions.MultiMatchResourcesHealthyCondition(metav1.ConditionFalse,
				v1alpha1.MultiMatchConditionHealthyReason,
				fmt.Sprintf("condition status: %s, message: %s", singleCondition.Status, singleCondition.Message))
			return &condition
		}
	}
	for _, matchFieldRule := range rule.MatchFields {
		matches, _ := selector.Matches(matchFieldRule.FieldSelectorRequirement, stampedObject.UnstructuredContent())
		if matches {
			condition := conditions.MultiMatchResourcesHealthyCondition(metav1.ConditionFalse,
				v1alpha1.MultiMatchFieldHealthyReason,
				messageForMatchingFieldRequirement(matchFieldRule, stampedObject))
			return &condition
		}
	}
	return nil
}

func allHealthyMatchCondition(rule v1alpha1.HealthMatchRule, stampedObject *unstructured.Unstructured) *metav1.Condition {
	var firstReason string
	var message string
	for _, conditionRule := range rule.MatchConditions {
		resourceCondition := utils.ExtractConditions(stampedObject).ConditionWithType(conditionRule.Type)
		if resourceCondition == nil || resourceCondition.Status != conditionRule.Status {
			return nil
		}
		if firstReason == "" {
			firstReason = v1alpha1.MultiMatchConditionHealthyReason
			message = fmt.Sprintf("condition status: %s, message: %s", resourceCondition.Status, resourceCondition.Message)
		}
	}
	for _, matchFieldRule := range rule.MatchFields {
		matches, err := selector.Matches(matchFieldRule.FieldSelectorRequirement, stampedObject.UnstructuredContent())
		if err != nil || !matches {
			return nil
		}
		if firstReason == "" {
			firstReason = v1alpha1.MultiMatchFieldHealthyReason
			message = messageForMatchingFieldRequirement(matchFieldRule, stampedObject)
		}
	}
	condition := conditions.MultiMatchResourcesHealthyCondition(metav1.ConditionTrue, firstReason, message)
	return &condition
}
