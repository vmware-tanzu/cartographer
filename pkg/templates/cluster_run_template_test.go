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
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

var _ = Describe("ClusterRunTemplate", func() {
	FDescribe("GetLatestSuccessfulOutput", func() {
		var (
			//	apiTemplate                                                         *v1alpha1.ClusterRunTemplate
			//	template                                                            templates.ClusterRunTemplate
			//	firstStampedObject, secondStampedObject, unconditionedStampedObject *unstructured.Unstructured
			//	stampedObjects                                                      []*unstructured.Unstructured
			serializer runtime.Serializer
		)
		//
		BeforeEach(func() {
			serializer = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
			//
			//	apiTemplate = &v1alpha1.ClusterRunTemplate{}
			//
			//	firstStampedObject = &unstructured.Unstructured{}
			//	firstStampedObjectManifest := utils.HereYamlF(`
			//		apiVersion: thing/v1
			//		kind: Thing
			//		metadata:
			//		  name: named-thing
			//		  namespace: somens
			//		  creationTimestamp: "2021-09-17T16:02:30Z"
			//		spec:
			//		  simple: is a string
			//		  complex:
			//			type: object
			//			name: complex object
			//		  only-exists-on-first-object: populated
			//		status:
			//		  conditions:
			//		    - type: Succeeded
			//		      status: "True"
			//	`)
			//
			//	_, _, err := serializer.Decode([]byte(firstStampedObjectManifest), nil, firstStampedObject)
			//	Expect(err).NotTo(HaveOccurred())
			//
			//	secondStampedObject = &unstructured.Unstructured{}
			//	secondStampedObjectManifest := utils.HereYamlF(`
			//		apiVersion: thing/v1
			//		kind: Thing
			//		metadata:
			//		  name: named-thing
			//		  namespace: somens
			//		  creationTimestamp: "2021-09-17T16:02:40Z"
			//		spec:
			//		  simple: 2nd-simple
			//		  complex: 2nd-complex
			//		status:
			//		  conditions:
			//		    - type: Succeeded
			//		      status: "True"
			//	`)
			//
			//	_, _, err = serializer.Decode([]byte(secondStampedObjectManifest), nil, secondStampedObject)
			//	Expect(err).NotTo(HaveOccurred())
			//
			//	unconditionedStampedObject = &unstructured.Unstructured{}
			//	unconditionedStampedObjectManifest := utils.HereYamlF(`
			//		apiVersion: thing/v1
			//		kind: Thing
			//		metadata:
			//		  name: named-thing
			//		  namespace: somens
			//		spec:
			//		  simple: val
			//		  complex: other-val
			//	`)
			//
			//	_, _, err = serializer.Decode([]byte(unconditionedStampedObjectManifest), nil, unconditionedStampedObject)
			//	Expect(err).NotTo(HaveOccurred())
		})

		//Context("when there is one object", func() {
		//	BeforeEach(func() {
		//		stampedObjects = []*unstructured.Unstructured{firstStampedObject}
		//	})
		//
		//	Context("with no outputs", func() {
		//		It("returns an empty list", func() {
		//			template := templates.NewRunTemplateModel(apiTemplate)
		//			outputs, evaluatedStampedObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
		//			Expect(err).NotTo(HaveOccurred())
		//			Expect(outputs).To(BeEmpty())
		//			Expect(evaluatedStampedObject).To(Equal(firstStampedObject))
		//		})
		//	})
		//
		//	Context("with valid output paths defined", func() {
		//		BeforeEach(func() {
		//			apiTemplate.Spec.Outputs = map[string]string{
		//				"simplistic": "spec.simple",
		//				"complexish": "spec.complex",
		//			}
		//		})
		//
		//		Context("when the object has not succeeded", func() {
		//			BeforeEach(func() {
		//				Expect(utils.AlterFieldOfNestedStringMaps(firstStampedObject.Object, "status.conditions.[0]status", "False")).To(Succeed()) // TODO: fix this notation or start using a jsonpath parser
		//			})
		//			It("returns empty outputs", func() {
		//				template := templates.NewRunTemplateModel(apiTemplate)
		//				outputs, evaluatedStampedObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
		//				Expect(err).NotTo(HaveOccurred())
		//				Expect(outputs).To(BeEmpty())
		//				Expect(evaluatedStampedObject).To(BeNil())
		//			})
		//		})
		//		Context("when the object has no conditions", func() {
		//			BeforeEach(func() {
		//				firstStampedObject = &unstructured.Unstructured{}
		//				firstStampedObjectManifest := utils.HereYamlF(`
		//					apiVersion: thing/v1
		//					kind: Thing
		//					metadata:
		//					  name: named-thing
		//					  namespace: somens
		//					  creationTimestamp: "2021-09-17T16:02:30Z"
		//					spec:
		//					  simple: is a string
		//					  complex:
		//						type: object
		//						name: complex object
		//					  only-exists-on-first-object: populated
		//					status: {}
		//				`)
		//
		//				_, _, err := serializer.Decode([]byte(firstStampedObjectManifest), nil, firstStampedObject)
		//				Expect(err).NotTo(HaveOccurred())
		//			})
		//			It("returns empty outputs", func() {
		//				template := templates.NewRunTemplateModel(apiTemplate)
		//				outputs, evaluatedStampedObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
		//				Expect(err).NotTo(HaveOccurred())
		//				Expect(outputs).To(BeEmpty())
		//				Expect(evaluatedStampedObject).To(BeNil())
		//			})
		//		})
		//
		//		Context("when the object has no succeeded condition", func() {
		//			BeforeEach(func() {
		//				firstStampedObject = &unstructured.Unstructured{}
		//				firstStampedObjectManifest := utils.HereYamlF(`
		//					apiVersion: thing/v1
		//					kind: Thing
		//					metadata:
		//					  name: named-thing
		//					  namespace: somens
		//					  creationTimestamp: "2021-09-17T16:02:30Z"
		//					spec:
		//					  simple: is a string
		//					  complex:
		//						type: object
		//						name: complex object
		//					  only-exists-on-first-object: populated
		//					status:
		//				      conditions:
		//						- type: Fredded
		//						  status: True
		//				`)
		//
		//				_, _, err := serializer.Decode([]byte(firstStampedObjectManifest), nil, firstStampedObject)
		//				Expect(err).NotTo(HaveOccurred())
		//			})
		//			It("returns empty outputs", func() {
		//				template := templates.NewRunTemplateModel(apiTemplate)
		//				outputs, evaluatedStampedObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
		//				Expect(err).NotTo(HaveOccurred())
		//				Expect(outputs).To(BeEmpty())
		//				Expect(evaluatedStampedObject).To(BeNil())
		//			})
		//		})
		//
		//		It("returns the new outputs", func() {
		//			template := templates.NewRunTemplateModel(apiTemplate)
		//			outputs, evaluatedStampedObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
		//			Expect(err).NotTo(HaveOccurred())
		//			Expect(outputs["simplistic"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"is a string"`)}))
		//			Expect(outputs["complexish"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`{"name":"complex object","type":"object"}`)}))
		//			Expect(evaluatedStampedObject).To(Equal(firstStampedObject))
		//		})
		//	})
		//
		//	Context("with invalid output paths defined", func() {
		//		BeforeEach(func() {
		//			apiTemplate.Spec.Outputs = map[string]string{
		//				"complexish": "spec.nonexistant",
		//			}
		//		})
		//		It("returns an error", func() {
		//			template := templates.NewRunTemplateModel(apiTemplate)
		//			_, _, err := template.GetLatestSuccessfulOutput(stampedObjects)
		//			Expect(err).To(HaveOccurred())
		//			Expect(err.Error()).To(Equal("failed to evaluate path [spec.nonexistant]: jsonpath returned empty list: spec.nonexistant"))
		//		})
		//
		//		It("has no outputs", func() {
		//			template := templates.NewRunTemplateModel(apiTemplate)
		//			outputs, _, _ := template.GetLatestSuccessfulOutput(stampedObjects)
		//			Expect(outputs).To(BeEmpty())
		//		})
		//	})
		//})
		//
		//Context("when there are multiple objects", func() {
		//	BeforeEach(func() {
		//		stampedObjects = []*unstructured.Unstructured{secondStampedObject, firstStampedObject}
		//
		//		apiTemplate.Spec.Outputs = map[string]string{
		//			"simplistic": "spec.simple",
		//			"complexish": "spec.complex",
		//		}
		//	})
		//
		//	Context("when none have succeeded", func() {
		//		BeforeEach(func() {
		//			Expect(utils.AlterFieldOfNestedStringMaps(firstStampedObject.Object, "status.conditions.[0]status", "False")).To(Succeed())
		//			Expect(utils.AlterFieldOfNestedStringMaps(secondStampedObject.Object, "status.conditions.[0]status", "False")).To(Succeed())
		//		})
		//		It("returns empty outputs", func() {
		//			template := templates.NewRunTemplateModel(apiTemplate)
		//			outputs, evaluatedStampedObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
		//			Expect(err).NotTo(HaveOccurred())
		//			Expect(outputs).To(BeEmpty())
		//			Expect(evaluatedStampedObject).To(BeNil())
		//		})
		//	})
		//	Context("when only the least recently has succeeded", func() {
		//		BeforeEach(func() {
		//			Expect(utils.AlterFieldOfNestedStringMaps(secondStampedObject.Object, "status.conditions.[0]status", "False")).To(Succeed())
		//		})
		//		It("returns the output of the earlier submitted and successful object", func() {
		//			template := templates.NewRunTemplateModel(apiTemplate)
		//			outputs, evaluatedStampedObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
		//			Expect(err).NotTo(HaveOccurred())
		//			Expect(outputs["simplistic"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"is a string"`)}))
		//			Expect(outputs["complexish"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`{"name":"complex object","type":"object"}`)}))
		//			Expect(evaluatedStampedObject).To(Equal(firstStampedObject))
		//		})
		//	})
		//	Context("when all have succeeded", func() {
		//		It("returns the output of the most recently submitted and successful object", func() {
		//			template := templates.NewRunTemplateModel(apiTemplate)
		//			outputs, evaluatedStampedObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
		//			Expect(err).NotTo(HaveOccurred())
		//			Expect(outputs["simplistic"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"2nd-simple"`)}))
		//			Expect(outputs["complexish"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"2nd-complex"`)}))
		//			Expect(evaluatedStampedObject).To(Equal(secondStampedObject))
		//		})
		//	})
		//	Context("when the field of one object don't match the declared output fields", func() {
		//		BeforeEach(func() {
		//			apiTemplate.Spec.Outputs = map[string]string{
		//				"simplistic": "spec.only-exists-on-first-object",
		//			}
		//		})
		//
		//		It("returns the output of the most recently submitted, successful, non-error inducing object", func() {
		//			template := templates.NewRunTemplateModel(apiTemplate)
		//			outputs, evaluatedStampedObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
		//			Expect(err).NotTo(HaveOccurred())
		//			Expect(outputs["simplistic"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"populated"`)}))
		//			Expect(evaluatedStampedObject).To(Equal(firstStampedObject))
		//		})
		//	})
		//	Context("when the fields of all objects don't match the declared output fields", func() {
		//		BeforeEach(func() {
		//			apiTemplate.Spec.Outputs = map[string]string{
		//				"simplistic": "spec.nonexistant",
		//			}
		//		})
		//		It("returns a helpful error", func() {
		//			template := templates.NewRunTemplateModel(apiTemplate)
		//			_, _, err := template.GetLatestSuccessfulOutput(stampedObjects)
		//			Expect(err).To(HaveOccurred())
		//			Expect(err.Error()).To(Equal("failed to evaluate path [spec.nonexistant]: jsonpath returned empty list: spec.nonexistant"))
		//		})
		//
		//		Context("and one does not have succeeded condition", func() {
		//			BeforeEach(func() {
		//				stampedObjects = []*unstructured.Unstructured{unconditionedStampedObject, firstStampedObject}
		//			})
		//			It("returns a helpful error", func() {
		//				template := templates.NewRunTemplateModel(apiTemplate)
		//				_, _, err := template.GetLatestSuccessfulOutput(stampedObjects)
		//				Expect(err).To(HaveOccurred())
		//				Expect(err.Error()).To(ContainSubstring("failed to evaluate path [spec.nonexistant]: jsonpath returned empty list: spec.nonexistant"))
		//			})
		//		})
		//	})
		//})

		// ----- reworking
		var (
			template       templates.ClusterRunTemplate
			stampedObjects []*unstructured.Unstructured
		)

		var makeTemplate = func(outputs map[string]string) templates.ClusterRunTemplate {
			apiTemplate := &v1alpha1.ClusterRunTemplate{}
			apiTemplate.Spec.Outputs = outputs

			return templates.NewRunTemplateModel(apiTemplate)
		}

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

				It("returns no output", func() { // Todo: returning an error it shouldn't
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

					// You might expect an error here, but you cant
					// it's always possible that the spec in the template changed, and so the latest output
					// becomes invalid until the latest stamped object is success: true
					It("returns no output, but the matching object", func() {
						outputs, outputSourceObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
						Expect(err).NotTo(HaveOccurred())
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

					It("returns the no outputs and the matched object", func() {
						outputs, outputSourceObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
						Expect(err).NotTo(HaveOccurred())
						Expect(outputs).To(BeEmpty())
						Expect(outputSourceObject).To(Equal(firstObject))
					})
				})

				//Context("that do not match the outputs in the template", func() {
				//	It("returns an error", func() {})
				//})
				//
				Context("that matches the outputs in the template", func() {
					BeforeEach(func() {
						template = makeTemplate(map[string]string{
							"an-output": "status.simple-result",
						})
						stampedObjects = []*unstructured.Unstructured{}
					})
					It("returns the earliest matched outputs and the earliest matched object", func() {
						outputs, outputSourceObject, err := template.GetLatestSuccessfulOutput(stampedObjects)
						Expect(err).NotTo(HaveOccurred())
						Expect(outputs["an-output"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"first result"`)}))
						Expect(outputSourceObject).To(Equal(firstObject))
					})
				})

			})
			//
			//Context("with [succeeded:true, succeeded:true] conditions", func() {
			//	Context("with no output specified in the template", func() {
			//		It("returns an empty output and the matched object", func() {})
			//	})
			//
			//	Context("neither match the outputs", func() {
			//		It("returns an error", func() {})
			//	})
			//
			//	Context("both match the outputs", func() {
			//		It("returns the latest matched outputs and the latest matched object", func() {})
			//	})
			//
			//	Context("the earliest matches the outputs, the latest does not", func() {
			//		It("returns the earliest matched outputs and the earliest matched object", func() {})
			//	})
			//})
		})
	})
})
