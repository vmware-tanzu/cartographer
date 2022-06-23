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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	. "github.com/vmware-tanzu/cartographer/pkg/utils"
	"github.com/vmware-tanzu/cartographer/tests/helpers"
	"github.com/vmware-tanzu/cartographer/tests/resources"
	v1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apiserver/pkg/storage/names"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

var _ = Describe("Stamping a resource on Runnable Creation", func() {
	var (
		ctx                   context.Context
		runnableDefinition    *unstructured.Unstructured
		runTemplateDefinition *unstructured.Unstructured
		serviceAccountName    string
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

		serviceAccountName = "my-service-account"

		serviceAccountSecretYaml := HereYamlF(`---
			apiVersion: v1
			kind: Secret
			metadata:
			  namespace: %s
			  name: my-service-account-secret
			  annotations:
				kubernetes.io/service-account.name: my-service-account
			data:
			  token: ZXlKaGJHY2lPaUpTVXpJMU5pSXNJbXRwWkNJNklubFNWM1YxVDNSRldESnZVRE4wTUd0R1EzQmlVVlJOVWtkMFNGb3RYMGh2VUhKYU1FRnVOR0Y0WlRBaWZRLmV5SnBjM01pT2lKcmRXSmxjbTVsZEdWekwzTmxjblpwWTJWaFkyTnZkVzUwSWl3aWEzVmlaWEp1WlhSbGN5NXBieTl6WlhKMmFXTmxZV05qYjNWdWRDOXVZVzFsYzNCaFkyVWlPaUprWldaaGRXeDBJaXdpYTNWaVpYSnVaWFJsY3k1cGJ5OXpaWEoyYVdObFlXTmpiM1Z1ZEM5elpXTnlaWFF1Ym1GdFpTSTZJbTE1TFhOaExYUnZhMlZ1TFd4dVkzRndJaXdpYTNWaVpYSnVaWFJsY3k1cGJ5OXpaWEoyYVdObFlXTmpiM1Z1ZEM5elpYSjJhV05sTFdGalkyOTFiblF1Ym1GdFpTSTZJbTE1TFhOaElpd2lhM1ZpWlhKdVpYUmxjeTVwYnk5elpYSjJhV05sWVdOamIzVnVkQzl6WlhKMmFXTmxMV0ZqWTI5MWJuUXVkV2xrSWpvaU9HSXhNV1V3WldNdFlURTVOeTAwWVdNeUxXRmpORFF0T0RjelpHSmpOVE13TkdKbElpd2ljM1ZpSWpvaWMzbHpkR1Z0T25ObGNuWnBZMlZoWTJOdmRXNTBPbVJsWm1GMWJIUTZiWGt0YzJFaWZRLmplMzRsZ3hpTUtnd0QxUGFhY19UMUZNWHdXWENCZmhjcVhQMEE2VUV2T0F6ek9xWGhpUUdGN2poY3RSeFhmUVFJVEs0Q2tkVmZ0YW5SUjNPRUROTUxVMVBXNXVsV3htVTZTYkMzdmZKT3ozLVJPX3BOVkNmVW8tZURpblN1Wm53bjNzMjNjZU9KM3IzYk04cnBrMHZZZFgyRVlQRGItMnd4cjIzZ1RxUjVxZU5ULW11cS1qYktXVE8wYnRYVl9wVHNjTnFXUkZIVzJBVTVHYVBpbmNWVXg1bXExLXN0SFdOOGtjTG96OF96S2RnUnJGYV92clFjb3NWZzZCRW5MSEt2NW1fVEhaR3AybU8wYmtIV3J1Q2xEUDdLc0tMOFVaZWxvTDN4Y3dQa000VlBBb2V0bDl5MzlvUi1KbWh3RUlIcS1hX3BzaVh5WE9EQU44STcybEZpUSU=
			type: kubernetes.io/service-account-token
			`,
			testNS)

		_ = createNamespacedObject(ctx, serviceAccountSecretYaml, testNS)

		serviceAccountYaml := HereYamlF(`---
			apiVersion: v1
			kind: ServiceAccount
			metadata:
			  namespace: %s
			  name: %s
			secrets:
			- name: my-service-account-secret
			`,
			testNS, serviceAccountName)

		_ = createNamespacedObject(ctx, serviceAccountYaml, testNS)
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
				testNS,
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
					  serviceAccountName: %s
					  runTemplateRef: 
					    name: my-run-template
					    namespace: %s
					    kind: ClusterRunTemplate
					  inputs:
					    key: val
					`,
					testNS, serviceAccountName, testNS)

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
					}).Should(Equal(2))

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
					  serviceAccountName: %s
					  runTemplateRef: 
					    name: my-run-template-does-not-match
					    namespace: %s
					    kind: ClusterRunTemplate
					`,
					testNS, serviceAccountName, testNS)

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
				`)

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
					  serviceAccountName: %s

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
					`, serviceAccountName, testNS)

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
					  serviceAccountName: %s

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
					`, serviceAccountName, testNS)

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
				`)

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
					  serviceAccountName: %s
					  runTemplateRef: 
					    name: my-run-template
					    namespace: %s
					    kind: ClusterRunTemplate
					`,
				testNS, serviceAccountName, testNS)

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

			Eventually(getRunnableTestStatus).Should(MatchFields(IgnoreExtras, Fields{
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

			Eventually(getRunnableTestStatus).Should(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal("Ready"),
				"Status": Equal(metav1.ConditionStatus("True")),
				"Reason": Equal("LifeIsGreat"),
			}))
		})
	})

	Describe("Latest stampedObject is the status", func() {
		BeforeEach(func() {
			runTemplateYaml := HereYamlF(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterRunTemplate
				metadata:
				  name: my-run-template
				spec:
				  outputs:
					test-status: status.conditions[?(@.type=="Succeeded")]
				  template:
					apiVersion: test.run/v1alpha1
					kind: TestObj
					metadata:
					  generateName: test-crd-
					  labels:
					    gen: "1"
					spec:
					  foo: $(runnable.spec.inputs.foo)$
				`)

			runTemplateDefinition = &unstructured.Unstructured{}
			err := yaml.Unmarshal([]byte(runTemplateYaml), runTemplateDefinition)
			Expect(err).NotTo(HaveOccurred())

			err = c.Create(ctx, runTemplateDefinition, &client.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := c.Delete(ctx, runTemplateDefinition)
			Expect(err).NotTo(HaveOccurred())
		})

		It("populates the runnable.Status.outputs properly", func() {
			listOpts := []client.ListOption{
				client.InNamespace(testNS),
				client.MatchingLabels(map[string]string{"carto.run/runnable-name": "my-runnable"}),
			}

			var runnableObject = &v1alpha1.Runnable{}
			By("creating the runnable", func() {
				runnableYaml := HereYamlF(`---
					apiVersion: carto.run/v1alpha1
					kind: Runnable
					metadata:
					  namespace: %s
					  name: my-runnable
					  labels:
					    some-val: first
					spec:
				      retentionPolicy: {maxFailedRuns: 10, maxSuccessfulRuns: 10}
					  serviceAccountName: %s
					  inputs:
				        foo: input-at-time-1
					  runTemplateRef: 
					    name: my-run-template
					    namespace: %s
					    kind: ClusterRunTemplate
					`,
					testNS, serviceAccountName, testNS)

				err := yaml.Unmarshal([]byte(runnableYaml), runnableObject)
				Expect(err).NotTo(HaveOccurred())

				err = c.Create(ctx, runnableObject, &client.CreateOptions{})
				Expect(err).NotTo(HaveOccurred())
			})
			By("showing that the Runnable status is unknown", func() {
				Eventually(func() (v1alpha1.RunnableStatus, error) {
					runnable := &v1alpha1.Runnable{}
					err := c.Get(ctx, client.ObjectKey{Namespace: testNS, Name: "my-runnable"}, runnable)
					return runnable.Status, err
				}).Should(
					MatchFields(IgnoreExtras,
						Fields{
							"Conditions": ContainElements(
								MatchFields(IgnoreExtras,
									Fields{
										"Type":   Equal("Ready"),
										"Status": Equal(metav1.ConditionUnknown),
									},
								),
								MatchFields(IgnoreExtras,
									Fields{
										"Type":   Equal("StampedObjectCondition"),
										"Status": Equal(metav1.ConditionUnknown),
									},
								),
								MatchFields(IgnoreExtras,
									Fields{
										"Type":   Equal("RunTemplateReady"),
										"Status": Equal(metav1.ConditionTrue),
									},
								),
							),
						},
					),
				)
			})

			var firstStampedObject resources.TestObj
			var secondStampedObject resources.TestObj

			By("seeing one stamped object", func() {
				testsList := &resources.TestObjList{}

				Eventually(func() ([]resources.TestObj, error) {
					err := c.List(ctx, testsList, listOpts...)
					return testsList.Items, err
				}).Should(HaveLen(1))

				firstStampedObject = testsList.Items[0]
			})
			By("changing the first stamped object's status to false", func() {
				firstStampedObject.Status = resources.TestStatus{
					ObservedGeneration: 1,
					Conditions: []metav1.Condition{
						{
							Type:               "Succeeded",
							Status:             "False",
							ObservedGeneration: 1,
							LastTransitionTime: metav1.Now(),
							Reason:             "FirstStampFailed",
							Message:            "not a happy first stamped object",
						},
					},
				}

				Eventually(func() error {
					return c.Status().Update(ctx, &firstStampedObject)
				}).ShouldNot(HaveOccurred())
			})
			By("seeing that the runnable status is false", func() {
				Eventually(func() ([]metav1.Condition, error) {
					runnable := &v1alpha1.Runnable{}
					err := c.Get(ctx, client.ObjectKey{Namespace: testNS, Name: "my-runnable"}, runnable)
					return runnable.Status.Conditions, err
				}).Should(
					ContainElements(
						MatchFields(IgnoreExtras,
							Fields{
								"Type":   Equal("Ready"),
								"Status": Equal(metav1.ConditionFalse),
							},
						),
						MatchFields(IgnoreExtras,
							Fields{
								"Type":   Equal("StampedObjectCondition"),
								"Status": Equal(metav1.ConditionFalse),
								"Reason": Equal("SucceededCondition"),
							},
						),
						MatchFields(IgnoreExtras,
							Fields{
								"Type":   Equal("RunTemplateReady"),
								"Status": Equal(metav1.ConditionTrue),
							},
						),
					),
				)
			})

			By("changing the input for the Runnable", func() {
				Eventually(func() error {
					err := c.Get(ctx, client.ObjectKey{
						Namespace: runnableObject.GetNamespace(),
						Name:      runnableObject.GetName(),
					}, runnableObject)

					if err != nil {
						return err
					}

					runnableObject.Spec.Inputs["foo"] = apiextensionsv1.JSON{
						Raw: []byte(`"input-at-time-2"`),
					}

					return c.Update(ctx, runnableObject, &client.UpdateOptions{})
				}).ShouldNot(HaveOccurred())
			})
			By("seeing that there is a new stampedObject", func() {
				testsList := &resources.TestObjList{}
				Eventually(func() ([]resources.TestObj, error) {
					err := c.List(ctx, testsList, listOpts...)
					return testsList.Items, err
				}).Should(HaveLen(2))

				for _, so := range testsList.Items {
					if so.Spec.Foo == "input-at-time-2" {
						secondStampedObject = so
						continue
					}
				}
			})
			By("seeing that the runnable status is unknown", func() {
				Eventually(func() ([]metav1.Condition, error) {
					runnable := &v1alpha1.Runnable{}
					err := c.Get(ctx, client.ObjectKey{Namespace: testNS, Name: "my-runnable"}, runnable)
					return runnable.Status.Conditions, err
				}).Should(
					ContainElements(
						MatchFields(IgnoreExtras,
							Fields{
								"Type":   Equal("Ready"),
								"Status": Equal(metav1.ConditionUnknown),
							},
						),
						MatchFields(IgnoreExtras,
							Fields{
								"Type":   Equal("StampedObjectCondition"),
								"Status": Equal(metav1.ConditionUnknown),
								"Reason": Equal("Unknown"),
							},
						),
						MatchFields(IgnoreExtras,
							Fields{
								"Type":   Equal("RunTemplateReady"),
								"Status": Equal(metav1.ConditionTrue),
							},
						),
					),
				)
			})
			By("changing the second stampedObject's status to True", func() {
				secondStampedObject.Status = resources.TestStatus{
					ObservedGeneration: 1,
					Conditions: []metav1.Condition{
						{
							Type:               "Succeeded",
							Status:             "True",
							ObservedGeneration: 1,
							LastTransitionTime: metav1.Now(),
							Reason:             "SecondStampWorked",
							Message:            "happy second stamped object",
						},
					},
				}

				Eventually(func() error {
					return c.Status().Update(ctx, &secondStampedObject)
				}).ShouldNot(HaveOccurred())
			})
			By("seeing that the runnable status is true", func() {
				Eventually(func() ([]metav1.Condition, error) {
					runnable := &v1alpha1.Runnable{}
					err := c.Get(ctx, client.ObjectKey{Namespace: testNS, Name: "my-runnable"}, runnable)
					return runnable.Status.Conditions, err
				}).Should(
					ContainElements(
						MatchFields(IgnoreExtras,
							Fields{
								"Type":   Equal("Ready"),
								"Status": Equal(metav1.ConditionTrue),
							},
						),
						MatchFields(IgnoreExtras,
							Fields{
								"Type":   Equal("StampedObjectCondition"),
								"Status": Equal(metav1.ConditionTrue),
								"Reason": Equal("SucceededCondition"),
							},
						),
						MatchFields(IgnoreExtras,
							Fields{
								"Type":   Equal("RunTemplateReady"),
								"Status": Equal(metav1.ConditionTrue),
							},
						),
					),
				)

			})

			By("changing the input for the Runnable", func() {
				Eventually(func() error {
					err := c.Get(ctx, client.ObjectKey{
						Namespace: runnableObject.GetNamespace(),
						Name:      runnableObject.GetName(),
					}, runnableObject)

					if err != nil {
						return err
					}

					runnableObject.Spec.Inputs["foo"] = apiextensionsv1.JSON{
						Raw: []byte(`"input-at-time-3"`),
					}

					return c.Update(ctx, runnableObject, &client.UpdateOptions{})
				}).ShouldNot(HaveOccurred())

			})
			By("seeing that there is a new stampedObject", func() {
				testsList := &resources.TestObjList{}
				Eventually(func() ([]resources.TestObj, error) {
					err := c.List(ctx, testsList, listOpts...)
					return testsList.Items, err
				}).Should(HaveLen(3))
			})
			By("seeing that the runnable status is unknown", func() {
				Eventually(func() ([]metav1.Condition, error) {
					runnable := &v1alpha1.Runnable{}
					err := c.Get(ctx, client.ObjectKey{Namespace: testNS, Name: "my-runnable"}, runnable)
					return runnable.Status.Conditions, err
				}).Should(
					ContainElements(
						MatchFields(IgnoreExtras,
							Fields{
								"Type":   Equal("Ready"),
								"Status": Equal(metav1.ConditionUnknown),
							},
						),
						MatchFields(IgnoreExtras,
							Fields{
								"Type":   Equal("StampedObjectCondition"),
								"Status": Equal(metav1.ConditionUnknown),
								"Reason": Equal("Unknown"),
							},
						),
						MatchFields(IgnoreExtras,
							Fields{
								"Type":   Equal("RunTemplateReady"),
								"Status": Equal(metav1.ConditionTrue),
							},
						),
					),
				)
			})
		})
	})
})
