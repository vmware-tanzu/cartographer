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
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vmware-tanzu/cartographer/pkg/stamp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/eval"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	"github.com/vmware-tanzu/cartographer/pkg/templates/templatesfakes"
)

var _ = Describe("ClusterDeploymentTemplate", func() {
	var (
		err                error
		deploymentTemplate *v1alpha1.ClusterDeploymentTemplate
		happyPathValue     string
	)

	BeforeEach(func() {
		happyPathValue = "All Good"

		deploymentTemplate = &v1alpha1.ClusterDeploymentTemplate{
			Spec: v1alpha1.DeploymentSpec{},
		}
	})

	ItReturnsAHelpfulError := func(expectedErrorSubstring string) {
		It("returns a helpful error", func() {
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(expectedErrorSubstring))
		})
	}

	ItDoesNotReturnAnError := func() {
		It("does not return an error", func() {
			Expect(err).ShouldNot(HaveOccurred())
		})
	}

	Describe("GetOutput", func() {
		var (
			output        *templates.Output
			stampedObject *unstructured.Unstructured
			evaluator     *templatesfakes.FakeEvaluator
			inputs        *templates.Inputs
		)

		BeforeEach(func() {
			stampedObject = &unstructured.Unstructured{}
			evaluator = &templatesfakes.FakeEvaluator{}
		})

		JustBeforeEach(func() {
			clusterDeploymentTemplateModel := templates.NewClusterDeploymentTemplateModel(deploymentTemplate, evaluator)
			clusterDeploymentTemplateModel.SetStampedObject(stampedObject)
			clusterDeploymentTemplateModel.SetInputs(inputs)
			output, err = clusterDeploymentTemplateModel.GetOutput()
		})

		Context("observedCompletion", func() {
			BeforeEach(func() {
				deploymentTemplate.Spec.ObservedCompletion = &v1alpha1.ObservedCompletion{
					SucceededCondition: v1alpha1.Condition{
						Key:   "completion.path",
						Value: happyPathValue,
					},
				}
			})
			When("stampedObject has reconciled (generation == observedGeneration)", func() {
				BeforeEach(func() {
					evaluator.EvaluateJsonPathReturnsOnCall(0, 42, nil)
					evaluator.EvaluateJsonPathReturnsOnCall(1, 42, nil)
				})

				When("success criterion is met", func() {
					BeforeEach(func() {
						evaluator.EvaluateJsonPathReturnsOnCall(2, happyPathValue, nil)
					})

					When("templating context includes a deployment", func() {
						BeforeEach(func() {
							inputs = &templates.Inputs{
								Deployment: &templates.SourceInput{
									URL:      "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
									Revision: "prod",
								},
							}
						})
						It("returns an appropriate output", func() {
							expectedOutput := templates.Source{
								URL:      "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
								Revision: "prod",
							}
							Expect(*output.Source).To(Equal(expectedOutput))
						})

						ItDoesNotReturnAnError()

						When("failure criterion is set", func() {
							BeforeEach(func() {
								deploymentTemplate.Spec.ObservedCompletion.FailedCondition = &v1alpha1.Condition{
									Key:   "failure.path",
									Value: "some sad path value",
								}
							})

							When("failure criterion is met", func() {
								BeforeEach(func() {
									evaluator.EvaluateJsonPathReturnsOnCall(2, "some sad path value", nil)
								})

								ItReturnsAHelpfulError("deployment failure condition [failure.path] was: some sad path value")
							})

							When("failure criterion path exists but is not met", func() {
								BeforeEach(func() {
									evaluator.EvaluateJsonPathReturnsOnCall(2, "", nil)
									evaluator.EvaluateJsonPathReturnsOnCall(3, happyPathValue, nil)
								})

								It("does not return an error", func() {
									Expect(err).To(BeNil())
								})
							})

							When("failure criterion path does not exist", func() {
								BeforeEach(func() {
									evaluator.EvaluateJsonPathReturnsOnCall(2, "", eval.JsonPathDoesNotExistError{Path: "failure.path"})
									evaluator.EvaluateJsonPathReturnsOnCall(3, happyPathValue, nil)
								})

								It("does not return an error", func() {
									Expect(err).To(BeNil())
								})
							})

							When("evaluating failure criterion path errors", func() {
								BeforeEach(func() {
									evaluator.EvaluateJsonPathReturnsOnCall(2, "", fmt.Errorf("some-error"))
									evaluator.EvaluateJsonPathReturnsOnCall(3, happyPathValue, nil)
								})

								It("returns an error", func() {
									Expect(err).To(HaveOccurred())
								})
							})
						})
					})

					When("templating context includes an incomplete deployment", func() {
						BeforeEach(func() {
							inputs = &templates.Inputs{
								Deployment: &templates.SourceInput{
									Revision: "prod",
								},
							}
						})
						It("returns the incomplete deployment as a source", func() {
							expectedOutput := templates.Source{
								Revision: "prod",
							}
							Expect(*output.Source).To(Equal(expectedOutput))
						})
					})

					When("templating context does not include a deployment", func() {
						BeforeEach(func() {
							inputs = &templates.Inputs{}
						})
						ItReturnsAHelpfulError("deployment not found in upstream template")
					})
				})

				When("success criterion is not met", func() {
					BeforeEach(func() {
						evaluator.EvaluateJsonPathReturnsOnCall(2, "some sad path value", nil)
					})

					ItReturnsAHelpfulError("deployment success condition [completion.path] was: some sad path value, expected: All Good")
				})

				When("success criterion path does not exist", func() {
					BeforeEach(func() {
						evaluator.EvaluateJsonPathReturnsOnCall(2, "", fmt.Errorf("some error"))
					})

					It("does not return an output", func() {
						Expect(output).To(BeNil())
					})

					It("returns an error which identifies the failing json path expression", func() {
						deploymentConditionError, ok := err.(stamp.DeploymentConditionError)
						Expect(ok).To(BeTrue())
						Expect(deploymentConditionError.Error()).To(ContainSubstring("completion.path"))
					})

					ItReturnsAHelpfulError("some error")
				})
			})

			When("stampedObject has not reconciled (generation != observedGeneration)", func() {
				BeforeEach(func() {
					evaluator.EvaluateJsonPathReturnsOnCall(0, 100, nil)
					evaluator.EvaluateJsonPathReturnsOnCall(1, 99, nil)
				})

				ItReturnsAHelpfulError("status.observedGeneration does not equal metadata.generation: %!s(int=99) != %!s(int=100)")
			})

			When("stampedObject does not have a generation)", func() {
				BeforeEach(func() {
					evaluator.EvaluateJsonPathReturnsOnCall(0, 0, fmt.Errorf("some error"))
				})

				ItReturnsAHelpfulError("evaluate json path 'metadata.generation': failed to evaluate metadata.generation: some error")
			})

			When("stampedObject does not have an observedGeneration)", func() {
				BeforeEach(func() {
					evaluator.EvaluateJsonPathReturnsOnCall(0, 100, nil)
					evaluator.EvaluateJsonPathReturnsOnCall(1, 0, fmt.Errorf("some error"))
				})

				ItReturnsAHelpfulError("failed to evaluate status.observedGeneration: some error")
			})
		})

		Context("observedMatches", func() {
			BeforeEach(func() {
				deploymentTemplate.Spec.ObservedMatches = []v1alpha1.ObservedMatch{
					{
						Input:  "input.path",
						Output: "output.path",
					},
				}
			})

			When("when inputs and outputs do match", func() {
				BeforeEach(func() {
					evaluator.EvaluateJsonPathReturnsOnCall(0, "a match!", nil)
					evaluator.EvaluateJsonPathReturnsOnCall(1, "a match!", nil)
				})

				When("templating context includes a deployment", func() {
					BeforeEach(func() {
						inputs = &templates.Inputs{
							Deployment: &templates.SourceInput{
								URL:      "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
								Revision: "prod",
							},
						}
					})
					It("returns an appropriate output", func() {
						expectedOutput := templates.Source{
							URL:      "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
							Revision: "prod",
						}
						Expect(*output.Source).To(Equal(expectedOutput))
					})

					ItDoesNotReturnAnError()
				})

				When("templating context includes an incomplete deployment", func() {
					BeforeEach(func() {
						inputs = &templates.Inputs{
							Deployment: &templates.SourceInput{
								Revision: "prod",
							},
						}
					})
					It("returns the incomplete deployment as a source", func() {
						expectedOutput := templates.Source{
							Revision: "prod",
						}
						Expect(*output.Source).To(Equal(expectedOutput))
					})
				})

				When("templating context does not include a deployment", func() {
					BeforeEach(func() {
						inputs = &templates.Inputs{}
					})
					ItReturnsAHelpfulError("deployment not found in upstream template")
				})
			})
			When("input cannot be found", func() {
				BeforeEach(func() {
					evaluator.EvaluateJsonPathReturnsOnCall(0, "", fmt.Errorf("some-error"))
				})
				ItReturnsAHelpfulError("could not find value on input [input.path]: some-error")
			})
			When("input but not output can be found", func() {
				BeforeEach(func() {
					evaluator.EvaluateJsonPathReturnsOnCall(0, "we could have had something beautiful", nil)
					evaluator.EvaluateJsonPathReturnsOnCall(1, "", fmt.Errorf("some-error"))
				})
				ItReturnsAHelpfulError("could not find value on output [output.path]: some-error")
			})
			When("values at input and output do not match", func() {
				BeforeEach(func() {
					evaluator.EvaluateJsonPathReturnsOnCall(0, "we could have had something beautiful", nil)
					evaluator.EvaluateJsonPathReturnsOnCall(1, "but it wasn't meant to be", nil)
				})
				ItReturnsAHelpfulError("input [input.path] and output [output.path] do not match: we could have had something beautiful != but it wasn't meant to be")
			})
		})
	})
})
