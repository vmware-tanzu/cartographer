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
	version   = "development"
	directory string
	verbose   bool
)

var rootCmd = &cobra.Command{
	Use:     "cartotest",
	Version: version,
	Short:   "cartotest - test Cartographer files",
	Long: `cartotest is a CLI to verify the output of your Cartographer files
This version tests only templates with the 'templates' command
Read more at cartographer.sh`,
	Args: cobra.NoArgs,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.SetFormatter(&log.TextFormatter{})
		if verbose {
			log.SetLevel(log.DebugLevel)
		}
	},
}

var templateCmd = &cobra.Command{
	Use:     "templates",
	Version: version,
	Short:   "templates tests directories of templates",
	Long: `the templates command allows assertion that a given template will create an expected object.
Users can mock the templating context that would be available from the supply chain.
Read more at cartographer.sh`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true

		return CliTest(directory)
	},
}

func CliTest(directory string) error {
	baseTestCase := TemplateTestCase{}
	testSuite, err := buildTestSuite(&baseTestCase, directory)
	if err != nil {
		return fmt.Errorf("build test cases: %w", err)
	}

	passedTests, failedTests := testSuite.Assert()

	return reportTestResults(passedTests, failedTests, testSuite.HasFocusedTests())
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "output logs and increase test failure verbosity")

	rootCmd.AddCommand(templateCmd)
	templateCmd.Flags().StringVarP(&directory, "directory", "d", "", "directory to test")

	_ = templateCmd.MarkFlagRequired("directory")
	_ = templateCmd.MarkFlagDirname("directory")
}
