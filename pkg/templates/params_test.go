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
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

var _ = Describe("Params", func() {
	DescribeTable("ParamsBuilder",
		func(templateParam *v1alpha1.TemplateParam,
			blueprintParam *v1alpha1.OverridableParam,
			resourceParam *v1alpha1.OverridableParam,
			orderParam *v1alpha1.Param,
			expected string) {
			var (
				templateParams  []v1alpha1.TemplateParam
				resourceParams  []v1alpha1.OverridableParam
				blueprintParams []v1alpha1.OverridableParam
				orderParams     []v1alpha1.Param
			)

			if templateParam != nil {
				templateParams = append(templateParams, *templateParam)
			}
			if resourceParam != nil {
				resourceParams = append(resourceParams, *resourceParam)
			}
			if blueprintParam != nil {
				blueprintParams = append(blueprintParams, *blueprintParam)
			}
			if orderParam != nil {
				orderParams = append(orderParams, *orderParam)
			}

			actual := templates.ParamsBuilder(
				templateParams,
				blueprintParams,
				resourceParams,
				orderParams)

			if expected == "" {
				Expect(actual).To(BeEmpty())
			} else {
				value, ok := actual["target-name"]
				Expect(ok).To(BeTrue())
				Expect(string(value.Raw)).To(Equal(expected))
			}
		},

		Entry("value only in template",
			&v1alpha1.TemplateParam{
				Name:         "target-name",
				DefaultValue: apiextensionsv1.JSON{Raw: []byte("from the template")},
			},
			&v1alpha1.OverridableParam{},
			&v1alpha1.OverridableParam{},
			&v1alpha1.Param{},
			"from the template"),

		Entry("no value on template, values elsewhere",
			nil,
			&v1alpha1.OverridableParam{
				Param: v1alpha1.Param{
					Name:  "target-name",
					Value: apiextensionsv1.JSON{Raw: []byte("from the resource")},
				},
				OverridableFlag: false,
			},
			&v1alpha1.OverridableParam{},
			&v1alpha1.Param{},
			""),

		Entry("value in template, resource, and workload; resource is not overridable",
			&v1alpha1.TemplateParam{
				Name:         "target-name",
				DefaultValue: apiextensionsv1.JSON{Raw: []byte("from the template")},
			},
			&v1alpha1.OverridableParam{},
			&v1alpha1.OverridableParam{
				Param: v1alpha1.Param{
					Name:  "target-name",
					Value: apiextensionsv1.JSON{Raw: []byte("from the resource")},
				},
				OverridableFlag: false,
			},
			&v1alpha1.Param{
				Name:  "target-name",
				Value: apiextensionsv1.JSON{Raw: []byte("from the workload")},
			},
			"from the resource"),

		Entry("value in template, blueprint, resource, and workload; blueprint and resource are not overridable",
			&v1alpha1.TemplateParam{
				Name:         "target-name",
				DefaultValue: apiextensionsv1.JSON{Raw: []byte("from the template")},
			},
			&v1alpha1.OverridableParam{
				Param: v1alpha1.Param{
					Name:  "target-name",
					Value: apiextensionsv1.JSON{Raw: []byte("from the blueprint")},
				},
				OverridableFlag: false,
			},
			&v1alpha1.OverridableParam{
				Param: v1alpha1.Param{
					Name:  "target-name",
					Value: apiextensionsv1.JSON{Raw: []byte("from the resource")},
				},
				OverridableFlag: false,
			},
			&v1alpha1.Param{
				Name:  "target-name",
				Value: apiextensionsv1.JSON{Raw: []byte("from the workload")},
			},
			"from the resource"),

		Entry("value in template, blueprint, and workload; blueprint is not overridable",
			&v1alpha1.TemplateParam{
				Name:         "target-name",
				DefaultValue: apiextensionsv1.JSON{Raw: []byte("from the template")},
			},
			&v1alpha1.OverridableParam{
				Param: v1alpha1.Param{
					Name:  "target-name",
					Value: apiextensionsv1.JSON{Raw: []byte("from the blueprint")},
				},
				OverridableFlag: false,
			},
			&v1alpha1.OverridableParam{},
			&v1alpha1.Param{
				Name:  "target-name",
				Value: apiextensionsv1.JSON{Raw: []byte("from the workload")},
			},
			"from the blueprint"),

		Entry("value in template, blueprint, resource, and workload; blueprint is not overridable, resource is",
			&v1alpha1.TemplateParam{
				Name:         "target-name",
				DefaultValue: apiextensionsv1.JSON{Raw: []byte("from the template")},
			},
			&v1alpha1.OverridableParam{
				Param: v1alpha1.Param{
					Name:  "target-name",
					Value: apiextensionsv1.JSON{Raw: []byte("from the blueprint")},
				},
				OverridableFlag: false,
			},
			&v1alpha1.OverridableParam{
				Param: v1alpha1.Param{
					Name:  "target-name",
					Value: apiextensionsv1.JSON{Raw: []byte("from the resource")},
				},
				OverridableFlag: true,
			},
			&v1alpha1.Param{
				Name:  "target-name",
				Value: apiextensionsv1.JSON{Raw: []byte("from the workload")},
			},
			"from the workload"),

		Entry("value in template, resource, and workload; resource is not overridable",
			&v1alpha1.TemplateParam{
				Name:         "target-name",
				DefaultValue: apiextensionsv1.JSON{Raw: []byte("from the template")},
			},
			&v1alpha1.OverridableParam{},
			&v1alpha1.OverridableParam{
				Param: v1alpha1.Param{
					Name:  "target-name",
					Value: apiextensionsv1.JSON{Raw: []byte("from the resource")},
				},
				OverridableFlag: true,
			},
			&v1alpha1.Param{
				Name:  "target-name",
				Value: apiextensionsv1.JSON{Raw: []byte("from the workload")},
			},
			"from the workload"),

		Entry("value in template, blueprint, resource, and workload; blueprint is overridable, resource is not",
			&v1alpha1.TemplateParam{
				Name:         "target-name",
				DefaultValue: apiextensionsv1.JSON{Raw: []byte("from the template")},
			},
			&v1alpha1.OverridableParam{
				Param: v1alpha1.Param{
					Name:  "target-name",
					Value: apiextensionsv1.JSON{Raw: []byte("from the blueprint")},
				},
				OverridableFlag: true,
			},
			&v1alpha1.OverridableParam{
				Param: v1alpha1.Param{
					Name:  "target-name",
					Value: apiextensionsv1.JSON{Raw: []byte("from the resource")},
				},
				OverridableFlag: false,
			},
			&v1alpha1.Param{
				Name:  "target-name",
				Value: apiextensionsv1.JSON{Raw: []byte("from the workload")},
			},
			"from the resource"),

		Entry("value in template, blueprint, and workload; blueprint is overridable",
			&v1alpha1.TemplateParam{
				Name:         "target-name",
				DefaultValue: apiextensionsv1.JSON{Raw: []byte("from the template")},
			},
			&v1alpha1.OverridableParam{
				Param: v1alpha1.Param{
					Name:  "target-name",
					Value: apiextensionsv1.JSON{Raw: []byte("from the blueprint")},
				},
				OverridableFlag: true,
			},
			&v1alpha1.OverridableParam{},
			&v1alpha1.Param{
				Name:  "target-name",
				Value: apiextensionsv1.JSON{Raw: []byte("from the workload")},
			},
			"from the workload"),

		Entry("value in template, blueprint, resource; blueprint and resource are overridable",
			&v1alpha1.TemplateParam{
				Name:         "target-name",
				DefaultValue: apiextensionsv1.JSON{Raw: []byte("from the template")},
			},
			&v1alpha1.OverridableParam{
				Param: v1alpha1.Param{
					Name:  "target-name",
					Value: apiextensionsv1.JSON{Raw: []byte("from the blueprint")},
				},
				OverridableFlag: true,
			},
			&v1alpha1.OverridableParam{
				Param: v1alpha1.Param{
					Name:  "target-name",
					Value: apiextensionsv1.JSON{Raw: []byte("from the resource")},
				},
				OverridableFlag: true,
			},
			&v1alpha1.Param{},
			"from the resource"),

		Entry("value in template, blueprint, resource, and workload; blueprint and resource are overridable"+
			"ovrdbl-supply-chain-ovrdbl-resource-on-workload",
			&v1alpha1.TemplateParam{
				Name:         "target-name",
				DefaultValue: apiextensionsv1.JSON{Raw: []byte("from the template")},
			},
			&v1alpha1.OverridableParam{
				Param: v1alpha1.Param{
					Name:  "target-name",
					Value: apiextensionsv1.JSON{Raw: []byte("from the blueprint")},
				},
				OverridableFlag: true,
			},
			&v1alpha1.OverridableParam{
				Param: v1alpha1.Param{
					Name:  "target-name",
					Value: apiextensionsv1.JSON{Raw: []byte("from the resource")},
				},
				OverridableFlag: true,
			},
			&v1alpha1.Param{
				Name:  "target-name",
				Value: apiextensionsv1.JSON{Raw: []byte("from the workload")},
			},
			"from the workload"),

		Entry("value in template and workload",
			&v1alpha1.TemplateParam{
				Name:         "target-name",
				DefaultValue: apiextensionsv1.JSON{Raw: []byte("from the template")},
			},
			&v1alpha1.OverridableParam{},
			&v1alpha1.OverridableParam{},
			&v1alpha1.Param{
				Name:  "target-name",
				Value: apiextensionsv1.JSON{Raw: []byte("from the workload")},
			},
			"from the template"),
	)
})
