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
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

var _ = Describe("ClusterDeploymentTemplate", func() {
	Describe("Webhook Validation", func() {
		var (
			template *v1alpha1.ClusterDeploymentTemplate
		)

		BeforeEach(func() {
			template = &v1alpha1.ClusterDeploymentTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-template",
				},
			}
		})

		Describe("#Create and #Update", func() {
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
				Context("with conditions correct", func() {
					Context("with observedCompletion", func() {
						BeforeEach(func() {
							template.Spec.ObservedCompletion = &v1alpha1.ObservedCompletion{
								SucceededCondition: v1alpha1.Condition{
									Key:   "some-key",
									Value: "some-value",
								},
							}
						})

						It("create succeeds", func() {
							_, err := template.ValidateCreate()
							Expect(err).To(Succeed())
						})

						It("update succeeds", func() {
							_, err := template.ValidateUpdate(nil)
							Expect(err).To(Succeed())
						})
					})
					Context("with observedMatches", func() {
						BeforeEach(func() {
							template.Spec.ObservedMatches = []v1alpha1.ObservedMatch{
								{
									Input:  "some-input",
									Output: "some-output",
								},
							}
						})

						It("create succeeds", func() {
							_, err := template.ValidateCreate()
							Expect(err).To(Succeed())
						})

						It("update succeeds", func() {
							_, err := template.ValidateUpdate(nil)
							Expect(err).To(Succeed())
						})
					})
				})
				Context("with conditions incorrect", func() {
					Context("with both observedCompletion and observedMatches", func() {
						BeforeEach(func() {
							template.Spec.ObservedCompletion = &v1alpha1.ObservedCompletion{
								SucceededCondition: v1alpha1.Condition{
									Key:   "some-key",
									Value: "some-value",
								},
							}
							template.Spec.ObservedMatches = []v1alpha1.ObservedMatch{
								{
									Input:  "some-input",
									Output: "some-output",
								},
							}
						})

						It("create returns an error", func() {
							_, err := template.ValidateCreate()
							Expect(err).
								To(MatchError("invalid spec: must set exactly one of spec.ObservedMatches and spec.ObservedCompletion"))
						})

						It("update returns an error", func() {
							_, err := template.ValidateUpdate(nil)
							Expect(err).
								To(MatchError("invalid spec: must set exactly one of spec.ObservedMatches and spec.ObservedCompletion"))
						})
					})
					Context("with neither observedCompletion or observedMatches", func() {
						It("create returns an error", func() {
							_, err := template.ValidateCreate()
							Expect(err).
								To(MatchError("invalid spec: must set exactly one of spec.ObservedMatches and spec.ObservedCompletion"))
						})

						It("update returns an error", func() {
							_, err := template.ValidateUpdate(nil)
							Expect(err).
								To(MatchError("invalid spec: must set exactly one of spec.ObservedMatches and spec.ObservedCompletion"))
						})
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

					template.Spec.ObservedMatches = []v1alpha1.ObservedMatch{
						{
							Input:  "some-input",
							Output: "some-output",
						},
					}
				})

				It("create returns an error", func() {
					_, err := template.ValidateCreate()
					Expect(err).
						To(MatchError("invalid template: template should not set metadata.namespace on the child object"))
				})

				It("update returns an error", func() {
					_, err := template.ValidateUpdate(nil) 
					Expect(err).
						To(MatchError("invalid template: template should not set metadata.namespace on the child object"))
				})
			})
		})

		Describe("#Delete", func() {
			Context("Any template", func() {
				var anyTemplate *v1alpha1.ClusterDeploymentTemplate
				It("always succeeds", func() {
					_, err := anyTemplate.ValidateDelete()
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
})
