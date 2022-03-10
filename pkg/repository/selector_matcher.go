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
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/selector"
)

type SelectingObject interface {
	GetSelectors() v1alpha1.LegacySelector
	GetObjectKind() schema.ObjectKind
	GetName() string
}

func LegacySelectorToSelector(legacySelector v1alpha1.LegacySelector) v1alpha1.Selector {
	return v1alpha1.Selector{
		LabelSelector: metav1.LabelSelector{
			MatchLabels:      legacySelector.Selector,
			MatchExpressions: legacySelector.SelectorMatchExpressions,
		},
		MatchFields: legacySelector.SelectorMatchFields,
	}
}

func selectingObjectsSelectors(selectingObjects []SelectingObject) []v1alpha1.Selector {
	selectors := make([]v1alpha1.Selector, len(selectingObjects))
	for idx, selectingObject := range selectingObjects {
		selectors[idx] = LegacySelectorToSelector(selectingObject.GetSelectors())
	}
	return selectors
}

// BestSelectorMatch attempts at finding the selectors that best match their selectors
// against the selectors.
func BestSelectorMatch(selectable selector.Selectable, selectingObjects []SelectingObject) ([]SelectingObject, error) {
	bestMatchingSelectingObjectIndices, err := selector.BestSelectorMatchIndices(selectable, selectingObjectsSelectors(selectingObjects))

	if err != nil {
		target := selectingObjects[err.SelectorIndex()]
		return nil, fmt.Errorf(
			"error handling selectors, selectorMatchExpressions or selectorMatchFields of [%s/%s]: %w",
			target.GetObjectKind().GroupVersionKind().Kind,
			target.GetName(),
			err,
		)
	}

	numMatches := len(bestMatchingSelectingObjectIndices)

	if numMatches == 0 {
		return nil, nil
	}

	matches := make([]SelectingObject, numMatches)
	for idx, selectingObjectIdx := range bestMatchingSelectingObjectIndices {
		matches[idx] = selectingObjects[selectingObjectIdx]
	}
	return matches, nil
}
