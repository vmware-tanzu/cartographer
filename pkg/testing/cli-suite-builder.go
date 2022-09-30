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

func buildTestSuite(testCase TemplateTestCase, directory string) (TemplateTestSuite, error) {
	testCase.Given.TemplateFile = replaceIfFound(testCase.Given.WorkloadFile, directory, "template.yaml")
	testCase.Given.WorkloadFile = replaceIfFound(testCase.Given.WorkloadFile, directory, "workload.yaml")
	testCase.Expect.ExpectedFile = replaceIfFound(testCase.Expect.ExpectedFile, directory, "expected.yaml")
	testCase.Given.SupplyChainInputsFile = replaceIfFound(testCase.Given.SupplyChainInputsFile, directory, "inputs.yaml")
	testCase.Given.BlueprintParamsFile = replaceIfFound(testCase.Given.BlueprintParamsFile, directory, "params.yaml")
	testCase.Given.YttFiles = replaceYttIfFound(testCase.Given.YttFiles)

	testCase.Focus = testCase.Focus || trueIfFound(directory, "focus")
	testCase.IgnoreMetadata = testCase.IgnoreMetadata || trueIfFound(directory, "ignoreMetadata")
	testCase.IgnoreOwnerRefs = testCase.IgnoreOwnerRefs || trueIfFound(directory, "ignoreOwnerRefs")
	testCase.IgnoreLabels = testCase.IgnoreLabels || trueIfFound(directory, "ignoreLabels")

	newIgnoreFields, err := getIgnoredFields(directory)
	if err != nil {
		return nil, fmt.Errorf("get ignored fields: %w", err)
	}

	testCase.IgnoreMetadataFields = append(testCase.IgnoreMetadataFields, newIgnoreFields...)

	subdirectories, err := getSubdirectories(directory)
	if err != nil {
		return nil, fmt.Errorf("get subdirectories: %w", err)
	}

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

func trueIfFound(directory string, file string) bool {
	filePath := filepath.Join(directory, file)
	_, err := os.Stat(filePath)
	if !errors.Is(err, fs.ErrNotExist) {
		log.Debugf("set %s for directory %s and its subdirectories", file, directory)
		return true
	}
	return false
}

func replaceYttIfFound(yttFiles []string) []string {
	yttFilepath := filepath.Join(directory, "ytt-values.yaml")
	_, err := os.Stat(yttFilepath)
	if !errors.Is(err, fs.ErrNotExist) {
		if len(yttFiles) > 0 {
			log.Debugf("%s concatenated with yttfiles in a parent directory", yttFilepath)
			return append(yttFiles, yttFilepath)
		} else {
			return []string{yttFilepath}
		}
	}
	return yttFiles
}

func replaceIfFound(originalPath string, directory, filename string) string {
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

func getIgnoredFields(directory string) ([]string, error) {
	var (
		ignoredFields []string
		err           error
	)

	ignoreFieldsFilepath := filepath.Join(directory, "ignore.yaml")
	ignoreFieldsFile, err := os.ReadFile(ignoreFieldsFilepath)
	if err != nil {
		return ignoredFields, nil
	}

	err = yaml.Unmarshal(ignoreFieldsFile, &ignoredFields)
	if err != nil {
		return nil, fmt.Errorf("unmarshall ignore.yaml: %w", err)
	}

	return ignoredFields, nil
}
