package testing

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

type TestFailError struct {
	err error
}

func (t TestFailError) Error() string {
	if t.err != nil {
		return t.err.Error()
	}
	return ""
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		switch err.(type) {
		case TestFailError:
		default:
			_, err = fmt.Fprintf(os.Stderr, "error while executing: %s\n", err.Error())
		}

		if err != nil {
			os.Exit(3)
		}
		os.Exit(1)
	}
}

var (
	version                      = "0.5.3" // TODO remove hard coding
	directory, template, yttFile string
	verbose                      bool
)

var rootCmd = &cobra.Command{
	Use:     "cartotest",
	Version: version,
	Short:   "cartotest - test Cartographer files",
	Long: `cartotest is a CLI to verify the output of your Cartographer files,
   
Read more at cartographer.sh`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := validateFilesExist()
		if err != nil {
			return fmt.Errorf("failed to validate all files exist: %w", err)
		}

		//var testCases []TemplateTestCase
		//testCases = buildTestCases(TemplateTestCase{},directory)

		testCase := TemplateTestCase{
			Given: TemplateTestGivens{
				TemplateFile: template,
				WorkloadFile: filepath.Join(directory, "workload.yaml"),
			},
			Expect: TemplateTestExpectation{
				ExpectedFile: filepath.Join(directory, "expected.yaml"),
			},
		}

		yttFilepath := filepath.Join(directory, "ytt-values.yaml")
		_, err = os.Stat(yttFilepath)
		if !errors.Is(err, fs.ErrNotExist) {
			testCase.Given.YttFiles = []string{yttFilepath}
		}

		inputsFilepath := filepath.Join(directory, "inputs.yaml")
		_, err = os.Stat(inputsFilepath)
		if !errors.Is(err, fs.ErrNotExist) {
			testCase.Given.SupplyChainInputsFile = inputsFilepath
		}

		paramsFilepath := filepath.Join(directory, "params.yaml")
		_, err = os.Stat(paramsFilepath)
		if !errors.Is(err, fs.ErrNotExist) {
			testCase.Given.BlueprintParamsFile = paramsFilepath
		}

		focusFilepath := filepath.Join(directory, "focus")
		_, err = os.Stat(focusFilepath)
		if !errors.Is(err, fs.ErrNotExist) {
			testCase.Focus = true
		}

		ignoreMetadataFilepath := filepath.Join(directory, "ignoreMetadata")
		_, err = os.Stat(ignoreMetadataFilepath)
		if !errors.Is(err, fs.ErrNotExist) {
			testCase.IgnoreMetadata = true
		}

		ignoreOwnerRefsFilepath := filepath.Join(directory, "ignoreOwnerRefs")
		_, err = os.Stat(ignoreOwnerRefsFilepath)
		if !errors.Is(err, fs.ErrNotExist) {
			testCase.IgnoreOwnerRefs = true
		}

		ignoreLabelsFilepath := filepath.Join(directory, "ignoreLabels")
		_, err = os.Stat(ignoreLabelsFilepath)
		if !errors.Is(err, fs.ErrNotExist) {
			testCase.IgnoreLabels = true
		}

		testCase.IgnoreMetadataFields, err = getIgnoredFields()
		if err != nil {
			return fmt.Errorf("get ignored fields: %w", err)
		}

		testSuite := TemplateTestSuite{
			directory: &testCase,
		}

		cmd.SilenceUsage = true
		cmd.SilenceErrors = true

		passedTests, failedTests := testSuite.Assert()

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

		if testSuite.HasFocusedTests() {
			reportString = fmt.Sprintf("%s\n\ntest suite failed due to focused test, check individual test case status", reportString)
			errorOccurred = true
		}

		if errorOccurred {
			reportString = fmt.Sprintf("%s\nFAIL", reportString)
		} else {
			reportString = fmt.Sprintf("%s\nPASS", reportString)
		}

		_, err = fmt.Fprintf(os.Stderr, "%s\n", reportString)
		if err != nil {
			return fmt.Errorf("write to stdErr failed")
		}

		if errorOccurred {
			return TestFailError{}
		}

		return nil
	},
}

type testCaseReporter struct {
	err                     error
	path, name, description string
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

func getIgnoredFields() ([]string, error) {
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

func validateFilesExist() error {
	var (
		err      error
		fileList []string
	)

	fileList = append(fileList, directory, template)
	fileList = append(fileList, filepath.Join(directory, "workload.yaml"))
	fileList = append(fileList, filepath.Join(directory, "expected.yaml"))
	if yttFile != "" {
		fileList = append(fileList, yttFile)
	}

	for _, file := range fileList {
		_, err = os.Stat(file)
		if os.IsNotExist(err) {
			return err
		}
	}

	return nil
}

func init() {
	rootCmd.Flags().StringVarP(&directory, "directory", "d", "", "directory with subdirectories of workloads and expected outcome files")
	rootCmd.Flags().StringVarP(&template, "template", "t", "", "template file to test")
	rootCmd.Flags().StringVarP(&yttFile, "ytt", "", "", "file to ytt apply with template")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "increase test verbosity")

	_ = rootCmd.MarkFlagRequired("directory")
	_ = rootCmd.MarkFlagRequired("template")
}
