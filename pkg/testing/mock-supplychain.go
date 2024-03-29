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

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/realizer"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

// MockSupplyChain implements SupplyChain
// SupplyChainInputs simulate expected inputs that are the outputs from earlier resources in the supply chain
// SupplyChainParams supplies params as if defined in the supply chain
type MockSupplyChain struct {
	Params SupplyChainParams
	Inputs SupplyChainInputs
}

func (i *MockSupplyChain) stamp(ctx context.Context, workload *v1alpha1.Workload, apiTemplate ValidatableTemplate, template templates.Reader) (*unstructured.Unstructured, error) {
	labels := completeLabels(*workload, apiTemplate.GetName(), apiTemplate.GetObjectKind().GroupVersionKind().Kind)

	var (
		err error
	)

	blueprintParams := make([]v1alpha1.BlueprintParam, 0)

	if i.Params != nil {
		blueprintParams, err = i.Params.GetParams()
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

	stampContext := templates.StamperBuilder(workload, templatingContext, labels)
	actualStampedObject, err := stampContext.Stamp(ctx, template.GetResourceTemplate())
	if err != nil {
		return nil, fmt.Errorf("could not stamp: %w", err)
	}

	return actualStampedObject, nil
}

func completeLabels(workload v1alpha1.Workload, name string, kind string) map[string]string {
	labels := make(map[string]string)

	labels["carto.run/workload-name"] = workload.GetName()
	labels["carto.run/workload-namespace"] = workload.GetNamespace()
	labels["carto.run/template-kind"] = kind
	labels["carto.run/cluster-template-name"] = name

	return labels
}

func (i *MockSupplyChain) createTemplatingContext(workload v1alpha1.Workload, params map[string]apiextensionsv1.JSON) (map[string]interface{}, error) {
	var (
		inputs *Inputs
		err    error
	)

	inputs = &Inputs{}

	if i.Inputs != nil {
		inputs, err = i.Inputs.GetInputs()
		if err != nil {
			return nil, fmt.Errorf("get supply chain inputs: %w", err)
		}
	}

	templatingContext := map[string]interface{}{
		"workload": workload,
		"params":   params,
		"sources":  inputs.Sources,
		"images":   inputs.Images,
		"configs":  inputs.Configs,
		//"deployment": // not implemented yet,
	}

	if len(inputs.Sources) == 1 {
		for _, source := range inputs.Sources {
			templatingContext["source"] = &source
		}
	}

	if len(inputs.Images) == 1 {
		for _, image := range inputs.Images {
			templatingContext["image"] = image.Image
		}
	}

	if len(inputs.Configs) == 1 {
		for _, config := range inputs.Configs {
			templatingContext["config"] = config.Config
		}
	}
	return templatingContext, nil
}

type SupplyChainParams interface {
	GetParams() ([]v1alpha1.BlueprintParam, error)
}

type SupplyChainParamsObject struct {
	Params []v1alpha1.BlueprintParam
}

func (p *SupplyChainParamsObject) GetParams() ([]v1alpha1.BlueprintParam, error) {
	return p.Params, nil
}

type SupplyChainParamsFile struct {
	Path string
}

func (p *SupplyChainParamsFile) GetParams() ([]v1alpha1.BlueprintParam, error) {
	paramsFile, err := os.ReadFile(p.Path)
	if err != nil {
		return nil, fmt.Errorf("could not read supplyChainParamsFile: %w", err)
	}

	var paramsData []v1alpha1.BlueprintParam

	err = yaml.Unmarshal(paramsFile, &paramsData)
	if err != nil {
		return nil, fmt.Errorf("unmarshall params: %w", err)
	}

	return paramsData, nil // TODO: document
}

// StringParam is a helper struct for use with the BuildSupplyChainStringParams method
// Either a Value or a DefaultValue should be specified for every StringParam
// A Name is required for every StringParam
type StringParam struct {
	Name         string
	Value        string
	DefaultValue string
}

// BuildSupplyChainStringParams is a helper method for creating string SupplyChainParams.
// SupplyChainParams that hold other valid JSON values must be constructed by hand.
func BuildSupplyChainStringParams(candidateParams []StringParam) ([]v1alpha1.BlueprintParam, error) {
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

type SupplyChainInputs interface {
	GetInputs() (*Inputs, error)
}

type SupplyChainInputsObject struct {
	Inputs *Inputs
}

func (i *SupplyChainInputsObject) GetInputs() (*Inputs, error) {
	return i.Inputs, nil
}

type SupplyChainInputsFile struct {
	Path string
}

func (p *SupplyChainInputsFile) GetInputs() (*Inputs, error) {
	inputsFile, err := os.ReadFile(p.Path)
	if err != nil {
		return nil, fmt.Errorf("could not read supplyChainInputsFile %s: %w", p.Path, err)
	}

	var inputs Inputs

	err = yaml.Unmarshal(inputsFile, &inputs)
	if err != nil {
		return nil, fmt.Errorf("unmarshall params: %w", err)
	}

	return &inputs, nil
}
