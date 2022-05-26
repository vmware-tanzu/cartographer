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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

func DetermineHealthCondition(rule *v1alpha1.HealthRule, realizedResource *v1alpha1.RealizedResource, stampedObject *unstructured.Unstructured) metav1.Condition {
	if rule == nil {
		if realizedResource == nil {
			return conditions.NoResourceResourcesHealthyCondition()
		} else if len(realizedResource.Outputs) > 0 {
			return conditions.OutputAvailableResourcesHealthyCondition()
		} else if realizedResource.TemplateRef != nil && realizedResource.TemplateRef.Kind == "ClusterTemplate" && realizedResource.TemplateRef.APIVersion == "carto.run/v1alpha1" {
			return conditions.AlwaysHealthyResourcesHealthyCondition()
		}
		return conditions.OutputNotAvailableResourcesHealthyCondition()
	} else {
		if rule.AlwaysHealthy != nil {
			return conditions.AlwaysHealthyResourcesHealthyCondition()
		}
		if rule.SingleConditionType != "" && stampedObject != nil {
			jsonpathQuery := fmt.Sprintf("{.status.conditions[?(@.type==\"%s\")].status}", rule.SingleConditionType)
			result, err := utils.SinglePathEvaluate(jsonpathQuery, stampedObject.UnstructuredContent())
			if err != nil {
				return conditions.SingleConditionTypeEvaluationErrorCondition(err)
			}
			if len(result) == 0 {
				return conditions.SingleConditionTypeNoResultResourcesCondition()
			}
			if resultString, ok := result[0].(string); ok {
				conditionStatus := metav1.ConditionStatus(resultString)
				if conditionStatus == metav1.ConditionFalse || conditionStatus == metav1.ConditionTrue {
					return conditions.SingleConditionMatchCondition(conditionStatus, rule.SingleConditionType)
				} else {
					return conditions.SingleConditionMatchCondition(metav1.ConditionUnknown, rule.SingleConditionType)
				}
			}
			return conditions.SingleConditionMatchCondition(metav1.ConditionUnknown, "")
		}
	}
	return conditions.UnknownResourcesHealthyCondition()
}
