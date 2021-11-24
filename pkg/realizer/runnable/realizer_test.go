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
	"reflect"

	. "github.com/MakeNowJust/heredoc/dot"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/runnable"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
	"github.com/vmware-tanzu/cartographer/tests/resources"
)

var _ = Describe("Realizer", func() {
	var (
		ctx                 context.Context
		systemRepo          *repositoryfakes.FakeRepository
		runnableRepo        *repositoryfakes.FakeRepository
		rlzr                realizer.Realizer
		runnable            *v1alpha1.Runnable
		createdUnstructured *unstructured.Unstructured
	)

	BeforeEach(func() {
		ctx = context.Background()
		systemRepo = &repositoryfakes.FakeRepository{}
		runnableRepo = &repositoryfakes.FakeRepository{}
		rlzr = realizer.NewRealizer()

		runnable = &v1alpha1.Runnable{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-runnable",
				Namespace: "my-important-ns",
			},
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
			testObj := resources.TestObj{
				TypeMeta: metav1.TypeMeta{
					Kind:       "TestObj",
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

			systemRepo.GetRunTemplateReturns(templateAPI, nil)

			createdUnstructured = &unstructured.Unstructured{}

			runnableRepo.EnsureObjectExistsOnClusterStub = func(ctx context.Context, obj *unstructured.Unstructured, allowUpdate bool) error {
				createdUnstructured.Object = obj.Object
				return nil
			}

			runnableRepo.ListUnstructuredReturns([]*unstructured.Unstructured{createdUnstructured}, nil)
		})

		It("stamps out the resource from the template", func() {
			_, _, _ = rlzr.Realize(ctx, runnable, systemRepo, runnableRepo)

			Expect(systemRepo.GetRunTemplateCallCount()).To(Equal(1))
			actualCtx, actualTemplate := systemRepo.GetRunTemplateArgsForCall(0)
			Expect(actualCtx).To(Equal(ctx))
			Expect(actualTemplate).To(MatchFields(IgnoreExtras,
				Fields{
					"Kind": Equal("ClusterRunTemplate"),
					"Name": Equal("my-template"),
				},
			))

			Expect(runnableRepo.EnsureObjectExistsOnClusterCallCount()).To(Equal(1))
			actualCtx, stamped, allowUpdate := runnableRepo.EnsureObjectExistsOnClusterArgsForCall(0)
			Expect(actualCtx).To(Equal(ctx))
			Expect(allowUpdate).To(BeFalse())
			Expect(stamped.Object).To(
				MatchKeys(IgnoreExtras, Keys{
					"metadata": MatchKeys(IgnoreExtras, Keys{
						"generateName": Equal("my-stamped-resource-"),
					}),
					"apiVersion": Equal("test.run/v1alpha1"),
					"kind":       Equal("TestObj"),
					"spec": MatchKeys(IgnoreExtras, Keys{
						"foo": Equal("is a string"),
					}),
				}),
			)
		})

		It("does not return an error", func() {
			_, _, err := rlzr.Realize(ctx, runnable, systemRepo, runnableRepo)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns the outputs", func() {
			_, outputs, _ := rlzr.Realize(ctx, runnable, systemRepo, runnableRepo)
			Expect(outputs["myout"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"is a string"`)}))
		})

		It("returns the stampedObject", func() {
			stampedObject, _, _ := rlzr.Realize(ctx, runnable, systemRepo, runnableRepo)
			Expect(stampedObject.Object["spec"]).To(Equal(map[string]interface{}{
				"foo":   "is a string",
				"value": nil,
			}))
			Expect(stampedObject.Object["apiVersion"]).To(Equal("test.run/v1alpha1"))
			Expect(stampedObject.Object["kind"]).To(Equal("TestObj"))
		})

		Context("error on EnsureObjectExistsOnCluster", func() {
			BeforeEach(func() {
				runnableRepo.EnsureObjectExistsOnClusterReturns(errors.New("some bad error"))
			})

			It("returns ApplyStampedObjectError", func() {
				_, _, err := rlzr.Realize(ctx, runnable, systemRepo, runnableRepo)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("some bad error"))
				Expect(reflect.TypeOf(err).String()).To(Equal("runnable.ApplyStampedObjectError"))
			})
		})

		Context("listing previously created objects fails", func() {
			BeforeEach(func() {
				runnableRepo.ListUnstructuredReturns(nil, errors.New("some list error"))
			})

			It("returns ListCreatedObjectsError", func() {
				_, _, err := rlzr.Realize(ctx, runnable, systemRepo, runnableRepo)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("some list error"))
				Expect(reflect.TypeOf(err).String()).To(Equal("runnable.ListCreatedObjectsError"))
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
				runnableRepo.ListUnstructuredReturns([]*unstructured.Unstructured{{map[string]interface{}{"useful-value": "from-selected-object"}}}, nil)
			})

			It("makes the selected object available in the templating context", func() {
				_, _, _ = rlzr.Realize(ctx, runnable, systemRepo, runnableRepo)

				Expect(runnableRepo.ListUnstructuredCallCount()).To(Equal(2))
				actualCtx, clientQueryObjectForSelector := runnableRepo.ListUnstructuredArgsForCall(0)
				Expect(actualCtx).To(Equal(ctx))

				Expect(clientQueryObjectForSelector.GetAPIVersion()).To(Equal("apiversion-to-be-selected"))
				Expect(clientQueryObjectForSelector.GetKind()).To(Equal("kind-to-be-selected"))
				Expect(clientQueryObjectForSelector.GetLabels()).To(Equal(map[string]string{"expected-label": "expected-value"}))

				Expect(runnableRepo.EnsureObjectExistsOnClusterCallCount()).To(Equal(1))
				actualCtx, stamped, allowUpdate := runnableRepo.EnsureObjectExistsOnClusterArgsForCall(0)
				Expect(actualCtx).To(Equal(ctx))
				Expect(allowUpdate).To(BeFalse())
				Expect(stamped.Object).To(
					MatchKeys(IgnoreExtras, Keys{
						"metadata": MatchKeys(IgnoreExtras, Keys{
							"generateName": Equal("my-stamped-resource-"),
						}),
						"apiVersion": Equal("test.run/v1alpha1"),
						"kind":       Equal("TestObj"),
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
				runnableRepo.ListUnstructuredReturns([]*unstructured.Unstructured{{map[string]interface{}{}}, {map[string]interface{}{}}}, nil)
			})

			It("returns ResolveSelectorError", func() {
				_, _, err := rlzr.Realize(ctx, runnable, systemRepo, runnableRepo)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(`unable to resolve selector '(apiVersion:apiversion-to-be-selected kind:kind-to-be-selected labels:map[expected-label:expected-value])': 'selector matched multiple objects'`))
				Expect(reflect.TypeOf(err).String()).To(Equal("runnable.ResolveSelectorError"))
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
				runnableRepo.ListUnstructuredReturns([]*unstructured.Unstructured{}, nil)
			})

			It("returns ResolveSelectorError", func() {
				_, _, err := rlzr.Realize(ctx, runnable, systemRepo, runnableRepo)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(`unable to resolve selector '(apiVersion:apiversion-to-be-selected kind:kind-to-be-selected labels:map[expected-label:expected-value])': 'selector did not match any objects'`))
				Expect(reflect.TypeOf(err).String()).To(Equal("runnable.ResolveSelectorError"))
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
				runnableRepo.ListUnstructuredReturns(nil, fmt.Errorf("listing unstructured is hard"))
			})

			It("returns ResolveSelectorError", func() {
				_, _, err := rlzr.Realize(ctx, runnable, systemRepo, runnableRepo)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(`unable to resolve selector '(apiVersion:apiversion-to-be-selected kind:kind-to-be-selected labels:map[expected-label:expected-value])': 'could not list objects matching selector: listing unstructured is hard'`))
				Expect(reflect.TypeOf(err).String()).To(Equal("runnable.ResolveSelectorError"))
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

			systemRepo.GetRunTemplateReturns(templateAPI, nil)

			createdUnstructured = &unstructured.Unstructured{}

			runnableRepo.EnsureObjectExistsOnClusterStub = func(ctx context.Context, obj *unstructured.Unstructured, allowUpdate bool) error {
				createdUnstructured.Object = obj.Object
				return nil
			}

			runnableRepo.ListUnstructuredReturns([]*unstructured.Unstructured{createdUnstructured}, nil)
		})

		It("returns RetrieveOutputError", func() {
			_, _, err := rlzr.Realize(ctx, runnable, systemRepo, runnableRepo)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`unable to retrieve outputs from stamped object for runnable 'my-important-ns/my-runnable': failed to evaluate path [data.hasnot]: evaluate: find results: hasnot is not found`))
			Expect(reflect.TypeOf(err).String()).To(Equal("runnable.RetrieveOutputError"))
		})
	})

	Context("with an invalid ClusterRunTemplate", func() {
		BeforeEach(func() {
			templateAPI := &v1alpha1.ClusterRunTemplate{
				Spec: v1alpha1.ClusterRunTemplateSpec{
					Template: runtime.RawExtension{},
				},
			}
			systemRepo.GetRunTemplateReturns(templateAPI, nil)
		})

		It("returns StampError", func() {
			_, _, err := rlzr.Realize(ctx, runnable, systemRepo, runnableRepo)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`unable to stamp object 'my-important-ns/my-runnable': 'failed to unmarshal json resource template: unexpected end of JSON input'`))
			Expect(reflect.TypeOf(err).String()).To(Equal("runnable.StampError"))
		})
	})

	Context("the ClusterRunTemplate cannot be fetched", func() {
		BeforeEach(func() {
			systemRepo.GetRunTemplateReturns(nil, errors.New("Errol mcErrorFace"))

			runnable.Spec = v1alpha1.RunnableSpec{
				RunTemplateRef: v1alpha1.TemplateReference{
					Kind: "ClusterRunTemplate",
					Name: "my-template",
				},
			}
		})

		It("returns GetRunTemplateError", func() {
			_, _, err := rlzr.Realize(ctx, runnable, systemRepo, runnableRepo)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`unable to get runnable 'my-important-ns/my-runnable': 'Errol mcErrorFace'`))
			Expect(reflect.TypeOf(err).String()).To(Equal("runnable.GetRunTemplateError"))
		})
	})
})
