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
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/events"
	"github.com/vmware-tanzu/cartographer/pkg/events/eventsfakes"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
	"github.com/vmware-tanzu/cartographer/tests/resources"
)

var _ = Describe("repository", func() {
	var (
		repo  repository.Repository
		cache *repositoryfakes.FakeRepoCache
		rm    *repositoryfakes.FakeRESTMapper
		rec   *eventsfakes.FakeOwnerEventRecorder
		ctx   context.Context
		out   *Buffer
	)

	BeforeEach(func() {
		out = NewBuffer()
		logger := zap.New(zap.WriteTo(out))
		ctx = logr.NewContext(context.Background(), logger)

		rec = &eventsfakes.FakeOwnerEventRecorder{}
		ctx = events.NewContext(ctx, rec)

		rm = &repositoryfakes.FakeRESTMapper{}

		cache = &repositoryfakes.FakeRepoCache{}
	})

	Describe("tests using counterfeiter client", func() {
		var cl *repositoryfakes.FakeClient

		BeforeEach(func() {
			cl = &repositoryfakes.FakeClient{}
			cl.RESTMapperReturns(rm)
			rm.RESTMappingReturns(&meta.RESTMapping{
				Resource: schema.GroupVersionResource{
					Group:    "example.com",
					Resource: "thing",
				},
			}, nil)
			repo = repository.NewRepository(cl, cache)
		})

		Context("EnsureMutableObjectExistsOnCluster", func() {
			var (
				stampedObj *unstructured.Unstructured
			)

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
				Expect(repo.EnsureMutableObjectExistsOnCluster(ctx, stampedObj)).To(Succeed())

				Expect(cl.GetCallCount()).To(Equal(1))

				_, namespacedName, obj, _ := cl.GetArgsForCall(0)
				Expect(namespacedName).To(Equal(types.NamespacedName{Namespace: "default", Name: "hello"}))
				Expect(obj.GetObjectKind().GroupVersionKind()).To(Equal(stampedObj.GroupVersionKind()))
			})

			Context("when the apiServer errors when trying to get the object", func() {
				BeforeEach(func() {
					cl.GetReturns(errors.New("some-error"))
				})

				It("returns a helpful error", func() {
					err := repo.EnsureMutableObjectExistsOnCluster(ctx, stampedObj)
					Expect(err).To(MatchError(ContainSubstring("failed to get unstructured [default/hello] from api server: some-error")))
				})

				It("does not create or patch any objects", func() {
					_ = repo.EnsureMutableObjectExistsOnCluster(ctx, stampedObj)
					Expect(cl.CreateCallCount()).To(Equal(0))
					Expect(cl.PatchCallCount()).To(Equal(0))
				})

				It("does not write to the submitted or persisted cache", func() {
					_ = repo.EnsureMutableObjectExistsOnCluster(ctx, stampedObj)
					Expect(cache.SetCallCount()).To(Equal(0))
				})

				It("does not record any events", func() {
					_ = repo.EnsureMutableObjectExistsOnCluster(ctx, stampedObj)
					Expect(rec.Invocations()).To(BeEmpty())
				})
			})

			Context("and the apiServer attempts to get the object and it doesn't exist", func() {
				BeforeEach(func() {
					cl.GetReturns(kerrors.NewNotFound(schema.GroupResource{}, ""))
				})

				It("attempts to create the object", func() {
					Expect(repo.EnsureMutableObjectExistsOnCluster(ctx, stampedObj)).To(Succeed())

					Expect(cl.CreateCallCount()).To(Equal(1))
					_, createCallObj, _ := cl.CreateArgsForCall(0)
					Expect(createCallObj).To(Equal(stampedObj))
				})

				Context("and the apiServer errors when creating the object", func() {
					BeforeEach(func() {
						cl.CreateReturns(errors.New("some-error"))
					})

					It("returns a helpful error", func() {
						err := repo.EnsureMutableObjectExistsOnCluster(ctx, stampedObj)
						Expect(err).To(MatchError(ContainSubstring("create: some-error")))
					})

					It("does not write to the submitted or persisted cache", func() {
						_ = repo.EnsureMutableObjectExistsOnCluster(ctx, stampedObj)
						Expect(cache.SetCallCount()).To(Equal(0))
					})

					It("does not record any events", func() {
						_ = repo.EnsureMutableObjectExistsOnCluster(ctx, stampedObj)
						Expect(rec.Invocations()).To(BeEmpty())
					})
				})

				Context("and the apiServer succeeds", func() {
					var returnedCreatedObj *unstructured.Unstructured
					BeforeEach(func() {
						returnedCreatedObj = stampedObj.DeepCopy()
						Expect(utils.AlterFieldOfNestedStringMaps(returnedCreatedObj.Object, "spec.template.spec.restartPolicy", "Never")).To(Succeed())
						cl.CreateStub = func(ctx context.Context, obj client.Object, _ ...client.CreateOption) error {
							objVal := reflect.ValueOf(obj)
							returnVal := reflect.ValueOf(returnedCreatedObj)

							reflect.Indirect(objVal).Set(reflect.Indirect(returnVal))
							return nil
						}
					})

					It("does not return an error", func() {
						Expect(repo.EnsureMutableObjectExistsOnCluster(ctx, stampedObj)).To(Succeed())
					})

					It("caches the submitted and persisted objects with no ownerDiscriminant, as the persisted one may be modified by mutating webhooks", func() {
						originalStampedObj := stampedObj.DeepCopy()

						Expect(repo.EnsureMutableObjectExistsOnCluster(ctx, stampedObj)).To(Succeed())
						Expect(cache.SetCallCount()).To(Equal(1))
						submitted, persisted, ownerDiscriminant := cache.SetArgsForCall(0)
						Expect(*submitted).To(Equal(*originalStampedObj))
						Expect(*persisted).To(Equal(*returnedCreatedObj))
						Expect(ownerDiscriminant).To(Equal(""))
					})

					It("records an StampedObjectApplied event", func() {
						_ = repo.EnsureMutableObjectExistsOnCluster(ctx, stampedObj)
						Expect(rec.EventfCallCount()).To(Equal(1))
						eventType, reason, message, fmtArgs := rec.EventfArgsForCall(0)
						Expect(eventType).To(Equal("Normal"))
						Expect(reason).To(Equal("StampedObjectApplied"))
						Expect(message).To(Equal("Created object [%s.%s/%s]"))
						Expect(fmtArgs).To(Equal([]interface{}{"thing", "example.com", "hello"}))
					})
				})
			})

			Context("and apiServer succeeds in getting the object", func() {
				var (
					existingObj *unstructured.Unstructured
				)

				BeforeEach(func() {
					existingObj = &unstructured.Unstructured{}
					existingObj.SetName("hello")
					existingObj.SetNamespace("default")
					existingObj.SetGeneration(5)

					cl.GetStub = func(ctx context.Context, key types.NamespacedName, obj client.Object, _ ...client.GetOption) error {
						objVal := reflect.ValueOf(obj)
						existingVal := reflect.ValueOf(existingObj)

						reflect.Indirect(objVal).Set(reflect.Indirect(existingVal))
						return nil
					}
				})

				It("the cache is consulted to see if there was a change since the last time the cache was updated", func() {
					Expect(repo.EnsureMutableObjectExistsOnCluster(ctx, stampedObj)).To(Succeed())
					Expect(cache.UnchangedSinceCachedCallCount()).To(Equal(1))

					submitted, persisted := cache.UnchangedSinceCachedArgsForCall(0)
					Expect(*submitted).To(Equal(*stampedObj))
					Expect(persisted).To(Equal(existingObj))
				})

				Context("and the cache determines there has been no change since the last update", func() {
					BeforeEach(func() {
						cache.UnchangedSinceCachedReturns(existingObj)
					})

					It("does not create or patch any objects", func() {
						Expect(repo.EnsureMutableObjectExistsOnCluster(ctx, stampedObj)).To(Succeed())
						Expect(cl.CreateCallCount()).To(Equal(0))
						Expect(cl.PatchCallCount()).To(Equal(0))
					})

					It("does not write to the submitted or persisted cache", func() {
						Expect(repo.EnsureMutableObjectExistsOnCluster(ctx, stampedObj)).To(Succeed())
						Expect(cache.SetCallCount()).To(Equal(0))
					})

					It("populates the object passed into the function with the object in apiServer", func() {
						originalStampedObj := stampedObj.DeepCopy()

						_ = repo.EnsureMutableObjectExistsOnCluster(ctx, stampedObj)

						Expect(stampedObj).To(Equal(existingObj))
						Expect(stampedObj).NotTo(Equal(originalStampedObj))
					})

					It("does not record any events", func() {
						_ = repo.EnsureMutableObjectExistsOnCluster(ctx, stampedObj)
						Expect(rec.Invocations()).To(BeEmpty())
					})
				})

				Context("and the cache determines there has been a change since the last update", func() {
					BeforeEach(func() {
						cache.UnchangedSinceCachedReturns(nil)
					})

					Context("and allowUpdate is true", func() {
						Context("list has exactly one object", func() {
							It("patches the object", func() {
								Expect(repo.EnsureMutableObjectExistsOnCluster(ctx, stampedObj)).To(Succeed())
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
									Expect(repo.EnsureMutableObjectExistsOnCluster(ctx, stampedObj)).To(Succeed())
								})

								It("caches the submitted and persisted objects with no ownerDiscriminant, as the persisted one may be modified by mutating webhooks", func() {
									originalStampedObj := stampedObj.DeepCopy()

									Expect(repo.EnsureMutableObjectExistsOnCluster(ctx, stampedObj)).To(Succeed())
									Expect(cache.SetCallCount()).To(Equal(1))
									submitted, persisted, ownerDiscriminant := cache.SetArgsForCall(0)
									Expect(*submitted).To(Equal(*originalStampedObj))
									Expect(*persisted).To(Equal(*returnedPatchedObj))
									Expect(ownerDiscriminant).To(Equal(""))
								})

								It("records an StampedObjectApplied event", func() {
									_ = repo.EnsureMutableObjectExistsOnCluster(ctx, stampedObj)
									Expect(rec.EventfCallCount()).To(Equal(1))
									eventType, reason, message, fmtArgs := rec.EventfArgsForCall(0)
									Expect(eventType).To(Equal("Normal"))
									Expect(reason).To(Equal("StampedObjectApplied"))
									Expect(message).To(Equal("Patched object [%s.%s/%s]"))
									Expect(fmtArgs).To(Equal([]interface{}{"thing", "example.com", "hello"}))
								})
							})

							Context("and the patch fails", func() {
								BeforeEach(func() {
									cl.PatchReturns(errors.New("some-error"))
								})
								It("returns a helpful error", func() {
									err := repo.EnsureMutableObjectExistsOnCluster(ctx, stampedObj)
									Expect(err).To(MatchError(ContainSubstring("patch: some-error")))
								})

								It("does not write to the submitted or persisted cache", func() {
									_ = repo.EnsureMutableObjectExistsOnCluster(ctx, stampedObj)
									Expect(cache.SetCallCount()).To(Equal(0))
								})

								It("does not record any events", func() {
									_ = repo.EnsureMutableObjectExistsOnCluster(ctx, stampedObj)
									Expect(rec.Invocations()).To(BeEmpty())
								})
							})
						})
					})
				})
			})
		})

		Context("EnsureImmutableObjectExistsOnCluster", func() {
			var (
				stampedObj *unstructured.Unstructured
				labels     map[string]string
			)

			BeforeEach(func() {
				labels = map[string]string{"quux": "xyzzy", "foo": "bar", "waldo": "fred"}
				stampedObj = &unstructured.Unstructured{}
				stampedObjManifest := `
apiVersion: batch/v1
kind: Job
metadata:
  name: hello
  namespace: default
  labels:
    foo: bar
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

			Context("and apiServer succeeds in getting the list of objects", func() {
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

					cl.ListStub = func(ctx context.Context, list client.ObjectList, options ...client.ListOption) error {
						listVal := reflect.ValueOf(list)
						existingVal := reflect.ValueOf(existingObjList)

						reflect.Indirect(listVal).Set(reflect.Indirect(existingVal))
						return nil
					}
				})

				It("the cache is consulted (with an ownerDiscriminant of the labels) to see if there was a change since the last time the cache was updated", func() {
					Expect(repo.EnsureImmutableObjectExistsOnCluster(ctx, stampedObj, labels)).To(Succeed())
					Expect(cache.UnchangedSinceCachedFromListCallCount()).To(Equal(1))

					submitted, persisted, ownerDiscriminant := cache.UnchangedSinceCachedFromListArgsForCall(0)
					Expect(*submitted).To(Equal(*stampedObj))
					Expect(persisted[0]).To(Equal(existingObj))
					Expect(ownerDiscriminant).To(Equal("{foo:bar}{quux:xyzzy}{waldo:fred}"))
				})

				Context("and the cache determines there has been no change since the last update", func() {
					BeforeEach(func() {
						cache.UnchangedSinceCachedFromListReturns(existingObj)
					})

					It("does not create any objects", func() {
						Expect(repo.EnsureImmutableObjectExistsOnCluster(ctx, stampedObj, labels)).To(Succeed())
						Expect(cl.CreateCallCount()).To(Equal(0))
					})

					It("does not write to the submitted or persisted cache", func() {
						Expect(repo.EnsureImmutableObjectExistsOnCluster(ctx, stampedObj, labels)).To(Succeed())
						Expect(cache.SetCallCount()).To(Equal(0))
					})

					It("populates the object passed into the function with the object in apiServer", func() {
						originalStampedObj := stampedObj.DeepCopy()

						_ = repo.EnsureImmutableObjectExistsOnCluster(ctx, stampedObj, labels)

						Expect(stampedObj).To(Equal(existingObj))
						Expect(stampedObj).NotTo(Equal(originalStampedObj))
					})

					It("does not record any events", func() {
						_ = repo.EnsureImmutableObjectExistsOnCluster(ctx, stampedObj, labels)
						Expect(rec.Invocations()).To(BeEmpty())
					})
				})

				Context("and the cache determines there has been a change since the last update", func() {
					BeforeEach(func() {
						cache.UnchangedSinceCachedReturns(nil)
					})

					It("creates a new object", func() {
						Expect(repo.EnsureImmutableObjectExistsOnCluster(ctx, stampedObj, labels)).To(Succeed())
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
							Expect(repo.EnsureImmutableObjectExistsOnCluster(ctx, stampedObj, labels)).To(Succeed())
						})

						It("caches the submitted and persisted objects with an ownerDiscriminant of the labels, as the persisted one may be modified by mutating webhooks", func() {
							originalStampedObj := stampedObj.DeepCopy()

							Expect(repo.EnsureImmutableObjectExistsOnCluster(ctx, stampedObj, labels)).To(Succeed())
							Expect(cache.SetCallCount()).To(Equal(1))
							submitted, persisted, ownerDiscriminant := cache.SetArgsForCall(0)
							Expect(*submitted).To(Equal(*originalStampedObj))
							Expect(*persisted).To(Equal(*returnedCreatedObj))
							Expect(ownerDiscriminant).To(Equal("{foo:bar}{quux:xyzzy}{waldo:fred}"))
						})

						It("records a StampedObjectApplied event", func() {
							_ = repo.EnsureImmutableObjectExistsOnCluster(ctx, stampedObj, labels)
							Expect(rec.EventfCallCount()).To(Equal(1))
							eventType, reason, message, fmtArgs := rec.EventfArgsForCall(0)
							Expect(eventType).To(Equal("Normal"))
							Expect(reason).To(Equal("StampedObjectApplied"))
							Expect(message).To(Equal("Created object [%s.%s/%s]"))
							Expect(fmtArgs).To(Equal([]interface{}{"thing", "example.com", "hello"}))
						})
					})

					Context("and the create fails", func() {
						BeforeEach(func() {
							cl.CreateReturns(errors.New("some-error"))
						})
						It("returns a helpful error", func() {
							err := repo.EnsureImmutableObjectExistsOnCluster(ctx, stampedObj, labels)
							Expect(err).To(MatchError(ContainSubstring("create: some-error")))
						})

						It("does not write to the submitted or persisted cache", func() {
							_ = repo.EnsureImmutableObjectExistsOnCluster(ctx, stampedObj, labels)
							Expect(cache.SetCallCount()).To(Equal(0))
						})

						It("does not record any events", func() {
							_ = repo.EnsureImmutableObjectExistsOnCluster(ctx, stampedObj, labels)
							Expect(rec.Invocations()).To(BeEmpty())
						})
					})
				})
			})
		})

		Context("ListUnstructured", func() {
			It("attempts to list objects from the apiServer", func() {
				namespace := "some-namespace"
				labels := map[string]string{"some-key": "some-value"}
				gvk := schema.GroupVersionKind{}

				_, err := repo.ListUnstructured(ctx, gvk, namespace, labels)
				Expect(err).NotTo(HaveOccurred())

				Expect(cl.ListCallCount()).To(Equal(1))

				expectedListOptions := []client.ListOption{
					client.InNamespace(namespace),
					client.MatchingLabels(labels),
				}

				_, objectList, options := cl.ListArgsForCall(0)
				unstructuredList, ok := objectList.(*unstructured.UnstructuredList)
				Expect(ok).To(BeTrue())
				Expect(len(unstructuredList.Items)).To(Equal(0))
				Expect(options).To(Equal(expectedListOptions))
				Expect(unstructuredList.GetObjectKind().GroupVersionKind()).To(Equal(gvk))
			})
		})

		Context("GetSupplyChainsForWorkload", func() {
			BeforeEach(func() {
				cl.ListReturns(errors.New("some list error"))
			})

			It("attempts to list the object from the apiServer", func() {
				_, err := repo.GetSupplyChainsForWorkload(ctx, &v1alpha1.Workload{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unable to list supply chains from api server: some list error"))
			})
		})

		Context("GetSupplyChain", func() {
			BeforeEach(func() {
				cl.GetReturns(errors.New("some get error"))
			})

			It("attempts to get the object from the apiServer", func() {
				_, err := repo.GetSupplyChain(ctx, "sc-name")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to get supply chain object from api server [sc-name]: failed to get object [sc-name] from api server: some get error"))
			})
		})

		Context("GetUnstructured", func() {
			Context("get returns an error", func() {
				Context("the error is of type IsNotFound", func() {
					BeforeEach(func() {
						cl.GetReturns(kerrors.NewNotFound(schema.GroupResource{}, ""))
					})

					It("returns a nil object and no error", func() {
						obj := &unstructured.Unstructured{}

						obj.SetGroupVersionKind(schema.GroupVersionKind{
							Group:   "my-group",
							Version: "my-version",
							Kind:    "my-kind",
						})
						obj.SetNamespace("my-ns")
						obj.SetName("my-name")
						returnedObj, err := repo.GetUnstructured(ctx, obj)
						Expect(err).NotTo(HaveOccurred())
						Expect(returnedObj).To(BeNil())
					})
				})

				Context("the error is not of type IsNotFound", func() {
					BeforeEach(func() {
						cl.GetReturns(errors.New("some get error"))
					})

					It("errors with a helfpul error message", func() {
						obj := &unstructured.Unstructured{}

						obj.SetGroupVersionKind(schema.GroupVersionKind{
							Group:   "my-group",
							Version: "my-version",
							Kind:    "my-kind",
						})
						obj.SetNamespace("my-ns")
						obj.SetName("my-name")
						_, err := repo.GetUnstructured(ctx, obj)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("failed to get unstructured [my-ns/my-name] from api server: some get error"))
					})
				})
			})

			Context("get does not return an error", func() {
				var existingObj *unstructured.Unstructured
				BeforeEach(func() {
					existingObj = &unstructured.Unstructured{}
					existingObj.SetName("hello")
					existingObj.SetNamespace("default")
					existingObj.SetGeneration(5)

					cl.GetStub = func(ctx context.Context, key types.NamespacedName, obj client.Object, _ ...client.GetOption) error {
						objVal := reflect.ValueOf(obj)
						existingVal := reflect.ValueOf(existingObj)

						reflect.Indirect(objVal).Set(reflect.Indirect(existingVal))
						return nil
					}
				})

				It("successfully gets unstructured from api server", func() {
					obj := &unstructured.Unstructured{}

					obj.SetGroupVersionKind(schema.GroupVersionKind{
						Group:   "my-group",
						Version: "my-version",
						Kind:    "my-kind",
					})
					obj.SetNamespace("my-ns")
					obj.SetName("my-name")
					returnedObj, err := repo.GetUnstructured(ctx, obj)
					Expect(err).NotTo(HaveOccurred())
					Expect(returnedObj).To(Equal(existingObj))
				})
			})
		})

		Context("GetTemplate", func() {
			Context("when the template reference kind is not in our gvk", func() {
				It("returns a helpful error", func() {
					reference := v1alpha1.DeliveryTemplateReference{
						Kind: "some-unsupported-kind",
						Name: "my-template",
					}
					_, err := repo.GetTemplate(ctx, reference.Name, reference.Kind)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("unable to get api template [some-unsupported-kind/my-template]: resource does not have valid kind: some-unsupported-kind"))
				})
			})

			Context("when the client returns an error on the get", func() {
				BeforeEach(func() {
					cl.GetReturns(errors.New("some bad get error"))
				})
				It("returns a helpful error", func() {
					reference := v1alpha1.DeliveryTemplateReference{
						Kind: "ClusterImageTemplate",
						Name: "image-template",
					}
					_, err := repo.GetTemplate(ctx, reference.Name, reference.Kind)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to get template object from api server [ClusterImageTemplate/image-template]: failed to get object [image-template] from api server: some bad get error"))
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
					_, err := repo.GetDelivery(ctx, "my-delivery")
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
				It("returns a nil result without error and without logging an error", func() {
					delivery, err := repo.GetDelivery(ctx, "my-delivery")
					Expect(err).NotTo(HaveOccurred())
					Expect(delivery).To(BeNil())

					Expect(out.Contents()).To(BeEmpty())
				})
			})

			Context("one matching delivery", func() {
				var apiDelivery *v1alpha1.ClusterDelivery

				BeforeEach(func() {
					apiDelivery = &v1alpha1.ClusterDelivery{}
					//nolint:staticcheck,ineffassign
					cl.GetStub = func(_ context.Context, _ client.ObjectKey, obj client.Object, _ ...client.GetOption) error {
						obj = apiDelivery
						return nil
					}
				})

				It("asks for the delivery by name", func() {
					_, err := repo.GetDelivery(ctx, "my-delivery")
					Expect(err).NotTo(HaveOccurred())

					Expect(cl.GetCallCount()).To(Equal(1))
					_, key, _, _ := cl.GetArgsForCall(0)
					Expect(key).To(Equal(client.ObjectKey{
						Name:      "my-delivery",
						Namespace: "",
					}))
				})

				It("returns the delivery without error", func() {
					delivery, err := repo.GetDelivery(ctx, "my-delivery")
					Expect(err).NotTo(HaveOccurred())

					Expect(delivery).To(Equal(apiDelivery))
				})
			})

		})

		Describe("GetServiceAccount", func() {
			Context("when the service account exists", func() {
				var (
					serviceAccount     *v1.ServiceAccount
					serviceAccountName string
					serviceAccountNS   string
				)

				BeforeEach(func() {
					serviceAccountName = "my-service-account"
					serviceAccountNS = "best-namespace-ever"

					serviceAccount = &v1.ServiceAccount{
						TypeMeta: metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{
							Name:      serviceAccountName,
							Namespace: serviceAccountNS,
						},
					}

					cl.GetStub = func(_ context.Context, key client.ObjectKey, obj client.Object, _ ...client.GetOption) error {
						if key.Name == serviceAccountName && key.Namespace == serviceAccountNS {
							bytes, _ := json.Marshal(serviceAccount)
							_ = json.Unmarshal(bytes, obj)
						} else {
							Fail(fmt.Sprintf("unexpected get call for name %s", key.Name))
						}
						return nil
					}
				})

				It("returns the named service account retrieved from the apiserver", func() {
					serviceAccount, err := repo.GetServiceAccount(context.TODO(), serviceAccountName, serviceAccountNS)
					Expect(err).NotTo(HaveOccurred())

					Expect(cl.GetCallCount()).To(Equal(1))
					_, getKey, obj, _ := cl.GetArgsForCall(0)
					Expect(getKey.Name).To(Equal(serviceAccountName))
					Expect(getKey.Namespace).To(Equal(serviceAccountNS))
					_, isGettingServiceAccount := obj.(*v1.ServiceAccount)
					Expect(isGettingServiceAccount).To(BeTrue())

					Expect(serviceAccount).To(Equal(serviceAccount))
				})
			})

			Context("when there is an error getting the service account", func() {
				BeforeEach(func() {
					cl.GetReturnsOnCall(0, fmt.Errorf("some error"))
				})

				It("returns a helpful error message", func() {
					_, err := repo.GetServiceAccount(context.TODO(), "some-service-account", "dont-matter")
					Expect(err).To(HaveOccurred())

					Expect(err.Error()).To(ContainSubstring("failed to get service account object from api server [dont-matter/some-service-account]: failed to get object [dont-matter/some-service-account] from api server: some error"))
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
			Expect(resources.AddToScheme(scheme)).To(Succeed())
			Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
			Expect(v1.AddToScheme(scheme)).To(Succeed())
		})

		JustBeforeEach(func() {
			fakeClientBuilder = fake.NewClientBuilder().WithScheme(scheme).WithObjects(clientObjects...)
			cl = fakeClientBuilder.Build()

			repo = repository.NewRepository(cl, cache)
		})

		Context("GetTemplate", func() {
			BeforeEach(func() {
				template := &v1alpha1.ClusterSourceTemplate{
					ObjectMeta: metav1.ObjectMeta{
						Name: "some-name",
					},
				}
				clientObjects = []client.Object{template}
			})

			It("gets the template successfully", func() {
				templateRef := v1alpha1.SupplyChainTemplateReference{
					Kind: "ClusterSourceTemplate",
					Name: "some-name",
				}
				template, err := repo.GetTemplate(ctx, templateRef.Name, templateRef.Kind)
				Expect(err).ToNot(HaveOccurred())
				Expect(template.GetName()).To(Equal("some-name"))
			})
		})

		Context("GetTemplate", func() {
			BeforeEach(func() {
				template := &v1alpha1.ClusterSourceTemplate{
					ObjectMeta: metav1.ObjectMeta{
						Name: "some-name",
					},
				}
				clientObjects = []client.Object{template}
			})

			It("gets the template successfully", func() {
				templateRef := v1alpha1.DeliveryTemplateReference{
					Kind: "ClusterSourceTemplate",
					Name: "some-name",
				}
				template, err := repo.GetTemplate(ctx, templateRef.Name, templateRef.Kind)
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
				template, err := repo.GetRunTemplate(ctx, templateRef)
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
				workload, err := repo.GetWorkload(ctx, "workload-name", "workload-namespace")
				Expect(err).ToNot(HaveOccurred())
				Expect(workload.GetName()).To(Equal("workload-name"))
			})

			Context("workload doesnt exist", func() {
				It("returns nil workload without logging an error", func() {
					workload, err := repo.GetWorkload(ctx, "workload-that-does-not-exist-name", "workload-namespace")
					Expect(err).NotTo(HaveOccurred())
					Expect(workload).To(BeNil())

					Expect(out.Contents()).To(BeEmpty())
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
				workload, err := repo.GetDeliverable(ctx, "deliverable-name", "deliverable-namespace")
				Expect(err).ToNot(HaveOccurred())
				Expect(workload.GetName()).To(Equal("deliverable-name"))
			})

			Context("deliverable doesnt exist", func() {
				It("returns nil deliverable without logging an error", func() {
					deliverable, err := repo.GetDeliverable(ctx, "deliverable-that-does-not-exist-name", "deliverable-namespace")
					Expect(err).NotTo(HaveOccurred())
					Expect(deliverable).To(BeNil())

					Expect(out.Contents()).To(BeEmpty())
				})
			})
		})

		Context("GetRunnable", func() {
			BeforeEach(func() {
				runnable := &v1alpha1.Runnable{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "runnable-name",
						Namespace: "runnable-namespace",
					},
				}
				clientObjects = []client.Object{runnable}
			})

			It("gets the runnable successfully", func() {
				runnable, err := repo.GetRunnable(ctx, "runnable-name", "runnable-namespace")
				Expect(err).ToNot(HaveOccurred())
				Expect(runnable.GetName()).To(Equal("runnable-name"))
			})

			Context("runnable doesnt exist", func() {
				It("returns nil runnable without logging an error", func() {
					runnable, err := repo.GetRunnable(ctx, "runnable-that-does-not-exist-name", "runnable-namespace")
					Expect(err).NotTo(HaveOccurred())
					Expect(runnable).To(BeNil())

					Expect(out.Contents()).To(BeEmpty())
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
				sc, err := repo.GetSupplyChain(ctx, "sc-name")
				Expect(err).ToNot(HaveOccurred())
				Expect(sc.GetName()).To(Equal("sc-name"))
			})

			Context("supply chain doesnt exist", func() {
				It("returns no error without logging an error", func() {
					sc, err := repo.GetSupplyChain(ctx, "sc-that-does-not-exist-name")
					Expect(err).ToNot(HaveOccurred())
					Expect(sc).To(BeNil())

					Expect(out.Contents()).To(BeEmpty())
				})
			})
		})

		Context("GetSupplyChainsForWorkload", func() {
			Context("When a matchFields key is invalid", func() {
				BeforeEach(func() {
					var supplyChain = &v1alpha1.ClusterSupplyChain{
						TypeMeta: metav1.TypeMeta{
							Kind: "ClusterSupplyChain",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "supplychain-name",
						},
						Spec: v1alpha1.SupplyChainSpec{
							LegacySelector: v1alpha1.LegacySelector{
								SelectorMatchFields: []v1alpha1.FieldSelectorRequirement{
									{
										Key:      "spec.env[asdfasdfadkf3",
										Operator: "Exists",
									},
								},
							},
						},
					}
					supplyChain.GetObjectKind()
					clientObjects = []client.Object{supplyChain}
				})

				It("returns an error", func() {
					workload := &v1alpha1.Workload{
						TypeMeta: metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "workload-name",
							Namespace: "myNS",
							Labels:    map[string]string{"foo": "bar"},
						},
						Spec:   v1alpha1.WorkloadSpec{},
						Status: v1alpha1.WorkloadStatus{},
					}
					_, err := repo.GetSupplyChainsForWorkload(ctx, workload)
					Expect(err).To(MatchError(ContainSubstring("evaluating supply chain selectors against workload [myNS/workload-name] failed")))
					Expect(err).To(MatchError(ContainSubstring("error handling selectors, selectorMatchExpressions or selectorMatchFields of [ClusterSupplyChain/supplychain-name]")))
					Expect(err).To(MatchError(ContainSubstring("unable to match field requirement with key [spec.env[asdfasdfadkf3] operator [Exists] values [[]]")))
				})
			})

			Context("One supply chain", func() {
				BeforeEach(func() {
					supplyChain := &v1alpha1.ClusterSupplyChain{
						TypeMeta: metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{
							Name: "supplychain-name",
						},
						Spec: v1alpha1.SupplyChainSpec{
							LegacySelector: v1alpha1.LegacySelector{
								Selector: map[string]string{"foo": "bar"},
							},
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
					supplyChains, err := repo.GetSupplyChainsForWorkload(ctx, workload)
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
							LegacySelector: v1alpha1.LegacySelector{
								Selector: map[string]string{"foo": "bar"},
							},
						},
					}
					supplyChain2 := &v1alpha1.ClusterSupplyChain{
						TypeMeta: metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{
							Name: "supplychain-name2",
						},
						Spec: v1alpha1.SupplyChainSpec{
							LegacySelector: v1alpha1.LegacySelector{
								Selector: map[string]string{"foo": "baz"},
							},
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
					supplyChains, err := repo.GetSupplyChainsForWorkload(ctx, workload)
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
					supplyChains, err := repo.GetSupplyChainsForWorkload(ctx, workload)
					Expect(err).ToNot(HaveOccurred())
					Expect(len(supplyChains)).To(Equal(0))
				})
			})
		})

		Context("Delete", func() {
			Context("when the object to be deleted does not exist", func() {
				It("logs and returns an error if the object is not present", func() {
					testObj := &unstructured.Unstructured{}
					stampedObjManifest := utils.HereYaml(`
						apiVersion: test.run/v1alpha1
						kind: TestObj
						metadata:
						  name: hello
						  namespace: default
						`)
					dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
					_, _, err := dec.Decode([]byte(stampedObjManifest), nil, testObj)
					Expect(err).NotTo(HaveOccurred())
					err = repo.Delete(ctx, testObj)

					Expect(err).To(MatchError("failed to delete object [default/hello]: testobjs.test.run \"hello\" not found"))
					Expect(out).To(Say("failed to delete object"))
				})
			})

			Context("when the object to be deleted exists", func() {
				var testObj *unstructured.Unstructured

				BeforeEach(func() {
					testObj = &unstructured.Unstructured{}
					stampedObjManifest := utils.HereYaml(`
						apiVersion: test.run/v1alpha1
						kind: TestObj
						metadata:
						  name: hello
						  namespace: default
						`)
					dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
					_, _, err := dec.Decode([]byte(stampedObjManifest), nil, testObj)
					Expect(err).NotTo(HaveOccurred())

					clientObjects = []client.Object{testObj}
				})

				It("throws no error and the object is gone", func() {
					err := repo.Delete(ctx, testObj)
					Expect(err).NotTo(HaveOccurred())

					obj := &resources.TestObj{}
					err = cl.Get(ctx, client.ObjectKey{
						Namespace: "default",
						Name:      "hello",
					}, obj)
					Expect(err).To(MatchError("testobjs.test.run \"hello\" not found"))
				})
			})
		})
	})
})
