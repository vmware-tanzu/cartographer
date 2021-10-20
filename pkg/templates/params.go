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

func ParamsBuilder(defaultParams v1alpha1.DefaultParams, resourceParams []v1alpha1.Param) Params {
	newParams := Params{}
	for _, param := range defaultParams {
		newParams[param.Name] = param.DefaultValue
	}

	for key := range newParams {
		for _, override := range resourceParams {
			if key == override.Name {
				newParams[key] = override.Value
			}
		}
	}
	return newParams
}
