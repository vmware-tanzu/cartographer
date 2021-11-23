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

type SelectorGetter interface {
	GetSelector() map[string]string
}

type LabelsGetter interface {
	GetLabels() map[string]string
}

// BestLabelMatch attempts at finding the targets that best match the label set
// of the source.
//
func BestLabelMatches(source LabelsGetter, targets []SelectorGetter) []SelectorGetter {
	if len(targets) == 0 {
		return nil
	}

	// count the number of matches
	matchCounter := make([]int, len(targets))
	for idx, target := range targets {
		if !subsetOf(source.GetLabels(), target.GetSelector()) {
			continue
		}

		for key, value := range target.GetSelector() {
			srcValue, found := source.GetLabels()[key]
			if !found || srcValue != value {
				continue
			}

			matchCounter[idx] += 1
		}
	}

	// keep just those that have the highest amount of matches
	var selectors []SelectorGetter
	if highestMatch := maxSlice(matchCounter); highestMatch > 0 {
		for idx := range matchCounter {
			if matchCounter[idx] == highestMatch {
				selectors = append(selectors, targets[idx])
			}
		}
	}

	// filter down to the most specific set
	selectorsCount := make([]int, len(selectors))
	for idx, selector := range selectors {
		selectorsCount[idx] = len(selector.GetSelector())
	}

	var res []SelectorGetter
	minSelectorCount := minSlice(selectorsCount)
	for _, selector := range selectors {
		if len(selector.GetSelector()) == minSelectorCount {
			res = append(res, selector)
		}
	}

	return res
}

// minSlice gets the minimum value in a given slice (or 999, otherwise)
//
func minSlice(slice []int) int {
	min := 999

	for idx, element := range slice {
		if idx == 0 || element < min {
			min = element
		}
	}

	return min
}

// maxSlice gets the maximum value in a given slice (or 0, otherwise)
//
func maxSlice(slice []int) int {
	max := 0

	for idx, element := range slice {
		if idx == 0 || element > max {
			max = element
		}
	}

	return max
}

// subsetOf verifies whether `a` is a subset of `b`
//
func subsetOf(a, b map[string]string) bool {
	for key, value := range b {
		if a[key] != value {
			return false
		}
	}

	return true
}
