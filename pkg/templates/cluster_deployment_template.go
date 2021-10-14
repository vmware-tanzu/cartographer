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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

type clusterDeploymentTemplate struct {
	template  *v1alpha1.ClusterDeploymentTemplate
	evaluator evaluator
}

func (t clusterDeploymentTemplate) GetKind() string {
	return t.template.Kind
}

func NewClusterDeploymentTemplateModel(template *v1alpha1.ClusterDeploymentTemplate, eval evaluator) *clusterDeploymentTemplate {
	return &clusterDeploymentTemplate{template: template, evaluator: eval}
}

func (t clusterDeploymentTemplate) GetName() string {
	return t.template.Name
}

func (t clusterDeploymentTemplate) GetOutput(stampedObject *unstructured.Unstructured) (*Output, error) {
	return &Output{}, nil
}

func (t clusterDeploymentTemplate) GetResourceTemplate() v1alpha1.TemplateSpec {
	return t.template.Spec
}

func (t clusterDeploymentTemplate) GetDefaultParams() v1alpha1.DefaultParams {
	return t.template.Spec.Params
}
