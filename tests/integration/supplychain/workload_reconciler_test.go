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

package supplychain_test

import (
	"context"
	"encoding/json"
	eventsv1 "k8s.io/api/events/v1"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
	"github.com/vmware-tanzu/cartographer/tests/resources"
)

var _ = Describe("WorkloadReconciler", func() {
	var templateBytes = func() []byte {
		configMap := &corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "example-config-map",
			},
			Data: map[string]string{},
		}

		templateBytes, err := json.Marshal(configMap)
		Expect(err).ToNot(HaveOccurred())
		return templateBytes
	}

	var newClusterSupplyChain = func(name string, selector map[string]string) *v1alpha1.ClusterSupplyChain {
		return &v1alpha1.ClusterSupplyChain{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: v1alpha1.SupplyChainSpec{
				Resources: []v1alpha1.SupplyChainResource{},
				LegacySelector: v1alpha1.LegacySelector{
					Selector: selector,
				},
			},
		}
	}

	var reconcileAgain = func() {
		time.Sleep(1 * time.Second) //metav1.Time unmarshals with 1 second accuracy so this sleep avoids a race condition

		workload := &v1alpha1.Workload{}
		err := c.Get(context.Background(), client.ObjectKey{Name: "workload-bob", Namespace: testNS}, workload)
		Expect(err).NotTo(HaveOccurred())

		workload.Spec.ServiceAccountName = "my-service-account"
		workload.Spec.Params = []v1alpha1.OwnerParam{{Name: "foo", Value: apiextensionsv1.JSON{
			Raw: []byte(`"definitelybar"`),
		}}}
		err = c.Update(context.Background(), workload)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() bool {
			workload := &v1alpha1.Workload{}
			err := c.Get(context.Background(), client.ObjectKey{Name: "workload-bob", Namespace: testNS}, workload)
			Expect(err).NotTo(HaveOccurred())
			return workload.Status.ObservedGeneration == workload.Generation
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

	Context("Has the source template and workload installed", func() {
		BeforeEach(func() {
			workload := &v1alpha1.Workload{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "workload-bob",
					Namespace: testNS,
					Labels: map[string]string{
						"name": "webapp",
					},
				},
				Spec: v1alpha1.WorkloadSpec{ServiceAccountName: "my-service-account"},
			}

			cleanups = append(cleanups, workload)
			err := c.Create(ctx, workload, &client.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not update the lastTransitionTime on subsequent reconciliation if the status does not change", func() {
			var lastConditions []metav1.Condition

			Eventually(func() bool {
				workload := &v1alpha1.Workload{}
				err := c.Get(context.Background(), client.ObjectKey{Name: "workload-bob", Namespace: testNS}, workload)
				Expect(err).NotTo(HaveOccurred())
				lastConditions = workload.Status.Conditions
				return workload.Status.ObservedGeneration == workload.Generation
			}).Should(BeTrue())

			reconcileAgain()

			workload := &v1alpha1.Workload{}
			err := c.Get(ctx, client.ObjectKey{Name: "workload-bob", Namespace: testNS}, workload)
			Expect(err).NotTo(HaveOccurred())

			Expect(workload.Status.Conditions).To(Equal(lastConditions))
		})

		Context("when reconciliation will end in an unknown status", func() {
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

				supplyChain := newClusterSupplyChain("supplychain-bob", map[string]string{"name": "webapp"})
				supplyChain.Spec.Resources = []v1alpha1.SupplyChainResource{
					{
						Name: "fred-resource",
						TemplateRef: v1alpha1.SupplyChainTemplateReference{
							Kind: "ClusterSourceTemplate",
							Name: "proper-template-bob",
						},
					},
				}
				cleanups = append(cleanups, supplyChain)

				err = c.Create(ctx, supplyChain, &client.CreateOptions{})
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not error if the reconciliation ends in an unknown status", func() {
				Eventually(func() []metav1.Condition {
					obj := &v1alpha1.Workload{}
					err := c.Get(ctx, client.ObjectKey{Name: "workload-bob", Namespace: testNS}, obj)
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
	})

	Context("a supply chain with a template that has stamped a test crd", func() {
		var (
			test *resources.TestObj
		)

		BeforeEach(func() {
			templateYaml := utils.HereYaml(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterConfigTemplate
				metadata:
				  name: my-config-template
				spec:
				  configPath: status.conditions[?(@.type=="Ready")]
			      template:
					apiVersion: test.run/v1alpha1
					kind: TestObj
					metadata:
					  name: test-resource
					spec:
					  foo: "bar"
			`)

			template := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, templateYaml)
			cleanups = append(cleanups, template)

			supplyChainYaml := utils.HereYaml(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterSupplyChain
				metadata:
				  name: my-supply-chain
				spec:
				  selector:
					"some-key": "some-value"
			      resources:
			        - name: my-first-resource
					  templateRef:
				        kind: ClusterConfigTemplate
				        name: my-config-template
			`)

			supplyChain := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, supplyChainYaml)
			cleanups = append(cleanups, supplyChain)

			workload := &v1alpha1.Workload{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "workload-joe",
					Namespace: testNS,
					Labels: map[string]string{
						"some-key": "some-value",
					},
				},
				Spec: v1alpha1.WorkloadSpec{ServiceAccountName: "my-service-account"},
			}

			cleanups = append(cleanups, workload)
			err := c.Create(ctx, workload, &client.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			test = &resources.TestObj{}

			// FIXME: make this more obvious
			Eventually(func() ([]metav1.Condition, error) {
				err := c.Get(ctx, client.ObjectKey{Name: "test-resource", Namespace: testNS}, test)
				return test.Status.Conditions, err
			}).Should(BeNil())

			Eventually(func() []metav1.Condition {
				obj := &v1alpha1.Workload{}
				err := c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, obj)
				Expect(err).NotTo(HaveOccurred())

				return obj.Status.Conditions
			}).Should(ContainElements(
				MatchFields(IgnoreExtras, Fields{
					"Type":   Equal("SupplyChainReady"),
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
					obj := &v1alpha1.Workload{}
					err := c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, obj)
					Expect(err).NotTo(HaveOccurred())

					return obj.Status.Conditions
				}).Should(ContainElements(
					MatchFields(IgnoreExtras, Fields{
						"Type":   Equal("SupplyChainReady"),
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
				events := &eventsv1.EventList{}
				err := c.List(ctx, events)
				Expect(err).NotTo(HaveOccurred())
				Expect(events.Items).To(ContainElement(MatchFields(IgnoreExtras, Fields{
					"Reason": Equal("StampedObjectApplied"),
					"Note":   Equal("Created object [testobjs.test.run/test-resource]"),
					"Regarding": MatchFields(IgnoreExtras, Fields{
						"APIVersion": Equal("carto.run/v1alpha1"),
						"Kind":       Equal("Workload"),
						"Namespace":  Equal(testNS),
						"Name":       Equal("workload-joe"),
					}),
				})))
			})
		})
	})

	Context("a supply chain with an immutable template", func() {
		BeforeEach(func() {
			templateYaml := utils.HereYaml(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterConfigTemplate
				metadata:
				  name: my-config-template
				spec:
				  configPath: spec.fork
			      lifecycle: immutable
			      template:
					apiVersion: test.run/v1alpha1
					kind: TestObj
					metadata:
					  generateName: test-resource-
					spec:
					  foo: $(workload.spec.source.image)$
			`)

			template := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, templateYaml)
			cleanups = append(cleanups, template)

			supplyChainYaml := utils.HereYaml(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterSupplyChain
				metadata:
				  name: my-supply-chain
				spec:
				  selector:
					"some-key": "some-value"
			      resources:
			        - name: my-first-resource
					  templateRef:
				        kind: ClusterConfigTemplate
				        name: my-config-template
			`)

			supplyChain := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, supplyChainYaml)
			cleanups = append(cleanups, supplyChain)

			image := "some-address"

			workload := &v1alpha1.Workload{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workload",
					APIVersion: "carto.run/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "workload-joe",
					Namespace: testNS,
					Labels: map[string]string{
						"some-key": "some-value",
					},
				},
				Spec: v1alpha1.WorkloadSpec{
					ServiceAccountName: "my-service-account",
					Source: &v1alpha1.Source{
						Image: &image,
					},
				},
			}

			cleanups = append(cleanups, workload)
			err := c.Create(ctx, workload, &client.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() []metav1.Condition {
				obj := &v1alpha1.Workload{}
				err := c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, obj)
				Expect(err).NotTo(HaveOccurred())

				return obj.Status.Conditions
			}).Should(ContainElements(
				MatchFields(IgnoreExtras, Fields{
					"Type":   Equal("SupplyChainReady"),
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

		It("stamps the templated object once", func() {
			testList := &resources.TestObjList{}

			Eventually(func() (int, error) {
				err := c.List(ctx, testList, &client.ListOptions{Namespace: testNS})
				return len(testList.Items), err
			}).Should(Equal(1))

			Consistently(func() (int, error) {
				err := c.List(ctx, testList, &client.ListOptions{Namespace: testNS})
				return len(testList.Items), err
			}, "2s").Should(BeNumerically("<=", 1))

			Expect(testList.Items[0].Name).To(ContainSubstring("test-resource-"))
			Expect(testList.Items[0].Spec.Foo).To(Equal("some-address"))
		})

		Context("and the workload is updated", func() {
			BeforeEach(func() {
				// ensure first object has been created
				testList := &resources.TestObjList{}

				Eventually(func() (int, error) {
					err := c.List(ctx, testList, &client.ListOptions{Namespace: testNS})
					return len(testList.Items), err
				}, "2s").Should(Equal(1))

				obj := &v1alpha1.Workload{}

				err := c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, obj)
				Expect(err).NotTo(HaveOccurred())

				image := "a-different-image"

				newWorkload := &v1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "workload-joe",
						Namespace: testNS,
						Labels: map[string]string{
							"some-key": "some-value",
						},
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       "Workload",
						APIVersion: "carto.run/v1alpha1",
					},
					Spec: v1alpha1.WorkloadSpec{
						Source:             &v1alpha1.Source{Image: &image},
						ServiceAccountName: "my-service-account",
					},
				}

				Expect(c.Patch(ctx, newWorkload, client.MergeFromWithOptions(obj, client.MergeFromWithOptimisticLock{}))).To(Succeed())

				Eventually(func() (int, error) {
					err := c.List(ctx, testList, &client.ListOptions{Namespace: testNS})
					return len(testList.Items), err
				}).Should(Equal(2))
			})

			It("creates a second object alongside the first", func() {
				testList := &resources.TestObjList{}

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

	Context("supply chain with immutable template", func() {
		var (
			expectedValue           string
			healthRuleSpecification string
			lifecycleSpecification  string
			immutableTemplateBase   string
		)

		BeforeEach(func() {
			immutableTemplateBase = `
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterConfigTemplate
				metadata:
				  name: my-config-template
				spec:
				  configPath: spec.foo
			      lifecycle: %s
			      template:
					apiVersion: test.run/v1alpha1
					kind: TestObj
					metadata:
					  generateName: test-resource-
					spec:
					  foo: $(workload.spec.source.image)$
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
					apiVersion: test.run/v1alpha1
					kind: TestObj
					metadata:
					  name: follow-object
					spec:
					  foo: $(config)$
			`)

			followOnTemplate := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, followOnTemplateYaml)
			cleanups = append(cleanups, followOnTemplate)

			supplyChainYaml := utils.HereYaml(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterSupplyChain
				metadata:
				  name: my-supply-chain
				spec:
				  selector:
					"some-key": "some-value"
			      resources:
			        - name: my-first-resource
					  templateRef:
				        kind: ClusterConfigTemplate
				        name: my-config-template
			        - name: follow-on-resource
					  templateRef:
				        kind: ClusterTemplate
				        name: follow-on-template
					  configs:
			            - resource: my-first-resource
			              name: config
			`)

			supplyChain := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, supplyChainYaml)
			cleanups = append(cleanups, supplyChain)

			expectedValue = "some-address"

			workload := &v1alpha1.Workload{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workload",
					APIVersion: "carto.run/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "workload-joe",
					Namespace: testNS,
					Labels: map[string]string{
						"some-key": "some-value",
					},
				},
				Spec: v1alpha1.WorkloadSpec{
					ServiceAccountName: "my-service-account",
					Source: &v1alpha1.Source{
						Image: &expectedValue,
					},
				},
			}

			cleanups = append(cleanups, workload)
			err := c.Create(ctx, workload, &client.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})

		var itResultsInAHealthyWorkload = func() {
			Eventually(func() []metav1.Condition {
				obj := &v1alpha1.Workload{}
				err := c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, obj)
				Expect(err).NotTo(HaveOccurred())

				return obj.Status.Conditions
			}).Should(ContainElements(
				MatchFields(IgnoreExtras, Fields{
					"Type":   Equal("SupplyChainReady"),
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
				obj := &v1alpha1.Workload{}
				err := c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, obj)
				Expect(err).NotTo(HaveOccurred())

				return obj.Status.Conditions
			}).Should(ContainElements(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal("Ready"),
				"Reason": Equal("Ready"),
				"Status": Equal(metav1.ConditionTrue),
			})))
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

				It("results in a healthy workload", func() {
					itResultsInAHealthyWorkload()
				})
			})

			Context("with an alwaysHealthy healthRule", func() {
				BeforeEach(func() {
					healthRuleSpecification = "healthRule:\n    alwaysHealthy: {}"
					templateYaml := utils.HereYamlF(immutableTemplateBase, lifecycleSpecification, healthRuleSpecification)
					template := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, templateYaml)
					cleanups = append(cleanups, template)
				})

				It("results in a healthy workload", func() {
					itResultsInAHealthyWorkload()
				})
			})

			Context("with a healthRule that must be satisfied", func() {
				BeforeEach(func() {
					healthRuleSpecification = "healthRule:\n    singleConditionType: Ready"
					templateYaml := utils.HereYamlF(immutableTemplateBase, lifecycleSpecification, healthRuleSpecification)
					template := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, templateYaml)
					cleanups = append(cleanups, template)
				})

				Context("which is not satisfied", func() {
					It("results in an unhealthy workload", func() {
						Eventually(func() []metav1.Condition {
							obj := &v1alpha1.Workload{}
							err := c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, obj)
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
							obj := &v1alpha1.Workload{}
							err := c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, obj)
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
						It("results in a healthy workload", func() {
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
							itResultsInAHealthyWorkload()
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

				When("the stamped object's succeeded condition has status == true", func() {
					It("results in a healthy workload", func() {
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
						itResultsInAHealthyWorkload()
					})
				})

				When("the stamped object's succeeded condition is not yet true", func() {
					It("results in a resource with an unknown healthy status workload", func() {
						Eventually(func() []metav1.Condition {
							obj := &v1alpha1.Workload{}
							err := c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, obj)
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
							obj := &v1alpha1.Workload{}
							err := c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, obj)
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

	Context("a supply chain with a tekton template with no healthRule specified", func() {

	})
})
