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
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
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
			obj := &v1alpha1.Workload{}
			err = c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, obj)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(obj.Status.Resources)).To(Equal(1))
			Expect(obj.Status.Resources[0].RealizedResource.Name).To(Equal("my-first-resource"))
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

		Context("the workload is changed to a new supply chain with resources in a new order", func() {
			BeforeEach(func() {
				simpleTemplateYaml := utils.HereYaml(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterTemplate
				metadata:
				  name: simple-template
				spec:
			      template:
					apiVersion: test.run/v1alpha1
					kind: TestObj
					metadata:
					  name: another-resource-$(params.suffix)$
					spec:
					  baz: "qux"
			`)
				template := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, simpleTemplateYaml)
				cleanups = append(cleanups, template)

				supplyChainYaml := utils.HereYaml(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterSupplyChain
				metadata:
				  name: new-supply-chain
				spec:
				  selector:
					"another-key": "another-value"
			      resources:
			        - name: my-zeroth-resource
					  templateRef:
				        kind: ClusterTemplate
				        name: simple-template
					  params:
					    - name: suffix
					      default: 0
			        - name: my-first-resource
					  templateRef:
				        kind: ClusterConfigTemplate
				        name: my-config-template
			        - name: another-resource
					  templateRef:
				        kind: ClusterTemplate
				        name: simple-template
					  params:
					    - name: suffix
					      default: another
			`)
				supplyChain := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, supplyChainYaml)
				cleanups = append(cleanups, supplyChain)

				workload := &v1alpha1.Workload{}
				err := c.Get(context.Background(), client.ObjectKey{Name: "workload-joe", Namespace: testNS}, workload)
				Expect(err).NotTo(HaveOccurred())

				workload.ObjectMeta.Labels = map[string]string{
					"another-key": "another-value",
				}

				err = c.Update(context.Background(), workload)
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() int {
					obj := &v1alpha1.Workload{}
					err := c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, obj)
					Expect(err).NotTo(HaveOccurred())

					return len(obj.Status.Resources)
				}).Should(Equal(3))
			})

			It("it updates to use new supply-chain", func() {
				obj := &v1alpha1.Workload{}
				err := c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, obj)
				Expect(err).NotTo(HaveOccurred())

				Expect(obj.Status.Resources[0].RealizedResource.Name).To(Equal("my-zeroth-resource"))
				Expect(obj.Status.Resources[1].RealizedResource.Name).To(Equal("my-first-resource"))
				Expect(obj.Status.Resources[2].RealizedResource.Name).To(Equal("another-resource"))
			})

			Context("supply chain deletes earlier resources", func() {
				BeforeEach(func() {
					supplyChainYaml := utils.HereYaml(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterSupplyChain
				metadata:
				  name: smaller-supply-chain
				spec:
				  selector:
					"what-kept": "later"
			      resources:
			        - name: another-resource
					  templateRef:
				        kind: ClusterTemplate
				        name: simple-template
					  params:
					    - name: suffix
					      default: another
			`)
					supplyChain := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, supplyChainYaml)
					cleanups = append(cleanups, supplyChain)

					workload := &v1alpha1.Workload{}
					err := c.Get(context.Background(), client.ObjectKey{Name: "workload-joe", Namespace: testNS}, workload)
					Expect(err).NotTo(HaveOccurred())

					workload.ObjectMeta.Labels = map[string]string{
						"what-kept": "later",
					}

					err = c.Update(context.Background(), workload)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns only the existing resource in the workload status", func() {
					Eventually(func() int {
						obj := &v1alpha1.Workload{}
						err := c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, obj)
						Expect(err).NotTo(HaveOccurred())

						return len(obj.Status.Resources)
					}).Should(Equal(1))

					obj := &v1alpha1.Workload{}
					err := c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, obj)
					Expect(err).NotTo(HaveOccurred())

					Expect(obj.Status.Resources[0].RealizedResource.Name).To(Equal("another-resource"))
				})
			})
			Context("supply chain deletes later resources", func() {
				BeforeEach(func() {
					supplyChainYaml := utils.HereYaml(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterSupplyChain
				metadata:
				  name: smaller-supply-chain
				spec:
				  selector:
					"what-kept": "earlier"
			      resources:
			        - name: my-zeroth-resource
					  templateRef:
				        kind: ClusterTemplate
				        name: simple-template
					  params:
					    - name: suffix
					      default: 0
			`)
					supplyChain := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, supplyChainYaml)
					cleanups = append(cleanups, supplyChain)

					workload := &v1alpha1.Workload{}
					err := c.Get(context.Background(), client.ObjectKey{Name: "workload-joe", Namespace: testNS}, workload)
					Expect(err).NotTo(HaveOccurred())

					workload.ObjectMeta.Labels = map[string]string{
						"what-kept": "earlier",
					}

					err = c.Update(context.Background(), workload)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns only the existing resource in the workload status", func() {
					Eventually(func() int {
						obj := &v1alpha1.Workload{}
						err := c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, obj)
						Expect(err).NotTo(HaveOccurred())

						return len(obj.Status.Resources)
					}).Should(Equal(1))

					obj := &v1alpha1.Workload{}
					err := c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, obj)
					Expect(err).NotTo(HaveOccurred())

					Expect(obj.Status.Resources[0].RealizedResource.Name).To(Equal("my-zeroth-resource"))
				})
			})
		})
	})

	Context("supply chain with immutable template", func() {
		var (
			expectedValue             string
			healthRuleSpecification   string
			lifecycleSpecification    string
			immutableTemplateBase     string
			workload                  v1alpha1.Workload
			configPathThatWillBeFound string
		)

		BeforeEach(func() {
			immutableTemplateBase = `
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterConfigTemplate
				metadata:
				  name: my-config-template
				spec:
				  configPath: %s
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

			configPathThatWillBeFound = "spec.foo"

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

			workload = v1alpha1.Workload{
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

			cleanups = append(cleanups, &workload)
			err := c.Create(ctx, &workload, &client.CreateOptions{})
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
					templateYaml := utils.HereYamlF(immutableTemplateBase, configPathThatWillBeFound, lifecycleSpecification, healthRuleSpecification)
					template := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, templateYaml)
					cleanups = append(cleanups, template)
				})

				It("results in a healthy workload and propagates outputs", func() {
					itResultsInAHealthyWorkload()

					workload := &v1alpha1.Workload{}
					err := c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, workload)
					Expect(err).NotTo(HaveOccurred())

					Expect(workload.Status.Resources[0].Outputs).To(HaveLen(1))
					Expect(workload.Status.Resources[0].Outputs[0]).To(MatchFields(IgnoreExtras, Fields{
						"Name":    Equal("config"),
						"Preview": Equal("some-address\n"),
					}))
				})

				It("stamps the templated object once", func() {
					itStampsTheTemplatedObjectOnce()
				})

				Context("and the workload is updated", func() {
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

						workload.Spec.Source.Image = &image

						utils.UpdateObjectOnCluster(ctx, c, &workload, &v1alpha1.Workload{})
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
					templateYaml := utils.HereYamlF(immutableTemplateBase, configPathThatWillBeFound, lifecycleSpecification, healthRuleSpecification)
					template := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, templateYaml)
					cleanups = append(cleanups, template)
				})

				It("results in a healthy workload and propagates outputs", func() {
					itResultsInAHealthyWorkload()

					workload := &v1alpha1.Workload{}
					err := c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, workload)
					Expect(err).NotTo(HaveOccurred())

					Expect(workload.Status.Resources[0].Outputs).To(HaveLen(1))
					Expect(workload.Status.Resources[0].Outputs[0]).To(MatchFields(IgnoreExtras, Fields{
						"Name":    Equal("config"),
						"Preview": Equal("some-address\n"),
					}))
				})

				It("stamps the templated object once", func() {
					itStampsTheTemplatedObjectOnce()
				})
			})

			Context("with a healthRule that must be satisfied", func() {
				Context("which is not satisfied", func() {
					BeforeEach(func() {
						healthRuleSpecification = "healthRule:\n    singleConditionType: Ready"
						templateYaml := utils.HereYamlF(immutableTemplateBase, configPathThatWillBeFound, lifecycleSpecification, healthRuleSpecification)
						template := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, templateYaml)
						cleanups = append(cleanups, template)
					})

					It("stamps the templated object once", func() {
						itStampsTheTemplatedObjectOnce()
					})
					It("results in proper status", func() {
						createdObject := getTestObjAtIndex(ctx, testNS, 0, 1)

						Eventually(func() []metav1.Condition {
							workload := &v1alpha1.Workload{}
							err := c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, workload)
							Expect(err).NotTo(HaveOccurred())

							if len(workload.Status.Resources) < 2 {
								return []metav1.Condition{}
							}

							return workload.Status.Resources[0].Conditions
						}).Should(MatchAllElements(getConditionOfType, Elements{
							"ResourceSubmitted": MatchFields(IgnoreExtras, Fields{
								"Status":  Equal(metav1.ConditionUnknown),
								"Reason":  Equal(v1alpha1.MissingValueAtPathResourcesSubmittedReason),
								"Message": Equal(fmt.Sprintf("waiting to read value [spec.foo] from object [testobjs.test.run/%s] in namespace [%s]", createdObject.Name, testNS)),
							}),
							"Healthy": MatchFields(IgnoreExtras, Fields{
								"Status":  Equal(metav1.ConditionUnknown),
								"Reason":  Equal("ReadyCondition"),
								"Message": Equal("condition with type [Ready] not found on resource status"),
							}),
							"Ready": MatchFields(IgnoreExtras, Fields{
								"Status": Equal(metav1.ConditionUnknown),
								"Reason": Equal(v1alpha1.MissingValueAtPathResourcesSubmittedReason),
							}),
						}))

						workload := &v1alpha1.Workload{}
						err := c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, workload)
						Expect(err).NotTo(HaveOccurred())

						Expect(workload.Status.Conditions).To(MatchAllElements(getConditionOfType, Elements{
							"SupplyChainReady": MatchFields(IgnoreExtras, Fields{
								"Status": Equal(metav1.ConditionTrue),
							}),
							"ResourcesSubmitted": MatchFields(IgnoreExtras, Fields{
								"Status": Equal(metav1.ConditionUnknown),
								"Reason": Equal(v1alpha1.MissingValueAtPathResourcesSubmittedReason),
							}),
							"ResourcesHealthy": MatchFields(IgnoreExtras, Fields{
								"Status": Equal(metav1.ConditionUnknown),
								"Reason": Equal("HealthyConditionRule"),
							}),
							"Ready": MatchFields(IgnoreExtras, Fields{
								"Status": Equal(metav1.ConditionUnknown),
							}),
						}))

						Expect(workload.Status.Resources[0].Outputs).To(HaveLen(0))
					})

					When("the healthRule is subsequently satisfied", func() {
						It("results in a healthy workload and propagates outputs", func() {
							// update the object
							testToUpdate := getTestObjAtIndex(ctx, testNS, 0, 1)
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

							itResultsInAHealthyWorkload()

							Eventually(func() v1alpha1.Output {
								workload := &v1alpha1.Workload{}
								Expect(c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, workload)).To(Succeed())
								if len(workload.Status.Resources[0].Outputs) < 1 {
									return v1alpha1.Output{}
								}
								return workload.Status.Resources[0].Outputs[0]
							}).Should(MatchFields(IgnoreExtras, Fields{
								"Name":    Equal("config"),
								"Preview": Equal("some-address\n"),
							}))
						})
					})
					When("the healthRule is subsequently violated", func() {
						It("results in an unhealthy workload that reports the absence of outputs", func() {
							// update the object
							testToUpdate := getTestObjAtIndex(ctx, testNS, 0, 1)
							testToUpdate.Status.Conditions = []metav1.Condition{
								{
									Type:               "Ready",
									Status:             "False",
									Reason:             "SomeReason",
									LastTransitionTime: metav1.Now(),
								},
							}

							err := c.Status().Update(ctx, testToUpdate)
							Expect(err).NotTo(HaveOccurred())

							// assert expected state
							Eventually(func() []metav1.Condition {
								workload := &v1alpha1.Workload{}
								err := c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, workload)
								Expect(err).NotTo(HaveOccurred())

								if len(workload.Status.Resources) < 2 {
									return []metav1.Condition{}
								}

								return workload.Status.Resources[0].Conditions
							}).Should(MatchAllElements(getConditionOfType, Elements{
								"ResourceSubmitted": MatchFields(IgnoreExtras, Fields{
									"Status": Equal(metav1.ConditionUnknown),
									"Reason": Equal(v1alpha1.MissingValueAtPathResourcesSubmittedReason),
									"Message": Equal(fmt.Sprintf(
										"cannot read value [spec.foo] from unhealthy object [testobjs.test.run/%s] in namespace [%s], examine object, particularly whether it is receiving proper inputs",
										testToUpdate.Name,
										testNS,
									)),
								}),
								"Healthy": MatchFields(IgnoreExtras, Fields{
									"Status": Equal(metav1.ConditionFalse),
									"Reason": Equal("ReadyCondition"),
								}),
								"Ready": MatchFields(IgnoreExtras, Fields{
									"Status": Equal(metav1.ConditionFalse),
									"Reason": Equal("ReadyCondition"),
								}),
							}))

							workload := &v1alpha1.Workload{}
							err = c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, workload)
							Expect(err).NotTo(HaveOccurred())

							Expect(workload.Status.Conditions).To(MatchAllElements(getConditionOfType, Elements{
								"SupplyChainReady": MatchFields(IgnoreExtras, Fields{
									"Status": Equal(metav1.ConditionTrue),
								}),
								"ResourcesSubmitted": MatchFields(IgnoreExtras, Fields{
									"Status": Equal(metav1.ConditionUnknown),
									"Reason": Equal(v1alpha1.MissingValueAtPathResourcesSubmittedReason),
									"Message": Equal(fmt.Sprintf(
										"cannot read value [spec.foo] from unhealthy object [testobjs.test.run/%s] in namespace [%s], examine object, particularly whether it is receiving proper inputs",
										testToUpdate.Name,
										testNS,
									)),
								}),
								"ResourcesHealthy": MatchFields(IgnoreExtras, Fields{
									"Status": Equal(metav1.ConditionFalse),
									"Reason": Equal("HealthyConditionRule"),
								}),
								"Ready": MatchFields(IgnoreExtras, Fields{
									"Status": Equal(metav1.ConditionFalse),
								}),
							}))

							Expect(workload.Status.Resources[0].Outputs).To(HaveLen(0))
						})
					})
				})
			})

			Context("whose output will not be found", func() {
				var createdObj *resources.TestObj
				BeforeEach(func() {
					configPathThatWillNotBeFound := "status.someOutput"
					healthRuleSpecification = "healthRule:\n    singleConditionType: Ready"
					templateYaml := utils.HereYamlF(immutableTemplateBase, configPathThatWillNotBeFound, lifecycleSpecification, healthRuleSpecification)
					template := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, templateYaml)
					cleanups = append(cleanups, template)
				})

				Context("while healthy", func() {
					BeforeEach(func() {
						createdObj = getTestObjAtIndex(ctx, testNS, 0, 1)
						createdObj.Status.Conditions = []metav1.Condition{
							{
								Type:               "Ready",
								Status:             "True",
								Reason:             "Ready",
								LastTransitionTime: metav1.Now(),
							},
						}

						Expect(c.Status().Update(ctx, createdObj)).To(Succeed())
					})

					It("returns an ResourceSubmitted error directing the reader to a Platform Eng", func() {
						Eventually(func() []metav1.Condition {
							workload := &v1alpha1.Workload{}
							Expect(c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, workload)).To(Succeed())

							if len(workload.Status.Resources) < 2 {
								return []metav1.Condition{}
							}
							return workload.Status.Resources[0].Conditions
						}).Should(MatchElements(getConditionOfType, IgnoreExtras, Elements{
							"ResourceSubmitted": MatchFields(IgnoreExtras, Fields{
								"Status": Equal(metav1.ConditionUnknown),
								"Reason": Equal(v1alpha1.MissingValueAtPathResourcesSubmittedReason),
								"Message": Equal(fmt.Sprintf(
									"cannot read value [status.someOutput] from healthy object [testobjs.test.run/%s] in namespace [%s], contact Platform Eng",
									createdObj.Name,
									testNS,
								)),
							}),
						}))

						workload := &v1alpha1.Workload{}
						Expect(c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, workload)).To(Succeed())

						Expect(workload.Status.Conditions).To(MatchElements(getConditionOfType, IgnoreExtras, Elements{
							"ResourcesSubmitted": MatchFields(IgnoreExtras, Fields{
								"Status": Equal(metav1.ConditionUnknown),
								"Reason": Equal(v1alpha1.MissingValueAtPathResourcesSubmittedReason),
								"Message": Equal(fmt.Sprintf(
									"cannot read value [status.someOutput] from healthy object [testobjs.test.run/%s] in namespace [%s], contact Platform Eng",
									createdObj.Name,
									testNS,
								)),
							}),
						}))
					})
				})

				Context("while unhealthy", func() {
					BeforeEach(func() {
						createdObj = getTestObjAtIndex(ctx, testNS, 0, 1)
						createdObj.Status.Conditions = []metav1.Condition{
							{
								Type:               "Ready",
								Status:             "False",
								Reason:             "SomeReason",
								LastTransitionTime: metav1.Now(),
							},
						}

						Expect(c.Status().Update(ctx, createdObj)).To(Succeed())
					})

					It("returns an ResourceSubmitted error directing the reader to examine inputs", func() {
						Eventually(func() []metav1.Condition {
							workload := &v1alpha1.Workload{}
							Expect(c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, workload)).To(Succeed())

							if len(workload.Status.Resources) < 2 {
								return []metav1.Condition{}
							}
							return workload.Status.Resources[0].Conditions
						}).Should(MatchElements(getConditionOfType, IgnoreExtras, Elements{
							"ResourceSubmitted": MatchFields(IgnoreExtras, Fields{
								"Status": Equal(metav1.ConditionUnknown),
								"Reason": Equal(v1alpha1.MissingValueAtPathResourcesSubmittedReason),
								"Message": Equal(fmt.Sprintf(
									"cannot read value [status.someOutput] from unhealthy object [testobjs.test.run/%s] in namespace [%s], examine object, particularly whether it is receiving proper inputs",
									createdObj.Name,
									testNS,
								)),
							}),
						}))

						workload := &v1alpha1.Workload{}
						Expect(c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, workload)).To(Succeed())

						Expect(workload.Status.Conditions).To(MatchElements(getConditionOfType, IgnoreExtras, Elements{
							"ResourcesSubmitted": MatchFields(IgnoreExtras, Fields{
								"Status": Equal(metav1.ConditionUnknown),
								"Reason": Equal(v1alpha1.MissingValueAtPathResourcesSubmittedReason),
								"Message": Equal(fmt.Sprintf(
									"cannot read value [status.someOutput] from unhealthy object [testobjs.test.run/%s] in namespace [%s], examine object, particularly whether it is receiving proper inputs",
									createdObj.Name,
									testNS,
								)),
							}),
						}))
					})
				})

				Context("while health unknown", func() {
					BeforeEach(func() {
						createdObj = getTestObjAtIndex(ctx, testNS, 0, 1)
						Expect(c.Status().Update(ctx, createdObj)).To(Succeed())
					})
					It("returns an ResourceSubmitted error suggesting the reader wait", func() {
						Eventually(func() []metav1.Condition {
							workload := &v1alpha1.Workload{}
							Expect(c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, workload)).To(Succeed())

							if len(workload.Status.Resources) < 2 {
								return []metav1.Condition{}
							}
							return workload.Status.Resources[0].Conditions
						}).Should(MatchElements(getConditionOfType, IgnoreExtras, Elements{
							"ResourceSubmitted": MatchFields(IgnoreExtras, Fields{
								"Status": Equal(metav1.ConditionUnknown),
								"Reason": Equal(v1alpha1.MissingValueAtPathResourcesSubmittedReason),
								"Message": Equal(fmt.Sprintf(
									"waiting to read value [status.someOutput] from object [testobjs.test.run/%s] in namespace [%s]",
									createdObj.Name,
									testNS,
								)),
							}),
						}))

						workload := &v1alpha1.Workload{}
						Expect(c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, workload)).To(Succeed())

						Expect(workload.Status.Conditions).To(MatchElements(getConditionOfType, IgnoreExtras, Elements{
							"ResourcesSubmitted": MatchFields(IgnoreExtras, Fields{
								"Status": Equal(metav1.ConditionUnknown),
								"Reason": Equal(v1alpha1.MissingValueAtPathResourcesSubmittedReason),
								"Message": Equal(fmt.Sprintf(
									"waiting to read value [status.someOutput] from object [testobjs.test.run/%s] in namespace [%s]",
									createdObj.Name,
									testNS,
								)),
							}),
						}))
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
					templateYaml := utils.HereYamlF(immutableTemplateBase, configPathThatWillBeFound, lifecycleSpecification, healthRuleSpecification)
					template := utils.CreateObjectOnClusterFromYamlDefinition(ctx, c, templateYaml)
					cleanups = append(cleanups, template)
				})

				It("stamps the templated object once", func() {
					itStampsTheTemplatedObjectOnce()
				})

				When("the stamped object's succeeded condition has status == true", func() {
					It("results in a healthy workload and propagates outputs", func() {
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

						workload := &v1alpha1.Workload{}
						err = c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, workload)
						Expect(err).NotTo(HaveOccurred())

						Expect(workload.Status.Resources[0].Outputs).To(HaveLen(1))
						Expect(workload.Status.Resources[0].Outputs[0]).To(MatchFields(IgnoreExtras, Fields{
							"Name":    Equal("config"),
							"Preview": Equal("some-address\n"),
						}))
					})
				})

				When("a successful stamp is followed by an unsuccessful stamp", func() {
					It("passes the outputs of the first object and reports an unhealthy object", func() {
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

						image := "a-different-image"
						workload.Spec.Source.Image = &image
						utils.UpdateObjectOnCluster(ctx, c, &workload, &v1alpha1.Workload{})

						getConditionOfType := func(element interface{}) string {
							return element.(metav1.Condition).Type
						}

						Eventually(func() []metav1.Condition {
							workload := &v1alpha1.Workload{}
							err := c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, workload)
							Expect(err).NotTo(HaveOccurred())

							if len(workload.Status.Resources) < 2 {
								return []metav1.Condition{}
							}

							return workload.Status.Resources[0].Conditions
						}).Should(MatchAllElements(getConditionOfType, Elements{
							"ResourceSubmitted": MatchFields(IgnoreExtras, Fields{
								"Status": Equal(metav1.ConditionTrue),
								"Reason": Equal("ResourceSubmissionComplete"),
							}),
							"Healthy": MatchFields(IgnoreExtras, Fields{
								"Status":  Equal(metav1.ConditionUnknown),
								"Reason":  Equal(v1alpha1.SucceededStampedObjectConditionReason),
								"Message": ContainSubstring("condition with type [Succeeded] not found on resource status"),
							}),
							"Ready": MatchFields(IgnoreExtras, Fields{
								"Status": Equal(metav1.ConditionUnknown),
								"Reason": Equal(v1alpha1.SucceededStampedObjectConditionReason),
							}),
						}))

						workload := &v1alpha1.Workload{}
						err = c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, workload)
						Expect(err).NotTo(HaveOccurred())

						Expect(workload.Status.Conditions).To(MatchAllElements(getConditionOfType, Elements{
							"SupplyChainReady": MatchFields(IgnoreExtras, Fields{
								"Status": Equal(metav1.ConditionTrue),
							}),
							"ResourcesSubmitted": MatchFields(IgnoreExtras, Fields{
								"Status": Equal(metav1.ConditionTrue),
							}),
							"ResourcesHealthy": MatchFields(IgnoreExtras, Fields{
								"Status": Equal(metav1.ConditionUnknown),
								"Reason": Equal("HealthyConditionRule"),
							}),
							"Ready": MatchFields(IgnoreExtras, Fields{
								"Status": Equal(metav1.ConditionUnknown),
							}),
						}))

						Expect(workload.Status.Resources[0].Outputs).To(HaveLen(1))
						Expect(workload.Status.Resources[0].Outputs[0]).To(MatchFields(IgnoreExtras, Fields{
							"Name":    Equal("config"),
							"Preview": Equal("some-address\n"),
						}))
					})
				})

				When("the stamped object's succeeded condition is not yet true", func() {
					It("results in proper status", func() {
						getConditionOfType := func(element interface{}) string {
							return element.(metav1.Condition).Type
						}

						createdObject := getTestObjAtIndex(ctx, testNS, 0, 1)

						Eventually(func() []metav1.Condition {
							workload := &v1alpha1.Workload{}
							err := c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, workload)
							Expect(err).NotTo(HaveOccurred())

							if len(workload.Status.Resources) < 2 {
								return []metav1.Condition{}
							}

							return workload.Status.Resources[0].Conditions
						}).Should(MatchAllElements(getConditionOfType, Elements{
							"ResourceSubmitted": MatchFields(IgnoreExtras, Fields{
								"Status":  Equal(metav1.ConditionUnknown),
								"Reason":  Equal(v1alpha1.MissingValueAtPathResourcesSubmittedReason),
								"Message": Equal(fmt.Sprintf("waiting to read value [spec.foo] from object [testobjs.test.run/%s] in namespace [%s]", createdObject.Name, testNS)),
							}),
							"Healthy": MatchFields(IgnoreExtras, Fields{
								"Status":  Equal(metav1.ConditionUnknown),
								"Reason":  Equal(v1alpha1.SucceededStampedObjectConditionReason),
								"Message": ContainSubstring("condition with type [Succeeded] not found on resource status"),
							}),
							"Ready": MatchFields(IgnoreExtras, Fields{
								"Status": Equal(metav1.ConditionUnknown),
								"Reason": Equal(v1alpha1.MissingValueAtPathResourcesSubmittedReason),
							}),
						}))

						workload := &v1alpha1.Workload{}
						err := c.Get(ctx, client.ObjectKey{Name: "workload-joe", Namespace: testNS}, workload)
						Expect(err).NotTo(HaveOccurred())

						Expect(workload.Status.Conditions).To(MatchAllElements(getConditionOfType, Elements{
							"SupplyChainReady": MatchFields(IgnoreExtras, Fields{
								"Status": Equal(metav1.ConditionTrue),
							}),
							"ResourcesSubmitted": MatchFields(IgnoreExtras, Fields{
								"Status": Equal(metav1.ConditionUnknown),
								"Reason": Equal(v1alpha1.MissingValueAtPathResourcesSubmittedReason),
							}),
							"ResourcesHealthy": MatchFields(IgnoreExtras, Fields{
								"Status": Equal(metav1.ConditionUnknown),
								"Reason": Equal("HealthyConditionRule"),
							}),
							"Ready": MatchFields(IgnoreExtras, Fields{
								"Status": Equal(metav1.ConditionUnknown),
							}),
						}))

						Expect(workload.Status.Resources[0].Outputs).To(HaveLen(0))
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
				kind: ClusterConfigTemplate
				metadata:
				  name: my-config-template
				spec:
				  configPath: data.foo
			      template:
					apiVersion: v1
					kind: ConfigMap
					metadata:
					  name: mutable-test-obj
					data:
					  foo: hard-coded-other-val
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

			workload := v1alpha1.Workload{
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

			cleanups = append(cleanups, &workload)
			err := c.Create(ctx, &workload, &client.CreateOptions{})
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
				kind: ClusterConfigTemplate
				metadata:
				  name: my-config-template
				spec:
				  configPath: spec.foo
			      lifecycle: immutable
			      template:
					apiVersion: test.run/v1alpha1
					kind: TestObj
					metadata:
					  generateName: test-resource-
					spec:
					  foo: $(params.foo)$
			`)

				utils.UpdateObjectOnClusterFromYamlDefinition(ctx, c, immutableTemplateYaml, testNS, &v1alpha1.ClusterConfigTemplate{})
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
})

func getTestObjAtIndex(ctx context.Context, namespace string, index int, numObjectsExpected int) *resources.TestObj {
	opts := []client.ListOption{
		client.InNamespace(namespace),
	}

	testsList := &resources.TestObjList{}

	Eventually(func() ([]resources.TestObj, error) {
		err := c.List(ctx, testsList, opts...)
		return testsList.Items, err
	}).Should(HaveLen(numObjectsExpected))

	return &testsList.Items[index]
}

func getConditionOfType(element interface{}) string {
	return element.(metav1.Condition).Type
}
