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
	"crypto/sha256"
	"fmt"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/strings"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

type clusterConfigTemplate struct {
	template  *v1alpha1.ClusterConfigTemplate
	evaluator evaluator
}

func (t *clusterConfigTemplate) GetOutputPaths() map[string]string {
	return map[string]string{
		"config": t.template.Spec.ConfigPath,
	}
}

func NewClusterConfigTemplateModel(template *v1alpha1.ClusterConfigTemplate, eval evaluator) *clusterConfigTemplate {
	return &clusterConfigTemplate{template: template, evaluator: eval}
}

func (t *clusterConfigTemplate) SetInputs(_ Inputs) {}

func (t *clusterConfigTemplate) GetOutput(stampedObject *unstructured.Unstructured) (*Output, error) {
	config, err := t.evaluator.EvaluateJsonPath(t.template.Spec.ConfigPath, stampedObject.UnstructuredContent())
	if err != nil {
		return nil, JsonPathError{
			Err: fmt.Errorf("failed to evaluate spec.configPath [%s]: %w",
				t.template.Spec.ConfigPath, err),
			expression: t.template.Spec.ConfigPath,
		}
	}

	return &Output{
		Config: config,
	}, nil
}

func (t *clusterConfigTemplate) GenerateResourceOutput(output *Output) ([]v1alpha1.Output, error) {
	if output == nil || output.Config == nil {
		return nil, nil
	}

	configBytes, err := yaml.Marshal(output.Config)
	if err != nil {
		return nil, err
	}

	configSHA := sha256.Sum256(configBytes)

	return []v1alpha1.Output{
		{
			Name:    "config",
			Preview: strings.ShortenString(string(configBytes), PREVIEW_CHARACTER_LIMIT),
			Digest:  fmt.Sprintf("sha256:%x", configSHA),
		},
	}, nil
}

func (t *clusterConfigTemplate) GetResourceTemplate() v1alpha1.TemplateSpec {
	return t.template.Spec.TemplateSpec
}

func (t *clusterConfigTemplate) GetDefaultParams() v1alpha1.TemplateParams {
	return t.template.Spec.Params
}

func (t *clusterConfigTemplate) GetHealthRule() *v1alpha1.HealthRule {
	return t.template.Spec.HealthRule
}
