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
	"strings"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/realizer"
)

type testInfo struct {
	Metadata       testInfoMetadata       `yaml:"metadata"`
	Given          testInfoGiven          `yaml:"given"`
	Expected       *string                `yaml:"expected"`
	Focus          *bool                  `yaml:"focus"`
	CompareOptions testInfoCompareOptions `yaml:"compareOptions"`
}

type testInfoMetadata struct {
	Name        *string `yaml:"name"`
	Description *string `yaml:"description"`
}

type testInfoCompareOptions struct {
	IgnoreMetadata       *bool    `yaml:"ignoreMetadata"`
	IgnoreOwnerRefs      *bool    `yaml:"ignoreOwnerRefs"`
	IgnoreLabels         *bool    `yaml:"ignoreLabels"`
	IgnoreMetadataFields []string `yaml:"ignoreMetadataFields"`
	NamedCMPOptionFuncs  []string `yaml:"namedCMPOptionFuncs"`
}

type testInfoGiven struct {
	Template        testInfoTemplate    `yaml:"template"`
	Workload        *string             `yaml:"workload"`
	MockSupplyChain testInfoMockSC      `yaml:"mockSupplyChain"`
	SupplyChain     testInfoSupplyChain `yaml:"supplyChain"`
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
	Paths              []string          `yaml:"paths"`
	YttPath            *string           `yaml:"yttPath"`
	TargetResourceName *string           `yaml:"targetResourceName"`
	PreviousOutputs    *realizer.Outputs `yaml:"previousOutputs"`
}

const (
	templateDefaultFilename             = "template.yaml"
	workloadDefaultFilename             = "workload.yaml"
	expectedDefaultFilename             = "expected.yaml"
	templateYttValuesDefaultFilename    = "template-ytt-values.yaml"
	supplyChainYttValuesDefaultFilename = "supply-chain-ytt-values.yaml"
)

