package testing

import (
	"context"
	"fmt"
	"github.com/vmware-tanzu/cartographer/pkg/realizer"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os"

	"sigs.k8s.io/yaml"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

func (i *TemplateTestGivens) mockedBlueprintStamp(ctx context.Context, workload *v1alpha1.Workload, apiTemplate ValidatableTemplate, template templates.Reader) (*unstructured.Unstructured, error) {
	i.completeLabels(*workload, apiTemplate.GetName(), apiTemplate.GetObjectKind().GroupVersionKind().Kind)

	var (
		blueprintParams []v1alpha1.BlueprintParam
		err             error
	)

	if i.BlueprintParams != nil {
		blueprintParams, err = i.BlueprintParams.GetBlueprintParams()
		if err != nil {
			return nil, fmt.Errorf("get blueprint params failed: %w", err)
		}
	}

	paramMerger := realizer.NewParamMerger([]v1alpha1.BlueprintParam{}, blueprintParams, workload.Spec.Params)
	params := paramMerger.Merge(template)

	templatingContext, err := i.createTemplatingContext(*workload, params)
	if err != nil {
		return nil, fmt.Errorf("create templating context: %w", err)
	}

	stampContext := templates.StamperBuilder(workload, templatingContext, i.labels)
	actualStampedObject, err := stampContext.Stamp(ctx, template.GetResourceTemplate())
	if err != nil {
		return nil, fmt.Errorf("could not stamp: %w", err)
	}

	return actualStampedObject, nil
}

func (i *TemplateTestGivens) completeLabels(workload v1alpha1.Workload, name string, kind string) {
	i.labels = map[string]string{}

	i.labels["carto.run/workload-name"] = workload.GetName()
	i.labels["carto.run/workload-namespace"] = workload.GetNamespace()
	i.labels["carto.run/template-kind"] = kind
	i.labels["carto.run/cluster-template-name"] = name
}

type BlueprintParams interface {
	GetBlueprintParams() ([]v1alpha1.BlueprintParam, error)
}

type BlueprintParamsObject struct {
	BlueprintParams []v1alpha1.BlueprintParam
}

func (p *BlueprintParamsObject) GetBlueprintParams() ([]v1alpha1.BlueprintParam, error) {
	return p.BlueprintParams, nil
}

type BlueprintParamsFile struct {
	Path string
}

func (p *BlueprintParamsFile) GetBlueprintParams() ([]v1alpha1.BlueprintParam, error) {
	paramsFile, err := os.ReadFile(p.Path)
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
