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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

		It("requires a selector", func() {
			selectorField, found := supplyChainSpecType.FieldByName("Selector")
			Expect(found).To(BeTrue())
			jsonValue := selectorField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("selector"))
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

	Describe("Webhook Validation", func() {
		Context("supply chain without options", func() {
			var (
				supplyChain    *v1alpha1.ClusterSupplyChain
				oldSupplyChain *v1alpha1.ClusterSupplyChain
			)

			BeforeEach(func() {
				supplyChain = &v1alpha1.ClusterSupplyChain{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "responsible-ops---default-params",
						Namespace: "default",
					},
					Spec: v1alpha1.SupplyChainSpec{
						Resources: []v1alpha1.SupplyChainResource{
							{
								Name: "source-provider",
								TemplateRef: v1alpha1.SupplyChainTemplateReference{
									Kind: "ClusterSourceTemplate",
									Name: "git-template---default-params",
								},
							},
							{
								Name: "other-source-provider",
								TemplateRef: v1alpha1.SupplyChainTemplateReference{
									Kind: "ClusterSourceTemplate",
									Name: "git-template---default-params",
								},
							},
						},
						Selector: map[string]string{"integration-test": "workload-no-supply-chain"},
						Params: []v1alpha1.BlueprintParam{
							{
								Name:  "some-param",
								Value: &apiextensionsv1.JSON{Raw: []byte(`"some value"`)},
							},
						},
					},
				}
			})

			Context("Well formed supply chain", func() {
				It("creates without error", func() {
					Expect(supplyChain.ValidateCreate()).NotTo(HaveOccurred())
				})

				It("updates without error", func() {
					Expect(supplyChain.ValidateUpdate(oldSupplyChain)).NotTo(HaveOccurred())
				})

				It("deletes without error", func() {
					Expect(supplyChain.ValidateDelete()).NotTo(HaveOccurred())
				})
			})

			Context("Supply chain with a resource reference that does not exist", func() {
				BeforeEach(func() {
					supplyChain.Spec.Resources[1].Sources = []v1alpha1.ResourceReference{
						{
							Name:     "some-source",
							Resource: "some-nonexistent-resource",
						},
					}
				})

				It("on create, returns an error", func() {
					Expect(supplyChain.ValidateCreate()).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: invalid sources for resource [other-source-provider]: [some-source] is provided by unknown resource [some-nonexistent-resource]",
					))
				})

				It("on update, returns an error", func() {
					Expect(supplyChain.ValidateUpdate(oldSupplyChain)).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: invalid sources for resource [other-source-provider]: [some-source] is provided by unknown resource [some-nonexistent-resource]",
					))
				})

				It("deletes without error", func() {
					Expect(supplyChain.ValidateDelete()).NotTo(HaveOccurred())
				})
			})

			Context("Two resources with the same name", func() {
				BeforeEach(func() {
					for i := range supplyChain.Spec.Resources {
						supplyChain.Spec.Resources[i].Name = "some-duplicate-name"
					}
				})

				It("on create, it rejects the Resource", func() {
					Expect(supplyChain.ValidateCreate()).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: duplicate resource name [some-duplicate-name] found",
					))
				})

				It("on update, it rejects the Resource", func() {
					Expect(supplyChain.ValidateUpdate(oldSupplyChain)).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: duplicate resource name [some-duplicate-name] found",
					))
				})

				It("deletes without error", func() {
					Expect(supplyChain.ValidateDelete()).NotTo(HaveOccurred())
				})
			})

			Context("SupplyChain with malformed params", func() {
				Context("Top level params are malformed", func() {
					Context("param does not specify a value or default", func() {
						BeforeEach(func() {
							supplyChain.Spec.Params = []v1alpha1.BlueprintParam{
								{
									Name: "some-param",
								},
							}
						})
						It("on create, returns an error", func() {
							Expect(supplyChain.ValidateCreate()).To(MatchError(
								"error validating clustersupplychain [responsible-ops---default-params]: param [some-param] is invalid: must set exactly one of value and default",
							))
						})

						It("on update, returns an error", func() {
							Expect(supplyChain.ValidateUpdate(oldSupplyChain)).To(MatchError(
								"error validating clustersupplychain [responsible-ops---default-params]: param [some-param] is invalid: must set exactly one of value and default",
							))
						})

						It("deletes without error", func() {
							Expect(supplyChain.ValidateDelete()).NotTo(HaveOccurred())
						})
					})

					Context("param specifies both a value and a default", func() {
						BeforeEach(func() {
							supplyChain.Spec.Params = []v1alpha1.BlueprintParam{
								{
									Name:         "some-param",
									Value:        &apiextensionsv1.JSON{Raw: []byte(`"some value"`)},
									DefaultValue: &apiextensionsv1.JSON{Raw: []byte(`"some value"`)},
								},
							}
						})

						It("on create, returns an error", func() {
							Expect(supplyChain.ValidateCreate()).To(MatchError(
								"error validating clustersupplychain [responsible-ops---default-params]: param [some-param] is invalid: must set exactly one of value and default",
							))
						})

						It("on update, returns an error", func() {
							Expect(supplyChain.ValidateUpdate(oldSupplyChain)).To(MatchError(
								"error validating clustersupplychain [responsible-ops---default-params]: param [some-param] is invalid: must set exactly one of value and default",
							))
						})

						It("deletes without error", func() {
							Expect(supplyChain.ValidateDelete()).NotTo(HaveOccurred())
						})
					})
				})

				Context("Params of an individual resource are malformed", func() {
					Context("param does not specify a value or default", func() {
						BeforeEach(func() {
							supplyChain.Spec.Resources[0].Params = []v1alpha1.BlueprintParam{
								{
									Name: "some-param",
								},
							}
						})
						It("on create, returns an error", func() {
							Expect(supplyChain.ValidateCreate()).To(MatchError(
								"error validating clustersupplychain [responsible-ops---default-params]: resource [source-provider] is invalid: param [some-param] is invalid: must set exactly one of value and default",
							))
						})

						It("on update, returns an error", func() {
							Expect(supplyChain.ValidateUpdate(oldSupplyChain)).To(MatchError(
								"error validating clustersupplychain [responsible-ops---default-params]: resource [source-provider] is invalid: param [some-param] is invalid: must set exactly one of value and default",
							))
						})

						It("deletes without error", func() {
							Expect(supplyChain.ValidateDelete()).NotTo(HaveOccurred())
						})
					})

					Context("param specifies both a value and a default", func() {
						BeforeEach(func() {
							supplyChain.Spec.Resources[0].Params = []v1alpha1.BlueprintParam{
								{
									Name:         "some-param",
									Value:        &apiextensionsv1.JSON{Raw: []byte(`"some value"`)},
									DefaultValue: &apiextensionsv1.JSON{Raw: []byte(`"some value"`)},
								},
							}
						})
						It("on create, returns an error", func() {
							Expect(supplyChain.ValidateCreate()).To(MatchError(
								"error validating clustersupplychain [responsible-ops---default-params]: resource [source-provider] is invalid: param [some-param] is invalid: must set exactly one of value and default",
							))
						})

						It("on update, returns an error", func() {
							Expect(supplyChain.ValidateUpdate(oldSupplyChain)).To(MatchError(
								"error validating clustersupplychain [responsible-ops---default-params]: resource [source-provider] is invalid: param [some-param] is invalid: must set exactly one of value and default",
							))
						})

						It("deletes without error", func() {
							Expect(supplyChain.ValidateDelete()).NotTo(HaveOccurred())
						})
					})
				})
			})

			Describe("Template inputs must reference a resource with a matching type", func() {
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
									TemplateRef: v1alpha1.SupplyChainTemplateReference{
										Name: "output-template",
									},
								},
								{
									Name: "input-consumer",
									TemplateRef: v1alpha1.SupplyChainTemplateReference{
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

						// Create
						createErr := supplyChain.ValidateCreate()

						// Update
						updateErr := supplyChain.ValidateUpdate(oldSupplyChain)

						// Delete
						deleteErr := supplyChain.ValidateDelete()

						if happy {
							Expect(createErr).NotTo(HaveOccurred())
							Expect(updateErr).NotTo(HaveOccurred())
							Expect(deleteErr).NotTo(HaveOccurred())
						} else {
							Expect(createErr).To(HaveOccurred())
							Expect(createErr).To(MatchError(fmt.Sprintf(
								"error validating clustersupplychain [responsible-ops---default-params]: invalid %ss for resource [input-consumer]: resource [input-provider] providing [input-name] must reference a %s",
								strings.ToLower(inputReferenceType),
								consumerToProviderMapping[inputReferenceType]),
							))

							Expect(updateErr).To(HaveOccurred())
							Expect(updateErr).To(MatchError(fmt.Sprintf(
								"error validating clustersupplychain [responsible-ops---default-params]: invalid %ss for resource [input-consumer]: resource [input-provider] providing [input-name] must reference a %s",
								strings.ToLower(inputReferenceType),
								consumerToProviderMapping[inputReferenceType]),
							))

							Expect(deleteErr).NotTo(HaveOccurred())
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

		Context("supply chain with options", func() {
			var (
				supplyChain    *v1alpha1.ClusterSupplyChain
				oldSupplyChain *v1alpha1.ClusterSupplyChain
			)

			BeforeEach(func() {
				supplyChain = &v1alpha1.ClusterSupplyChain{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "responsible-ops---default-params",
						Namespace: "default",
					},
					Spec: v1alpha1.SupplyChainSpec{
						Resources: []v1alpha1.SupplyChainResource{
							{
								Name: "source-provider",
								TemplateRef: v1alpha1.SupplyChainTemplateReference{
									Kind: "ClusterSourceTemplate",
									Options: []v1alpha1.TemplateOption{
										{
											Name: "source-1",
											Selector: v1alpha1.Selector{
												MatchFields: []v1alpha1.FieldSelectorRequirement{
													{
														Key:      "spec.source.git.url",
														Operator: v1alpha1.FieldSelectorOpExists,
													},
												},
											},
										},
										{
											Name: "source-2",
											Selector: v1alpha1.Selector{
												MatchFields: []v1alpha1.FieldSelectorRequirement{
													{
														Key:      "spec.source.git.url",
														Operator: v1alpha1.FieldSelectorOpDoesNotExist,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}
			})

			Context("Well formed supply chain", func() {
				It("creates without error", func() {
					Expect(supplyChain.ValidateCreate()).NotTo(HaveOccurred())
				})

				It("updates without error", func() {
					Expect(supplyChain.ValidateUpdate(oldSupplyChain)).NotTo(HaveOccurred())
				})

				It("deletes without error", func() {
					Expect(supplyChain.ValidateDelete()).NotTo(HaveOccurred())
				})
			})

			Context("Two options with the same name", func() {
				BeforeEach(func() {
					supplyChain.Spec.Resources[0].TemplateRef.Options[0].Name = supplyChain.Spec.Resources[0].TemplateRef.Options[1].Name
				})

				It("on create, it rejects the Resource", func() {
					Expect(supplyChain.ValidateCreate()).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: duplicate template name [source-2] found in options for resource [source-provider]",
					))
				})

				It("on update, it rejects the Resource", func() {
					Expect(supplyChain.ValidateUpdate(oldSupplyChain)).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: duplicate template name [source-2] found in options for resource [source-provider]",
					))
				})

				It("deletes without error", func() {
					Expect(supplyChain.ValidateDelete()).NotTo(HaveOccurred())
				})
			})

			Context("only one option is specified", func() {
				BeforeEach(func() {
					supplyChain.Spec.Resources[0].TemplateRef.Options = []v1alpha1.TemplateOption{
						{
							Name:     "only-option",
							Selector: v1alpha1.Selector{},
						},
					}
				})

				It("on create, it rejects the Resource", func() {
					Expect(supplyChain.ValidateCreate()).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: templateRef.Options must have more than one option",
					))
				})

				It("on update, it rejects the Resource", func() {
					Expect(supplyChain.ValidateUpdate(oldSupplyChain)).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: templateRef.Options must have more than one option",
					))
				})

				It("deletes without error", func() {
					Expect(supplyChain.ValidateDelete()).NotTo(HaveOccurred())
				})
			})

			Context("operator values", func() {
				Context("operator is Exists and has values", func() {
					BeforeEach(func() {
						supplyChain.Spec.Resources[0].TemplateRef.Options[0].Selector.MatchFields[0] = v1alpha1.FieldSelectorRequirement{
							Key:      "something",
							Operator: v1alpha1.FieldSelectorOpExists,
							Values:   []string{"bad"},
						}
					})

					It("on create, it rejects the Resource", func() {
						Expect(supplyChain.ValidateCreate()).To(MatchError(
							"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1]: cannot specify values with operator [Exists]",
						))
					})

					It("on update, it rejects the Resource", func() {
						Expect(supplyChain.ValidateUpdate(oldSupplyChain)).To(MatchError(
							"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1]: cannot specify values with operator [Exists]",
						))
					})

					It("deletes without error", func() {
						Expect(supplyChain.ValidateDelete()).NotTo(HaveOccurred())
					})
				})

				Context("operator is NotExists and has values", func() {
					BeforeEach(func() {
						supplyChain.Spec.Resources[0].TemplateRef.Options[0].Selector.MatchFields[0] = v1alpha1.FieldSelectorRequirement{
							Key:      "something",
							Operator: v1alpha1.FieldSelectorOpDoesNotExist,
							Values:   []string{"bad"},
						}
					})

					It("on create, it rejects the Resource", func() {
						Expect(supplyChain.ValidateCreate()).To(MatchError(
							"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1]: cannot specify values with operator [DoesNotExist]",
						))
					})

					It("on update, it rejects the Resource", func() {
						Expect(supplyChain.ValidateUpdate(oldSupplyChain)).To(MatchError(
							"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1]: cannot specify values with operator [DoesNotExist]",
						))
					})

					It("deletes without error", func() {
						Expect(supplyChain.ValidateDelete()).NotTo(HaveOccurred())
					})
				})
				Context("operator is In and does NOT have values", func() {
					BeforeEach(func() {
						supplyChain.Spec.Resources[0].TemplateRef.Options[0].Selector.MatchFields[0] = v1alpha1.FieldSelectorRequirement{
							Key:      "something",
							Operator: v1alpha1.FieldSelectorOpIn,
						}
					})

					It("on create, it rejects the Resource", func() {
						Expect(supplyChain.ValidateCreate()).To(MatchError(
							"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1]: must specify values with operator [In]",
						))
					})

					It("on update, it rejects the Resource", func() {
						Expect(supplyChain.ValidateUpdate(oldSupplyChain)).To(MatchError(
							"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1]: must specify values with operator [In]",
						))
					})

					It("deletes without error", func() {
						Expect(supplyChain.ValidateDelete()).NotTo(HaveOccurred())
					})
				})
				Context("operator is NotIn and does NOT have values", func() {
					BeforeEach(func() {
						supplyChain.Spec.Resources[0].TemplateRef.Options[0].Selector.MatchFields[0] = v1alpha1.FieldSelectorRequirement{
							Key:      "something",
							Operator: v1alpha1.FieldSelectorOpNotIn,
						}
					})

					It("on create, it rejects the Resource", func() {
						Expect(supplyChain.ValidateCreate()).To(MatchError(
							"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1]: must specify values with operator [NotIn]",
						))
					})

					It("on update, it rejects the Resource", func() {
						Expect(supplyChain.ValidateUpdate(oldSupplyChain)).To(MatchError(
							"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1]: must specify values with operator [NotIn]",
						))
					})

					It("deletes without error", func() {
						Expect(supplyChain.ValidateDelete()).NotTo(HaveOccurred())
					})
				})
			})

			Context("2 options with identical requirements", func() {
				Context("selectors are identical", func() {
					BeforeEach(func() {
						supplyChain.Spec.Resources[0].TemplateRef.Options[0].Selector = supplyChain.Spec.Resources[0].TemplateRef.Options[1].Selector
					})

					It("on create, it rejects the Resource", func() {
						Expect(supplyChain.ValidateCreate()).To(MatchError(
							"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: duplicate selector found in options [source-1, source-2]",
						))
					})

					It("on update, it rejects the Resource", func() {
						Expect(supplyChain.ValidateUpdate(oldSupplyChain)).To(MatchError(
							"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: duplicate selector found in options [source-1, source-2]",
						))
					})

					It("deletes without error", func() {
						Expect(supplyChain.ValidateDelete()).NotTo(HaveOccurred())
					})
				})
			})

			Context("option points to key that doesn't exist in spec", func() {
				BeforeEach(func() {
					supplyChain.Spec.Resources[0].TemplateRef.Options[0].Selector.MatchFields[0].Key = "spec.does.not.exist"
				})

				It("on create, it rejects the Resource", func() {
					Expect(supplyChain.ValidateCreate()).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1]: requirement key [spec.does.not.exist] is not a valid path",
					))
				})

				It("on update, it rejects the Resource", func() {
					Expect(supplyChain.ValidateUpdate(oldSupplyChain)).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1]: requirement key [spec.does.not.exist] is not a valid path",
					))
				})

				It("deletes without error", func() {
					Expect(supplyChain.ValidateDelete()).NotTo(HaveOccurred())
				})
			})

			Context("option points to key that is a valid prefix into an array", func() {
				BeforeEach(func() {
					supplyChain.Spec.Resources[0].TemplateRef.Options[0].Selector.MatchFields[0].Key = `spec.env[?(@.name=="some-name")].value`
				})

				It("on create, it does not reject the Resource", func() {
					err := supplyChain.ValidateCreate()
					Expect(err).NotTo(HaveOccurred())
				})

				It("on update, it does not reject the Resource", func() {
					err := supplyChain.ValidateUpdate(oldSupplyChain)
					Expect(err).NotTo(HaveOccurred())
				})

				It("deletes without error", func() {
					Expect(supplyChain.ValidateDelete()).NotTo(HaveOccurred())
				})
			})

			Context("both name and options specified", func() {
				BeforeEach(func() {
					supplyChain.Spec.Resources[0].TemplateRef.Name = "some-name"
				})

				It("on create, it rejects the Resource", func() {
					Expect(supplyChain.ValidateCreate()).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: exactly one of templateRef.Name or templateRef.Options must be specified, found both",
					))
				})

				It("on update, it rejects the Resource", func() {
					Expect(supplyChain.ValidateUpdate(oldSupplyChain)).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: exactly one of templateRef.Name or templateRef.Options must be specified, found both",
					))
				})

				It("deletes without error", func() {
					Expect(supplyChain.ValidateDelete()).NotTo(HaveOccurred())
				})
			})

			Context("neither name and options are specified", func() {
				BeforeEach(func() {
					supplyChain.Spec.Resources[0].TemplateRef.Options = nil
				})

				It("on create, it rejects the Resource", func() {
					Expect(supplyChain.ValidateCreate()).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: exactly one of templateRef.Name or templateRef.Options must be specified, found neither",
					))
				})

				It("on update, it rejects the Resource", func() {
					Expect(supplyChain.ValidateUpdate(oldSupplyChain)).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: exactly one of templateRef.Name or templateRef.Options must be specified, found neither",
					))
				})

				It("deletes without error", func() {
					Expect(supplyChain.ValidateDelete()).NotTo(HaveOccurred())
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
