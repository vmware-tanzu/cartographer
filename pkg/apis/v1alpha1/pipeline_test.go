package v1alpha1_test

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

var _ = Describe("Pipeline", func() {
	Describe("PipelineSpec", func() {
		var (
			pipelineSpec     v1alpha1.PipelineSpec
			pipelineSpecType reflect.Type
		)

		BeforeEach(func() {
			pipelineSpecType = reflect.TypeOf(pipelineSpec)
		})

		It("requires runTemplate", func() {
			componentsField, found := pipelineSpecType.FieldByName("RunTemplateRef")
			Expect(found).To(BeTrue())
			jsonValue := componentsField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("runTemplate"))
			Expect(jsonValue).NotTo(ContainSubstring("omitempty"))
		})
	})

	Describe("TemplateReference", func() {
		var (
			templateReference     v1alpha1.TemplateReference
			templateReferenceType reflect.Type
		)

		BeforeEach(func() {
			templateReferenceType = reflect.TypeOf(templateReference)
		})

		It("requires a name", func() {
			nameField, found := templateReferenceType.FieldByName("Name")
			Expect(found).To(BeTrue())
			jsonValue := nameField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("name"))
			Expect(jsonValue).NotTo(ContainSubstring("omitempty"))
		})

		It("requires a kind", func() {
			kindField, found := templateReferenceType.FieldByName("Kind")
			Expect(found).To(BeTrue())
			jsonValue := kindField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("kind"))
			Expect(jsonValue).NotTo(ContainSubstring("omitempty"))
		})

		It("requires a namespace", func() {
			namespaceField, found := templateReferenceType.FieldByName("Namespace")
			Expect(found).To(BeTrue())
			jsonValue := namespaceField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("namespace"))
			Expect(jsonValue).NotTo(ContainSubstring("omitempty"))
		})
	})
})
