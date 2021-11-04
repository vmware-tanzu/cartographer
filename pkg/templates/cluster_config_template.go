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
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

type clusterConfigTemplate struct {
	template  *v1alpha1.ClusterConfigTemplate
	evaluator evaluator
}

func (t clusterConfigTemplate) GetKind() string {
	return t.template.Kind
}

func NewClusterConfigTemplateModel(template *v1alpha1.ClusterConfigTemplate, eval evaluator) *clusterConfigTemplate {
	return &clusterConfigTemplate{template: template, evaluator: eval}
}

func (t clusterConfigTemplate) GetName() string {
	return t.template.Name
}

func (t clusterConfigTemplate) GetOutput(stampedObject *unstructured.Unstructured, templatingContext map[string]interface{}) (*Output, error) {
	config, err := t.evaluator.EvaluateJsonPath(t.template.Spec.ConfigPath, stampedObject.UnstructuredContent())
	if err != nil {
		return nil, JsonPathError{
			Err:        fmt.Errorf("evaluate config url json path: %w", err),
			expression: t.template.Spec.ConfigPath,
		}
	}

	return &Output{
		Config: config,
	}, nil
}

func (t clusterConfigTemplate) GetResourceTemplate() v1alpha1.TemplateSpec {
	return t.template.Spec.TemplateSpec
}

func (t clusterConfigTemplate) GetDefaultParams() v1alpha1.DefaultParams {
	return t.template.Spec.Params
}
