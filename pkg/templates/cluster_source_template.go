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

type clusterSourceTemplate struct {
	template  *v1alpha1.ClusterSourceTemplate
	evaluator evaluator
}

func (t clusterSourceTemplate) GetKind() string {
	return t.template.Kind
}

func NewClusterSourceTemplateModel(template *v1alpha1.ClusterSourceTemplate, eval evaluator) *clusterSourceTemplate {
	return &clusterSourceTemplate{template: template, evaluator: eval}
}

func (t clusterSourceTemplate) GetName() string {
	return t.template.Name
}

func (t clusterSourceTemplate) GetOutput(stampedObject *unstructured.Unstructured, templatingContext map[string]interface{}) (*Output, error) {
	url, err := t.evaluator.EvaluateJsonPath(t.template.Spec.URLPath, stampedObject.UnstructuredContent())
	if err != nil {
		return nil, JsonPathError{
			Err:        fmt.Errorf("evaluate source url json path: %w", err),
			expression: t.template.Spec.URLPath,
		}
	}

	revision, err := t.evaluator.EvaluateJsonPath(t.template.Spec.RevisionPath, stampedObject.UnstructuredContent())
	if err != nil {
		return nil, JsonPathError{
			Err:        fmt.Errorf("evaluate source revision json path: %w", err),
			expression: t.template.Spec.RevisionPath,
		}
	}
	return &Output{
		Source: &Source{
			URL:      url,
			Revision: revision,
		},
	}, nil
}

func (t clusterSourceTemplate) GetResourceTemplate() v1alpha1.TemplateSpec {
	return t.template.Spec.TemplateSpec
}

func (t clusterSourceTemplate) GetDefaultParams() v1alpha1.DefaultParams {
	return t.template.Spec.Params
}
