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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	. "github.com/vmware-tanzu/cartographer/pkg/utils"
)

var _ = Describe("Stamping a resource on Pipeline Creation", func() {
	var (
		ctx                   = context.Background()
		pipelineDefinition    *unstructured.Unstructured
		runTemplateDefinition *unstructured.Unstructured
	)

	// TODO: ask team about inspecting template.metadata.name and warning/blocking for UX
	Context("a RunTemplate that produces a Resource", func() {
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
				    kind: ConfigMap
				    metadata:
					  generateName: my-stamped-resource-
				    data:
					  has: data
				`,
				testNS,
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

		Context("a Pipeline that matches the RunTemplateRef", func() {
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

			It("Stamps a new Resource", func() {
				resourceList := &v1.ConfigMapList{}

				Eventually(func() (int, error) {
					err := c.List(ctx, resourceList, &client.ListOptions{Namespace: testNS})
					return len(resourceList.Items), err
				}).Should(BeNumerically(">", 0))

				// TODO: comment this in and make it pass
				//Consistently(func() (int, error) {
				//	err := c.List(ctx, resourceList, &client.ListOptions{Namespace: testNS})
				//	return len(resourceList.Items), err
				//}, "5s").Should(BeNumerically("<=",1))

				Expect(resourceList.Items[0].Name).To(ContainSubstring("my-stamped-resource-"))
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

				Consistently(func() ([]v1.ConfigMap, error) {
					err := c.List(ctx, resourceList, &client.ListOptions{Namespace: testNS})
					return resourceList.Items, err
				}, "5s").Should(HaveLen(0))
			})
		})
	})
})
