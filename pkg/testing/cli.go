package testing

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.SetFormatter(&log.TextFormatter{})
		if verbose {
			log.SetLevel(log.DebugLevel)
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := validateFilesExist()
		if err != nil {
			return fmt.Errorf("failed to validate all files exist: %w", err)
		}

		var testSuite TemplateTestSuite
		testSuite, err = buildTestCases(TemplateTestCase{Given: TemplateTestGivens{TemplateFile: template}}, directory)
		if err != nil {
			return fmt.Errorf("build test cases: %w", err)
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

func buildTestCases(testCase TemplateTestCase, directory string) (TemplateTestSuite, error) {
	var err error

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
			tempTestSuite, err = buildTestCases(newCase, filepath.Join(directory, subdirectory))
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
