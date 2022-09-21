package template_tester

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
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

type templateType interface {
	ValidateCreate() error
	client.Object
}

type TemplateTestSuite map[string]*TemplateTestCase

func (s *TemplateTestSuite) Run(t *testing.T) {
	testsToRun := s.getTestsToRun()

	for name, testCase := range testsToRun {
		tc := testCase
		t.Run(name, func(t *testing.T) {
			tc.Run(t)
		})
	}
}

func (s *TemplateTestSuite) getTestsToRun() TemplateTestSuite {
	testsToRun := *s
	focusedCases := make(map[string]*TemplateTestCase, len(*s))

	for name, testCase := range *s {
		if testCase.Focus {
			focusedCases[name] = testCase
		}
	}

	if len(focusedCases) > 0 {
		testsToRun = focusedCases
	}
	return testsToRun
}

func (s *TemplateTestSuite) RunConcurrently(t *testing.T) {
	testsToRun := s.getTestsToRun()

	for name, testCase := range testsToRun {
		tc := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			tc.Run(t)
		})
	}
}

type TemplateTestCase struct {
	Inputs               TemplateTestInputs
	Expectations         TemplateTestExpectations
	IgnoreMetadata       bool
	IgnoreOwnerRefs      bool
	IgnoreLabels         bool
	IgnoreMetadataFields []string
	Focus                bool
}

type TemplateTestInputs struct {
	TemplateFile    string
	Template        templateType
	WorkloadFile    string
	Workload        *v1alpha1.Workload
	BlueprintParams []v1alpha1.BlueprintParam
	YttValues       Values
	YttFiles        []string
	labels          map[string]string
}

type TemplateTestExpectations struct {
	ExpectedObjectFile string
	ExpectedObject     client.Object
}

func (c *TemplateTestCase) Run(t *testing.T) {
	expectedObject, err := c.Expectations.getExpectedObject()
	if err != nil {
		t.Fatalf("failed to get expected object: %v", err)
	}

	actualObject, err := c.Inputs.getActualObject()
	if err != nil {
		t.Fatalf("failed to get actual object: %v", err)
	}

	c.stripIgnoredFields(expectedObject, actualObject)

	if diff := cmp.Diff(expectedObject.Object, actualObject.Object); diff != "" {
		t.Fatalf("expected does not equal actual: (-expected +actual):\n%s", diff)
	}
}

func (i *TemplateTestInputs) getActualObject() (*unstructured.Unstructured, error) {
	workload, err := i.getWorkload()
	if err != nil {
		return nil, fmt.Errorf("get workload failed: %v", err)
	}

	apiTemplate, err := i.getPopulatedTemplate()
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

	params := templates.ParamsBuilder(template.GetDefaultParams(), i.BlueprintParams, []v1alpha1.BlueprintParam{}, workload.Spec.Params)

	templatingContext := createTemplatingContext(*workload, params)

	stampContext := templates.StamperBuilder(workload, templatingContext, i.labels)
	ctx := context.TODO()
	actualStampedObject, err := stampContext.Stamp(ctx, template.GetResourceTemplate())
	if err != nil {
		return nil, fmt.Errorf("could not stamp: %v", err)
	}

	return actualStampedObject, nil
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

func (i *TemplateTestInputs) getWorkload() (*v1alpha1.Workload, error) {
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

func (i *TemplateTestInputs) completeLabels(workload v1alpha1.Workload, template templates.Template) {
	i.labels = map[string]string{}

	i.labels["carto.run/workload-name"] = workload.GetName()
	i.labels["carto.run/workload-namespace"] = workload.GetNamespace()
	i.labels["carto.run/template-kind"] = template.GetKind()
	i.labels["carto.run/cluster-template-name"] = template.GetName()
}

func (i *TemplateTestInputs) getPopulatedTemplate() (templateType, error) {
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
		templateFile, err = i.preprocessYtt()
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

func (e *TemplateTestExpectations) getExpectedObject() (*unstructured.Unstructured, error) {
	if (e.ExpectedObjectFile == "" && e.ExpectedObject == nil) ||
		(e.ExpectedObjectFile != "" && e.ExpectedObject != nil) {
		return nil, fmt.Errorf("exactly one of template or templateFile must be set")
	}

	if e.ExpectedObjectFile != "" {
		return e.getExpectedObjectFromFile()
	}

	unstruct, err := runtime.DefaultUnstructuredConverter.ToUnstructured(e.ExpectedObject)
	if err != nil {
		return nil, fmt.Errorf("failed to convert template to unstructured: %w", err)
	}

	return &unstructured.Unstructured{Object: unstruct}, nil
}

func (e *TemplateTestExpectations) getExpectedObjectFromFile() (*unstructured.Unstructured, error) {
	expectedStampedObjectYaml, err := os.ReadFile(e.ExpectedObjectFile)
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

func (i *TemplateTestInputs) preprocessYtt() (string, error) {
	yt := YTT()
	yt.Values(i.YttValues)
	yt.F(i.TemplateFile)
	for _, yttfile := range i.YttFiles {
		yt.F(yttfile)
	}
	f, err := yt.ToTempFile(context.TODO())
	if err != nil {
		return "", fmt.Errorf("failed to create temp fail by ytt: %w", err)
	}

	return f.Name(), nil
}

func createTemplatingContext(workload v1alpha1.Workload, params templates.Params) map[string]interface{} {
	sources := map[string]templates.SourceInput{}
	images := map[string]templates.ImageInput{}
	configs := map[string]templates.ConfigInput{}

	inputs := templates.Inputs{
		Sources: sources,
		Images:  images,
		Configs: configs,
	}

	templatingContext := map[string]interface{}{
		"workload": workload,
		"params":   params,
		"sources":  sources,
		"images":   images,
		"configs":  configs,
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
	return templatingContext
}

type StringParams []struct {
	Name         string
	Value        string
	DefaultValue string
}

func BuildBlueprintStringParams(candidateParams StringParams) ([]v1alpha1.BlueprintParam, error) {
	var completeParams []v1alpha1.BlueprintParam

	for _, stringParam := range candidateParams {
		newParam, err := BuildBlueprintStringParam(stringParam.Name, stringParam.Value, stringParam.DefaultValue)
		if err != nil {
			return nil, fmt.Errorf("failed to build param: %w", err)
		}
		completeParams = append(completeParams, *newParam)
	}

	return completeParams, nil
}

func BuildBlueprintStringParam(name string, value string, defaultValue string) (*v1alpha1.BlueprintParam, error) {
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
