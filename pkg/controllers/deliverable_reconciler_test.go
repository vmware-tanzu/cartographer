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

package controllers_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"

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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/conditions/conditionsfakes"
	"github.com/vmware-tanzu/cartographer/pkg/controllers"
	cerrors "github.com/vmware-tanzu/cartographer/pkg/errors"
	"github.com/vmware-tanzu/cartographer/pkg/realizer"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/realizerfakes"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/statuses"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
	"github.com/vmware-tanzu/cartographer/pkg/satoken/satokenfakes"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/dependency/dependencyfakes"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/stamped/stampedfakes"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

var _ = Describe("DeliverableReconciler", func() {
	var (
		out               *Buffer
		reconciler        controllers.DeliverableReconciler
		ctx               context.Context
		req               ctrl.Request
		repo              *repositoryfakes.FakeRepository
		tokenManager      *satokenfakes.FakeTokenManager
		conditionManager  *conditionsfakes.FakeConditionManager
		rlzr              *realizerfakes.FakeRealizer
		dl                *v1alpha1.Deliverable
		deliverableLabels map[string]string
		stampedTracker    *stampedfakes.FakeStampedTracker
		dependencyTracker *dependencyfakes.FakeDependencyTracker

		builtResourceRealizer           *realizerfakes.FakeResourceRealizer
		labelerForBuiltResourceRealizer realizer.ResourceLabeler
		resourceRealizerAuthToken       string
		deliverableServiceAccount       *corev1.ServiceAccount
		resourceRealizerBuilderError    error
		deliverableServiceAccountName   = "service-account-name-for-deliverable"
		deliverableServiceAccountToken  = "deliverable-sa-token"
	)

	BeforeEach(func() {
		out = NewBuffer()
		logger := zap.New(zap.WriteTo(out))
		ctx = logr.NewContext(context.Background(), logger)

		conditionManager = &conditionsfakes.FakeConditionManager{}

		fakeConditionManagerBuilder := func(string, []metav1.Condition) conditions.ConditionManager {
			return conditionManager
		}

		rlzr = &realizerfakes.FakeRealizer{}
		rlzr.RealizeReturns(nil)

		stampedTracker = &stampedfakes.FakeStampedTracker{}
		dependencyTracker = &dependencyfakes.FakeDependencyTracker{}

		repo = &repositoryfakes.FakeRepository{}
		scheme := runtime.NewScheme()
		err := utils.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())
		repo.GetSchemeReturns(scheme)

		tokenManager = &satokenfakes.FakeTokenManager{}

		deliverableServiceAccount = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Name: deliverableServiceAccountName},
		}

		repo.GetServiceAccountReturns(deliverableServiceAccount, nil)
		tokenManager.GetServiceAccountTokenReturns(deliverableServiceAccountToken, nil)

		resourceRealizerBuilderError = nil

		resourceRealizerBuilder := func(authToken string, owner client.Object, ownerParams []v1alpha1.OwnerParam, systemRepo repository.Repository, blueprintParams []v1alpha1.BlueprintParam, resourceLabeler realizer.ResourceLabeler) (realizer.ResourceRealizer, error) {
			labelerForBuiltResourceRealizer = resourceLabeler
			if resourceRealizerBuilderError != nil {
				return nil, resourceRealizerBuilderError
			}
			resourceRealizerAuthToken = authToken
			return builtResourceRealizer, nil
		}

		reconciler = controllers.DeliverableReconciler{
			Repo:                    repo,
			TokenManager:            tokenManager,
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
				ServiceAccountName: deliverableServiceAccountName,
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
			deliveryName     string
			delivery         v1alpha1.ClusterDelivery
			resourceStatuses statuses.ResourceStatuses
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

			resourceStatuses = statuses.NewResourceStatuses(nil, conditions.AddConditionForResourceSubmittedDeliverable)
			resourceStatuses.Add(
				&v1alpha1.RealizedResource{
					Name: "resource1",
					StampedRef: &v1alpha1.StampedRef{
						ObjectReference: &corev1.ObjectReference{
							Kind:       "MyThing",
							APIVersion: "thing.io/alphabeta1",
						},
						Resource: "mything",
					},
					TemplateRef: &corev1.ObjectReference{
						Kind:       "my-image-kind",
						Name:       "my-image-template",
						APIVersion: "carto.run/v1alpha1",
					},
				}, nil)
			resourceStatuses.Add(
				&v1alpha1.RealizedResource{
					Name: "resource2",
					StampedRef: &v1alpha1.StampedRef{
						ObjectReference: &corev1.ObjectReference{
							Kind:       "NiceToSeeYou",
							APIVersion: "hello.io/goodbye",
						},
						Resource: "nicetoseeyou",
					},
					TemplateRef: &corev1.ObjectReference{
						Kind:       "my-config-kind",
						Name:       "my-config-template",
						APIVersion: "carto.run/v1alpha1",
					},
				}, nil)

			rlzr.RealizeStub = func(ctx context.Context, resourceRealizer realizer.ResourceRealizer, deliveryName string, resources []realizer.OwnerResource, statuses statuses.ResourceStatuses) error {
				statusesVal := reflect.ValueOf(statuses)
				existingVal := reflect.ValueOf(resourceStatuses)

				reflect.Indirect(statusesVal).Set(reflect.Indirect(existingVal))
				return nil
			}
		})

		It("labels owner resources", func() {
			_, _ = reconciler.Reconcile(ctx, req)
			Expect(labelerForBuiltResourceRealizer).To(Not(BeNil()))

			resource := realizer.OwnerResource{
				TemplateRef: v1alpha1.TemplateReference{
					Kind: "be-kind",
					Name: "no-names",
				},
				Name: "fine-i-have-a-name",
			}

			labels := labelerForBuiltResourceRealizer(resource)
			Expect(labels).To(Equal(templates.Labels{
				"carto.run/deliverable-name":      "my-deliverable",
				"carto.run/deliverable-namespace": "my-namespace",
				"carto.run/delivery-name":         "some-delivery",
				"carto.run/resource-name":         resource.Name,
				"carto.run/template-kind":         resource.TemplateRef.Kind,
				"carto.run/cluster-template-name": resource.TemplateRef.Name,
			}))
		})

		It("updates the status of the owner with the realizedResources", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(repo.StatusUpdateCallCount()).To(Equal(1))
			_, dl := repo.StatusUpdateArgsForCall(0)
			var resource1Status v1alpha1.ResourceStatus
			var resource2Status v1alpha1.ResourceStatus

			currentStatuses := resourceStatuses.GetCurrent()
			Expect(currentStatuses).To(HaveLen(2))

			for i := range currentStatuses {
				switch currentStatuses[i].Name {
				case "resource1":
					resource1Status = currentStatuses[i]
				case "resource2":
					resource2Status = currentStatuses[i]
				}
			}

			var deliverableResource1Status v1alpha1.ResourceStatus
			var deliverableResource2Status v1alpha1.ResourceStatus

			for i := range dl.(*v1alpha1.Deliverable).Status.Resources {
				switch dl.(*v1alpha1.Deliverable).Status.Resources[i].Name {
				case "resource1":
					deliverableResource1Status = dl.(*v1alpha1.Deliverable).Status.Resources[i]
				case "resource2":
					deliverableResource2Status = dl.(*v1alpha1.Deliverable).Status.Resources[i]
				}
			}

			Expect(deliverableResource1Status.RealizedResource).To(Equal(resource1Status.RealizedResource))
			Expect(deliverableResource2Status.RealizedResource).To(Equal(resource2Status.RealizedResource))
		})

		It("dynamically creates a resource realizer", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(rlzr.RealizeCallCount()).To(Equal(1))
			_, resourceRealizer, _, _, _ := rlzr.RealizeArgsForCall(0)
			Expect(resourceRealizer).To(Equal(builtResourceRealizer))
		})

		It("uses the service account specified by the deliverable for realizing resources", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(repo.GetServiceAccountCallCount()).To(Equal(1))
			_, serviceAccountNameArg, serviceAccountNS := repo.GetServiceAccountArgsForCall(0)
			Expect(serviceAccountNameArg).To(Equal(deliverableServiceAccountName))
			Expect(serviceAccountNS).To(Equal("my-namespace"))

			Expect(tokenManager.GetServiceAccountTokenArgsForCall(0)).To(Equal(deliverableServiceAccount))
			Expect(resourceRealizerAuthToken).To(Equal(deliverableServiceAccountToken))
		})

		Context("the deliverable does not specify a service account", func() {
			BeforeEach(func() {
				dl.Spec.ServiceAccountName = ""
			})

			Context("the delivery provides a service account", func() {
				deliveryServiceAccountName := "some-delivery-service-account"
				deliveryServiceAccountToken := "delivery-service-account-token"
				deliveryServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{Name: deliveryServiceAccountName},
				}

				BeforeEach(func() {
					delivery.Spec.ServiceAccountRef.Name = deliveryServiceAccountName

					repo.GetServiceAccountReturns(deliveryServiceAccount, nil)
					tokenManager.GetServiceAccountTokenReturns(deliveryServiceAccountToken, nil)
				})

				It("uses the delivery service account in the deliverable's namespace", func() {
					_, _ = reconciler.Reconcile(ctx, req)

					Expect(repo.GetServiceAccountCallCount()).To(Equal(1))
					_, serviceAccountNameArg, serviceAccountNS := repo.GetServiceAccountArgsForCall(0)
					Expect(serviceAccountNameArg).To(Equal(deliveryServiceAccountName))
					Expect(serviceAccountNS).To(Equal("my-namespace"))

					Expect(tokenManager.GetServiceAccountTokenArgsForCall(0)).To(Equal(deliveryServiceAccount))
					Expect(resourceRealizerAuthToken).To(Equal(deliveryServiceAccountToken))
				})

				Context("the delivery specifies a namespace", func() {
					BeforeEach(func() {
						delivery.Spec.ServiceAccountRef.Namespace = "some-delivery-namespace"
					})

					It("uses the delivery service account in the specified namespace", func() {
						_, _ = reconciler.Reconcile(ctx, req)

						Expect(repo.GetServiceAccountCallCount()).To(Equal(1))
						_, serviceAccountNameArg, serviceAccountNS := repo.GetServiceAccountArgsForCall(0)
						Expect(serviceAccountNameArg).To(Equal(deliveryServiceAccountName))
						Expect(serviceAccountNS).To(Equal("some-delivery-namespace"))

						Expect(tokenManager.GetServiceAccountTokenArgsForCall(0)).To(Equal(deliveryServiceAccount))
						Expect(resourceRealizerAuthToken).To(Equal(deliveryServiceAccountToken))
					})
				})
			})

			Context("the delivery does not provide a service account", func() {
				defaultServiceAccountToken := "default-service-account-token"
				defaultServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{Name: "default"},
				}

				BeforeEach(func() {
					repo.GetServiceAccountReturns(defaultServiceAccount, nil)
					tokenManager.GetServiceAccountTokenReturns(defaultServiceAccountToken, nil)
				})

				It("defaults to the default service account in the deliverables namespace", func() {
					_, _ = reconciler.Reconcile(ctx, req)

					Expect(repo.GetServiceAccountCallCount()).To(Equal(1))
					_, serviceAccountNameArg, serviceAccountNS := repo.GetServiceAccountArgsForCall(0)
					Expect(serviceAccountNameArg).To(Equal("default"))
					Expect(serviceAccountNS).To(Equal("my-namespace"))

					Expect(tokenManager.GetServiceAccountTokenArgsForCall(0)).To(Equal(defaultServiceAccount))
					Expect(resourceRealizerAuthToken).To(Equal(defaultServiceAccountToken))
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
			Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(conditions.DeliveryReadyCondition()))
		})

		It("calls the condition manager to report the resources have been submitted", func() {
			_, _ = reconciler.Reconcile(ctx, req)
			Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(conditions.ResourcesSubmittedCondition(true)))
		})

		It("watches the stampedObjects kinds", func() {
			_, _ = reconciler.Reconcile(ctx, req)
			Expect(stampedTracker.WatchCallCount()).To(Equal(2))

			var gvks []schema.GroupVersionKind

			_, obj, hndl, _ := stampedTracker.WatchArgsForCall(0)
			gvks = append(gvks, obj.GetObjectKind().GroupVersionKind())
			Expect(hndl).To(Equal(&handler.EnqueueRequestForOwner{OwnerType: &v1alpha1.Deliverable{}}))

			_, obj, hndl, _ = stampedTracker.WatchArgsForCall(1)
			gvks = append(gvks, obj.GetObjectKind().GroupVersionKind())
			Expect(hndl).To(Equal(&handler.EnqueueRequestForOwner{OwnerType: &v1alpha1.Deliverable{}}))

			currentStatuses := resourceStatuses.GetCurrent()
			Expect(currentStatuses).To(HaveLen(2))
			var resource1Status v1alpha1.ResourceStatus
			var resource2Status v1alpha1.ResourceStatus

			for i := range currentStatuses {
				switch currentStatuses[i].Name {
				case "resource1":
					resource1Status = currentStatuses[i]
				case "resource2":
					resource2Status = currentStatuses[i]
				}
			}

			Expect(gvks).To(ContainElements(resource1Status.StampedRef.GetObjectKind().GroupVersionKind(),
				resource2Status.StampedRef.GetObjectKind().GroupVersionKind()))
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
				Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(conditions.MissingReadyInDeliveryCondition(expectedCondition)))
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
					templateError = cerrors.GetTemplateError{
						ResourceName: "some-name",
						Err:          errors.New("some error"),
					}
					rlzr.RealizeReturns(templateError)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(conditions.TemplateObjectRetrievalFailureCondition(true, templateError)))
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
				var stampError cerrors.StampError
				BeforeEach(func() {
					stampError = cerrors.StampError{
						Err:           errors.New("some error"),
						ResourceName:  "some-name",
						BlueprintName: "some-delivery",
						BlueprintType: cerrors.Delivery,
						TemplateName:  "some-template",
						TemplateKind:  "ClusterDeploymentTemplate",
					}
					resourceStatuses.Add(
						&v1alpha1.RealizedResource{
							Name:       "resource1",
							StampedRef: nil,
							TemplateRef: &corev1.ObjectReference{
								Kind:       "my-image-kind",
								Name:       "my-image-template",
								APIVersion: "carto.run/v1alpha1",
							},
						}, nil,
					)
					resourceStatuses.Add(
						&v1alpha1.RealizedResource{
							Name:       "resource2",
							StampedRef: nil,
							TemplateRef: &corev1.ObjectReference{
								Kind:       "my-config-kind",
								Name:       "my-config-template",
								APIVersion: "carto.run/v1alpha1",
							},
						}, nil,
					)

					rlzr.RealizeStub = func(ctx context.Context, resourceRealizer realizer.ResourceRealizer, deliveryName string, resources []realizer.OwnerResource, statuses statuses.ResourceStatuses) error {
						statusesVal := reflect.ValueOf(statuses)
						existingVal := reflect.ValueOf(resourceStatuses)

						reflect.Indirect(statusesVal).Set(reflect.Indirect(existingVal))
						return stampError
					}
				})

				It("does not try to watch the stampedObjects", func() {
					_, err := reconciler.Reconcile(ctx, req)
					Expect(err).NotTo(HaveOccurred())

					Expect(stampedTracker.WatchCallCount()).To(Equal(0))
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(conditions.TemplateStampFailureCondition(true, stampError)))
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
					Expect(out).To(Say(`"handled error":"unable to stamp object for resource \[some-name\] for template \[ClusterDeploymentTemplate/some-template\] in delivery \[some-delivery\]: some error"`))
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
				var stampedObjectError cerrors.ApplyStampedObjectError
				BeforeEach(func() {
					stampedObjectError = cerrors.ApplyStampedObjectError{
						Err:           errors.New("some error"),
						StampedObject: &unstructured.Unstructured{},
						ResourceName:  "some-name",
					}
					rlzr.RealizeStub = func(ctx context.Context, resourceRealizer realizer.ResourceRealizer, deliveryName string, resources []realizer.OwnerResource, statuses statuses.ResourceStatuses) error {
						statusesVal := reflect.ValueOf(statuses)
						existingVal := reflect.ValueOf(resourceStatuses)

						reflect.Indirect(statusesVal).Set(reflect.Indirect(existingVal))
						return stampedObjectError
					}
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(conditions.TemplateRejectedByAPIServerCondition(true, stampedObjectError)))
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
				var stampedObjectError cerrors.ApplyStampedObjectError
				BeforeEach(func() {
					status := &metav1.Status{
						Message: "fantastic error",
						Reason:  metav1.StatusReasonForbidden,
						Code:    403,
					}
					stampedObject1 := &unstructured.Unstructured{}
					stampedObject1.SetNamespace("a-namespace")
					stampedObject1.SetName("a-name")

					stampedObjectError = cerrors.ApplyStampedObjectError{
						Err:           kerrors.FromObject(status),
						StampedObject: stampedObject1,
						ResourceName:  "some-name",
						BlueprintName: deliveryName,
						BlueprintType: cerrors.Delivery,
					}

					rlzr.RealizeStub = func(ctx context.Context, resourceRealizer realizer.ResourceRealizer, deliveryName string, resources []realizer.OwnerResource, statuses statuses.ResourceStatuses) error {
						statusesVal := reflect.ValueOf(statuses)
						existingVal := reflect.ValueOf(resourceStatuses)

						reflect.Indirect(statusesVal).Set(reflect.Indirect(existingVal))
						return stampedObjectError
					}
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(conditions.TemplateRejectedByAPIServerCondition(true, stampedObjectError)))
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
				var retrieveError cerrors.RetrieveOutputError
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

					retrieveError = cerrors.RetrieveOutputError{
						Err:           wrappedError,
						ResourceName:  "some-resource",
						BlueprintName: deliveryName,
						BlueprintType: cerrors.Delivery,
						StampedObject: stampedObject,
						ResourceType:  "mything.thing.io",
					}

					rlzr.RealizeStub = func(ctx context.Context, resourceRealizer realizer.ResourceRealizer, deliveryName string, resources []realizer.OwnerResource, statuses statuses.ResourceStatuses) error {
						statusesVal := reflect.ValueOf(statuses)
						existingVal := reflect.ValueOf(resourceStatuses)

						reflect.Indirect(statusesVal).Set(reflect.Indirect(existingVal))
						return retrieveError
					}
				})

				Context("which wraps an ObservedGenerationError", func() {
					BeforeEach(func() {
						wrappedError = templates.NewObservedGenerationError(errors.New("some error"))
					})

					It("calls the condition manager to report", func() {
						_, _ = reconciler.Reconcile(ctx, req)
						Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(conditions.TemplateStampFailureByObservedGenerationCondition(retrieveError)))
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
						Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(conditions.DeploymentConditionNotMetCondition(retrieveError)))
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
						Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(conditions.DeploymentFailedConditionMetCondition(retrieveError)))
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
						Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(conditions.MissingValueAtPathCondition(true, stampedObject, "this.wont.find.anything", "mything.thing.io")))
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
						Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(conditions.UnknownResourceErrorCondition(true, retrieveError)))
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
				var resolveOptionErr cerrors.ResolveTemplateOptionError
				BeforeEach(func() {
					jsonPathError := templates.NewJsonPathError("this.wont.find.anything", errors.New("some error"))
					resolveOptionErr = cerrors.ResolveTemplateOptionError{
						Err:           jsonPathError,
						BlueprintName: deliveryName,
						BlueprintType: cerrors.Delivery,
						ResourceName:  "some-resource",
						OptionName:    "some-option",
					}
					rlzr.RealizeReturns(resolveOptionErr)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(
						Equal(conditions.ResolveTemplateOptionsErrorCondition(true, resolveOptionErr)))
				})

				It("does not return an error", func() {
					_, err := reconciler.Reconcile(ctx, req)
					Expect(err).NotTo(HaveOccurred())
				})

				It("logs the handled error message", func() {
					_, _ = reconciler.Reconcile(ctx, req)

					Expect(out).To(Say(`"level":"info"`))
					Expect(out).To(Say(`"msg":"handled error reconciling deliverable"`))
					Expect(out).To(Say(`"handled error":"error matching against template option \[some-option\] for resource \[some-resource\] in delivery \[some-delivery\]: failed to evaluate json path 'this.wont.find.anything': some error"`))
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
				var templateOptionsMatchErr cerrors.TemplateOptionsMatchError
				BeforeEach(func() {
					templateOptionsMatchErr = cerrors.TemplateOptionsMatchError{
						BlueprintName: deliveryName,
						BlueprintType: cerrors.Delivery,
						ResourceName:  "some-resource",
						OptionNames:   []string{"option1", "option2"},
					}
					rlzr.RealizeReturns(templateOptionsMatchErr)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(
						Equal(conditions.TemplateOptionsMatchErrorCondition(true, templateOptionsMatchErr)))
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
						rlzr.RealizeReturns(templateOptionsMatchErr)

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
					rlzr.RealizeStub = func(ctx context.Context, resourceRealizer realizer.ResourceRealizer, deliveryName string, resources []realizer.OwnerResource, statuses statuses.ResourceStatuses) error {
						statusesVal := reflect.ValueOf(statuses)
						existingVal := reflect.ValueOf(resourceStatuses)

						reflect.Indirect(statusesVal).Set(reflect.Indirect(existingVal))
						return realizerError
					}
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(conditions.UnknownResourceErrorCondition(true, realizerError)))
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

		Context("but the repo returns an error when requesting the service account", func() {
			var repoError error
			BeforeEach(func() {
				repoError = errors.New("some error")
				repo.GetServiceAccountReturns(nil, repoError)
			})

			It("calls the condition manager to add a service account not found condition", func() {
				_, _ = reconciler.Reconcile(ctx, req)
				Expect(conditionManager.AddPositiveCallCount()).To(BeNumerically(">", 1))
				Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(conditions.ServiceAccountNotFoundCondition(repoError)))
			})

			It("handles the error and logs it", func() {
				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				Expect(out).To(Say(`"level":"info"`))
				Expect(out).To(Say(`"handled error":"failed to get service account \[my-namespace/service-account-name-for-deliverable\]: some error"`))
			})
		})

		Context("but the token manager returns an error when requesting a token for the service account", func() {
			var tokenError error
			BeforeEach(func() {
				tokenError = errors.New("some error")
				tokenManager.GetServiceAccountTokenReturns("", tokenError)
			})

			It("calls the condition manager to add a service account not found condition", func() {
				_, _ = reconciler.Reconcile(ctx, req)
				Expect(conditionManager.AddPositiveCallCount()).To(BeNumerically(">", 1))
				Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(conditions.ServiceAccountTokenErrorCondition(tokenError)))
			})

			It("handles the error and logs it", func() {
				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				Expect(out).To(Say(`"level":"info"`))
				Expect(out).To(Say(`"handled error":"failed to get token for service account \[my-namespace/service-account-name-for-deliverable\]: some error"`))
			})
		})

		Context("but the resource realizer builder fails", func() {
			BeforeEach(func() {
				resourceRealizerBuilderError = errors.New("some error")
			})

			It("calls the condition manager to add a resource realizer builder error condition", func() {
				_, _ = reconciler.Reconcile(ctx, req)
				Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(conditions.ResourceRealizerBuilderErrorCondition(resourceRealizerBuilderError)))
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
			Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(conditions.DeliverableMissingLabelsCondition()))
		})
	})

	Context("and repo returns an empty list of deliveries", func() {
		It("does not return an error", func() {
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
		})

		It("calls the condition manager to add a delivery not found condition", func() {
			_, _ = reconciler.Reconcile(ctx, req)
			Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(conditions.DeliveryNotFoundCondition(deliverableLabels)))
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
			Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(conditions.TooManyDeliveryMatchesCondition()))
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

			rlzr.RealizeReturns(nil)

			resourceStatuses := statuses.NewResourceStatuses(nil, conditions.AddConditionForResourceSubmittedDeliverable)
			resourceStatuses.Add(
				&v1alpha1.RealizedResource{
					Name: "some-resource",
					StampedRef: &v1alpha1.StampedRef{
						ObjectReference: &corev1.ObjectReference{
							APIVersion: "some-api-version",
							Kind:       "some-kind",
							Name:       "some-new-stamped-obj-name",
						},
						Resource: "some-kind",
					},
				}, nil,
			)
			rlzr.RealizeStub = func(ctx context.Context, resourceRealizer realizer.ResourceRealizer, deliveryName string, resources []realizer.OwnerResource, statuses statuses.ResourceStatuses) error {
				statusesVal := reflect.ValueOf(statuses)
				existingVal := reflect.ValueOf(resourceStatuses)

				reflect.Indirect(statusesVal).Set(reflect.Indirect(existingVal))
				return nil
			}
		})
		Context("template does not change so there are no orphaned objects", func() {
			BeforeEach(func() {
				dl.Status.Resources = []v1alpha1.ResourceStatus{
					{
						RealizedResource: v1alpha1.RealizedResource{
							Name: "some-resource",
							StampedRef: &v1alpha1.StampedRef{
								ObjectReference: &corev1.ObjectReference{
									APIVersion: "some-api-version",
									Kind:       "some-kind",
									Name:       "some-new-stamped-obj-name",
								},
								Resource: "some-kind",
							},
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
				dl.Status.Resources = []v1alpha1.ResourceStatus{
					{
						RealizedResource: v1alpha1.RealizedResource{
							Name: "some-resource",
							StampedRef: &v1alpha1.StampedRef{
								ObjectReference: &corev1.ObjectReference{
									APIVersion: "some-api-version",
									Kind:       "some-kind",
									Name:       "some-old-stamped-obj-name",
								},
								Resource: "some-kind",
							},
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
