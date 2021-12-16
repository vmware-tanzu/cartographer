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
		fakeLogger           *repositoryfakes.FakeLogger
		submitted, persisted *unstructured.Unstructured
	)

	BeforeEach(func() {
		fakeLogger = &repositoryfakes.FakeLogger{}
		cache = repository.NewCache(fakeLogger)

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

	Describe("UnchangedSinceCachedFromList", func() {
		Context("when the submitted object has a name", func() {
			var existingObjsOnAPIServer []*unstructured.Unstructured

			BeforeEach(func() {
				existingObjsOnAPIServer = append(existingObjsOnAPIServer, persisted.DeepCopy())
			})

			Context("when the submitted object is not present in the cache", func() {
				It("is false", func() {
					Expect(cache.UnchangedSinceCachedFromList(submitted, existingObjsOnAPIServer)).To(BeNil())
				})
			})

			Context("when the submitted object differs from the cached submitted object", func() {
				BeforeEach(func() {
					cache.Set(submitted, persisted)
				})

				It("is false", func() {
					newSubmission := submitted.DeepCopy()
					newSubmission.SetLabels(map[string]string{"now-with": "funky-labels"})
					Expect(cache.UnchangedSinceCachedFromList(newSubmission, existingObjsOnAPIServer)).To(BeNil())
				})
			})

			Context("when the submitted object is the same as the cached submitted object", func() {
				BeforeEach(func() {
					cache.Set(submitted, persisted)
				})

				Context("when the existing object has no spec", func() {
					It("is false", func() {
						Expect(cache.UnchangedSinceCachedFromList(submitted, existingObjsOnAPIServer)).To(BeNil())
					})
				})

				Context("when the existing object has a spec", func() {
					BeforeEach(func() {
						existingObjsOnAPIServer[0].UnstructuredContent()["spec"] = map[string]interface{}{"oh-look": "its-a-spec"}
					})

					Context("when the persisted object has no spec", func() {
						BeforeEach(func() {
							cache.Set(submitted, persisted)
						})

						It("is false", func() {
							Expect(cache.UnchangedSinceCachedFromList(submitted, existingObjsOnAPIServer)).To(BeNil())
						})
					})

					Context("when the persisted object has a spec", func() {
						Context("when the existing object spec is the same as the cached submitted object spec", func() {
							BeforeEach(func() {
								persisted.UnstructuredContent()["spec"] = existingObjsOnAPIServer[0].UnstructuredContent()["spec"]
								cache.Set(submitted, persisted)
							})

							It("is true", func() {
								Expect(cache.UnchangedSinceCachedFromList(submitted, existingObjsOnAPIServer)).ToNot(BeNil())
							})
						})

						Context("when the existing object spec differs from the cached submitted object spec", func() {
							BeforeEach(func() {
								persisted.UnstructuredContent()["spec"] = map[string]interface{}{"oh-wait": "this-spec-is-different"}
								cache.Set(submitted, persisted)
							})

							It("is false", func() {
								Expect(cache.UnchangedSinceCachedFromList(submitted, existingObjsOnAPIServer)).To(BeNil())
							})
						})
					})
				})
			})
		})

		Context("when the submitted object has no name", func() {
			var existingObjsOnAPIServer []*unstructured.Unstructured

			BeforeEach(func() {
				submitted.SetName("")
				submitted.SetGenerateName("this-is-generate-name-")
				submitted.UnstructuredContent()["spec"] = map[string]interface{}{"ooo": "a-spec"}

				persisted.SetName("this-is-generate-name-abcdef")
				persisted.SetGenerateName("")
				persisted.UnstructuredContent()["spec"] = submitted.UnstructuredContent()["spec"]

				cache.Set(submitted, persisted)
				existingObjsOnAPIServer = append(existingObjsOnAPIServer, persisted.DeepCopy())
			})

			It("the cache matches against the generateName instead", func() {
				Expect(cache.UnchangedSinceCachedFromList(submitted, existingObjsOnAPIServer)).ToNot(BeNil())
				submitted.SetGenerateName("another-generate-name-")
				Expect(cache.UnchangedSinceCachedFromList(submitted, existingObjsOnAPIServer)).To(BeNil())
			})
		})
	})
})
