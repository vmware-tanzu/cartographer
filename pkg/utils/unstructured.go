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

package utils

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ConditionList []metav1.Condition

func (c ConditionList) ConditionWithType(conditionType string) *metav1.Condition {
	for _, condition := range c {
		if condition.Type == conditionType {
			return &condition
		}
	}
	return nil
}

func ExtractConditions(stampedObject *unstructured.Unstructured) ConditionList {
	var conditionList ConditionList
	maybeStatus := stampedObject.UnstructuredContent()["status"]
	if unstructuredStatus, statusOk := maybeStatus.(map[string]interface{}); statusOk {
		maybeConditions := unstructuredStatus["conditions"]
		maybeConditionsJSON, err := json.Marshal(maybeConditions)
		if err == nil {
			err = json.Unmarshal(maybeConditionsJSON, &conditionList)
			if err != nil {
				return ConditionList{}
			}
		}
	}
	return conditionList
}
