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

package selector

import (
	"fmt"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/eval"
)


func Matches(req v1alpha1.FieldSelectorRequirement, context interface{}) (bool, error) {
	evaluator := eval.EvaluatorBuilder()
	actualValue, err := evaluator.EvaluateJsonPath(req.Key, context)
	if err != nil {
		return false, err
	}

	switch req.Operator {
	case v1alpha1.FieldSelectorOpIn:
		for _, v := range req.Values {
			if actualValue == v {
				return true, nil
			}
		}
		return false, nil
	case v1alpha1.FieldSelectorOpNotIn:
		for _, v := range req.Values {
			if actualValue == v {
				return false, nil
			}
		}
		return true, nil
	case v1alpha1.FieldSelectorOpExists:
		return actualValue != nil, nil
	case v1alpha1.FieldSelectorOpDoesNotExist:
		return actualValue == nil, nil
	default:
		return false, fmt.Errorf("invalid operator %s for field selector", req.Operator)
	}
}