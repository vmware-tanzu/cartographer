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
	"k8s.io/utils/strings"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

type clusterSourceTemplate struct {
	template  *v1alpha1.ClusterSourceTemplate
	evaluator evaluator
}

func NewClusterSourceTemplateModel(template *v1alpha1.ClusterSourceTemplate, eval evaluator) *clusterSourceTemplate {
	return &clusterSourceTemplate{template: template, evaluator: eval}
}

func (t *clusterSourceTemplate) GenerateResourceOutput(output *Output) ([]v1alpha1.Output, error) {
	if output == nil || output.Source == nil {
		return nil, nil
	}

	urlBytes, err := yaml.Marshal(output.Source.URL)
	if err != nil {
		return nil, err
	}

	urlSHA := sha256.Sum256(urlBytes)

	revBytes, err := yaml.Marshal(output.Source.Revision)
	if err != nil {
		return nil, err
	}

	revSHA := sha256.Sum256(revBytes)

	return []v1alpha1.Output{
		{
			Name:    "url",
			Preview: strings.ShortenString(string(urlBytes), PREVIEW_CHARACTER_LIMIT),
			Digest:  fmt.Sprintf("sha256:%x", urlSHA),
		},
		{
			Name:    "revision",
			Preview: strings.ShortenString(string(revBytes), PREVIEW_CHARACTER_LIMIT),
			Digest:  fmt.Sprintf("sha256:%x", revSHA),
		},
	}, nil
}

func (t *clusterSourceTemplate) GetResourceTemplate() v1alpha1.TemplateSpec {
	return t.template.Spec.TemplateSpec
}

func (t *clusterSourceTemplate) GetDefaultParams() v1alpha1.TemplateParams {
	return t.template.Spec.Params
}

func (t *clusterSourceTemplate) GetHealthRule() *v1alpha1.HealthRule {
	return t.template.Spec.HealthRule
}
