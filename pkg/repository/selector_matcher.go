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

package repository

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/selector"
)

type SelectingObject interface {
	GetSelectors() v1alpha1.Selectors
	GetObjectKind() schema.ObjectKind
	GetName() string
}

type Selectable interface {
	GetLabels() map[string]string
}

// BestSelectorMatch attempts at finding the selectors that best match their selectors
// against the selectors.
func BestSelectorMatch(selectable Selectable, blueprints []SelectingObject) ([]SelectingObject, error) {

	sc := &v1alpha1.ClusterSupplyChain{}

	sc.GetSelectors()
	sc.GetObjectKind()
	sc.GetName()
	if len(blueprints) == 0 {
		return nil, nil
	}

	var matchingSelectors = map[int][]SelectingObject{}
	var highWaterMark = 0

	for _, target := range blueprints {
		selectors := target.GetSelectors()

		size := 0
		labelSelector := &metav1.LabelSelector{
			MatchLabels:      selectors.Selector,
			MatchExpressions: selectors.SelectorMatchExpressions,
		}

		// -- Labels
		sel, err := metav1.LabelSelectorAsSelector(labelSelector)
		if err != nil {
			return nil, fmt.Errorf(
				"selectorMatchExpressions or selectors of [%s/%s] are not valid: %w",
				target.GetObjectKind().GroupVersionKind().Kind,
				target.GetName(),
				err,
			)
		}
		if !sel.Matches(labels.Set(selectable.GetLabels())) {
			continue // Bail early!
		}

		size += len(labelSelector.MatchLabels)
		size += len(labelSelector.MatchExpressions)

		// -- Fields
		allFieldsMatched, err := matchesAllFields(selectable, selectors.SelectorMatchFields)
		if err != nil {
			// Todo: test in unit test
			return nil, fmt.Errorf(
				"failed to evaluate all matched fields of [%s/%s]: %w",
				target.GetObjectKind().GroupVersionKind().Kind,
				target.GetName(),
				err,
			)
		}
		if !allFieldsMatched {
			continue // Bail early!
		}
		size += len(selectors.SelectorMatchFields)

		// -- decision time
		if size > 0 {
			if matchingSelectors[size] == nil {
				matchingSelectors[size] = []SelectingObject{}
			}
			if size > highWaterMark {
				highWaterMark = size
			}
			matchingSelectors[size] = append(matchingSelectors[size], target)
		}
	}

	return matchingSelectors[highWaterMark], nil
}

func matchesAllFields(source Selectable, fields []v1alpha1.FieldSelectorRequirement) (bool, error) {
	for _, requirement := range fields {
		match, err := selector.Matches(requirement, source)
		if err != nil {
			return false, fmt.Errorf("unable to match field requirement with key [%s] operator [%s] values [%v]: %w", requirement.Key, requirement.Operator, requirement.Values, err)
		}
		if !match {
			return false, nil
		}
	}
	return true, nil
}
