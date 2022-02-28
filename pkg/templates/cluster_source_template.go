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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/strings"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

type clusterSourceTemplate struct {
	template      *v1alpha1.ClusterSourceTemplate
	evaluator     evaluator
	stampedObject *unstructured.Unstructured
}

func (t *clusterSourceTemplate) GetKind() string {
	return t.template.Kind
}

func NewClusterSourceTemplateModel(template *v1alpha1.ClusterSourceTemplate, eval evaluator) *clusterSourceTemplate {
	return &clusterSourceTemplate{template: template, evaluator: eval}
}

func (t *clusterSourceTemplate) GetName() string {
	return t.template.Name
}

func (t *clusterSourceTemplate) SetInputs(_ *Inputs) {}

func (t *clusterSourceTemplate) SetStampedObject(stampedObject *unstructured.Unstructured) {
	t.stampedObject = stampedObject
}

func (t *clusterSourceTemplate) GetOutput() (*Output, error) {
	url, err := t.evaluator.EvaluateJsonPath(t.template.Spec.URLPath, t.stampedObject.UnstructuredContent())
	if err != nil {
		return nil, JsonPathError{
			Err: fmt.Errorf("failed to evaluate the url path [%s]: %w",
				t.template.Spec.URLPath, err),
			expression: t.template.Spec.URLPath,
		}
	}

	revision, err := t.evaluator.EvaluateJsonPath(t.template.Spec.RevisionPath, t.stampedObject.UnstructuredContent())
	if err != nil {
		return nil, JsonPathError{
			Err: fmt.Errorf("failed to evaluate the revision path [%s]: %w",
				t.template.Spec.RevisionPath, err),
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

func (t *clusterSourceTemplate) GenerateResourceOutput(output *Output) []v1alpha1.Output {
	if output == nil {
		return nil
	}

	url := fmt.Sprintf("%v", output.Source.URL)
	urlSHA := sha256.Sum256([]byte(url))

	revision := fmt.Sprintf("%v", output.Source.Revision)
	revisionSHA := sha256.Sum256([]byte(revision))

	return []v1alpha1.Output{
		{
			Name:    "url",
			Preview: strings.ShortenString(url, 200),
			Digest:  fmt.Sprintf("sha256:%x", urlSHA),
		},
		{
			Name:    "revision",
			Preview: strings.ShortenString(revision, 200),
			Digest:  fmt.Sprintf("sha256:%x", revisionSHA),
		},
	}
}

func (t *clusterSourceTemplate) GetResourceTemplate() v1alpha1.TemplateSpec {
	return t.template.Spec.TemplateSpec
}

func (t *clusterSourceTemplate) GetDefaultParams() v1alpha1.TemplateParams {
	return t.template.Spec.Params
}
