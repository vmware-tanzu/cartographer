package testing

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"sigs.k8s.io/yaml"
)

type testCaseReporter struct {
	err                     error
	path, name, description string
}

func reportTestResults(passedTests []string, failedTests []*FailedTest, hasFocusedTests bool) error {
	var (
		tests         []testCaseReporter
		errorOccurred bool
		reportString  string
	)

	for _, passedTest := range passedTests {
		tests = append(tests, *newTestCaseReporter(passedTest, nil))
	}

	for _, failedTest := range failedTests {
		tests = append(tests, *newTestCaseReporter(failedTest.name, failedTest.err))
		errorOccurred = true
	}

	for _, test := range tests {
		reportString = fmt.Sprintf("%s\n%s", reportString, test.report(verbose))
	}

	if hasFocusedTests {
		reportString = fmt.Sprintf("%s\n\ntest suite failed due to focused test, check individual test case status", reportString)
		errorOccurred = true
	}

	if errorOccurred {
		reportString = fmt.Sprintf("%s\nFAIL", reportString)
	} else {
		reportString = fmt.Sprintf("%s\nPASS", reportString)
	}

	_, err := fmt.Fprintf(os.Stderr, "%s\n", reportString)
	if err != nil {
		return fmt.Errorf("write to stdErr failed")
	}

	if errorOccurred {
		return TestFailError{}
	}

	return nil
}

func newTestCaseReporter(path string, err error) *testCaseReporter {
	tcr := testCaseReporter{
		err:  err,
		path: path,
	}
	tcr.name, _ = getTestInfo(path, "name")
	tcr.description, _ = getTestInfo(path, "description")

	return &tcr
}

func (t *testCaseReporter) report(verbose bool) string {
	if t.err == nil {
		return t.reportPassed(verbose)
	}
	return t.reportFailed(verbose)
}

func (t *testCaseReporter) reportPassed(verbose bool) string {
	if verbose && t.name != "" {
		return fmt.Sprintf("PASS: %s %s", t.path, t.name)
	}
	return fmt.Sprintf("PASS: %s", t.path)
}

func (t *testCaseReporter) reportFailed(verbose bool) string {
	returnString := fmt.Sprintf("FAIL: %s", t.path)

	if verbose && t.name != "" {
		returnString = fmt.Sprintf("%s\nName: %s", returnString, t.name)
	}

	if verbose && t.description != "" {
		returnString = fmt.Sprintf("%s\nDescription: %s", returnString, t.description)
	}

	if verbose && t.err != nil {
		returnString = fmt.Sprintf("%s\nError: %s", returnString, t.err.Error())
	}

	return returnString
}

func getTestInfo(path string, field string) (string, error) {
	infoFilepath := filepath.Join(path, "info.yaml")

	infoFile, err := os.ReadFile(infoFilepath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("file does not exist: %s", infoFilepath)
		}
		return "", fmt.Errorf("unable to read file: %s", infoFilepath)
	}

	infoData := make(map[string]interface{})

	err = yaml.Unmarshal(infoFile, &infoData)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal %s", infoFilepath)
	}

	if val, ok := infoData[field]; ok {
		if stringVal, ok := val.(string); ok {
			return stringVal, nil
		}
	}

	return "", fmt.Errorf("field not found: %s", field)
}
