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

package delivery_test

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/controller/delivery"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/dependency/dependencyfakes"
)

var _ = Describe("delivery reconciler", func() {
	var (
		repo              *repositoryfakes.FakeRepository
		dependencyTracker *dependencyfakes.FakeDependencyTracker
		reconciler        delivery.Reconciler
		ctx               context.Context
		req               reconcile.Request
		out               *Buffer
	)

	BeforeEach(func() {
		repo = &repositoryfakes.FakeRepository{}

		dependencyTracker = &dependencyfakes.FakeDependencyTracker{}

		reconciler = delivery.Reconciler{
			Repo:              repo,
			DependencyTracker: dependencyTracker,
		}

		out = NewBuffer()
		logger := zap.New(zap.WriteTo(out))
		ctx = logr.NewContext(context.Background(), logger)

		req = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      "my-new-delivery",
				Namespace: "default",
			},
		}
	})

	Context("with a well formed delivery", func() {
		var (
			apiDelivery *v1alpha1.ClusterDelivery
		)
		BeforeEach(func() {
			apiDelivery = &v1alpha1.ClusterDelivery{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "my-new-delivery",
					Generation: 99,
				},
				Spec: v1alpha1.DeliverySpec{},
			}
			repo.GetDeliveryReturns(apiDelivery, nil)
		})

		Context("all referenced templates exist", func() {
			var (
				firstTemplate  *v1alpha1.ClusterSourceTemplate
				secondTemplate *v1alpha1.ClusterTemplate
				thirdTemplate  *v1alpha1.ClusterTemplate
			)
			BeforeEach(func() {
				firstTemplate = &v1alpha1.ClusterSourceTemplate{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterSourceTemplate",
						APIVersion: "carto.run/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "my-source-template",
					},
				}

				secondTemplate = &v1alpha1.ClusterTemplate{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterTemplate",
						APIVersion: "carto.run/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "my-final-template",
					},
				}

				thirdTemplate = &v1alpha1.ClusterTemplate{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterTemplate",
						APIVersion: "carto.run/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "my-final-template-option2",
					},
				}

				apiDelivery.Spec.Resources = []v1alpha1.DeliveryResource{
					{
						Name: "first-resource",
						TemplateRef: v1alpha1.DeliveryTemplateReference{
							Kind: "ClusterSourceTemplate",
							Name: "my-source-template",
						},
					},
					{
						Name: "second-resource",
						TemplateRef: v1alpha1.DeliveryTemplateReference{
							Kind: "ClusterTemplate",
							Options: []v1alpha1.TemplateOption{
								{
									Name: "my-final-template-option1",
								},
								{
									Name: "my-final-template-option2",
								},
							},
						},
					},
				}
				repo.GetDeliveryTemplateReturnsOnCall(0, firstTemplate, nil)
				repo.GetDeliveryTemplateReturnsOnCall(1, secondTemplate, nil)
				repo.GetDeliveryTemplateReturnsOnCall(2, thirdTemplate, nil)
			})

			It("Attaches a ready/true status", func() {
				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				Expect(repo.GetDeliveryCallCount()).To(Equal(1))
				_, name := repo.GetDeliveryArgsForCall(0)

				Expect(name).To(Equal("my-new-delivery"))

				Expect(repo.StatusUpdateCallCount()).To(Equal(1))

				_, obj := repo.StatusUpdateArgsForCall(0)
				deliveryObject, ok := obj.(*v1alpha1.ClusterDelivery)

				Expect(ok).To(BeTrue())

				Expect(deliveryObject).To(Equal(apiDelivery))
				Expect(deliveryObject.Status.Conditions).To(ContainElements(
					MatchFields(
						IgnoreExtras,
						Fields{
							"Type":   Equal("Ready"),
							"Status": Equal(metav1.ConditionTrue),
							"Reason": Equal("Ready"),
						},
					),
					MatchFields(
						IgnoreExtras,
						Fields{
							"Type":   Equal("TemplatesReady"),
							"Status": Equal(metav1.ConditionTrue),
							"Reason": Equal("Ready"),
						},
					),
				))
			})

			It("updates the status.observedGeneration to equal metadata.generation", func() {
				_, _ = reconciler.Reconcile(ctx, req)

				_, updatedDelivery := repo.StatusUpdateArgsForCall(0)

				Expect(*updatedDelivery.(*v1alpha1.ClusterDelivery)).To(MatchFields(IgnoreExtras, Fields{
					"Status": MatchFields(IgnoreExtras, Fields{
						"ObservedGeneration": BeEquivalentTo(99),
					}),
				}))
			})

			It("watches the templates", func() {
				_, _ = reconciler.Reconcile(ctx, req)

				Expect(dependencyTracker.TrackCallCount()).To(Equal(3))
				firstTemplateKey, _ := dependencyTracker.TrackArgsForCall(0)
				Expect(firstTemplateKey.String()).To(Equal("ClusterSourceTemplate.carto.run//my-source-template"))

				secondTemplateKey, _ := dependencyTracker.TrackArgsForCall(1)
				Expect(secondTemplateKey.String()).To(Equal("ClusterTemplate.carto.run//my-final-template-option1"))

				thirdTemplateKey, _ := dependencyTracker.TrackArgsForCall(2)
				Expect(thirdTemplateKey.String()).To(Equal("ClusterTemplate.carto.run//my-final-template-option2"))
			})

			It("does not return an error", func() {
				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("get cluster template fails", func() {
			BeforeEach(func() {
				apiDelivery.Spec.Resources = []v1alpha1.DeliveryResource{
					{
						Name: "first-resource",
						TemplateRef: v1alpha1.DeliveryTemplateReference{
							Kind: "ClusterSourceTemplate",
							Name: "my-source-template",
						},
					},
				}

				repo.GetDeliveryTemplateReturnsOnCall(0, nil, errors.New("getting templates is hard"))
			})

			It("returns an error and requeues", func() {
				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).To(MatchError(ContainSubstring("getting templates is hard")))
			})
		})

		Context("cannot find cluster template", func() {
			BeforeEach(func() {
				apiDelivery.Spec.Resources = []v1alpha1.DeliveryResource{
					{
						Name: "first-resource",
						TemplateRef: v1alpha1.DeliveryTemplateReference{
							Kind: "ClusterSourceTemplate",
							Name: "my-source-template",
						},
					},
				}

				repo.GetDeliveryTemplateReturnsOnCall(0, nil, nil)
			})

			It("adds a positive templates NOT found condition", func() {
				_, _ = reconciler.Reconcile(ctx, req)

				_, obj := repo.StatusUpdateArgsForCall(0)

				deliveryObject, ok := obj.(*v1alpha1.ClusterDelivery)
				Expect(ok).To(BeTrue())

				Expect(deliveryObject.Status.Conditions).To(ContainElements(
					MatchFields(
						IgnoreExtras,
						Fields{
							"Type":   Equal("Ready"),
							"Status": Equal(metav1.ConditionFalse),
							"Reason": Equal("TemplatesNotFound"),
						},
					),
					MatchFields(
						IgnoreExtras,
						Fields{
							"Type":   Equal("TemplatesReady"),
							"Status": Equal(metav1.ConditionFalse),
							"Reason": Equal("TemplatesNotFound"),
						},
					),
				))
			})

			It("does not return an error", func() {
				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		It("Starts and Finishes cleanly", func() {
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())

			Eventually(out).Should(Say(`"msg":"started"`))
			Eventually(out).Should(Say(`"msg":"finished"`))
		})
	})

	Context("repo.GetDelivery fails", func() {
		It("returns an error and requeues", func() {
			repo.GetDeliveryReturns(nil, errors.New("repo.GetDelivery failed"))

			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to get delivery [default/my-new-delivery]: repo.GetDelivery failed"))
		})
	})

	Context("repo.StatusUpdate fails", func() {
		It("returns an error and requeues", func() {
			apiDelivery := &v1alpha1.ClusterDelivery{}
			repo.GetDeliveryReturns(apiDelivery, nil)
			repo.StatusUpdateReturns(errors.New("repo.StatusUpdate failed"))

			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to update status for delivery: repo.StatusUpdate failed"))
		})
	})

	Context("when the delivery has been deleted from the apiServer", func() {
		BeforeEach(func() {
			repo.GetDeliveryReturns(nil, nil)
		})

		It("does not return an error", func() {
			_, err := reconciler.Reconcile(ctx, req)

			Expect(err).NotTo(HaveOccurred())
		})
	})
})
