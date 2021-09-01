package matchers

import (
	"strings"

	"github.com/onsi/gomega/types"

	"fmt"
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
