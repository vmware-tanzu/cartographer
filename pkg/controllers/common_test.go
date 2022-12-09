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

package controllers_test

import (
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

type lifecycleReader struct {
	lifecycle templates.Lifecycle
}

func (l *lifecycleReader) GetLifecycle() *templates.Lifecycle {
	return &l.lifecycle
}

func (l *lifecycleReader) GetDefaultParams() v1alpha1.TemplateParams {
	panic("not implemented")
}
func (l *lifecycleReader) GetResourceTemplate() v1alpha1.TemplateSpec {
	panic("not implemented")
}
func (l *lifecycleReader) GetHealthRule() *v1alpha1.HealthRule {
	panic("not implemented")
}
func (l *lifecycleReader) IsYTTTemplate() bool {
	panic("not implemented")
}
func (l *lifecycleReader) GetRetentionPolicy() v1alpha1.RetentionPolicy {
	panic("not implemented")
}
