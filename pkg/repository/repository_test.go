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
	"strings"

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
)

func alterFieldOfNestedStringMaps(obj interface{}, key string, value string) error {
	aMap, ok := obj.(map[string]interface{})
	if !ok {
		return errors.New("field not found")
	}

	i := strings.Index(key, ".")
	if i < 0 {
		_, ok = aMap[key]
		if !ok {
			return errors.New("field not found")
		}
		aMap[key] = value
	} else {
		keyPrefix := key[:i]
		keySuffix := key[i+1:]
		subMap, ok := aMap[keyPrefix]
		if !ok {
			return errors.New("field not found")
		}
		return alterFieldOfNestedStringMaps(subMap, keySuffix, value)
	}

	return nil
}

var _ = Describe("Repository", func() {
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

		Context("CreateOrPatchUnstructuredObject", func() {
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
				Expect(repo.CreateOrPatchUnstructuredObject(stampedObj)).To(Succeed())

				Expect(cl.GetCallCount()).To(Equal(1))

				stampedObjKey := client.ObjectKey{
					Namespace: stampedObj.GetNamespace(),
					Name:      stampedObj.GetName(),
				}

				_, getCallKey, blankObjToBePopulatedByClient := cl.GetArgsForCall(0)
				Expect(getCallKey).To(Equal(stampedObjKey))
				Expect(blankObjToBePopulatedByClient.GetObjectKind().GroupVersionKind()).To(Equal(stampedObj.GroupVersionKind()))
			})

			Context("when the apiServer errors when trying to get the object", func() {
				BeforeEach(func() {
					cl.GetReturns(errors.New("some-error"))
				})

				It("returns a helpful error", func() {
					err := repo.CreateOrPatchUnstructuredObject(stampedObj)
					Expect(err).To(MatchError(ContainSubstring("get: some-error")))
				})

				It("does not create or patch any objects", func() {
					_ = repo.CreateOrPatchUnstructuredObject(stampedObj)
					Expect(cl.CreateCallCount()).To(Equal(0))
					Expect(cl.PatchCallCount()).To(Equal(0))
				})

				It("does not write to the submitted or persisted cache", func() {
					_ = repo.CreateOrPatchUnstructuredObject(stampedObj)
					Expect(cache.SetCallCount()).To(Equal(0))
				})
			})

			Context("and the apiServer attempts to get the object and it doesn't exist", func() {
				BeforeEach(func() {
					groupResource := schema.GroupResource{
						Group:    stampedObj.GetAPIVersion(),
						Resource: stampedObj.GetKind(),
					}
					cl.GetReturns(kerrors.NewNotFound(groupResource, stampedObj.GetName()))
				})

				It("attempts to create the object", func() {
					Expect(repo.CreateOrPatchUnstructuredObject(stampedObj)).To(Succeed())

					Expect(cl.CreateCallCount()).To(Equal(1))
					_, createCallObj, _ := cl.CreateArgsForCall(0)
					Expect(createCallObj).To(Equal(stampedObj))
				})

				Context("and the apiServer errors when creating the object", func() {
					BeforeEach(func() {
						cl.CreateReturns(errors.New("some-error"))
					})

					It("returns a helpful error", func() {
						err := repo.CreateOrPatchUnstructuredObject(stampedObj)
						Expect(err).To(MatchError(ContainSubstring("create: some-error")))
					})

					It("does not write to the submitted or persisted cache", func() {
						_ = repo.CreateOrPatchUnstructuredObject(stampedObj)
						Expect(cache.SetCallCount()).To(Equal(0))
					})
				})

				Context("and the apiServer succeeds", func() {
					var returnedCreatedObj *unstructured.Unstructured
					BeforeEach(func() {
						returnedCreatedObj = stampedObj.DeepCopy()
						Expect(alterFieldOfNestedStringMaps(returnedCreatedObj.Object, "spec.template.spec.restartPolicy", "Never")).To(Succeed())
						cl.CreateStub = func(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
							objVal := reflect.ValueOf(obj)
							returnVal := reflect.ValueOf(returnedCreatedObj)

							reflect.Indirect(objVal).Set(reflect.Indirect(returnVal))
							return nil
						}
					})

					It("does not return an error", func() {
						Expect(repo.CreateOrPatchUnstructuredObject(stampedObj)).To(Succeed())
					})

					It("caches the submitted and persisted objects, as the persisted one may be modified by mutating webhooks", func() {
						originalStampedObj := stampedObj.DeepCopy()

						Expect(repo.CreateOrPatchUnstructuredObject(stampedObj)).To(Succeed())
						Expect(cache.SetCallCount()).To(Equal(1))
						submitted, persisted := cache.SetArgsForCall(0)
						Expect(*submitted).To(Equal(*originalStampedObj))
						Expect(*persisted).To(Equal(*returnedCreatedObj))
					})
				})
			})

			Context("and apiServer succeeds in getting the object", func() {
				var existingObj *unstructured.Unstructured

				BeforeEach(func() {
					existingObj = &unstructured.Unstructured{}
					existingObj.SetName("im-the-previous-one")
					cl.GetStub = func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						objVal := reflect.ValueOf(obj)
						prevVal := reflect.ValueOf(existingObj)

						reflect.Indirect(objVal).Set(reflect.Indirect(prevVal))
						return nil
					}
				})

				It("the cache is consulted to see if there was a change since the last time the cache was updated", func() {
					Expect(repo.CreateOrPatchUnstructuredObject(stampedObj)).To(Succeed())
					Expect(cache.UnchangedSinceCachedCallCount()).To(Equal(1))

					submitted, persisted := cache.UnchangedSinceCachedArgsForCall(0)
					Expect(*submitted).To(Equal(*stampedObj))
					Expect(*persisted).To(Equal(*existingObj))
				})

				Context("and the cache determines there has been no change since the last update", func() {
					BeforeEach(func() {
						cache.UnchangedSinceCachedReturns(true)
					})

					It("does not create or patch any objects", func() {
						Expect(repo.CreateOrPatchUnstructuredObject(stampedObj)).To(Succeed())
						Expect(cl.CreateCallCount()).To(Equal(0))
						Expect(cl.PatchCallCount()).To(Equal(0))
					})

					It("does not write to the submitted or persisted cache", func() {
						Expect(repo.CreateOrPatchUnstructuredObject(stampedObj)).To(Succeed())
						Expect(cache.SetCallCount()).To(Equal(0))
					})

					It("refreshes the cache entry", func() {
						Expect(repo.CreateOrPatchUnstructuredObject(stampedObj)).To(Succeed())
						Expect(cache.RefreshCallCount()).To(Equal(1))
						Expect(cache.RefreshArgsForCall(0)).To(Equal(stampedObj))
					})
				})

				Context("and the cache determines there has been a change since the last update", func() {
					BeforeEach(func() {
						cache.UnchangedSinceCachedReturns(false)
					})

					It("patches the object", func() {
						Expect(repo.CreateOrPatchUnstructuredObject(stampedObj)).To(Succeed())
						Expect(cl.PatchCallCount()).To(Equal(1))
					})

					Context("and the patch succeeds", func() {
						var returnedPatchedObj *unstructured.Unstructured

						BeforeEach(func() {
							returnedPatchedObj = stampedObj.DeepCopy()
							Expect(alterFieldOfNestedStringMaps(returnedPatchedObj.Object, "spec.template.spec.restartPolicy", "Never")).To(Succeed())
							cl.PatchStub = func(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
								objVal := reflect.ValueOf(obj)
								returnVal := reflect.ValueOf(returnedPatchedObj)

								reflect.Indirect(objVal).Set(reflect.Indirect(returnVal))
								return nil
							}
						})

						It("does not return an error", func() {
							Expect(repo.CreateOrPatchUnstructuredObject(stampedObj)).To(Succeed())
						})

						It("caches the submitted and persisted objects, as the persisted one may be modified by mutating webhooks", func() {
							originalStampedObj := stampedObj.DeepCopy()

							Expect(repo.CreateOrPatchUnstructuredObject(stampedObj)).To(Succeed())
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
							err := repo.CreateOrPatchUnstructuredObject(stampedObj)
							Expect(err).To(MatchError(ContainSubstring("patch: some-error")))
						})

						It("does not write to the submitted or persisted cache", func() {
							_ = repo.CreateOrPatchUnstructuredObject(stampedObj)
							Expect(cache.SetCallCount()).To(Equal(0))
						})
					})
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
				templateRef := v1alpha1.TemplateReference{
					Kind: "ClusterSourceTemplate",
					Name: "some-name",
				}
				template, err := repo.GetTemplate(templateRef)
				Expect(err).ToNot(HaveOccurred())
				Expect(template.GetName()).To(Equal("some-name"))
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
