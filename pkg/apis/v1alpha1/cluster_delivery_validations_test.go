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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

var _ = Describe("Delivery Validation", func() {

	Context("delivery without options", func() {
		var (
			delivery    *v1alpha1.ClusterDelivery
			oldDelivery *v1alpha1.ClusterDelivery
		)

		BeforeEach(func() {
			delivery = &v1alpha1.ClusterDelivery{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "delivery-resource",
					Namespace: "default",
				},
				Spec: v1alpha1.DeliverySpec{
					Selector: map[string]string { "requires": "at-least-one" },
					Resources: []v1alpha1.DeliveryResource{
						{
							Name: "source-provider",
							TemplateRef: v1alpha1.DeliveryTemplateReference{
								Kind: "ClusterSourceTemplate",
								Name: "source-template",
							},
						},
						{
							Name: "other-source-provider",
							TemplateRef: v1alpha1.DeliveryTemplateReference{
								Kind: "ClusterSourceTemplate",
								Name: "source-template",
							},
						},
					},
				},
			}
		})
		Context("Well formed delivery", func() {
			It("creates without error", func() {
				Expect(delivery.ValidateCreate()).NotTo(HaveOccurred())
			})

			It("updates without error", func() {
				Expect(delivery.ValidateUpdate(oldDelivery)).NotTo(HaveOccurred())
			})
		})

		Context("Duplicate resource names", func() {
			BeforeEach(func() {
				for i := range delivery.Spec.Resources {
					delivery.Spec.Resources[i].Name = "source-provider"
				}
			})

			It("on create, it rejects the Resource", func() {
				Expect(delivery.ValidateCreate()).To(MatchError(`error validating clusterdelivery [delivery-resource]: spec.resources[1].name "source-provider" cannot appear twice`))
			})

			It("on update, it rejects the Resource", func() {
				Expect(delivery.ValidateUpdate(oldDelivery)).To(MatchError(`error validating clusterdelivery [delivery-resource]: spec.resources[1].name "source-provider" cannot appear twice`))
			})
		})

		Context("Delivery with malformed params", func() {
			Context("Top level params are malformed", func() {
				Context("param does not specify a value or default", func() {
					BeforeEach(func() {
						delivery.Spec.Params = []v1alpha1.BlueprintParam{
							{
								Name: "some-param",
							},
						}
					})
					It("on create, it rejects the Resource", func() {
						Expect(delivery.ValidateCreate()).To(MatchError(
							"error validating clusterdelivery [delivery-resource]: param [some-param] is invalid: must set exactly one of value and default",
						))
					})

					It("on update, it rejects the Resource", func() {
						Expect(delivery.ValidateUpdate(oldDelivery)).To(MatchError(
							"error validating clusterdelivery [delivery-resource]: param [some-param] is invalid: must set exactly one of value and default",
						))
					})
				})

				Context("param specifies both a value and a default", func() {
					BeforeEach(func() {
						delivery.Spec.Params = []v1alpha1.BlueprintParam{
							{
								Name:         "some-param",
								Value:        &apiextensionsv1.JSON{Raw: []byte(`"some value"`)},
								DefaultValue: &apiextensionsv1.JSON{Raw: []byte(`"some value"`)},
							},
						}
					})

					It("on create, it rejects the Resource", func() {
						Expect(delivery.ValidateCreate()).To(MatchError(
							"error validating clusterdelivery [delivery-resource]: param [some-param] is invalid: must set exactly one of value and default",
						))
					})

					It("on update, it rejects the Resource", func() {
						Expect(delivery.ValidateUpdate(oldDelivery)).To(MatchError(
							"error validating clusterdelivery [delivery-resource]: param [some-param] is invalid: must set exactly one of value and default",
						))
					})
				})
			})

			Context("Params of an individual resource are malformed", func() {
				Context("param does not specify a value or default", func() {
					BeforeEach(func() {
						delivery.Spec.Resources[0].Params = []v1alpha1.BlueprintParam{
							{
								Name: "some-param",
							},
						}
					})
					It("on create, it rejects the Resource", func() {
						Expect(delivery.ValidateCreate()).To(MatchError(
							"error validating clusterdelivery [delivery-resource]: resource [source-provider] is invalid: param [some-param] is invalid: must set exactly one of value and default",
						))
					})

					It("on update, it rejects the Resource", func() {
						Expect(delivery.ValidateUpdate(oldDelivery)).To(MatchError(
							"error validating clusterdelivery [delivery-resource]: resource [source-provider] is invalid: param [some-param] is invalid: must set exactly one of value and default",
						))
					})
				})

				Context("param specifies both a value and a default", func() {
					BeforeEach(func() {
						delivery.Spec.Resources[0].Params = []v1alpha1.BlueprintParam{
							{
								Name:         "some-param",
								Value:        &apiextensionsv1.JSON{Raw: []byte(`"some value"`)},
								DefaultValue: &apiextensionsv1.JSON{Raw: []byte(`"some value"`)},
							},
						}
					})
					It("on create, it rejects the Resourcer", func() {
						Expect(delivery.ValidateCreate()).To(MatchError(
							"error validating clusterdelivery [delivery-resource]: resource [source-provider] is invalid: param [some-param] is invalid: must set exactly one of value and default",
						))
					})

					It("on update, it rejects the Resourcer", func() {
						Expect(delivery.ValidateUpdate(oldDelivery)).To(MatchError(
							"error validating clusterdelivery [delivery-resource]: resource [source-provider] is invalid: param [some-param] is invalid: must set exactly one of value and default",
						))
					})
				})
			})
		})
	})

	Context("delivery with options", func() {
		var (
			delivery    *v1alpha1.ClusterDelivery
			oldDelivery *v1alpha1.ClusterDelivery
		)

		BeforeEach(func() {
			delivery = &v1alpha1.ClusterDelivery{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "responsible-ops---default-params",
					Namespace: "default",
				},
				Spec: v1alpha1.DeliverySpec{
					Selector: map[string]string { "one-selector-of-any-kind": "is-needed" },
					Resources: []v1alpha1.DeliveryResource{
						{
							Name: "source-provider",
							TemplateRef: v1alpha1.DeliveryTemplateReference{
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

		Context("Well formed delivery", func() {
			It("creates without error", func() {
				Expect(delivery.ValidateCreate()).NotTo(HaveOccurred())
			})

			It("updates without error", func() {
				Expect(delivery.ValidateUpdate(oldDelivery)).NotTo(HaveOccurred())
			})

			It("deletes without error", func() {
				Expect(delivery.ValidateDelete()).NotTo(HaveOccurred())
			})
		})

		Context("Two options with the same name", func() {
			BeforeEach(func() {
				delivery.Spec.Resources[0].TemplateRef.Options[0].Name = delivery.Spec.Resources[0].TemplateRef.Options[1].Name
			})

			It("on create, it rejects the Resource", func() {
				Expect(delivery.ValidateCreate()).To(MatchError(
					"error validating clusterdelivery [responsible-ops---default-params]: duplicate template name [source-2] found in options for resource [source-provider]",
				))
			})

			It("on update, it rejects the Resource", func() {
				Expect(delivery.ValidateUpdate(oldDelivery)).To(MatchError(
					"error validating clusterdelivery [responsible-ops---default-params]: duplicate template name [source-2] found in options for resource [source-provider]",
				))
			})

			It("deletes without error", func() {
				Expect(delivery.ValidateDelete()).NotTo(HaveOccurred())
			})
		})

		Context("only one option is specified", func() {
			BeforeEach(func() {
				delivery.Spec.Resources[0].TemplateRef.Options = []v1alpha1.TemplateOption{
					{
						Name:     "only-option",
						Selector: v1alpha1.Selector{},
					},
				}
			})

			It("on create, it rejects the Resource", func() {
				Expect(delivery.ValidateCreate()).To(MatchError(
					"error validating clusterdelivery [responsible-ops---default-params]: error validating resource [source-provider]: templateRef.Options must have more than one option",
				))
			})

			It("on update, it rejects the Resource", func() {
				Expect(delivery.ValidateUpdate(oldDelivery)).To(MatchError(
					"error validating clusterdelivery [responsible-ops---default-params]: error validating resource [source-provider]: templateRef.Options must have more than one option",
				))
			})

			It("deletes without error", func() {
				Expect(delivery.ValidateDelete()).NotTo(HaveOccurred())
			})
		})

		Context("operator values", func() {
			Context("operator is Exists and has values", func() {
				BeforeEach(func() {
					delivery.Spec.Resources[0].TemplateRef.Options[0].Selector.MatchFields[0] = v1alpha1.FieldSelectorRequirement{
						Key:      "something",
						Operator: v1alpha1.FieldSelectorOpExists,
						Values:   []string{"bad"},
					}
				})

				It("on create, it rejects the Resource", func() {
					Expect(delivery.ValidateCreate()).To(MatchError(
						"error validating clusterdelivery [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1]: cannot specify values with operator [Exists]",
					))
				})

				It("on update, it rejects the Resource", func() {
					Expect(delivery.ValidateUpdate(oldDelivery)).To(MatchError(
						"error validating clusterdelivery [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1]: cannot specify values with operator [Exists]",
					))
				})

				It("deletes without error", func() {
					Expect(delivery.ValidateDelete()).NotTo(HaveOccurred())
				})
			})

			Context("operator is NotExists and has values", func() {
				BeforeEach(func() {
					delivery.Spec.Resources[0].TemplateRef.Options[0].Selector.MatchFields[0] = v1alpha1.FieldSelectorRequirement{
						Key:      "something",
						Operator: v1alpha1.FieldSelectorOpDoesNotExist,
						Values:   []string{"bad"},
					}
				})

				It("on create, it rejects the Resource", func() {
					Expect(delivery.ValidateCreate()).To(MatchError(
						"error validating clusterdelivery [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1]: cannot specify values with operator [DoesNotExist]",
					))
				})

				It("on update, it rejects the Resource", func() {
					Expect(delivery.ValidateUpdate(oldDelivery)).To(MatchError(
						"error validating clusterdelivery [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1]: cannot specify values with operator [DoesNotExist]",
					))
				})

				It("deletes without error", func() {
					Expect(delivery.ValidateDelete()).NotTo(HaveOccurred())
				})
			})
			Context("operator is In and does NOT have values", func() {
				BeforeEach(func() {
					delivery.Spec.Resources[0].TemplateRef.Options[0].Selector.MatchFields[0] = v1alpha1.FieldSelectorRequirement{
						Key:      "something",
						Operator: v1alpha1.FieldSelectorOpIn,
					}
				})

				It("on create, it rejects the Resource", func() {
					Expect(delivery.ValidateCreate()).To(MatchError(
						"error validating clusterdelivery [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1]: must specify values with operator [In]",
					))
				})

				It("on update, it rejects the Resource", func() {
					Expect(delivery.ValidateUpdate(oldDelivery)).To(MatchError(
						"error validating clusterdelivery [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1]: must specify values with operator [In]",
					))
				})

				It("deletes without error", func() {
					Expect(delivery.ValidateDelete()).NotTo(HaveOccurred())
				})
			})
			Context("operator is NotIn and does NOT have values", func() {
				BeforeEach(func() {
					delivery.Spec.Resources[0].TemplateRef.Options[0].Selector.MatchFields[0] = v1alpha1.FieldSelectorRequirement{
						Key:      "something",
						Operator: v1alpha1.FieldSelectorOpNotIn,
					}
				})

				It("on create, it rejects the Resource", func() {
					Expect(delivery.ValidateCreate()).To(MatchError(
						"error validating clusterdelivery [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1]: must specify values with operator [NotIn]",
					))
				})

				It("on update, it rejects the Resource", func() {
					Expect(delivery.ValidateUpdate(oldDelivery)).To(MatchError(
						"error validating clusterdelivery [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1]: must specify values with operator [NotIn]",
					))
				})

				It("deletes without error", func() {
					Expect(delivery.ValidateDelete()).NotTo(HaveOccurred())
				})
			})
		})

		Context("2 options with identical requirements", func() {
			Context("selectors are identical", func() {
				BeforeEach(func() {
					delivery.Spec.Resources[0].TemplateRef.Options[0].Selector = delivery.Spec.Resources[0].TemplateRef.Options[1].Selector
				})

				It("on create, it rejects the Resource", func() {
					Expect(delivery.ValidateCreate()).To(MatchError(
						"error validating clusterdelivery [responsible-ops---default-params]: error validating resource [source-provider]: duplicate selector found in options [source-1, source-2]",
					))
				})

				It("on update, it rejects the Resource", func() {
					Expect(delivery.ValidateUpdate(oldDelivery)).To(MatchError(
						"error validating clusterdelivery [responsible-ops---default-params]: error validating resource [source-provider]: duplicate selector found in options [source-1, source-2]",
					))
				})

				It("deletes without error", func() {
					Expect(delivery.ValidateDelete()).NotTo(HaveOccurred())
				})
			})
		})

		Context("option points to key that doesn't exist in spec", func() {
			BeforeEach(func() {
				delivery.Spec.Resources[0].TemplateRef.Options[0].Selector.MatchFields[0].Key = "spec.does.not.exist"
			})

			It("on create, it rejects the Resource", func() {
				Expect(delivery.ValidateCreate()).To(MatchError(
					"error validating clusterdelivery [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1]: requirement key [spec.does.not.exist] is not a valid path",
				))
			})

			It("on update, it rejects the Resource", func() {
				Expect(delivery.ValidateUpdate(oldDelivery)).To(MatchError(
					"error validating clusterdelivery [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [source-1]: requirement key [spec.does.not.exist] is not a valid path",
				))
			})

			It("deletes without error", func() {
				Expect(delivery.ValidateDelete()).NotTo(HaveOccurred())
			})
		})

		Context("option points to key that is a valid prefix into an array", func() {
			BeforeEach(func() {
				delivery.Spec.Resources[0].TemplateRef.Options[0].Selector.MatchFields[0].Key = `spec.params[?(@.name=="some-name")].value`
			})

			It("on create, it does not reject the Resource", func() {
				err := delivery.ValidateCreate()
				Expect(err).NotTo(HaveOccurred())
			})

			It("on update, it does not reject the Resource", func() {
				err := delivery.ValidateUpdate(oldDelivery)
				Expect(err).NotTo(HaveOccurred())
			})

			It("deletes without error", func() {
				Expect(delivery.ValidateDelete()).NotTo(HaveOccurred())
			})
		})

		Context("both name and options specified", func() {
			BeforeEach(func() {
				delivery.Spec.Resources[0].TemplateRef.Name = "some-name"
			})

			It("on create, it rejects the Resource", func() {
				Expect(delivery.ValidateCreate()).To(MatchError(
					"error validating clusterdelivery [responsible-ops---default-params]: error validating resource [source-provider]: exactly one of templateRef.Name or templateRef.Options must be specified, found both",
				))
			})

			It("on update, it rejects the Resource", func() {
				Expect(delivery.ValidateUpdate(oldDelivery)).To(MatchError(
					"error validating clusterdelivery [responsible-ops---default-params]: error validating resource [source-provider]: exactly one of templateRef.Name or templateRef.Options must be specified, found both",
				))
			})

			It("deletes without error", func() {
				Expect(delivery.ValidateDelete()).NotTo(HaveOccurred())
			})
		})

		Context("neither name and options are specified", func() {
			BeforeEach(func() {
				delivery.Spec.Resources[0].TemplateRef.Options = nil
			})

			It("on create, it rejects the Resource", func() {
				Expect(delivery.ValidateCreate()).To(MatchError(
					"error validating clusterdelivery [responsible-ops---default-params]: error validating resource [source-provider]: exactly one of templateRef.Name or templateRef.Options must be specified, found neither",
				))
			})

			It("on update, it rejects the Resource", func() {
				Expect(delivery.ValidateUpdate(oldDelivery)).To(MatchError(
					"error validating clusterdelivery [responsible-ops---default-params]: error validating resource [source-provider]: exactly one of templateRef.Name or templateRef.Options must be specified, found neither",
				))
			})

			It("deletes without error", func() {
				Expect(delivery.ValidateDelete()).NotTo(HaveOccurred())
			})
		})
	})

	Describe("OneOf Selector, SelectorMatchExpressions, or SelectorMatchFields", func() {
		var deliveryFactory = func(selector map[string]string, expressions []metav1.LabelSelectorRequirement, fields []v1alpha1.FieldSelectorRequirement) *v1alpha1.ClusterDelivery {
			return &v1alpha1.ClusterDelivery{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "delivery-resource",
					Namespace: "default",
				},
				Spec: v1alpha1.DeliverySpec{
					Selector: selector,
					SelectorMatchExpressions: expressions,
					SelectorMatchFields: fields,
					Resources: []v1alpha1.DeliveryResource{
						{
							Name: "source-provider",
							TemplateRef: v1alpha1.DeliveryTemplateReference{
								Kind: "ClusterSourceTemplate",
								Name: "source-template",
							},
						},
						{
							Name: "other-source-provider",
							TemplateRef: v1alpha1.DeliveryTemplateReference{
								Kind: "ClusterSourceTemplate",
								Name: "source-template",
							},
						},
					},
				},
			}

		}
		Context("No selection", func() {
			var delivery *v1alpha1.ClusterDelivery
			BeforeEach(func() {
				delivery = deliveryFactory(nil, nil, nil)
			})

			It("on create, returns an error", func() {
				Expect(delivery.ValidateCreate()).To(MatchError(
					"error validating clusterdelivery [delivery-resource]: at least one selector, selectorMatchExpression, selectorMatchField must be specified",
				))
			})

			It("on update, returns an error", func() {
				Expect(delivery.ValidateUpdate(nil)).To(MatchError(
					"error validating clusterdelivery [delivery-resource]: at least one selector, selectorMatchExpression, selectorMatchField must be specified",
				))
			})

			It("deletes without error", func() {
				Expect(delivery.ValidateDelete()).NotTo(HaveOccurred())
			})
		})
		Context("Empty selection", func() {
			var delivery *v1alpha1.ClusterDelivery
			BeforeEach(func() {
				delivery = deliveryFactory(map[string]string{}, []metav1.LabelSelectorRequirement{}, []v1alpha1.FieldSelectorRequirement{})
			})

			It("on create, returns an error", func() {
				Expect(delivery.ValidateCreate()).To(MatchError(
					"error validating clusterdelivery [delivery-resource]: at least one selector, selectorMatchExpression, selectorMatchField must be specified",
				))
			})

			It("on update, returns an error", func() {
				Expect(delivery.ValidateUpdate(nil)).To(MatchError(
					"error validating clusterdelivery [delivery-resource]: at least one selector, selectorMatchExpression, selectorMatchField must be specified",
				))
			})

			It("deletes without error", func() {
				Expect(delivery.ValidateDelete()).NotTo(HaveOccurred())
			})
		})
		Context("A Selector", func() {
			var delivery *v1alpha1.ClusterDelivery
			BeforeEach(func() {
				delivery = deliveryFactory(map[string]string{"foo":"bar"}, nil, nil)
			})

			It("creates without error", func() {
				Expect(delivery.ValidateCreate()).NotTo(HaveOccurred())
			})

			It("on update, returns an error", func() {
				Expect(delivery.ValidateUpdate(nil)).NotTo(HaveOccurred())
			})

			It("deletes without error", func() {
				Expect(delivery.ValidateDelete()).NotTo(HaveOccurred())
			})
		})
		Context("A SelectorMatchExpression", func() {
			var delivery *v1alpha1.ClusterDelivery
			BeforeEach(func() {
				delivery = deliveryFactory(nil, []metav1.LabelSelectorRequirement{{Key: "whatever", Operator: "Exists" }}, nil)
			})

			It("creates without error", func() {
				Expect(delivery.ValidateCreate()).NotTo(HaveOccurred())
			})

			It("updates without error", func() {
				Expect(delivery.ValidateUpdate(nil)).NotTo(HaveOccurred())
			})

			It("deletes without error", func() {
				Expect(delivery.ValidateDelete()).NotTo(HaveOccurred())
			})
		})
		Context("A SelectorMatchFields", func() {
			var delivery *v1alpha1.ClusterDelivery
			BeforeEach(func() {
				delivery = deliveryFactory(nil, nil, []v1alpha1.FieldSelectorRequirement{{Key: "whatever", Operator: "Exists" }})
			})

			It("creates without error", func() {
				Expect(delivery.ValidateCreate()).NotTo(HaveOccurred())
			})

			It("updates without error", func() {
				Expect(delivery.ValidateUpdate(nil)).NotTo(HaveOccurred())
			})

			It("deletes without error", func() {
				Expect(delivery.ValidateDelete()).NotTo(HaveOccurred())
			})
		})
	})

})

var _ = Describe("DeliveryTemplateReference", func() {
	It("has valid references", func() {
		Expect(v1alpha1.ValidDeliveryTemplates).To(HaveLen(4))

		Expect(v1alpha1.ValidDeliveryTemplates).To(ContainElements(
			&v1alpha1.ClusterSourceTemplate{},
			&v1alpha1.ClusterDeploymentTemplate{},
			&v1alpha1.ClusterTemplate{},
			&v1alpha1.ClusterConfigTemplate{},
		))
	})

	It("has a matching valid enum for Kind", func() {

		mrkrs, err := markersFor(
			"./cluster_delivery.go",
			"DeliveryTemplateReference",
			"Kind",
			"kubebuilder:validation:Enum",
		)

		Expect(err).NotTo(HaveOccurred())

		enumMarkers, ok := mrkrs.(crdmarkers.Enum)
		Expect(ok).To(BeTrue())

		Expect(enumMarkers).To(HaveLen(len(v1alpha1.ValidDeliveryTemplates)))
		for _, validTemplate := range v1alpha1.ValidDeliveryTemplates {
			typ := reflect.TypeOf(validTemplate)
			templateName := typ.Elem().Name()
			Expect(enumMarkers).To(ContainElement(templateName))
		}
	})
})
