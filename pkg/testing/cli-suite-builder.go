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

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/realizer"
)

type testInfo struct {
	metadata             testInfoMetadata `yaml:"metadata"`
	Given                testInfoGiven    `yaml:"given"`
	Expected             *string          `yaml:"expected"`
	Focus                *bool            `yaml:"focus"`
	IgnoreMetadata       *bool            `yaml:"ignoreMetadata"`
	IgnoreOwnerRefs      *bool            `yaml:"ignoreOwnerRefs"`
	IgnoreLabels         *bool            `yaml:"ignoreLabels"`
	IgnoreMetadataFields []string         `yaml:"ignoreMetadataFields"`
}

type testInfoMetadata struct {
	Name        *string `yaml:"name"`
	Description *string `yaml:"description"`
}

type testInfoGiven struct {
	Template        testInfoTemplate    `yaml:"template"`
	Workload        *string             `yaml:"workload"`
	MockSupplyChain testInfoMockSC      `yaml:"mockSupplyChain"`
	SupplyChain     testInfoSupplyChain `yaml:"SupplyChain"`
}

type testInfoTemplate struct {
	Path    *string `yaml:"path"`
	YttPath *string `yaml:"yttPath"`
}

type testInfoMockSC struct {
	BlueprintInputs *Inputs                   `yaml:"blueprintInputs"`
	BlueprintParams []v1alpha1.BlueprintParam `yaml:"blueprintParams"`
}

type testInfoSupplyChain struct {
	Path               *string           `yaml:"path"`
	YttPath            *string           `yaml:"yttPath"`
	TargetResourceName *string           `yaml:"targetResourceName"`
	PreviousOutputs    *realizer.Outputs `yaml:"previousOutputs"`
}

const (
	templateDefaultFilename  = "template.yaml"
	workloadDefaultFilename  = "workload.yaml"
	expectedDefaultFilename  = "expected.yaml"
	yttValuesDefaultFilename = "ytt-values.yaml"
)

func buildTestSuite(testCase *TemplateTestCase, directory string) (TemplateTestSuite, error) {
	var err error

	testCase, err = populateTestCase(testCase, directory)
	if err != nil {
		return nil, fmt.Errorf("populate testCase: %w", err)
	}

	subdirectories, err := getSubdirectories(directory)
	if err != nil {
		return nil, fmt.Errorf("get subdirectories: %w", err)
	}

	// recurse
	if len(subdirectories) > 0 {
		testSuite := make(TemplateTestSuite)
		for _, subdirectory := range subdirectories {
			newCase := *testCase
			var tempTestSuite TemplateTestSuite
			tempTestSuite, err = buildTestSuite(&newCase, filepath.Join(directory, subdirectory))
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
		directory: testCase,
	}, nil
}

func populateTestCase(testCase *TemplateTestCase, directory string) (*TemplateTestCase, error) {
	info, err := populateInfo(directory)
	if err != nil {
		return nil, fmt.Errorf("populate info: %w", err)
	}

	testCase, err = populateTestCaseTemplate(testCase, directory, info)
	if err != nil {
		return nil, fmt.Errorf("populate testCase template: %w", err)
	}

	testCase, err = populateTestCaseWorkload(testCase, directory, info)
	if err != nil {
		return nil, fmt.Errorf("populate testCase workload: %w", err)
	}

	newExpectedFilePath, err := replaceIfFound(directory, expectedDefaultFilename, info.Expected)
	if err != nil {
		return nil, fmt.Errorf("replace expected file in directory %s: %w", directory, err)
	}
	if newExpectedFilePath != "" {
		testCase.Expect = &ExpectedFile{Path: newExpectedFilePath}
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

	var (
		mockSupplyChainSpecified bool
		supplyChainSpecified     bool
	)

	testCase, mockSupplyChainSpecified = populateTestCaseMockSupplyChain(testCase, info)
	testCase, supplyChainSpecified, err = populateTestCaseSupplyChain(testCase, directory, info)
	if err != nil {
		return nil, fmt.Errorf("populate testCase supplychain: %w", err)
	}

	if mockSupplyChainSpecified && supplyChainSpecified {
		return nil, fmt.Errorf("only one of mock supply chain and real supply chain may be specified")
	}

	return testCase, nil
}

func populateTestCaseWorkload(testCase *TemplateTestCase, directory string, info *testInfo) (*TemplateTestCase, error) {
	newWorkloadValue, err := replaceIfFound(directory, workloadDefaultFilename, info.Given.Workload)
	if err != nil {
		return nil, fmt.Errorf("replace workload file in directory %s: %w", directory, err)
	}
	if newWorkloadValue != "" {
		testCase.Given.Workload = &WorkloadFile{Path: newWorkloadValue}
	}
	return testCase, nil
}

func populateTestCaseTemplate(testCase *TemplateTestCase, directory string, info *testInfo) (*TemplateTestCase, error) {
	newTemplateValue, err := replaceIfFound(directory, templateDefaultFilename, info.Given.Template.Path)
	if err != nil {
		return nil, fmt.Errorf("replace template file in directory %s: %w", directory, err)
	}
	if newTemplateValue != "" {
		var previousYttFile []string
		previousTemplateFile, ok := testCase.Given.Template.(*TemplateFile)
		if ok {
			previousYttFile = previousTemplateFile.YttFiles
		}

		newTemplateFile := TemplateFile{Path: newTemplateValue}

		newYTTFile, err := replaceIfFound(directory, yttValuesDefaultFilename, info.Given.Template.YttPath)
		if err != nil {
			return nil, fmt.Errorf("replace workload file in directory %s: %w", directory, err)
		}
		if newYTTFile != "" {
			newTemplateFile.YttFiles = []string{newYTTFile}
		} else {
			newTemplateFile.YttFiles = previousYttFile
		}

		testCase.Given.Template = &newTemplateFile
	}
	return testCase, nil
}

func populateTestCaseMockSupplyChain(testCase *TemplateTestCase, info *testInfo) (*TemplateTestCase, bool) {
	var isMockSupplyChainSpecified bool
	mockSupplyChain := MockSupplyChain{}

	if info.Given.MockSupplyChain.BlueprintInputs != nil {
		mockSupplyChain.BlueprintInputs = &BlueprintInputsObject{BlueprintInputs: info.Given.MockSupplyChain.BlueprintInputs}
		isMockSupplyChainSpecified = true
	}

	if info.Given.MockSupplyChain.BlueprintParams != nil {
		mockSupplyChain.BlueprintParams = &BlueprintParamsObject{BlueprintParams: info.Given.MockSupplyChain.BlueprintParams}
		isMockSupplyChainSpecified = true
	}

	testCase.Given.SupplyChain = &mockSupplyChain

	return testCase, isMockSupplyChainSpecified
}

func populateTestCaseSupplyChain(testCase *TemplateTestCase, directory string, info *testInfo) (*TemplateTestCase, bool, error) {
	panic("not implemented")
	//return testCase, false, nil
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

func replaceIfFound(directory, filename string, priorityPath *string) (string, error) {
	if priorityPath != nil {
		return filepath.Join(directory, *priorityPath), nil
	}
	candidatePath := filepath.Join(directory, filename)
	_, err := os.Stat(candidatePath)
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("failed while getting file info on %s: %w", candidatePath, err)
	}
	return candidatePath, nil
}
