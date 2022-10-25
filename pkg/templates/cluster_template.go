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
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

type clusterTemplate struct {
	template *v1alpha1.ClusterTemplate
}

func (t *clusterTemplate) GetLifecycle() *Lifecycle {
	lifecycle := convertLifecycle(t.template.Spec.Lifecycle)
	return &lifecycle
}

func NewClusterTemplateReader(template *v1alpha1.ClusterTemplate) *clusterTemplate {
	return &clusterTemplate{template: template}
}

func (t *clusterTemplate) GetResourceTemplate() v1alpha1.TemplateSpec {
	return t.template.Spec
}

func (t *clusterTemplate) GetDefaultParams() v1alpha1.TemplateParams {
	return t.template.Spec.Params
}

func (t *clusterTemplate) GetHealthRule() *v1alpha1.HealthRule {
	return t.template.Spec.HealthRule
}

func (t *clusterTemplate) IsYTTTemplate() bool {
	return t.template.Spec.Ytt != ""
}
