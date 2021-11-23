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

package delivery_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
	"github.com/vmware-tanzu/cartographer/tests/resources"
)

var _ = Describe("Deliveries", func() {
	var (
		ctx      context.Context
		delivery *unstructured.Unstructured
		cleanups []client.Object
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	AfterEach(func() {
		for _, obj := range cleanups {
			_ = c.Delete(ctx, obj, &client.DeleteOptions{})
		}
	})

	Describe("I can define a delivery with a resource", func() {
		BeforeEach(func() {
			deliveryYaml := utils.HereYaml(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterDelivery
				metadata:
				  name: my-delivery
				spec:
				  selector:
					"some-key": "some-value"
			      resources:
			        - name: my-first-resource
					  templateRef:
				        kind: ClusterSourceTemplate
				        name: my-source-template
			`)

			delivery = &unstructured.Unstructured{}
			err := yaml.Unmarshal([]byte(deliveryYaml), delivery)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("the referenced resource exists", func() {
			BeforeEach(func() {
				clusterSourceTemplateYaml := utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterSourceTemplate
					metadata:
					  name: my-source-template
					spec:
					  urlPath: .spec.value.foo
					  revisionPath: .spec.value.foo
					  template:
					    apiVersion: test.run/v1alpha1
					    kind: TestObj
					    metadata:
					      name: test-deliverable-source
					    spec:
					      value:
					        foo: bar
			    `)

				clusterSourceTemplate := &unstructured.Unstructured{}
				err := yaml.Unmarshal([]byte(clusterSourceTemplateYaml), clusterSourceTemplate)
				Expect(err).NotTo(HaveOccurred())

				err = c.Create(ctx, clusterSourceTemplate, &client.CreateOptions{})
				Expect(err).NotTo(HaveOccurred())

				cleanups = append(cleanups, clusterSourceTemplate)
			})

			It("sets the status to Ready True", func() {
				err := c.Create(ctx, delivery, &client.CreateOptions{})
				Expect(err).NotTo(HaveOccurred())

				cleanups = append(cleanups, delivery)

				Eventually(func() []metav1.Condition {
					persistedDelivery := &v1alpha1.ClusterDelivery{}
					err := c.Get(ctx, client.ObjectKey{Name: "my-delivery"}, persistedDelivery)
					Expect(err).NotTo(HaveOccurred())
					return persistedDelivery.Status.Conditions
				}, 5*time.Second).Should(
					ContainElements(
						MatchFields(
							IgnoreExtras,
							Fields{
								"Type":   Equal("Ready"),
								"Status": Equal(metav1.ConditionTrue),
								"Reason": Equal("Ready"),
							},
						),
						MatchFields(
							IgnoreExtras,
							Fields{
								"Type":   Equal("TemplatesReady"),
								"Status": Equal(metav1.ConditionTrue),
								"Reason": Equal("Ready"),
							},
						),
					),
				)
			})
		})

		Context("the referenced resource does not exist", func() {
			It("sets the status to Ready False", func() {
				err := c.Create(ctx, delivery, &client.CreateOptions{})
				Expect(err).NotTo(HaveOccurred())

				cleanups = append(cleanups, delivery)

				Eventually(func() []metav1.Condition {
					persistedDelivery := &v1alpha1.ClusterDelivery{}
					err := c.Get(ctx, client.ObjectKey{Name: "my-delivery"}, persistedDelivery)
					Expect(err).NotTo(HaveOccurred())
					return persistedDelivery.Status.Conditions
				}, 5*time.Second).Should(
					ContainElements(
						MatchFields(
							IgnoreExtras,
							Fields{
								"Type":   Equal("Ready"),
								"Status": Equal(metav1.ConditionFalse),
								"Reason": Equal("TemplatesNotFound"),
							},
						),
						MatchFields(
							IgnoreExtras,
							Fields{
								"Type":   Equal("TemplatesReady"),
								"Status": Equal(metav1.ConditionFalse),
								"Reason": Equal("TemplatesNotFound"),
							},
						),
					),
				)
			})
		})
	})

	Describe("I cannot define identical resource names", func() {
		It("rejects the delivery with an error", func() {
			deliveryYaml := utils.HereYaml(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterDelivery
				metadata:
				  name: my-delivery
				spec:
				  selector:
					foo: bar
			      resources:
			        - name: my-first-resource
					  templateRef:
						kind: ClusterSourceTemplate
						name: my-source-template
			        - name: my-first-resource
					  templateRef:
						kind: ClusterSourceTemplate
						name: my-other-template
			`)

			delivery = &unstructured.Unstructured{}
			err := yaml.Unmarshal([]byte(deliveryYaml), delivery)
			Expect(err).NotTo(HaveOccurred())

			err = c.Create(ctx, delivery, &client.CreateOptions{})
			Expect(err).To(HaveOccurred())

			cleanups = append(cleanups, delivery)

			Expect(err).To(MatchError(ContainSubstring(`spec.resources[1].name "my-first-resource" cannot appear twice`)))
		})
	})

	Describe("I can expect ClusterDelivery to not keep updating it's status", func() {

		var (
			lastConditions []metav1.Condition
		)

		BeforeEach(func() {
			deliveryYaml := utils.HereYaml(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterDelivery
				metadata:
				  name: my-delivery
				spec:
				  selector:
				    "some-key": "some-value"
				  resources:
				    - name: my-first-resource
				      templateRef:
				        kind: ClusterSourceTemplate
				        name: my-source-template
			`)

			delivery = &unstructured.Unstructured{}
			err := yaml.Unmarshal([]byte(deliveryYaml), delivery)
			Expect(err).NotTo(HaveOccurred())

			err = c.Create(ctx, delivery, &client.CreateOptions{})
			cleanups = append(cleanups, delivery)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() bool {
				persistedDelivery := &v1alpha1.ClusterDelivery{}
				err := c.Get(ctx, client.ObjectKey{Name: "my-delivery"}, persistedDelivery)
				Expect(err).NotTo(HaveOccurred())
				lastConditions = persistedDelivery.Status.Conditions

				return persistedDelivery.Status.ObservedGeneration == persistedDelivery.Generation
			}, 5*time.Second).Should(BeTrue())
		})

		It("does not update the lastTransitionTime on subsequent reconciliation if the status does not change", func() {
			time.Sleep(1 * time.Second) //metav1.Time unmarshals with 1 second accuracy so this sleep avoids a race condition

			persistedDelivery := &v1alpha1.ClusterDelivery{}
			err := c.Get(context.Background(), client.ObjectKey{Name: "my-delivery"}, persistedDelivery)
			Expect(err).NotTo(HaveOccurred())

			persistedDelivery.Spec.Selector = map[string]string{"some-key": "blah"}
			err = c.Update(context.Background(), persistedDelivery)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() int64 {
				persistedDelivery = &v1alpha1.ClusterDelivery{}
				err := c.Get(context.Background(), client.ObjectKey{Name: "my-delivery"}, persistedDelivery)
				Expect(err).NotTo(HaveOccurred())
				return persistedDelivery.Status.ObservedGeneration
			}).Should(Equal(persistedDelivery.Generation))

			err = c.Get(context.Background(), client.ObjectKey{Name: "my-delivery"}, persistedDelivery)
			Expect(err).NotTo(HaveOccurred())
			Expect(persistedDelivery.Status.Conditions).To(Equal(lastConditions))
		})
	})

	Context("when reconciling a delivery with template references", func() {
		BeforeEach(func() {
			deliveryYaml := utils.HereYaml(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterDelivery
				metadata:
				  name: my-delivery
				spec:
				  selector:
					"some-key": "some-value"
			      resources:
			        - name: my-first-resource
					  templateRef:
				        kind: ClusterTemplate
				        name: my-terminal-template
			`)

			delivery := &unstructured.Unstructured{}
			err := yaml.Unmarshal([]byte(deliveryYaml), delivery)
			Expect(err).NotTo(HaveOccurred())

			err = c.Create(ctx, delivery, &client.CreateOptions{})
			cleanups = append(cleanups, delivery)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() []metav1.Condition {
				delivery := &v1alpha1.ClusterDelivery{}
				err := c.Get(ctx, client.ObjectKey{Name: "my-delivery"}, delivery)
				Expect(err).NotTo(HaveOccurred())

				return delivery.Status.Conditions

			}, 5*time.Second).Should(
				ContainElement(
					MatchFields(IgnoreExtras, Fields{
						"Type":   Equal("TemplatesReady"),
						"Status": Equal(metav1.ConditionFalse),
						"Reason": Equal("TemplatesNotFound"),
					}),
				),
			)
		})

		Context("a change to the delivery occurs that does not cause the status to change", func() {
			var conditionsBeforeMutation []metav1.Condition

			BeforeEach(func() {
				// metav1.Time unmarshals with 1 second accuracy so this sleep ensures
				// the transition time is noticeable if it changes
				time.Sleep(1 * time.Second)

				delivery := &v1alpha1.ClusterDelivery{}
				err := c.Get(context.Background(), client.ObjectKey{Name: "my-delivery"}, delivery)
				Expect(err).NotTo(HaveOccurred())

				conditionsBeforeMutation = delivery.Status.Conditions

				delivery.Spec.Selector = map[string]string{"blah": "blah"}
				err = c.Update(context.Background(), delivery)
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() int64 {
					delivery := &v1alpha1.ClusterDelivery{}
					err := c.Get(context.Background(), client.ObjectKey{Name: "my-delivery"}, delivery)
					Expect(err).NotTo(HaveOccurred())
					return delivery.Status.ObservedGeneration
				}).Should(Equal(delivery.Generation))
			})

			It("does not update the lastTransitionTime", func() {
				delivery := &v1alpha1.ClusterDelivery{}
				err := c.Get(ctx, client.ObjectKey{Name: "my-delivery"}, delivery)
				Expect(err).NotTo(HaveOccurred())
				Expect(delivery.Status.Conditions).To(Equal(conditionsBeforeMutation))
			})
		})

		Context("a missing referenced template is created", func() {
			BeforeEach(func() {
				sourceTemplateYaml := utils.HereYaml(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterTemplate
				metadata:
				  name: my-terminal-template
				spec:
					template: {}
				`)

				sourceTemplate := &unstructured.Unstructured{}
				err := yaml.Unmarshal([]byte(sourceTemplateYaml), sourceTemplate)
				Expect(err).NotTo(HaveOccurred())

				err = c.Create(ctx, sourceTemplate, &client.CreateOptions{})
				cleanups = append(cleanups, sourceTemplate)
				Expect(err).NotTo(HaveOccurred())
			})

			It("immediately updates the delivery status", func() {
				Eventually(func() []metav1.Condition {
					delivery := &v1alpha1.ClusterDelivery{}
					err := c.Get(ctx, client.ObjectKey{Name: "my-delivery"}, delivery)
					Expect(err).NotTo(HaveOccurred())

					return delivery.Status.Conditions

				}, 3*time.Second).Should(
					ContainElements(
						MatchFields(IgnoreExtras, Fields{
							"Type":   Equal("Ready"),
							"Status": Equal(metav1.ConditionTrue),
						}),
						MatchFields(IgnoreExtras, Fields{
							"Type":   Equal("TemplatesReady"),
							"Status": Equal(metav1.ConditionTrue),
						}),
					),
				)
			})
		})
	})

	Context("a delivery with a template that has stamped a test crd", func() {
		var (
			test *resources.TestObj
		)

		BeforeEach(func() {
			templateYaml := utils.HereYaml(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterSourceTemplate
				metadata:
				  name: my-source-template
				spec:
				  urlPath: status.conditions[?(@.type=="Ready")]
				  revisionPath: status.conditions[?(@.type=="Succeeded")]
			      template:
					apiVersion: test.run/v1alpha1
					kind: TestObj
					metadata:
					  name: test-resource
					spec:
					  foo: "bar"
			`)

			template := &unstructured.Unstructured{}
			err := yaml.Unmarshal([]byte(templateYaml), template)
			Expect(err).NotTo(HaveOccurred())

			err = c.Create(ctx, template, &client.CreateOptions{})
			cleanups = append(cleanups, template)
			Expect(err).NotTo(HaveOccurred())

			deliveryYaml := utils.HereYaml(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterDelivery
				metadata:
				  name: my-delivery
				spec:
				  selector:
					"some-key": "some-value"
			      resources:
			        - name: my-first-resource
					  templateRef:
				        kind: ClusterSourceTemplate
				        name: my-source-template
			`)

			delivery := &unstructured.Unstructured{}
			err = yaml.Unmarshal([]byte(deliveryYaml), delivery)
			Expect(err).NotTo(HaveOccurred())

			err = c.Create(ctx, delivery, &client.CreateOptions{})
			cleanups = append(cleanups, delivery)
			Expect(err).NotTo(HaveOccurred())

			deliverable := &v1alpha1.Deliverable{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "deliverable-joe",
					Namespace: testNS,
					Labels: map[string]string{
						"some-key": "some-value",
					},
				},
			}

			cleanups = append(cleanups, deliverable)
			err = c.Create(ctx, deliverable, &client.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			test = &resources.TestObj{}

			Eventually(func() ([]metav1.Condition, error) {
				err := c.Get(ctx, client.ObjectKey{Name: "test-resource", Namespace: testNS}, test)
				return test.Status.Conditions, err
			}).Should(BeNil())

			Eventually(func() []metav1.Condition {
				obj := &v1alpha1.Deliverable{}
				err := c.Get(ctx, client.ObjectKey{Name: "deliverable-joe", Namespace: testNS}, obj)
				Expect(err).NotTo(HaveOccurred())

				return obj.Status.Conditions
			}, 5*time.Second).Should(ContainElements(
				MatchFields(IgnoreExtras, Fields{
					"Type":   Equal("DeliveryReady"),
					"Reason": Equal("Ready"),
					"Status": Equal(metav1.ConditionTrue),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Type":   Equal("ResourcesSubmitted"),
					"Reason": Equal("MissingValueAtPath"),
					"Status": Equal(metav1.ConditionStatus("Unknown")),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Type":   Equal("Ready"),
					"Reason": Equal("MissingValueAtPath"),
					"Status": Equal(metav1.ConditionStatus("Unknown")),
				}),
			))
		})

		Context("a stamped object has changed", func() {

			BeforeEach(func() {
				test.Status.Conditions = []metav1.Condition{
					{
						Type:               "Ready",
						Status:             "True",
						Reason:             "LifeIsGood",
						LastTransitionTime: metav1.Now(),
					},
					{
						Type:               "Succeeded",
						Status:             "True",
						Reason:             "Success",
						LastTransitionTime: metav1.Now(),
					},
				}
				err := c.Status().Update(ctx, test)
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() ([]metav1.Condition, error) {
					err := c.Get(ctx, client.ObjectKey{Name: "test-resource", Namespace: testNS}, test)
					return test.Status.Conditions, err
				}).Should(Not(BeNil()))
			})

			It("immediately reconciles", func() {
				Eventually(func() []metav1.Condition {
					obj := &v1alpha1.Deliverable{}
					err := c.Get(ctx, client.ObjectKey{Name: "deliverable-joe", Namespace: testNS}, obj)
					Expect(err).NotTo(HaveOccurred())

					return obj.Status.Conditions
				}, 5*time.Second).Should(ContainElements(
					MatchFields(IgnoreExtras, Fields{
						"Type":   Equal("DeliveryReady"),
						"Reason": Equal("Ready"),
						"Status": Equal(metav1.ConditionStatus("True")),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Type":   Equal("ResourcesSubmitted"),
						"Reason": Equal("ResourceSubmissionComplete"),
						"Status": Equal(metav1.ConditionStatus("True")),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Type":   Equal("Ready"),
						"Reason": Equal("Ready"),
						"Status": Equal(metav1.ConditionStatus("True")),
					}),
				))
			})
		})
	})
})
