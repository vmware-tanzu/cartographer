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

package deliverable_test

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/conditions/conditionsfakes"
	"github.com/vmware-tanzu/cartographer/pkg/controller/deliverable"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/deliverable"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/deliverable/deliverablefakes"
	"github.com/vmware-tanzu/cartographer/pkg/registrar"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/trackerfakes"
)

var _ = Describe("Reconciler", func() {
	var (
		out               *Buffer
		reconciler        deliverable.Reconciler
		ctx               context.Context
		req               ctrl.Request
		repo              *repositoryfakes.FakeRepository
		conditionManager  *conditionsfakes.FakeConditionManager
		rlzr              *deliverablefakes.FakeRealizer
		dl                *v1alpha1.Deliverable
		deliverableLabels map[string]string
		dynamicTracker    *trackerfakes.FakeDynamicTracker

		builtResourceRealizer        *deliverablefakes.FakeResourceRealizer
		resourceRealizerSecret       *corev1.Secret
		serviceAccountSecret         *corev1.Secret
		serviceAccountName           string
		resourceRealizerBuilderError error
	)

	BeforeEach(func() {
		out = NewBuffer()
		logger := zap.New(zap.WriteTo(out))
		ctx = logr.NewContext(context.Background(), logger)

		conditionManager = &conditionsfakes.FakeConditionManager{}

		fakeConditionManagerBuilder := func(string, []metav1.Condition) conditions.ConditionManager {
			return conditionManager
		}

		rlzr = &deliverablefakes.FakeRealizer{}
		rlzr.RealizeReturns(nil, nil)

		dynamicTracker = &trackerfakes.FakeDynamicTracker{}

		repo = &repositoryfakes.FakeRepository{}
		scheme := runtime.NewScheme()
		err := registrar.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())
		repo.GetSchemeReturns(scheme)

		serviceAccountSecret = &corev1.Secret{
			StringData: map[string]string{"foo": "bar"},
		}
		repo.GetServiceAccountSecretReturns(serviceAccountSecret, nil)

		resourceRealizerBuilderError = nil
		resourceRealizerBuilder := func(secret *corev1.Secret, deliverable *v1alpha1.Deliverable, systemRepo repository.Repository, deliveryParams []v1alpha1.DelegatableParam) (realizer.ResourceRealizer, error) {
			if resourceRealizerBuilderError != nil {
				return nil, resourceRealizerBuilderError
			}
			resourceRealizerSecret = secret
			builtResourceRealizer = &deliverablefakes.FakeResourceRealizer{}
			return builtResourceRealizer, nil
		}

		reconciler = deliverable.Reconciler{
			Repo:                    repo,
			ConditionManagerBuilder: fakeConditionManagerBuilder,
			ResourceRealizerBuilder: resourceRealizerBuilder,
			Realizer:                rlzr,
			DynamicTracker:          dynamicTracker,
		}

		req = ctrl.Request{
			NamespacedName: types.NamespacedName{Name: "my-deliverable-name", Namespace: "my-namespace"},
		}

		deliverableLabels = map[string]string{"some-key": "some-val"}

		serviceAccountName = "service-account-name-for-deliverable"

		dl = &v1alpha1.Deliverable{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "my-deliverable",
				Namespace:  "my-ns",
				Generation: 1,
				Labels:     deliverableLabels,
			},
			Spec: v1alpha1.DeliverableSpec{
				ServiceAccountName: serviceAccountName,
			},
		}
		repo.GetDeliverableReturns(dl, nil)
	})

	It("logs that it's begun", func() {
		_, _ = reconciler.Reconcile(ctx, req)

		Expect(out).To(Say(`"msg":"started"`))
	})

	It("logs that it's finished", func() {
		_, _ = reconciler.Reconcile(ctx, req)

		Expect(out).To(Say(`"msg":"finished"`))
	})

	It("updates the status of the deliverable", func() {
		_, _ = reconciler.Reconcile(ctx, req)

		Expect(repo.StatusUpdateCallCount()).To(Equal(1))
	})

	It("updates the status.observedGeneration to equal metadata.generation", func() {
		_, _ = reconciler.Reconcile(ctx, req)

		_, updatedDeliverable := repo.StatusUpdateArgsForCall(0)

		Expect(*updatedDeliverable.(*v1alpha1.Deliverable)).To(MatchFields(IgnoreExtras, Fields{
			"Status": MatchFields(IgnoreExtras, Fields{
				"ObservedGeneration": BeEquivalentTo(1),
			}),
		}))
	})

	It("updates the conditions based on the condition manager", func() {
		someConditions := []metav1.Condition{
			{
				Type:               "some type",
				Status:             "True",
				LastTransitionTime: metav1.Time{},
				Reason:             "great causes",
				Message:            "good going",
			},
			{
				Type:               "another type",
				Status:             "False",
				LastTransitionTime: metav1.Time{},
				Reason:             "sad omens",
				Message:            "gotta get fixed",
			},
		}

		conditionManager.FinalizeReturns(someConditions, true)

		_, _ = reconciler.Reconcile(ctx, req)

		_, updatedDeliverable := repo.StatusUpdateArgsForCall(0)

		Expect(*updatedDeliverable.(*v1alpha1.Deliverable)).To(MatchFields(IgnoreExtras, Fields{
			"Status": MatchFields(IgnoreExtras, Fields{
				"Conditions": Equal(someConditions),
			}),
		}))
	})

	It("requests deliveries from the repo", func() {
		_, _ = reconciler.Reconcile(ctx, req)

		_, passedDeliverable := repo.GetDeliveriesForDeliverableArgsForCall(0)
		Expect(passedDeliverable).To(Equal(dl))
	})

	Context("and the repo returns a single matching delivery for the deliverable", func() {
		var (
			deliveryName   string
			delivery       v1alpha1.ClusterDelivery
			stampedObject1 *unstructured.Unstructured
			stampedObject2 *unstructured.Unstructured
		)
		BeforeEach(func() {
			deliveryName = "some-delivery"
			delivery = v1alpha1.ClusterDelivery{
				ObjectMeta: metav1.ObjectMeta{
					Name: deliveryName,
				},
				Status: v1alpha1.ClusterDeliveryStatus{
					Conditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             "True",
							LastTransitionTime: metav1.Time{},
							Reason:             "Ready",
							Message:            "Ready",
						},
					},
				},
			}
			repo.GetDeliveriesForDeliverableReturns([]*v1alpha1.ClusterDelivery{&delivery}, nil)
			stampedObject1 = &unstructured.Unstructured{}
			stampedObject1.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "thing.io",
				Version: "alphabeta1",
				Kind:    "MyThing",
			})
			stampedObject1.SetNamespace("my-ns")
			stampedObject1.SetName("my-name")
			stampedObject2 = &unstructured.Unstructured{}
			stampedObject2.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "hello.io",
				Version: "goodbye",
				Kind:    "NiceToSeeYou",
			})
			stampedObject2.SetNamespace("my-ns-two")
			stampedObject2.SetName("my-name-two")
			rlzr.RealizeReturns([]*unstructured.Unstructured{stampedObject1, stampedObject2}, nil)
		})

		It("uses the service account specified by the workload for realizing resources", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(repo.GetServiceAccountSecretCallCount()).To(Equal(1))
			_, serviceAccountNameArg, serviceAccountNS := repo.GetServiceAccountSecretArgsForCall(0)
			Expect(serviceAccountNameArg).To(Equal(serviceAccountName))
			Expect(serviceAccountNS).To(Equal("my-ns"))
			Expect(resourceRealizerSecret).To(Equal(serviceAccountSecret))

			Expect(rlzr.RealizeCallCount()).To(Equal(1))
			_, resourceRealizer, _ := rlzr.RealizeArgsForCall(0)
			Expect(resourceRealizer).To(Equal(builtResourceRealizer))
		})

		It("uses the default service account in the deliverables namespace if there is no service account specified", func() {
			dl.Spec.ServiceAccountName = ""
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(repo.GetServiceAccountSecretCallCount()).To(Equal(1))
			_, serviceAccountNameArg, serviceAccountNS := repo.GetServiceAccountSecretArgsForCall(0)
			Expect(serviceAccountNameArg).To(Equal("default"))
			Expect(serviceAccountNS).To(Equal("my-ns"))
			Expect(resourceRealizerSecret).To(Equal(serviceAccountSecret))

			Expect(rlzr.RealizeCallCount()).To(Equal(1))
			_, resourceRealizer, _ := rlzr.RealizeArgsForCall(0)
			Expect(resourceRealizer).To(Equal(builtResourceRealizer))
		})

		It("sets the DeliveryRef", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(dl.Status.DeliveryRef.Kind).To(Equal("ClusterDelivery"))
			Expect(dl.Status.DeliveryRef.Name).To(Equal(deliveryName))
		})

		It("calls the condition manager to specify the delivery is ready", func() {
			_, _ = reconciler.Reconcile(ctx, req)
			Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(deliverable.DeliveryReadyCondition()))
		})

		It("calls the condition manager to report the resources have been submitted", func() {
			_, _ = reconciler.Reconcile(ctx, req)
			Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(deliverable.ResourcesSubmittedCondition()))
		})

		It("watches the stampedObjects kinds", func() {
			_, _ = reconciler.Reconcile(ctx, req)
			Expect(dynamicTracker.WatchCallCount()).To(Equal(2))
			_, obj, hndl := dynamicTracker.WatchArgsForCall(0)

			Expect(obj).To(Equal(stampedObject1))
			Expect(hndl).To(Equal(&handler.EnqueueRequestForOwner{OwnerType: &v1alpha1.Deliverable{}}))

			_, obj, hndl = dynamicTracker.WatchArgsForCall(1)

			Expect(obj).To(Equal(stampedObject2))
			Expect(hndl).To(Equal(&handler.EnqueueRequestForOwner{OwnerType: &v1alpha1.Deliverable{}}))
		})

		Context("but getting the object GVK fails", func() {
			BeforeEach(func() {
				repo.GetSchemeReturns(runtime.NewScheme())
			})

			It("returns an unhandled error and requeues", func() {
				_, err := reconciler.Reconcile(ctx, req)

				Expect(err.Error()).To(ContainSubstring("failed to get object gvk for delivery [some-delivery]"))
			})
		})

		Context("but the delivery is not in a ready state", func() {
			BeforeEach(func() {
				delivery.Status.Conditions = []metav1.Condition{
					{
						Type:    "Ready",
						Status:  "False",
						Reason:  "SomeReason",
						Message: "some informative message",
					},
				}
				repo.GetDeliveriesForDeliverableReturns([]*v1alpha1.ClusterDelivery{&delivery}, nil)
			})

			It("does not return an error", func() {
				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
			})

			It("calls the condition manager to report delivery not ready", func() {
				_, _ = reconciler.Reconcile(ctx, req)
				expectedCondition := metav1.Condition{
					Type:               v1alpha1.DeliverableDeliveryReady,
					Status:             metav1.ConditionFalse,
					ObservedGeneration: 0,
					LastTransitionTime: metav1.Time{},
					Reason:             "SomeReason",
					Message:            "some informative message",
				}
				Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(deliverable.MissingReadyInDeliveryCondition(expectedCondition)))
			})

			It("logs the handled error message", func() {
				_, _ = reconciler.Reconcile(ctx, req)

				Expect(out).To(Say(`"level":"info"`))
				Expect(out).To(Say(`"deliverable":"my-namespace/my-deliverable-name"`))
				Expect(out).To(Say(`"delivery":"some-delivery"`))
				Expect(out).To(Say(`"handled error":"delivery \[some-delivery\] is not in ready state"`))
			})
		})

		Context("but the realizer returns an error", func() {
			Context("of type GetClusterTemplateError", func() {
				var templateError error
				BeforeEach(func() {
					templateError = realizer.GetDeliveryClusterTemplateError{
						Err: errors.New("some error"),
					}
					rlzr.RealizeReturns(nil, templateError)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(deliverable.TemplateObjectRetrievalFailureCondition(templateError)))
				})

				It("returns an unhandled error and requeues", func() {
					_, err := reconciler.Reconcile(ctx, req)

					Expect(err.Error()).To(ContainSubstring("unable to get template"))
				})
			})

			Context("of type StampError", func() {
				var stampError realizer.StampError
				BeforeEach(func() {
					stampError = realizer.StampError{
						Err:      errors.New("some error"),
						Resource: &v1alpha1.ClusterDeliveryResource{Name: "some-name"},
					}
					rlzr.RealizeReturns(nil, stampError)
				})

				It("does not try to watch the stampedObjects", func() {
					_, err := reconciler.Reconcile(ctx, req)
					Expect(err).NotTo(HaveOccurred())

					Expect(dynamicTracker.WatchCallCount()).To(Equal(0))
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(deliverable.TemplateStampFailureCondition(stampError)))
				})

				It("does not return an error", func() {
					_, err := reconciler.Reconcile(ctx, req)
					Expect(err).NotTo(HaveOccurred())
				})

				It("logs the handled error message", func() {
					_, _ = reconciler.Reconcile(ctx, req)

					Expect(out).To(Say(`"level":"info"`))
					Expect(out).To(Say(`"msg":"handled error reconciling deliverable"`))
					Expect(out).To(Say(`"deliverable":"my-namespace/my-deliverable-name"`))
					Expect(out).To(Say(`"delivery":"some-delivery"`))
					Expect(out).To(Say(`"handled error":"unable to stamp object for resource \[some-name\]: some error"`))
				})
			})

			Context("of type ApplyStampedObjectError", func() {
				var stampedObjectError realizer.ApplyStampedObjectError
				BeforeEach(func() {
					stampedObjectError = realizer.ApplyStampedObjectError{
						Err:           errors.New("some error"),
						StampedObject: &unstructured.Unstructured{},
					}
					rlzr.RealizeReturns(nil, stampedObjectError)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(deliverable.TemplateRejectedByAPIServerCondition(stampedObjectError)))
				})

				It("returns an unhandled error and requeues", func() {
					_, err := reconciler.Reconcile(ctx, req)

					Expect(err.Error()).To(ContainSubstring("unable to apply object"))
				})
			})

			Context("of type ApplyStampedObjectError where the user did not have proper permissions", func() {
				var stampedObjectError realizer.ApplyStampedObjectError
				BeforeEach(func() {
					status := &metav1.Status{
						Message: "fantastic error",
						Reason:  metav1.StatusReasonForbidden,
						Code:    403,
					}
					stampedObject1 = &unstructured.Unstructured{}
					stampedObject1.SetNamespace("a-namespace")
					stampedObject1.SetName("a-name")

					stampedObjectError = realizer.ApplyStampedObjectError{
						Err:           kerrors.FromObject(status),
						StampedObject: stampedObject1,
					}

					rlzr.RealizeReturns(nil, stampedObjectError)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(deliverable.TemplateRejectedByAPIServerCondition(stampedObjectError)))
				})

				It("handles the error and logs it", func() {
					_, err := reconciler.Reconcile(ctx, req)
					Expect(err).NotTo(HaveOccurred())

					Expect(out).To(Say(`"level":"info"`))
					Expect(out).To(Say(`"handled error":"unable to apply object \[a-namespace/a-name\]: fantastic error"`))
				})
			})

			Context("of type RetrieveOutputError", func() {
				var retrieveError realizer.RetrieveOutputError
				var wrappedError error

				JustBeforeEach(func() {
					retrieveError = realizer.RetrieveOutputError{
						Err:      wrappedError,
						Resource: &v1alpha1.ClusterDeliveryResource{Name: "some-resource"},
					}

					rlzr.RealizeReturns(nil, retrieveError)
				})

				Context("which wraps an ObservedGenerationError", func() {
					BeforeEach(func() {
						wrappedError = templates.NewObservedGenerationError(errors.New("some error"))
					})

					It("calls the condition manager to report", func() {
						_, _ = reconciler.Reconcile(ctx, req)
						Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(deliverable.TemplateStampFailureByObservedGenerationCondition(retrieveError)))
					})

					It("does not return an error", func() {
						_, err := reconciler.Reconcile(ctx, req)
						Expect(err).NotTo(HaveOccurred())
					})

					It("logs the handled error message", func() {
						_, _ = reconciler.Reconcile(ctx, req)

						Expect(out).To(Say(`"level":"info"`))
						Expect(out).To(Say(`"msg":"handled error reconciling deliverable"`))
						Expect(out).To(Say(`"deliverable":"my-namespace/my-deliverable-name"`))
						Expect(out).To(Say(`"delivery":"some-delivery"`))
						Expect(out).To(Say(`"handled error":"unable to retrieve outputs from stamped object for resource \[some-resource\]: some error"`))
					})
				})

				Context("which wraps an DeploymentConditionError", func() {
					BeforeEach(func() {
						wrappedError = templates.NewDeploymentConditionError(errors.New("some error"))
					})

					It("calls the condition manager to report", func() {
						_, _ = reconciler.Reconcile(ctx, req)
						Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(deliverable.DeploymentConditionNotMetCondition(retrieveError)))
					})

					It("does not return an error", func() {
						_, err := reconciler.Reconcile(ctx, req)
						Expect(err).NotTo(HaveOccurred())
					})

					It("logs the handled error message", func() {
						_, _ = reconciler.Reconcile(ctx, req)

						Expect(out).To(Say(`"level":"info"`))
						Expect(out).To(Say(`"msg":"handled error reconciling deliverable"`))
						Expect(out).To(Say(`"deliverable":"my-namespace/my-deliverable-name"`))
						Expect(out).To(Say(`"delivery":"some-delivery"`))
						Expect(out).To(Say(`"handled error":"unable to retrieve outputs from stamped object for resource \[some-resource\]: some error"`))
					})
				})

				Context("which wraps an DeploymentFailedConditionMetError", func() {
					BeforeEach(func() {
						wrappedError = templates.NewDeploymentFailedConditionMetError(errors.New("some error"))
					})

					It("calls the condition manager to report", func() {
						_, _ = reconciler.Reconcile(ctx, req)
						Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(deliverable.DeploymentFailedConditionMetCondition(retrieveError)))
					})

					It("does not return an error", func() {
						_, err := reconciler.Reconcile(ctx, req)
						Expect(err).NotTo(HaveOccurred())
					})

					It("logs the handled error message", func() {
						_, _ = reconciler.Reconcile(ctx, req)

						Expect(out).To(Say(`"level":"info"`))
						Expect(out).To(Say(`"msg":"handled error reconciling deliverable"`))
						Expect(out).To(Say(`"deliverable":"my-namespace/my-deliverable-name"`))
						Expect(out).To(Say(`"delivery":"some-delivery"`))
						Expect(out).To(Say(`"handled error":"unable to retrieve outputs from stamped object for resource \[some-resource\]: some error"`))
					})
				})

				Context("which wraps a json path error", func() {
					BeforeEach(func() {
						wrappedError = templates.NewJsonPathError("this.wont.find.anything", errors.New("some error"))
					})

					It("calls the condition manager to report", func() {
						_, _ = reconciler.Reconcile(ctx, req)
						Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(deliverable.MissingValueAtPathCondition("some-resource", "this.wont.find.anything")))
					})

					It("does not return an error", func() {
						_, err := reconciler.Reconcile(ctx, req)
						Expect(err).NotTo(HaveOccurred())
					})

					It("logs the handled error message", func() {
						_, _ = reconciler.Reconcile(ctx, req)

						Expect(out).To(Say(`"level":"info"`))
						Expect(out).To(Say(`"msg":"handled error reconciling deliverable"`))
						Expect(out).To(Say(`"deliverable":"my-namespace/my-deliverable-name"`))
						Expect(out).To(Say(`"delivery":"some-delivery"`))
						Expect(out).To(Say(`"handled error":"unable to retrieve outputs from stamped object for resource \[some-resource\]: evaluate json path 'this.wont.find.anything': some error"`))
					})
				})

				Context("which wraps any other error", func() {
					BeforeEach(func() {
						wrappedError = errors.New("some error")
					})

					It("calls the condition manager to report", func() {
						_, _ = reconciler.Reconcile(ctx, req)
						Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(deliverable.UnknownResourceErrorCondition(retrieveError)))
					})

					It("does not return an error", func() {
						_, err := reconciler.Reconcile(ctx, req)
						Expect(err).NotTo(HaveOccurred())
					})

					It("logs the handled error message", func() {
						_, _ = reconciler.Reconcile(ctx, req)

						Expect(out).To(Say(`"level":"info"`))
						Expect(out).To(Say(`"msg":"handled error reconciling deliverable"`))
						Expect(out).To(Say(`"deliverable":"my-namespace/my-deliverable-name"`))
						Expect(out).To(Say(`"delivery":"some-delivery"`))
						Expect(out).To(Say(`"handled error":"unable to retrieve outputs from stamped object for resource \[some-resource\]: some error"`))
					})
				})
			})

			Context("of unknown type", func() {
				var realizerError error
				BeforeEach(func() {
					realizerError = errors.New("some error")
					rlzr.RealizeReturns(nil, realizerError)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(deliverable.UnknownResourceErrorCondition(realizerError)))
				})

				It("returns an unhandled error and requeues", func() {
					_, err := reconciler.Reconcile(ctx, req)

					Expect(err.Error()).To(ContainSubstring("some error"))
				})
			})
		})

		Context("but the watcher returns an error", func() {
			BeforeEach(func() {
				dynamicTracker.WatchReturns(errors.New("could not watch"))
			})

			It("logs the error message", func() {
				_, _ = reconciler.Reconcile(ctx, req)

				Expect(out).To(Say(`"level":"error"`))
				Expect(out).To(Say(`"msg":"failed to add informer for object"`))
				Expect(out).To(Say(`"deliverable":"my-namespace/my-deliverable-name"`))
				Expect(out).To(Say(`"delivery":"some-delivery"`))
				Expect(out).To(Say(`"object":{"apiVersion":"hello.io/goodbye","kind":"NiceToSeeYou","namespace":"my-ns-two","name":"my-name-two"}`))
			})

			It("returns an unhandled error and requeues", func() {
				_, err := reconciler.Reconcile(ctx, req)

				Expect(err.Error()).To(ContainSubstring("could not watch"))
			})
		})

		Context("but the repo returns an error when requesting the service account secret", func() {
			var repoError error
			BeforeEach(func() {
				repoError = errors.New("some error")
				repo.GetServiceAccountSecretReturns(nil, repoError)
			})

			It("calls the condition manager to add a service account secret not found condition", func() {
				_, _ = reconciler.Reconcile(ctx, req)
				Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(deliverable.ServiceAccountSecretNotFoundCondition(repoError)))
			})

			It("handles the error and logs it", func() {
				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				Expect(out).To(Say(`"level":"info"`))
				Expect(out).To(Say(`"handled error":"get secret for service account 'service-account-name-for-deliverable': some error"`))
			})
		})

		Context("but the resource realizer builder fails", func() {
			BeforeEach(func() {
				resourceRealizerBuilderError = errors.New("some error")
			})

			It("calls the condition manager to add a resource realizer builder error condition", func() {
				_, _ = reconciler.Reconcile(ctx, req)
				Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(deliverable.ResourceRealizerBuilderErrorCondition(resourceRealizerBuilderError)))
			})

			It("returns an unhandled error", func() {
				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("build resource realizer: some error"))
			})
		})
	})

	Context("but the deliverable has no label to match with the delivery", func() {
		BeforeEach(func() {
			dl.Labels = nil
		})

		It("does not return an error", func() {
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
		})

		It("logs the handled error message", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(out).To(Say(`"level":"info"`))
			Expect(out).To(Say(`"msg":"handled error reconciling deliverable"`))
			Expect(out).To(Say(`"deliverable":"my-namespace/my-deliverable-name"`))
			Expect(out).To(Say(`"handled error":"deliverable \[my-ns/my-deliverable\] is missing required labels"`))
		})

		It("calls the condition manager to report the bad state", func() {
			_, _ = reconciler.Reconcile(ctx, req)
			Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(deliverable.DeliverableMissingLabelsCondition()))
		})
	})

	Context("and repo returns an empty list of deliveries", func() {
		It("does not return an error", func() {
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
		})

		It("calls the condition manager to add a delivery not found condition", func() {
			_, _ = reconciler.Reconcile(ctx, req)
			Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(deliverable.DeliveryNotFoundCondition(deliverableLabels)))
		})

		It("logs the handled error message", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(out).To(Say(`"level":"info"`))
			Expect(out).To(Say(`"msg":"handled error reconciling deliverable"`))
			Expect(out).To(Say(`"deliverable":"my-namespace/my-deliverable-name"`))
			Expect(out).To(Say(`"handled error":"no delivery \[my-ns/my-deliverable\] found where full selector is satisfied by labels: map\[some-key:some-val\]"`))
		})
	})

	Context("and repo returns an an error when requesting deliveries", func() {
		BeforeEach(func() {
			repo.GetDeliveriesForDeliverableReturns(nil, errors.New("some error"))
		})

		It("returns an unhandled error and requeues", func() {
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err.Error()).To(ContainSubstring("failed to get deliveries for deliverable [my-ns/my-deliverable]: some error"))
		})
	})

	Context("and the repo returns multiple deliveries", func() {
		BeforeEach(func() {
			delivery := v1alpha1.ClusterDelivery{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-name",
					Namespace: "my-ns",
				},
			}
			delivery.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "thing.io",
				Version: "alphabeta1",
				Kind:    "MyThing",
			})
			repo.GetDeliveriesForDeliverableReturns([]*v1alpha1.ClusterDelivery{&delivery, &delivery}, nil)
		})

		It("does not return an error", func() {
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
		})

		It("calls the condition manager to report too mane deliveries matched", func() {
			_, _ = reconciler.Reconcile(ctx, req)
			Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(deliverable.TooManyDeliveryMatchesCondition()))
		})

		It("logs the handled error message", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(out).To(Say(`"level":"info"`))
			Expect(out).To(Say(`"msg":"handled error reconciling deliverable"`))
			Expect(out).To(Say(`"deliverable":"my-namespace/my-deliverable-name"`))
			Expect(out).To(Say(`"handled error":"more than one delivery selected for deliverable \[my-ns/my-deliverable\]: \[my-name my-name\]"`))
		})
	})

	Context("but status update fails", func() {
		BeforeEach(func() {
			repo.StatusUpdateReturns(errors.New("some error"))
		})

		It("returns an unhandled error and requeues", func() {
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).To(MatchError(ContainSubstring("failed to update status for deliverable: some error")))
		})
	})

	Context("getting the deliverable returns an error", func() {
		BeforeEach(func() {
			repositoryError := errors.New("RepositoryError")
			repo.GetDeliverableReturns(nil, repositoryError)
		})

		It("returns the error from the repository", func() {
			_, err := reconciler.Reconcile(ctx, req)

			Expect(err).To(MatchError(ContainSubstring("RepositoryError")))
		})
	})

	Context("deliverable is deleted", func() {
		BeforeEach(func() {
			repo.GetDeliverableReturns(nil, nil)
		})

		It("finishes the reconcile and does not requeue", func() {
			_, err := reconciler.Reconcile(ctx, req)

			Expect(err).NotTo(HaveOccurred())
		})
	})
})
