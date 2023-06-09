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
