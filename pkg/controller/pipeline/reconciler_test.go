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

package pipeline_test

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
	pipeline2 "github.com/vmware-tanzu/cartographer/pkg/realizer/pipeline"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/controller/pipeline"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/pipeline/pipelinefakes"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
)

var _ = Describe("Reconcile", func() {
	var (
		out        *Buffer
		ctx        context.Context
		reconciler pipeline.Reconciler
		request    ctrl.Request
		repository *repositoryfakes.FakeRepository
		realizer   *pipelinefakes.FakeRealizer
	)

	BeforeEach(func() {
		out = NewBuffer()
		logger := zap.New(zap.WriteTo(out))
		ctx = logr.NewContext(context.Background(), logger)
		repository = &repositoryfakes.FakeRepository{}
		realizer = &pipelinefakes.FakeRealizer{}

		reconciler = pipeline.NewReconciler(repository, realizer)

		request = ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      "my-pipeline",
				Namespace: "my-namespace",
			},
		}
	})

	Context("reconcile a new valid Pipeline", func() {
		BeforeEach(func() {
			repository.GetPipelineReturns(&v1alpha1.Pipeline{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pipeline",
					APIVersion: "carto.run/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-pipeline",
					Namespace: "my-namespace",
				},
				Spec: v1alpha1.PipelineSpec{
					RunTemplateRef: v1alpha1.TemplateReference{
						Kind:      "RunTemplateRef",
						Name:      "my-run-template",
						Namespace: "ns1",
					},
				},
			}, nil)

		})

		Context("no outputs were returned from the realizer", func() {
			BeforeEach(func() {
				realizer.RealizeReturns(pipeline2.RunTemplateReadyCondition(), nil)
			})
			It("fetches the pipeline", func() {
				_, _ = reconciler.Reconcile(ctx, request)

				Expect(repository.GetPipelineCallCount()).To(Equal(1))
				actualName, actualNamespace := repository.GetPipelineArgsForCall(0)
				Expect(actualName).To(Equal("my-pipeline"))
				Expect(actualNamespace).To(Equal("my-namespace"))
			})

			It("Starts and Finishes cleanly", func() {
				_, err := reconciler.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())

				Expect(out).To(Say(`"msg":"started","name":"my-pipeline","namespace":"my-namespace"`))
				Expect(out).To(Say(`"msg":"finished","name":"my-pipeline","namespace":"my-namespace"`))
			})

			It("Updates the status with a Happy Condition", func() {
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
						"Type":   Equal("RunTemplateReady"),
						"Status": Equal(metav1.ConditionTrue),
						"Reason": Equal("Ready"),
					}),
				))

			})
		})

		Context("outputs are returned from the realizer", func() {
			BeforeEach(func() {
				realizer.RealizeReturns(
					pipeline2.RunTemplateReadyCondition(),
					templates.Outputs{
						"an-output": apiextensionsv1.JSON{Raw: []byte(`"the value"`)},
					})
			})

			It("Updates the status with the outputs", func() {
				_, err := reconciler.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())

				Expect(repository.StatusUpdateCallCount()).To(Equal(1))
				statusObject, ok := repository.StatusUpdateArgsForCall(0).(*v1alpha1.Pipeline)
				Expect(ok).To(BeTrue())

				Expect(statusObject.Status.Outputs).To(HaveLen(1))
				Expect(statusObject.Status.Outputs["an-output"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"the value"`)}))

			})
		})

		Context("updating the status fails", func() {
			BeforeEach(func() {
				realizer.RealizeReturns(pipeline2.RunTemplateReadyCondition(), nil)
				repository.StatusUpdateReturns(errors.New("bad status update error"))
			})

			It("Starts and Finishes cleanly", func() {
				_, _ = reconciler.Reconcile(ctx, request)
				Expect(out).To(Say(`"msg":"started","name":"my-pipeline","namespace":"my-namespace"`))
				Expect(out).To(Say(`"msg":"finished","name":"my-pipeline","namespace":"my-namespace"`))
			})

			It("returns a status error", func() {
				result, err := reconciler.Reconcile(ctx, request)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("update pipeline status"))
				Expect(result).To(Equal(ctrl.Result{}))
			})
		})
	})

	Context("the pipeline goes away", func() {
		BeforeEach(func() {
			repository.GetPipelineReturns(nil, kerrors.NewNotFound(
				schema.GroupResource{
					Group:    "carto.run",
					Resource: "Pipeline",
				},
				"my-pipeline",
			))
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

	Context("the pipeline fetch is in error", func() {
		BeforeEach(func() {
			repository.GetPipelineReturns(nil, errors.New("very bad pipeline"))
		})

		It("returns a helpful error", func() {
			_, err := reconciler.Reconcile(ctx, request)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("very bad pipeline"))
		})
	})
})
