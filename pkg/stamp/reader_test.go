package stamp_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/stamp"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

type noInputImpl struct{}

func (noInputImpl) GetDeployment() *templates.SourceInput {
	return nil
}

var _ = Describe("Reader", func() {

	Context("using a source reader", func() {
		Context("where the evaluator can return a value", func() {
			It("returns the output", func() {

			})

		})

		Context("where the evaluator can not return a value", func() {
			It("returns a nil output", func() {

			})
			It("returns an error", func() {

			})
		})
	})

	Context("using an image reader", func() {
		var (
			template *v1alpha1.ClusterImageTemplate
			reader   stamp.Reader
		)

		BeforeEach(func() {
			template = &v1alpha1.ClusterImageTemplate{
				Spec: v1alpha1.ImageTemplateSpec{
					TemplateSpec: v1alpha1.TemplateSpec{},
					ImagePath:    ".data.image",
				},
			}

			var err error
			reader, err = stamp.NewReader(template, noInputImpl{})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("where the evaluator can return a value", func() {
			var stampedObject *unstructured.Unstructured

			BeforeEach(func() {
				unstructuredContent := map[string]interface{}{
					"data": map[string]interface{}{
						"image": "my-image",
					},
				}

				stampedObject = &unstructured.Unstructured{}
				stampedObject.SetUnstructuredContent(unstructuredContent)
			})

			It("returns the output", func() {
				output, err := reader.GetOutput(stampedObject)
				Expect(err).NotTo(HaveOccurred())
				Expect(output.Image).To(Equal("my-image"))
			})
		})

		Context("where the evaluator can not return a value", func() {
			var stampedObject *unstructured.Unstructured

			BeforeEach(func() {
				unstructuredContent := map[string]interface{}{}

				stampedObject = &unstructured.Unstructured{}
				stampedObject.SetUnstructuredContent(unstructuredContent)
			})

			It("returns a nil output", func() {
				output, _ := reader.GetOutput(stampedObject)
				Expect(output).To(BeNil())
			})

			It("returns an error", func() {
				_, err := reader.GetOutput(stampedObject)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to evaluate json path"))
				Expect(err.Error()).To(ContainSubstring(".data.image"))
			})
		})
	})

	Context("using a config reader", func() {
		var (
			template *v1alpha1.ClusterConfigTemplate
			reader   stamp.Reader
		)

		BeforeEach(func() {
			template = &v1alpha1.ClusterConfigTemplate{
				Spec: v1alpha1.ConfigTemplateSpec{
					TemplateSpec: v1alpha1.TemplateSpec{},
					ConfigPath:   ".data.config",
				},
			}

			var err error
			reader, err = stamp.NewReader(template, noInputImpl{})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("where the evaluator can return a value", func() {
			var stampedObject *unstructured.Unstructured

			BeforeEach(func() {
				unstructuredContent := map[string]interface{}{
					"data": map[string]interface{}{
						"config": "my-config",
					},
				}

				stampedObject = &unstructured.Unstructured{}
				stampedObject.SetUnstructuredContent(unstructuredContent)
			})

			It("returns the output", func() {
				output, err := reader.GetOutput(stampedObject)
				Expect(err).NotTo(HaveOccurred())
				Expect(output.Config).To(Equal("my-config"))
			})
		})

		Context("where the evaluator can not return a value", func() {
			var stampedObject *unstructured.Unstructured

			BeforeEach(func() {
				unstructuredContent := map[string]interface{}{}

				stampedObject = &unstructured.Unstructured{}
				stampedObject.SetUnstructuredContent(unstructuredContent)
			})

			It("returns a nil output", func() {
				output, _ := reader.GetOutput(stampedObject)
				Expect(output).To(BeNil())
			})

			It("returns an error", func() {
				_, err := reader.GetOutput(stampedObject)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to evaluate json path"))
				Expect(err.Error()).To(ContainSubstring(".data.config"))
			})
		})
	})

	Context("using a deployment reader", func() {
		Context("where the evaluator can return a value", func() {
			It("returns the output", func() {

			})
		})

		Context("where the evaluator can not return a value", func() {
			It("returns a nil output", func() {

			})
			It("returns an error", func() {

			})
		})
	})

	Context("using a no output reader", func() {
		It("returns an empty output", func() {

		})

		It("does not error", func() {

		})
	})
})
