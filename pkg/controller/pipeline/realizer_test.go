package pipeline_test

import (
	"errors"

	. "github.com/MakeNowJust/heredoc/dot"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	. "github.com/vmware-tanzu/cartographer/pkg/controller/pipeline"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

var _ = Describe("Reconcile", func() {
	var (
		out        *Buffer
		repository *repositoryfakes.FakeRepository
		logger     logr.Logger
		realizer   Realizer
		pipeline   *v1alpha1.Pipeline
	)

	BeforeEach(func() {
		out = NewBuffer()
		logger = zap.New(zap.WriteTo(out))
		repository = &repositoryfakes.FakeRepository{}
		realizer = NewRealizer()
	})

	Context("with a valid RunTemplate", func() {
		BeforeEach(func() {
			repository.GetTemplateStub = func(reference v1alpha1.TemplateReference) (templates.Template, error) {
				template := templates.NewRunTemplateModel(&v1alpha1.RunTemplate{
					Spec: v1alpha1.RunTemplateSpec{
						Template: runtime.RawExtension{
							Raw: []byte(D(`{
								"apiVersion": "v1",
								"kind": "ConfigMap",
								"metadata": { "generateName": "my-stamped-resource-" },
								"data": { "has": "data" }
							}`)),
						},
					},
				})
				return template, nil
			}

			pipeline = &v1alpha1.Pipeline{
				Spec: v1alpha1.PipelineSpec{
					RunTemplate: v1alpha1.TemplateReference{
						Kind:      "RunTemplate",
						Name:      "my-template",
						Namespace: "some-ns",
					},
				},
			}

		})

		It("stamps out the resource from the template", func() {
			_ = realizer.Realize(pipeline, logger, repository)

			Expect(repository.GetTemplateCallCount()).To(Equal(1))
			Expect(repository.GetTemplateArgsForCall(0)).To(MatchFields(IgnoreExtras,
				Fields{
					"Kind": Equal("RunTemplate"),
					"Name": Equal("my-template"),
				},
			))

			Expect(repository.CreateCallCount()).To(Equal(1))
			Expect(repository.CreateCallCount()).To(Equal(1))
			stamped := repository.CreateArgsForCall(0).Object
			Expect(stamped).To(
				MatchKeys(IgnoreExtras, Keys{
					"metadata": MatchKeys(IgnoreExtras, Keys{
						"generateName": Equal("my-stamped-resource-"),
					}),
					"apiVersion": Equal("v1"),
					"kind":       Equal("ConfigMap"),
					"data": MatchKeys(IgnoreExtras, Keys{
						"has": Equal("data"),
					}),
				}),
			)
		})

		It("returns a happy condition", func() {
			condition := realizer.Realize(pipeline, logger, repository)
			Expect(*condition).To(
				MatchFields(IgnoreExtras, Fields{
					"Type":   Equal("RunTemplateReady"),
					"Status": Equal(metav1.ConditionTrue),
					"Reason": Equal("Ready"),
				}),
			)
		})

		Context("error on Create", func() {
			BeforeEach(func() {
				repository.CreateReturns(errors.New("some bad error"))
			})

			It("logs the error", func() {
				_ = realizer.Realize(pipeline, logger, repository)

				Expect(out).To(Say(`"msg":"could not create object"`))
				Expect(out).To(Say(`"error":"some bad error"`))
			})

			It("returns a condition stating that it failed to create", func() {
				condition := realizer.Realize(pipeline, logger, repository)

				Expect(*condition).To(
					MatchFields(IgnoreExtras, Fields{
						"Type":    Equal("RunTemplateReady"),
						"Status":  Equal(metav1.ConditionFalse),
						"Reason":  Equal("StampedObjectRejectedByAPIServer"),
						"Message": Equal("could not create object: some bad error"),
					}),
				)
			})

		})

	})

	Context("the RunTemplate cannot be fetched", func() {
		BeforeEach(func() {
			repository.GetTemplateStub = func(reference v1alpha1.TemplateReference) (templates.Template, error) {
				return nil, errors.New("Errol mcErrorFace")
			}

			pipeline = &v1alpha1.Pipeline{
				Spec: v1alpha1.PipelineSpec{
					// FIXME should this be `RunTemplateRef`
					RunTemplate: v1alpha1.TemplateReference{
						Kind:      "RunTemplate",
						Name:      "my-template",
						Namespace: "some-ns",
					},
				},
			}
		})

		It("logs the error", func() {
			_ = realizer.Realize(pipeline, logger, repository)

			Expect(out).To(Say(`"msg":"could not get RunTemplate 'my-template'"`))
			Expect(out).To(Say(`"error":"Errol mcErrorFace"`))
		})

		It("return the condition for a missing RunTemplate", func() {
			condition := realizer.Realize(pipeline, logger, repository)

			Expect(*condition).To(
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal("RunTemplateReady"),
					"Status":  Equal(metav1.ConditionFalse),
					"Reason":  Equal("RunTemplateNotFound"),
					"Message": Equal("could not get RunTemplate 'my-template': Errol mcErrorFace"),
				}),
			)
		})
	})
})
