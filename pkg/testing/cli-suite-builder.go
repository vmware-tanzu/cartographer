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

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

type testInfo struct {
	Name                 *string                   `yaml:"name"`
	Description          *string                   `yaml:"description"`
	Template             *string                   `yaml:"template"`
	Workload             *string                   `yaml:"workload"`
	Expected             *string                   `yaml:"expected"`
	Ytt                  *string                   `yaml:"ytt"`
	SupplyChainInputs    *templates.Inputs         `yaml:"supplyChainInputs"`
	BlueprintParams      []v1alpha1.BlueprintParam `yaml:"blueprintParams"`
	Focus                *bool                     `yaml:"focus"`
	IgnoreMetadata       *bool                     `yaml:"ignoreMetadata"`
	IgnoreOwnerRefs      *bool                     `yaml:"ignoreOwnerRefs"`
	IgnoreLabels         *bool                     `yaml:"ignoreLabels"`
	IgnoreMetadataFields []string                  `yaml:"ignoreMetadataFields"`
}

const (
	templateDefaultFilename  = "template.yaml"
	workloadDefaultFilename  = "workload.yaml"
	expectedDefaultFilename  = "expected.yaml"
	yttValuesDefaultFilename = "ytt-values.yaml"
)

func buildTestSuite(testCase TemplateTestCase, directory string) (TemplateTestSuite, error) {
	info, err := populateInfo(directory)
	if err != nil {
		return nil, fmt.Errorf("populate info: %w", err)
	}

	testCase.Given.TemplateFile = replaceIfFound(testCase.Given.TemplateFile, directory, templateDefaultFilename, info.Template)
	testCase.Given.WorkloadFile = replaceIfFound(testCase.Given.WorkloadFile, directory, workloadDefaultFilename, info.Workload)
	testCase.Expect.ExpectedFile = replaceIfFound(testCase.Expect.ExpectedFile, directory, expectedDefaultFilename, info.Expected)

	yttFile := ""
	if testCase.Given.YttFiles != nil {
		yttFile = testCase.Given.YttFiles[0]
	}
	yttFile = replaceIfFound(yttFile, directory, yttValuesDefaultFilename, info.Ytt)
	if yttFile != "" {
		testCase.Given.YttFiles = []string{yttFile}
	}

	if info.Focus != nil {
		testCase.Focus = *info.Focus
	}
	if info.IgnoreMetadata != nil {
		testCase.IgnoreMetadata = *info.IgnoreMetadata
	}
	if info.IgnoreOwnerRefs != nil {
		testCase.IgnoreOwnerRefs = *info.IgnoreOwnerRefs
	}
	if info.IgnoreLabels != nil {
		testCase.IgnoreLabels = *info.IgnoreLabels
	}
	if info.IgnoreMetadataFields != nil {
		testCase.IgnoreMetadataFields = info.IgnoreMetadataFields
	}

	if info.SupplyChainInputs != nil {
		testCase.Given.SupplyChainInputs = info.SupplyChainInputs
	}

	if info.BlueprintParams != nil {
		testCase.Given.BlueprintParams = info.BlueprintParams
	}

	subdirectories, err := getSubdirectories(directory)
	if err != nil {
		return nil, fmt.Errorf("get subdirectories: %w", err)
	}

	// recurse
	if len(subdirectories) > 0 {
		testSuite := make(TemplateTestSuite)
		for _, subdirectory := range subdirectories {
			newCase := testCase
			var tempTestSuite TemplateTestSuite
			tempTestSuite, err = buildTestSuite(newCase, filepath.Join(directory, subdirectory))
			if err != nil {
				return nil, fmt.Errorf("failed building test case for subdirectory: %s: %w", subdirectory, err)
			}
			for name, aCase := range tempTestSuite {
				testSuite[name] = aCase
			}
		}
		return testSuite, nil
	}

	return TemplateTestSuite{
		directory: &testCase,
	}, nil
}

func getSubdirectories(directory string) ([]string, error) {
	var subdirectories []string
	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			subdirectories = append(subdirectories, file.Name())
		}
	}
	return subdirectories, nil
}

func replaceIfFound(originalPath string, directory, filename string, priorityPath *string) string {
	if priorityPath != nil {
		return filepath.Join(directory, *priorityPath)
	}
	candidatePath := filepath.Join(directory, filename)
	_, err := os.Stat(candidatePath)
	if errors.Is(err, fs.ErrNotExist) {
		return originalPath
	}
	if originalPath != "" {
		log.Debugf("%s replaced a value found in a parent directory", candidatePath)
	}
	return candidatePath
}
