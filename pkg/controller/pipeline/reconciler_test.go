package pipeline_test

import (
	"context"
	"errors"

	. "github.com/MakeNowJust/heredoc/dot"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	. "github.com/vmware-tanzu/cartographer/pkg/controller/pipeline"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var _ = Describe("Reconcile", func() {
	var (
		out        *Buffer
		ctx        context.Context
		reconciler *Reconciler
		request    ctrl.Request
		repository *repositoryfakes.FakeRepository
	)

	BeforeEach(func() {
		out = NewBuffer()
		logger := zap.New(zap.WriteTo(out))
		ctx = logr.NewContext(context.Background(), logger)
		repository = &repositoryfakes.FakeRepository{}
		reconciler = &Reconciler{Repository: repository}

		request = ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      "my-pipeline",
				Namespace: "my-namespace",
			},
		}
	})

	Context("reconcile a new valid Pipeline", func() {
		BeforeEach(func() {
			repository.GetPipelineStub = func(name, namespace string) (*v1alpha1.Pipeline, error) {
				pipeline := &v1alpha1.Pipeline{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pipeline",
						APIVersion: "carto.run/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace,
					},
					Spec: v1alpha1.PipelineSpec{
						RunTemplate: v1alpha1.TemplateReference{
							Kind:      "RunTemplate",
							Name:      "my-run-template",
							Namespace: "ns1",
						},
					},
				}
				return pipeline, nil
			}
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
			})

			It("stamps out the resource from the template", func() {
				_, _ = reconciler.Reconcile(ctx, request)

				Expect(repository.GetPipelineCallCount()).To(Equal(1))
				actualName, actualNamespace := repository.GetPipelineArgsForCall(0)
				Expect(actualName).To(Equal("my-pipeline"))
				Expect(actualNamespace).To(Equal("my-namespace"))

				Expect(repository.GetTemplateCallCount()).To(Equal(1))
				Expect(repository.GetTemplateArgsForCall(0)).To(MatchFields(IgnoreExtras,
					Fields{
						"Kind": Equal("RunTemplate"),
						"Name": Equal("my-run-template"),
					},
				))

				Expect(repository.AlwaysCreateOnClusterCallCount()).To(Equal(1))
				Expect(repository.AlwaysCreateOnClusterCallCount()).To(Equal(1))
				stamped := repository.AlwaysCreateOnClusterArgsForCall(0).Object
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

			It("Starts and Finishes cleanly", func() {
				_, err := reconciler.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())

				Expect(out).To(Say(`"msg":"started","name":"my-pipeline","namespace":"my-namespace"`))
				Expect(out).To(Say(`"msg":"finished","name":"my-pipeline","namespace":"my-namespace"`))
			})
		})

		Context("the RunTemplate cannot be fetched", func() {
			BeforeEach(func() {
				repository.GetTemplateStub = func(reference v1alpha1.TemplateReference) (templates.Template, error) {
					return nil, errors.New("Errol mcErrorFace")
				}

				repository.StatusUpdateStub = func(object client.Object) error {
					return nil
				}
			})

			It("logs the error", func() {
				result, err := reconciler.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))

				Expect(out).To(Say(`"msg":"could not get RunTemplate 'my-run-template'"`))
				Expect(out).To(Say(`"error":"Errol mcErrorFace"`))
			})

			It("sets the status to something clear", func() {
				_, err := reconciler.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())

				Expect(repository.StatusUpdateCallCount()).To(Equal(1))
				statusObject, ok := repository.StatusUpdateArgsForCall(0).(*v1alpha1.Pipeline)

				Expect(ok).To(BeTrue())

				Expect(statusObject.GetObjectKind().GroupVersionKind()).To(MatchFields(0, Fields{
					"Kind":    Equal("Pipeline"),
					"Version": Equal("v1alpha1"),
					"Group":   Equal("carto.run"),
				}))

				Expect(statusObject.Name).To(Equal("my-pipeline"))

				Expect(statusObject.Status.Conditions).To(ContainElements(
					MatchFields(IgnoreExtras, Fields{
						"Type":    Equal("Ready"),
						"Status":  Equal(metav1.ConditionFalse),
						"Reason":  Equal("ConditionsUnmet"),
						"Message": Equal("not all conditions are met"),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Type":    Equal("RunTemplateReady"),
						"Status":  Equal(metav1.ConditionFalse),
						"Reason":  Equal("RunTemplateNotFound"),
						"Message": Equal("could not get RunTemplate 'my-run-template': Errol mcErrorFace"),
					}),
				))
			})

			It("Starts and Finishes cleanly", func() {
				_, err := reconciler.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())

				Expect(out).To(Say(`"msg":"started","name":"my-pipeline","namespace":"my-namespace"`))
				Expect(out).To(Say(`"msg":"finished","name":"my-pipeline","namespace":"my-namespace"`))
			})

			Context("updating the status fails", func() {

			})
		})
	})

	Context("the pipeline goes away", func() {
		BeforeEach(func() {
			repository.GetPipelineStub = func(name, namespace string) (*v1alpha1.Pipeline, error) {
				return nil, kerrors.NewNotFound(
					schema.GroupResource{
						Group:    "carto.run",
						Resource: "Pipeline",
					},
					"my-pipeline",
				)
			}
		})

		It("considers the reconcile complete", func() {
			result, err := reconciler.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})

		It("logs that we saw the pipeline go away", func() {
			_, err := reconciler.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())

			Expect(out).To(Say(`"msg":"pipeline no longer exists","name":"my-pipeline","namespace":"my-namespace"`))
		})

		It("Starts and Finishes cleanly", func() {
			_, err := reconciler.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())

			Eventually(out).Should(Say(`"msg":"started","name":"my-pipeline","namespace":"my-namespace"`))
			Eventually(out).Should(Say(`"msg":"finished","name":"my-pipeline","namespace":"my-namespace"`))
		})
	})
})
