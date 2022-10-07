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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

type TemplateParams interface {
	GetDefaultParams() v1alpha1.TemplateParams
}

type ContextParams interface {
	GetParams(templateParams TemplateParams) map[string]apiextensionsv1.JSON
}

func NewParamGenerator(resourceParams []v1alpha1.BlueprintParam, blueprintParams []v1alpha1.BlueprintParam, ownerParams []v1alpha1.OwnerParam) ContextParams {
	return &ParamGenerator{
		blueprintParams: blueprintParams,
		resourceParams:  resourceParams,
		ownerParams:     ownerParams,
	}
}

type ParamGenerator struct {
	blueprintParams []v1alpha1.BlueprintParam
	resourceParams  []v1alpha1.BlueprintParam
	ownerParams     []v1alpha1.OwnerParam
}

func (p ParamGenerator) GetParams(templateParams TemplateParams) map[string]apiextensionsv1.JSON {
	newParams := map[string]apiextensionsv1.JSON{}

	for _, param := range templateParams.GetDefaultParams() {
		newParams[param.Name] = param.DefaultValue
	}

	protectedFromOwnerOverride := make(map[string]bool)

	for _, blueprintOverride := range p.blueprintParams {
		key := blueprintOverride.Name
		if blueprintOverride.Value != nil {
			newParams[key] = *blueprintOverride.Value
			protectedFromOwnerOverride[key] = true
		} else {
			newParams[key] = *blueprintOverride.DefaultValue
			protectedFromOwnerOverride[key] = false
		}
	}

	for _, resourceOverride := range p.resourceParams {
		key := resourceOverride.Name
		if resourceOverride.Value != nil {
			newParams[key] = *resourceOverride.Value
			protectedFromOwnerOverride[key] = true
		} else {
			newParams[key] = *resourceOverride.DefaultValue
			protectedFromOwnerOverride[key] = false
		}
	}

	for _, ownerOverride := range p.ownerParams {
		key := ownerOverride.Name
		if ownerCanOverride(protectedFromOwnerOverride, key) {
			newParams[key] = ownerOverride.Value
		}
	}

	return newParams

}

func ownerCanOverride(isProtected map[string]bool, key string) bool {
	protected, written := isProtected[key]
	return !written || !protected
}
