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

package pipeline_service_test

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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	. "github.com/vmware-tanzu/cartographer/pkg/utils"
	"github.com/vmware-tanzu/cartographer/tests/resources"
)

var _ = Describe("Stamping a resource on Pipeline Creation", func() {
	var (
		ctx                   context.Context
		pipelineDefinition    *unstructured.Unstructured
		runTemplateDefinition *unstructured.Unstructured
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	getPipelineTestStatus := func() (metav1.Condition, error) {
		pipeline := &v1alpha1.Pipeline{}
		err := c.Get(ctx, client.ObjectKey{Name: "my-pipeline", Namespace: testNS}, pipeline)
		if err != nil {
			return metav1.Condition{}, err
		}
		testStatusCondition := &metav1.Condition{}
		testStatusConditionJson := pipeline.Status.Outputs["test-status"].Raw
		err = json.Unmarshal(testStatusConditionJson, testStatusCondition)
		return *testStatusCondition, err
	}

	Describe("when a ClusterRunTemplate that produces a Resource leverages a Pipeline field", func() {
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
						  values: [$(pipeline.spec.inputs.key)$]
				`,
				testNS, testNS,
			)

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

		Context("and a Pipeline matches the RunTemplateRef", func() {
			BeforeEach(func() {
				pipelineYaml := HereYamlF(`---
					apiVersion: carto.run/v1alpha1
					kind: Pipeline
					metadata:
					  namespace: %s
					  name: my-pipeline
					spec:
					  runTemplateRef: 
					    name: my-run-template
					    namespace: %s
					    kind: ClusterRunTemplate
					  inputs:
					    key: val
					`,
					testNS, testNS)

				pipelineDefinition = &unstructured.Unstructured{}
				err := yaml.Unmarshal([]byte(pipelineYaml), pipelineDefinition)
				Expect(err).NotTo(HaveOccurred())

				err = c.Create(ctx, pipelineDefinition, &client.CreateOptions{})
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				err := c.Delete(ctx, pipelineDefinition)
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

			Context("and the Pipeline object is updated", func() {
				BeforeEach(func() {
					Expect(AlterFieldOfNestedStringMaps(pipelineDefinition.Object, "spec.inputs.key", "new-val")).To(Succeed())
					Expect(c.Update(ctx, pipelineDefinition, &client.UpdateOptions{})).To(Succeed())
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
					}, "2s").Should(BeNumerically("<=", 2))

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

					// Ensure that first object has been stamped, and status update reconcile of pipeline has occurred
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

		Context("a Pipeline that does not match the RunTemplateRef", func() {
			BeforeEach(func() {
				pipelineYaml := HereYamlF(`---
					apiVersion: carto.run/v1alpha1
					kind: Pipeline
					metadata:
					  namespace: %s
					  name: my-pipeline
					spec:
					  runTemplateRef: 
					    name: my-run-template-does-not-match
					    namespace: %s
					    kind: ClusterRunTemplate
					`,
					testNS, testNS)

				pipelineDefinition = &unstructured.Unstructured{}
				err := yaml.Unmarshal([]byte(pipelineYaml), pipelineDefinition)
				Expect(err).NotTo(HaveOccurred())

				err = c.Create(ctx, pipelineDefinition, &client.CreateOptions{})
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				err := c.Delete(ctx, pipelineDefinition)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Does not stamp a new Resource", func() {
				resourceList := &v1.ConfigMapList{}

				Consistently(func() (int, error) {
					err := c.List(ctx, resourceList, &client.ListOptions{Namespace: testNS})
					return len(resourceList.Items), err
				}).Should(Equal(0))
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
					kind: Test
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

			pipelineYaml := HereYamlF(`---
					apiVersion: carto.run/v1alpha1
					kind: Pipeline
					metadata:
					  namespace: %s
					  name: my-pipeline
					  labels:
					    some-val: first
					spec:
					  runTemplateRef: 
					    name: my-run-template
					    namespace: %s
					    kind: ClusterRunTemplate
					`,
				testNS, testNS)

			pipelineDefinition = &unstructured.Unstructured{}
			err = yaml.Unmarshal([]byte(pipelineYaml), pipelineDefinition)
			Expect(err).NotTo(HaveOccurred())

			err = c.Create(ctx, pipelineDefinition, &client.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := c.Delete(ctx, pipelineDefinition)
			Expect(err).NotTo(HaveOccurred())

			err = c.Delete(ctx, runTemplateDefinition)
			Expect(err).NotTo(HaveOccurred())
		})

		It("populates the pipeline.Status.outputs properly", func() {
			opts := []client.ListOption{
				client.InNamespace(testNS),
				client.MatchingLabels(map[string]string{"carto.run/pipeline-name": "my-pipeline"}),
			}

			testsList := &resources.TestList{}

			Eventually(func() ([]resources.Test, error) {
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

			Eventually(getPipelineTestStatus, "10s").Should(MatchFields(IgnoreExtras, Fields{
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

			Consistently(getPipelineTestStatus).Should(MatchFields(IgnoreExtras, Fields{
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

			Eventually(getPipelineTestStatus, "10s").Should(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal("Ready"),
				"Status": Equal(metav1.ConditionStatus("True")),
				"Reason": Equal("LifeIsGreat"),
			}))
		})
	})

	Describe("Multiple objects created", func() {
		BeforeEach(func() {
			pipelineYaml := HereYamlF(`---
					apiVersion: carto.run/v1alpha1
					kind: Pipeline
					metadata:
					  namespace: %s
					  name: my-pipeline
					  labels:
					    some-val: first
					spec:
					  runTemplateRef: 
					    name: my-run-template
					    namespace: %s
					    kind: ClusterRunTemplate
					`,
				testNS, testNS)

			pipelineDefinition = &unstructured.Unstructured{}
			err := yaml.Unmarshal([]byte(pipelineYaml), pipelineDefinition)
			Expect(err).NotTo(HaveOccurred())

			err = c.Create(ctx, pipelineDefinition, &client.CreateOptions{})
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
					kind: Test
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
				client.MatchingLabels(map[string]string{"carto.run/pipeline-name": "my-pipeline"}),
			}

			testsList := &resources.TestList{}

			Eventually(func() ([]resources.Test, error) {
				err := c.List(ctx, testsList, opts...)
				return testsList.Items, err
			}).Should(HaveLen(1))

			// This is in order to ensure gen 1 object and gen 2 object have different creationTimestamps
			time.Sleep(time.Second)

			Expect(AlterFieldOfNestedStringMaps(runTemplateDefinition.Object, "spec.template.metadata.labels.gen", "2")).To(Succeed())

			err = c.Update(ctx, runTemplateDefinition, &client.UpdateOptions{})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() ([]resources.Test, error) {
				err := c.List(ctx, testsList, opts...)
				return testsList.Items, err
			}).Should(HaveLen(2))

			// This is in order to ensure gen 2 object and gen 3 object have different creationTimestamps
			// Gen 3 object is needed to demonstrate behaviour when the most recently submitted is not successful
			time.Sleep(time.Second)

			Expect(AlterFieldOfNestedStringMaps(runTemplateDefinition.Object, "spec.template.metadata.labels.gen", "3")).To(Succeed())

			err = c.Update(ctx, runTemplateDefinition, &client.UpdateOptions{})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() ([]resources.Test, error) {
				err := c.List(ctx, testsList, opts...)
				return testsList.Items, err
			}).Should(HaveLen(3))
		})

		AfterEach(func() {
			err := c.Delete(ctx, pipelineDefinition)
			Expect(err).NotTo(HaveOccurred())

			err = c.Delete(ctx, runTemplateDefinition)
			Expect(err).NotTo(HaveOccurred())
		})

		It("populates the pipeline.Status.outputs properly", func() {
			By("updating pipeline status based on the most recently submitted and successful object")
			opts := []client.ListOption{
				client.InNamespace(testNS),
				client.MatchingLabels(map[string]string{"gen": "2"}),
			}

			testsList := &resources.TestList{}

			Eventually(func() ([]resources.Test, error) {
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

			Eventually(getPipelineTestStatus, "10s").Should(MatchFields(IgnoreExtras, Fields{
				"Message": Equal("this is generation 2"),
			}))

			By("not updating pipeline status based on the less recently submitted and successful objects")
			opts = []client.ListOption{
				client.InNamespace(testNS),
				client.MatchingLabels(map[string]string{"gen": "1"}),
			}

			Eventually(func() ([]resources.Test, error) {
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

			Consistently(getPipelineTestStatus, "1s").Should(MatchFields(IgnoreExtras, Fields{
				"Message": And(
					Equal("this is generation 2"),
					Not(Equal("but this is earlier generation 1"))),
			}))
		})
	})
})
