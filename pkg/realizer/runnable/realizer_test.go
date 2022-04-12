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
	"time"

	. "github.com/MakeNowJust/heredoc/dot"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/runnable"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/runnable/runnablefakes"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
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
		discoveryClient     *runnablefakes.FakeDiscoveryInterface
	)

	BeforeEach(func() {
		ctx = context.Background()
		systemRepo = &repositoryfakes.FakeRepository{}
		runnableRepo = &repositoryfakes.FakeRepository{}
		discoveryClient = &runnablefakes.FakeDiscoveryInterface{}
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
				Spec: v1alpha1.RunTemplateSpec{
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

			runnableRepo.EnsureImmutableObjectExistsOnClusterStub = func(ctx context.Context, obj *unstructured.Unstructured, labels map[string]string) error {
				createdUnstructured.Object = obj.Object
				return nil
			}

			runnableRepo.ListUnstructuredReturns([]*unstructured.Unstructured{createdUnstructured}, nil)

			discoveryClient.ServerResourcesForGroupVersionReturns(&metav1.APIResourceList{
				APIResources: []metav1.APIResource{
					{
						Kind:       "kind-to-be-selected",
						Namespaced: true,
					},
				},
			}, nil)
		})

		It("stamps out the resource from the template", func() {
			_, _, _ = rlzr.Realize(ctx, runnable, systemRepo, runnableRepo, discoveryClient)

			Expect(systemRepo.GetRunTemplateCallCount()).To(Equal(1))
			_, actualTemplate := systemRepo.GetRunTemplateArgsForCall(0)
			Expect(actualTemplate).To(MatchFields(IgnoreExtras,
				Fields{
					"Kind": Equal("ClusterRunTemplate"),
					"Name": Equal("my-template"),
				},
			))

			Expect(runnableRepo.EnsureImmutableObjectExistsOnClusterCallCount()).To(Equal(1))
			_, stamped, labels := runnableRepo.EnsureImmutableObjectExistsOnClusterArgsForCall(0)
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
			Expect(labels).To(Equal(map[string]string{
				"carto.run/runnable-name": "my-runnable",
			}))
		})

		It("does not return an error", func() {
			_, _, err := rlzr.Realize(ctx, runnable, systemRepo, runnableRepo, discoveryClient)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns the outputs", func() {
			_, outputs, _ := rlzr.Realize(ctx, runnable, systemRepo, runnableRepo, discoveryClient)
			Expect(outputs["myout"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"is a string"`)}))
		})

		It("returns the stampedObject", func() {
			stampedObject, _, _ := rlzr.Realize(ctx, runnable, systemRepo, runnableRepo, discoveryClient)
			Expect(stampedObject.Object["spec"]).To(Equal(map[string]interface{}{
				"foo":   "is a string",
				"value": nil,
			}))
			Expect(stampedObject.Object["apiVersion"]).To(Equal("test.run/v1alpha1"))
			Expect(stampedObject.Object["kind"]).To(Equal("TestObj"))
		})

		It("garbage collects failed and successful runnable stamped objects according to retention policy", func() {
			runnable.Spec.RetentionPolicy.MaxFailedRuns = 1
			runnable.Spec.RetentionPolicy.MaxSuccessfulRuns = 1

			success1 := &unstructured.Unstructured{}
			success2 := &unstructured.Unstructured{}
			failed1 := &unstructured.Unstructured{}
			failed2 := &unstructured.Unstructured{}
			t0 := time.Now()

			stampedObjManifest := utils.HereYaml(`
				apiVersion: test.run/v1alpha1
				kind: TestObj
				metadata:
				  name: success2
				  namespace: default
				  creationTimestamp: ` + t0.Add(-1*time.Hour).Format(time.RFC3339) + `
				status:
				  conditions:
					- type: Succeeded
					  status: "True"
			`)
			dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
			_, _, err := dec.Decode([]byte(stampedObjManifest), nil, success2)
			Expect(err).NotTo(HaveOccurred())
			stampedObjManifest = utils.HereYaml(`
				apiVersion: test.run/v1alpha1
				kind: TestObj
				metadata:
				  name: failed1
				  namespace: default
				  creationTimestamp: ` + t0.Add(-1*time.Hour).Format(time.RFC3339) + `
				status:
				  conditions:
					- type: Succeeded
					  status: "False"
			`)
			_, _, err = dec.Decode([]byte(stampedObjManifest), nil, failed1)
			Expect(err).NotTo(HaveOccurred())
			stampedObjManifest = utils.HereYaml(`
				apiVersion: test.run/v1alpha1
				kind: TestObj
				metadata:
				  name: failed2
				  namespace: default
				  creationTimestamp: ` + t0.Add(-2*time.Hour).Format(time.RFC3339) + `
				status:
				  conditions:
					- type: Succeeded
					  status: "False"
				`)
			_, _, err = dec.Decode([]byte(stampedObjManifest), nil, failed2)
			Expect(err).NotTo(HaveOccurred())

			runnableRepo.ListUnstructuredReturns([]*unstructured.Unstructured{success2, failed1, failed2}, nil)

			runnableRepo.EnsureImmutableObjectExistsOnClusterStub = func(ctx context.Context, obj *unstructured.Unstructured, labels map[string]string) error {
				success1.Object = obj.Object
				success1.SetName("success1")
				success1.SetCreationTimestamp(metav1.Time{Time: t0})
				success1.Object["status"] = map[string]interface{}{
					"conditions": []map[string]interface{}{
						{
							"type":   "Succeeded",
							"status": "True",
						},
					},
				}
				runnableRepo.ListUnstructuredReturns([]*unstructured.Unstructured{success1, success2, failed1, failed2}, nil)
				return nil
			}

			_, _, err = rlzr.Realize(ctx, runnable, systemRepo, runnableRepo, discoveryClient)
			Expect(err).NotTo(HaveOccurred())

			Expect(runnableRepo.DeleteCallCount()).To(Equal(2))

			_, deleted1 := runnableRepo.DeleteArgsForCall(0)
			_, deleted2 := runnableRepo.DeleteArgsForCall(1)
			allDeletedObjects := []*unstructured.Unstructured{deleted2, deleted1}
			Expect(allDeletedObjects).To(ConsistOf(success2, failed2))
		})

		Context("error on EnsureImmutableObjectExistsOnCluster", func() {
			BeforeEach(func() {
				runnableRepo.EnsureImmutableObjectExistsOnClusterReturns(errors.New("some bad error"))
			})

			It("returns ApplyStampedObjectError", func() {
				_, _, err := rlzr.Realize(ctx, runnable, systemRepo, runnableRepo, discoveryClient)
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
				_, _, err := rlzr.Realize(ctx, runnable, systemRepo, runnableRepo, discoveryClient)
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

			Context("the selected object is namespaced", func() {
				BeforeEach(func() {
					discoveryClient.ServerResourcesForGroupVersionReturns(&metav1.APIResourceList{
						APIResources: []metav1.APIResource{
							{
								Kind:       "kind-to-be-selected",
								Namespaced: true,
							},
						},
					}, nil)
				})

				It("makes the selected object available in the templating context", func() {
					_, _, _ = rlzr.Realize(ctx, runnable, systemRepo, runnableRepo, discoveryClient)

					Expect(runnableRepo.ListUnstructuredCallCount()).To(Equal(2))
					_, gvk, namespace, labels := runnableRepo.ListUnstructuredArgsForCall(0)

					Expect(gvk.Version).To(Equal("apiversion-to-be-selected"))
					Expect(gvk.Kind).To(Equal("kind-to-be-selected"))
					Expect(labels).To(Equal(map[string]string{"expected-label": "expected-value"}))
					Expect(namespace).To(Equal("my-important-ns"))

					Expect(runnableRepo.EnsureImmutableObjectExistsOnClusterCallCount()).To(Equal(1))
					_, stamped, labels := runnableRepo.EnsureImmutableObjectExistsOnClusterArgsForCall(0)
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
					Expect(labels).To(Equal(map[string]string{
						"carto.run/runnable-name": "my-runnable",
					}))
				})
			})

			Context("the selected object is cluster scoped", func() {
				BeforeEach(func() {
					discoveryClient.ServerResourcesForGroupVersionReturns(&metav1.APIResourceList{
						APIResources: []metav1.APIResource{
							{
								Kind:       "kind-to-be-selected",
								Namespaced: false,
							},
						},
					}, nil)
				})

				It("makes the selected object available in the templating context", func() {
					_, _, _ = rlzr.Realize(ctx, runnable, systemRepo, runnableRepo, discoveryClient)

					Expect(runnableRepo.ListUnstructuredCallCount()).To(Equal(2))
					_, gvk, namespace, labels := runnableRepo.ListUnstructuredArgsForCall(0)

					Expect(gvk.Version).To(Equal("apiversion-to-be-selected"))
					Expect(gvk.Kind).To(Equal("kind-to-be-selected"))
					Expect(labels).To(Equal(map[string]string{"expected-label": "expected-value"}))
					Expect(namespace).To(Equal(""))

					Expect(runnableRepo.EnsureImmutableObjectExistsOnClusterCallCount()).To(Equal(1))
					_, stamped, labels := runnableRepo.EnsureImmutableObjectExistsOnClusterArgsForCall(0)
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
					Expect(labels).To(Equal(map[string]string{
						"carto.run/runnable-name": "my-runnable",
					}))
				})
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
				_, _, err := rlzr.Realize(ctx, runnable, systemRepo, runnableRepo, discoveryClient)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(`unable to resolve selector [map[expected-label:expected-value]], apiVersion [apiversion-to-be-selected], kind [kind-to-be-selected]: selector matched multiple objects`))
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
				_, _, err := rlzr.Realize(ctx, runnable, systemRepo, runnableRepo, discoveryClient)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(`unable to resolve selector [map[expected-label:expected-value]], apiVersion [apiversion-to-be-selected], kind [kind-to-be-selected]: selector did not match any objects`))
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
				_, _, err := rlzr.Realize(ctx, runnable, systemRepo, runnableRepo, discoveryClient)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(`unable to resolve selector [map[expected-label:expected-value]], apiVersion [apiversion-to-be-selected], kind [kind-to-be-selected]: failed to list objects in namespace matching selector [map[expected-label:expected-value]]: listing unstructured is hard`))
				Expect(reflect.TypeOf(err).String()).To(Equal("runnable.ResolveSelectorError"))
			})
		})
	})

	Context("with unsatisfied output paths", func() {
		BeforeEach(func() {
			templateAPI := &v1alpha1.ClusterRunTemplate{
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

			systemRepo.GetRunTemplateReturns(templateAPI, nil)

			createdUnstructured = &unstructured.Unstructured{}

			runnableRepo.EnsureImmutableObjectExistsOnClusterStub = func(ctx context.Context, obj *unstructured.Unstructured, labels map[string]string) error {
				createdUnstructured.Object = obj.Object
				return nil
			}

			runnableRepo.ListUnstructuredReturns([]*unstructured.Unstructured{createdUnstructured}, nil)
		})

		It("returns RetrieveOutputError", func() {
			_, _, err := rlzr.Realize(ctx, runnable, systemRepo, runnableRepo, discoveryClient)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`unable to retrieve outputs from stamped object [my-important-ns/my-stamped-resource-] of type [configmap] for run template [my-template]: failed to evaluate path [data.hasnot]: jsonpath returned empty list: data.hasnot`))
			Expect(reflect.TypeOf(err).String()).To(Equal("runnable.RetrieveOutputError"))
		})
	})

	Context("with an invalid ClusterRunTemplate", func() {
		BeforeEach(func() {
			templateAPI := &v1alpha1.ClusterRunTemplate{
				Spec: v1alpha1.RunTemplateSpec{
					Template: runtime.RawExtension{},
				},
			}
			systemRepo.GetRunTemplateReturns(templateAPI, nil)
		})

		It("returns StampError", func() {
			_, _, err := rlzr.Realize(ctx, runnable, systemRepo, runnableRepo, discoveryClient)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`unable to stamp object for run template [my-template]: failed to unmarshal json resource template: unexpected end of JSON input`))
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
			_, _, err := rlzr.Realize(ctx, runnable, systemRepo, runnableRepo, discoveryClient)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`unable to get run template [my-template]: Errol mcErrorFace`))
			Expect(reflect.TypeOf(err).String()).To(Equal("runnable.GetRunTemplateError"))
		})
	})
})
