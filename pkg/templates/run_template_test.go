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

	"github.com/vmware-tanzu/cartographer/pkg/templates"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

var _ = Describe("RunTemplate", func() {
	Describe("GetOutput", func() {
		var (
			apiTemplate                                                         *v1alpha1.RunTemplate
			firstStampedObject, secondStampedObject, unconditionedStampedObject *unstructured.Unstructured
			stampedObjects                                                      []*unstructured.Unstructured
			originalOutputs                                                     map[string]apiextensionsv1.JSON
		)

		BeforeEach(func() {
			apiTemplate = &v1alpha1.RunTemplate{}

			firstStampedObject = &unstructured.Unstructured{}
			firstStampedObjectManifest := utils.HereYamlF(`
				apiVersion: thing/v1
				kind: Thing
				metadata:
				  name: named-thing
				  namespace: somens
				  creationTimestamp: "2021-09-17T16:02:30Z"
				spec:
				  simple: is a string
				  complex: 
					type: object
					name: complex object
				  only-exists-on-first-object: populated
				status:
				  conditions:
				    - type: Succeeded
				      status: "True"
			`)

			dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
			_, _, err := dec.Decode([]byte(firstStampedObjectManifest), nil, firstStampedObject)
			Expect(err).NotTo(HaveOccurred())

			secondStampedObject = &unstructured.Unstructured{}
			secondStampedObjectManifest := utils.HereYamlF(`
				apiVersion: thing/v1
				kind: Thing
				metadata:
				  name: named-thing
				  namespace: somens
				  creationTimestamp: "2021-09-17T16:02:40Z"
				spec:
				  simple: 2nd-simple
				  complex: 2nd-complex
				status:
				  conditions:
				    - type: Succeeded
				      status: "True"
			`)

			dec = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
			_, _, err = dec.Decode([]byte(secondStampedObjectManifest), nil, secondStampedObject)
			Expect(err).NotTo(HaveOccurred())

			unconditionedStampedObject = &unstructured.Unstructured{}
			unconditionedStampedObjectManifest := utils.HereYamlF(`
				apiVersion: thing/v1
				kind: Thing
				metadata:
				  name: named-thing
				  namespace: somens
				spec:
				  simple: val
				  complex: other-val
			`)

			dec = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
			_, _, err = dec.Decode([]byte(unconditionedStampedObjectManifest), nil, unconditionedStampedObject)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the previous outputs are empty", func() {
			BeforeEach(func() {
				originalOutputs = make(map[string]apiextensionsv1.JSON)
			})

			Context("when there is one object", func() {
				BeforeEach(func() {
					stampedObjects = []*unstructured.Unstructured{firstStampedObject}
				})

				Context("with no outputs", func() {
					It("returns an empty list", func() {
						template := templates.NewRunTemplateModel(apiTemplate)
						outputs, err := template.GetOutput(originalOutputs, stampedObjects)
						Expect(err).NotTo(HaveOccurred())
						Expect(outputs).To(BeEmpty())
					})
				})

				Context("with valid output paths defined", func() {
					BeforeEach(func() {
						apiTemplate.Spec.Outputs = map[string]string{
							"simplistic": "spec.simple",
							"complexish": "spec.complex",
						}
					})

					Context("when the object has not succeeded", func() {
						BeforeEach(func() {
							Expect(utils.AlterFieldOfNestedStringMaps(firstStampedObject.Object, "status.conditions.[0]status", "False")).To(Succeed()) // TODO: fix this notation or start using a jsonpath parser
						})
						It("returns an empty list", func() {
							template := templates.NewRunTemplateModel(apiTemplate)
							outputs, err := template.GetOutput(originalOutputs, stampedObjects)
							Expect(err).NotTo(HaveOccurred())
							Expect(outputs).To(BeEmpty())
						})
					})

					Context("when the object has succeeded", func() {
						It("returns the outputs", func() {
							template := templates.NewRunTemplateModel(apiTemplate)
							outputs, err := template.GetOutput(originalOutputs, stampedObjects)
							Expect(err).NotTo(HaveOccurred())
							Expect(outputs["simplistic"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"is a string"`)}))
							Expect(outputs["complexish"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`{"name":"complex object","type":"object"}`)}))
						})
					})
				})

				Context("with invalid output paths defined", func() {
					BeforeEach(func() {
						apiTemplate.Spec.Outputs = map[string]string{
							"complexish": "spec.nonexistant",
						}
					})
					It("returns an error", func() {
						template := templates.NewRunTemplateModel(apiTemplate)
						_, err := template.GetOutput(originalOutputs, stampedObjects)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(Equal("get output: evaluate: find results: nonexistant is not found"))
					})
				})
			})
		})

		Context("when the previous outputs had a value", func() {
			BeforeEach(func() {
				originalOutputs = map[string]apiextensionsv1.JSON{
					"simplistic": apiextensionsv1.JSON{Raw: []byte(`"original simple"`)},
				}
			})

			Context("when there is one object", func() {
				BeforeEach(func() {
					stampedObjects = []*unstructured.Unstructured{firstStampedObject}
				})

				Context("with no outputs", func() {
					It("returns an empty list", func() {
						template := templates.NewRunTemplateModel(apiTemplate)
						outputs, err := template.GetOutput(originalOutputs, stampedObjects)
						Expect(err).NotTo(HaveOccurred())
						Expect(outputs).To(BeEmpty())
					})
				})

				Context("with valid output paths defined", func() {
					BeforeEach(func() {
						apiTemplate.Spec.Outputs = map[string]string{
							"simplistic": "spec.simple",
							"complexish": "spec.complex",
						}
					})

					Context("when the object has not succeeded", func() {
						BeforeEach(func() {
							Expect(utils.AlterFieldOfNestedStringMaps(firstStampedObject.Object, "status.conditions.[0]status", "False")).To(Succeed()) // TODO: fix this notation or start using a jsonpath parser
						})
						It("returns the pipeline's previous output", func() {
							template := templates.NewRunTemplateModel(apiTemplate)
							outputs, err := template.GetOutput(originalOutputs, stampedObjects)
							Expect(err).NotTo(HaveOccurred())
							Expect(outputs["simplistic"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"original simple"`)}))
						})
					})

					It("returns the new outputs", func() {
						template := templates.NewRunTemplateModel(apiTemplate)
						outputs, err := template.GetOutput(originalOutputs, stampedObjects)
						Expect(err).NotTo(HaveOccurred())
						Expect(outputs["simplistic"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"is a string"`)}))
						Expect(outputs["complexish"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`{"name":"complex object","type":"object"}`)}))
					})
				})

				Context("with invalid output paths defined", func() {
					BeforeEach(func() {
						apiTemplate.Spec.Outputs = map[string]string{
							"complexish": "spec.nonexistant",
						}
					})
					It("returns an error", func() {
						template := templates.NewRunTemplateModel(apiTemplate)
						_, err := template.GetOutput(originalOutputs, stampedObjects)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(Equal("get output: evaluate: find results: nonexistant is not found"))
					})
				})
			})

			Context("when there are multiple objects", func() {
				BeforeEach(func() {
					stampedObjects = []*unstructured.Unstructured{secondStampedObject, firstStampedObject}

					apiTemplate.Spec.Outputs = map[string]string{
						"simplistic": "spec.simple",
						"complexish": "spec.complex",
					}
				})

				Context("when none have succeeded", func() {
					BeforeEach(func() {
						Expect(utils.AlterFieldOfNestedStringMaps(firstStampedObject.Object, "status.conditions.[0]status", "False")).To(Succeed())
						Expect(utils.AlterFieldOfNestedStringMaps(secondStampedObject.Object, "status.conditions.[0]status", "False")).To(Succeed())
					})
					It("returns the pipeline's previous output", func() {
						template := templates.NewRunTemplateModel(apiTemplate)
						outputs, err := template.GetOutput(originalOutputs, stampedObjects)
						Expect(err).NotTo(HaveOccurred())
						Expect(outputs["simplistic"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"original simple"`)}))
					})
				})
				Context("when only the least recently has succeeded", func() {
					BeforeEach(func() {
						Expect(utils.AlterFieldOfNestedStringMaps(secondStampedObject.Object, "status.conditions.[0]status", "False")).To(Succeed())
					})
					It("returns the output of the earlier submitted and successful object", func() {
						template := templates.NewRunTemplateModel(apiTemplate)
						outputs, err := template.GetOutput(originalOutputs, stampedObjects)
						Expect(err).NotTo(HaveOccurred())
						Expect(outputs["simplistic"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"is a string"`)}))
						Expect(outputs["complexish"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`{"name":"complex object","type":"object"}`)}))
					})
				})
				Context("when all have succeeded", func() {
					It("returns the output of the most recently submitted and successful object", func() {
						template := templates.NewRunTemplateModel(apiTemplate)
						outputs, err := template.GetOutput(originalOutputs, stampedObjects)
						Expect(err).NotTo(HaveOccurred())
						Expect(outputs["simplistic"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"2nd-simple"`)}))
						Expect(outputs["complexish"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"2nd-complex"`)}))
					})
				})
				Context("when the field of one object don't match the declared output fields", func() {
					BeforeEach(func() {
						apiTemplate.Spec.Outputs = map[string]string{
							"simplistic": "spec.only-exists-on-first-object",
						}
					})

					It("returns the output of the most recently submitted, successful, non-error inducing object", func() {
						template := templates.NewRunTemplateModel(apiTemplate)
						outputs, err := template.GetOutput(originalOutputs, stampedObjects)
						Expect(err).NotTo(HaveOccurred())
						Expect(outputs["simplistic"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"populated"`)}))
					})
				})
				Context("when the fields of all object don't match the declared output fields", func() {
					BeforeEach(func() {
						apiTemplate.Spec.Outputs = map[string]string{
							"simplistic": "spec.nonexistant",
						}
					})
					It("returns a helpful error", func() {
						template := templates.NewRunTemplateModel(apiTemplate)
						_, err := template.GetOutput(originalOutputs, stampedObjects)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(Equal("get output: evaluate: find results: nonexistant is not found"))
					})

					Context("and one does not have succeeded condition", func() {
						BeforeEach(func() {
							stampedObjects = []*unstructured.Unstructured{unconditionedStampedObject, firstStampedObject}
						})
						It("returns a helpful error", func() {
							template := templates.NewRunTemplateModel(apiTemplate)
							_, err := template.GetOutput(originalOutputs, stampedObjects)
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(Equal("get output: evaluate: find results: nonexistant is not found"))
						})
					})
				})
			})
		})
	})
})
