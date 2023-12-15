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
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

var _ = Describe("ClusterTemplate", func() {
	Describe("Webhook Validation", func() {
		var (
			template *v1alpha1.ClusterTemplate
		)

		BeforeEach(func() {
			template = &v1alpha1.ClusterTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-template",
					Namespace: "default",
				},
			}
		})

		Describe("#Create", func() {
			Context("template is well formed", func() {
				BeforeEach(func() {
					raw, err := json.Marshal(&ArbitraryObject{
						TypeMeta: metav1.TypeMeta{
							Kind:       "some-kind",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "some-name",
						},
						Spec: ArbitrarySpec{
							SomeKey: "some-val",
						},
					})
					Expect(err).NotTo(HaveOccurred())
					template.Spec.Template = &runtime.RawExtension{Raw: raw}
				})

				It("succeeds", func() {
					_, err := template.ValidateCreate()
					Expect(err).To(Succeed())
				})
			})

			Describe("Health Rule validation", func() {
				BeforeEach(func() {
					raw, err := json.Marshal(&ArbitraryObject{
						TypeMeta: metav1.TypeMeta{
							Kind:       "some-kind",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "some-name",
						},
						Spec: ArbitrarySpec{
							SomeKey: "some-val",
						},
					})
					Expect(err).NotTo(HaveOccurred())
					template.Spec.Template = &runtime.RawExtension{Raw: raw}
					template.Spec.HealthRule = &v1alpha1.HealthRule{}
				})

				It("returns an error if no types are specified", func() {
					_, err := template.ValidateCreate()
					Expect(err).
						To(MatchError("invalid health rule: must specify one of alwaysHealthy, singleConditionType or multiMatch, found neither"))
				})

				DescribeTable("returns an error if multiple types are specified",
					func(alwaysHealthy, singleConditionType, multiMatchRule bool) {
						if alwaysHealthy {
							template.Spec.HealthRule.AlwaysHealthy = &runtime.RawExtension{Raw: []byte{}}
						} else {
							template.Spec.HealthRule.AlwaysHealthy = nil
						}
						if singleConditionType {
							template.Spec.HealthRule.SingleConditionType = "CantHaveThisToo"
						} else {
							template.Spec.HealthRule.SingleConditionType = ""
						}
						if multiMatchRule {
							template.Spec.HealthRule.MultiMatch = &v1alpha1.MultiMatchHealthRule{
								Healthy: v1alpha1.HealthMatchRule{
									MatchConditions: []v1alpha1.ConditionRequirement{
										{
											Type:   "BuildStatus",
											Status: "Succeeded",
										},
									},
								},
								Unhealthy: v1alpha1.HealthMatchRule{
									MatchFields: []v1alpha1.HealthMatchFieldSelectorRequirement{
										{
											FieldSelectorRequirement: v1alpha1.FieldSelectorRequirement{
												Key:      ".status.greenlight",
												Operator: "Exists",
											},
											MessagePath: ".somewhere.out.there",
										},
									},
								},
							}
						} else {
							template.Spec.HealthRule.MultiMatch = nil
						}
						_, err := template.ValidateCreate()
						Expect(err).
							To(MatchError("invalid health rule: must specify one of alwaysHealthy, singleConditionType or multiMatch, found multiple"))

					},
					Entry("All types", true, true, true),
					Entry("AlwaysHealthy and SingleConditionType", true, true, false),
					Entry("AlwaysHealthy and MultiMatch", true, false, true),
					Entry("SingleConditionType and MultiMatch", false, true, true),
				)

				It("succeeds when AlwaysHealthy is set", func() {
					template.Spec.HealthRule = &v1alpha1.HealthRule{
						AlwaysHealthy: &runtime.RawExtension{Raw: []byte{}},
					}
					_, err := template.ValidateCreate()
					Expect(err).To(Succeed())
				})

				It("succeeds when SingleConditionType is set", func() {
					template.Spec.HealthRule = &v1alpha1.HealthRule{
						SingleConditionType: "ThisWorksAlone",
					}
					_, err := template.ValidateCreate()
					Expect(err).To(Succeed())
				})

				It("succeeds when MultiMatch is set", func() {
					template.Spec.HealthRule = &v1alpha1.HealthRule{
						MultiMatch: &v1alpha1.MultiMatchHealthRule{
							Healthy: v1alpha1.HealthMatchRule{
								MatchConditions: []v1alpha1.ConditionRequirement{
									{
										Type:   "BuildStatus",
										Status: "Succeeded",
									},
								},
							},
							Unhealthy: v1alpha1.HealthMatchRule{
								MatchFields: []v1alpha1.HealthMatchFieldSelectorRequirement{
									{
										FieldSelectorRequirement: v1alpha1.FieldSelectorRequirement{
											Key:      ".status.greenlight",
											Operator: "Exists",
										},
										MessagePath: ".somewhere.out.there",
									},
								},
							},
						},
					}
					_, err := template.ValidateCreate()
					Expect(err).To(Succeed())
				})

				Context("Invalid MultiMatch rules", func() {
					BeforeEach(func() {
						template.Spec.HealthRule = &v1alpha1.HealthRule{
							MultiMatch: &v1alpha1.MultiMatchHealthRule{
								Healthy: v1alpha1.HealthMatchRule{
									MatchConditions: []v1alpha1.ConditionRequirement{
										{
											Type:   "BuildStatus",
											Status: "Succeeded",
										},
									},
								},
								Unhealthy: v1alpha1.HealthMatchRule{
									MatchFields: []v1alpha1.HealthMatchFieldSelectorRequirement{
										{
											FieldSelectorRequirement: v1alpha1.FieldSelectorRequirement{
												Key:      ".status.greenlight",
												Operator: "Exists",
											},
											MessagePath: ".somewhere.out.there",
										},
									},
								},
							},
						}
					})

					It("returns an error if Unhealthy has no condition or field requirements", func() {
						template.Spec.HealthRule.MultiMatch.Unhealthy = v1alpha1.HealthMatchRule{
							MatchFields:     []v1alpha1.HealthMatchFieldSelectorRequirement{},
							MatchConditions: []v1alpha1.ConditionRequirement{},
						}
						_, err := template.ValidateCreate()
						Expect(err).
							To(MatchError("invalid multi match health rule: unhealthy rule has no matchFields or matchConditions"))
					})

					It("returns an error if Healthy has no condition or field requirements", func() {
						template.Spec.HealthRule.MultiMatch.Healthy = v1alpha1.HealthMatchRule{
							MatchFields:     []v1alpha1.HealthMatchFieldSelectorRequirement{},
							MatchConditions: []v1alpha1.ConditionRequirement{},
						}
						_, err := template.ValidateCreate()
						Expect(err).
							To(MatchError("invalid multi match health rule: healthy rule has no matchFields or matchConditions"))
					})
				})
			})

			Context("template sets object namespace", func() {
				BeforeEach(func() {
					raw, err := json.Marshal(&ArbitraryObject{
						TypeMeta: metav1.TypeMeta{
							Kind:       "some-kind",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "some-name",
							Namespace: "some-namespace",
						},
						Spec: ArbitrarySpec{
							SomeKey: "some-val",
						},
					})
					Expect(err).NotTo(HaveOccurred())
					template.Spec.Template = &runtime.RawExtension{Raw: raw}
				})

				It("returns an error", func() {
					_, err := template.ValidateCreate()
					Expect(err).
						To(MatchError("invalid template: template should not set metadata.namespace on the child object"))
				})
			})

			Context("template missing", func() {
				It("succeeds", func() {
					_, err := template.ValidateCreate()
					Expect(err).
						To(MatchError("invalid template: must specify one of template or ytt, found neither"))
				})
			})

			Context("template over specified", func() {
				BeforeEach(func() {
					raw, err := json.Marshal(&ArbitraryObject{
						TypeMeta: metav1.TypeMeta{
							Kind:       "some-kind",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "some-name",
						},
						Spec: ArbitrarySpec{
							SomeKey: "some-val",
						},
					})
					Expect(err).NotTo(HaveOccurred())
					template.Spec.Template = &runtime.RawExtension{Raw: raw}
					template.Spec.Ytt = `hello: #@ data.values.hello`
				})

				It("succeeds", func() {
					_, err := template.ValidateCreate()
					Expect(err).
						To(MatchError("invalid template: must specify one of template or ytt, found both"))
				})
			})

			Context("lifecycle", func() {
				BeforeEach(func() {
					raw, err := json.Marshal(&ArbitraryObject{
						TypeMeta: metav1.TypeMeta{
							Kind:       "some-kind",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "some-name",
						},
						Spec: ArbitrarySpec{
							SomeKey: "some-val",
						},
					})
					Expect(err).NotTo(HaveOccurred())
					template.Spec.Template = &runtime.RawExtension{Raw: raw}
				})
				Context("is immutable", func() {
					BeforeEach(func() {
						template.Spec.Lifecycle = "immutable"
					})
					Context("a retention policy is set", func() {
						BeforeEach(func() {
							template.Spec.RetentionPolicy = &v1alpha1.RetentionPolicy{
								MaxFailedRuns:     1,
								MaxSuccessfulRuns: 1,
							}
						})
						It("does not return an error", func() {
							_, err := template.ValidateCreate()
							Expect(err).To(Succeed())
						})
					})

					Context("a retention policy is not set", func() {
						It("does not return an error", func() {
							_, err := template.ValidateCreate()
							Expect(err).To(Succeed())
						})
					})
				})

				Context("is tekton", func() {
					BeforeEach(func() {
						template.Spec.Lifecycle = "tekton"
					})
					Context("a retention policy is set", func() {
						BeforeEach(func() {
							template.Spec.RetentionPolicy = &v1alpha1.RetentionPolicy{
								MaxFailedRuns:     1,
								MaxSuccessfulRuns: 1,
							}
						})
						It("does not return an error", func() {
							_, err := template.ValidateCreate()
							Expect(err).To(Succeed())
						})
					})

					Context("a retention policy is not set", func() {
						It("does not return an error", func() {
							_, err := template.ValidateCreate()
							Expect(err).To(Succeed())
						})
					})
				})

				Context("is mutable", func() {
					BeforeEach(func() {
						template.Spec.Lifecycle = "mutable"
					})
					Context("a retention policy is set", func() {
						BeforeEach(func() {
							template.Spec.RetentionPolicy = &v1alpha1.RetentionPolicy{}
						})
						It("returns a helpful error", func() {
							_, err := template.ValidateCreate()
							Expect(err).To(HaveOccurred())
							Expect(err).To(MatchError("invalid template: if lifecycle is mutable, no retention policy may be set"))
						})
					})

					Context("a retention policy is not set", func() {
						It("does not return an error", func() {
							_, err := template.ValidateCreate()
							Expect(err).To(Succeed())
						})
					})
				})
			})
		})

		Describe("#Update", func() {
			Context("template is well formed", func() {
				BeforeEach(func() {
					raw, err := json.Marshal(&ArbitraryObject{
						TypeMeta: metav1.TypeMeta{
							Kind:       "some-kind",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "some-name",
						},
						Spec: ArbitrarySpec{
							SomeKey: "some-val",
						},
					})
					Expect(err).NotTo(HaveOccurred())
					template.Spec.Template = &runtime.RawExtension{Raw: raw}
				})

				It("succeeds", func() {
					_, err := template.ValidateUpdate(nil)
					Expect(err).To(Succeed())
				})
			})

			Describe("Health Rule validation", func() {
				BeforeEach(func() {
					raw, err := json.Marshal(&ArbitraryObject{
						TypeMeta: metav1.TypeMeta{
							Kind:       "some-kind",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "some-name",
						},
						Spec: ArbitrarySpec{
							SomeKey: "some-val",
						},
					})
					Expect(err).NotTo(HaveOccurred())
					template.Spec.Template = &runtime.RawExtension{Raw: raw}
					template.Spec.HealthRule = &v1alpha1.HealthRule{}
				})

				It("returns an error if no types are specified", func() {
					_, err := template.ValidateUpdate(nil)
					Expect(err).
						To(MatchError("invalid health rule: must specify one of alwaysHealthy, singleConditionType or multiMatch, found neither"))
				})

				DescribeTable("returns an error if multiple types are specified",
					func(alwaysHealthy, singleConditionType, multiMatchRule bool) {
						if alwaysHealthy {
							template.Spec.HealthRule.AlwaysHealthy = &runtime.RawExtension{Raw: []byte{}}
						} else {
							template.Spec.HealthRule.AlwaysHealthy = nil
						}
						if singleConditionType {
							template.Spec.HealthRule.SingleConditionType = "CantHaveThisToo"
						} else {
							template.Spec.HealthRule.SingleConditionType = ""
						}
						if multiMatchRule {
							template.Spec.HealthRule.MultiMatch = &v1alpha1.MultiMatchHealthRule{
								Healthy: v1alpha1.HealthMatchRule{
									MatchConditions: []v1alpha1.ConditionRequirement{
										{
											Type:   "BuildStatus",
											Status: "Succeeded",
										},
									},
								},
								Unhealthy: v1alpha1.HealthMatchRule{
									MatchFields: []v1alpha1.HealthMatchFieldSelectorRequirement{
										{
											FieldSelectorRequirement: v1alpha1.FieldSelectorRequirement{
												Key:      ".status.greenlight",
												Operator: "Exists",
											},
											MessagePath: ".somewhere.out.there",
										},
									},
								},
							}
						} else {
							template.Spec.HealthRule.MultiMatch = nil
						}
						_, err := template.ValidateUpdate(nil)
						Expect(err).
							To(MatchError("invalid health rule: must specify one of alwaysHealthy, singleConditionType or multiMatch, found multiple"))

					},
					Entry("All types", true, true, true),
					Entry("AlwaysHealthy and SingleConditionType", true, true, false),
					Entry("AlwaysHealthy and MultiMatch", true, false, true),
					Entry("SingleConditionType and MultiMatch", false, true, true),
				)

				It("succeeds when AlwaysHealthy is set", func() {
					template.Spec.HealthRule = &v1alpha1.HealthRule{
						AlwaysHealthy: &runtime.RawExtension{Raw: []byte{}},
					}
					_, err := template.ValidateUpdate(nil)
					Expect(err).To(Succeed())
				})

				It("succeeds when SingleConditionType is set", func() {
					template.Spec.HealthRule = &v1alpha1.HealthRule{
						SingleConditionType: "ThisWorksAlone",
					}
					_, err := template.ValidateUpdate(nil)
					Expect(err).To(Succeed())
				})

				It("succeeds when MultiMatch is set", func() {
					template.Spec.HealthRule = &v1alpha1.HealthRule{
						MultiMatch: &v1alpha1.MultiMatchHealthRule{
							Healthy: v1alpha1.HealthMatchRule{
								MatchConditions: []v1alpha1.ConditionRequirement{
									{
										Type:   "BuildStatus",
										Status: "Succeeded",
									},
								},
							},
							Unhealthy: v1alpha1.HealthMatchRule{
								MatchFields: []v1alpha1.HealthMatchFieldSelectorRequirement{
									{
										FieldSelectorRequirement: v1alpha1.FieldSelectorRequirement{
											Key:      ".status.greenlight",
											Operator: "Exists",
										},
										MessagePath: ".somewhere.out.there",
									},
								},
							},
						},
					}
					_, err := template.ValidateUpdate(nil)
					Expect(err).To(Succeed())
				})

				Context("Invalid MultiMatch rules", func() {
					BeforeEach(func() {
						template.Spec.HealthRule = &v1alpha1.HealthRule{
							MultiMatch: &v1alpha1.MultiMatchHealthRule{
								Healthy: v1alpha1.HealthMatchRule{
									MatchConditions: []v1alpha1.ConditionRequirement{
										{
											Type:   "BuildStatus",
											Status: "Succeeded",
										},
									},
								},
								Unhealthy: v1alpha1.HealthMatchRule{
									MatchFields: []v1alpha1.HealthMatchFieldSelectorRequirement{
										{
											FieldSelectorRequirement: v1alpha1.FieldSelectorRequirement{
												Key:      ".status.greenlight",
												Operator: "Exists",
											},
											MessagePath: ".somewhere.out.there",
										},
									},
								},
							},
						}
					})

					It("returns an error if Unhealthy has no condition or field requirements", func() {
						template.Spec.HealthRule.MultiMatch.Unhealthy = v1alpha1.HealthMatchRule{
							MatchFields:     []v1alpha1.HealthMatchFieldSelectorRequirement{},
							MatchConditions: []v1alpha1.ConditionRequirement{},
						}
						_, err := template.ValidateUpdate(nil)
						Expect(err).
							To(MatchError("invalid multi match health rule: unhealthy rule has no matchFields or matchConditions"))
					})

					It("returns an error if Healthy has no condition or field requirements", func() {
						template.Spec.HealthRule.MultiMatch.Healthy = v1alpha1.HealthMatchRule{
							MatchFields:     []v1alpha1.HealthMatchFieldSelectorRequirement{},
							MatchConditions: []v1alpha1.ConditionRequirement{},
						}
						_, err := template.ValidateUpdate(nil)
						Expect(err).
							To(MatchError("invalid multi match health rule: healthy rule has no matchFields or matchConditions"))
					})
				})
			})

			Context("template sets object namespace", func() {
				BeforeEach(func() {
					raw, err := json.Marshal(&ArbitraryObject{
						TypeMeta: metav1.TypeMeta{
							Kind:       "some-kind",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "some-name",
							Namespace: "some-namespace",
						},
						Spec: ArbitrarySpec{
							SomeKey: "some-val",
						},
					})
					Expect(err).NotTo(HaveOccurred())
					template.Spec.Template = &runtime.RawExtension{Raw: raw}
				})

				It("returns an error", func() {
					_, err := template.ValidateUpdate(nil)
					Expect(err).
						To(MatchError("invalid template: template should not set metadata.namespace on the child object"))
				})
			})

			Context("template missing", func() {
				It("succeeds", func() {
					_, err := template.ValidateUpdate(nil)
					Expect(err).
						To(MatchError("invalid template: must specify one of template or ytt, found neither"))
				})
			})

			Context("template over specified", func() {
				BeforeEach(func() {
					raw, err := json.Marshal(&ArbitraryObject{
						TypeMeta: metav1.TypeMeta{
							Kind:       "some-kind",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "some-name",
						},
						Spec: ArbitrarySpec{
							SomeKey: "some-val",
						},
					})
					Expect(err).NotTo(HaveOccurred())
					template.Spec.Template = &runtime.RawExtension{Raw: raw}
					template.Spec.Ytt = `hello: #@ data.values.hello`
				})

				It("succeeds", func() {
					_, err := template.ValidateUpdate(nil)
					Expect(err).
						To(MatchError("invalid template: must specify one of template or ytt, found both"))
				})
			})
		})

		Context("#Delete", func() {
			Context("Any template", func() {
				var anyTemplate *v1alpha1.ClusterTemplate
				It("always succeeds", func() {
					_, err := anyTemplate.ValidateDelete()
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})

	It("has a matching valid enum for lifecycle", func() {
		expectedEnumVals := []string{"mutable", "immutable", "tekton"}

		mrkrs, err := markersFor(
			"cluster_template.go",
			"./...",
			"TemplateSpec",
			"Lifecycle",
			"kubebuilder:validation:Enum",
		)

		Expect(err).NotTo(HaveOccurred())

		enumMarkers, ok := mrkrs.(crdmarkers.Enum)
		Expect(ok).To(BeTrue())

		Expect(enumMarkers).To(HaveLen(len(expectedEnumVals)))
		for _, expectedEnumVal := range expectedEnumVals {
			Expect(enumMarkers).To(ContainElement(expectedEnumVal))
		}
	})
})
