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

//go:generate go run -modfile ../../hack/tools/go.mod github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"crypto/sha256"
	"fmt"

	"gopkg.in/yaml.v3"
	"k8s.io/utils/strings"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

type clusterImageTemplate struct {
	template  *v1alpha1.ClusterImageTemplate
	evaluator evaluator
}

func NewClusterImageTemplateModel(template *v1alpha1.ClusterImageTemplate, eval evaluator) *clusterImageTemplate {
	return &clusterImageTemplate{template: template, evaluator: eval}
}

func (t *clusterImageTemplate) GenerateResourceOutput(output *Output) ([]v1alpha1.Output, error) {
	if output == nil || output.Image == nil {
		return nil, nil
	}

	imageBytes, err := yaml.Marshal(output.Image)
	if err != nil {
		return nil, err
	}

	imageSHA := sha256.Sum256(imageBytes)

	return []v1alpha1.Output{
		{
			Name:    "image",
			Preview: strings.ShortenString(string(imageBytes), PREVIEW_CHARACTER_LIMIT),
			Digest:  fmt.Sprintf("sha256:%x", imageSHA),
		},
	}, nil
}

func (t *clusterImageTemplate) GetResourceTemplate() v1alpha1.TemplateSpec {
	return t.template.Spec.TemplateSpec
}

func (t *clusterImageTemplate) GetDefaultParams() v1alpha1.TemplateParams {
	return t.template.Spec.Params
}

func (t *clusterImageTemplate) GetHealthRule() *v1alpha1.HealthRule {
	return t.template.Spec.HealthRule
}
