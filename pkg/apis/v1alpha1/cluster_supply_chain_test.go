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
	"fmt"
	"reflect"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"

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

		It("requires a selector", func() {
			selectorField, found := supplyChainSpecType.FieldByName("Selector")
			Expect(found).To(BeTrue())
			jsonValue := selectorField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("selector"))
			Expect(jsonValue).NotTo(ContainSubstring("omitempty"))
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

	Describe("Webhook Validation", func() {
		Describe("#Create", func() {
			Context("Well formed supply chain", func() {
				var wellFormedSupplyChain *v1alpha1.ClusterSupplyChain
				BeforeEach(func() {
					wellFormedSupplyChain = &v1alpha1.ClusterSupplyChain{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "responsible-ops---default-params",
							Namespace: "default",
						},
						Spec: v1alpha1.SupplyChainSpec{
							Resources: []v1alpha1.SupplyChainResource{
								{
									Name: "source-provider",
									TemplateRef: v1alpha1.ClusterTemplateReference{
										Kind: "ClusterSourceTemplate",
										Name: "git-template---default-params",
									},
								},
								{
									Name: "other-source-provider",
									TemplateRef: v1alpha1.ClusterTemplateReference{
										Kind: "ClusterSourceTemplate",
										Name: "git-template---default-params",
									},
								},
							},
							Selector: map[string]string{"integration-test": "workload-no-supply-chain"},
						},
					}
				})
				It("does not return an error", func() {
					Expect(wellFormedSupplyChain.ValidateCreate()).NotTo(HaveOccurred())
				})
			})

			Context("Supply chain with a resource reference that does not exist", func() {
				var supplyChain *v1alpha1.ClusterSupplyChain

				BeforeEach(func() {
					supplyChain = &v1alpha1.ClusterSupplyChain{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "responsible-ops---default-params",
							Namespace: "default",
						},
						Spec: v1alpha1.SupplyChainSpec{
							Resources: []v1alpha1.SupplyChainResource{
								{
									Name: "some-resource",
									TemplateRef: v1alpha1.ClusterTemplateReference{
										Kind: "ClusterSourceTemplate",
										Name: "some-template",
									},
								},
								{
									Name: "other-resource",
									TemplateRef: v1alpha1.ClusterTemplateReference{
										Kind: "ClusterTemplate",
										Name: "some-other-template",
									},
									Sources: []v1alpha1.ResourceReference{
										{
											Name:     "some-source",
											Resource: "some-nonexistent-resource",
										},
									},
								},
							},
						},
					}
				})

				It("returns an error", func() {
					Expect(supplyChain.ValidateCreate()).To(MatchError(
						"invalid sources for resource 'other-resource': 'some-source' is provided by unknown resource 'some-nonexistent-resource'",
					))
				})
			})

			Context("Two resources with the same name", func() {
				var supplyChainWithDuplicateResourceNames *v1alpha1.ClusterSupplyChain
				BeforeEach(func() {
					supplyChainWithDuplicateResourceNames = &v1alpha1.ClusterSupplyChain{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "responsible-ops---default-params",
							Namespace: "default",
						},
						Spec: v1alpha1.SupplyChainSpec{
							Resources: []v1alpha1.SupplyChainResource{
								{
									Name: "source-provider",
									TemplateRef: v1alpha1.ClusterTemplateReference{
										Kind: "ClusterSourceTemplate",
										Name: "git-template---default-params",
									},
								},
								{
									Name: "source-provider",
									TemplateRef: v1alpha1.ClusterTemplateReference{
										Kind: "ClusterSourceTemplate",
										Name: "git-template---default-params",
									},
								},
							},
							Selector: map[string]string{"integration-test": "workload-no-supply-chain"},
						},
					}
				})

				It("rejects the Resource", func() {
					err := supplyChainWithDuplicateResourceNames.ValidateCreate()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("duplicate resource name 'source-provider' found in clustersupplychain 'responsible-ops---default-params'"))
				})
			})

			Describe("Template inputs must reference a resource with a matching type", func() {
				var supplyChain *v1alpha1.ClusterSupplyChain
				var consumerToProviderMapping = map[string]string{
					"Source": "ClusterSourceTemplate",
					"Config": "ClusterConfigTemplate",
					"Image":  "ClusterImageTemplate",
				}
				BeforeEach(func() {
					supplyChain = &v1alpha1.ClusterSupplyChain{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "responsible-ops---default-params",
							Namespace: "default",
						},
						Spec: v1alpha1.SupplyChainSpec{
							Resources: []v1alpha1.SupplyChainResource{
								{
									Name: "input-provider",
									TemplateRef: v1alpha1.ClusterTemplateReference{
										Name: "output-template",
									},
								},
								{
									Name: "input-consumer",
									TemplateRef: v1alpha1.ClusterTemplateReference{
										Kind: "ClusterTemplate",
										Name: "consuming-template",
									},
								},
							},
							Selector: map[string]string{"integration-test": "workload-no-supply-chain"},
						},
					}
				})
				DescribeTable("template input does not match template type",
					func(firstResourceKind string, inputReferenceType string, happy bool) {
						supplyChain.Spec.Resources[0].TemplateRef.Kind = firstResourceKind

						reference := v1alpha1.ResourceReference{
							Name:     "input-name",
							Resource: "input-provider",
						}

						switch inputReferenceType {
						case "Source":
							supplyChain.Spec.Resources[1].Sources = []v1alpha1.ResourceReference{reference}
						case "Image":
							supplyChain.Spec.Resources[1].Images = []v1alpha1.ResourceReference{reference}
						case "Config":
							supplyChain.Spec.Resources[1].Configs = []v1alpha1.ResourceReference{reference}
						}

						err := supplyChain.ValidateCreate()

						if happy {
							Expect(err).NotTo(HaveOccurred())
						} else {
							Expect(err).To(HaveOccurred())
							Expect(err).To(MatchError(fmt.Sprintf(
								"invalid %ss for resource 'input-consumer': resource 'input-provider' providing 'input-name' must reference a %s",
								strings.ToLower(inputReferenceType),
								consumerToProviderMapping[inputReferenceType]),
							))
						}

					},
					Entry("Config cannot be a source provider", "ClusterTemplate", "Source", false),
					Entry("Config cannot be a image provider", "ClusterTemplate", "Image", false),
					Entry("Config cannot be an config provider", "ClusterTemplate", "Config", false),
					Entry("Build cannot be a source provider", "ClusterImageTemplate", "Source", false),
					Entry("Build can be a image provider", "ClusterImageTemplate", "Image", true),
					Entry("Build cannot be a config provider", "ClusterImageTemplate", "Config", false),
					Entry("Source can be a source provider", "ClusterSourceTemplate", "Source", true),
					Entry("Source cannot be a image provider", "ClusterSourceTemplate", "Image", false),
					Entry("Source cannot be a config provider", "ClusterSourceTemplate", "Config", false),
					Entry("Config cannot be a source provider", "ClusterConfigTemplate", "Source", false),
					Entry("Config cannot be a image provider", "ClusterConfigTemplate", "Image", false),
					Entry("Config can be a config provider", "ClusterConfigTemplate", "Config", true),
				)
			})
		})

		Describe("#Update", func() {
			var oldSupplyChain *v1alpha1.ClusterSupplyChain
			Context("Well formed supply chain", func() {
				var wellFormedSupplyChain *v1alpha1.ClusterSupplyChain
				BeforeEach(func() {
					wellFormedSupplyChain = &v1alpha1.ClusterSupplyChain{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "responsible-ops---default-params",
							Namespace: "default",
						},
						Spec: v1alpha1.SupplyChainSpec{
							Resources: []v1alpha1.SupplyChainResource{
								{
									Name: "source-provider",
									TemplateRef: v1alpha1.ClusterTemplateReference{
										Kind: "ClusterSourceTemplate",
										Name: "git-template---default-params",
									},
								},
								{
									Name: "other-source-provider",
									TemplateRef: v1alpha1.ClusterTemplateReference{
										Kind: "ClusterSourceTemplate",
										Name: "git-template---default-params",
									},
								},
							},
							Selector: map[string]string{"integration-test": "workload-no-supply-chain"},
						},
					}
				})
				It("does not return an error", func() {
					Expect(wellFormedSupplyChain.ValidateUpdate(oldSupplyChain)).NotTo(HaveOccurred())
				})
			})

			Context("Two resources with the same name", func() {
				var supplyChainWithDuplicateResourceNames *v1alpha1.ClusterSupplyChain
				BeforeEach(func() {
					supplyChainWithDuplicateResourceNames = &v1alpha1.ClusterSupplyChain{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "responsible-ops---default-params",
							Namespace: "default",
						},
						Spec: v1alpha1.SupplyChainSpec{
							Resources: []v1alpha1.SupplyChainResource{
								{
									Name: "source-provider",
									TemplateRef: v1alpha1.ClusterTemplateReference{
										Kind: "ClusterSourceTemplate",
										Name: "git-template---default-params",
									},
								},
								{
									Name: "source-provider",
									TemplateRef: v1alpha1.ClusterTemplateReference{
										Kind: "ClusterSourceTemplate",
										Name: "git-template---default-params",
									},
								},
							},
							Selector: map[string]string{"integration-test": "workload-no-supply-chain"},
						},
					}
				})

				It("rejects the Resource", func() {
					err := supplyChainWithDuplicateResourceNames.ValidateUpdate(oldSupplyChain)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("duplicate resource name 'source-provider' found in clustersupplychain 'responsible-ops---default-params'"))
				})
			})

			Context("Template inputs do not reference a resource with a matching type", func() {
				var invalidSupplyChain *v1alpha1.ClusterSupplyChain
				BeforeEach(func() {
					invalidSupplyChain = &v1alpha1.ClusterSupplyChain{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "responsible-ops---default-params",
							Namespace: "default",
						},
						Spec: v1alpha1.SupplyChainSpec{
							Resources: []v1alpha1.SupplyChainResource{
								{
									Name: "image-provider",
									TemplateRef: v1alpha1.ClusterTemplateReference{
										Name: "image-output-template",
										Kind: "ClusterImageTemplate",
									},
								},
								{
									Name: "input-consumer",
									TemplateRef: v1alpha1.ClusterTemplateReference{
										Kind: "ClusterTemplate",
										Name: "consuming-template",
									},
									Sources: []v1alpha1.ResourceReference{
										{
											Name:     "source-name",
											Resource: "image-provider",
										},
									},
								},
							},
							Selector: map[string]string{"integration-test": "workload-no-supply-chain"},
						},
					}
				})
				It("validates on update as well", func() {
					err := invalidSupplyChain.ValidateUpdate(&v1alpha1.ClusterSupplyChain{})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("invalid sources for resource 'input-consumer': resource 'image-provider' providing 'source-name' must reference a ClusterSourceTemplate"))
				})
			})
		})

		Describe("#Delete", func() {
			Context("Any supply chain", func() {
				var anyOldSupplyChain *v1alpha1.ClusterSupplyChain
				It("does not return an error", func() {
					Expect(anyOldSupplyChain.ValidateDelete()).NotTo(HaveOccurred())
				})
			})
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

	Describe("ClusterTemplateReference", func() {
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
			// an extension of go parser for loading markers.
			// Loads the package but only with the contents of the cluster_supply_chain file
			packages, err := loader.LoadRoots("./cluster_supply_chain.go")
			Expect(err).NotTo(HaveOccurred())
			Expect(packages).To(HaveLen(1))
			Expect(packages[0].GoFiles).To(HaveLen(1))

			// create a registry of CRD markers
			reg := &markers.Registry{}
			err = crdmarkers.Register(reg)
			Expect(err).NotTo(HaveOccurred())

			// and a collector which `EachType` requires
			coll := &markers.Collector{Registry: reg}

			// find the "kubebuilder:validation:Enum" marker for the ClusterTemplateReference.Kind field
			var enumMarkers crdmarkers.Enum
			err = markers.EachType(coll, packages[0], func(info *markers.TypeInfo) {
				if info.Name == "ClusterTemplateReference" {
					for _, fieldInfo := range info.Fields {
						if fieldInfo.Name == "Kind" {
							var ok bool
							enumMarkers, ok = fieldInfo.Markers.Get("kubebuilder:validation:Enum").(crdmarkers.Enum)
							Expect(ok).To(BeTrue())
						}
					}
				}
			})
			Expect(err).NotTo(HaveOccurred())

			// There should be 4, all found in the ValidSupplyChainTemplates List
			Expect(enumMarkers).To(HaveLen(len(v1alpha1.ValidSupplyChainTemplates)))
			for _, validTemplate := range v1alpha1.ValidSupplyChainTemplates {
				typ := reflect.TypeOf(validTemplate)
				templateName := typ.Elem().Name()
				Expect(enumMarkers).To(ContainElement(templateName))
			}
		})
	})
})
