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
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

type runTemplate struct {
	template *v1alpha1.RunTemplate
}

func NewRunTemplateModel(template *v1alpha1.RunTemplate) *runTemplate {
	return &runTemplate{template: template}
}

func (t runTemplate) GetName() string {
	return t.template.Name
}

// TODO: not yet implemented
func (t runTemplate) GetOutput(stampedObject *unstructured.Unstructured) (*Output, error) {
	return &Output{}, nil
}

func (t runTemplate) GetResourceTemplate() runtime.RawExtension {
	return t.template.Spec.Template
}

// TODO: not yet implemented
func (t runTemplate) GetDefaultParams() v1alpha1.DefaultParams {
	return v1alpha1.DefaultParams{}
}
