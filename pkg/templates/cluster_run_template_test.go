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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

func makeTemplate(outputs map[string]string) templates.ClusterRunTemplate {
	apiTemplate := &v1alpha1.ClusterRunTemplate{}
	apiTemplate.Spec.Outputs = outputs

	return templates.NewRunTemplateModel(apiTemplate)
}

var _ = FDescribe("ClusterRunTemplate", func() {
	Describe("GetLatestSuccessfulOutput", func() {
		var (
			serializer     runtime.Serializer
			template       templates.ClusterRunTemplate
			stampedObjects []*unstructured.Unstructured
		)
		BeforeEach(func() {
			serializer = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		})

		Context("No stamped objects", func() {
			BeforeEach(func() {
				template = makeTemplate(map[string]string{
					"an-output": "status.simple-result",
				})
				stampedObjects = []*unstructured.Unstructured{}
			})

			It("returns no output", func() {
				outputs, outputSourceObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
				Expect(err).NotTo(HaveOccurred())
				Expect(outputs).To(BeEmpty())
				Expect(outputSourceObject).To(BeNil())
			})
		})

		Context("One stamped object", func() {
			var stampedObject *unstructured.Unstructured

			BeforeEach(func() {
				template = makeTemplate(map[string]string{
					"an-output": "status.simple-result",
				})
			})

			Context("with no succeeded condition", func() {
				BeforeEach(func() {
					stampedObject = &unstructured.Unstructured{}
					stampedObjectYaml := utils.HereYamlF(`
						apiVersion: thing/v1
						kind: Thing
						metadata:
						  name: named-thing
						  namespace: somens
						  creationTimestamp: "2021-09-17T16:02:30Z"
						status: 
						  conditions: {}
					`)
					_, _, err := serializer.Decode([]byte(stampedObjectYaml), nil, stampedObject)
					Expect(err).NotTo(HaveOccurred())
					stampedObjects = []*unstructured.Unstructured{stampedObject}
				})

				It("returns no output", func() {
					outputs, outputSourceObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
					Expect(err).NotTo(HaveOccurred())
					Expect(outputs).To(BeEmpty())
					Expect(outputSourceObject).To(BeNil())
				})
			})

			Context("with a succeeded:false condition", func() {
				BeforeEach(func() {
					stampedObject = &unstructured.Unstructured{}
					stampedObjectYaml := utils.HereYamlF(`
						apiVersion: thing/v1
						kind: Thing
						metadata:
						  name: named-thing
						  namespace: somens
						  creationTimestamp: "2021-09-17T16:02:30Z"
						status: 
						  conditions:
							- type: Succeeded
							  status: "False"
					`)

					_, _, err := serializer.Decode([]byte(stampedObjectYaml), nil, stampedObject)
					Expect(err).NotTo(HaveOccurred())
					stampedObjects = []*unstructured.Unstructured{stampedObject}
				})

				It("returns no output", func() {
					outputs, outputSourceObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
					Expect(err).NotTo(HaveOccurred())
					Expect(outputs).To(BeEmpty())
					Expect(outputSourceObject).To(BeNil())
				})
			})

			Context("with a succeeded:true condition", func() {
				BeforeEach(func() {
					stampedObject = &unstructured.Unstructured{}
					stampedObjectYaml := utils.HereYamlF(`
						apiVersion: thing/v1
						kind: Thing
						metadata:
						  name: named-thing
						  namespace: somens
						  creationTimestamp: "2021-09-17T16:02:30Z"
						status: 
						  conditions:
							- type: Succeeded
							  status: "True"
						  simple-result: a thing
					`)

					_, _, err := serializer.Decode([]byte(stampedObjectYaml), nil, stampedObject)
					Expect(err).NotTo(HaveOccurred())
					stampedObjects = []*unstructured.Unstructured{stampedObject}
				})

				Context("that does not match the outputs", func() {
					BeforeEach(func() {
						template = makeTemplate(map[string]string{
							"an-output": "status.nonexistant",
						})
					})

					It("returns no output, the matching object and an error", func() {
						outputs, outputSourceObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
						Expect(err).To(MatchError("failed to evaluate path [status.nonexistant]: jsonpath returned empty list: status.nonexistant"))
						Expect(outputs).To(BeEmpty())
						Expect(outputSourceObject).To(Equal(stampedObjects[0]))
					})
				})

				Context("no output specified in the template", func() {
					BeforeEach(func() {
						template = makeTemplate(map[string]string{})
					})
					It("returns an empty output and the matched object", func() {
						outputs, outputSourceObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
						Expect(err).NotTo(HaveOccurred())
						Expect(outputs).To(BeEmpty())
						Expect(outputSourceObject).To(Equal(stampedObjects[0]))
					})
				})

				Context("that matches the outputs", func() {
					It("returns the outputs and the matched object", func() {
						outputs, outputSourceObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
						Expect(err).NotTo(HaveOccurred())
						Expect(outputs["an-output"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"a thing"`)}))
						Expect(outputSourceObject).To(Equal(stampedObjects[0]))
					})
				})
			})
		})

		Context("two stamped objects", func() {
			Context("with no conditions", func() {
				BeforeEach(func() {
					firstObject := &unstructured.Unstructured{}
					firstObjectYaml := utils.HereYamlF(`
						apiVersion: thing/v1
						kind: Thing
						metadata:
						  name: first-thing
						  namespace: somens
						  creationTimestamp: "2021-09-17T16:02:30Z"
						status: 
						  conditions: {}
					`)
					_, _, err := serializer.Decode([]byte(firstObjectYaml), nil, firstObject)
					Expect(err).NotTo(HaveOccurred())

					secondObject := &unstructured.Unstructured{}
					secondObjectYaml := utils.HereYamlF(`
						apiVersion: thing/v1
						kind: Thing
						metadata:
						  name: second-thing
						  namespace: somens
						  creationTimestamp: "2021-09-17T17:02:30Z"
						status: 
						  conditions: {}
					`)
					_, _, err = serializer.Decode([]byte(secondObjectYaml), nil, secondObject)
					Expect(err).NotTo(HaveOccurred())

					// Out of order deliberately
					stampedObjects = []*unstructured.Unstructured{secondObject, firstObject}
				})

				It("returns no output", func() {
					outputs, outputSourceObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
					Expect(err).NotTo(HaveOccurred())
					Expect(outputs).To(BeEmpty())
					Expect(outputSourceObject).To(BeNil())
				})
			})

			Context("with [succeeded:false, succeeded:false] conditions", func() {
				BeforeEach(func() {
					firstObject := &unstructured.Unstructured{}
					firstObjectYaml := utils.HereYamlF(`
						apiVersion: thing/v1
						kind: Thing
						metadata:
						  name: first-thing
						  namespace: somens
						  creationTimestamp: "2021-09-17T16:02:30Z"
						status: 
						  conditions:
							- type: Succeeded
							  status: "False"
						  simple-result: first result
					`)
					_, _, err := serializer.Decode([]byte(firstObjectYaml), nil, firstObject)
					Expect(err).NotTo(HaveOccurred())

					secondObject := &unstructured.Unstructured{}
					secondObjectYaml := utils.HereYamlF(`
						apiVersion: thing/v1
						kind: Thing
						metadata:
						  name: second-thing
						  namespace: somens
						  creationTimestamp: "2021-09-17T17:02:30Z"
						status: 
						  conditions:
							- type: Succeeded
							  status: "False"
						  simple-result: second result
					`)
					_, _, err = serializer.Decode([]byte(secondObjectYaml), nil, secondObject)
					Expect(err).NotTo(HaveOccurred())

					// Out of order deliberately
					stampedObjects = []*unstructured.Unstructured{secondObject, firstObject}
				})

				It("returns no output", func() {
					outputs, outputSourceObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
					Expect(err).NotTo(HaveOccurred())
					Expect(outputs).To(BeEmpty())
					Expect(outputSourceObject).To(BeNil())
				})
			})

			Context("with [succeeded:true, succeeded:false] conditions", func() {
				var firstObject, secondObject *unstructured.Unstructured
				BeforeEach(func() {
					firstObject = &unstructured.Unstructured{}
					firstObjectYaml := utils.HereYamlF(`
						apiVersion: thing/v1
						kind: Thing
						metadata:
						  name: first-thing
						  namespace: somens
						  creationTimestamp: "2021-09-17T16:02:30Z"
						status: 
						  conditions:
							- type: Succeeded
							  status: "True"
						  simple-result: first result
					`)
					_, _, err := serializer.Decode([]byte(firstObjectYaml), nil, firstObject)
					Expect(err).NotTo(HaveOccurred())

					secondObject = &unstructured.Unstructured{}
					secondObjectYaml := utils.HereYamlF(`
						apiVersion: thing/v1
						kind: Thing
						metadata:
						  name: second-thing
						  namespace: somens
						  creationTimestamp: "2021-09-17T17:02:30Z"
						status: 
						  conditions:
							- type: Succeeded
							  status: "False"
						  simple-result: second result
					`)
					_, _, err = serializer.Decode([]byte(secondObjectYaml), nil, secondObject)
					Expect(err).NotTo(HaveOccurred())

					// Out of order deliberately
					stampedObjects = []*unstructured.Unstructured{secondObject, firstObject}
				})

				Context("with no output specified in the template", func() {
					BeforeEach(func() {
						template = makeTemplate(map[string]string{})
					})

					It("returns the empty outputs and the matched object", func() {
						outputs, outputSourceObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
						Expect(err).NotTo(HaveOccurred())
						Expect(outputs).To(BeEmpty())
						Expect(outputSourceObject).To(Equal(firstObject))
					})
				})

				Context("that do not match the outputs in the template", func() {
					BeforeEach(func() {
						template = makeTemplate(map[string]string{
							"an-output": "status.nonexistant",
						})
					})

					It("returns no output, the matching object and an error", func() {
						outputs, outputSourceObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
						Expect(err).To(MatchError("failed to evaluate path [status.nonexistant]: jsonpath returned empty list: status.nonexistant"))
						Expect(outputs).To(BeEmpty())
						Expect(outputSourceObject).To(Equal(firstObject))
					})

				})

				Context("that matches the outputs in the template", func() {
					BeforeEach(func() {
						template = makeTemplate(map[string]string{
							"an-output": "status.simple-result",
						})
					})

					It("returns the earliest matched outputs and the earliest matched object", func() {
						outputs, outputSourceObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
						Expect(err).NotTo(HaveOccurred())
						Expect(outputs["an-output"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"first result"`)}))
						Expect(outputSourceObject).To(Equal(firstObject))
					})
				})

			})

			Context("with [succeeded:true, succeeded:true] conditions", func() {
				var firstObject, secondObject *unstructured.Unstructured

				BeforeEach(func() {
					firstObject = &unstructured.Unstructured{}
					firstObjectYaml := utils.HereYamlF(`
						apiVersion: thing/v1
						kind: Thing
						metadata:
						  name: first-thing
						  namespace: somens
						  creationTimestamp: "2021-09-17T16:02:30Z"
						status: 
						  conditions:
							- type: Succeeded
							  status: "True"
						  simple-result: first result
					      first-only: first only result
					`)
					_, _, err := serializer.Decode([]byte(firstObjectYaml), nil, firstObject)
					Expect(err).NotTo(HaveOccurred())

					secondObject = &unstructured.Unstructured{}
					secondObjectYaml := utils.HereYamlF(`
						apiVersion: thing/v1
						kind: Thing
						metadata:
						  name: second-thing
						  namespace: somens
						  creationTimestamp: "2021-09-17T17:02:30Z"
						status: 
						  conditions:
							- type: Succeeded
							  status: "True"
						  simple-result: second result
					      second-only: second only result
					`)
					_, _, err = serializer.Decode([]byte(secondObjectYaml), nil, secondObject)
					Expect(err).NotTo(HaveOccurred())

					// Out of order deliberately
					stampedObjects = []*unstructured.Unstructured{secondObject, firstObject}

				})
				Context("with no output specified in the template", func() {
					BeforeEach(func() {
						template = makeTemplate(map[string]string{})
					})

					It("returns an empty output and the latest matched object", func() {
						outputs, outputSourceObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
						Expect(err).NotTo(HaveOccurred())
						Expect(outputs).To(BeEmpty())
						Expect(outputSourceObject).To(Equal(secondObject))
					})
				})

				Context("neither match the outputs", func() {
					BeforeEach(func() {
						template = makeTemplate(map[string]string{
							"an-output": "status.nonexistant",
						})
					})

					It("returns an error and the latest object", func() {
						outputs, outputSourceObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
						Expect(err).To(MatchError("failed to evaluate path [status.nonexistant]: jsonpath returned empty list: status.nonexistant"))
						Expect(outputs).To(BeEmpty())
						Expect(outputSourceObject).To(Equal(secondObject))
					})
				})

				Context("later does not match the outputs", func() {
					BeforeEach(func() {
						template = makeTemplate(map[string]string{
							"an-output": "status.first-only",
						})
					})
					It("returns an error and the latest object", func() {
						outputs, outputSourceObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
						Expect(err).To(MatchError("failed to evaluate path [status.first-only]: jsonpath returned empty list: status.first-only"))
						Expect(outputs).To(BeEmpty())
						Expect(outputSourceObject).To(Equal(secondObject))
					})
				})

				Context("earlier does not match the outputs", func() {
					BeforeEach(func() {
						template = makeTemplate(map[string]string{
							"an-output": "status.second-only",
						})
					})

					It("returns the latest", func() {
						outputs, outputSourceObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
						Expect(err).NotTo(HaveOccurred())
						Expect(outputs["an-output"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"second only result"`)}))
						Expect(outputSourceObject).To(Equal(secondObject))
					})
				})

				Context("both match the outputs", func() {
					BeforeEach(func() {
						template = makeTemplate(map[string]string{
							"an-output": "status.simple-result",
						})
					})

					It("returns the latest matched output and the latest matched object", func() {
						outputs, outputSourceObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
						Expect(err).NotTo(HaveOccurred())
						Expect(outputs["an-output"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"second result"`)}))
						Expect(outputSourceObject).To(Equal(secondObject))
					})
				})
			})
		})

		Describe("supports complex output objects", func() {
			var stampedObject *unstructured.Unstructured

			BeforeEach(func() {
				template = makeTemplate(map[string]string{
					"my-complex-output": "status.complex-result",
				})

				stampedObject = &unstructured.Unstructured{}
				stampedObjectYaml := utils.HereYamlF(`
						apiVersion: thing/v1
						kind: Thing
						metadata:
						  name: named-thing
						  namespace: somens
						  creationTimestamp: "2021-09-17T16:02:30Z"
						status: 
						  conditions:
							- type: Succeeded
							  status: "True"
						  complex-result:
							- name: item1
							  value: 
								field1: one
								field2: two
							- name: item2
							  value: a string
					`)

				_, _, err := serializer.Decode([]byte(stampedObjectYaml), nil, stampedObject)
				Expect(err).NotTo(HaveOccurred())
				stampedObjects = []*unstructured.Unstructured{stampedObject}
			})

			It("returns the output", func() {
				outputs, _, err := template.GetLatestSuccessfulOutput(stampedObjects)
				Expect(err).NotTo(HaveOccurred())

				Expect(outputs["my-complex-output"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"[{\"name\":\"item1\",\"value\":{\"field1\":\"one\",\"field2\":\"two\"}},{\"name\":\"item2\",\"value\":\"a string\"}]"`)}))
			})

		})
	})
})
