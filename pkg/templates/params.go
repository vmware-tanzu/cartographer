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

type Param struct {
	Name  string               `json:"name"`
	Value apiextensionsv1.JSON `json:"value"`
}

type Params []Param

func ParamsBuilder(defaultParams v1alpha1.DefaultParams, componentParams []v1alpha1.SupplyChainParam) Params {
	newParams := Params{}
	for _, param := range defaultParams {
		newParams = append(newParams, Param{
			Name:  param.Name,
			Value: param.DefaultValue,
		})
	}

	for i, param := range newParams {
		for _, override := range componentParams {
			if param.Name == override.Name {
				newParams[i].Value = override.Value
			}
		}
	}
	return newParams
}
