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

package stamp_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/stamp"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

type noInputImpl struct{}

func (noInputImpl) GetDeployment() *templates.SourceInput {
	return nil
}

type deploymentInputImpl struct{}

func (deploymentInputImpl) GetDeployment() *templates.SourceInput {
	return &templates.SourceInput{
		URL:      "my-url",
		Revision: "my-revision",
		Name:     "my-resource",
	}
}

var _ = Describe("Reader", func() {

	Context("using a source reader", func() {
		var (
			template *v1alpha1.ClusterSourceTemplate
			reader   stamp.Reader
		)

		BeforeEach(func() {
			template = &v1alpha1.ClusterSourceTemplate{
				Spec: v1alpha1.SourceTemplateSpec{
					TemplateSpec: v1alpha1.TemplateSpec{},
					URLPath:      ".data.url",
					RevisionPath: ".data.revision",
				},
			}

			var err error
			reader, err = stamp.NewReader(template, noInputImpl{})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("where the evaluator can return a value", func() {
			var stampedObject *unstructured.Unstructured

			BeforeEach(func() {
				unstructuredContent := map[string]interface{}{
					"data": map[string]interface{}{
						"url":      "my-url",
						"revision": "my-revision",
					},
				}

				stampedObject = &unstructured.Unstructured{}
				stampedObject.SetUnstructuredContent(unstructuredContent)
			})

			It("returns the output", func() {
				output, err := reader.GetOutput(stampedObject)
				Expect(err).NotTo(HaveOccurred())
				Expect(output.Source.URL).To(Equal("my-url"))
				Expect(output.Source.Revision).To(Equal("my-revision"))
			})
		})

		Context("where the evaluator can not return a value", func() {
			var stampedObject *unstructured.Unstructured

			BeforeEach(func() {
				unstructuredContent := map[string]interface{}{}

				stampedObject = &unstructured.Unstructured{}
				stampedObject.SetUnstructuredContent(unstructuredContent)
			})

			It("returns a nil output", func() {
				output, _ := reader.GetOutput(stampedObject)
				Expect(output).To(BeNil())
			})

			It("returns an error", func() {
				_, err := reader.GetOutput(stampedObject)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to evaluate json path"))
				Expect(err.Error()).To(ContainSubstring(".data.url"))
			})
		})
	})

	Context("using an image reader", func() {
		var (
			template *v1alpha1.ClusterImageTemplate
			reader   stamp.Reader
		)

		BeforeEach(func() {
			template = &v1alpha1.ClusterImageTemplate{
				Spec: v1alpha1.ImageTemplateSpec{
					TemplateSpec: v1alpha1.TemplateSpec{},
					ImagePath:    ".data.image",
				},
			}

			var err error
			reader, err = stamp.NewReader(template, noInputImpl{})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("where the evaluator can return a value", func() {
			var stampedObject *unstructured.Unstructured

			BeforeEach(func() {
				unstructuredContent := map[string]interface{}{
					"data": map[string]interface{}{
						"image": "my-image",
					},
				}

				stampedObject = &unstructured.Unstructured{}
				stampedObject.SetUnstructuredContent(unstructuredContent)
			})

			It("returns the output", func() {
				output, err := reader.GetOutput(stampedObject)
				Expect(err).NotTo(HaveOccurred())
				Expect(output.Image).To(Equal("my-image"))
			})
		})

		Context("where the evaluator can not return a value", func() {
			var stampedObject *unstructured.Unstructured

			BeforeEach(func() {
				unstructuredContent := map[string]interface{}{}

				stampedObject = &unstructured.Unstructured{}
				stampedObject.SetUnstructuredContent(unstructuredContent)
			})

			It("returns a nil output", func() {
				output, _ := reader.GetOutput(stampedObject)
				Expect(output).To(BeNil())
			})

			It("returns an error", func() {
				_, err := reader.GetOutput(stampedObject)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to evaluate json path"))
				Expect(err.Error()).To(ContainSubstring(".data.image"))
			})
		})
	})

	Context("using a config reader", func() {
		var (
			template *v1alpha1.ClusterConfigTemplate
			reader   stamp.Reader
		)

		BeforeEach(func() {
			template = &v1alpha1.ClusterConfigTemplate{
				Spec: v1alpha1.ConfigTemplateSpec{
					TemplateSpec: v1alpha1.TemplateSpec{},
					ConfigPath:   ".data.config",
				},
			}

			var err error
			reader, err = stamp.NewReader(template, noInputImpl{})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("where the evaluator can return a value", func() {
			var stampedObject *unstructured.Unstructured

			BeforeEach(func() {
				unstructuredContent := map[string]interface{}{
					"data": map[string]interface{}{
						"config": "my-config",
					},
				}

				stampedObject = &unstructured.Unstructured{}
				stampedObject.SetUnstructuredContent(unstructuredContent)
			})

			It("returns the output", func() {
				output, err := reader.GetOutput(stampedObject)
				Expect(err).NotTo(HaveOccurred())
				Expect(output.Config).To(Equal("my-config"))
			})
		})

		Context("where the evaluator can not return a value", func() {
			var stampedObject *unstructured.Unstructured

			BeforeEach(func() {
				unstructuredContent := map[string]interface{}{}

				stampedObject = &unstructured.Unstructured{}
				stampedObject.SetUnstructuredContent(unstructuredContent)
			})

			It("returns a nil output", func() {
				output, _ := reader.GetOutput(stampedObject)
				Expect(output).To(BeNil())
			})

			It("returns an error", func() {
				_, err := reader.GetOutput(stampedObject)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to evaluate json path"))
				Expect(err.Error()).To(ContainSubstring(".data.config"))
			})
		})
	})

	Context("using a deployment reader", func() {
		var (
			template      *v1alpha1.ClusterDeploymentTemplate
			reader        stamp.Reader
			stampedObject *unstructured.Unstructured
		)

		BeforeEach(func() {
			template = &v1alpha1.ClusterDeploymentTemplate{
				Spec: v1alpha1.DeploymentSpec{},
			}
		})

		Context("observedCompletion", func() {
			var unstructuredContent map[string]interface{}

			BeforeEach(func() {
				template = &v1alpha1.ClusterDeploymentTemplate{
					Spec: v1alpha1.DeploymentSpec{
						ObservedCompletion: &v1alpha1.ObservedCompletion{
							SucceededCondition: v1alpha1.Condition{
								Key:   "completion.path",
								Value: "All Good",
							},
						},
					},
				}
			})

			Context("stampedObject has reconciled (generation == observedGeneration)", func() {

				BeforeEach(func() {
					unstructuredContent = map[string]interface{}{
						"metadata": map[string]interface{}{
							"generation": 1,
						},
						"status": map[string]interface{}{
							"observedGeneration": 1,
						},
						"completion": map[string]interface{}{
							"path": "All Good",
						},
						"failure": map[string]interface{}{
							"path": "some sad path value",
						},
					}

					stampedObject = &unstructured.Unstructured{}
					stampedObject.SetUnstructuredContent(unstructuredContent)
				})

				Context("success criterion is met", func() {

					Context("where the deployment is found", func() {
						BeforeEach(func() {
							var err error
							reader, err = stamp.NewReader(template, deploymentInputImpl{})
							Expect(err).NotTo(HaveOccurred())
						})

						It("returns the output", func() {
							output, err := reader.GetOutput(stampedObject)
							Expect(err).NotTo(HaveOccurred())
							Expect(output.Source.URL).To(Equal("my-url"))
							Expect(output.Source.Revision).To(Equal("my-revision"))
						})

						Context("failure criterion is set", func() {
							BeforeEach(func() {
								template.Spec.ObservedCompletion.FailedCondition = &v1alpha1.Condition{
									Key:   "failure.path",
									Value: "some sad path value",
								}
							})

							Context("failure criterion is met", func() {
								It("returns an error", func() {
									_, err := reader.GetOutput(stampedObject)
									Expect(err).To(HaveOccurred())
									Expect(err.Error()).To(ContainSubstring("deployment failure condition [failure.path] was: some sad path value"))
								})
							})

							Context("failure criterion path exists but is not met", func() {
								BeforeEach(func() {
									template.Spec.ObservedCompletion.FailedCondition = &v1alpha1.Condition{
										Key:   "failure.path",
										Value: "some other sad path value",
									}
								})
								It("returns the output", func() {
									output, err := reader.GetOutput(stampedObject)
									Expect(err).NotTo(HaveOccurred())

									Expect(output.Source.URL).To(Equal("my-url"))
									Expect(output.Source.Revision).To(Equal("my-revision"))
								})
							})

							Context("failure criterion path does not exist", func() {
								BeforeEach(func() {
									template.Spec.ObservedCompletion.FailedCondition = &v1alpha1.Condition{
										Key:   "failure.path-does-not-exist",
										Value: "some sad path value",
									}
								})
								It("returns the output", func() {
									output, err := reader.GetOutput(stampedObject)
									Expect(err).NotTo(HaveOccurred())

									Expect(output.Source.URL).To(Equal("my-url"))
									Expect(output.Source.Revision).To(Equal("my-revision"))
								})
							})

							Context("evaluating failure criterion path errors", func() {
								XIt("returns an error", func() {
									// TODO: I'm not sure how to force a non JsonPathDoesNotExistError error
									_, err := reader.GetOutput(stampedObject)
									Expect(err).To(HaveOccurred())
									Expect(err.Error()).To(ContainSubstring("failed to evaluate"))
								})
							})
						})
					})

					Context("where the deployment is not found", func() {
						BeforeEach(func() {
							var err error
							reader, err = stamp.NewReader(template, noInputImpl{})
							Expect(err).NotTo(HaveOccurred())
						})

						It("returns an error", func() {
							_, err := reader.GetOutput(stampedObject)
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("deployment not found in upstream template"))
						})
					})
				})

				Context("success criterion is not met", func() {

					BeforeEach(func() {
						unstructuredContent["completion"].(map[string]interface{})["path"] = "some sad path value"
						stampedObject = &unstructured.Unstructured{}
						stampedObject.SetUnstructuredContent(unstructuredContent)

						var err error
						reader, err = stamp.NewReader(template, deploymentInputImpl{})
						Expect(err).NotTo(HaveOccurred())
					})

					It("returns an error", func() {
						_, err := reader.GetOutput(stampedObject)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("deployment success condition [completion.path] was: some sad path value, expected: All Good"))
					})

					Context("success criterion path does not exist", func() {
						BeforeEach(func() {
							unstructuredContent["completion"] = "some sad path value"
							stampedObject = &unstructured.Unstructured{}
							stampedObject.SetUnstructuredContent(unstructuredContent)

						})

						It("returns an error", func() {
							_, err := reader.GetOutput(stampedObject)
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("failed to evaluate succeededCondition.Key [completion.path]: jsonpath returned empty list: completion.path"))

						})
					})
				})
			})

			Context("stampedObject has not reconciled (generation != observedGeneration)", func() {

				BeforeEach(func() {
					unstructuredContent = map[string]interface{}{
						"metadata": map[string]interface{}{
							"generation": 1,
						},
						"status": map[string]interface{}{
							"observedGeneration": 2,
						},
					}

					stampedObject = &unstructured.Unstructured{}
					stampedObject.SetUnstructuredContent(unstructuredContent)

					var err error
					reader, err = stamp.NewReader(template, deploymentInputImpl{})
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					_, err := reader.GetOutput(stampedObject)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("status.observedGeneration does not equal metadata.generation"))
				})
			})

			Context("stampedObject does not have a generation", func() {
				BeforeEach(func() {
					unstructuredContent = map[string]interface{}{
						"metadata": map[string]interface{}{},
						"status": map[string]interface{}{
							"observedGeneration": 2,
						},
					}

					stampedObject = &unstructured.Unstructured{}
					stampedObject.SetUnstructuredContent(unstructuredContent)

					var err error
					reader, err = stamp.NewReader(template, deploymentInputImpl{})
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					_, err := reader.GetOutput(stampedObject)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to evaluate json path 'metadata.generation'"))
				})
			})

			Context("stampedObject does not have an observedGeneration", func() {
				BeforeEach(func() {
					unstructuredContent = map[string]interface{}{
						"metadata": map[string]interface{}{
							"generation": 1,
						},
						"status": map[string]interface{}{},
					}

					stampedObject = &unstructured.Unstructured{}
					stampedObject.SetUnstructuredContent(unstructuredContent)

					var err error
					reader, err = stamp.NewReader(template, deploymentInputImpl{})
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					_, err := reader.GetOutput(stampedObject)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to evaluate status.observedGeneration"))
				})

			})

		})

		Context("observedMatches set", func() {
			BeforeEach(func() {
				template = &v1alpha1.ClusterDeploymentTemplate{
					Spec: v1alpha1.DeploymentSpec{
						ObservedMatches: []v1alpha1.ObservedMatch{
							{
								Input:  "input.path",
								Output: "output.path",
							},
						},
					},
				}
			})

			Context("when inputs and outputs do not match", func() {
				BeforeEach(func() {
					unstructuredContent := map[string]interface{}{
						"input": map[string]interface{}{
							"path": "happy",
						},
						"output": map[string]interface{}{
							"path": "happy",
						},
					}

					stampedObject = &unstructured.Unstructured{}
					stampedObject.SetUnstructuredContent(unstructuredContent)
				})

				Context("where the deployment is found", func() {
					BeforeEach(func() {
						var err error
						reader, err = stamp.NewReader(template, deploymentInputImpl{})
						Expect(err).NotTo(HaveOccurred())
					})

					It("returns an output", func() {
						output, err := reader.GetOutput(stampedObject)
						Expect(err).NotTo(HaveOccurred())
						Expect(output.Source.URL).To(Equal("my-url"))
						Expect(output.Source.Revision).To(Equal("my-revision"))
					})
				})

				Context("where the deployment is not found", func() {
					BeforeEach(func() {
						var err error
						reader, err = stamp.NewReader(template, noInputImpl{})
						Expect(err).NotTo(HaveOccurred())
					})

					It("returns an error", func() {
						_, err := reader.GetOutput(stampedObject)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("deployment not found in upstream template"))
					})
				})
			})

			Context("input cannot be found", func() {
				BeforeEach(func() {
					unstructuredContent := map[string]interface{}{
						"output": map[string]interface{}{
							"path": "happy",
						},
					}

					stampedObject = &unstructured.Unstructured{}
					stampedObject.SetUnstructuredContent(unstructuredContent)

					var err error
					reader, err = stamp.NewReader(template, deploymentInputImpl{})
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					_, err := reader.GetOutput(stampedObject)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("could not find value on input [input.path]:"))
				})
			})

			Context("input but not output can be found", func() {
				BeforeEach(func() {
					unstructuredContent := map[string]interface{}{
						"input": map[string]interface{}{
							"path": "happy",
						},
					}

					stampedObject = &unstructured.Unstructured{}
					stampedObject.SetUnstructuredContent(unstructuredContent)

					var err error
					reader, err = stamp.NewReader(template, deploymentInputImpl{})
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					_, err := reader.GetOutput(stampedObject)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("could not find value on output [output.path]:"))
				})
			})

			Context("values at input and output do not match", func() {
				BeforeEach(func() {
					unstructuredContent := map[string]interface{}{
						"input": map[string]interface{}{
							"path": "happy",
						},
						"output": map[string]interface{}{
							"path": "not happy",
						},
					}

					stampedObject = &unstructured.Unstructured{}
					stampedObject.SetUnstructuredContent(unstructuredContent)

					var err error
					reader, err = stamp.NewReader(template, deploymentInputImpl{})
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					_, err := reader.GetOutput(stampedObject)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("input [input.path] and output [output.path] do not match: happy != not happy"))
				})
			})
		})
	})

	Context("using a no output reader", func() {
		var (
			template      *v1alpha1.ClusterTemplate
			reader        stamp.Reader
			stampedObject *unstructured.Unstructured
		)

		BeforeEach(func() {
			template = &v1alpha1.ClusterTemplate{}

			var err error
			reader, err = stamp.NewReader(template, noInputImpl{})
			Expect(err).NotTo(HaveOccurred())

			stampedObject = &unstructured.Unstructured{}
		})

		It("returns an empty output", func() {
			output, _ := reader.GetOutput(stampedObject)
			Expect(output.Source).To(BeNil())
			Expect(output.Image).To(BeNil())
			Expect(output.Config).To(BeNil())
		})

		It("does not return an error", func() {
			_, err := reader.GetOutput(stampedObject)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
