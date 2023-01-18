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

type clusterConfigTemplate struct {
	template *v1alpha1.ClusterConfigTemplate
}

func (t *clusterConfigTemplate) GetLifecycle() *Lifecycle {
	lifecycle := convertLifecycle(t.template.Spec.Lifecycle)
	return &lifecycle
}

func NewClusterConfigTemplateReader(template *v1alpha1.ClusterConfigTemplate) *clusterConfigTemplate {
	return &clusterConfigTemplate{template: template}
}

func (t *clusterConfigTemplate) GetRetentionPolicy() v1alpha1.RetentionPolicy {
	if t.template.Spec.RetentionPolicy == nil {
		return v1alpha1.RetentionPolicy{
			MaxFailedRuns:     10,
			MaxSuccessfulRuns: 10,
		}
	}
	return *t.template.Spec.RetentionPolicy
}

func (t *clusterConfigTemplate) GetResourceTemplate() v1alpha1.TemplateSpec {
	return t.template.Spec.TemplateSpec
}

func (t *clusterConfigTemplate) GetDefaultParams() v1alpha1.TemplateParams {
	return t.template.Spec.Params
}

func (t *clusterConfigTemplate) GetHealthRule() *v1alpha1.HealthRule {
	if t.template.Spec.HealthRule == nil && t.template.Spec.Lifecycle == "tekton" {
		return &v1alpha1.HealthRule{SingleConditionType: "Succeeded"}
	}

	return t.template.Spec.HealthRule
}

func (t *clusterConfigTemplate) IsYTTTemplate() bool {
	return t.template.Spec.Ytt != ""
}
