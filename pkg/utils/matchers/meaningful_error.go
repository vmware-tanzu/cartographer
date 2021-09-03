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

package matchers

import (
	"fmt"
	"strings"

	"github.com/onsi/gomega/types"
)

func BeMeaningful(expected string) types.GomegaMatcher {
	return &beMeaningfulMatcher{
		expected: expected,
	}
}

type beMeaningfulMatcher struct {
	expected string
}

func (matcher *beMeaningfulMatcher) Match(actual interface{}) (success bool, err error) {
	if actual == nil {
		return true, nil
	}

	actualError, ok := actual.(error)
	if !ok {
		return false, fmt.Errorf("BeMeaningful matcher expects an error")
	}

	return strings.Contains(actualError.Error(), matcher.expected), nil
}

func (matcher *beMeaningfulMatcher) FailureMessage(actual interface{}) (message string) {
	if actual == nil {
		return "Expected error to have occurred"
	} else {
		return fmt.Sprintf("Expected\n\t%#v\nto contain \n\t%#v", actual.(error).Error(), matcher.expected)
	}
}

func (matcher *beMeaningfulMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	if actual == nil {
		return "Expected error not to have occurred"
	} else {
		return fmt.Sprintf("Expected\n\t%#v\n not to contain \n\t%#v", actual.(error).Error(), matcher.expected)
	}
}
