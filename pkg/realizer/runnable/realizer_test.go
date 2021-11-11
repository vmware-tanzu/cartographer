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

package runnable_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	. "github.com/MakeNowJust/heredoc/dot"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/runnable"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
	"github.com/vmware-tanzu/cartographer/tests/resources"
)

var _ = Describe("Realizer", func() {
	var (
		out                 *Buffer
		repository          *repositoryfakes.FakeRepository
		logger              logr.Logger
		rlzr                realizer.Realizer
		runnable            *v1alpha1.Runnable
		createdUnstructured *unstructured.Unstructured
	)

	BeforeEach(func() {
		out = NewBuffer()
		logger = zap.New(zap.WriteTo(out))
		repository = &repositoryfakes.FakeRepository{}
		rlzr = realizer.NewRealizer()

		runnable = &v1alpha1.Runnable{
			Spec: v1alpha1.RunnableSpec{
				RunTemplateRef: v1alpha1.TemplateReference{
					Kind: "ClusterRunTemplate",
					Name: "my-template",
				},
			},
		}
	})

	Context("with a valid ClusterRunTemplate", func() {
		BeforeEach(func() {
			testObj := resources.Test{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Test",
					APIVersion: "test.run/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "my-stamped-resource-",
				},
				Spec: resources.TestSpec{
					Foo:   "is a string",
					Value: runtime.RawExtension{Raw: []byte(`"$(selected)$"`)},
				},
				Status: resources.TestStatus{
					ObservedGeneration: 1,
					Conditions: []metav1.Condition{{
						Type:               "Succeeded",
						Status:             "True",
						LastTransitionTime: metav1.Now(),
						Reason:             "",
					}},
				},
			}
			dbytes, err := json.Marshal(testObj)
			Expect(err).ToNot(HaveOccurred())

			var templateAPI = &v1alpha1.ClusterRunTemplate{
				Spec: v1alpha1.ClusterRunTemplateSpec{
					Outputs: map[string]string{
						"myout": "spec.foo",
					},
					Template: runtime.RawExtension{
						Raw: dbytes,
					},
				},
			}

			repository.GetRunTemplateReturns(templateAPI, nil)

			createdUnstructured = &unstructured.Unstructured{}

			repository.EnsureObjectExistsOnClusterStub = func(obj *unstructured.Unstructured, allowUpdate bool) error {
				createdUnstructured.Object = obj.Object
				return nil
			}

			repository.ListUnstructuredReturns([]*unstructured.Unstructured{createdUnstructured}, nil)
		})

		It("stamps out the resource from the template", func() {
			_, _, _ = rlzr.Realize(context.TODO(), runnable, logger, repository)

			Expect(repository.GetRunTemplateCallCount()).To(Equal(1))
			Expect(repository.GetRunTemplateArgsForCall(0)).To(MatchFields(IgnoreExtras,
				Fields{
					"Kind": Equal("ClusterRunTemplate"),
					"Name": Equal("my-template"),
				},
			))

			Expect(repository.EnsureObjectExistsOnClusterCallCount()).To(Equal(1))
			stamped, allowUpdate := repository.EnsureObjectExistsOnClusterArgsForCall(0)
			Expect(allowUpdate).To(BeFalse())
			Expect(stamped.Object).To(
				MatchKeys(IgnoreExtras, Keys{
					"metadata": MatchKeys(IgnoreExtras, Keys{
						"generateName": Equal("my-stamped-resource-"),
					}),
					"apiVersion": Equal("test.run/v1alpha1"),
					"kind":       Equal("Test"),
					"spec": MatchKeys(IgnoreExtras, Keys{
						"foo": Equal("is a string"),
					}),
				}),
			)
		})

		It("returns a happy condition", func() {
			condition, _, _ := rlzr.Realize(context.TODO(), runnable, logger, repository)
			Expect(*condition).To(
				MatchFields(IgnoreExtras, Fields{
					"Type":   Equal("RunTemplateReady"),
					"Status": Equal(metav1.ConditionTrue),
					"Reason": Equal("Ready"),
				}),
			)
		})

		It("returns the outputs", func() {
			_, outputs, _ := rlzr.Realize(context.TODO(), runnable, logger, repository)
			Expect(outputs["myout"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"is a string"`)}))
		})

		It("returns the stampedObject", func() {
			_, _, stampedObject := rlzr.Realize(context.TODO(), runnable, logger, repository)
			Expect(stampedObject.Object["spec"]).To(Equal(map[string]interface{}{
				"foo":   "is a string",
				"value": nil,
			}))
			Expect(stampedObject.Object["apiVersion"]).To(Equal("test.run/v1alpha1"))
			Expect(stampedObject.Object["kind"]).To(Equal("Test"))
		})

		Context("error on Create", func() {
			BeforeEach(func() {
				repository.EnsureObjectExistsOnClusterReturns(errors.New("some bad error"))
			})

			It("logs the error", func() {
				_, _, _ = rlzr.Realize(context.TODO(), runnable, logger, repository)

				Expect(out).To(Say(`"msg":"could not create object"`))
				Expect(out).To(Say(`"error":"some bad error"`))
			})

			It("returns a condition stating that it failed to create", func() {
				condition, _, _ := rlzr.Realize(context.TODO(), runnable, logger, repository)

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

		Context("listing previously created objects fails", func() {
			BeforeEach(func() {
				repository.ListUnstructuredReturns(nil, errors.New("some list error"))
			})

			It("logs the error", func() {
				_, _, _ = rlzr.Realize(context.TODO(), runnable, logger, repository)

				Expect(out).To(Say(`"msg":"could not list runnable objects: some list error"`))
			})

			It("returns a condition stating that it failed to list created objects", func() {
				condition, _, _ := rlzr.Realize(context.TODO(), runnable, logger, repository)

				Expect(*condition).To(
					MatchFields(IgnoreExtras, Fields{
						"Type":    Equal("RunTemplateReady"),
						"Status":  Equal(metav1.ConditionFalse),
						"Reason":  Equal("FailedToListCreatedObjects"),
						"Message": Equal("could not list runnable objects: some list error"),
					}),
				)
			})
		})

		Context("runnable selector resolves successfully", func() {
			BeforeEach(func() {
				runnable.Spec.Selector = &v1alpha1.ResourceSelector{
					Resource: v1alpha1.ResourceType{
						APIVersion: "apiversion-to-be-selected",
						Kind:       "kind-to-be-selected",
					},
					MatchingLabels: map[string]string{"expected-label": "expected-value"},
				}
				repository.ListUnstructuredReturns([]*unstructured.Unstructured{{map[string]interface{}{"useful-value": "from-selected-object"}}}, nil)
			})

			It("makes the selected object available in the templating context", func() {
				_, _, _ = rlzr.Realize(context.TODO(), runnable, logger, repository)

				Expect(repository.ListUnstructuredCallCount()).To(Equal(2))
				clientQueryObjectForSelector := repository.ListUnstructuredArgsForCall(0)
				Expect(clientQueryObjectForSelector.GetAPIVersion()).To(Equal("apiversion-to-be-selected"))
				Expect(clientQueryObjectForSelector.GetKind()).To(Equal("kind-to-be-selected"))
				Expect(clientQueryObjectForSelector.GetLabels()).To(Equal(map[string]string{"expected-label": "expected-value"}))

				Expect(repository.EnsureObjectExistsOnClusterCallCount()).To(Equal(1))
				stamped, allowUpdate := repository.EnsureObjectExistsOnClusterArgsForCall(0)
				Expect(allowUpdate).To(BeFalse())
				Expect(stamped.Object).To(
					MatchKeys(IgnoreExtras, Keys{
						"metadata": MatchKeys(IgnoreExtras, Keys{
							"generateName": Equal("my-stamped-resource-"),
						}),
						"apiVersion": Equal("test.run/v1alpha1"),
						"kind":       Equal("Test"),
						"spec": MatchKeys(IgnoreExtras, Keys{
							"value": MatchKeys(IgnoreExtras, Keys{
								"useful-value": Equal("from-selected-object"),
							}),
						}),
					}),
				)
			})
		})

		Context("runnable selector matches too many objects", func() {
			BeforeEach(func() {
				runnable.Spec.Selector = &v1alpha1.ResourceSelector{
					Resource: v1alpha1.ResourceType{
						APIVersion: "apiversion-to-be-selected",
						Kind:       "kind-to-be-selected",
					},
					MatchingLabels: map[string]string{"expected-label": "expected-value"},
				}
				repository.ListUnstructuredReturns([]*unstructured.Unstructured{{map[string]interface{}{}}, {map[string]interface{}{}}}, nil)
			})

			It("logs the error", func() {
				_, _, _ = rlzr.Realize(context.TODO(), runnable, logger, repository)

				Expect(out).To(Say(`"msg":"could not resolve selector \(apiVersion:apiversion-to-be-selected kind:kind-to-be-selected labels:map\[expected-label:expected-value\]\)"`))
				Expect(out).To(Say(`"error":"selector matched multiple objects"`))
			})

			It("returns a condition stating that it failed to create", func() {
				condition, _, _ := rlzr.Realize(context.TODO(), runnable, logger, repository)

				Expect(*condition).To(
					MatchFields(IgnoreExtras, Fields{
						"Type":    Equal("RunTemplateReady"),
						"Status":  Equal(metav1.ConditionFalse),
						"Reason":  Equal("TemplateStampFailure"),
						"Message": Equal("could not resolve selector (apiVersion:apiversion-to-be-selected kind:kind-to-be-selected labels:map[expected-label:expected-value]): selector matched multiple objects"),
					}),
				)
			})
		})

		Context("runnable selector does not match any objects", func() {
			BeforeEach(func() {
				runnable.Spec.Selector = &v1alpha1.ResourceSelector{
					Resource: v1alpha1.ResourceType{
						APIVersion: "apiversion-to-be-selected",
						Kind:       "kind-to-be-selected",
					},
					MatchingLabels: map[string]string{"expected-label": "expected-value"},
				}
				repository.ListUnstructuredReturns([]*unstructured.Unstructured{}, nil)
			})

			It("logs the error", func() {
				_, _, _ = rlzr.Realize(context.TODO(), runnable, logger, repository)

				Expect(out).To(Say(`"msg":"could not resolve selector \(apiVersion:apiversion-to-be-selected kind:kind-to-be-selected labels:map\[expected-label:expected-value\]\)"`))
				Expect(out).To(Say(`"error":"selector did not match any objects"`))
			})

			It("returns a condition stating that it failed to create", func() {
				condition, _, _ := rlzr.Realize(context.TODO(), runnable, logger, repository)

				Expect(*condition).To(
					MatchFields(IgnoreExtras, Fields{
						"Type":    Equal("RunTemplateReady"),
						"Status":  Equal(metav1.ConditionFalse),
						"Reason":  Equal("TemplateStampFailure"),
						"Message": Equal("could not resolve selector (apiVersion:apiversion-to-be-selected kind:kind-to-be-selected labels:map[expected-label:expected-value]): selector did not match any objects"),
					}),
				)
			})
		})

		Context("runnable selector cannot be resolved", func() {
			BeforeEach(func() {
				runnable.Spec.Selector = &v1alpha1.ResourceSelector{
					Resource: v1alpha1.ResourceType{
						APIVersion: "apiversion-to-be-selected",
						Kind:       "kind-to-be-selected",
					},
					MatchingLabels: map[string]string{"expected-label": "expected-value"},
				}
				repository.ListUnstructuredReturns(nil, fmt.Errorf("listing unstructured is hard"))
			})

			It("logs the error", func() {
				_, _, _ = rlzr.Realize(context.TODO(), runnable, logger, repository)

				Expect(out).To(Say(`"msg":"could not resolve selector \(apiVersion:apiversion-to-be-selected kind:kind-to-be-selected labels:map\[expected-label:expected-value\]\)"`))
				Expect(out).To(Say(`"error":"could not list objects matching selector: listing unstructured is hard"`))
			})

			It("returns a condition stating that it failed to create", func() {
				condition, _, _ := rlzr.Realize(context.TODO(), runnable, logger, repository)

				Expect(*condition).To(
					MatchFields(IgnoreExtras, Fields{
						"Type":    Equal("RunTemplateReady"),
						"Status":  Equal(metav1.ConditionFalse),
						"Reason":  Equal("TemplateStampFailure"),
						"Message": Equal("could not resolve selector (apiVersion:apiversion-to-be-selected kind:kind-to-be-selected labels:map[expected-label:expected-value]): could not list objects matching selector: listing unstructured is hard"),
					}),
				)
			})
		})
	})

	Context("with unsatisfied output paths", func() {
		BeforeEach(func() {
			templateAPI := &v1alpha1.ClusterRunTemplate{
				Spec: v1alpha1.ClusterRunTemplateSpec{
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

			repository.GetRunTemplateReturns(templateAPI, nil)

			createdUnstructured = &unstructured.Unstructured{}

			repository.EnsureObjectExistsOnClusterStub = func(obj *unstructured.Unstructured, allowUpdate bool) error {
				createdUnstructured.Object = obj.Object
				return nil
			}

			repository.ListUnstructuredReturns([]*unstructured.Unstructured{createdUnstructured}, nil)
		})

		It("logs info about the missing outputs", func() {
			_, _, _ = rlzr.Realize(context.TODO(), runnable, logger, repository)

			// FIXME need a `Log` matcher so we dont have multiline matches.
			Expect(out).To(Say(`"level":"info"`))
			Expect(out).To(Say(`"msg":"could not get output: get output: evaluate: find results: hasnot is not found"`))
		})

		It("returns a condition stating that it failed to get outputs", func() {
			condition, outputs, _ := rlzr.Realize(context.TODO(), runnable, logger, repository)

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

	Context("with an invalid ClusterRunTemplate", func() {
		BeforeEach(func() {
			templateAPI := &v1alpha1.ClusterRunTemplate{
				Spec: v1alpha1.ClusterRunTemplateSpec{
					Template: runtime.RawExtension{},
				},
			}
			repository.GetRunTemplateReturns(templateAPI, nil)
		})

		It("logs the error", func() {
			_, _, _ = rlzr.Realize(context.TODO(), runnable, logger, repository)

			Expect(out).To(Say(`"msg":"could not stamp template"`))
			Expect(out).To(Say(`"error":"unmarshal to JSON: unexpected end of JSON input"`))
		})

		It("returns a condition stating that it failed to stamp", func() {
			condition, _, _ := rlzr.Realize(context.TODO(), runnable, logger, repository)

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

	Context("the ClusterRunTemplate cannot be fetched", func() {
		BeforeEach(func() {
			repository.GetRunTemplateReturns(nil, errors.New("Errol mcErrorFace"))

			runnable = &v1alpha1.Runnable{
				Spec: v1alpha1.RunnableSpec{
					RunTemplateRef: v1alpha1.TemplateReference{
						Kind: "ClusterRunTemplate",
						Name: "my-template",
					},
				},
			}
		})

		It("logs the error", func() {
			_, _, _ = rlzr.Realize(context.TODO(), runnable, logger, repository)

			Expect(out).To(Say(`"msg":"could not get ClusterRunTemplate 'my-template'"`))
			Expect(out).To(Say(`"error":"Errol mcErrorFace"`))
		})

		It("return the condition for a missing ClusterRunTemplate", func() {
			condition, _, _ := rlzr.Realize(context.TODO(), runnable, logger, repository)

			Expect(*condition).To(
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal("RunTemplateReady"),
					"Status":  Equal(metav1.ConditionFalse),
					"Reason":  Equal("RunTemplateNotFound"),
					"Message": Equal("could not get ClusterRunTemplate 'my-template': Errol mcErrorFace"),
				}),
			)
		})
	})
})
