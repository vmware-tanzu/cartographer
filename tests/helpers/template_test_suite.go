package helpers

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

type templateType interface {
	ValidateCreate() error
	client.Object
}

type TemplateTestSuite struct {
	TemplateFile       string
	ExpectedObjectFile string
	WorkloadFile       string
	Workload           *v1alpha1.Workload
	BlueprintParams    []v1alpha1.BlueprintParam
	Labels             map[string]string
	IgnoreMetadata     bool
	IgnoreOwnerRefs    bool
	IgnoreLabels       bool
}

func (ts *TemplateTestSuite) Run(t *testing.T) {
	if err := ts.verifySuite(); err != nil {
		t.Fatalf("TemplateTestSuite invalid: %v", err)
	}

	workload, err := ts.getWorkload()
	if err != nil {
		t.Fatalf("get workload failed: %v", err)
	}

	apiTemplate, err := getPopulatedTemplate(ts.TemplateFile)
	if err != nil {
		t.Fatalf("get populated template failed: %v", err)
	}

	if err = apiTemplate.ValidateCreate(); err != nil {
		t.Fatalf("template validation failed: %v", err)
	}

	template, err := templates.NewModelFromAPI(apiTemplate)
	if err != nil {
		t.Fatalf("failed to get cluster template")
	}

	completeLabels(ts.Labels, *workload, template)

	params := templates.ParamsBuilder(template.GetDefaultParams(), ts.BlueprintParams, []v1alpha1.BlueprintParam{}, workload.Spec.Params)

	templatingContext := createTemplatingContext(*workload, params)

	stampContext := templates.StamperBuilder(workload, templatingContext, ts.Labels)
	ctx := context.TODO()
	actualStampedObject, err := stampContext.Stamp(ctx, template.GetResourceTemplate())
	if err != nil {
		t.Fatalf("could not stamp: %v", err)
	}

	expectedStampedObjectYaml, err := os.ReadFile(ts.ExpectedObjectFile)
	if err != nil {
		t.Fatalf("could not read expected yaml: %v", err)
	}

	expectedStampedObject := getExpectedObject(t, err, expectedStampedObjectYaml)

	stripIgnoredFields(ts, expectedStampedObject, actualStampedObject)

	if diff := cmp.Diff(expectedStampedObject.Object, actualStampedObject.Object); diff != "" {
		t.Fatalf("expected does not equal actual: (-expected +actual):\n%s", diff)
	}
}

func stripIgnoredFields(ts *TemplateTestSuite, expected unstructured.Unstructured, actual *unstructured.Unstructured) {
	if ts.IgnoreMetadata {
		expected.Object["metadata"] = nil
		actual.Object["metadata"] = nil
	}

	if ts.IgnoreOwnerRefs {
		if expected.Object["metadata"] != nil {
			metadata := expected.Object["metadata"].(map[string]interface{})
			metadata["ownerReferences"] = nil
		}
		if actual.Object["metadata"] != nil {
			metadata := actual.Object["metadata"].(map[string]interface{})
			metadata["ownerReferences"] = nil
		}
	}

	if ts.IgnoreLabels {
		expected.SetLabels(nil)
		actual.SetLabels(nil)
	}
}

func (ts *TemplateTestSuite) verifySuite() error {
	if ts.Workload == nil && ts.WorkloadFile == "" {
		return fmt.Errorf("exactly one of Workload or WorkloadFile must be specified")
	}

	if ts.Workload != nil && ts.WorkloadFile != "" {
		return fmt.Errorf("exactly one of Workload or WorkloadFile must be specified")
	}

	if ts.Labels == nil {
		ts.Labels = map[string]string{}
	}

	if ts.BlueprintParams == nil {
		ts.BlueprintParams = []v1alpha1.BlueprintParam{}
	}

	return nil
}

func (ts *TemplateTestSuite) getWorkload() (*v1alpha1.Workload, error) {
	if ts.Workload != nil {
		return ts.Workload, nil
	}

	workload := &v1alpha1.Workload{}

	workloadData, err := os.ReadFile(ts.WorkloadFile)
	if err != nil {
		return nil, fmt.Errorf("could not read workload file: %v", err)
	}

	if err = yaml.Unmarshal(workloadData, workload); err != nil {
		return nil, fmt.Errorf("unmarshall template: %v", err)
	}

	return workload, nil
}

func getExpectedObject(t *testing.T, err error, expectedStampedObjectYaml []byte) unstructured.Unstructured {
	expectedJson, err := yaml.YAMLToJSON(expectedStampedObjectYaml)
	if err != nil {
		t.Fatalf("convert yaml to json: %v", err)
	}

	expectedStampedObject := unstructured.Unstructured{}

	if err = expectedStampedObject.UnmarshalJSON(expectedJson); err != nil {
		t.Fatalf("unmarshall json: %v", err)
	}
	return expectedStampedObject
}

func completeLabels(labels map[string]string, workload v1alpha1.Workload, template templates.Template) {
	labels["carto.run/workload-name"] = workload.GetName()
	labels["carto.run/workload-namespace"] = workload.GetNamespace()
	labels["carto.run/template-kind"] = template.GetKind()
	labels["carto.run/cluster-template-name"] = template.GetName()
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

func getPopulatedTemplate(templateFile string) (templateType, error) {
	templateData, err := os.ReadFile(templateFile)

	if err != nil {
		return nil, fmt.Errorf("could not read template file: %v", err)
	}

	unknownTemplate := unstructured.Unstructured{}

	templateJson, err := yaml.YAMLToJSON(templateData)
	if err != nil {
		return nil, fmt.Errorf("convert yaml to json: %v", err)
	}

	if err = unknownTemplate.UnmarshalJSON(templateJson); err != nil {
		return nil, fmt.Errorf("unmarshall json: %v", err)
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
		return nil, fmt.Errorf("unmarshall template: %v", err)
	}

	return apiTemplate, nil
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
