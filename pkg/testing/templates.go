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
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/realizer"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

type templateType interface {
	ValidateCreate() error
	client.Object
}

// TemplateTestSuite is a collection of named template tests which may be run together
type TemplateTestSuite map[string]*TemplateTestCase

type FailedTest struct {
	name string
	err  error
}

// Assert allows testing a TemplateTestSuite when a *testing.T is not available,
// e.g. when tests are not run from 'go test'
// It returns a list of the named tests that passed and a list of the named tests that failed with their errors
func (s *TemplateTestSuite) Assert() ([]string, []*FailedTest) {
	var (
		passedTests []string
		failedTests []*FailedTest
	)

	testsToRun, _ := s.getTestsToRun()

	for name, testCase := range testsToRun {
		err := testCase.Run()
		if err != nil {
			failedTests = append(failedTests, &FailedTest{name: name, err: err})
		} else {
			passedTests = append(passedTests, name)
		}
	}

	return passedTests, failedTests
}

func (s *TemplateTestSuite) HasFocusedTests() bool {
	_, focused := s.getTestsToRun()
	return focused
}

func (s *TemplateTestSuite) Run(t *testing.T) {
	testsToRun, focused := s.getTestsToRun()

	if focused {
		defer t.Fatalf("test suite failed due to focused test, check individual test case status")
	}

	for name, testCase := range testsToRun {
		tc := testCase
		t.Run(name, func(t *testing.T) {
			err := tc.Run()
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func (s *TemplateTestSuite) RunConcurrently(t *testing.T) {
	testsToRun, focused := s.getTestsToRun()

	if focused {
		defer t.Fatalf("test suite failed due to focused test, check individual test case status")
	}

	for name, testCase := range testsToRun {
		tc := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := tc.Run()
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func (s *TemplateTestSuite) getTestsToRun() (TemplateTestSuite, bool) {
	focused := false
	testsToRun := *s
	focusedCases := make(map[string]*TemplateTestCase, len(*s))

	for name, testCase := range *s {
		if testCase.Focus {
			focusedCases[name] = testCase
		}
	}

	if len(focusedCases) > 0 {
		testsToRun = focusedCases
		focused = true
	}
	return testsToRun, focused
}

// TemplateTestCase is an individual template test.
// Given and Expect values must be provided.
// Fields in the expected object's metadata may be ignored
// When run as part of a TemplateTestSuite, an individual case(s) may be focused.
// This will exercise the individual test(s).
// Note that the overall suite will fail (preventing focused tests from passing CI).
type TemplateTestCase struct {
	Given                TemplateTestGivens
	Expect               TemplateTestExpectation
	IgnoreMetadata       bool
	IgnoreOwnerRefs      bool
	IgnoreLabels         bool
	IgnoreMetadataFields []string
	Focus                bool
}

func (c *TemplateTestCase) Run() error {
	expectedObject, err := c.Expect.getExpectedObject()
	if err != nil {
		return fmt.Errorf("failed to get expected object: %v", err)
	}

	actualObject, err := c.Given.getActualObject()
	if err != nil {
		return fmt.Errorf("failed to get actual object: %v", err)
	}

	c.stripIgnoredFields(expectedObject, actualObject)

	if diff := cmp.Diff(expectedObject.Object, actualObject.Object); diff != "" {
		return fmt.Errorf("expected does not equal actual: (-expected +actual):\n%s", diff)
	}

	return nil
}

func (c *TemplateTestCase) stripIgnoredFields(expected *unstructured.Unstructured, actual *unstructured.Unstructured) {
	delete(expected.Object, "status")
	delete(actual.Object, "status")

	if c.IgnoreLabels {
		expected.SetLabels(nil)
		actual.SetLabels(nil)
	}

	if c.IgnoreMetadata {
		delete(expected.Object, "metadata")
		delete(actual.Object, "metadata")
	}

	var expectedMetadata, actualMetadata map[string]interface{}

	if expected.Object["metadata"] != nil {
		expectedMetadata = expected.Object["metadata"].(map[string]interface{})
	}
	if actual.Object["metadata"] != nil {
		actualMetadata = actual.Object["metadata"].(map[string]interface{})
	}

	if c.IgnoreOwnerRefs {
		delete(expectedMetadata, "ownerReferences")
		delete(actualMetadata, "ownerReferences")
	}

	for _, field := range c.IgnoreMetadataFields {
		delete(expectedMetadata, field)
		delete(actualMetadata, field)
	}
}

// TemplateTestExpectation must provide the expected object as
// an object,
// an unstructured.Unstructured, or as
// a yaml file.
type TemplateTestExpectation struct {
	ExpectedFile         string
	ExpectedObject       client.Object
	ExpectedUnstructured *unstructured.Unstructured
}

func (e *TemplateTestExpectation) getExpectedObject() (*unstructured.Unstructured, error) {
	populatedFieldCount := 0
	if e.ExpectedFile != "" {
		populatedFieldCount++
	}
	if e.ExpectedObject != nil {
		populatedFieldCount++
	}
	if e.ExpectedUnstructured != nil {
		populatedFieldCount++
	}

	if populatedFieldCount != 1 {
		return nil, fmt.Errorf("exactly one of ExpectedFile, ExpectedObject or ExpectedUnstructured must be set")
	}

	if e.ExpectedUnstructured != nil {
		return e.ExpectedUnstructured, nil
	}

	if e.ExpectedFile != "" {
		return e.getExpectedObjectFromFile()
	}

	unstruct, err := runtime.DefaultUnstructuredConverter.ToUnstructured(e.ExpectedObject)
	if err != nil {
		return nil, fmt.Errorf("failed to convert template to unstructured: %w", err)
	}

	return &unstructured.Unstructured{Object: unstruct}, nil
}

func (e *TemplateTestExpectation) getExpectedObjectFromFile() (*unstructured.Unstructured, error) {
	expectedStampedObjectYaml, err := os.ReadFile(e.ExpectedFile)
	if err != nil {
		return nil, fmt.Errorf("could not read expected yaml: %w", err)
	}

	expectedJson, err := yaml.YAMLToJSON(expectedStampedObjectYaml)
	if err != nil {
		return nil, fmt.Errorf("convert yaml to json: %w", err)
	}

	expectedStampedObject := unstructured.Unstructured{}

	if err = expectedStampedObject.UnmarshalJSON(expectedJson); err != nil {
		return nil, fmt.Errorf("unmarshall json: %w", err)
	}

	return &expectedStampedObject, nil
}

// TemplateTestGivens must specify a template and a workload.
// These can be specified as yaml files or as objects.
// If the template is a yaml file, it may be pre-processed with ytt and values provided
// as objects or in a values yaml file.
// Any outputs expected from earlier templates in a supply chain may be provided in SupplyChainInputs.
// Params may be specified in the BlueprintParams
type TemplateTestGivens struct {
	TemplateFile          string
	Template              templateType
	WorkloadFile          string
	Workload              *v1alpha1.Workload
	BlueprintParams       []v1alpha1.BlueprintParam
	BlueprintParamsFile   string
	YttValues             Values
	YttFiles              []string
	labels                map[string]string
	SupplyChainInputs     *templates.Inputs
	SupplyChainInputsFile string
}

func (i *TemplateTestGivens) getActualObject() (*unstructured.Unstructured, error) {
	ctx := context.Background()

	workload, err := i.getWorkload()
	if err != nil {
		return nil, fmt.Errorf("get workload failed: %v", err)
	}

	apiTemplate, err := i.getPopulatedTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("get populated template failed: %v", err)
	}

	if err = apiTemplate.ValidateCreate(); err != nil {
		return nil, fmt.Errorf("template validation failed: %v", err)
	}

	template, err := templates.NewModelFromAPI(apiTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster template")
	}

	i.completeLabels(*workload, template)

	blueprintParams, err := i.getBlueprintParams()
	if err != nil {
		return nil, fmt.Errorf("get blueprint params failed: %v", err)
	}

	paramGenerator := realizer.NewParamGenerator([]v1alpha1.BlueprintParam{}, blueprintParams, workload.Spec.Params)
	params := paramGenerator.GetParams(template)

	templatingContext, err := i.createTemplatingContext(*workload, params)
	if err != nil {
		return nil, fmt.Errorf("create templating context: %w", err)
	}

	stampContext := templates.StamperBuilder(workload, templatingContext, i.labels)
	actualStampedObject, err := stampContext.Stamp(ctx, template.GetResourceTemplate())
	if err != nil {
		return nil, fmt.Errorf("could not stamp: %v", err)
	}

	return actualStampedObject, nil
}

func (i *TemplateTestGivens) getWorkload() (*v1alpha1.Workload, error) {
	if (i.Workload == nil && i.WorkloadFile == "") ||
		(i.Workload != nil && i.WorkloadFile != "") {
		return nil, fmt.Errorf("exactly one of Workload or WorkloadFile must be specified")
	}

	if i.Workload != nil {
		return i.Workload, nil
	}

	workload := &v1alpha1.Workload{}

	workloadData, err := os.ReadFile(i.WorkloadFile)
	if err != nil {
		return nil, fmt.Errorf("could not read workload file: %w", err)
	}

	if err = yaml.Unmarshal(workloadData, workload); err != nil {
		return nil, fmt.Errorf("unmarshall template: %w", err)
	}

	return workload, nil
}

func (i *TemplateTestGivens) getPopulatedTemplate(ctx context.Context) (templateType, error) {
	if (i.TemplateFile == "" && i.Template == nil) ||
		(i.TemplateFile != "" && i.Template != nil) {
		return nil, fmt.Errorf("exactly one of template or templateFile must be set")
	}

	if i.Template != nil {
		return i.Template, nil
	}

	var (
		templateFile string
		err          error
	)

	if len(i.YttValues) != 0 || len(i.YttFiles) != 0 {
		templateFile, err = i.preprocessYtt(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to preprocess ytt: %w", err)
		}
		defer os.RemoveAll(templateFile)
	} else {
		templateFile = i.TemplateFile
	}

	templateData, err := os.ReadFile(templateFile)

	if err != nil {
		return nil, fmt.Errorf("could not read template file: %w", err)
	}

	unknownTemplate := unstructured.Unstructured{}

	templateJson, err := yaml.YAMLToJSON(templateData)
	if err != nil {
		return nil, fmt.Errorf("convert yaml to json: %w", err)
	}

	if err = unknownTemplate.UnmarshalJSON(templateJson); err != nil {
		return nil, fmt.Errorf("unmarshall json: %w", err)
	}

	var apiTemplate templateType

	switch templateKind := unknownTemplate.GetKind(); templateKind {
	case "ClusterSourceTemplate":
		apiTemplate = &v1alpha1.ClusterSourceTemplate{}
	case "ClusterImageTemplate":
		apiTemplate = &v1alpha1.ClusterImageTemplate{}
	case "ClusterConfigTemplate":
		apiTemplate = &v1alpha1.ClusterConfigTemplate{}
	case "ClusterTemplate":
		apiTemplate = &v1alpha1.ClusterTemplate{}
	default:
		return nil, fmt.Errorf("template kind not found")
	}

	if err = yaml.Unmarshal(templateData, apiTemplate); err != nil {
		return nil, fmt.Errorf("unmarshall template: %w", err)
	}

	return apiTemplate, nil
}

func (i *TemplateTestGivens) preprocessYtt(ctx context.Context) (string, error) {
	yt := YTT()
	yt.Values(i.YttValues)
	yt.F(i.TemplateFile)
	for _, yttfile := range i.YttFiles {
		yt.F(yttfile)
	}
	f, err := yt.ToTempFile(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file by ytt: %w", err)
	}

	return f.Name(), nil
}

func (i *TemplateTestGivens) completeLabels(workload v1alpha1.Workload, template templates.Template) {
	i.labels = map[string]string{}

	i.labels["carto.run/workload-name"] = workload.GetName()
	i.labels["carto.run/workload-namespace"] = workload.GetNamespace()
	i.labels["carto.run/template-kind"] = template.GetKind()
	i.labels["carto.run/cluster-template-name"] = template.GetName()
}

func (i *TemplateTestGivens) createTemplatingContext(workload v1alpha1.Workload, params map[string]apiextensionsv1.JSON) (map[string]interface{}, error) {
	var inputs *templates.Inputs

	inputs, err := i.getSupplyChainInputs()
	if err != nil {
		return nil, fmt.Errorf("get supply chain inputs: %w", err)
	}

	templatingContext := map[string]interface{}{
		"workload": workload,
		"params":   params,
		"sources":  inputs.Sources,
		"images":   inputs.Images,
		"configs":  inputs.Configs,
		//"deployment": // not implemented yet,
	}

	if inputs.OnlyConfig() != nil {
		templatingContext["config"] = inputs.OnlyConfig()
	}
	if inputs.OnlyImage() != nil {
		templatingContext["image"] = inputs.OnlyImage()
	}
	if inputs.OnlySource() != nil {
		templatingContext["source"] = inputs.OnlySource()
	}
	return templatingContext, nil
}

func (i *TemplateTestGivens) getBlueprintParams() ([]v1alpha1.BlueprintParam, error) {
	if i.BlueprintParamsFile != "" && i.BlueprintParams != nil {
		return nil, fmt.Errorf("only one of blueprintParams or blueprintParamsFile may be set")
	}

	if i.BlueprintParamsFile == "" && i.BlueprintParams == nil {
		return []v1alpha1.BlueprintParam{}, nil
	}

	if i.BlueprintParams != nil {
		return i.BlueprintParams, nil
	}

	paramsFile, err := os.ReadFile(i.BlueprintParamsFile)
	if err != nil {
		return nil, fmt.Errorf("could not read blueprintParamsFile: %w", err)
	}

	var paramsData []v1alpha1.BlueprintParam

	err = yaml.Unmarshal(paramsFile, &paramsData)
	if err != nil {
		return nil, fmt.Errorf("unmarshall params: %w", err)
	}

	return paramsData, nil // TODO: document
}

func (i *TemplateTestGivens) getSupplyChainInputs() (*templates.Inputs, error) {
	if i.SupplyChainInputsFile != "" && i.SupplyChainInputs != nil {
		return nil, fmt.Errorf("only one of blueprintParams or blueprintParamsFile may be set")
	}

	if i.SupplyChainInputsFile == "" && i.SupplyChainInputs == nil {
		return &templates.Inputs{}, nil
	}

	if i.SupplyChainInputs != nil {
		return i.SupplyChainInputs, nil
	}

	inputsFile, err := os.ReadFile(i.SupplyChainInputsFile)
	if err != nil {
		return nil, fmt.Errorf("could not read blueprintParamsFile: %w", err)
	}

	inputData := make(map[string]interface{})

	err = yaml.Unmarshal(inputsFile, &inputData)
	if err != nil {
		return nil, fmt.Errorf("unmarshall params: %w", err)
	}

	var inputs templates.Inputs

	if val, ok := inputData["sources"]; ok {
		switch sources := val.(type) {
		case map[string]interface{}:
			inputs.Sources, err = extractSources(sources)
			if err != nil {
				return nil, fmt.Errorf("extract sources: %w", err)
			}
		default:
			return nil, fmt.Errorf("type assertion failed, expected \"sources\" to be type map[string]interface{} but found: %v", val)
		}
	}

	if val, ok := inputData["images"]; ok {
		switch images := val.(type) {
		case map[string]interface{}:
			inputs.Images, err = extractImages(images)
			if err != nil {
				return nil, fmt.Errorf("extract images: %w", err)
			}
		default:
			return nil, fmt.Errorf("type assertion failed, expected \"images\" to be type map[string]interface{} but found: %v", val)
		}
	}

	if val, ok := inputData["configs"]; ok {
		switch configs := val.(type) {
		case map[string]interface{}:
			inputs.Configs, err = extractConfigs(configs)
			if err != nil {
				return nil, fmt.Errorf("extract configs: %w", err)
			}
		default:
			return nil, fmt.Errorf("type assertion failed, expected \"configs\" to be type map[string]interface{} but found: %v", val)
		}
	}

	return &inputs, nil
}

func extractSources(sources map[string]interface{}) (map[string]templates.SourceInput, error) {
	inputSources := make(map[string]templates.SourceInput)
	for name, v := range sources {
		switch source := v.(type) {
		case map[string]interface{}:
			inputSource := templates.SourceInput{}
			inputSource.Name = name
			if url, ok := source["url"]; ok {
				inputSource.URL = url
			}
			if revision, ok := source["revision"]; ok {
				inputSource.Revision = revision
			}
			inputSources[name] = inputSource
		default:
			return nil, fmt.Errorf("type assertion failed: expected sources[\"%s\"] to be map[string]interface{}, but found: %v", name, source)
		}
	}
	return inputSources, nil
}

func extractImages(images map[string]interface{}) (map[string]templates.ImageInput, error) {
	inputImages := make(map[string]templates.ImageInput)
	for name, v := range images {
		switch image := v.(type) {
		case map[string]interface{}:
			inputImage := templates.ImageInput{}
			inputImage.Name = name
			if i, ok := image["image"]; ok {
				inputImage.Image = i
			}
			inputImages[name] = inputImage
		default:
			return nil, fmt.Errorf("type assertion failed: expected images[\"%s\"] to be map[string]interface{}, but found: %v", name, image)
		}
	}
	return inputImages, nil
}

func extractConfigs(configs map[string]interface{}) (map[string]templates.ConfigInput, error) {
	inputConfigs := make(map[string]templates.ConfigInput)
	for name, v := range configs {
		switch config := v.(type) {
		case map[string]interface{}:
			inputConfig := templates.ConfigInput{}
			inputConfig.Name = name
			if c, ok := config["config"]; ok {
				inputConfig.Config = c
			}
			inputConfigs[name] = inputConfig
		default:
			return nil, fmt.Errorf("type assertion failed: expected configs[\"%s\"] to be map[string]interface{}, but found: %v", name, config)
		}
	}
	return inputConfigs, nil
}

// StringParam is a helper struct for use with the BuildBlueprintStringParams method
// Either a Value or a DefaultValue should be specified for every StringParam
// A Name is required for every StringParam
type StringParam struct {
	Name         string
	Value        string
	DefaultValue string
}

// BuildBlueprintStringParams is a helper method for creating string BlueprintParams for Givens.
// BlueprintParams that hold other valid JSON values must be constructed by hand.
func BuildBlueprintStringParams(candidateParams []StringParam) ([]v1alpha1.BlueprintParam, error) {
	var completeParams []v1alpha1.BlueprintParam

	for _, stringParam := range candidateParams {
		newParam, err := buildBlueprintStringParam(stringParam.Name, stringParam.Value, stringParam.DefaultValue)
		if err != nil {
			return nil, fmt.Errorf("failed to build param: %w", err)
		}
		completeParams = append(completeParams, *newParam)
	}

	return completeParams, nil
}

func buildBlueprintStringParam(name string, value string, defaultValue string) (*v1alpha1.BlueprintParam, error) {
	if (value == "" && defaultValue == "") ||
		value != "" && defaultValue != "" {
		return nil, fmt.Errorf("exactly one of value or defaultValue must be set")
	}

	if name == "" {
		return nil, fmt.Errorf("name must be set")
	}

	param := v1alpha1.BlueprintParam{
		Name: name,
	}

	if value != "" {
		param.Value = &apiextensionsv1.JSON{Raw: []byte(fmt.Sprintf("%#v", value))}
	} else {
		param.DefaultValue = &apiextensionsv1.JSON{Raw: []byte(fmt.Sprintf("%#v", defaultValue))}
	}

	return &param, nil
}
