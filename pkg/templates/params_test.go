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
	resourceParam := v1alpha1.Param{
		Name:  "target-name",
		Value: apiextensionsv1.JSON{Raw: []byte("from the resource")},
	}

	templateParam := &v1alpha1.TemplateParam{
		Name:         "target-name",
		DefaultValue: apiextensionsv1.JSON{Raw: []byte("from the template")},
	}

	blueprintParam := v1alpha1.Param{
		Name:  "target-name",
		Value: apiextensionsv1.JSON{Raw: []byte("from the blueprint")},
	}

	ownerParam := &v1alpha1.Param{
		Name:  "target-name",
		Value: apiextensionsv1.JSON{Raw: []byte("from the owner")},
	}

	DescribeTable("ParamsBuilder",
		func(templateParam *v1alpha1.TemplateParam,
			blueprintParam *v1alpha1.OverridableParam,
			resourceParam *v1alpha1.OverridableParam,
			ownerParam *v1alpha1.Param,
			expected string) {
			var (
				templateParams  []v1alpha1.TemplateParam
				resourceParams  []v1alpha1.OverridableParam
				blueprintParams []v1alpha1.OverridableParam
				ownerParams     []v1alpha1.Param
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
			if ownerParam != nil {
				ownerParams = append(ownerParams, *ownerParam)
			}

			actual := templates.ParamsBuilder(
				templateParams,
				blueprintParams,
				resourceParams,
				ownerParams)

			if expected == "" {
				Expect(actual).To(BeEmpty())
			} else {
				value, ok := actual["target-name"]
				Expect(ok).To(BeTrue())
				Expect(string(value.Raw)).To(Equal(expected))
			}
		},

		Entry("value only in template",
			templateParam,
			&v1alpha1.OverridableParam{},
			&v1alpha1.OverridableParam{},
			&v1alpha1.Param{},
			"from the template"),

		Entry("no value on template, values elsewhere",
			nil,
			&v1alpha1.OverridableParam{Param: resourceParam, OverridableFlag: false},
			&v1alpha1.OverridableParam{},
			&v1alpha1.Param{},
			""),

		Entry("value in template, resource, and owner; resource is not overridable",
			templateParam,
			&v1alpha1.OverridableParam{},
			&v1alpha1.OverridableParam{Param: resourceParam, OverridableFlag: false},
			ownerParam,
			"from the resource"),

		Entry("value in template, blueprint, resource, and owner; blueprint and resource are not overridable",
			templateParam,
			&v1alpha1.OverridableParam{Param: blueprintParam, OverridableFlag: false},
			&v1alpha1.OverridableParam{Param: resourceParam, OverridableFlag: false},
			ownerParam,
			"from the resource"),

		Entry("value in template, blueprint, and owner; blueprint is not overridable",
			templateParam,
			&v1alpha1.OverridableParam{Param: blueprintParam, OverridableFlag: false},
			&v1alpha1.OverridableParam{},
			ownerParam,
			"from the blueprint"),

		Entry("value in template, blueprint, resource, and owner; blueprint is not overridable, resource is",
			templateParam,
			&v1alpha1.OverridableParam{Param: blueprintParam, OverridableFlag: false},
			&v1alpha1.OverridableParam{Param: resourceParam, OverridableFlag: true},
			ownerParam,
			"from the owner"),

		Entry("value in template, resource, and owner; resource is not overridable",
			templateParam,
			&v1alpha1.OverridableParam{},
			&v1alpha1.OverridableParam{Param: resourceParam, OverridableFlag: true},
			ownerParam,
			"from the owner"),

		Entry("value in template, blueprint, resource, and owner; blueprint is overridable, resource is not",
			templateParam,
			&v1alpha1.OverridableParam{Param: blueprintParam, OverridableFlag: true},
			&v1alpha1.OverridableParam{Param: resourceParam, OverridableFlag: false},
			ownerParam,
			"from the resource"),

		Entry("value in template, blueprint, and owner; blueprint is overridable",
			templateParam,
			&v1alpha1.OverridableParam{Param: blueprintParam, OverridableFlag: true},
			&v1alpha1.OverridableParam{},
			ownerParam,
			"from the owner"),

		Entry("value in template, blueprint, resource; blueprint and resource are overridable",
			templateParam,
			&v1alpha1.OverridableParam{Param: blueprintParam, OverridableFlag: true},
			&v1alpha1.OverridableParam{Param: resourceParam, OverridableFlag: true},
			&v1alpha1.Param{},
			"from the resource"),

		Entry("value in template, blueprint, resource, and owner; blueprint and resource are overridable",
			templateParam,
			&v1alpha1.OverridableParam{Param: blueprintParam, OverridableFlag: true},
			&v1alpha1.OverridableParam{Param: resourceParam, OverridableFlag: true},
			ownerParam,
			"from the owner"),

		Entry("value in template and owner",
			templateParam,
			&v1alpha1.OverridableParam{},
			&v1alpha1.OverridableParam{},
			ownerParam,
			"from the template"),
	)
})