func populateTestCase(testCase *Test, directory string) (*Test, error) {
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

	newExpectedFilePath, err := getLocallySpecifiedPath(directory, expectedDefaultFilename, info.Expected)
	if err != nil {
		return nil, fmt.Errorf("get expected file specified in directory %s: %w", directory, err)
	}
	if newExpectedFilePath != "" {
		testCase.Expect = &ExpectedFile{Path: newExpectedFilePath}
	}

	if info.Focus != nil {
		testCase.Focus = *info.Focus
	}

	testCase = populateCompareOptions(testCase, info)

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

func populateCompareOptions(testCase *Test, info *testInfo) *Test {
	if info.CompareOptions.IgnoreMetadata != nil {
		if testCase.CompareOptions == nil {
			testCase.CompareOptions = &CompareOptions{}
		}
		testCase.CompareOptions.IgnoreMetadata = *info.CompareOptions.IgnoreMetadata
	}
	if info.CompareOptions.IgnoreOwnerRefs != nil {
		if testCase.CompareOptions == nil {
			testCase.CompareOptions = &CompareOptions{}
		}
		testCase.CompareOptions.IgnoreOwnerRefs = *info.CompareOptions.IgnoreOwnerRefs
	}
	if info.CompareOptions.IgnoreLabels != nil {
		if testCase.CompareOptions == nil {
			testCase.CompareOptions = &CompareOptions{}
		}
		testCase.CompareOptions.IgnoreLabels = *info.CompareOptions.IgnoreLabels
	}
	if info.CompareOptions.IgnoreMetadataFields != nil {
		if testCase.CompareOptions == nil {
			testCase.CompareOptions = &CompareOptions{}
		}
		testCase.CompareOptions.IgnoreMetadataFields = info.CompareOptions.IgnoreMetadataFields
	}

	for _, optionFunc := range info.CompareOptions.NamedCMPOptionFuncs {
		if optionFunc == "ConvertNumbersToFloatsDuringComparison" {
			if testCase.CompareOptions == nil {
				testCase.CompareOptions = &CompareOptions{}
			}
			testCase.CompareOptions.CMPOption = ConvertNumbersToFloatsDuringComparison
		}
	}

	return testCase
}

func populateTestCaseWorkload(testCase *Test, directory string, info *testInfo) (*Test, error) {
	newWorkloadValue, err := getLocallySpecifiedPath(directory, workloadDefaultFilename, info.Given.Workload)
	if err != nil {
		return nil, fmt.Errorf("get workload file specified in directory %s: %w", directory, err)
	}
	if newWorkloadValue != "" {
		testCase.Given.Workload = &WorkloadFile{Path: newWorkloadValue}
	}
	return testCase, nil
}

func populateTestCaseTemplate(testCase *Test, directory string, info *testInfo) (*Test, error) {
	newTemplateFile := TemplateFile{}

	if previousTemplateFile, prevTemplateFileExisted := testCase.Given.Template.(*TemplateFile); prevTemplateFileExisted {
		newTemplateFile = *previousTemplateFile
	}

	newTemplateFilepath, err := getLocallySpecifiedPath(directory, templateDefaultFilename, info.Given.Template.Path)
	if err != nil {
		return nil, fmt.Errorf("get template file specified in directory %s: %w", directory, err)
	}
	if newTemplateFilepath != "" {
		newTemplateFile.Path = newTemplateFilepath
	}

	yttFile, err := getLocallySpecifiedPath(directory, templateYttValuesDefaultFilename, info.Given.Template.YttPath)
	if err != nil {
		return nil, fmt.Errorf("get template ytt file specified in directory %s: %w", directory, err)
	}
	if yttFile != "" {
		newTemplateFile.YttFiles = []string{yttFile}
	}

	testCase.Given.Template = &newTemplateFile

	return testCase, nil
}

func populateTestCaseMockSupplyChain(testCase *Test, info *testInfo) (*Test, bool) {
	var mockSupplyChainSpecified bool
	mockSupplyChain := MockSupplyChain{}

	if info.Given.MockSupplyChain.BlueprintInputs != nil {
		mockSupplyChain.Inputs = &SupplyChainInputsObject{Inputs: info.Given.MockSupplyChain.BlueprintInputs}
		mockSupplyChainSpecified = true
	}

	if info.Given.MockSupplyChain.BlueprintParams != nil {
		mockSupplyChain.Params = &SupplyChainParamsObject{Params: info.Given.MockSupplyChain.BlueprintParams}
		mockSupplyChainSpecified = true
	}

	testCase.Given.SupplyChain = &mockSupplyChain

	return testCase, mockSupplyChainSpecified
}

func populateTestCaseSupplyChain(testCase *Test, directory string, info *testInfo) (*Test, bool, error) {
	var supplyChainSpecified bool

	newSupplyChain := SupplyChainFileSet{}

	if previousSCSet, prevSCSetExisted := testCase.Given.SupplyChain.(*SupplyChainFileSet); prevSCSetExisted {
		newSupplyChain = *previousSCSet
	}

	if info.Given.SupplyChain.TargetResourceName != nil {
		newSupplyChain.TargetResourceName = *info.Given.SupplyChain.TargetResourceName
	}

	if info.Given.SupplyChain.PreviousOutputs != nil {
		newSupplyChain.PreviousOutputs = info.Given.SupplyChain.PreviousOutputs
	}

	yttFile, err := getLocallySpecifiedPath(directory, supplyChainYttValuesDefaultFilename, info.Given.SupplyChain.YttPath)
	if err != nil {
		return nil, false, fmt.Errorf("get supply chain ytt file specified in directory %s: %w", directory, err)
	}
	if yttFile != "" {
		newSupplyChain.YttFiles = []string{yttFile}
	}

	if len(info.Given.SupplyChain.Paths) > 0 {
		newSupplyChain.Paths = info.Given.SupplyChain.Paths
		supplyChainSpecified = true
	} else {
		filesPrefixedSupplyChainInDir, err := getFilesPrefixedSupplyChainInDir(directory)
		if err != nil {
			return nil, false, fmt.Errorf("get files prefixed with supply-chain in dir: %w", err)
		}
		if len(filesPrefixedSupplyChainInDir) > 0 {
			newSupplyChain.Paths = filesPrefixedSupplyChainInDir
			supplyChainSpecified = true
		}
	}

	if fileSetNotEmpty(&newSupplyChain) {
		testCase.Given.SupplyChain = &newSupplyChain
	}

	return testCase, supplyChainSpecified, nil
}

// getLocallySpecifiedPath returns the path specified in info.yaml, or if none specified there, the default
// filename (if the file exists)
func getLocallySpecifiedPath(directory, filename string, priorityPath *string) (string, error) {
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

func getFilesPrefixedSupplyChainInDir(directory string) ([]string, error) {
	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("os read directory: %w", err)
	}

	var supplyChainYamlFilesFound []string

	for _, file := range files {
		if strings.HasPrefix(file.Name(), "supply-chain") && strings.HasSuffix(file.Name(), ".yaml") {
			supplyChainYamlFilesFound = append(supplyChainYamlFilesFound, filepath.Join(directory, file.Name()))
		}
	}

	return supplyChainYamlFilesFound, nil
}

func fileSetNotEmpty(supplyChainfileSet *SupplyChainFileSet) bool {
	return len(supplyChainfileSet.YttFiles) > 0 ||
		len(supplyChainfileSet.Paths) > 0 ||
		supplyChainfileSet.TargetResourceName != "" ||
		supplyChainfileSet.PreviousOutputs != nil
}
