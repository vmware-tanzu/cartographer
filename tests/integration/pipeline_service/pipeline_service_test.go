package pipeline_service_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/vmware-tanzu/cartographer/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
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

		Context("a Pipeline that matches the RunTemplate", func() {
			BeforeEach(func() {
				pipelineYaml := HereYamlF(`---
					apiVersion: carto.run/v1alpha1
					kind: Pipeline
					metadata:
					  namespace: %s
					  name: my-pipeline
					spec:
					  runTemplate: 
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

				Eventually(func() ([]v1.ConfigMap, error) {
					err := c.List(ctx, resourceList, &client.ListOptions{Namespace: testNS})
					return resourceList.Items, err
				}).Should(HaveLen(1))

				Expect(resourceList.Items[0].Name).To(ContainSubstring("my-stamped-resource-"))
			})
		})

		Context("a Pipeline that does not match the RunTemplate", func() {
			XIt("Does not stamp a new Job", func() {})
		})
	})
})
