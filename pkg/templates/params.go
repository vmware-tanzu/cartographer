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
	blueprintParams []v1alpha1.OverridableParam,
	resourceParams []v1alpha1.OverridableParam,
	orderParams []v1alpha1.Param,
) Params {
	newParams := Params{}
	for _, param := range templateParams {
		newParams[param.Name] = param.DefaultValue
	}

	overridableByWorkload := make(map[string]bool)

	for key := range newParams {
		for _, supplyChainOverride := range blueprintParams {
			if key == supplyChainOverride.Name {
				newParams[key] = supplyChainOverride.Value
				overridableByWorkload[key] = supplyChainOverride.OverridableFlag
			}
		}

		for _, resourceOverride := range resourceParams {
			if key == resourceOverride.Name {
				newParams[key] = resourceOverride.Value
				overridableByWorkload[key] = resourceOverride.OverridableFlag
			}
		}

		for _, workloadOverride := range orderParams {
			if key == workloadOverride.Name && workloadCanOverride(overridableByWorkload, key) {
				newParams[key] = workloadOverride.Value
			}
		}
	}

	return newParams
}

func workloadCanOverride(isOverridable map[string]bool, key string) bool {
	overridable, written := isOverridable[key]
	return written && overridable
}
