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
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

type clusterImageTemplate struct {
	template *v1alpha1.ClusterImageTemplate
}

func (t *clusterImageTemplate) GetLifecycle() *Lifecycle {
	lifecycle := convertLifecycle(t.template.Spec.Lifecycle)
	return &lifecycle
}

func NewClusterImageTemplateReader(template *v1alpha1.ClusterImageTemplate) *clusterImageTemplate {
	return &clusterImageTemplate{template: template}
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

func (t *clusterImageTemplate) IsYTTTemplate() bool {
	return t.template.Spec.Ytt != ""
}
