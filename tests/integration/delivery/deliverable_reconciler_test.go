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
	"encoding/json"
	"fmt"
	eventsv1 "k8s.io/api/events/v1"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
	"github.com/vmware-tanzu/cartographer/tests/resources"
)

var _ = Describe("DeliverableReconciler", func() {
	var templateBytes = func() []byte {
		configMap := &corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "example-config-map",
			},
			Data: map[string]string{
				"existent-field": "a-value",
			},
		}

		templateBytes, err := json.Marshal(configMap)
		Expect(err).ToNot(HaveOccurred())
		return templateBytes
	}

	var createObject = func(ctx context.Context, objYaml, namespace string) *unstructured.Unstructured {
		obj := &unstructured.Unstructured{}
		err := yaml.Unmarshal([]byte(objYaml), obj)
		Expect(err).NotTo(HaveOccurred())
		if namespace != "" {
			obj.SetNamespace(namespace)
		}

		err = c.Create(ctx, obj, &client.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())

		return obj
	}

	var assertObjectExistsWithCorrectSpec = func(ctx context.Context, expectedObjYaml string) {
		expectedObj := &unstructured.Unstructured{}
		err := yaml.Unmarshal([]byte(expectedObjYaml), expectedObj)
		Expect(err).NotTo(HaveOccurred())
		Eventually(func() interface{} {
			obj := &unstructured.Unstructured{}
			obj.SetGroupVersionKind(expectedObj.GroupVersionKind())
			_ = c.Get(ctx, client.ObjectKey{Name: expectedObj.GetName(), Namespace: testNS}, obj)
			return obj.UnstructuredContent()["spec"]
		}).Should(Equal(expectedObj.UnstructuredContent()["spec"]), fmt.Sprintf("failed on obj name: %s", expectedObj.GetName()))
	}

	var updateObservedGenerationOfTest = func(ctx context.Context, name string) {
		testToUpdate := &resources.TestObj{}

		Eventually(func() error {
			err := c.Get(ctx, client.ObjectKey{Name: name, Namespace: testNS}, testToUpdate)
			return err
		}).Should(BeNil())

		testToUpdate.Status.ObservedGeneration = testToUpdate.Generation
		err := c.Status().Update(ctx, testToUpdate)
		Expect(err).NotTo(HaveOccurred())
	}

	var setConditionOfTest = func(ctx context.Context, name, conditionType string, conditionStatus metav1.ConditionStatus) {
		testToUpdate := &resources.TestObj{}

		Eventually(func() error {
			err := c.Get(ctx, client.ObjectKey{Name: name, Namespace: testNS}, testToUpdate)
			return err
		}).Should(BeNil())

		if testToUpdate.Status.Conditions == nil {
			testToUpdate.Status.Conditions = []metav1.Condition{}
		}

		testToUpdate.Status.Conditions = append(testToUpdate.Status.Conditions, metav1.Condition{
			Type:               conditionType,
			Status:             conditionStatus,
			LastTransitionTime: metav1.Now(),
			Reason:             "SetByTest",
			Message:            "",
		})

		err := c.Status().Update(ctx, testToUpdate)
		Expect(err).NotTo(HaveOccurred())
	}

	var newClusterDelivery = func(name string, selector map[string]string) *v1alpha1.ClusterDelivery {
		return &v1alpha1.ClusterDelivery{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: v1alpha1.DeliverySpec{
				Resources: []v1alpha1.DeliveryResource{},
				LegacySelector: v1alpha1.LegacySelector{
					Selector: selector,
				},
			},
		}
	}

	var reconcileAgain = func() {
		time.Sleep(1 * time.Second) //metav1.Time unmarshals with 1 second accuracy so this sleep avoids a race condition

		deliverable := &v1alpha1.Deliverable{}
		err := c.Get(context.Background(), client.ObjectKey{Name: "deliverable-bob", Namespace: testNS}, deliverable)
		Expect(err).NotTo(HaveOccurred())

		deliverable.Spec.Params = []v1alpha1.OwnerParam{{Name: "foo", Value: apiextensionsv1.JSON{
			Raw: []byte(`"definitelybar"`),
		}}}
		err = c.Update(context.Background(), deliverable)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() bool {
			deliverable := &v1alpha1.Deliverable{}
			err := c.Get(context.Background(), client.ObjectKey{Name: "deliverable-bob", Namespace: testNS}, deliverable)
			Expect(err).NotTo(HaveOccurred())
			return deliverable.Status.ObservedGeneration == deliverable.Generation
		}).Should(BeTrue())
	}

	var (
		ctx      context.Context
		cleanups []client.Object
	)

	BeforeEach(func() {
		ctx = context.Background()

		myServiceAccount := &corev1.ServiceAccount{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-service-account",
				Namespace: testNS,
			},
			Secrets: []corev1.ObjectReference{
				{
					Name: "my-service-account-secret",
				},
			},
		}

		cleanups = append(cleanups, myServiceAccount)
		err := c.Create(ctx, myServiceAccount, &client.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		for _, obj := range cleanups {
			_ = c.Delete(ctx, obj, &client.DeleteOptions{})
		}
	})

	Context("when the deliverable is installed", func() {
		BeforeEach(func() {
			deliverableYaml := utils.HereYaml(`
				---
				apiVersion: carto.run/v1alpha1
				kind: Deliverable
				metadata:
				  name: deliverable-bob
				  labels:
					name: webapp
				spec:
				  serviceAccountName: my-service-account
				  source:
					git:
					  url: https://github.com/ekcasey/hello-world-ops
					  ref:
						branch: prod
			`)

			deliverable := createObject(ctx, deliverableYaml, testNS)
			cleanups = append(cleanups, deliverable)
		})

		It("does not update the lastTransitionTime on subsequent reconciliation if the status does not change", func() {
			var lastConditions []metav1.Condition

			Eventually(func() bool {
				deliverable := &v1alpha1.Deliverable{}
				err := c.Get(context.Background(), client.ObjectKey{Name: "deliverable-bob", Namespace: testNS}, deliverable)
				Expect(err).NotTo(HaveOccurred())
				lastConditions = deliverable.Status.Conditions
				return deliverable.Status.ObservedGeneration == deliverable.Generation
			}).Should(BeTrue())

			reconcileAgain()

			deliverable := &v1alpha1.Deliverable{}
			err := c.Get(ctx, client.ObjectKey{Name: "deliverable-bob", Namespace: testNS}, deliverable)
			Expect(err).NotTo(HaveOccurred())

			Expect(deliverable.Status.Conditions).To(Equal(lastConditions))
		})

		Context("and reconciliation will end in an unknown status", func() {
			BeforeEach(func() {
				template := &v1alpha1.ClusterSourceTemplate{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "proper-template-bob",
					},
					Spec: v1alpha1.SourceTemplateSpec{
						TemplateSpec: v1alpha1.TemplateSpec{
							Template: &runtime.RawExtension{Raw: templateBytes()},
						},
						URLPath: "nonexistant.path",
					},
				}

				cleanups = append(cleanups, template)
				err := c.Create(ctx, template, &client.CreateOptions{})
				Expect(err).NotTo(HaveOccurred())

				delivery := newClusterDelivery("delivery-bob", map[string]string{"name": "webapp"})
				delivery.Spec.Resources = []v1alpha1.DeliveryResource{
					{
						Name: "fred-resource",
						TemplateRef: v1alpha1.DeliveryTemplateReference{
							Kind: "ClusterSourceTemplate",
							Name: "proper-template-bob",
						},
					},
				}
				cleanups = append(cleanups, delivery)

				err = c.Create(ctx, delivery, &client.CreateOptions{})
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not error if the reconciliation ends in an unknown status", func() {
				Eventually(func() []metav1.Condition {
					obj := &v1alpha1.Deliverable{}
					err := c.Get(ctx, client.ObjectKey{Name: "deliverable-bob", Namespace: testNS}, obj)
					Expect(err).NotTo(HaveOccurred())

					return obj.Status.Conditions
				}).Should(ContainElements(
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
				Expect(controllerBuffer).NotTo(gbytes.Say("Reconciler error.*unable to retrieve outputs from stamped object for resource"))
			})
		})

		Context("along with a delivery with a source, deployment, and bare template", func() {
			BeforeEach(func() {
				// Create Source
				clusterSourceTemplateYaml := utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterSourceTemplate
					metadata:
					  name: source
					spec:
					  urlPath: .spec.value.url
					  revisionPath: .spec.value.ref
					
					  template:
						apiVersion: test.run/v1alpha1
						kind: TestObj
						metadata:
						  name: $(deliverable.metadata.name)$
						spec:
						  value:
							url: $(deliverable.spec.source.git.url)$
							ref: $(deliverable.spec.source.git.ref)$
			    `)

				clusterSourceTemplate := createObject(ctx, clusterSourceTemplateYaml, "")
				cleanups = append(cleanups, clusterSourceTemplate)

				// Create Bare Template
				clusterTemplateYaml := utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterTemplate
					metadata:
					  name: git-merge
					spec:
					  template:
						apiVersion: test.run/v1alpha1
						kind: TestObj
						metadata:
						  name: $(deliverable.metadata.name)$-merge
						spec:
						  value:
							merge-key: $(source.url)$
			    `)

				clusterTemplate := createObject(ctx, clusterTemplateYaml, "")
				cleanups = append(cleanups, clusterTemplate)

				// Create Delivery
				clusterDeliveryYaml := utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterDelivery
					metadata:
					  name: delivery
					spec:
					  selector:
						name: webapp
					  resources:
						- name: config-provider
						  templateRef:
							kind: ClusterSourceTemplate
							name: source
						- name: deployer
						  templateRef:
							kind: ClusterDeploymentTemplate
							name: app-deploy
						  deployment:
							resource: config-provider
						- name: promoter
						  templateRef:
							kind: ClusterTemplate
							name: git-merge
						  sources:
							- resource: deployer
							  name: deployer
			    `)

				clusterDelivery := createObject(ctx, clusterDeliveryYaml, "")
				cleanups = append(cleanups, clusterDelivery)
			})

			Context("and the deployment has only a succeeded condition", func() {
				BeforeEach(func() {
					// Create Deployment
					clusterDeploymentTemplateYaml := utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterDeploymentTemplate
					metadata:
					  name: app-deploy
					spec:
					  observedCompletion:
						succeeded:
						  key: 'status.conditions[?(@.type=="Succeeded")].status'
						  value: "True"
					  template:
						apiVersion: test.run/v1alpha1
						kind: TestObj
						metadata:
						  name: $(deliverable.metadata.name)$-1
						spec:
						  value:
							some-key: $(deployment.url)$
			    `)

					clusterDeploymentTemplate := createObject(ctx, clusterDeploymentTemplateYaml, "")
					cleanups = append(cleanups, clusterDeploymentTemplate)
				})

				Context("and the object does not have an observedGeneration", func() {
					It("cannot find the objects stamped from templates consuming the deployment outputs", func() {
						resourceNotYetStamped := &resources.TestObj{}

						Consistently(func() error {
							err := c.Get(ctx, client.ObjectKey{Name: "deliverable-bob-merge", Namespace: testNS}, resourceNotYetStamped)
							return err
						}).Should(HaveOccurred())
					})

					It("reports the error on the deliverable Ready condition", func() {
						deliverable := &v1alpha1.Deliverable{}

						id := func(element interface{}) string {
							return element.(metav1.Condition).Type
						}

						Eventually(func() []metav1.Condition {
							_ = c.Get(ctx, client.ObjectKey{Name: "deliverable-bob", Namespace: testNS}, deliverable)
							return deliverable.Status.Conditions
						}).Should(MatchElements(id, IgnoreExtras, Elements{
							"ResourcesSubmitted": MatchFields(IgnoreExtras, Fields{
								"Status":  Equal(metav1.ConditionFalse),
								"Reason":  Equal("TemplateStampFailure"),
								"Message": ContainSubstring("resource [deployer] cannot satisfy observedCompletion without observedGeneration in object status"),
							}),
						}))
					})
				})

				Context("and the object has an observedGeneration, but the success condition is not met", func() {
					BeforeEach(func() {
						updateObservedGenerationOfTest(ctx, "deliverable-bob-1")
						setConditionOfTest(ctx, "deliverable-bob-1", "Succeeded", metav1.ConditionFalse)
					})
					It("cannot find the objects stamped from templates consuming the deployment outputs", func() {
						resourceNotYetStamped := &resources.TestObj{}

						Consistently(func() error {
							err := c.Get(ctx, client.ObjectKey{Name: "deliverable-bob-merge", Namespace: testNS}, resourceNotYetStamped)
							return err
						}).Should(HaveOccurred())
					})

					It("the deliverable has an unknown Ready condition", func() {
						deliverable := &v1alpha1.Deliverable{}

						id := func(element interface{}) string {
							return element.(metav1.Condition).Type
						}

						Eventually(func() []metav1.Condition {
							_ = c.Get(ctx, client.ObjectKey{Name: "deliverable-bob", Namespace: testNS}, deliverable)
							return deliverable.Status.Conditions
						}).Should(MatchElements(id, IgnoreExtras, Elements{
							"ResourcesSubmitted": MatchFields(IgnoreExtras, Fields{
								"Status":  Equal(metav1.ConditionUnknown),
								"Reason":  Equal("ConditionNotMet"),
								"Message": ContainSubstring("resource [deployer] condition not met: deployment success condition [status.conditions[?(@.type==\"Succeeded\")].status] was: False, expected: True"),
							}),
						}))
					})
				})

				Context("and the object has an observedGeneration, but the success key is not found", func() {
					BeforeEach(func() {
						updateObservedGenerationOfTest(ctx, "deliverable-bob-1")
					})
					It("cannot find the objects stamped from templates consuming the deployment outputs", func() {
						resourceNotYetStamped := &resources.TestObj{}

						Consistently(func() error {
							err := c.Get(ctx, client.ObjectKey{Name: "deliverable-bob-merge", Namespace: testNS}, resourceNotYetStamped)
							return err
						}).Should(HaveOccurred())
					})
					It("reports the MissingValue on the deliverable Ready condition", func() {
						deliverable := &v1alpha1.Deliverable{}

						id := func(element interface{}) string {
							return element.(metav1.Condition).Type
						}

						Eventually(func() []metav1.Condition {
							_ = c.Get(ctx, client.ObjectKey{Name: "deliverable-bob", Namespace: testNS}, deliverable)
							return deliverable.Status.Conditions
						}).Should(MatchElements(id, IgnoreExtras, Elements{
							"ResourcesSubmitted": MatchFields(IgnoreExtras, Fields{
								"Status":  Equal(metav1.ConditionUnknown),
								"Reason":  Equal("ConditionNotMet"),
								"Message": ContainSubstring(`resource [deployer] condition not met: failed to evaluate succeededCondition.Key [status.conditions[?(@.type=="Succeeded")].status]: jsonpath returned empty list: status.conditions[?(@.type=="Succeeded")].status`),
							}),
						}))
					})
				})

				Context("after the deployment stamped object has reconciled", func() {
					BeforeEach(func() {
						updateObservedGenerationOfTest(ctx, "deliverable-bob-1")
						setConditionOfTest(ctx, "deliverable-bob-1", "Succeeded", metav1.ConditionTrue)
					})

					It("can find the objects stamped from templates consuming the deployment outputs", func() {
						assertObjectExistsWithCorrectSpec(ctx, utils.HereYaml(`
							---
							apiVersion: test.run/v1alpha1
							kind: TestObj
							metadata:
							  name: deliverable-bob-merge
							spec:
							  value:
								merge-key: https://github.com/ekcasey/hello-world-ops
						`))
					})

					It("emits a StampedObjectApplied event", func() {
						Eventually(func() []eventsv1.Event {
							events := &eventsv1.EventList{}
							err := c.List(ctx, events)
							Expect(err).NotTo(HaveOccurred())
							return events.Items
						}).Should(ContainElement(MatchFields(IgnoreExtras, Fields{
							"Reason": Equal("StampedObjectApplied"),
							"Note":   Equal("Created object [testobjs.test.run/deliverable-bob-1]"),
							"Regarding": MatchFields(IgnoreExtras, Fields{
								"APIVersion": Equal("carto.run/v1alpha1"),
								"Kind":       Equal("Deliverable"),
								"Namespace":  Equal(testNS),
								"Name":       Equal("deliverable-bob"),
							}),
						})))
					})

					It("reports the deliverable Ready condition as True", func() {
						deliverable := &v1alpha1.Deliverable{}

						id := func(element interface{}) string {
							return element.(metav1.Condition).Type
						}

						Eventually(func() []metav1.Condition {
							_ = c.Get(ctx, client.ObjectKey{Name: "deliverable-bob", Namespace: testNS}, deliverable)
							return deliverable.Status.Conditions
						}).Should(MatchElements(id, IgnoreExtras, Elements{
							"Ready": MatchFields(IgnoreExtras, Fields{
								"Status": BeEquivalentTo("True"),
							}),
						}))
					})
				})

				It("finds the templated objects", func() {
					assertObjectExistsWithCorrectSpec(ctx, utils.HereYaml(`
					---
					apiVersion: test.run/v1alpha1
					kind: TestObj
					metadata:
					  name: deliverable-bob
					spec:
					  value:
						url: https://github.com/ekcasey/hello-world-ops
						ref:
						  branch: prod
			    `))

					assertObjectExistsWithCorrectSpec(ctx, utils.HereYaml(`
					---
					apiVersion: test.run/v1alpha1
					kind: TestObj
					metadata:
					  name: deliverable-bob-1
					spec:
					  value:
						some-key: https://github.com/ekcasey/hello-world-ops
			    `))
				})
			})

			Context("and the deployment has succeeded and failed conditions", func() {
				BeforeEach(func() {
					// Create Deployment
					clusterDeploymentTemplateYaml := utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterDeploymentTemplate
					metadata:
					  name: app-deploy
					spec:
					  observedCompletion:
						succeeded:
						  key: 'status.conditions[?(@.type=="Succeeded")].status'
						  value: "True"
						failed:
						  key: 'status.conditions[?(@.type=="Failed")].status'
						  value: "True"
					  template:
						apiVersion: test.run/v1alpha1
						kind: TestObj
						metadata:
						  name: $(deliverable.metadata.name)$-1
						spec:
						  value:
							some-key: $(deployment.url)$
			   		`)

					clusterDeploymentTemplate := createObject(ctx, clusterDeploymentTemplateYaml, "")
					cleanups = append(cleanups, clusterDeploymentTemplate)
				})

				Context("and the object has an observedGeneration, and both the succeeded and failed conditions are met", func() {
					BeforeEach(func() {
						updateObservedGenerationOfTest(ctx, "deliverable-bob-1")
						setConditionOfTest(ctx, "deliverable-bob-1", "Failed", metav1.ConditionTrue)
						setConditionOfTest(ctx, "deliverable-bob-1", "Succeeded", metav1.ConditionTrue)
					})

					It("cannot find the objects stamped from templates consuming the deployment outputs", func() {
						resourceNotYetStamped := &resources.TestObj{}

						Consistently(func() error {
							err := c.Get(ctx, client.ObjectKey{Name: "deliverable-bob-merge", Namespace: testNS}, resourceNotYetStamped)
							return err
						}).Should(HaveOccurred())
					})

					It("the deliverable has a failed Ready condition", func() {
						deliverable := &v1alpha1.Deliverable{}

						id := func(element interface{}) string {
							return element.(metav1.Condition).Type
						}

						Eventually(func() []metav1.Condition {
							_ = c.Get(ctx, client.ObjectKey{Name: "deliverable-bob", Namespace: testNS}, deliverable)
							return deliverable.Status.Conditions
						}).Should(MatchElements(id, IgnoreExtras, Elements{
							"ResourcesSubmitted": MatchFields(IgnoreExtras, Fields{
								"Status":  Equal(metav1.ConditionFalse),
								"Reason":  Equal("FailedConditionMet"),
								"Message": ContainSubstring("resource [deployer] failed condition met: deployment failure condition [status.conditions[?(@.type==\"Failed\")].status] was: True"),
							}),
						}))
					})
				})

				Context("and the object has an observedGeneration, and only the failed condition is met", func() {
					BeforeEach(func() {
						updateObservedGenerationOfTest(ctx, "deliverable-bob-1")
						setConditionOfTest(ctx, "deliverable-bob-1", "Failed", metav1.ConditionTrue)
						setConditionOfTest(ctx, "deliverable-bob-1", "Succeeded", metav1.ConditionFalse)
					})

					It("cannot find the objects stamped from templates consuming the deployment outputs", func() {
						resourceNotYetStamped := &resources.TestObj{}

						Consistently(func() error {
							err := c.Get(ctx, client.ObjectKey{Name: "deliverable-bob-merge", Namespace: testNS}, resourceNotYetStamped)
							return err
						}).Should(HaveOccurred())
					})

					It("the deliverable has a failed Ready condition", func() {
						deliverable := &v1alpha1.Deliverable{}

						id := func(element interface{}) string {
							return element.(metav1.Condition).Type
						}

						Eventually(func() []metav1.Condition {
							_ = c.Get(ctx, client.ObjectKey{Name: "deliverable-bob", Namespace: testNS}, deliverable)
							return deliverable.Status.Conditions
						}).Should(MatchElements(id, IgnoreExtras, Elements{
							"ResourcesSubmitted": MatchFields(IgnoreExtras, Fields{
								"Status":  Equal(metav1.ConditionFalse),
								"Reason":  Equal("FailedConditionMet"),
								"Message": ContainSubstring("resource [deployer] failed condition met: deployment failure condition [status.conditions[?(@.type==\"Failed\")].status] was: True"),
							}),
						}))
					})
				})

				Context("and the object has an observedGeneration, and only the succeeded condition is met", func() {
					BeforeEach(func() {
						updateObservedGenerationOfTest(ctx, "deliverable-bob-1")
						setConditionOfTest(ctx, "deliverable-bob-1", "Failed", metav1.ConditionFalse)
						setConditionOfTest(ctx, "deliverable-bob-1", "Succeeded", metav1.ConditionTrue)
					})

					It("can find the objects stamped from templates consuming the deployment outputs", func() {
						assertObjectExistsWithCorrectSpec(ctx, utils.HereYaml(`
					---
					apiVersion: test.run/v1alpha1
					kind: TestObj
					metadata:
					  name: deliverable-bob-merge
					spec:
					  value:
						merge-key: https://github.com/ekcasey/hello-world-ops
				`))
					})

					It("reports the deliverable Ready condition as True", func() {
						deliverable := &v1alpha1.Deliverable{}

						id := func(element interface{}) string {
							return element.(metav1.Condition).Type
						}

						Eventually(func() []metav1.Condition {
							_ = c.Get(ctx, client.ObjectKey{Name: "deliverable-bob", Namespace: testNS}, deliverable)
							return deliverable.Status.Conditions
						}).Should(MatchElements(id, IgnoreExtras, Elements{
							"Ready": MatchFields(IgnoreExtras, Fields{
								"Status": BeEquivalentTo("True"),
							}),
						}))
					})
				})
			})

		})
	})

	Context("mutable template", func() {
		BeforeEach(func() {
			templateYaml := utils.HereYaml(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterTemplate
				metadata:
				  name: my-template
				spec:
			      template:
					apiVersion: v1
					kind: ConfigMap
					metadata:
					  name: mutable-test-obj
					data:
					  foo: hard-coded-other-val
					  additionalField: $(params.health)$
			`)

			template := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, templateYaml)
			cleanups = append(cleanups, template)

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
				        name: my-template
			`)

			delivery := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, deliveryYaml)
			cleanups = append(cleanups, delivery)

			deliverable := v1alpha1.Deliverable{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Deliverable",
					APIVersion: "carto.run/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "deliverable-jaylen",
					Namespace: testNS,
					Labels: map[string]string{
						"some-key": "some-value",
					},
				},
				Spec: v1alpha1.DeliverableSpec{
					ServiceAccountName: "my-service-account",
					Params: []v1alpha1.OwnerParam{
						{
							Name:  "foo",
							Value: apiextensionsv1.JSON{Raw: []byte(`"bar"`)},
						},
						{
							Name:  "health",
							Value: apiextensionsv1.JSON{Raw: []byte(`"healthy"`)},
						},
					},
				},
			}

			cleanups = append(cleanups, &deliverable)
			err := c.Create(ctx, &deliverable, &client.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})

		It("creates the object", func() {
			configList := &corev1.ConfigMapList{}

			Eventually(func() (int, error) {
				err := c.List(ctx, configList, &client.ListOptions{Namespace: testNS})
				return len(configList.Items), err
			}).Should(Equal(1))

			Expect(configList.Items[0].Name).To(Equal("mutable-test-obj"))
			Expect(configList.Items[0].Data["foo"]).To(Equal("hard-coded-other-val"))
		})
		When("the template is changed to an immutable template", func() {
			BeforeEach(func() {
				opts := []client.ListOption{
					client.InNamespace(testNS),
				}
				configMapList := &corev1.ConfigMapList{}
				Eventually(func() (int, error) {
					err := c.List(ctx, configMapList, opts...)
					return len(configMapList.Items), err
				}).Should(Equal(1))

				immutableTemplateYaml := utils.HereYaml(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterTemplate
				metadata:
				  name: my-template
				spec:
			      lifecycle: immutable
			      template:
					apiVersion: test.run/v1alpha1
					kind: TestObj
					metadata:
					  generateName: test-resource-
					spec:
					  foo: $(params.foo)$
					  additionalField: $(params.health)$
			`)

				utils.UpdateObjectOnClusterFromYamlDefinition(ctx, c, immutableTemplateYaml, testNS, &v1alpha1.ClusterTemplate{})
			})
			It("deletes the original mutable stamped object", func() {
				type testAssertion struct {
					TestObjectsCount     int
					FoundTestObjectName  string
					FoundTestObjFieldVal string
					ConfigMapCount       int
				}

				Eventually(func() (testAssertion, error) {
					testList := &resources.TestObjList{}
					var ta testAssertion
					err := c.List(ctx, testList, &client.ListOptions{Namespace: testNS})
					ta.TestObjectsCount = len(testList.Items)
					if len(testList.Items) != 1 || err != nil {
						return ta, err
					}

					ta.FoundTestObjectName = testList.Items[0].Name
					ta.FoundTestObjFieldVal = testList.Items[0].Spec.Foo

					configMapList := &corev1.ConfigMapList{}
					err = c.List(ctx, configMapList, &client.ListOptions{Namespace: testNS})
					ta.ConfigMapCount = len(configMapList.Items)
					return ta, err
				}).Should(MatchAllFields(Fields{
					"TestObjectsCount":     Equal(1),
					"FoundTestObjectName":  ContainSubstring("test-resource-"),
					"FoundTestObjFieldVal": Equal("bar"),
					"ConfigMapCount":       Equal(0),
				}))
			})
		})
	})

	Context("delivery with immutable template", func() {
		var (
			expectedValue           string
			healthRuleSpecification string
			lifecycleSpecification  string
			immutableTemplateBase   string
			deliverable             v1alpha1.Deliverable
		)

		BeforeEach(func() {
			immutableTemplateBase = `
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterSourceTemplate
				metadata:
				  name: my-template
				spec:
				  urlPath: spec.foo
				  revisionPath: spec.foo
			      lifecycle: %s
			      template:
					apiVersion: test.run/v1alpha1
					kind: TestObj
					metadata:
					  generateName: test-resource-
					spec:
					  foo: $(deliverable.spec.source.image)$
				  %s
			`

			followOnTemplateYaml := utils.HereYaml(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterTemplate
				metadata:
				  name: follow-on-template
				spec:
			      template:
					apiVersion: v1
					kind: ConfigMap
					metadata:
					  name: follow-object
					data:
					  foo: $(sources.source.url)$
			`)

			followOnTemplate := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, followOnTemplateYaml)
			cleanups = append(cleanups, followOnTemplate)

			deliveryYaml := utils.HereYaml(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterDelivery
				metadata:
				  name: my-supply-chain
				spec:
				  selector:
					"some-key": "some-value"
			      resources:
			        - name: my-first-resource
					  templateRef:
				        kind: ClusterSourceTemplate
				        name: my-template
			        - name: follow-on-resource
					  templateRef:
				        kind: ClusterTemplate
				        name: follow-on-template
					  sources:
			            - resource: my-first-resource
			              name: source
			`)

			delivery := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, deliveryYaml)
			cleanups = append(cleanups, delivery)

			expectedValue = "some-address"

			deliverable = v1alpha1.Deliverable{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Deliverable",
					APIVersion: "carto.run/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "deliverable-jaylen",
					Namespace: testNS,
					Labels: map[string]string{
						"some-key": "some-value",
					},
				},
				Spec: v1alpha1.DeliverableSpec{
					ServiceAccountName: "my-service-account",
					Source: &v1alpha1.Source{
						Image: &expectedValue,
					},
				},
			}

			cleanups = append(cleanups, &deliverable)
			err := c.Create(ctx, &deliverable, &client.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})

		var itResultsInAHealthyDeliverable = func() {
			Eventually(func() []metav1.Condition {
				obj := &v1alpha1.Deliverable{}
				err := c.Get(ctx, client.ObjectKey{Name: "deliverable-jaylen", Namespace: testNS}, obj)
				Expect(err).NotTo(HaveOccurred())

				return obj.Status.Conditions
			}).Should(ContainElements(
				MatchFields(IgnoreExtras, Fields{
					"Type":   Equal("DeliveryReady"),
					"Reason": Equal("Ready"),
					"Status": Equal(metav1.ConditionTrue),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Type":   Equal("ResourcesSubmitted"),
					"Reason": Equal("ResourceSubmissionComplete"),
					"Status": Equal(metav1.ConditionTrue),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Type":   Equal("Ready"),
					"Reason": Equal("Ready"),
					"Status": Equal(metav1.ConditionTrue),
				}),
			))

			Consistently(func() []metav1.Condition {
				obj := &v1alpha1.Deliverable{}
				err := c.Get(ctx, client.ObjectKey{Name: "deliverable-jaylen", Namespace: testNS}, obj)
				Expect(err).NotTo(HaveOccurred())

				return obj.Status.Conditions
			}).Should(ContainElements(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal("Ready"),
				"Reason": Equal("Ready"),
				"Status": Equal(metav1.ConditionTrue),
			})))
		}

		var itStampsTheTemplatedObjectOnce = func() {
			testList := &resources.TestObjList{}

			Eventually(func() (int, error) {
				err := c.List(ctx, testList, &client.ListOptions{Namespace: testNS})
				return len(testList.Items), err
			}).Should(Equal(1))

			Consistently(func() (int, error) {
				err := c.List(ctx, testList, &client.ListOptions{Namespace: testNS})
				return len(testList.Items), err
			}, "2s").Should(Equal(1))

			Expect(testList.Items[0].Name).To(ContainSubstring("test-resource-"))
			Expect(testList.Items[0].Spec.Foo).To(Equal("some-address"))
		}

		Context("generic immutable template", func() {
			BeforeEach(func() {
				lifecycleSpecification = "immutable"
			})
			Context("without a healthRule", func() {
				BeforeEach(func() {
					healthRuleSpecification = ""
					templateYaml := utils.HereYamlF(immutableTemplateBase, lifecycleSpecification, healthRuleSpecification)
					template := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, templateYaml)
					cleanups = append(cleanups, template)
				})

				It("results in a healthy deliverable", func() {
					itResultsInAHealthyDeliverable()
				})

				It("stamps the templated object once", func() {
					itStampsTheTemplatedObjectOnce()
				})

				Context("and the deliverable is updated", func() {
					BeforeEach(func() {
						// ensure first objects have been created
						Eventually(func() (map[string]int, error) {
							testList := &resources.TestObjList{}
							configmapList := &corev1.ConfigMapList{}

							err := c.List(ctx, configmapList, &client.ListOptions{Namespace: testNS})
							if err != nil {
								return nil, err
							}
							err = c.List(ctx, testList, &client.ListOptions{Namespace: testNS})
							objectCountMap := map[string]int{
								"testObj":   len(testList.Items),
								"configMap": len(configmapList.Items),
							}
							return objectCountMap, err
						}, "2s").Should(MatchAllKeys(Keys{
							"testObj":   Equal(1),
							"configMap": Equal(1),
						}))

						image := "a-different-image"

						deliverable.Spec.Source.Image = &image

						utils.UpdateObjectOnCluster(ctx, c, &deliverable, &v1alpha1.Deliverable{})
					})

					It("creates a second object alongside the first", func() {
						testList := &resources.TestObjList{}
						configmapList := &corev1.ConfigMapList{}

						Eventually(func() (map[string]int, error) {
							err := c.List(ctx, configmapList, &client.ListOptions{Namespace: testNS})
							if err != nil {
								return nil, err
							}
							err = c.List(ctx, testList, &client.ListOptions{Namespace: testNS})
							objectCountMap := map[string]int{
								"testObj":   len(testList.Items),
								"configMap": len(configmapList.Items),
							}
							return objectCountMap, err
						}, "2s").Should(MatchAllKeys(Keys{
							"testObj":   Equal(2),
							"configMap": Equal(1),
						}))

						Consistently(func() (int, error) {
							err := c.List(ctx, testList, &client.ListOptions{Namespace: testNS})
							return len(testList.Items), err
						}, "2s").Should(Equal(2))

						Expect(testList.Items[0].Name).To(ContainSubstring("test-resource-"))
						Expect(testList.Items[1].Name).To(ContainSubstring("test-resource-"))

						id := func(element interface{}) string {
							return element.(resources.TestObj).Spec.Foo
						}
						Expect(testList.Items).To(MatchAllElements(id, Elements{
							"a-different-image": Not(BeNil()),
							"some-address":      Not(BeNil()),
						}))
					})
				})
			})

			Context("with an alwaysHealthy healthRule", func() {
				BeforeEach(func() {
					healthRuleSpecification = "healthRule:\n    alwaysHealthy: {}"
					templateYaml := utils.HereYamlF(immutableTemplateBase, lifecycleSpecification, healthRuleSpecification)
					template := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, templateYaml)
					cleanups = append(cleanups, template)
				})

				It("results in a healthy deliverable", func() {
					itResultsInAHealthyDeliverable()
				})

				It("stamps the templated object once", func() {
					itStampsTheTemplatedObjectOnce()
				})
			})

			Context("with a healthRule that must be satisfied", func() {
				Context("which is not satisfied", func() {
					BeforeEach(func() {
						healthRuleSpecification = "healthRule:\n    singleConditionType: Ready"
						templateYaml := utils.HereYamlF(immutableTemplateBase, lifecycleSpecification, healthRuleSpecification)
						template := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, templateYaml)
						cleanups = append(cleanups, template)
					})

					It("stamps the templated object once", func() {
						itStampsTheTemplatedObjectOnce()
					})
					It("results in an unhealthy deliverable", func() {
						Eventually(func() []metav1.Condition {
							obj := &v1alpha1.Deliverable{}
							err := c.Get(ctx, client.ObjectKey{Name: "deliverable-jaylen", Namespace: testNS}, obj)
							Expect(err).NotTo(HaveOccurred())

							return obj.Status.Conditions
						}).Should(ContainElements(
							MatchFields(IgnoreExtras, Fields{
								"Type":   Equal("ResourcesHealthy"),
								"Reason": Equal("HealthyConditionRule"),
								"Status": Equal(metav1.ConditionUnknown),
							}),
						))
					})

					It("prevents reading fields of the stamped object", func() {
						Eventually(func() []metav1.Condition {
							obj := &v1alpha1.Deliverable{}
							err := c.Get(ctx, client.ObjectKey{Name: "deliverable-jaylen", Namespace: testNS}, obj)
							Expect(err).NotTo(HaveOccurred())

							return obj.Status.Conditions
						}).Should(ContainElements(
							MatchFields(IgnoreExtras, Fields{
								"Type":   Equal("ResourcesSubmitted"),
								"Reason": Equal("MissingValueAtPath"),
								"Status": Equal(metav1.ConditionUnknown),
							}),
						))
					})

					When("the healthRule is subsequently satisfied", func() {
						It("results in a healthy deliverable", func() {
							// update the object
							opts := []client.ListOption{
								client.InNamespace(testNS),
							}

							testsList := &resources.TestObjList{}

							Eventually(func() ([]resources.TestObj, error) {
								err := c.List(ctx, testsList, opts...)
								return testsList.Items, err
							}).Should(HaveLen(1))

							testToUpdate := &testsList.Items[0]
							testToUpdate.Status.Conditions = []metav1.Condition{
								{
									Type:               "Ready",
									Status:             "True",
									Reason:             "Ready",
									LastTransitionTime: metav1.Now(),
								},
							}

							err := c.Status().Update(ctx, testToUpdate)
							Expect(err).NotTo(HaveOccurred())

							// assert expected state
							itResultsInAHealthyDeliverable()
						})
					})
				})
			})
		})

		Context("tekton template", func() {
			BeforeEach(func() {
				lifecycleSpecification = "tekton"
			})

			Context("without a healthRule", func() {
				BeforeEach(func() {
					healthRuleSpecification = ""
					templateYaml := utils.HereYamlF(immutableTemplateBase, lifecycleSpecification, healthRuleSpecification)
					template := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, templateYaml)
					cleanups = append(cleanups, template)
				})

				It("stamps the templated object once", func() {
					itStampsTheTemplatedObjectOnce()
				})

				When("the stamped object's succeeded condition has status == true", func() {
					It("results in a healthy deliverable", func() {
						// update the object
						opts := []client.ListOption{
							client.InNamespace(testNS),
						}

						testsList := &resources.TestObjList{}

						Eventually(func() ([]resources.TestObj, error) {
							err := c.List(ctx, testsList, opts...)
							return testsList.Items, err
						}).Should(HaveLen(1))

						testToUpdate := &testsList.Items[0]
						testToUpdate.Status.Conditions = []metav1.Condition{
							{
								Type:               "Succeeded",
								Status:             "True",
								Reason:             "SomeGoodReason",
								LastTransitionTime: metav1.Now(),
							},
						}

						err := c.Status().Update(ctx, testToUpdate)
						Expect(err).NotTo(HaveOccurred())

						// assert expected state
						itResultsInAHealthyDeliverable()
					})
				})

				When("the stamped object's succeeded condition is not yet true", func() {
					It("results in a resource with an unknown healthy status deliverable", func() {
						Eventually(func() []metav1.Condition {
							obj := &v1alpha1.Deliverable{}
							err := c.Get(ctx, client.ObjectKey{Name: "deliverable-jaylen", Namespace: testNS}, obj)
							Expect(err).NotTo(HaveOccurred())

							return obj.Status.Conditions
						}).Should(ContainElements(
							MatchFields(IgnoreExtras, Fields{
								"Type":   Equal("ResourcesHealthy"),
								"Reason": Equal("HealthyConditionRule"),
								"Status": Equal(metav1.ConditionUnknown),
							}),
						))
					})

					It("prevents reading fields of the stamped object", func() {
						Eventually(func() []metav1.Condition {
							obj := &v1alpha1.Deliverable{}
							err := c.Get(ctx, client.ObjectKey{Name: "deliverable-jaylen", Namespace: testNS}, obj)
							Expect(err).NotTo(HaveOccurred())

							return obj.Status.Conditions
						}).Should(ContainElements(
							MatchFields(IgnoreExtras, Fields{
								"Type":   Equal("ResourcesSubmitted"),
								"Reason": Equal("MissingValueAtPath"),
								"Status": Equal(metav1.ConditionUnknown),
							}),
						))
					})
				})
			})
		})
	})
})
