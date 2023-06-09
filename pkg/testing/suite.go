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

package testing

import "testing"

// TemplateTestSuite is a collection of named template tests which may be run together
type TemplateTestSuite map[string]*TemplateTestCase

type FailedTest struct {
	name string
	err  error
}

// Assert allows testing a TemplateTestSuite when a *testing.T is not available,
// e.g. when tests are not run from 'go test'
// It returns a list of the named tests that passed and a list of the named tests that failed with their errors
func (s *TemplateTestSuite) Assert() ([]string, []*FailedTest) {
	var (
		passedTests []string
		failedTests []*FailedTest
	)

	testsToRun, _ := s.getTestsToRun()

	for name, testCase := range testsToRun {
		err := testCase.Run()
		if err != nil {
			failedTests = append(failedTests, &FailedTest{name: name, err: err})
		} else {
			passedTests = append(passedTests, name)
		}
	}

	return passedTests, failedTests
}

func (s *TemplateTestSuite) HasFocusedTests() bool {
	_, focused := s.getTestsToRun()
	return focused
}

func (s *TemplateTestSuite) Run(t *testing.T) {
	testsToRun, focused := s.getTestsToRun()

	if focused {
		defer t.Fatalf("test suite failed due to focused test, check individual test case status")
	}

	for name, testCase := range testsToRun {
		tc := testCase
		t.Run(name, func(t *testing.T) {
			err := tc.Run()
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func (s *TemplateTestSuite) RunConcurrently(t *testing.T) {
	testsToRun, focused := s.getTestsToRun()

	if focused {
		defer t.Fatalf("test suite failed due to focused test, check individual test case status")
	}

	for name, testCase := range testsToRun {
		tc := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := tc.Run()
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func (s *TemplateTestSuite) getTestsToRun() (TemplateTestSuite, bool) {
	focused := false
	testsToRun := *s
	focusedCases := make(map[string]*TemplateTestCase, len(*s))

	for name, testCase := range *s {
		if testCase.Focus {
			focusedCases[name] = testCase
		}
	}

	if len(focusedCases) > 0 {
		testsToRun = focusedCases
		focused = true
	}
	return testsToRun, focused
}
