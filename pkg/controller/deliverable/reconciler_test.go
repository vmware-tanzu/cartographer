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
	"fmt"

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
	"github.com/vmware-tanzu/cartographer/pkg/tracker/dependency/dependencyfakes"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/stamped/stampedfakes"
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
		stampedTracker    *stampedfakes.FakeStampedTracker
		dependencyTracker *dependencyfakes.FakeDependencyTracker

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

		stampedTracker = &stampedfakes.FakeStampedTracker{}
		dependencyTracker = &dependencyfakes.FakeDependencyTracker{}

		repo = &repositoryfakes.FakeRepository{}
		scheme := runtime.NewScheme()
		err := registrar.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())
		repo.GetSchemeReturns(scheme)

		serviceAccountName = "service-account-name-for-deliverable"

		serviceAccountSecret = &corev1.Secret{
			StringData: map[string]string{"foo": "bar"},
		}
		repo.GetServiceAccountSecretReturns(serviceAccountSecret, nil)

		resourceRealizerBuilderError = nil
		resourceRealizerBuilder := func(secret *corev1.Secret, deliverable *v1alpha1.Deliverable, systemRepo repository.Repository, deliveryParams []v1alpha1.BlueprintParam) (realizer.ResourceRealizer, error) {
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
			StampedTracker:          stampedTracker,
			DependencyTracker:       dependencyTracker,
		}

		req = ctrl.Request{
			NamespacedName: types.NamespacedName{Name: "my-deliverable-name", Namespace: "my-namespace"},
		}

		deliverableLabels = map[string]string{"some-key": "some-val"}

		dl = &v1alpha1.Deliverable{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "my-deliverable",
				Namespace:  "my-namespace",
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
				"OwnerStatus": MatchFields(IgnoreExtras, Fields{
					"ObservedGeneration": BeEquivalentTo(1),
				}),
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
				"OwnerStatus": MatchFields(IgnoreExtras, Fields{
					"Conditions": Equal(someConditions),
				}),
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
			deliveryName      string
			delivery          v1alpha1.ClusterDelivery
			realizedResources []v1alpha1.RealizedResource
		)
		BeforeEach(func() {
			deliveryName = "some-delivery"
			delivery = v1alpha1.ClusterDelivery{
				ObjectMeta: metav1.ObjectMeta{
					Name: deliveryName,
				},
				Status: v1alpha1.DeliveryStatus{
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

			realizedResources = []v1alpha1.RealizedResource{
				{
					StampedRef: &corev1.ObjectReference{
						Kind:       "MyThing",
						APIVersion: "thing.io/alphabeta1",
					},
					TemplateRef: &corev1.ObjectReference{
						Kind:       "my-image-kind",
						Name:       "my-image-template",
						APIVersion: "carto.run/v1alpha1",
					},
				},
				{
					StampedRef: &corev1.ObjectReference{
						Kind:       "NiceToSeeYou",
						APIVersion: "hello.io/goodbye",
					},
					TemplateRef: &corev1.ObjectReference{
						Kind:       "my-config-kind",
						Name:       "my-config-template",
						APIVersion: "carto.run/v1alpha1",
					},
				},
			}

			rlzr.RealizeReturns(realizedResources, nil)
		})

		It("updates the status of the workload with the realizedResources", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(repo.StatusUpdateCallCount()).To(Equal(1))
			_, dl := repo.StatusUpdateArgsForCall(0)
			Expect(dl.(*v1alpha1.Deliverable).Status.Resources).To(Equal(realizedResources))
		})

		It("dynamically creates a resource realizer", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(rlzr.RealizeCallCount()).To(Equal(1))
			_, resourceRealizer, _, _ := rlzr.RealizeArgsForCall(0)
			Expect(resourceRealizer).To(Equal(builtResourceRealizer))
		})

		It("uses the service account specified by the deliverable for realizing resources", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(repo.GetServiceAccountSecretCallCount()).To(Equal(1))
			_, serviceAccountNameArg, serviceAccountNS := repo.GetServiceAccountSecretArgsForCall(0)
			Expect(serviceAccountNameArg).To(Equal(serviceAccountName))
			Expect(serviceAccountNS).To(Equal("my-namespace"))

			Expect(resourceRealizerSecret).To(Equal(serviceAccountSecret))
		})

		Context("the deliverable does not specify a service account", func() {
			BeforeEach(func() {
				dl.Spec.ServiceAccountName = ""
			})

			Context("the delivery provides a service account", func() {
				var deliveryServiceAccountSecret *corev1.Secret

				BeforeEach(func() {
					delivery.Spec.ServiceAccountRef.Name = "some-delivery-service-account"

					deliveryServiceAccountSecret = &corev1.Secret{Data: map[string][]byte{"token": []byte(`some-delivery-service-account-token`)}}
					repo.GetServiceAccountSecretReturns(deliveryServiceAccountSecret, nil)
				})

				It("uses the delivery service account in the deliverable's namespace", func() {
					_, _ = reconciler.Reconcile(ctx, req)

					Expect(repo.GetServiceAccountSecretCallCount()).To(Equal(1))
					_, serviceAccountNameArg, serviceAccountNS := repo.GetServiceAccountSecretArgsForCall(0)
					Expect(serviceAccountNameArg).To(Equal("some-delivery-service-account"))
					Expect(serviceAccountNS).To(Equal("my-namespace"))

					Expect(resourceRealizerSecret).To(Equal(deliveryServiceAccountSecret))
				})

				Context("the delivery specifies a namespace", func() {
					BeforeEach(func() {
						delivery.Spec.ServiceAccountRef.Namespace = "some-delivery-namespace"
					})

					It("uses the delivery service account in the specified namespace", func() {
						_, _ = reconciler.Reconcile(ctx, req)

						Expect(repo.GetServiceAccountSecretCallCount()).To(Equal(1))
						_, serviceAccountNameArg, serviceAccountNS := repo.GetServiceAccountSecretArgsForCall(0)
						Expect(serviceAccountNameArg).To(Equal("some-delivery-service-account"))
						Expect(serviceAccountNS).To(Equal("some-delivery-namespace"))

						Expect(resourceRealizerSecret).To(Equal(deliveryServiceAccountSecret))
					})
				})
			})

			Context("the delivery does not provide a service account", func() {
				var defaultServiceAccountSecret *corev1.Secret

				BeforeEach(func() {
					defaultServiceAccountSecret = &corev1.Secret{Data: map[string][]byte{"token": []byte(`some-default-service-account-token`)}}
					repo.GetServiceAccountSecretReturns(defaultServiceAccountSecret, nil)
				})

				It("defaults to the default service account in the deliverables namespace", func() {
					_, _ = reconciler.Reconcile(ctx, req)

					Expect(repo.GetServiceAccountSecretCallCount()).To(Equal(1))
					_, serviceAccountNameArg, serviceAccountNS := repo.GetServiceAccountSecretArgsForCall(0)
					Expect(serviceAccountNameArg).To(Equal("default"))
					Expect(serviceAccountNS).To(Equal("my-namespace"))

					Expect(resourceRealizerSecret).To(Equal(defaultServiceAccountSecret))
				})
			})
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
			Expect(stampedTracker.WatchCallCount()).To(Equal(2))
			_, obj, hndl, _ := stampedTracker.WatchArgsForCall(0)

			Expect(obj.GetObjectKind().GroupVersionKind().Kind).To(Equal(realizedResources[0].StampedRef.GetObjectKind().GroupVersionKind().Kind))
			Expect(obj.GetObjectKind().GroupVersionKind().Version).To(Equal(realizedResources[0].StampedRef.GetObjectKind().GroupVersionKind().Version))
			Expect(hndl).To(Equal(&handler.EnqueueRequestForOwner{OwnerType: &v1alpha1.Deliverable{}}))

			_, obj, hndl, _ = stampedTracker.WatchArgsForCall(1)

			Expect(obj.GetObjectKind().GroupVersionKind().Kind).To(Equal(realizedResources[1].StampedRef.GetObjectKind().GroupVersionKind().Kind))
			Expect(obj.GetObjectKind().GroupVersionKind().Version).To(Equal(realizedResources[1].StampedRef.GetObjectKind().GroupVersionKind().Version))
			Expect(hndl).To(Equal(&handler.EnqueueRequestForOwner{OwnerType: &v1alpha1.Deliverable{}}))
		})

		It("clears the previously tracked objects for the deliverable", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(dependencyTracker.ClearTrackedCallCount()).To(Equal(1))
			obj := dependencyTracker.ClearTrackedArgsForCall(0)
			Expect(obj.Name).To(Equal("my-deliverable"))
			Expect(obj.Namespace).To(Equal("my-namespace"))
		})

		It("watches the templates and service account", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(dependencyTracker.TrackCallCount()).To(Equal(3))
			serviceAccountKey, _ := dependencyTracker.TrackArgsForCall(0)
			Expect(serviceAccountKey.String()).To(Equal("ServiceAccount/my-namespace/service-account-name-for-deliverable"))

			firstTemplateKey, _ := dependencyTracker.TrackArgsForCall(1)
			Expect(firstTemplateKey.String()).To(Equal("my-image-kind.carto.run//my-image-template"))

			secondTemplateKey, _ := dependencyTracker.TrackArgsForCall(2)
			Expect(secondTemplateKey.String()).To(Equal("my-config-kind.carto.run//my-config-template"))
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
			Context("of type GetTemplateError", func() {
				var templateError error
				BeforeEach(func() {
					templateError = realizer.GetTemplateError{
						Resource: &v1alpha1.DeliveryResource{Name: "some-name"},
						Err:      errors.New("some error"),
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

				It("does not track the template", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(dependencyTracker.TrackCallCount()).To(Equal(1))
					key, obj := dependencyTracker.TrackArgsForCall(0)
					Expect(key.String()).To(Equal("ServiceAccount/my-namespace/service-account-name-for-deliverable"))
					Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
				})
			})

			Context("of type StampError", func() {
				var stampError realizer.StampError
				BeforeEach(func() {
					stampError = realizer.StampError{
						Err:          errors.New("some error"),
						Resource:     &v1alpha1.DeliveryResource{Name: "some-name"},
						DeliveryName: "some-delivery",
					}
					realizedResources[0].StampedRef = nil
					realizedResources[1].StampedRef = nil
					rlzr.RealizeReturns(realizedResources, stampError)
				})

				It("does not try to watch the stampedObjects", func() {
					_, err := reconciler.Reconcile(ctx, req)
					Expect(err).NotTo(HaveOccurred())

					Expect(stampedTracker.WatchCallCount()).To(Equal(0))
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
					Expect(out).To(Say(`"handled error":"unable to stamp object for resource \[some-name\] in delivery \[some-delivery\]: some error"`))
				})

				It("does track the template", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(dependencyTracker.TrackCallCount()).To(Equal(3))
					key, obj := dependencyTracker.TrackArgsForCall(0)
					Expect(key.String()).To(Equal("ServiceAccount/my-namespace/service-account-name-for-deliverable"))
					Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
					key, obj = dependencyTracker.TrackArgsForCall(1)
					Expect(key.String()).To(Equal("my-image-kind.carto.run//my-image-template"))
					Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
					key, obj = dependencyTracker.TrackArgsForCall(2)
					Expect(key.String()).To(Equal("my-config-kind.carto.run//my-config-template"))
					Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
				})
			})

			Context("of type ApplyStampedObjectError", func() {
				var stampedObjectError realizer.ApplyStampedObjectError
				BeforeEach(func() {
					stampedObjectError = realizer.ApplyStampedObjectError{
						Err:           errors.New("some error"),
						StampedObject: &unstructured.Unstructured{},
						Resource:      &v1alpha1.DeliveryResource{Name: "some-name"},
					}
					rlzr.RealizeReturns(realizedResources, stampedObjectError)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(deliverable.TemplateRejectedByAPIServerCondition(stampedObjectError)))
				})

				It("returns an unhandled error and requeues", func() {
					_, err := reconciler.Reconcile(ctx, req)

					Expect(err.Error()).To(ContainSubstring("unable to apply object"))
				})

				It("does track the template", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(dependencyTracker.TrackCallCount()).To(Equal(3))
					key, obj := dependencyTracker.TrackArgsForCall(0)
					Expect(key.String()).To(Equal("ServiceAccount/my-namespace/service-account-name-for-deliverable"))
					Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
					key, obj = dependencyTracker.TrackArgsForCall(1)
					Expect(key.String()).To(Equal("my-image-kind.carto.run//my-image-template"))
					Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
					key, obj = dependencyTracker.TrackArgsForCall(2)
					Expect(key.String()).To(Equal("my-config-kind.carto.run//my-config-template"))
					Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
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
					stampedObject1 := &unstructured.Unstructured{}
					stampedObject1.SetNamespace("a-namespace")
					stampedObject1.SetName("a-name")

					stampedObjectError = realizer.ApplyStampedObjectError{
						Err:           kerrors.FromObject(status),
						StampedObject: stampedObject1,
						Resource:      &v1alpha1.DeliveryResource{Name: "some-name"},
						DeliveryName:  deliveryName,
					}

					rlzr.RealizeReturns(realizedResources, stampedObjectError)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(deliverable.TemplateRejectedByAPIServerCondition(stampedObjectError)))
				})

				It("handles the error and logs it", func() {
					_, err := reconciler.Reconcile(ctx, req)
					Expect(err).NotTo(HaveOccurred())

					Expect(out).To(Say(`"level":"info"`))
					Expect(out).To(Say(`"handled error":"unable to apply object \[a-namespace/a-name\] for resource \[some-name\] in delivery \[some-delivery\]: fantastic error"`))
				})

				It("does track the template", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(dependencyTracker.TrackCallCount()).To(Equal(3))
					key, obj := dependencyTracker.TrackArgsForCall(0)
					Expect(key.String()).To(Equal("ServiceAccount/my-namespace/service-account-name-for-deliverable"))
					Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
					key, obj = dependencyTracker.TrackArgsForCall(1)
					Expect(key.String()).To(Equal("my-image-kind.carto.run//my-image-template"))
					Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
					key, obj = dependencyTracker.TrackArgsForCall(2)
					Expect(key.String()).To(Equal("my-config-kind.carto.run//my-config-template"))
					Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
				})
			})

			Context("of type RetrieveOutputError", func() {
				var retrieveError realizer.RetrieveOutputError
				var wrappedError error
				var stampedObject *unstructured.Unstructured

				JustBeforeEach(func() {
					stampedObject = &unstructured.Unstructured{}
					stampedObject.SetGroupVersionKind(schema.GroupVersionKind{
						Group:   "thing.io",
						Version: "alphabeta1",
						Kind:    "MyThing",
					})
					stampedObject.SetName("my-obj")
					stampedObject.SetNamespace("my-ns")

					retrieveError = realizer.RetrieveOutputError{
						Err:           wrappedError,
						Resource:      &v1alpha1.DeliveryResource{Name: "some-resource"},
						DeliveryName:  deliveryName,
						StampedObject: stampedObject,
					}

					rlzr.RealizeReturns(realizedResources, retrieveError)
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
						Expect(out).To(Say(`"handled error":"unable to retrieve outputs from stamped object \[my-ns/my-obj\] of type \[mything.thing.io\] for resource \[some-resource\] in delivery \[some-delivery\]: some error"`))
					})

					It("does track the template", func() {
						_, _ = reconciler.Reconcile(ctx, req)
						Expect(dependencyTracker.TrackCallCount()).To(Equal(3))
						key, obj := dependencyTracker.TrackArgsForCall(0)
						Expect(key.String()).To(Equal("ServiceAccount/my-namespace/service-account-name-for-deliverable"))
						Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
						key, obj = dependencyTracker.TrackArgsForCall(1)
						Expect(key.String()).To(Equal("my-image-kind.carto.run//my-image-template"))
						Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
						key, obj = dependencyTracker.TrackArgsForCall(2)
						Expect(key.String()).To(Equal("my-config-kind.carto.run//my-config-template"))
						Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
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
						Expect(out).To(Say(`"handled error":"unable to retrieve outputs from stamped object \[my-ns/my-obj\] of type \[mything.thing.io\] for resource \[some-resource\] in delivery \[some-delivery\]: some error"`))
					})

					It("does track the template", func() {
						_, _ = reconciler.Reconcile(ctx, req)
						Expect(dependencyTracker.TrackCallCount()).To(Equal(3))
						key, obj := dependencyTracker.TrackArgsForCall(0)
						Expect(key.String()).To(Equal("ServiceAccount/my-namespace/service-account-name-for-deliverable"))
						Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
						key, obj = dependencyTracker.TrackArgsForCall(1)
						Expect(key.String()).To(Equal("my-image-kind.carto.run//my-image-template"))
						Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
						key, obj = dependencyTracker.TrackArgsForCall(2)
						Expect(key.String()).To(Equal("my-config-kind.carto.run//my-config-template"))
						Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
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
						Expect(out).To(Say(`"handled error":"unable to retrieve outputs from stamped object \[my-ns/my-obj\] of type \[mything.thing.io\] for resource \[some-resource\] in delivery \[some-delivery\]: some error"`))
					})

					It("does track the template", func() {
						_, _ = reconciler.Reconcile(ctx, req)
						Expect(dependencyTracker.TrackCallCount()).To(Equal(3))
						key, obj := dependencyTracker.TrackArgsForCall(0)
						Expect(key.String()).To(Equal("ServiceAccount/my-namespace/service-account-name-for-deliverable"))
						Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
						key, obj = dependencyTracker.TrackArgsForCall(1)
						Expect(key.String()).To(Equal("my-image-kind.carto.run//my-image-template"))
						Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
						key, obj = dependencyTracker.TrackArgsForCall(2)
						Expect(key.String()).To(Equal("my-config-kind.carto.run//my-config-template"))
						Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
					})
				})

				Context("which wraps a json path error", func() {
					BeforeEach(func() {
						wrappedError = templates.NewJsonPathError("this.wont.find.anything", errors.New("some error"))
					})

					It("calls the condition manager to report", func() {
						_, _ = reconciler.Reconcile(ctx, req)
						Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(deliverable.MissingValueAtPathCondition(stampedObject, "this.wont.find.anything")))
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
						Expect(out).To(Say(`"handled error":"unable to retrieve outputs \[this.wont.find.anything\] from stamped object \[my-ns/my-obj\] of type \[mything.thing.io\] for resource \[some-resource\] in delivery \[some-delivery\]: failed to evaluate json path 'this.wont.find.anything': some error"`))
					})

					It("does track the template", func() {
						_, _ = reconciler.Reconcile(ctx, req)
						Expect(dependencyTracker.TrackCallCount()).To(Equal(3))
						key, obj := dependencyTracker.TrackArgsForCall(0)
						Expect(key.String()).To(Equal("ServiceAccount/my-namespace/service-account-name-for-deliverable"))
						Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
						key, obj = dependencyTracker.TrackArgsForCall(1)
						Expect(key.String()).To(Equal("my-image-kind.carto.run//my-image-template"))
						Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
						key, obj = dependencyTracker.TrackArgsForCall(2)
						Expect(key.String()).To(Equal("my-config-kind.carto.run//my-config-template"))
						Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
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
						Expect(out).To(Say(`"handled error":"unable to retrieve outputs from stamped object \[my-ns/my-obj\] of type \[mything.thing.io\] for resource \[some-resource\] in delivery \[some-delivery\]: some error"`))
					})

					It("does track the template", func() {
						_, _ = reconciler.Reconcile(ctx, req)
						Expect(dependencyTracker.TrackCallCount()).To(Equal(3))
						key, obj := dependencyTracker.TrackArgsForCall(0)
						Expect(key.String()).To(Equal("ServiceAccount/my-namespace/service-account-name-for-deliverable"))
						Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
						key, obj = dependencyTracker.TrackArgsForCall(1)
						Expect(key.String()).To(Equal("my-image-kind.carto.run//my-image-template"))
						Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
						key, obj = dependencyTracker.TrackArgsForCall(2)
						Expect(key.String()).To(Equal("my-config-kind.carto.run//my-config-template"))
						Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
					})
				})
			})

			Context("of type ResolveTemplateOptionError", func() {
				var resolveOptionErr realizer.ResolveTemplateOptionError
				BeforeEach(func() {
					jsonPathError := templates.NewJsonPathError("this.wont.find.anything", errors.New("some error"))
					resolveOptionErr = realizer.ResolveTemplateOptionError{
						Err:          jsonPathError,
						DeliveryName: deliveryName,
						Resource:     &v1alpha1.DeliveryResource{Name: "some-resource"},
						OptionName:   "some-option",
						Key:          "some-key",
					}
					rlzr.RealizeReturns(nil, resolveOptionErr)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(
						Equal(deliverable.ResolveTemplateOptionsErrorCondition(resolveOptionErr)))
				})

				It("does not return an error", func() {
					_, err := reconciler.Reconcile(ctx, req)
					Expect(err).NotTo(HaveOccurred())
				})

				It("logs the handled error message", func() {
					_, _ = reconciler.Reconcile(ctx, req)

					Expect(out).To(Say(`"level":"info"`))
					Expect(out).To(Say(`"msg":"handled error reconciling deliverable"`))
					Expect(out).To(Say(`"handled error":"key \[some-key\] is invalid in template option \[some-option\] for resource \[some-resource\] in delivery \[some-delivery\]: failed to evaluate json path 'this.wont.find.anything': some error"`))
				})

				It("does not track the template", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(dependencyTracker.TrackCallCount()).To(Equal(1))
					key, obj := dependencyTracker.TrackArgsForCall(0)
					Expect(key.String()).To(Equal("ServiceAccount/my-namespace/service-account-name-for-deliverable"))
					Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
				})
			})

			Context("of type TemplateOptionsMatchError", func() {
				var templateOptionsMatchErr realizer.TemplateOptionsMatchError
				BeforeEach(func() {
					templateOptionsMatchErr = realizer.TemplateOptionsMatchError{
						DeliveryName: deliveryName,
						Resource:     &v1alpha1.DeliveryResource{Name: "some-resource"},
						OptionNames:  []string{"option1", "option2"},
					}
					rlzr.RealizeReturns(nil, templateOptionsMatchErr)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(
						Equal(deliverable.TemplateOptionsMatchErrorCondition(templateOptionsMatchErr)))
				})

				It("does not return an error", func() {
					_, err := reconciler.Reconcile(ctx, req)
					Expect(err).NotTo(HaveOccurred())
				})

				It("logs the handled error message", func() {
					_, _ = reconciler.Reconcile(ctx, req)

					Expect(out).To(Say(`"level":"info"`))
					Expect(out).To(Say(`"msg":"handled error reconciling deliverable"`))
					Expect(out).To(Say(`"handled error":"expected exactly 1 option to match, found \[2\] matching options \[option1, option2\] for resource \[some-resource\] in delivery \[some-delivery\]"`))
				})

				It("does not track the template", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(dependencyTracker.TrackCallCount()).To(Equal(1))
					key, obj := dependencyTracker.TrackArgsForCall(0)
					Expect(key.String()).To(Equal("ServiceAccount/my-namespace/service-account-name-for-deliverable"))
					Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
				})

				Context("there are no matching options", func() {
					It("logs the handled error message", func() {
						templateOptionsMatchErr.OptionNames = []string{}
						rlzr.RealizeReturns(realizedResources, templateOptionsMatchErr)

						_, _ = reconciler.Reconcile(ctx, req)

						Expect(out).To(Say(`"level":"info"`))
						Expect(out).To(Say(`"msg":"handled error reconciling deliverable"`))
						Expect(out).To(Say(`"handled error":"expected exactly 1 option to match, found \[0\] matching options for resource \[some-resource\] in delivery \[some-delivery\]"`))
					})

					It("does not track the template", func() {
						_, _ = reconciler.Reconcile(ctx, req)
						Expect(dependencyTracker.TrackCallCount()).To(Equal(1))
						key, obj := dependencyTracker.TrackArgsForCall(0)
						Expect(key.String()).To(Equal("ServiceAccount/my-namespace/service-account-name-for-deliverable"))
						Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
					})
				})
			})

			Context("of unknown type", func() {
				var realizerError error
				BeforeEach(func() {
					realizerError = errors.New("some error")
					rlzr.RealizeReturns(realizedResources, realizerError)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(deliverable.UnknownResourceErrorCondition(realizerError)))
				})

				It("returns an unhandled error and requeues", func() {
					_, err := reconciler.Reconcile(ctx, req)

					Expect(err.Error()).To(ContainSubstring("some error"))
				})

				It("does track the template", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(dependencyTracker.TrackCallCount()).To(Equal(3))
					key, obj := dependencyTracker.TrackArgsForCall(0)
					Expect(key.String()).To(Equal("ServiceAccount/my-namespace/service-account-name-for-deliverable"))
					Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
					key, obj = dependencyTracker.TrackArgsForCall(1)
					Expect(key.String()).To(Equal("my-image-kind.carto.run//my-image-template"))
					Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
					key, obj = dependencyTracker.TrackArgsForCall(2)
					Expect(key.String()).To(Equal("my-config-kind.carto.run//my-config-template"))
					Expect(obj.String()).To(Equal("my-namespace/my-deliverable"))
				})
			})
		})

		Context("but the watcher returns an error", func() {
			BeforeEach(func() {
				stampedTracker.WatchReturns(errors.New("could not watch"))
			})

			It("logs the error message", func() {
				_, _ = reconciler.Reconcile(ctx, req)

				Expect(out).To(Say(`"level":"error"`))
				Expect(out).To(Say(`"msg":"failed to add informer for object"`))
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
				Expect(out).To(Say(`"handled error":"failed to get secret for service account \[my-namespace/service-account-name-for-deliverable\]: some error"`))
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

				Expect(err.Error()).To(ContainSubstring("failed to build resource realizer: some error"))
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
			Expect(out).To(Say(`"handled error":"deliverable \[my-namespace/my-deliverable\] is missing required labels"`))
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
			Expect(out).To(Say(`"handled error":"no delivery \[my-namespace/my-deliverable\] found where full selector is satisfied by labels: map\[some-key:some-val\]"`))
		})
	})

	Context("and repo returns an an error when requesting deliveries", func() {
		BeforeEach(func() {
			repo.GetDeliveriesForDeliverableReturns(nil, errors.New("some error"))
		})

		It("returns an unhandled error and requeues", func() {
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err.Error()).To(ContainSubstring("failed to get deliveries for deliverable [my-namespace/my-deliverable]: some error"))
		})
	})

	Context("and the repo returns multiple deliveries", func() {
		BeforeEach(func() {
			delivery := v1alpha1.ClusterDelivery{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-name",
					Namespace: "my-namespace",
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
			Expect(out).To(Say(`"handled error":"more than one delivery selected for deliverable \[my-namespace/my-deliverable\]: \[my-name my-name\]"`))
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
		It("clears the previously tracked objects for the deliverable", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(dependencyTracker.ClearTrackedCallCount()).To(Equal(1))
			obj := dependencyTracker.ClearTrackedArgsForCall(0)
			Expect(obj.Name).To(Equal("my-deliverable-name"))
			Expect(obj.Namespace).To(Equal("my-namespace"))
		})
	})

	Describe("cleaning up orphaned objects", func() {
		BeforeEach(func() {
			delivery := v1alpha1.ClusterDelivery{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-delivery",
				},
				Status: v1alpha1.DeliveryStatus{
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

			rlzr.RealizeReturns([]v1alpha1.RealizedResource{
				{
					Name: "some-resource",
					StampedRef: &corev1.ObjectReference{
						APIVersion: "some-api-version",
						Kind:       "some-kind",
						Name:       "some-new-stamped-obj-name",
					},
				},
			}, nil)
		})
		Context("template does not change so there are no orphaned objects", func() {
			BeforeEach(func() {
				dl.Status.Resources = []v1alpha1.RealizedResource{
					{
						Name: "some-resource",
						StampedRef: &corev1.ObjectReference{
							APIVersion: "some-api-version",
							Kind:       "some-kind",
							Name:       "some-new-stamped-obj-name",
						},
					},
				}
				repo.GetDeliverableReturns(dl, nil)
			})

			It("does not attempt to delete any objects", func() {
				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				Expect(repo.DeleteCallCount()).To(Equal(0))
			})
		})

		Context("a template changes so there are orphaned objects", func() {
			BeforeEach(func() {
				dl.Status.Resources = []v1alpha1.RealizedResource{
					{
						Name: "some-resource",
						StampedRef: &corev1.ObjectReference{
							APIVersion: "some-api-version",
							Kind:       "some-kind",
							Name:       "some-old-stamped-obj-name",
						},
					},
				}
				repo.GetDeliverableReturns(dl, nil)
			})

			It("deletes the orphaned objects", func() {
				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				Expect(repo.DeleteCallCount()).To(Equal(1))

				_, obj := repo.DeleteArgsForCall(0)
				Expect(obj.GetName()).To(Equal("some-old-stamped-obj-name"))
				Expect(obj.GetKind()).To(Equal("some-kind"))
			})

			Context("deleting the object fails", func() {
				BeforeEach(func() {
					repo.DeleteReturns(fmt.Errorf("some error"))
				})

				It("logs an error but does not requeue", func() {
					_, err := reconciler.Reconcile(ctx, req)
					Expect(err).NotTo(HaveOccurred())

					Expect(out).To(Say(`"msg":"failed to cleanup orphaned objects","deliverable":"my-namespace/my-deliverable-name"`))
				})
			})
		})
	})
})
