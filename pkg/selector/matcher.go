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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/eval"
)

type Selectable interface {
	GetLabels() map[string]string
}

type MatchError interface {
	error
	SelectorIndex() int
}

type selectorMatchError struct {
	Err                  error
	SelectingObjectIndex int
}

func (e selectorMatchError) Error() string {
	return e.Err.Error()
}

func (e selectorMatchError) SelectorIndex() int {
	return e.SelectingObjectIndex
}

// BestSelectorMatchIndices returns a slice of indices into the passed in `selectors` which match the
// `selectable` with the most specificity. Any error processing a selector includes the index of the
// offending Selector.
func BestSelectorMatchIndices(selectable Selectable, selectors []v1alpha1.Selector) ([]int, MatchError) {
	var mostSpecificMatchingSelectors []int
	var highWaterMark = 1

	for idx, selector := range selectors {
		matchScore := 0

		// -- Labels
		sel, err := metav1.LabelSelectorAsSelector(&selector.LabelSelector)
		if err != nil {
			return nil, selectorMatchError{
				Err:                  fmt.Errorf("selector labels or matchExpressions are not valid: %w", err),
				SelectingObjectIndex: idx,
			}
		}
		if !sel.Matches(labels.Set(selectable.GetLabels())) {
			continue // Bail early!
		}

		matchScore += len(selector.MatchLabels)
		matchScore += len(selector.MatchExpressions)

		// -- Fields
		allFieldsMatched, err := matchesAllFields(selectable, selector.MatchFields)
		if err != nil {
			return nil, selectorMatchError{
				Err:                  fmt.Errorf("failed to evaluate selector matchFields: %w", err),
				SelectingObjectIndex: idx,
			}
		}
		if !allFieldsMatched {
			continue // Bail early!
		}
		matchScore += len(selector.MatchFields)

		// -- decision time
		if matchScore == highWaterMark {
			mostSpecificMatchingSelectors = append(mostSpecificMatchingSelectors, idx)
		} else if matchScore > highWaterMark {
			highWaterMark = matchScore
			mostSpecificMatchingSelectors = []int{idx}
		}
	}

	return mostSpecificMatchingSelectors, nil
}

func matchesAllFields(source interface{}, requirements []v1alpha1.FieldSelectorRequirement) (bool, error) {
	for _, requirement := range requirements {
		match, err := Matches(requirement, source)
		if err != nil {
			if _, ok := err.(eval.JsonPathDoesNotExistError); !ok {
				return false, fmt.Errorf("unable to match field requirement with key [%s] operator [%s] values [%v]: %w", requirement.Key, requirement.Operator, requirement.Values, err)
			}
		}
		if !match {
			return false, nil
		}
	}
	return true, nil
}
