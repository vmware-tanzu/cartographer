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

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

type testCaseReporter struct {
	err  error
	path string
	info *testInfo
}

func reportTestResults(passedTests []string, failedTests []*FailedTest, hasFocusedTests bool) error {
	var (
		tests         []testCaseReporter
		errorOccurred bool
		reportString  string
	)

	for _, passedTest := range passedTests {
		testCase, err := newTestCaseReporter(passedTest, nil)
		if err != nil {
			return fmt.Errorf("failed to create new test case reporter for test %s: %w", passedTest, err)
		}
		tests = append(tests, *testCase)
	}

	for _, failedTest := range failedTests {
		testCase, err := newTestCaseReporter(failedTest.name, failedTest.err)
		if err != nil {
			return fmt.Errorf("failed to create new test case reporter for test %s: %w", failedTest.name, err)
		}
		tests = append(tests, *testCase)
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

func newTestCaseReporter(path string, err error) (*testCaseReporter, error) {
	tcr := testCaseReporter{
		err:  err,
		path: path,
	}

	info, err := populateInfo(path)
	if err != nil {
		return nil, fmt.Errorf("populate test info: %w", err)
	}

	tcr.info = info

	return &tcr, nil
}

func (t *testCaseReporter) report(verbose bool) string {
	if t.err == nil {
		return t.reportPassed(verbose)
	}
	return t.reportFailed(verbose)
}

func (t *testCaseReporter) reportPassed(verbose bool) string {
	if verbose && t.info.Metadata.Name != nil {
		return fmt.Sprintf("PASS: %s %s", t.path, *t.info.Metadata.Name)
	}
	return fmt.Sprintf("PASS: %s", t.path)
}

func (t *testCaseReporter) reportFailed(verbose bool) string {
	returnString := fmt.Sprintf("FAIL: %s", t.path)

	if verbose && t.info.Metadata.Name != nil {
		returnString = fmt.Sprintf("%s\nName: %s", returnString, *t.info.Metadata.Name)
	}

	if verbose && t.info.Metadata.Description != nil {
		returnString = fmt.Sprintf("%s\nDescription: %s", returnString, *t.info.Metadata.Description)
	}

	if verbose && t.err != nil {
		returnString = fmt.Sprintf("%s\nError: %s", returnString, t.err.Error())
	}

	return returnString
}

const infoFilename = "info.yaml"

func populateInfo(directory string) (*testInfo, error) {
	var infoStruct testInfo

	filePath := filepath.Join(directory, infoFilename)
	infoFile, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			log.Debugf("populate info failed, did not find %s", filePath)
			return &infoStruct, nil
		}
		return nil, fmt.Errorf("read file failed: %s: %w", filePath, err)
	}

	err = yaml.Unmarshal(infoFile, &infoStruct)
	if err != nil {
		return nil, fmt.Errorf("unmarshal failed: %s: %w", filePath, err)
	}

	return &infoStruct, nil
}
