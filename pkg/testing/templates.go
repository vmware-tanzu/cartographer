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
	"errors"
	"fmt"
	"os"
	"os/exec"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/realizer"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

type Inputs struct {
	Sources    map[string]templates.SourceInput
	Images     map[string]templates.ImageInput
	Configs    map[string]templates.ConfigInput
	Deployment *templates.SourceInput
}

type ValidatableTemplate interface {
	ValidateCreate() error
	client.Object
}

type FailedTest struct {
	name string
	err  error
}

// TemplateTestGivens must specify a template and a workload.
// These can be specified as yaml files or as objects.
// If the template is a yaml file, it may be pre-processed with ytt and values provided
// as objects or in a values yaml file.
// Any outputs expected from earlier templates in a supply chain may be provided in BlueprintInputs.
// Params may be specified in the BlueprintParams
type TemplateTestGivens struct {
	Template            Template
	Workload            Workload
	BlueprintParams     BlueprintParams
	labels              map[string]string
	BlueprintInputs     *Inputs
	BlueprintInputsFile string
	SupplyChain         SupplyChain
	TargetResource      TargetResource
	TTOutputs           TTOutputs
}

type Template interface {
	GetTemplate() (*ValidatableTemplate, error)
}

type TemplateObject struct {
	Template ValidatableTemplate
}

func (t *TemplateObject) GetTemplate() (*ValidatableTemplate, error) {
	return &t.Template, nil
}

type TemplateFile struct {
	Path      string
	YttValues Values
	YttFiles  []string
}

func (i *TemplateFile) GetTemplate() (*ValidatableTemplate, error) {
	var (
		templateFile string
		err          error
	)
	ctx := context.TODO()

	if len(i.YttValues) != 0 || len(i.YttFiles) != 0 {
		err = ensureYTTAvailable(ctx)

		if err != nil {
			return nil, fmt.Errorf("ensure ytt available: %w", err)
		}

		templateFile, err = i.preprocessYtt(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to preprocess ytt: %w", err)
		}
		defer os.RemoveAll(templateFile)
	} else {
		templateFile = i.Path
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

	var apiTemplate ValidatableTemplate

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

	return &apiTemplate, nil
}

type TargetResource interface {
	GetTargetResourceName() (string, error)
}

type TTOutputs interface {
	GetOutputs() (realizer.Outputs, error)
}

var yttNotFound = errors.New("ytt must be installed in PATH but was not found")

func ensureYTTAvailable(ctx context.Context) error {
	yttTestArgs := []string{"ytt", "--version"}
	_, _, err := Cmd(yttTestArgs...).RunWithOutput(ctx)
	if errors.Is(err, exec.ErrNotFound) {
		return yttNotFound
	} else if err != nil {
		return fmt.Errorf("run ytt test args: %w", err)
	}

	return nil
}

func (i *TemplateFile) preprocessYtt(ctx context.Context) (string, error) {
	yt := YTT()
	yt.Values(i.YttValues)
	yt.F(i.Path)
	for _, yttfile := range i.YttFiles {
		yt.F(yttfile)
	}
	f, err := yt.ToTempFile(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file by ytt: %w", err)
	}

	return f.Name(), nil
}

func (i *TemplateTestGivens) createTemplatingContext(workload v1alpha1.Workload, params map[string]apiextensionsv1.JSON) (map[string]interface{}, error) {
	var inputs *Inputs

	inputs, err := i.getBlueprintInputs()
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

func (i *TemplateTestGivens) getBlueprintInputs() (*Inputs, error) {
	if i.BlueprintInputsFile != "" && i.BlueprintInputs != nil {
		return nil, fmt.Errorf("only one of blueprintInputs or blueprintInputsFile may be set")
	}

	if i.BlueprintInputsFile == "" && i.BlueprintInputs == nil {
		return &Inputs{}, nil
	}

	if i.BlueprintInputs != nil {
		return i.BlueprintInputs, nil
	}

	inputsFile, err := os.ReadFile(i.BlueprintInputsFile)
	if err != nil {
		return nil, fmt.Errorf("could not read blueprintInputsFile: %w", err)
	}

	var inputs Inputs

	err = yaml.Unmarshal(inputsFile, &inputs)
	if err != nil {
		return nil, fmt.Errorf("unmarshall params: %w", err)
	}

	return &inputs, nil
}
