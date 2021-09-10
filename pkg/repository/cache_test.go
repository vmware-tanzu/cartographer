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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
)

var _ = Describe("Cache", func() {
	var (
		cache                repository.RepoCache
		fakeExpiringCache    *repositoryfakes.FakeExpiringCache
		submitted, persisted *unstructured.Unstructured
	)

	BeforeEach(func() {
		fakeExpiringCache = &repositoryfakes.FakeExpiringCache{}
		cache = repository.NewCache(fakeExpiringCache)

		objKind := "the-kind"
		objName := "its-name"
		objNamespace := "its-ns"
		submitted = &unstructured.Unstructured{}
		submitted.SetKind(objKind)
		submitted.SetName(objName)
		submitted.SetNamespace(objNamespace)

		persisted = &unstructured.Unstructured{}
		persisted.SetLabels(map[string]string{"something": "different-here"})
		persisted.SetKind(objKind + "-ignored-submitted-one-is-used")
		persisted.SetName(objName + "-ignored-submitted-one-is-used")
		persisted.SetNamespace(objNamespace + "-ignored-submitted-one-is-used")
	})

	Describe("Set", func() {
		It("stores the submitted and persisted values paired in the cache", func() {
			cache.Set(submitted, persisted)

			Expect(fakeExpiringCache.SetCallCount()).To(Equal(2))

			submittedKey, submittedValue, submittedExpiryDuration := fakeExpiringCache.SetArgsForCall(0)
			Expect(submittedKey).To(Equal("submitted:its-ns:the-kind:its-name"))
			Expect(submittedValue).To(Equal(*submitted))
			Expect(submittedExpiryDuration).To(Equal(repository.CacheExpiryDuration))
			persistedKey, persistedValue, persistedExpiryDuration := fakeExpiringCache.SetArgsForCall(1)
			Expect(persistedKey).To(Equal("persisted:its-ns:the-kind:its-name"))
			Expect(persistedValue).To(Equal(*persisted))
			Expect(persistedExpiryDuration).To(Equal(repository.CacheExpiryDuration))
		})

		Context("constructs a key using generateName", func() {
			BeforeEach(func() {
				submitted.SetName("")
				submitted.SetGenerateName("some-generated-")
			})

			It("stores the submitted and persisted values paired in the cache", func() {
				cache.Set(submitted, persisted)

				Expect(fakeExpiringCache.SetCallCount()).To(Equal(2))

				submittedKey, _, _ := fakeExpiringCache.SetArgsForCall(0)
				Expect(submittedKey).To(Equal("submitted:its-ns:the-kind:some-generated-"))
				persistedKey, _, _ := fakeExpiringCache.SetArgsForCall(1)
				Expect(persistedKey).To(Equal("persisted:its-ns:the-kind:some-generated-"))
			})
		})
	})

	Describe("functions that rely on the state of the cache", func() {
		var (
			submittedObjInCache, persistedObjInCache     interface{}
			submittedFoundInCache, persistedFoundInCache bool
		)

		BeforeEach(func() {
			fakeExpiringCache.GetCalls(func(key interface{}) (interface{}, bool) {
				if key == "submitted:its-ns:the-kind:its-name" {
					return submittedObjInCache, submittedFoundInCache
				}
				if key == "persisted:its-ns:the-kind:its-name" {
					return persistedObjInCache, persistedFoundInCache
				}
				panic("unexpected key")
			})
		})

		Describe("Refresh", func() {
			Context("when submitted and persisted exist in the cache", func() {
				BeforeEach(func() {
					persistedObjInCache = *persisted
					persistedFoundInCache = true
					submittedObjInCache = *submitted
					submittedFoundInCache = true
				})

				It("re-sets the entries in the cache with the expiry duration", func() {
					cache.Refresh(submitted)

					submittedKey, submittedValue, submittedExpiryDuration := fakeExpiringCache.SetArgsForCall(0)
					Expect(submittedKey).To(Equal("submitted:its-ns:the-kind:its-name"))
					Expect(submittedValue).To(Equal(*submitted))
					Expect(submittedExpiryDuration).To(Equal(repository.CacheExpiryDuration))
					persistedKey, persistedValue, persistedExpiryDuration := fakeExpiringCache.SetArgsForCall(1)
					Expect(persistedKey).To(Equal("persisted:its-ns:the-kind:its-name"))
					Expect(persistedValue).To(Equal(*persisted))
					Expect(persistedExpiryDuration).To(Equal(repository.CacheExpiryDuration))
				})
			})

			Context("when the submitted obj is not in the cache", func() {
				BeforeEach(func() {
					persistedObjInCache = *persisted
					persistedFoundInCache = true
					submittedObjInCache = nil
					submittedFoundInCache = false
				})

				It("makes no changes to the cache", func() {
					cache.Refresh(submitted)
					Expect(fakeExpiringCache.SetCallCount()).To(Equal(0))
				})
			})

			Context("when the submitted obj is not in the cache", func() {
				BeforeEach(func() {
					persistedObjInCache = nil
					persistedFoundInCache = false
					submittedObjInCache = *submitted
					submittedFoundInCache = true
				})

				It("makes no changes to the cache", func() {
					cache.Refresh(submitted)
					Expect(fakeExpiringCache.SetCallCount()).To(Equal(0))
				})
			})
		})

		Describe("UnchangedSinceCached", func() {
			var existingObjsOnAPIServer []unstructured.Unstructured

			BeforeEach(func() {
				persistedObjInCache = *persisted
				persistedFoundInCache = true
				submittedObjInCache = *submitted
				submittedFoundInCache = true

				existingObjsOnAPIServer = append(existingObjsOnAPIServer, *persisted.DeepCopy())
			})

			Context("when the submitted object is not present in the cache", func() {
				BeforeEach(func() {
					submittedObjInCache = nil
					submittedFoundInCache = false
				})

				It("is false", func() {
					Expect(cache.UnchangedSinceCached(submitted, existingObjsOnAPIServer)).To(BeNil())
				})
			})

			Context("when the submitted object differs from the cached submitted object", func() {
				It("is false", func() {
					newSubmission := submitted.DeepCopy()
					newSubmission.SetLabels(map[string]string{"now-with": "funky-labels"})
					Expect(cache.UnchangedSinceCached(newSubmission, existingObjsOnAPIServer)).To(BeNil())
				})
			})

			Context("when the submitted object is the same as the cached submitted object", func() {
				Context("when the existing object has no spec", func() {
					It("is false", func() {
						Expect(cache.UnchangedSinceCached(submitted, existingObjsOnAPIServer)).To(BeNil())
					})
				})

				Context("when the existing object has a spec", func() {
					BeforeEach(func() {
						existingObjsOnAPIServer[0].UnstructuredContent()["spec"] = map[string]interface{}{"oh-look": "its-a-spec"}
					})

					Context("when the persisted object has no spec", func() {
						It("is false", func() {
							Expect(cache.UnchangedSinceCached(submitted, existingObjsOnAPIServer)).To(BeNil())
						})
					})

					Context("when the persisted object is not present", func() {
						BeforeEach(func() {
							persistedObjInCache = nil
							persistedFoundInCache = false
						})

						It("is false", func() {
							Expect(cache.UnchangedSinceCached(submitted, existingObjsOnAPIServer)).To(BeNil())
						})
					})

					Context("when the persisted object is somehow not an unstructured object", func() {
						BeforeEach(func() {
							persistedObjInCache = "this is a string, not an unstructured"
						})

						It("is false", func() {
							Expect(cache.UnchangedSinceCached(submitted, existingObjsOnAPIServer)).To(BeNil())
						})
					})

					Context("when the persisted object has a spec", func() {
						Context("when the existing object spec is the same as the cached submitted object spec", func() {
							BeforeEach(func() {
								persisted.UnstructuredContent()["spec"] = existingObjsOnAPIServer[0].UnstructuredContent()["spec"]
							})

							It("is true", func() {
								Expect(cache.UnchangedSinceCached(submitted, existingObjsOnAPIServer)).ToNot(BeNil())
							})
						})

						Context("when the existing object spec differs from the cached submitted object spec", func() {
							BeforeEach(func() {
								persisted.UnstructuredContent()["spec"] = map[string]interface{}{"oh-wait": "this-spec-is-different"}
							})

							It("is false", func() {
								Expect(cache.UnchangedSinceCached(submitted, existingObjsOnAPIServer)).To(BeNil())
							})
						})
					})
				})
			})
		})
	})
})
