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

package pipeline_test

import (
	"errors"

	. "github.com/MakeNowJust/heredoc/dot"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/pipeline"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

var _ = Describe("Realizer", func() {
	var (
		out        *Buffer
		repository *repositoryfakes.FakeRepository
		logger     logr.Logger
		rlzr       realizer.Realizer
		pipeline   *v1alpha1.Pipeline
	)

	BeforeEach(func() {
		out = NewBuffer()
		logger = zap.New(zap.WriteTo(out))
		repository = &repositoryfakes.FakeRepository{}
		rlzr = realizer.NewRealizer()

		pipeline = &v1alpha1.Pipeline{
			Spec: v1alpha1.PipelineSpec{
				RunTemplateRef: v1alpha1.TemplateReference{
					Kind:      "RunTemplate",
					Name:      "my-template",
					Namespace: "some-ns",
				},
			},
		}
	})

	Context("with a valid RunTemplate", func() {
		BeforeEach(func() {
			templateAPI := &v1alpha1.RunTemplate{
				Spec: v1alpha1.RunTemplateSpec{
					Outputs: map[string]string{
						"myout": "data.has",
					},
					Template: runtime.RawExtension{
						Raw: []byte(D(`{
								"apiVersion": "v1",
								"kind": "ConfigMap",
								"metadata": { "generateName": "my-stamped-resource-" },
								"data": { "has": "is a string" }
							}`,
						)),
					},
				},
			}
			template := templates.NewRunTemplateModel(templateAPI)
			repository.GetRunTemplateReturns(template, nil)
		})

		It("stamps out the resource from the template", func() {
			_, _ = rlzr.Realize(pipeline, logger, repository)

			Expect(repository.GetRunTemplateCallCount()).To(Equal(1))
			Expect(repository.GetRunTemplateArgsForCall(0)).To(MatchFields(IgnoreExtras,
				Fields{
					"Kind": Equal("RunTemplate"),
					"Name": Equal("my-template"),
				},
			))

			Expect(repository.EnsureObjectExistsOnClusterCallCount()).To(Equal(1))
			Expect(repository.EnsureObjectExistsOnClusterCallCount()).To(Equal(1))
			stamped, allowUpdate := repository.EnsureObjectExistsOnClusterArgsForCall(0)
			Expect(allowUpdate).To(BeFalse())
			Expect(stamped.Object).To(
				MatchKeys(IgnoreExtras, Keys{
					"metadata": MatchKeys(IgnoreExtras, Keys{
						"generateName": Equal("my-stamped-resource-"),
					}),
					"apiVersion": Equal("v1"),
					"kind":       Equal("ConfigMap"),
					"data": MatchKeys(IgnoreExtras, Keys{
						"has": Equal("is a string"),
					}),
				}),
			)
		})

		It("returns a happy condition", func() {
			condition, _ := rlzr.Realize(pipeline, logger, repository)
			Expect(*condition).To(
				MatchFields(IgnoreExtras, Fields{
					"Type":   Equal("RunTemplateReady"),
					"Status": Equal(metav1.ConditionTrue),
					"Reason": Equal("Ready"),
				}),
			)
		})

		It("returns the outputs", func() {
			_, outputs := rlzr.Realize(pipeline, logger, repository)
			Expect(outputs["myout"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"is a string"`)}))
		})

		Context("error on Create", func() {
			BeforeEach(func() {
				repository.EnsureObjectExistsOnClusterReturns(errors.New("some bad error"))
			})

			It("logs the error", func() {
				_, _ = rlzr.Realize(pipeline, logger, repository)

				Expect(out).To(Say(`"msg":"could not create object"`))
				Expect(out).To(Say(`"error":"some bad error"`))
			})

			It("returns a condition stating that it failed to create", func() {
				condition, _ := rlzr.Realize(pipeline, logger, repository)

				Expect(*condition).To(
					MatchFields(IgnoreExtras, Fields{
						"Type":    Equal("RunTemplateReady"),
						"Status":  Equal(metav1.ConditionFalse),
						"Reason":  Equal("StampedObjectRejectedByAPIServer"),
						"Message": Equal("could not create object: some bad error"),
					}),
				)
			})

		})

	})

	Context("with unsatisfied output paths", func() {
		BeforeEach(func() {
			templateAPI := &v1alpha1.RunTemplate{
				Spec: v1alpha1.RunTemplateSpec{
					Outputs: map[string]string{
						"myout": "data.hasnot",
					},
					Template: runtime.RawExtension{
						Raw: []byte(D(`{
								"apiVersion": "v1",
								"kind": "ConfigMap",
								"metadata": { "generateName": "my-stamped-resource-" },
								"data": { "has": "is a string" }
							}`,
						)),
					},
				},
			}
			template := templates.NewRunTemplateModel(templateAPI)
			repository.GetRunTemplateReturns(template, nil)
		})

		It("logs info about the missing outputs", func() {
			_, _ = rlzr.Realize(pipeline, logger, repository)

			// FIXME need a `Log` matcher so we dont have multiline matches.
			Expect(out).To(Say(`"level":"info"`))
			Expect(out).To(Say(`"msg":"could not get output: get output: evaluate: find results: hasnot is not found"`))
		})


		It("returns a condition stating that it failed to get outputs", func() {
			condition, outputs := rlzr.Realize(pipeline, logger, repository)

			Expect(outputs).To(BeNil())

			Expect(*condition).To(
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal("RunTemplateReady"),
					"Status":  Equal(metav1.ConditionFalse),
					"Reason":  Equal("OutputPathNotSatisfied"),
					"Message": Equal("get output: evaluate: find results: hasnot is not found"),
				}),
			)
		})

	})

	Context("with an invalid RunTemplate", func() {
		BeforeEach(func() {
			templateAPI := &v1alpha1.RunTemplate{
				Spec: v1alpha1.RunTemplateSpec{
					Template: runtime.RawExtension{},
				},
			}
			template := templates.NewRunTemplateModel(templateAPI)
			repository.GetRunTemplateReturns(template, nil)
		})

		It("logs the error", func() {
			_, _ = rlzr.Realize(pipeline, logger, repository)

			Expect(out).To(Say(`"msg":"could not stamp template"`))
			Expect(out).To(Say(`"error":"unmarshal to JSON: unexpected end of JSON input"`))
		})

		It("returns a condition stating that it failed to stamp", func() {
			condition, _ := rlzr.Realize(pipeline, logger, repository)

			Expect(*condition).To(
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal("RunTemplateReady"),
					"Status":  Equal(metav1.ConditionFalse),
					"Reason":  Equal("TemplateStampFailure"),
					"Message": Equal("could not stamp template: unmarshal to JSON: unexpected end of JSON input"),
				}),
			)
		})

	})

	Context("the RunTemplate cannot be fetched", func() {
		BeforeEach(func() {
			repository.GetRunTemplateReturns(nil, errors.New("Errol mcErrorFace"))

			pipeline = &v1alpha1.Pipeline{
				Spec: v1alpha1.PipelineSpec{
					RunTemplateRef: v1alpha1.TemplateReference{
						Kind:      "RunTemplate",
						Name:      "my-template",
						Namespace: "some-ns",
					},
				},
			}
		})

		It("logs the error", func() {
			_, _ = rlzr.Realize(pipeline, logger, repository)

			Expect(out).To(Say(`"msg":"could not get RunTemplate 'my-template'"`))
			Expect(out).To(Say(`"error":"Errol mcErrorFace"`))
		})

		It("return the condition for a missing RunTemplate", func() {
			condition, _ := rlzr.Realize(pipeline, logger, repository)

			Expect(*condition).To(
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal("RunTemplateReady"),
					"Status":  Equal(metav1.ConditionFalse),
					"Reason":  Equal("RunTemplateNotFound"),
					"Message": Equal("could not get RunTemplate 'my-template': Errol mcErrorFace"),
				}),
			)
		})
	})
})
