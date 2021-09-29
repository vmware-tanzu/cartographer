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

package templates_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

var _ = Describe("OverrideDefaultParams", func() {
	Describe("ParamsBuilder", func() {
		It("makes Params using v1alpha.DefaultParams and overrides them with v1alpha.SupplyChainParam values", func() {
			defaultParams := v1alpha1.DefaultParams{
				{
					Name: "foo",
					DefaultValue: apiextensionsv1.JSON{
						Raw: []byte(`bar`),
					},
				},
				{
					Name: "fizz",
					DefaultValue: apiextensionsv1.JSON{
						Raw: []byte(`baz`),
					},
				},
			}
			resourceParams := []v1alpha1.SupplyChainParam{
				{
					Name: "fizz",
					Value: apiextensionsv1.JSON{
						Raw: []byte(`buzz`),
					},
				},
			}
			params := templates.ParamsBuilder(defaultParams, resourceParams)

			Expect(params).To(HaveLen(2))
			Expect(params["foo"].Raw).To(Equal([]byte("bar")))
			Expect(params["fizz"].Raw).To(Equal([]byte("buzz")))
		})
	})
})
