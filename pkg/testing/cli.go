package testing

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
			if err != nil {
				os.Exit(3)
			}
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
		err := validateFilesExist([]string{directory, template})
		if err != nil {
			return fmt.Errorf("failed to validate files exist: %w", err)
		}

		cmd.SilenceUsage = true

		baseTestCase := TemplateTestCase{Given: TemplateTestGivens{TemplateFile: template}}
		testSuite, err := buildTestSuite(baseTestCase, directory)
		if err != nil {
			return fmt.Errorf("build test cases: %w", err)
		}

		cmd.SilenceErrors = true

		passedTests, failedTests := testSuite.Assert()

		return reportTestResults(passedTests, failedTests, testSuite.HasFocusedTests())
	},
}

func validateFilesExist(filelist []string) error {
	for _, file := range filelist {
		_, err := os.Stat(file)
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
