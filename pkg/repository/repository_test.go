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

package repository_test

import (
	"context"
	"errors"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

var _ = Describe("repository", func() {
	var (
		repo  repository.Repository
		cache *repositoryfakes.FakeRepoCache
	)

	BeforeEach(func() {
		cache = &repositoryfakes.FakeRepoCache{}
	})

	Describe("tests using counterfeiter client", func() {
		var cl *repositoryfakes.FakeClient

		BeforeEach(func() {
			cl = &repositoryfakes.FakeClient{}
			repo = repository.NewRepository(cl, cache)
		})

		Context("EnsureObjectExistsOnCluster", func() {
			var stampedObj *unstructured.Unstructured

			BeforeEach(func() {
				stampedObj = &unstructured.Unstructured{}
				stampedObjManifest := `
apiVersion: batch/v1
kind: Job
metadata:
  name: hello
  namespace: default
spec:
  template:
    spec:
      containers:
      - name: hello
        image: busybox
        command: ['sh', '-c', 'echo "Hello, Kubernetes!" && sleep 3600']
      restartPolicy: OnFailure
`
				dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
				_, _, err := dec.Decode([]byte(stampedObjManifest), nil, stampedObj)
				Expect(err).NotTo(HaveOccurred())
			})

			It("attempts to get the object from the apiServer", func() {
				Expect(repo.EnsureObjectExistsOnCluster(stampedObj, true)).To(Succeed())

				Expect(cl.ListCallCount()).To(Equal(1))

				listOptions := []client.ListOption{
					client.InNamespace(stampedObj.GetNamespace()),
					client.MatchingLabels(stampedObj.GetLabels()),
				}

				_, objectList, options := cl.ListArgsForCall(0)
				unstructuredList, ok := objectList.(*unstructured.UnstructuredList)
				Expect(ok).To(BeTrue())
				Expect(len(unstructuredList.Items)).To(Equal(0))
				Expect(options).To(Equal(listOptions))
				Expect(unstructuredList.GetObjectKind().GroupVersionKind()).To(Equal(stampedObj.GroupVersionKind()))
			})

			Context("when the apiServer errors when trying to get the object", func() {
				BeforeEach(func() {
					cl.ListReturns(errors.New("some-error"))
				})

				It("returns a helpful error", func() {
					err := repo.EnsureObjectExistsOnCluster(stampedObj, true)
					Expect(err).To(MatchError(ContainSubstring("list: some-error")))
				})

				It("does not create or patch any objects", func() {
					_ = repo.EnsureObjectExistsOnCluster(stampedObj, true)
					Expect(cl.CreateCallCount()).To(Equal(0))
					Expect(cl.PatchCallCount()).To(Equal(0))
				})

				It("does not write to the submitted or persisted cache", func() {
					_ = repo.EnsureObjectExistsOnCluster(stampedObj, true)
					Expect(cache.SetCallCount()).To(Equal(0))
				})
			})

			Context("and the apiServer attempts to get the object and it doesn't exist", func() {
				BeforeEach(func() {
					// default behavior is empty list - no need to stub
				})
				It("attempts to create the object", func() {
					Expect(repo.EnsureObjectExistsOnCluster(stampedObj, true)).To(Succeed())

					Expect(cl.CreateCallCount()).To(Equal(1))
					_, createCallObj, _ := cl.CreateArgsForCall(0)
					Expect(createCallObj).To(Equal(stampedObj))
				})

				Context("and the apiServer errors when creating the object", func() {
					BeforeEach(func() {
						cl.CreateReturns(errors.New("some-error"))
					})

					It("returns a helpful error", func() {
						err := repo.EnsureObjectExistsOnCluster(stampedObj, true)
						Expect(err).To(MatchError(ContainSubstring("create: some-error")))
					})

					It("does not write to the submitted or persisted cache", func() {
						_ = repo.EnsureObjectExistsOnCluster(stampedObj, true)
						Expect(cache.SetCallCount()).To(Equal(0))
					})
				})

				Context("and the apiServer succeeds", func() {
					var returnedCreatedObj *unstructured.Unstructured
					BeforeEach(func() {
						returnedCreatedObj = stampedObj.DeepCopy()
						Expect(utils.AlterFieldOfNestedStringMaps(returnedCreatedObj.Object, "spec.template.spec.restartPolicy", "Never")).To(Succeed())
						cl.CreateStub = func(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
							objVal := reflect.ValueOf(obj)
							returnVal := reflect.ValueOf(returnedCreatedObj)

							reflect.Indirect(objVal).Set(reflect.Indirect(returnVal))
							return nil
						}
					})

					It("does not return an error", func() {
						Expect(repo.EnsureObjectExistsOnCluster(stampedObj, true)).To(Succeed())
					})

					It("caches the submitted and persisted objects, as the persisted one may be modified by mutating webhooks", func() {
						originalStampedObj := stampedObj.DeepCopy()

						Expect(repo.EnsureObjectExistsOnCluster(stampedObj, true)).To(Succeed())
						Expect(cache.SetCallCount()).To(Equal(1))
						submitted, persisted := cache.SetArgsForCall(0)
						Expect(*submitted).To(Equal(*originalStampedObj))
						Expect(*persisted).To(Equal(*returnedCreatedObj))
					})
				})
			})

			Context("and apiServer succeeds in getting the list of object(s)", func() {
				var (
					existingObj     *unstructured.Unstructured
					existingObjList unstructured.UnstructuredList
				)

				BeforeEach(func() {
					existingObj = &unstructured.Unstructured{}
					existingObj.SetName("hello")
					existingObj.SetNamespace("default")
					existingObj.SetGeneration(5)

					existingObjList = unstructured.UnstructuredList{
						Items: []unstructured.Unstructured{*existingObj},
					}
					cl.ListStub = func(ctx context.Context, list client.ObjectList, option ...client.ListOption) error {
						listVal := reflect.ValueOf(list)
						existingVal := reflect.ValueOf(existingObjList)

						reflect.Indirect(listVal).Set(reflect.Indirect(existingVal))
						return nil
					}
				})

				It("the cache is consulted to see if there was a change since the last time the cache was updated", func() {
					Expect(repo.EnsureObjectExistsOnCluster(stampedObj, true)).To(Succeed())
					Expect(cache.UnchangedSinceCachedCallCount()).To(Equal(1))

					submitted, persisted := cache.UnchangedSinceCachedArgsForCall(0)
					Expect(*submitted).To(Equal(*stampedObj))
					Expect(persisted[0]).To(Equal(existingObj))
				})

				Context("and the cache determines there has been no change since the last update", func() {
					BeforeEach(func() {
						cache.UnchangedSinceCachedReturns(existingObj)
					})

					It("does not create or patch any objects", func() {
						Expect(repo.EnsureObjectExistsOnCluster(stampedObj, true)).To(Succeed())
						Expect(cl.CreateCallCount()).To(Equal(0))
						Expect(cl.PatchCallCount()).To(Equal(0))
					})

					It("does not write to the submitted or persisted cache", func() {
						Expect(repo.EnsureObjectExistsOnCluster(stampedObj, true)).To(Succeed())
						Expect(cache.SetCallCount()).To(Equal(0))
					})

					It("refreshes the cache entry", func() {
						originalStampedObj := stampedObj.DeepCopy()

						Expect(repo.EnsureObjectExistsOnCluster(stampedObj, true)).To(Succeed())
						Expect(cache.RefreshCallCount()).To(Equal(1))
						Expect(cache.RefreshArgsForCall(0)).To(Equal(originalStampedObj))
					})

					It("populates the object passed into the function with the object in apiServer", func() {
						originalStampedObj := stampedObj.DeepCopy()

						_ = repo.EnsureObjectExistsOnCluster(stampedObj, true)

						Expect(stampedObj).To(Equal(existingObj))
						Expect(stampedObj).NotTo(Equal(originalStampedObj))
					})
				})

				Context("and the cache determines there has been a change since the last update", func() {
					BeforeEach(func() {
						cache.UnchangedSinceCachedReturns(nil)
					})

					Context("and allowUpdate is true", func() {
						Context("list has exactly one object", func() {
							It("patches the object", func() {
								Expect(repo.EnsureObjectExistsOnCluster(stampedObj, true)).To(Succeed())
								Expect(cl.PatchCallCount()).To(Equal(1))
							})

							Context("and the patch succeeds", func() {
								var returnedPatchedObj *unstructured.Unstructured

								BeforeEach(func() {
									returnedPatchedObj = stampedObj.DeepCopy()
									Expect(utils.AlterFieldOfNestedStringMaps(returnedPatchedObj.Object, "spec.template.spec.restartPolicy", "Never")).To(Succeed())
									cl.PatchStub = func(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
										objVal := reflect.ValueOf(obj)
										returnVal := reflect.ValueOf(returnedPatchedObj)

										reflect.Indirect(objVal).Set(reflect.Indirect(returnVal))
										return nil
									}
								})

								It("does not return an error", func() {
									Expect(repo.EnsureObjectExistsOnCluster(stampedObj, true)).To(Succeed())
								})

								It("caches the submitted and persisted objects, as the persisted one may be modified by mutating webhooks", func() {
									originalStampedObj := stampedObj.DeepCopy()

									Expect(repo.EnsureObjectExistsOnCluster(stampedObj, true)).To(Succeed())
									Expect(cache.SetCallCount()).To(Equal(1))
									submitted, persisted := cache.SetArgsForCall(0)
									Expect(*submitted).To(Equal(*originalStampedObj))
									Expect(*persisted).To(Equal(*returnedPatchedObj))
								})
							})

							Context("and the patch fails", func() {
								BeforeEach(func() {
									cl.PatchReturns(errors.New("some-error"))
								})
								It("returns a helpful error", func() {
									err := repo.EnsureObjectExistsOnCluster(stampedObj, true)
									Expect(err).To(MatchError(ContainSubstring("patch: some-error")))
								})

								It("does not write to the submitted or persisted cache", func() {
									_ = repo.EnsureObjectExistsOnCluster(stampedObj, true)
									Expect(cache.SetCallCount()).To(Equal(0))
								})
							})
						})

						Context("list has more than one object", func() {
							Context("and the list contains the correct object", func() {
								BeforeEach(func() {
									rogueObjectWithDuplicateLabels := existingObj.DeepCopy()
									Expect(utils.AlterFieldOfNestedStringMaps(rogueObjectWithDuplicateLabels.Object, "metadata.name", "goodbye")).To(Succeed())
									existingObjList = unstructured.UnstructuredList{
										Items: []unstructured.Unstructured{*existingObj, *rogueObjectWithDuplicateLabels},
									}
								})

								It("it patches", func() {
									Expect(repo.EnsureObjectExistsOnCluster(stampedObj, true)).To(Succeed())
									Expect(cl.PatchCallCount()).To(Equal(1))
								})
							})

							Context("and the list does not contain the correct object", func() {
								BeforeEach(func() {
									rogueObjectWithDuplicateLabels := existingObj.DeepCopy()
									Expect(utils.AlterFieldOfNestedStringMaps(rogueObjectWithDuplicateLabels.Object, "metadata.name", "goodbye")).To(Succeed())
									secondRogueObjectWithDuplicateLabels := existingObj.DeepCopy()
									Expect(utils.AlterFieldOfNestedStringMaps(secondRogueObjectWithDuplicateLabels.Object, "metadata.name", "farewell")).To(Succeed())
									existingObjList = unstructured.UnstructuredList{
										Items: []unstructured.Unstructured{*rogueObjectWithDuplicateLabels, *secondRogueObjectWithDuplicateLabels},
									}
								})
								It("it creates", func() {
									Expect(repo.EnsureObjectExistsOnCluster(stampedObj, true)).To(Succeed())
									Expect(cl.CreateCallCount()).To(Equal(1))
								})
							})
						})
					})

					Context("and allowUpate is false", func() {
						It("creates a new object", func() {
							Expect(repo.EnsureObjectExistsOnCluster(stampedObj, false)).To(Succeed())
							Expect(cl.PatchCallCount()).To(Equal(0))
							Expect(cl.CreateCallCount()).To(Equal(1))
						})

						Context("and the create succeeds", func() {
							var returnedCreatedObj *unstructured.Unstructured

							BeforeEach(func() {
								returnedCreatedObj = stampedObj.DeepCopy()
								Expect(utils.AlterFieldOfNestedStringMaps(returnedCreatedObj.Object, "spec.template.spec.restartPolicy", "Never")).To(Succeed())
								cl.CreateStub = func(ctx context.Context, object client.Object, option ...client.CreateOption) error {
									objVal := reflect.ValueOf(object)
									returnVal := reflect.ValueOf(returnedCreatedObj)

									reflect.Indirect(objVal).Set(reflect.Indirect(returnVal))
									return nil
								}
							})

							It("does not return an error", func() {
								Expect(repo.EnsureObjectExistsOnCluster(stampedObj, false)).To(Succeed())
							})

							It("caches the submitted and persisted objects, as the persisted one may be modified by mutating webhooks", func() {
								originalStampedObj := stampedObj.DeepCopy()

								Expect(repo.EnsureObjectExistsOnCluster(stampedObj, false)).To(Succeed())
								Expect(cache.SetCallCount()).To(Equal(1))
								submitted, persisted := cache.SetArgsForCall(0)
								Expect(*submitted).To(Equal(*originalStampedObj))
								Expect(*persisted).To(Equal(*returnedCreatedObj))
							})
						})

						Context("and the create fails", func() {
							BeforeEach(func() {
								cl.CreateReturns(errors.New("some-error"))
							})
							It("returns a helpful error", func() {
								err := repo.EnsureObjectExistsOnCluster(stampedObj, false)
								Expect(err).To(MatchError(ContainSubstring("create: some-error")))
							})

							It("does not write to the submitted or persisted cache", func() {
								_ = repo.EnsureObjectExistsOnCluster(stampedObj, false)
								Expect(cache.SetCallCount()).To(Equal(0))
							})
						})
					})
				})
			})
		})

		Context("GetSupplyChainsForWorkload", func() {
			BeforeEach(func() {
				cl.ListReturns(errors.New("some list error"))
			})

			It("attempts to list the object from the apiServer", func() {
				_, err := repo.GetSupplyChainsForWorkload(&v1alpha1.Workload{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("list supply chains:"))
			})
		})

		Context("GetSupplyChain", func() {
			BeforeEach(func() {
				cl.GetReturns(errors.New("some get error"))
			})

			It("attempts to get the object from the apiServer", func() {
				_, err := repo.GetSupplyChain("sc-name")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("get:"))
			})
		})

		Context("GetClusterTemplate", func() {
			Context("when the template reference kind is not in our gvk", func() {
				It("returns a helpful error", func() {
					reference := v1alpha1.ClusterTemplateReference{
						Kind: "some-unsupported-kind",
					}
					_, err := repo.GetClusterTemplate(reference)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("get api template:"))
				})
			})

			Context("when the client returns an error on the get", func() {
				BeforeEach(func() {
					cl.GetReturns(errors.New("some bad get error"))
				})
				It("returns a helpful error", func() {
					reference := v1alpha1.ClusterTemplateReference{
						Kind: "ClusterImageTemplate",
						Name: "image-template",
					}
					_, err := repo.GetClusterTemplate(reference)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("get:"))
				})
			})
		})

		Context("GetDeliveryClusterTemplate", func() {
			Context("when the template reference kind is not in our gvk", func() {
				It("returns a helpful error", func() {
					reference := v1alpha1.DeliveryClusterTemplateReference{
						Kind: "some-unsupported-kind",
					}
					_, err := repo.GetDeliveryClusterTemplate(reference)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("get api template:"))
				})
			})

			Context("when the client returns an error on the get", func() {
				BeforeEach(func() {
					cl.GetReturns(errors.New("some bad get error"))
				})
				It("returns a helpful error", func() {
					reference := v1alpha1.DeliveryClusterTemplateReference{
						Kind: "ClusterImageTemplate",
						Name: "image-template",
					}
					_, err := repo.GetDeliveryClusterTemplate(reference)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("get:"))
				})
			})
		})

		Describe("GetClusterDelivery", func() {
			Context("api errors on get", func() {
				var apiError error
				BeforeEach(func() {
					apiError = errors.New("my error")
					cl.GetReturns(apiError)
				})
				It("returns the error", func() {
					_, err := repo.GetDelivery("my-delivery")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(apiError.Error()))
				})
			})

			Context("no matching cluster delivery", func() {
				var apiError error
				BeforeEach(func() {
					apiError = kerrors.NewNotFound(
						schema.GroupResource{
							Group:    "carto.io/v1alpha1",
							Resource: "ClusterDelivery",
						},
						"my-delivery",
					)
					cl.GetReturns(apiError)
				})
				It("returns a nil result without error", func() {
					delivery, err := repo.GetDelivery("my-delivery")
					Expect(err).NotTo(HaveOccurred())
					Expect(delivery).To(BeNil())
				})
			})

			Context("one matching delivery", func() {
				var apiDelivery *v1alpha1.ClusterDelivery

				BeforeEach(func() {
					apiDelivery = &v1alpha1.ClusterDelivery{}
					//nolint:staticcheck,ineffassign
					cl.GetStub = func(_ context.Context, _ client.ObjectKey, obj client.Object) error {
						obj = apiDelivery
						return nil
					}
				})

				It("asks for the delivery by name", func() {
					_, err := repo.GetDelivery("my-delivery")
					Expect(err).NotTo(HaveOccurred())

					Expect(cl.GetCallCount()).To(Equal(1))
					_, key, _ := cl.GetArgsForCall(0)
					Expect(key).To(Equal(client.ObjectKey{
						Name:      "my-delivery",
						Namespace: "",
					}))
				})

				It("returns the delivery without error", func() {
					delivery, err := repo.GetDelivery("my-delivery")
					Expect(err).NotTo(HaveOccurred())

					Expect(delivery).To(Equal(apiDelivery))
				})
			})

		})

	})

	Describe("tests using apiMachinery fake client", func() {
		var (
			scheme            *runtime.Scheme
			fakeClientBuilder *fake.ClientBuilder
			clientObjects     []client.Object
			cl                client.Client
		)

		BeforeEach(func() {
			scheme = runtime.NewScheme()
			Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
			Expect(v1.AddToScheme(scheme)).To(Succeed())
		})

		JustBeforeEach(func() {
			fakeClientBuilder = fake.NewClientBuilder().WithScheme(scheme).WithObjects(clientObjects...)
			cl = fakeClientBuilder.Build()

			repo = repository.NewRepository(cl, cache)
		})

		Context("GetClusterTemplate", func() {
			BeforeEach(func() {
				template := &v1alpha1.ClusterSourceTemplate{
					ObjectMeta: metav1.ObjectMeta{
						Name: "some-name",
					},
				}
				clientObjects = []client.Object{template}
			})

			It("gets the template successfully", func() {
				templateRef := v1alpha1.ClusterTemplateReference{
					Kind: "ClusterSourceTemplate",
					Name: "some-name",
				}
				template, err := repo.GetClusterTemplate(templateRef)
				Expect(err).ToNot(HaveOccurred())
				Expect(template.GetName()).To(Equal("some-name"))
			})
		})

		Context("GetDeliveryClusterTemplate", func() {
			BeforeEach(func() {
				template := &v1alpha1.ClusterSourceTemplate{
					ObjectMeta: metav1.ObjectMeta{
						Name: "some-name",
					},
				}
				clientObjects = []client.Object{template}
			})

			It("gets the template successfully", func() {
				templateRef := v1alpha1.DeliveryClusterTemplateReference{
					Kind: "ClusterSourceTemplate",
					Name: "some-name",
				}
				template, err := repo.GetDeliveryClusterTemplate(templateRef)
				Expect(err).ToNot(HaveOccurred())
				Expect(template.GetName()).To(Equal("some-name"))
			})
		})

		Context("GetRunTemplate", func() {
			BeforeEach(func() {
				clientObjects = []client.Object{
					&v1alpha1.ClusterRunTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Name: "first-template",
						},
					},
					&v1alpha1.ClusterRunTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Name: "second-template",
						},
					}}
			})

			It("gets the template successfully", func() {
				templateRef := v1alpha1.TemplateReference{
					Kind: "ClusterRunTemplate",
					Name: "second-template",
				}
				template, err := repo.GetRunTemplate(templateRef)
				Expect(err).ToNot(HaveOccurred())
				Expect(template.GetName()).To(Equal("second-template"))
			})
		})

		Context("GetWorkload", func() {
			BeforeEach(func() {
				workload := &v1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "workload-name",
						Namespace: "workload-namespace",
					},
				}
				clientObjects = []client.Object{workload}
			})

			It("gets the workload successfully", func() {
				workload, err := repo.GetWorkload("workload-name", "workload-namespace")
				Expect(err).ToNot(HaveOccurred())
				Expect(workload.GetName()).To(Equal("workload-name"))
			})

			Context("workload doesnt exist", func() {
				It("returns an error", func() {
					_, err := repo.GetWorkload("workload-that-does-not-exist-name", "workload-namespace")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("get:"))
				})
			})
		})

		Context("GetDeliverable", func() {
			BeforeEach(func() {
				deliverable := &v1alpha1.Deliverable{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "deliverable-name",
						Namespace: "deliverable-namespace",
					},
				}
				clientObjects = []client.Object{deliverable}
			})

			It("gets the deliverable successfully", func() {
				workload, err := repo.GetDeliverable("deliverable-name", "deliverable-namespace")
				Expect(err).ToNot(HaveOccurred())
				Expect(workload.GetName()).To(Equal("deliverable-name"))
			})

			Context("deliverable doesnt exist", func() {
				It("returns an error", func() {
					_, err := repo.GetDeliverable("deliverable-that-does-not-exist-name", "deliverable-namespace")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("get:"))
				})
			})
		})

		Context("GetPipeline", func() {
			BeforeEach(func() {
				pipeline := &v1alpha1.Pipeline{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pipeline-name",
						Namespace: "pipeline-namespace",
					},
				}
				clientObjects = []client.Object{pipeline}
			})

			It("gets the pipeline successfully", func() {
				pipeline, err := repo.GetPipeline("pipeline-name", "pipeline-namespace")
				Expect(err).ToNot(HaveOccurred())
				Expect(pipeline.GetName()).To(Equal("pipeline-name"))
			})

			Context("pipeline doesnt exist", func() {
				It("returns an error", func() {
					_, err := repo.GetPipeline("pipeline-that-does-not-exist-name", "pipeline-namespace")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("get-pipeline:"))
				})
			})
		})

		Context("GetSupplyChain", func() {
			BeforeEach(func() {
				supplyChain := &v1alpha1.ClusterSupplyChain{
					ObjectMeta: metav1.ObjectMeta{
						Name: "sc-name",
					},
				}
				clientObjects = []client.Object{supplyChain}
			})

			It("gets the supply chain successfully", func() {
				sc, err := repo.GetSupplyChain("sc-name")
				Expect(err).ToNot(HaveOccurred())
				Expect(sc.GetName()).To(Equal("sc-name"))
			})

			Context("supply chain doesnt exist", func() {
				It("returns no error", func() {
					sc, err := repo.GetSupplyChain("sc-that-does-not-exist-name")
					Expect(err).ToNot(HaveOccurred())
					Expect(sc).To(BeNil())
				})
			})
		})

		Context("GetSupplyChainsForWorkload", func() {
			Context("One supply chain", func() {
				BeforeEach(func() {
					supplyChain := &v1alpha1.ClusterSupplyChain{
						TypeMeta: metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{
							Name: "supplychain-name",
						},
						Spec: v1alpha1.SupplyChainSpec{
							Selector: map[string]string{"foo": "bar"},
						},
					}
					clientObjects = []client.Object{supplyChain}
				})

				It("returns supply chains for workload", func() {
					workload := &v1alpha1.Workload{
						TypeMeta: metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{
							Name:   "workload-name",
							Labels: map[string]string{"foo": "bar"},
						},
						Spec:   v1alpha1.WorkloadSpec{},
						Status: v1alpha1.WorkloadStatus{},
					}
					supplyChains, err := repo.GetSupplyChainsForWorkload(workload)
					Expect(err).ToNot(HaveOccurred())
					Expect(len(supplyChains)).To(Equal(1))
					Expect(supplyChains[0].Name).To(Equal("supplychain-name"))
				})
			})

			Context("More than one supply chain", func() {
				BeforeEach(func() {
					supplyChain := &v1alpha1.ClusterSupplyChain{
						TypeMeta: metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{
							Name: "supplychain-name",
						},
						Spec: v1alpha1.SupplyChainSpec{
							Selector: map[string]string{"foo": "bar"},
						},
					}
					supplyChain2 := &v1alpha1.ClusterSupplyChain{
						TypeMeta: metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{
							Name: "supplychain-name2",
						},
						Spec: v1alpha1.SupplyChainSpec{
							Selector: map[string]string{"foo": "baz"},
						},
					}
					clientObjects = []client.Object{supplyChain, supplyChain2}
				})

				It("returns supply chains for workload", func() {
					workload := &v1alpha1.Workload{
						TypeMeta: metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{
							Name:   "workload-name",
							Labels: map[string]string{"foo": "bar"},
						},
						Spec:   v1alpha1.WorkloadSpec{},
						Status: v1alpha1.WorkloadStatus{},
					}
					supplyChains, err := repo.GetSupplyChainsForWorkload(workload)
					Expect(err).ToNot(HaveOccurred())
					Expect(len(supplyChains)).To(Equal(1))
					Expect(supplyChains[0].Name).To(Equal("supplychain-name"))
				})

				It("returns no supply chains for workload", func() {
					workload := &v1alpha1.Workload{
						TypeMeta: metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{
							Name:   "workload-name",
							Labels: map[string]string{"foo": "bat"},
						},
						Spec:   v1alpha1.WorkloadSpec{},
						Status: v1alpha1.WorkloadStatus{},
					}
					supplyChains, err := repo.GetSupplyChainsForWorkload(workload)
					Expect(err).ToNot(HaveOccurred())
					Expect(len(supplyChains)).To(Equal(0))
				})
			})
		})
	})
})
