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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	. "github.com/vmware-tanzu/cartographer/pkg/utils"
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

	// TODO: ask team about inspecting template.metadata.name and warning/blocking for UX
	Context("when a RunTemplate that produces a Resource leverages a Pipeline field", func() {
		BeforeEach(func() {
			runTemplateYaml := HereYamlF(`
				---
				apiVersion: carto.run/v1alpha1
				kind: RunTemplate
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
						  values: [$(pipeline.metadata.labels.some-val)$]
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
					  labels:
					    some-val: first
					spec:
					  runTemplateRef: 
					    name: my-run-template
					    namespace: %s
					    kind: RunTemplate
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
				Expect(resourceList.Items[0].Spec.ScopeSelector.MatchExpressions[0].Values).To(ConsistOf("first"))
			})

			Context("and the Pipeline object is updated", func() {
				BeforeEach(func() {
					Expect(AlterFieldOfNestedStringMaps(pipelineDefinition.Object, "metadata.labels.some-val", "second")).To(Succeed())
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
						"first":  Not(BeNil()),
						"second": Not(BeNil()),
					}))
				})
			})

			Context("and the RunTemplate object is updated", func() {
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
					    kind: RunTemplate
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
})
