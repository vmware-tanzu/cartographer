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
	Describe("#Create", func() {
		var delivery *v1alpha1.ClusterDelivery

		BeforeEach(func() {
			delivery = &v1alpha1.ClusterDelivery{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "delivery-resource",
					Namespace: "default",
				},
				Spec: v1alpha1.DeliverySpec{
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
			It("does not return an error", func() {
				Expect(delivery.ValidateCreate()).NotTo(HaveOccurred())
			})

		})

		Context("Duplicate resource names", func() {
			BeforeEach(func() {
				for i := range delivery.Spec.Resources {
					delivery.Spec.Resources[i].Name = "source-provider"
				}
			})

			It("returns an error", func() {
				Expect(delivery.ValidateCreate()).To(MatchError(`spec.resources[1].name "source-provider" cannot appear twice`))
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
					It("returns an error", func() {
						Expect(delivery.ValidateCreate()).To(MatchError(
							"param [some-param] is invalid: must set exactly one of value and default",
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

					It("returns an error", func() {
						Expect(delivery.ValidateCreate()).To(MatchError(
							"param [some-param] is invalid: must set exactly one of value and default",
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
					It("returns an error", func() {
						Expect(delivery.ValidateCreate()).To(MatchError(
							"resource [source-provider] is invalid: param [some-param] is invalid: must set exactly one of value and default",
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
					It("returns an error", func() {
						Expect(delivery.ValidateCreate()).To(MatchError(
							"resource [source-provider] is invalid: param [some-param] is invalid: must set exactly one of value and default",
						))
					})
				})
			})
		})
	})

	Describe("#Update", func() {
		var (
			previousDelivery *v1alpha1.ClusterDelivery
			newDelivery      *v1alpha1.ClusterDelivery
		)

		BeforeEach(func() {
			previousDelivery = &v1alpha1.ClusterDelivery{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "delivery-resource",
					Namespace: "default",
				},
				Spec: v1alpha1.DeliverySpec{
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

		Context("with a valid change", func() {
			BeforeEach(func() {
				newDelivery = previousDelivery.DeepCopy()
				newDelivery.Spec.Resources = newDelivery.Spec.Resources[:1]
			})

			It("does not return an error", func() {
				Expect(newDelivery.ValidateUpdate(previousDelivery)).NotTo(HaveOccurred())
			})
		})

		Context("Duplicate resource names", func() {
			BeforeEach(func() {
				newDelivery = previousDelivery.DeepCopy()
				newDelivery.Spec.Resources = append(newDelivery.Spec.Resources, v1alpha1.DeliveryResource{
					Name: "other-source-provider",
					TemplateRef: v1alpha1.DeliveryTemplateReference{
						Kind: "ClusterSourceTemplate",
						Name: "source-template",
					},
				})
			})

			It("returns an error", func() {
				err := newDelivery.ValidateUpdate(previousDelivery)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring(`spec.resources[2].name "other-source-provider" cannot appear twice`)))
			})

		})

	})

	Describe("#Delete", func() {

	})
})

var _ = Describe("DeliveryTemplateReference", func() {
	It("has valid references", func() {
		Expect(v1alpha1.ValidDeliveryTemplates).To(HaveLen(3))

		Expect(v1alpha1.ValidDeliveryTemplates).To(ContainElements(
			&v1alpha1.ClusterSourceTemplate{},
			&v1alpha1.ClusterDeploymentTemplate{},
			&v1alpha1.ClusterTemplate{},
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
