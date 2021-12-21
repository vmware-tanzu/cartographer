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

type Params map[string]apiextensionsv1.JSON

func ParamsBuilder(
	templateParams []v1alpha1.TemplateParam,
	blueprintParams []v1alpha1.BlueprintParam,
	resourceParams []v1alpha1.BlueprintParam,
	ownerParams []v1alpha1.OwnerParam,
) Params {
	newParams := Params{}
	for _, param := range templateParams {
		newParams[param.Name] = param.DefaultValue
	}

	protectedFromOwnerOverride := make(map[string]bool)

	for _, blueprintOverride := range blueprintParams {
		key := blueprintOverride.Name
		if blueprintOverride.Value != nil {
			newParams[key] = *blueprintOverride.Value
			protectedFromOwnerOverride[key] = true
		} else {
			newParams[key] = *blueprintOverride.DefaultValue
			protectedFromOwnerOverride[key] = false
		}
	}

	for _, resourceOverride := range resourceParams {
		key := resourceOverride.Name
		if resourceOverride.Value != nil {
			newParams[key] = *resourceOverride.Value
			protectedFromOwnerOverride[key] = true
		} else {
			newParams[key] = *resourceOverride.DefaultValue
			protectedFromOwnerOverride[key] = false
		}
	}

	for _, ownerOverride := range ownerParams {
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
