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

	newExpectedFilePath, err := getLocallySpecifiedPath(directory, expectedDefaultFilename, info.Expected)
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
	newWorkloadValue, err := getLocallySpecifiedPath(directory, workloadDefaultFilename, info.Given.Workload)
	if err != nil {
		return nil, fmt.Errorf("replace workload file in directory %s: %w", directory, err)
	}
	if newWorkloadValue != "" {
		testCase.Given.Workload = &WorkloadFile{Path: newWorkloadValue}
	}
	return testCase, nil
}

func populateTestCaseTemplate(testCase *TemplateTestCase, directory string, info *testInfo) (*TemplateTestCase, error) {
	newTemplateFilepath, err := getLocallySpecifiedPath(directory, templateDefaultFilename, info.Given.Template.Path)
	if err != nil {
		return nil, fmt.Errorf("replace template file in directory %s: %w", directory, err)
	}

	yttFile, err := getLocallySpecifiedPath(directory, yttValuesDefaultFilename, info.Given.Template.YttPath)
	if err != nil {
		return nil, fmt.Errorf("replace workload file in directory %s: %w", directory, err)
	}

	newTemplateFile := TemplateFile{}

	if previousTemplateFile, prevTemplateFileExisted := testCase.Given.Template.(*TemplateFile); prevTemplateFileExisted {
		newTemplateFile = *previousTemplateFile
	}

	if newTemplateFilepath != "" {
		newTemplateFile.Path = newTemplateFilepath
	}

	if yttFile != "" {
		newTemplateFile.YttFiles = []string{yttFile}
	}

	testCase.Given.Template = &newTemplateFile

	return testCase, nil
}

func populateTestCaseMockSupplyChain(testCase *TemplateTestCase, info *testInfo) (*TemplateTestCase, bool) {
	var mockSupplyChainSpecified bool
	mockSupplyChain := MockSupplyChain{}

	if info.Given.MockSupplyChain.BlueprintInputs != nil {
		mockSupplyChain.BlueprintInputs = &BlueprintInputsObject{BlueprintInputs: info.Given.MockSupplyChain.BlueprintInputs}
		mockSupplyChainSpecified = true
	}

	if info.Given.MockSupplyChain.BlueprintParams != nil {
		mockSupplyChain.BlueprintParams = &BlueprintParamsObject{BlueprintParams: info.Given.MockSupplyChain.BlueprintParams}
		mockSupplyChainSpecified = true
	}

	testCase.Given.SupplyChain = &mockSupplyChain

	return testCase, mockSupplyChainSpecified
}

func populateTestCaseSupplyChain(testCase *TemplateTestCase, directory string, info *testInfo) (*TemplateTestCase, bool, error) {
	panic("not implemented")
	//return testCase, false, nil
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
