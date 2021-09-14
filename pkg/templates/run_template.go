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

package templates

import (
	"encoding/json"
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/eval"
)

type Outputs map[string]apiextensionsv1.JSON

type RunTemplate interface {
	GetName() string
	GetResourceTemplate() v1alpha1.TemplateSpec
	GetOutput(stampedObject *unstructured.Unstructured) (Outputs, error)
}

type runTemplate struct {
	template *v1alpha1.RunTemplate
}

func (t runTemplate) GetOutput(stampedObject *unstructured.Unstructured) (Outputs, error) {
	evaluator := eval.EvaluatorBuilder()

	outputs := Outputs{}

	for key, path := range t.template.Spec.Outputs {
		output, err := evaluator.EvaluateJsonPath(path, stampedObject.UnstructuredContent())
		if err != nil {
			return nil, fmt.Errorf("get output: %w", err)
		}

		result, err := json.Marshal(output)
		if err != nil {
			return nil, fmt.Errorf("get output could not marshal jsonpath output: %w", err)
		}

		ext := apiextensionsv1.JSON{Raw: result}
		outputs[key] = ext

	}

	return outputs, nil
}

func NewRunTemplateModel(template *v1alpha1.RunTemplate) RunTemplate {
	return &runTemplate{template: template}
}

func (t runTemplate) GetName() string {
	return t.template.Name
}

func (t runTemplate) GetResourceTemplate() v1alpha1.TemplateSpec {
	return v1alpha1.TemplateSpec{
		Template: &t.template.Spec.Template,
	}
}
