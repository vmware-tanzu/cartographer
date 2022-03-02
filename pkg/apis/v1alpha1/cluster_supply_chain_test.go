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

package v1alpha1_test

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

var _ = Describe("ClusterSupplyChain", func() {
	Describe("SupplyChainSpec", func() {
		var (
			supplyChainSpec     v1alpha1.SupplyChainSpec
			supplyChainSpecType reflect.Type
		)

		BeforeEach(func() {
			supplyChainSpecType = reflect.TypeOf(supplyChainSpec)
		})

		It("requires resources", func() {
			resourcesField, found := supplyChainSpecType.FieldByName("Resources")
			Expect(found).To(BeTrue())
			jsonValue := resourcesField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("resources"))
			Expect(jsonValue).NotTo(ContainSubstring("omitempty"))
		})

		It("allows but does not require a service account ref", func() {
			serviceAccountNameField, found := supplyChainSpecType.FieldByName("ServiceAccountRef")
			Expect(found).To(BeTrue())
			jsonValue := serviceAccountNameField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("serviceAccountRef"))
			Expect(jsonValue).To(ContainSubstring("omitempty"))
		})
	})

	Describe("SupplyChainResource", func() {
		var (
			supplyChainResource     v1alpha1.SupplyChainResource
			supplyChainResourceType reflect.Type
		)

		BeforeEach(func() {
			supplyChainResourceType = reflect.TypeOf(supplyChainResource)
		})

		It("requires a name", func() {
			nameField, found := supplyChainResourceType.FieldByName("Name")
			Expect(found).To(BeTrue())
			jsonValue := nameField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("name"))
			Expect(jsonValue).NotTo(ContainSubstring("omitempty"))
		})

		It("requires a templateRef", func() {
			templateRefField, found := supplyChainResourceType.FieldByName("TemplateRef")
			Expect(found).To(BeTrue())
			jsonValue := templateRefField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("templateRef"))
			Expect(jsonValue).NotTo(ContainSubstring("omitempty"))
		})

		It("allows but does not require param", func() {
			paramField, found := supplyChainResourceType.FieldByName("Params")
			Expect(found).To(BeTrue())
			jsonValue := paramField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("params"))
			Expect(jsonValue).To(ContainSubstring("omitempty"))
		})

		It("does not require sources", func() {
			sourcesField, found := supplyChainResourceType.FieldByName("Sources")
			Expect(found).To(BeTrue())
			jsonValue := sourcesField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("sources"))
			Expect(jsonValue).To(ContainSubstring("omitempty"))
		})

		It("does not require images", func() {
			imagesField, found := supplyChainResourceType.FieldByName("Images")
			Expect(found).To(BeTrue())
			jsonValue := imagesField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("images"))
			Expect(jsonValue).To(ContainSubstring("omitempty"))
		})

		It("does not require configs", func() {
			configsField, found := supplyChainResourceType.FieldByName("Configs")
			Expect(found).To(BeTrue())
			jsonValue := configsField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("configs"))
			Expect(jsonValue).To(ContainSubstring("omitempty"))
		})
	})
	Describe("GetSelectorsFromObject", func() {
		var expectedSelectors, actualSelectors []string
		Context("when object is a supply chain", func() {
			var sc *v1alpha1.ClusterSupplyChain

			BeforeEach(func() {
				sc = &v1alpha1.ClusterSupplyChain{Spec: v1alpha1.SupplyChainSpec{}}
			})
			Context("given a supply chain with 0 selectors", func() {
				BeforeEach(func() {
					actualSelectors = v1alpha1.GetSelectorsFromObject(sc)
				})
				It("returns an empty list", func() {
					expectedSelectors = []string{}
					Expect(actualSelectors).To(ConsistOf(expectedSelectors))
				})
			})
			Context("given a supply chain with 1 selector", func() {
				BeforeEach(func() {
					sc.Spec.Selector = map[string]string{"some-key": "some-val"}
					actualSelectors = v1alpha1.GetSelectorsFromObject(sc)
				})

				It("returns a list with only the string concatenation of the key-value", func() {
					expectedSelectors = []string{"some-key: some-val"}
					Expect(actualSelectors).To(ConsistOf(expectedSelectors))
				})
			})
			Context("given a supply chain with many selectors", func() {
				BeforeEach(func() {
					sc.Spec.Selector = map[string]string{
						"some-key":    "some-val",
						"another-key": "another-val",
					}
					actualSelectors = v1alpha1.GetSelectorsFromObject(sc)
				})

				It("returns a list with string concatenations of each key-value", func() {
					expectedSelectors = []string{"some-key: some-val", "another-key: another-val"}
					Expect(actualSelectors).To(ConsistOf(expectedSelectors))
				})
			})
		})

		Context("when object is not a supply chain", func() {
			BeforeEach(func() {
				actualSelectors = v1alpha1.GetSelectorsFromObject(&v1alpha1.Workload{})
			})
			It("returns an empty list", func() {
				expectedSelectors = []string{}
				Expect(actualSelectors).To(ConsistOf(expectedSelectors))
			})
		})
	})

	Describe("SupplyChainTemplateReference", func() {
		It("has four valid references", func() {
			Expect(v1alpha1.ValidSupplyChainTemplates).To(HaveLen(4))

			Expect(v1alpha1.ValidSupplyChainTemplates).To(ContainElements(
				&v1alpha1.ClusterSourceTemplate{},
				&v1alpha1.ClusterConfigTemplate{},
				&v1alpha1.ClusterImageTemplate{},
				&v1alpha1.ClusterTemplate{},
			))
		})

		It("has a matching valid enum for Kind", func() {
			mrkrs, err := markersFor(
				"./cluster_supply_chain.go",
				"SupplyChainTemplateReference",
				"Kind",
				"kubebuilder:validation:Enum",
			)

			Expect(err).NotTo(HaveOccurred())

			enumMarkers, ok := mrkrs.(crdmarkers.Enum)
			Expect(ok).To(BeTrue())

			Expect(enumMarkers).To(HaveLen(len(v1alpha1.ValidSupplyChainTemplates)))
			for _, validTemplate := range v1alpha1.ValidSupplyChainTemplates {
				typ := reflect.TypeOf(validTemplate)
				templateName := typ.Elem().Name()
				Expect(enumMarkers).To(ContainElement(templateName))
			}
		})
	})
})
