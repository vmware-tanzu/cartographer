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

package runnable_test

import (
	"context"
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apiserver/pkg/storage/names"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	. "github.com/vmware-tanzu/cartographer/pkg/utils"
	"github.com/vmware-tanzu/cartographer/tests/helpers"
	"github.com/vmware-tanzu/cartographer/tests/resources"
)

var _ = Describe("Stamping a resource on Runnable Creation", func() {
	var (
		ctx                   context.Context
		runnableDefinition    *unstructured.Unstructured
		runTemplateDefinition *unstructured.Unstructured
	)

	var createNamespacedObject = func(ctx context.Context, objYaml, namespace string) *unstructured.Unstructured {
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

	BeforeEach(func() {
		ctx = context.Background()
	})

	getRunnableTestStatus := func() (metav1.Condition, error) {
		runnable := &v1alpha1.Runnable{}
		err := c.Get(ctx, client.ObjectKey{Name: "my-runnable", Namespace: testNS}, runnable)
		if err != nil {
			return metav1.Condition{}, err
		}
		testStatusCondition := &metav1.Condition{}
		testStatusConditionJson := runnable.Status.Outputs["test-status"].Raw
		err = json.Unmarshal(testStatusConditionJson, testStatusCondition)
		return *testStatusCondition, err
	}

	Describe("when a ClusterRunTemplate that produces a Resource leverages a Runnable field", func() {
		BeforeEach(func() {
			runTemplateYaml := HereYamlF(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterRunTemplate
				metadata:
				  namespace: %s
				  name: my-run-template
				spec:
				  template:
					apiVersion: v1
					kind: ResourceQuota
					metadata:
					  generateName: my-stamped-resource-
					  namespace: %s
					  labels:
					    focus: something-useful
					spec:
					  hard:
						cpu: "1000"
						memory: 200Gi
						pods: "10"
					  scopeSelector:
						matchExpressions:
						- operator : In
						  scopeName: PriorityClass
						  values: [$(runnable.spec.inputs.key)$]
				`,
				testNS, testNS,
			)

			runTemplateDefinition = createNamespacedObject(ctx, runTemplateYaml, testNS)
		})

		AfterEach(func() {
			err := c.Delete(ctx, runTemplateDefinition)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("and a Runnable matches the RunTemplateRef", func() {
			BeforeEach(func() {
				runnableYaml := HereYamlF(`---
					apiVersion: carto.run/v1alpha1
					kind: Runnable
					metadata:
					  namespace: %s
					  name: my-runnable
					spec:
					  runTemplateRef: 
					    name: my-run-template
					    namespace: %s
					    kind: ClusterRunTemplate
					  inputs:
					    key: val
					`,
					testNS, testNS)

				runnableDefinition = createNamespacedObject(ctx, runnableYaml, testNS)
			})

			AfterEach(func() {
				err := c.Delete(ctx, runnableDefinition)
				Expect(err).NotTo(HaveOccurred())
			})

			It("stamps the templated object once", func() {
				resourceList := &v1.ResourceQuotaList{}

				Eventually(func() (int, error) {
					err := c.List(ctx, resourceList, &client.ListOptions{Namespace: testNS})
					return len(resourceList.Items), err
				}).Should(Equal(1))

				Consistently(func() (int, error) {
					err := c.List(ctx, resourceList, &client.ListOptions{Namespace: testNS})
					return len(resourceList.Items), err
				}, "2s").Should(BeNumerically("<=", 1))

				Expect(resourceList.Items[0].Name).To(ContainSubstring("my-stamped-resource-"))
				Expect(resourceList.Items[0].Spec.ScopeSelector.MatchExpressions[0].Values).To(ConsistOf("val"))
			})

			Context("and the Runnable object is updated", func() {
				BeforeEach(func() {
					Expect(AlterFieldOfNestedStringMaps(runnableDefinition.Object, "spec.inputs.key", "new-val")).To(Succeed())
					Expect(c.Update(ctx, runnableDefinition, &client.UpdateOptions{})).To(Succeed())
				})
				It("creates a second object alongside the first", func() {
					resourceList := &v1.ResourceQuotaList{}

					Eventually(func() (int, error) {
						err := c.List(ctx, resourceList, &client.ListOptions{Namespace: testNS})
						return len(resourceList.Items), err
					}, "3s").Should(Equal(2))

					Consistently(func() (int, error) {
						err := c.List(ctx, resourceList, &client.ListOptions{Namespace: testNS})
						return len(resourceList.Items), err
					}, "2s").Should(Equal(2))

					Expect(resourceList.Items[0].Name).To(ContainSubstring("my-stamped-resource-"))
					Expect(resourceList.Items[1].Name).To(ContainSubstring("my-stamped-resource-"))

					id := func(element interface{}) string {
						return element.(v1.ResourceQuota).Spec.ScopeSelector.MatchExpressions[0].Values[0]
					}
					Expect(resourceList.Items).To(MatchAllElements(id, Elements{
						"val":     Not(BeNil()),
						"new-val": Not(BeNil()),
					}))
				})
			})

			Context("and the ClusterRunTemplate object is updated", func() {
				It("creates a second object alongside the first", func() {
					resourceList := &v1.ResourceQuotaList{}

					Eventually(func() (int, error) {
						err := c.List(ctx, resourceList, &client.ListOptions{Namespace: testNS})
						return len(resourceList.Items), err
					}).Should(Equal(1))

					// Ensure that first object has been stamped, and status update reconcile of runnable has occurred
					Consistently(func() (int, error) {
						err := c.List(ctx, resourceList, &client.ListOptions{Namespace: testNS})
						return len(resourceList.Items), err
					}, "2s").Should(BeNumerically("<=", 1))

					Expect(AlterFieldOfNestedStringMaps(runTemplateDefinition.Object, "spec.template.metadata.labels.focus", "other-things")).To(Succeed())
					Expect(c.Update(ctx, runTemplateDefinition, &client.UpdateOptions{})).To(Succeed())

					Eventually(func() (int, error) {
						err := c.List(ctx, resourceList, &client.ListOptions{Namespace: testNS})
						return len(resourceList.Items), err
					}).Should(Equal(2))

					Consistently(func() (int, error) {
						err := c.List(ctx, resourceList, &client.ListOptions{Namespace: testNS})
						return len(resourceList.Items), err
					}, "2s").Should(BeNumerically("<=", 2))

					Expect(resourceList.Items[0].Name).To(ContainSubstring("my-stamped-resource-"))
					Expect(resourceList.Items[1].Name).To(ContainSubstring("my-stamped-resource-"))

					Expect(resourceList.Items[0].UID).NotTo(Equal(resourceList.Items[1].UID))

					id := func(element interface{}) string {
						return element.(v1.ResourceQuota).ObjectMeta.Labels["focus"]
					}
					Expect(resourceList.Items).To(MatchAllElements(id, Elements{
						"something-useful": Not(BeNil()),
						"other-things":     Not(BeNil()),
					}))
				})
			})
		})

		Context("a Runnable that does not match the RunTemplateRef", func() {
			BeforeEach(func() {
				runnableYaml := HereYamlF(`---
					apiVersion: carto.run/v1alpha1
					kind: Runnable
					metadata:
					  namespace: %s
					  name: my-runnable
					spec:
					  runTemplateRef: 
					    name: my-run-template-does-not-match
					    namespace: %s
					    kind: ClusterRunTemplate
					`,
					testNS, testNS)

				runnableDefinition = &unstructured.Unstructured{}
				err := yaml.Unmarshal([]byte(runnableYaml), runnableDefinition)
				Expect(err).NotTo(HaveOccurred())

				err = c.Create(ctx, runnableDefinition, &client.CreateOptions{})
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				err := c.Delete(ctx, runnableDefinition)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Does not stamp a new Resource", func() {
				resourceList := &v1.ConfigMapList{} //TODO figure out if this is correct or should be a resourceQuotaList

				Consistently(func() (int, error) {
					err := c.List(ctx, resourceList, &client.ListOptions{Namespace: testNS})
					return len(resourceList.Items), err
				}).Should(Equal(0))
			})
		})
	})

	Describe("when a ClusterRunTemplate that produces a Resource leverages a Selector field", func() {
		BeforeEach(func() {
			runTemplateYaml := HereYamlF(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterRunTemplate
				metadata:
				  name: run-template---multi-label-selector
				spec:
				  template:
					apiVersion: v1
					kind: ResourceQuota
					metadata:
					  generateName: my-stamped-resource-
					  namespace: %s
					  labels:
					    focus: something-useful
					spec:
					  hard:
						cpu: "1000"
						memory: 200Gi
						pods: "10"
					  scopeSelector:
						matchExpressions:
						- operator : In
						  scopeName: PriorityClass
						  values: [$(selected.spec.value.answer)$]
				`, testNS)

			runTemplateDefinition = createNamespacedObject(ctx, runTemplateYaml, testNS)
		})

		AfterEach(func() {
			err := c.Delete(ctx, runTemplateDefinition)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("and a Runnable matches the RunTemplateRef and has a selector for a namespace scoped object", func() {
			BeforeEach(func() {
				runnableYaml := HereYamlF(`
					---
					apiVersion: carto.run/v1alpha1
					kind: Runnable
					metadata:
					  name: runnable---multi-label-selector
					  labels:
						my-label: this-is-it
					spec:
					  runTemplateRef:
						name: run-template---multi-label-selector
						namespace: %s
					
					  selector:
						resource:
						  apiVersion: test.run/v1alpha1
						  kind: TestObj
						matchingLabels:
						  runnables.carto.run/group: dev---multi-label-selector
						  runnables.carto.run/stage: production
					`, testNS)

				runnableDefinition = createNamespacedObject(ctx, runnableYaml, testNS)
			})

			AfterEach(func() {
				err := c.Delete(ctx, runnableDefinition)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("and only one object matches the selector in the namespace", func() {
				var selectedDefinition *unstructured.Unstructured

				BeforeEach(func() {
					selectedYaml := HereYaml(`
						---
						apiVersion: test.run/v1alpha1
						kind: TestObj
						metadata:
						  name: dummy-selected-2---multi-label-selector
						  labels:
							runnables.carto.run/group: dev---multi-label-selector
							runnables.carto.run/stage: production

						spec:
						  value:
							answer: polo-production-stage
					`)
					selectedDefinition = createNamespacedObject(ctx, selectedYaml, testNS)
				})

				AfterEach(func() {
					err := c.Delete(ctx, selectedDefinition)
					Expect(err).NotTo(HaveOccurred())
				})

				It("stamps the templated object once", func() {
					resourceList := &v1.ResourceQuotaList{}

					Eventually(func() (int, error) {
						err := c.List(ctx, resourceList, &client.ListOptions{Namespace: testNS})
						return len(resourceList.Items), err
					}).Should(Equal(1))

					Consistently(func() (int, error) {
						err := c.List(ctx, resourceList, &client.ListOptions{Namespace: testNS})
						return len(resourceList.Items), err
					}, "2s").Should(BeNumerically("<=", 1))

					Expect(resourceList.Items[0].Name).To(ContainSubstring("my-stamped-resource-"))
					Expect(resourceList.Items[0].Spec.ScopeSelector.MatchExpressions[0].Values).To(ConsistOf("polo-production-stage"))
				})

				Context("and the Selected object is updated", func() {
					BeforeEach(func() {
						Expect(AlterFieldOfNestedStringMaps(selectedDefinition.Object, "spec.value.answer", "buzkashi-production-stage")).To(Succeed())
						Expect(c.Update(ctx, selectedDefinition, &client.UpdateOptions{})).To(Succeed())
					})
					It("creates a second object alongside the first", func() {
						resourceList := &v1.ResourceQuotaList{}

						Eventually(func() (int, error) {
							err := c.List(ctx, resourceList, &client.ListOptions{Namespace: testNS})
							return len(resourceList.Items), err
						}, "3s").Should(Equal(2))

						Consistently(func() (int, error) {
							err := c.List(ctx, resourceList, &client.ListOptions{Namespace: testNS})
							return len(resourceList.Items), err
						}, "2s").Should(Equal(2))

						Expect(resourceList.Items[0].Name).To(ContainSubstring("my-stamped-resource-"))
						Expect(resourceList.Items[1].Name).To(ContainSubstring("my-stamped-resource-"))

						id := func(element interface{}) string {
							return element.(v1.ResourceQuota).Spec.ScopeSelector.MatchExpressions[0].Values[0]
						}
						Expect(resourceList.Items).To(MatchAllElements(id, Elements{
							"polo-production-stage":     Not(BeNil()),
							"buzkashi-production-stage": Not(BeNil()),
						}))
					})
				})
			})

			Context("and an object in this namespace and another match the selector", func() {
				var selectedDefinition1, selectedDefinition2 *unstructured.Unstructured
				var otherNamespace string

				BeforeEach(func() {
					otherNamespace = names.SimpleNameGenerator.GenerateName("other-namespace-")
					err := helpers.EnsureNamespace(otherNamespace, c)
					Expect(err).ToNot(HaveOccurred())

					selectedYamlTemplate := `
						---
						apiVersion: test.run/v1alpha1
						kind: TestObj
						metadata:
						  name: dummy-selected-distractor
						  labels:
							runnables.carto.run/group: dev---multi-label-selector
							runnables.carto.run/stage: production

						spec:
						  value:
							answer: %s`

					selectedYaml1 := HereYamlF(selectedYamlTemplate, "polo-production-stage")
					selectedDefinition1 = createNamespacedObject(ctx, selectedYaml1, testNS)

					selectedYaml2 := HereYamlF(selectedYamlTemplate, "a-value-not-to-be-propagated")
					selectedDefinition2 = createNamespacedObject(ctx, selectedYaml2, otherNamespace)
				})

				AfterEach(func() {
					err := c.Delete(ctx, selectedDefinition1)
					Expect(err).NotTo(HaveOccurred())

					err = c.Delete(ctx, selectedDefinition2)
					Expect(err).NotTo(HaveOccurred())

					err = helpers.DeleteNamespace(otherNamespace, c)
					Expect(err).NotTo(HaveOccurred())
				})

				It("stamps the templated object once", func() {
					resourceList := &v1.ResourceQuotaList{}

					Eventually(func() (int, error) {
						err := c.List(ctx, resourceList, &client.ListOptions{Namespace: testNS})
						return len(resourceList.Items), err
					}).Should(Equal(1))

					Consistently(func() (int, error) {
						err := c.List(ctx, resourceList, &client.ListOptions{Namespace: testNS})
						return len(resourceList.Items), err
					}, "2s").Should(BeNumerically("<=", 1))

					Expect(resourceList.Items[0].Name).To(ContainSubstring("my-stamped-resource-"))
					Expect(resourceList.Items[0].Spec.ScopeSelector.MatchExpressions[0].Values).To(ConsistOf("polo-production-stage"))
				})
			})
		})

		Context("and a Runnable matches the RunTemplateRef and has a selector for a cluster scoped object", func() {
			BeforeEach(func() {
				runnableYaml := HereYamlF(`
					---
					apiVersion: carto.run/v1alpha1
					kind: Runnable
					metadata:
					  name: runnable---multi-label-selector
					  labels:
						my-label: this-is-it
					spec:
					  runTemplateRef:
						name: run-template---multi-label-selector
						namespace: %s

					  selector:
						resource:
						  apiVersion: test.run/v1alpha1
						  kind: ClusterTestObj
						matchingLabels:
						  runnables.carto.run/group: dev---multi-label-selector
						  runnables.carto.run/stage: production
					`, testNS)

				runnableDefinition = createNamespacedObject(ctx, runnableYaml, testNS)
			})

			AfterEach(func() {
				err := c.Delete(ctx, runnableDefinition)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("and only one object matches the selector in the namespace", func() {
				var selectedDefinition *unstructured.Unstructured

				BeforeEach(func() {
					selectedYaml := HereYaml(`
						---
						apiVersion: test.run/v1alpha1
						kind: ClusterTestObj
						metadata:
						  name: dummy-selected-2---multi-label-selector
						  labels:
							runnables.carto.run/group: dev---multi-label-selector
							runnables.carto.run/stage: production

						spec:
						  value:
							answer: polo-production-stage
					`)
					selectedDefinition = createNamespacedObject(ctx, selectedYaml, "")
				})

				AfterEach(func() {
					err := c.Delete(ctx, selectedDefinition)
					Expect(err).NotTo(HaveOccurred())
				})

				It("stamps the templated object once", func() {
					resourceList := &v1.ResourceQuotaList{}

					Eventually(func() (int, error) {
						err := c.List(ctx, resourceList, &client.ListOptions{Namespace: testNS})
						return len(resourceList.Items), err
					}).Should(Equal(1))

					Consistently(func() (int, error) {
						err := c.List(ctx, resourceList, &client.ListOptions{Namespace: testNS})
						return len(resourceList.Items), err
					}, "2s").Should(BeNumerically("<=", 1))

					Expect(resourceList.Items[0].Name).To(ContainSubstring("my-stamped-resource-"))
					Expect(resourceList.Items[0].Spec.ScopeSelector.MatchExpressions[0].Values).To(ConsistOf("polo-production-stage"))
				})
			})
		})
	})

	Describe("A ClusterRunTemplate that selects for outputs that are eventually available", func() {
		BeforeEach(func() {
			runTemplateYaml := HereYamlF(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterRunTemplate
				metadata:
				  namespace: %s
				  name: my-run-template
				spec:
				  outputs:
					test-status: status.conditions[?(@.type=="Ready")]
				  template:
					apiVersion: test.run/v1alpha1
					kind: TestObj
					metadata:
					  name: test-crd
					spec:
					  foo: "bar"
				`,
				testNS,
			)

			runTemplateDefinition = &unstructured.Unstructured{}
			err := yaml.Unmarshal([]byte(runTemplateYaml), runTemplateDefinition)
			Expect(err).NotTo(HaveOccurred())

			err = c.Create(ctx, runTemplateDefinition, &client.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			runnableYaml := HereYamlF(`---
					apiVersion: carto.run/v1alpha1
					kind: Runnable
					metadata:
					  namespace: %s
					  name: my-runnable
					  labels:
					    some-val: first
					spec:
					  runTemplateRef: 
					    name: my-run-template
					    namespace: %s
					    kind: ClusterRunTemplate
					`,
				testNS, testNS)

			runnableDefinition = &unstructured.Unstructured{}
			err = yaml.Unmarshal([]byte(runnableYaml), runnableDefinition)
			Expect(err).NotTo(HaveOccurred())

			err = c.Create(ctx, runnableDefinition, &client.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := c.Delete(ctx, runnableDefinition)
			Expect(err).NotTo(HaveOccurred())

			err = c.Delete(ctx, runTemplateDefinition)
			Expect(err).NotTo(HaveOccurred())
		})

		It("populates the runnable.Status.outputs properly", func() {
			opts := []client.ListOption{
				client.InNamespace(testNS),
				client.MatchingLabels(map[string]string{"carto.run/runnable-name": "my-runnable"}),
			}

			testsList := &resources.TestObjList{}

			Eventually(func() ([]resources.TestObj, error) {
				err := c.List(ctx, testsList, opts...)
				return testsList.Items, err
			}).Should(HaveLen(1))

			By("reflecting status when succeeded is true")
			testToUpdate := &testsList.Items[0]
			testToUpdate.Status.Conditions = []metav1.Condition{
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
			err := c.Status().Update(ctx, testToUpdate)
			Expect(err).NotTo(HaveOccurred())

			Eventually(getRunnableTestStatus, "10s").Should(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal("Ready"),
				"Status": Equal(metav1.ConditionStatus("True")),
				"Reason": Equal("LifeIsGood"),
			}))

			By("reflecting the past status succeeded is no longer succeeded")
			testToUpdate.Status.Conditions = []metav1.Condition{
				{
					Type:               "Ready",
					Status:             "False",
					Reason:             "LifeIsSad",
					LastTransitionTime: metav1.Now(),
				},
				{
					Type:               "Succeeded",
					Status:             "False",
					Reason:             "Failure",
					LastTransitionTime: metav1.Now(),
				},
			}

			err = c.Status().Update(ctx, testToUpdate)
			Expect(err).NotTo(HaveOccurred())

			Consistently(getRunnableTestStatus).Should(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal("Ready"),
				"Status": Equal(metav1.ConditionStatus("True")),
				"Reason": Equal("LifeIsGood"),
			}))

			By("reflecting the most recent status when succeeded is true again")

			testToUpdate.Status.Conditions = []metav1.Condition{
				{
					Type:               "Ready",
					Status:             "True",
					Reason:             "LifeIsGreat",
					LastTransitionTime: metav1.Now(),
				},
				{
					Type:               "Succeeded",
					Status:             "True",
					Reason:             "Success",
					LastTransitionTime: metav1.Now(),
				},
			}

			err = c.Status().Update(ctx, testToUpdate)
			Expect(err).NotTo(HaveOccurred())

			Eventually(getRunnableTestStatus, "10s").Should(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal("Ready"),
				"Status": Equal(metav1.ConditionStatus("True")),
				"Reason": Equal("LifeIsGreat"),
			}))
		})
	})

	Describe("when a ClusterRunTemplate that produces a Resource leverages a Selected field", func() {
		BeforeEach(func() {
			runTemplateYaml := HereYamlF(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterRunTemplate
				metadata:
				  namespace: %s
				  name: my-run-template
				spec:
				  template:
					apiVersion: v1
					kind: ResourceQuota
					metadata:
					  generateName: my-stamped-resource-
					  namespace: %s
					  labels:
					    focus: something-useful
					spec:
					  hard:
						cpu: "1000"
						memory: 200Gi
						pods: "10"
					  scopeSelector:
						matchExpressions:
						- operator : In
						  scopeName: PriorityClass
						  values: [$(selected.spec.inputs.key)$]
				`,
				testNS, testNS,
			)

			runTemplateDefinition = &unstructured.Unstructured{}
			err := yaml.Unmarshal([]byte(runTemplateYaml), runTemplateDefinition)
			Expect(err).NotTo(HaveOccurred())

			err = c.Create(ctx, runTemplateDefinition, &client.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Multiple objects created", func() {
		BeforeEach(func() {
			runnableYaml := HereYamlF(`---
					apiVersion: carto.run/v1alpha1
					kind: Runnable
					metadata:
					  namespace: %s
					  name: my-runnable
					  labels:
					    some-val: first
					spec:
					  runTemplateRef: 
					    name: my-run-template
					    namespace: %s
					    kind: ClusterRunTemplate
					`,
				testNS, testNS)

			runnableDefinition = &unstructured.Unstructured{}
			err := yaml.Unmarshal([]byte(runnableYaml), runnableDefinition)
			Expect(err).NotTo(HaveOccurred())

			err = c.Create(ctx, runnableDefinition, &client.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			runTemplateYaml := HereYamlF(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterRunTemplate
				metadata:
				  namespace: %s
				  name: my-run-template
				spec:
				  outputs:
					test-status: status.conditions[?(@.type=="Ready")]
				  template:
					apiVersion: test.run/v1alpha1
					kind: TestObj
					metadata:
					  generateName: test-crd-
					  labels:
					    gen: "1"
					spec:
					  foo: "bar"
				`,
				testNS,
			)

			runTemplateDefinition = &unstructured.Unstructured{}
			err = yaml.Unmarshal([]byte(runTemplateYaml), runTemplateDefinition)
			Expect(err).NotTo(HaveOccurred())

			err = c.Create(ctx, runTemplateDefinition, &client.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			opts := []client.ListOption{
				client.InNamespace(testNS),
				client.MatchingLabels(map[string]string{"carto.run/runnable-name": "my-runnable"}),
			}

			testsList := &resources.TestObjList{}

			Eventually(func() ([]resources.TestObj, error) {
				err := c.List(ctx, testsList, opts...)
				return testsList.Items, err
			}).Should(HaveLen(1))

			// This is in order to ensure gen 1 object and gen 2 object have different creationTimestamps
			time.Sleep(time.Second)

			Expect(AlterFieldOfNestedStringMaps(runTemplateDefinition.Object, "spec.template.metadata.labels.gen", "2")).To(Succeed())

			err = c.Update(ctx, runTemplateDefinition, &client.UpdateOptions{})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() ([]resources.TestObj, error) {
				err := c.List(ctx, testsList, opts...)
				return testsList.Items, err
			}).Should(HaveLen(2))

			// This is in order to ensure gen 2 object and gen 3 object have different creationTimestamps
			// Gen 3 object is needed to demonstrate behaviour when the most recently submitted is not successful
			time.Sleep(time.Second)

			Expect(AlterFieldOfNestedStringMaps(runTemplateDefinition.Object, "spec.template.metadata.labels.gen", "3")).To(Succeed())

			err = c.Update(ctx, runTemplateDefinition, &client.UpdateOptions{})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() ([]resources.TestObj, error) {
				err := c.List(ctx, testsList, opts...)
				return testsList.Items, err
			}).Should(HaveLen(3))
		})

		AfterEach(func() {
			err := c.Delete(ctx, runnableDefinition)
			Expect(err).NotTo(HaveOccurred())

			err = c.Delete(ctx, runTemplateDefinition)
			Expect(err).NotTo(HaveOccurred())
		})

		It("populates the runnable.Status.outputs properly", func() {
			By("updating runnable status based on the most recently submitted and successful object")
			opts := []client.ListOption{
				client.InNamespace(testNS),
				client.MatchingLabels(map[string]string{"gen": "2"}),
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
					Reason:             "LifeIsGood",
					LastTransitionTime: metav1.Now(),
					Message:            "this is generation 2",
				},
				{
					Type:               "Succeeded",
					Status:             "True",
					Reason:             "Success",
					LastTransitionTime: metav1.Now(),
				},
			}
			err := c.Status().Update(ctx, testToUpdate)
			Expect(err).NotTo(HaveOccurred())

			Eventually(getRunnableTestStatus, "10s").Should(MatchFields(IgnoreExtras, Fields{
				"Message": Equal("this is generation 2"),
			}))

			By("not updating runnable status based on the less recently submitted and successful objects")
			opts = []client.ListOption{
				client.InNamespace(testNS),
				client.MatchingLabels(map[string]string{"gen": "1"}),
			}

			Eventually(func() ([]resources.TestObj, error) {
				err := c.List(ctx, testsList, opts...)
				return testsList.Items, err
			}).Should(HaveLen(1))

			testToUpdate = &testsList.Items[0]
			testToUpdate.Status.Conditions = []metav1.Condition{
				{
					Type:               "Ready",
					Status:             "True",
					Reason:             "LifeIsGood",
					LastTransitionTime: metav1.Now(),
					Message:            "but this is earlier generation 1",
				},
				{
					Type:               "Succeeded",
					Status:             "True",
					Reason:             "Success",
					LastTransitionTime: metav1.Now(),
				},
			}
			err = c.Status().Update(ctx, testToUpdate)
			Expect(err).NotTo(HaveOccurred())

			Consistently(getRunnableTestStatus, "1s").Should(MatchFields(IgnoreExtras, Fields{
				"Message": And(
					Equal("this is generation 2"),
					Not(Equal("but this is earlier generation 1"))),
			}))
		})
	})
})
