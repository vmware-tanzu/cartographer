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

package realizer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/realizer"
)

type Template struct {
	params v1alpha1.TemplateParams
}

func (t Template) GetDefaultParams() v1alpha1.TemplateParams {
	return t.params
}

var _ = Describe("Params", func() {
	template := Template{
		params: v1alpha1.TemplateParams{
			v1alpha1.TemplateParam{
				Name:         "target-name",
				DefaultValue: apiextensionsv1.JSON{Raw: []byte("from the template")},
			},
		},
	}

	delegatingBlueprintParam := &v1alpha1.BlueprintParam{
		Name:         "target-name",
		DefaultValue: &apiextensionsv1.JSON{Raw: []byte("from the blueprint")},
	}

	nonDelegatingBlueprintParam := &v1alpha1.BlueprintParam{
		Name:  "target-name",
		Value: &apiextensionsv1.JSON{Raw: []byte("from the blueprint")},
	}

	delegatingResourceParam := &v1alpha1.BlueprintParam{
		Name:         "target-name",
		DefaultValue: &apiextensionsv1.JSON{Raw: []byte("from the resource")},
	}

	nonDelegatingResourceParam := &v1alpha1.BlueprintParam{
		Name:  "target-name",
		Value: &apiextensionsv1.JSON{Raw: []byte("from the resource")},
	}

	ownerParam := &v1alpha1.OwnerParam{
		Name:  "target-name",
		Value: apiextensionsv1.JSON{Raw: []byte("from the owner")},
	}

	DescribeTable("ParamsBuilder",
		func(templateParams realizer.TemplateParams,
			blueprintParam *v1alpha1.BlueprintParam,
			resourceParam *v1alpha1.BlueprintParam,
			ownerParam *v1alpha1.OwnerParam,
			expected string) {
			var (
				resourceParams  []v1alpha1.BlueprintParam
				blueprintParams []v1alpha1.BlueprintParam
				ownerParams     []v1alpha1.OwnerParam
			)

			if resourceParam != nil {
				resourceParams = append(resourceParams, *resourceParam)
			}
			if blueprintParam != nil {
				blueprintParams = append(blueprintParams, *blueprintParam)
			}
			if ownerParam != nil {
				ownerParams = append(ownerParams, *ownerParam)
			}

			actual := realizer.NewParamGenerator(resourceParams, blueprintParams, ownerParams).GetParams(templateParams)

			if expected == "" {
				Expect(actual).To(BeEmpty())
			} else {
				value, ok := actual["target-name"]
				Expect(ok).To(BeTrue())
				Expect(string(value.Raw)).To(Equal(expected))
			}
		},

		Entry("value only in template",
			template,
			nil,
			nil,
			nil,
			"from the template"),

		Entry("no value on template, values elsewhere",
			nil,
			nonDelegatingBlueprintParam,
			nil,
			nil,
			"from the blueprint"),

		Entry("value in template, resource, and owner; resource is not overridable",
			template,
			nil,
			nonDelegatingResourceParam,
			ownerParam,
			"from the resource"),

		Entry("value in template, blueprint, resource, and owner; blueprint and resource are not overridable",
			template,
			nonDelegatingBlueprintParam,
			nonDelegatingResourceParam,
			ownerParam,
			"from the resource"),

		Entry("value in template, blueprint, and owner; blueprint is not overridable",
			template,
			nonDelegatingBlueprintParam,
			nil,
			ownerParam,
			"from the blueprint"),

		Entry("value in template, blueprint, resource, and owner; blueprint is not overridable, resource is",
			template,
			nonDelegatingBlueprintParam,
			delegatingResourceParam,
			ownerParam,
			"from the owner"),

		Entry("value in template, resource, and owner; resource is not overridable",
			template,
			nil,
			delegatingResourceParam,
			ownerParam,
			"from the owner"),

		Entry("value in template, blueprint, resource, and owner; blueprint is overridable, resource is not",
			template,
			delegatingBlueprintParam,
			nonDelegatingResourceParam,
			ownerParam,
			"from the resource"),

		Entry("value in template, blueprint, and owner; blueprint is overridable",
			template,
			delegatingBlueprintParam,
			nil,
			ownerParam,
			"from the owner"),

		Entry("value in template, blueprint, resource; blueprint and resource are overridable",
			template,
			delegatingBlueprintParam,
			delegatingResourceParam,
			nil,
			"from the resource"),

		Entry("value in template, blueprint, resource, and owner; blueprint and resource are overridable",
			template,
			delegatingBlueprintParam,
			delegatingResourceParam,
			ownerParam,
			"from the owner"),

		Entry("value in template and owner",
			template,
			nil,
			nil,
			ownerParam,
			"from the owner"),
	)
})
