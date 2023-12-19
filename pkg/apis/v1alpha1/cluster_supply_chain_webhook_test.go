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
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

var _ = Describe("Webhook Validation", func() {
	Context("supply chain without options", func() {
		var (
			supplyChain    *v1alpha1.ClusterSupplyChain
			oldSupplyChain *v1alpha1.ClusterSupplyChain
		)

		BeforeEach(func() {
			supplyChain = &v1alpha1.ClusterSupplyChain{
				ObjectMeta: metav1.ObjectMeta{
					Name: "responsible-ops---default-params",
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
					LegacySelector: v1alpha1.LegacySelector{
						Selector: map[string]string{"integration-test": "workload-no-supply-chain"},
					},
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
				_, err := supplyChain.ValidateCreate()
				Expect(err).NotTo(HaveOccurred())
			})

			It("updates without error", func() {
				_, err := supplyChain.ValidateUpdate(oldSupplyChain)
				Expect(err).NotTo(HaveOccurred())
			})

			It("deletes without error", func() {
				_, err := supplyChain.ValidateDelete()
				Expect(err).NotTo(HaveOccurred())
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
				_, err := supplyChain.ValidateCreate()
				Expect(err).To(MatchError(
					"error validating clustersupplychain [responsible-ops---default-params]: invalid sources for resource [other-source-provider]: [some-source] is provided by unknown resource [some-nonexistent-resource]",
				))
			})

			It("on update, returns an error", func() {
				_, err := supplyChain.ValidateUpdate(oldSupplyChain)
				Expect(err).To(MatchError(
					"error validating clustersupplychain [responsible-ops---default-params]: invalid sources for resource [other-source-provider]: [some-source] is provided by unknown resource [some-nonexistent-resource]",
				))
			})

			It("deletes without error", func() {
				_, err := supplyChain.ValidateDelete()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("Two resources with the same name", func() {
			BeforeEach(func() {
				for i := range supplyChain.Spec.Resources {
					supplyChain.Spec.Resources[i].Name = "some-duplicate-name"
				}
			})

			It("on create, it rejects the Resource", func() {
				_, err := supplyChain.ValidateCreate()
				Expect(err).To(MatchError(
					"error validating clustersupplychain [responsible-ops---default-params]: duplicate resource name [some-duplicate-name] found",
				))
			})

			It("on update, it rejects the Resource", func() {
				_, err := supplyChain.ValidateUpdate(oldSupplyChain)
				Expect(err).To(MatchError(
					"error validating clustersupplychain [responsible-ops---default-params]: duplicate resource name [some-duplicate-name] found",
				))
			})

			It("deletes without error", func() {
				_, err := supplyChain.ValidateDelete()
				Expect(err).NotTo(HaveOccurred())
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
						_, err := supplyChain.ValidateCreate()
						Expect(err).To(MatchError(
							"error validating clustersupplychain [responsible-ops---default-params]: param [some-param] is invalid: must set exactly one of value and default",
						))
					})

					It("on update, returns an error", func() {
						_, err := supplyChain.ValidateUpdate(oldSupplyChain)
						Expect(err).To(MatchError(
							"error validating clustersupplychain [responsible-ops---default-params]: param [some-param] is invalid: must set exactly one of value and default",
						))
					})

					It("deletes without error", func() {
						_, err := supplyChain.ValidateDelete()
						Expect(err).NotTo(HaveOccurred())
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
						_, err := supplyChain.ValidateCreate()
						Expect(err).To(MatchError(
							"error validating clustersupplychain [responsible-ops---default-params]: param [some-param] is invalid: must set exactly one of value and default",
						))
					})

					It("on update, returns an error", func() {
						_, err := supplyChain.ValidateUpdate(oldSupplyChain)
						Expect(err).To(MatchError(
							"error validating clustersupplychain [responsible-ops---default-params]: param [some-param] is invalid: must set exactly one of value and default",
						))
					})

					It("deletes without error", func() {
						_, err := supplyChain.ValidateDelete()
						Expect(err).NotTo(HaveOccurred())
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
						_, err := supplyChain.ValidateCreate()
						Expect(err).To(MatchError(
							"error validating clustersupplychain [responsible-ops---default-params]: resource [source-provider] is invalid: param [some-param] is invalid: must set exactly one of value and default",
						))
					})

					It("on update, returns an error", func() {
						_, err := supplyChain.ValidateUpdate(oldSupplyChain)
						Expect(err).To(MatchError(
							"error validating clustersupplychain [responsible-ops---default-params]: resource [source-provider] is invalid: param [some-param] is invalid: must set exactly one of value and default",
						))
					})

					It("deletes without error", func() {
						_, err := supplyChain.ValidateDelete()
						Expect(err).NotTo(HaveOccurred())
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
						_, err := supplyChain.ValidateCreate()
						Expect(err).To(MatchError(
							"error validating clustersupplychain [responsible-ops---default-params]: resource [source-provider] is invalid: param [some-param] is invalid: must set exactly one of value and default",
						))
					})

					It("on update, returns an error", func() {
						_, err := supplyChain.ValidateUpdate(oldSupplyChain)
						Expect(err).To(MatchError(
							"error validating clustersupplychain [responsible-ops---default-params]: resource [source-provider] is invalid: param [some-param] is invalid: must set exactly one of value and default",
						))
					})

					It("deletes without error", func() {
						_, err := supplyChain.ValidateDelete()
						Expect(err).NotTo(HaveOccurred())
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
						Name: "responsible-ops---default-params",
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
						LegacySelector: v1alpha1.LegacySelector{
							Selector: map[string]string{"integration-test": "workload-no-supply-chain"},
						},
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
					_, createErr := supplyChain.ValidateCreate()

					// Update
					_, updateErr := supplyChain.ValidateUpdate(oldSupplyChain)

					// Delete
					_, deleteErr := supplyChain.ValidateDelete()

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

	Describe("OneOf Selector, SelectorMatchExpressions, or SelectorMatchFields", func() {
		var supplyChainFactory = func(selector map[string]string, expressions []metav1.LabelSelectorRequirement, fields []v1alpha1.FieldSelectorRequirement) *v1alpha1.ClusterSupplyChain {
			return &v1alpha1.ClusterSupplyChain{
				ObjectMeta: metav1.ObjectMeta{
					Name: "responsible-ops---default-params",
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
					},
					LegacySelector: v1alpha1.LegacySelector{
						Selector:                 selector,
						SelectorMatchExpressions: expressions,
						SelectorMatchFields:      fields,
					},
				},
			}
		}
		Context("No selection", func() {
			var supplyChain *v1alpha1.ClusterSupplyChain
			BeforeEach(func() {
				supplyChain = supplyChainFactory(nil, nil, nil)
			})

			It("on create, returns an error", func() {
				_, err := supplyChain.ValidateCreate()
				Expect(err).To(MatchError(
					"error validating clustersupplychain [responsible-ops---default-params]: at least one selector, selectorMatchExpression, selectorMatchField must be specified",
				))
			})

			It("on update, returns an error", func() {
				_, err := supplyChain.ValidateUpdate(nil)
				Expect(err).To(MatchError(
					"error validating clustersupplychain [responsible-ops---default-params]: at least one selector, selectorMatchExpression, selectorMatchField must be specified",
				))
			})

			It("deletes without error", func() {
				_, err := supplyChain.ValidateDelete()
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("Empty selection", func() {
			var supplyChain *v1alpha1.ClusterSupplyChain
			BeforeEach(func() {
				supplyChain = supplyChainFactory(map[string]string{}, []metav1.LabelSelectorRequirement{}, []v1alpha1.FieldSelectorRequirement{})
			})

			It("on create, returns an error", func() {
				_, err := supplyChain.ValidateCreate()
				Expect(err).To(MatchError(
					"error validating clustersupplychain [responsible-ops---default-params]: at least one selector, selectorMatchExpression, selectorMatchField must be specified",
				))
			})

			It("on update, returns an error", func() {
				_, err := supplyChain.ValidateUpdate(nil)
				Expect(err).To(MatchError(
					"error validating clustersupplychain [responsible-ops---default-params]: at least one selector, selectorMatchExpression, selectorMatchField must be specified",
				))
			})

			It("deletes without error", func() {
				_, err := supplyChain.ValidateDelete()
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("A Selector", func() {
			var supplyChain *v1alpha1.ClusterSupplyChain
			BeforeEach(func() {
				supplyChain = supplyChainFactory(map[string]string{"foo": "bar"}, nil, nil)
			})

			It("creates without error", func() {
				_, err := supplyChain.ValidateCreate()
				Expect(err).NotTo(HaveOccurred())
			})

			It("updates without error", func() {
				_, err := supplyChain.ValidateUpdate(nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("deletes without error", func() {
				_, err := supplyChain.ValidateDelete()
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("A SelectorMatchExpression", func() {
			var supplyChain *v1alpha1.ClusterSupplyChain
			BeforeEach(func() {
				supplyChain = supplyChainFactory(nil, []metav1.LabelSelectorRequirement{{Key: "whatever", Operator: "Exists"}}, nil)
			})

			It("creates without error", func() {
				_, err := supplyChain.ValidateCreate()
				Expect(err).NotTo(HaveOccurred())
			})

			It("updates without error", func() {
				_, err := supplyChain.ValidateUpdate(nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("deletes without error", func() {
				_, err := supplyChain.ValidateDelete()
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("A SelectorMatchFields", func() {
			var supplyChain *v1alpha1.ClusterSupplyChain
			BeforeEach(func() {
				supplyChain = supplyChainFactory(nil, nil, []v1alpha1.FieldSelectorRequirement{{Key: "metadata.whatever", Operator: "Exists"}})
			})

			It("creates without error", func() {
				_, err := supplyChain.ValidateCreate()
				Expect(err).NotTo(HaveOccurred())
			})

			It("updates without error", func() {
				_, err := supplyChain.ValidateUpdate(nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("deletes without error", func() {
				_, err := supplyChain.ValidateDelete()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("selector validations", func() {
		var (
			supplyChain *v1alpha1.ClusterSupplyChain
		)
		BeforeEach(func() {
			supplyChain = &v1alpha1.ClusterSupplyChain{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-name",
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

		Context("selector is selecting on SelectorMatchFields", func() {
			Context("valid field selector", func() {
				BeforeEach(func() {
					supplyChain.Spec.LegacySelector = v1alpha1.LegacySelector{
						SelectorMatchFields: []v1alpha1.FieldSelectorRequirement{
							{
								Key:      "spec.env[0].something",
								Operator: "Exists",
							},
						},
					}
				})
				It("creates without error", func() {
					_, err := supplyChain.ValidateCreate()
					Expect(err).NotTo(HaveOccurred())
				})

			})

			Context("invalid json path in field selector", func() {
				BeforeEach(func() {
					supplyChain.Spec.LegacySelector = v1alpha1.LegacySelector{
						SelectorMatchFields: []v1alpha1.FieldSelectorRequirement{
							{
								Key:      "spec.env[0].{{}}",
								Operator: "Exists",
							},
						},
					}
				})
				It("rejects with an error", func() {
					_, err := supplyChain.ValidateCreate()
					Expect(err).To(MatchError("error validating clustersupplychain [some-name]: invalid jsonpath for key [spec.env[0].{{}}]: unrecognized character in action: U+007B '{'"))
				})
			})

			Context("field selector is not in accepted list", func() {
				BeforeEach(func() {
					supplyChain.Spec.LegacySelector = v1alpha1.LegacySelector{
						SelectorMatchFields: []v1alpha1.FieldSelectorRequirement{
							{
								Key:      "foo",
								Operator: "Exists",
							},
						},
					}
				})
				It("rejects with an error", func() {
					_, err := supplyChain.ValidateCreate()
					Expect(err).To(MatchError("error validating clustersupplychain [some-name]: requirement key [foo] is not a valid path"))
				})
			})
		})

		Context("selector is selecting on SelectorMatchExpressions", func() {
			Context("there is a valid selector", func() {
				BeforeEach(func() {
					supplyChain.Spec.LegacySelector = v1alpha1.LegacySelector{
						SelectorMatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "my-label",
								Operator: "Exists",
							},
						},
					}
				})

				It("creates without error", func() {
					_, err := supplyChain.ValidateCreate()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("there is an invalid selector", func() {
				BeforeEach(func() {
					supplyChain.Spec.LegacySelector = v1alpha1.LegacySelector{
						SelectorMatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "-my-label",
								Operator: "Exists",
							},
						},
					}
				})

				It("rejects with error", func() {
					_, err := supplyChain.ValidateCreate()
					Expect(err).To(MatchError(ContainSubstring("error validating clustersupplychain [some-name]: selectorMatchExpressions are not valid: key: Invalid value: \"-my-label\"")))
				})
			})
		})

		Context("selector is selecting on Selector", func() {
			Context("there is a valid selector", func() {
				BeforeEach(func() {
					supplyChain.Spec.LegacySelector = v1alpha1.LegacySelector{
						Selector: map[string]string{"my-label": "some-value"},
					}
				})

				It("creates without error", func() {
					_, err := supplyChain.ValidateCreate()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("there is an invalid selector", func() {
				BeforeEach(func() {
					supplyChain.Spec.LegacySelector = v1alpha1.LegacySelector{
						Selector: map[string]string{"-my-label": "some-value"},
					}
				})

				It("rejects with error", func() {
					_, err := supplyChain.ValidateCreate()
					Expect(err).To(MatchError(ContainSubstring("error validating clustersupplychain [some-name]: selector is not valid: key: Invalid value: \"-my-label\"")))
				})
			})
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
					Name: "responsible-ops---default-params",
				},
				Spec: v1alpha1.SupplyChainSpec{
					LegacySelector: v1alpha1.LegacySelector{
						Selector: map[string]string{"foo": "bar"},
					},
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
				_, err := supplyChain.ValidateCreate()
				Expect(err).NotTo(HaveOccurred())
			})

			It("updates without error", func() {
				_, err := supplyChain.ValidateUpdate(oldSupplyChain)
				Expect(err).NotTo(HaveOccurred())
			})

			It("deletes without error", func() {
				_, err := supplyChain.ValidateDelete()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("Two options with the same name", func() {
			BeforeEach(func() {
				supplyChain.Spec.Resources[0].TemplateRef.Options[0].Name = supplyChain.Spec.Resources[0].TemplateRef.Options[1].Name
			})

			It("on create, it rejects the Resource", func() {
				_, err := supplyChain.ValidateCreate()
				Expect(err).To(MatchError(
					"error validating clustersupplychain [responsible-ops---default-params]: duplicate template name [source-2] found in options for resource [source-provider]",
				))
			})

			It("on update, it rejects the Resource", func() {
				_, err := supplyChain.ValidateUpdate(oldSupplyChain)
				Expect(err).To(MatchError(
					"error validating clustersupplychain [responsible-ops---default-params]: duplicate template name [source-2] found in options for resource [source-provider]",
				))
			})

			It("deletes without error", func() {
				_, err := supplyChain.ValidateDelete()
				Expect(err).NotTo(HaveOccurred())
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
				_, err := supplyChain.ValidateCreate()
				Expect(err).To(MatchError(
					"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: templateRef.Options must have more than one option",
				))
			})

			It("on update, it rejects the Resource", func() {
				_, err := supplyChain.ValidateUpdate(oldSupplyChain)
				Expect(err).To(MatchError(
					"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: templateRef.Options must have more than one option",
				))
			})

			It("deletes without error", func() {
				_, err := supplyChain.ValidateDelete()
				Expect(err).NotTo(HaveOccurred())
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
					_, err := supplyChain.ValidateCreate()
					Expect(err).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1] selector: cannot specify values with operator [Exists]",
					))
				})

				It("on update, it rejects the Resource", func() {
					_, err := supplyChain.ValidateUpdate(oldSupplyChain)
					Expect(err).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1] selector: cannot specify values with operator [Exists]",
					))
				})

				It("deletes without error", func() {
					_, err := supplyChain.ValidateDelete()
					Expect(err).NotTo(HaveOccurred())
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
					_, err := supplyChain.ValidateCreate()
					Expect(err).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1] selector: cannot specify values with operator [DoesNotExist]",
					))
				})

				It("on update, it rejects the Resource", func() {
					_, err := supplyChain.ValidateUpdate(oldSupplyChain)
					Expect(err).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1] selector: cannot specify values with operator [DoesNotExist]",
					))
				})

				It("deletes without error", func() {
					_, err := supplyChain.ValidateDelete()
					Expect(err).NotTo(HaveOccurred())
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
					_, err := supplyChain.ValidateCreate()
					Expect(err).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1] selector: must specify values with operator [In]",
					))
				})

				It("on update, it rejects the Resource", func() {
					_, err := supplyChain.ValidateUpdate(oldSupplyChain)
					Expect(err).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1] selector: must specify values with operator [In]",
					))
				})

				It("deletes without error", func() {
					_, err := supplyChain.ValidateDelete()
					Expect(err).NotTo(HaveOccurred())
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
					_, err := supplyChain.ValidateCreate()
					Expect(err).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1] selector: must specify values with operator [NotIn]",
					))
				})

				It("on update, it rejects the Resource", func() {
					_, err := supplyChain.ValidateUpdate(oldSupplyChain)
					Expect(err).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1] selector: must specify values with operator [NotIn]",
					))
				})

				It("deletes without error", func() {
					_, err := supplyChain.ValidateDelete()
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("2 options with identical requirements", func() {
			Context("selectors are identical", func() {
				BeforeEach(func() {
					supplyChain.Spec.Resources[0].TemplateRef.Options[0].Selector = supplyChain.Spec.Resources[0].TemplateRef.Options[1].Selector
				})

				It("on create, it rejects the Resource", func() {
					_, err := supplyChain.ValidateCreate()
					Expect(err).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: duplicate selector found in options [source-1, source-2]",
					))
				})

				It("on update, it rejects the Resource", func() {
					_, err := supplyChain.ValidateUpdate(oldSupplyChain)
					Expect(err).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: duplicate selector found in options [source-1, source-2]",
					))
				})

				It("deletes without error", func() {
					_, err := supplyChain.ValidateDelete()
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("option points to key that doesn't exist in spec", func() {
			BeforeEach(func() {
				supplyChain.Spec.Resources[0].TemplateRef.Options[0].Selector.MatchFields[0].Key = "spec.does.not.exist"
			})

			It("on create, it rejects the Resource", func() {
				_, err := supplyChain.ValidateCreate()
				Expect(err).To(MatchError(
					"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1] selector: requirement key [spec.does.not.exist] is not a valid path",
				))
			})

			It("on update, it rejects the Resource", func() {
				_, err := supplyChain.ValidateUpdate(oldSupplyChain)
				Expect(err).To(MatchError(
					"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1] selector: requirement key [spec.does.not.exist] is not a valid path",
				))
			})

			It("deletes without error", func() {
				_, err := supplyChain.ValidateDelete()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("option has invalid label selector", func() {
			BeforeEach(func() {
				supplyChain.Spec.Resources[0].TemplateRef.Options[0].Selector.LabelSelector.MatchLabels = map[string]string{"not-valid-": "like-this-"}
			})

			It("on create, it rejects the Resource", func() {
				_, err := supplyChain.ValidateCreate()
				Expect(err).To(MatchError(ContainSubstring(
					`error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1] selector: matchLabels are not valid: [key: Invalid value: "not-valid-"`,
				)))
			})

			It("on update, it rejects the Resource", func() {
				_, err := supplyChain.ValidateUpdate(oldSupplyChain)
				Expect(err).To(MatchError(ContainSubstring(
					`error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1] selector: matchLabels are not valid: [key: Invalid value: "not-valid-"`,
				)))
			})

			It("deletes without error", func() {
				_, err := supplyChain.ValidateDelete()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("option points to key that is a valid prefix into an array", func() {
			BeforeEach(func() {
				supplyChain.Spec.Resources[0].TemplateRef.Options[0].Selector.MatchFields[0].Key = `spec.env[?(@.name=="some-name")].value`
			})

			It("on create, it does not reject the Resource", func() {
				_, err := supplyChain.ValidateCreate()
				Expect(err).NotTo(HaveOccurred())
			})

			It("on update, it does not reject the Resource", func() {
				_, err := supplyChain.ValidateUpdate(oldSupplyChain)
				Expect(err).NotTo(HaveOccurred())
			})

			It("deletes without error", func() {
				_, err := supplyChain.ValidateDelete()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("both name and options specified", func() {
			BeforeEach(func() {
				supplyChain.Spec.Resources[0].TemplateRef.Name = "some-name"
			})

			It("on create, it rejects the Resource", func() {
				_, err := supplyChain.ValidateCreate()
				Expect(err).To(MatchError(
					"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: exactly one of templateRef.Name or templateRef.Options must be specified, found both",
				))
			})

			It("on update, it rejects the Resource", func() {
				_, err := supplyChain.ValidateUpdate(oldSupplyChain)
				Expect(err).To(MatchError(
					"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: exactly one of templateRef.Name or templateRef.Options must be specified, found both",
				))
			})

			It("deletes without error", func() {
				_, err := supplyChain.ValidateDelete()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("neither name and options are specified", func() {
			BeforeEach(func() {
				supplyChain.Spec.Resources[0].TemplateRef.Options = nil
			})

			It("on create, it rejects the Resource", func() {
				_, err := supplyChain.ValidateCreate()
				Expect(err).To(MatchError(
					"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: exactly one of templateRef.Name or templateRef.Options must be specified, found neither",
				))
			})

			It("on update, it rejects the Resource", func() {
				_, err := supplyChain.ValidateUpdate(oldSupplyChain)
				Expect(err).To(MatchError(
					"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: exactly one of templateRef.Name or templateRef.Options must be specified, found neither",
				))
			})

			It("deletes without error", func() {
				_, err := supplyChain.ValidateDelete()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("option name and option pass through both not specified", func() {
			BeforeEach(func() {
				supplyChain.Spec.Resources[0].TemplateRef.Options[0].Name = ""
			})

			It("on create, it rejects the Resource", func() {
				_, err := supplyChain.ValidateCreate()
				Expect(err).To(MatchError(
					"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: exactly one of option.Name or option.PassThrough must be specified, found neither",
				))
			})

			It("on update, it rejects the Resource", func() {
				_, err := supplyChain.ValidateUpdate(oldSupplyChain)
				Expect(err).To(MatchError(
					"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: exactly one of option.Name or option.PassThrough must be specified, found neither",
				))
			})

			It("deletes without error", func() {
				_, err := supplyChain.ValidateDelete()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("options with pass through", func() {
			BeforeEach(func() {
				supplyChain.Spec.Resources[0].Sources = []v1alpha1.ResourceReference{
					{
						Name:     "not-empty",
						Resource: "another-resource",
					},
				}

				supplyChain.Spec.Resources = append(supplyChain.Spec.Resources, v1alpha1.SupplyChainResource{
					Name: "another-resource",
					TemplateRef: v1alpha1.SupplyChainTemplateReference{
						Name: "my-name",
						Kind: "ClusterSourceTemplate",
					},
				})
			})

			Context("pass through refers to an input that does not exist", func() {
				BeforeEach(func() {
					supplyChain.Spec.Resources[0].TemplateRef.Options[0].Name = ""
					supplyChain.Spec.Resources[0].TemplateRef.Options[0].PassThrough = "wrong-input"
				})

				It("on create, it rejects the Resource", func() {
					_, err := supplyChain.ValidateCreate()
					Expect(err).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: pass through [wrong-input] does not refer to a known input",
					))
				})

				It("on update, it rejects the Resource", func() {
					_, err := supplyChain.ValidateUpdate(oldSupplyChain)
					Expect(err).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: pass through [wrong-input] does not refer to a known input",
					))
				})

				It("deletes without error", func() {
					_, err := supplyChain.ValidateDelete()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("more than one pass through specified", func() {
				BeforeEach(func() {
					supplyChain.Spec.Resources[0].TemplateRef.Options[0].PassThrough = "not-empty"
					supplyChain.Spec.Resources[0].TemplateRef.Options[1].PassThrough = "not-empty-also"
				})

				It("on create, it rejects the Resource", func() {
					_, err := supplyChain.ValidateCreate()
					Expect(err).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: cannot have more than one pass through option, found 2",
					))
				})

				It("on update, it rejects the Resource", func() {
					_, err := supplyChain.ValidateUpdate(oldSupplyChain)
					Expect(err).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: cannot have more than one pass through option, found 2",
					))
				})

				It("deletes without error", func() {
					_, err := supplyChain.ValidateDelete()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("option name and option pass through both specified", func() {
				BeforeEach(func() {
					supplyChain.Spec.Resources[0].TemplateRef.Options[0].PassThrough = "not-empty"
				})

				It("on create, it rejects the Resource", func() {
					_, err := supplyChain.ValidateCreate()
					Expect(err).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: exactly one of option.Name or option.PassThrough must be specified, found both",
					))
				})

				It("on update, it rejects the Resource", func() {
					_, err := supplyChain.ValidateUpdate(oldSupplyChain)
					Expect(err).To(MatchError(
						"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: exactly one of option.Name or option.PassThrough must be specified, found both",
					))
				})

				It("deletes without error", func() {
					_, err := supplyChain.ValidateDelete()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("2 options with identical requirements", func() {
				Context("selectors are identical", func() {
					BeforeEach(func() {
						supplyChain.Spec.Resources[0].TemplateRef.Options[0].Selector = supplyChain.Spec.Resources[0].TemplateRef.Options[1].Selector
						supplyChain.Spec.Resources[0].TemplateRef.Options[1].PassThrough = "not-empty"
						supplyChain.Spec.Resources[0].TemplateRef.Options[1].Name = ""
					})

					It("on create, it rejects the Resource", func() {
						_, err := supplyChain.ValidateCreate()
						Expect(err).To(MatchError(
							"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: duplicate selector found in options [source-1, passThrough]",
						))
					})

					It("on update, it rejects the Resource", func() {
						_, err := supplyChain.ValidateUpdate(oldSupplyChain)
						Expect(err).To(MatchError(
							"error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: duplicate selector found in options [source-1, passThrough]",
						))
					})

					It("deletes without error", func() {
						_, err := supplyChain.ValidateDelete()
						Expect(err).NotTo(HaveOccurred())
					})
				})
			})

			Context("well formed", func() {
				BeforeEach(func() {
					supplyChain.Spec.Resources[0].TemplateRef.Options[0].Name = ""
					supplyChain.Spec.Resources[0].TemplateRef.Options[0].PassThrough = "not-empty"
				})

				It("creates without error", func() {
					_, err := supplyChain.ValidateCreate()
					Expect(err).NotTo(HaveOccurred())
				})

				It("updates without error", func() {
					_, err := supplyChain.ValidateUpdate(oldSupplyChain)
					Expect(err).NotTo(HaveOccurred())
				})

				It("deletes without error", func() {
					_, err := supplyChain.ValidateDelete()
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
})
