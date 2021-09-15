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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

var _ = Describe("RunTemplate", func() {
	Describe("GetOutput", func() {
		var (
			apiTemplate   *v1alpha1.RunTemplate
			stampedObject *unstructured.Unstructured
		)

		BeforeEach(func() {
			apiTemplate = &v1alpha1.RunTemplate{}
			stampedObject = &unstructured.Unstructured{}
			stampedObjectManifest := utils.HereYamlF(`
				apiVersion: thing/v1
				kind: Thing
				metadata:
				  name: named-thing
				  namespace: somens
				spec:
				  simple: is a string
				  complex: 
					type: object
					name: complex object
			`)
			dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
			_, _, err := dec.Decode([]byte(stampedObjectManifest), nil, stampedObject)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("no outputs", func() {
			It("returns an empty list", func() {
				template := templates.NewRunTemplateModel(apiTemplate)
				outputs, err := template.GetOutput(stampedObject)
				Expect(err).NotTo(HaveOccurred())
				Expect(outputs).To(BeEmpty())
			})
		})

		Context("valid output paths defined", func() {
			BeforeEach(func() {
				apiTemplate.Spec.Outputs = map[string]string{
					"simplistic": "spec.simple",
					"complexish": "spec.complex",
				}
			})
			It("returns the outputs", func() {
				template := templates.NewRunTemplateModel(apiTemplate)
				outputs, err := template.GetOutput(stampedObject)
				Expect(err).NotTo(HaveOccurred())
				Expect(outputs["simplistic"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"is a string"`)}))
				Expect(outputs["complexish"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`{"name":"complex object","type":"object"}`)}))
			})
		})

		Context("invalid output paths defined", func() {
			BeforeEach(func() {
				apiTemplate.Spec.Outputs = map[string]string{
					"complexish": "spec.nonexistant",
				}
			})
			It("returns the outputs", func() {
				template := templates.NewRunTemplateModel(apiTemplate)
				_, err := template.GetOutput(stampedObject)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("get output: evaluate: find results: nonexistant is not found"))
			})
		})
	})
})
