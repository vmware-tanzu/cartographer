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

package workload_test

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
	"github.com/vmware-tanzu/cartographer/pkg/controller/workload"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/workload"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/workload/workloadfakes"
	"github.com/vmware-tanzu/cartographer/pkg/registrar"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/dependency/dependencyfakes"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/stamped/stampedfakes"
)

var _ = Describe("Reconciler", func() {
	var (
		out                          *Buffer
		reconciler                   workload.Reconciler
		ctx                          context.Context
		req                          ctrl.Request
		repo                         *repositoryfakes.FakeRepository
		conditionManager             *conditionsfakes.FakeConditionManager
		rlzr                         *workloadfakes.FakeRealizer
		wl                           *v1alpha1.Workload
		workloadLabels               map[string]string
		stampedTracker               *stampedfakes.FakeStampedTracker
		dependencyTracker            *dependencyfakes.FakeDependencyTracker
		builtResourceRealizer        *workloadfakes.FakeResourceRealizer
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

		rlzr = &workloadfakes.FakeRealizer{}
		rlzr.RealizeReturns(nil, nil)

		stampedTracker = &stampedfakes.FakeStampedTracker{}
		dependencyTracker = &dependencyfakes.FakeDependencyTracker{}

		repo = &repositoryfakes.FakeRepository{}
		scheme := runtime.NewScheme()
		err := registrar.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())
		repo.GetSchemeReturns(scheme)

		serviceAccountName = "workload-service-account-name"

		serviceAccountSecret = &corev1.Secret{Data: map[string][]byte{"token": []byte(`blahblah`)}}
		repo.GetServiceAccountSecretReturns(serviceAccountSecret, nil)

		resourceRealizerBuilderError = nil
		resourceRealizerBuilder := func(secret *corev1.Secret, workload *v1alpha1.Workload, systemRepo repository.Repository, supplyChainParams []v1alpha1.BlueprintParam) (realizer.ResourceRealizer, error) {
			if resourceRealizerBuilderError != nil {
				return nil, resourceRealizerBuilderError
			}
			resourceRealizerSecret = secret
			builtResourceRealizer = &workloadfakes.FakeResourceRealizer{}
			return builtResourceRealizer, nil
		}

		reconciler = workload.Reconciler{
			Repo:                    repo,
			ConditionManagerBuilder: fakeConditionManagerBuilder,
			ResourceRealizerBuilder: resourceRealizerBuilder,
			Realizer:                rlzr,
			StampedTracker:          stampedTracker,
			DependencyTracker:       dependencyTracker,
		}

		req = ctrl.Request{
			NamespacedName: types.NamespacedName{Name: "my-workload-name", Namespace: "my-namespace"},
		}

		workloadLabels = map[string]string{"some-key": "some-val"}

		wl = &v1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Generation: 1,
				Labels:     workloadLabels,
				Name:       "my-workload-name",
				Namespace:  "my-namespace",
			},
			Spec: v1alpha1.WorkloadSpec{
				ServiceAccountName: serviceAccountName,
			},
		}
		repo.GetWorkloadReturns(wl, nil)
	})

	It("logs that it's begun", func() {
		_, _ = reconciler.Reconcile(ctx, req)

		Expect(out).To(Say(`"msg":"started"`))
	})

	It("logs that it's finished", func() {
		_, _ = reconciler.Reconcile(ctx, req)

		Expect(out).To(Say(`"msg":"finished"`))
	})

	It("updates the status of the workload", func() {
		_, _ = reconciler.Reconcile(ctx, req)

		Expect(repo.StatusUpdateCallCount()).To(Equal(1))
	})

	It("updates the status.observedGeneration to equal metadata.generation", func() {
		_, _ = reconciler.Reconcile(ctx, req)

		_, updatedWorkload := repo.StatusUpdateArgsForCall(0)

		Expect(*updatedWorkload.(*v1alpha1.Workload)).To(MatchFields(IgnoreExtras, Fields{
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

		_, updatedWorkload := repo.StatusUpdateArgsForCall(0)
		Expect(*updatedWorkload.(*v1alpha1.Workload)).To(MatchFields(IgnoreExtras, Fields{
			"Status": MatchFields(IgnoreExtras, Fields{
				"OwnerStatus": MatchFields(IgnoreExtras, Fields{
					"Conditions": Equal(someConditions),
				}),
			}),
		}))
	})

	It("requests supply chains from the repo", func() {
		_, _ = reconciler.Reconcile(ctx, req)
		_, workload := repo.GetSupplyChainsForWorkloadArgsForCall(0)
		Expect(workload).To(Equal(wl))
	})

	Context("and the repo returns a single matching supply-chain for the workload", func() {
		var (
			supplyChainName   string
			supplyChain       v1alpha1.ClusterSupplyChain
			realizedResources []v1alpha1.RealizedResource
		)
		BeforeEach(func() {
			supplyChainName = "some-supply-chain"
			supplyChain = v1alpha1.ClusterSupplyChain{
				ObjectMeta: metav1.ObjectMeta{
					Name: supplyChainName,
				},
				Status: v1alpha1.SupplyChainStatus{
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
			repo.GetSupplyChainsForWorkloadReturns([]*v1alpha1.ClusterSupplyChain{&supplyChain}, nil)

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
			_, wk := repo.StatusUpdateArgsForCall(0)
			Expect(wk.(*v1alpha1.Workload).Status.Resources).To(Equal(realizedResources))
		})

		It("dynamically creates a resource realizer", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(rlzr.RealizeCallCount()).To(Equal(1))
			_, resourceRealizer, _, _ := rlzr.RealizeArgsForCall(0)
			Expect(resourceRealizer).To(Equal(builtResourceRealizer))
		})

		It("uses the service account specified by the workload for realizing resources", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(repo.GetServiceAccountSecretCallCount()).To(Equal(1))
			_, serviceAccountNameArg, serviceAccountNS := repo.GetServiceAccountSecretArgsForCall(0)
			Expect(serviceAccountNameArg).To(Equal(serviceAccountName))
			Expect(serviceAccountNS).To(Equal("my-namespace"))

			Expect(resourceRealizerSecret).To(Equal(serviceAccountSecret))
		})

		Context("the workload does not specify a service account", func() {
			BeforeEach(func() {
				wl.Spec.ServiceAccountName = ""
			})

			Context("the supply chain provides a service account", func() {
				var supplyChainServiceAccountSecret *corev1.Secret

				BeforeEach(func() {
					supplyChain.Spec.ServiceAccountRef.Name = "some-supply-chain-service-account"

					supplyChainServiceAccountSecret = &corev1.Secret{Data: map[string][]byte{"token": []byte(`some-sc-service-account-token`)}}
					repo.GetServiceAccountSecretReturns(supplyChainServiceAccountSecret, nil)
				})

				It("uses the supply chain service account in the workload's namespace", func() {
					_, _ = reconciler.Reconcile(ctx, req)

					Expect(repo.GetServiceAccountSecretCallCount()).To(Equal(1))
					_, serviceAccountNameArg, serviceAccountNS := repo.GetServiceAccountSecretArgsForCall(0)
					Expect(serviceAccountNameArg).To(Equal("some-supply-chain-service-account"))
					Expect(serviceAccountNS).To(Equal("my-namespace"))

					Expect(resourceRealizerSecret).To(Equal(supplyChainServiceAccountSecret))
				})

				Context("the supply chain specifies a namespace", func() {
					BeforeEach(func() {
						supplyChain.Spec.ServiceAccountRef.Namespace = "some-supply-chain-namespace"
					})

					It("uses the supply chain service account in the specified namespace", func() {
						_, _ = reconciler.Reconcile(ctx, req)

						Expect(repo.GetServiceAccountSecretCallCount()).To(Equal(1))
						_, serviceAccountNameArg, serviceAccountNS := repo.GetServiceAccountSecretArgsForCall(0)
						Expect(serviceAccountNameArg).To(Equal("some-supply-chain-service-account"))
						Expect(serviceAccountNS).To(Equal("some-supply-chain-namespace"))

						Expect(resourceRealizerSecret).To(Equal(supplyChainServiceAccountSecret))
					})
				})
			})

			Context("the supply chain does not provide a service account", func() {
				var defaultServiceAccountSecret *corev1.Secret

				BeforeEach(func() {
					defaultServiceAccountSecret = &corev1.Secret{Data: map[string][]byte{"token": []byte(`some-default-service-account-token`)}}
					repo.GetServiceAccountSecretReturns(defaultServiceAccountSecret, nil)
				})

				It("defaults to the default service account in the workloads namespace", func() {
					_, _ = reconciler.Reconcile(ctx, req)

					Expect(repo.GetServiceAccountSecretCallCount()).To(Equal(1))
					_, serviceAccountNameArg, serviceAccountNS := repo.GetServiceAccountSecretArgsForCall(0)
					Expect(serviceAccountNameArg).To(Equal("default"))
					Expect(serviceAccountNS).To(Equal("my-namespace"))

					Expect(resourceRealizerSecret).To(Equal(defaultServiceAccountSecret))
				})
			})
		})

		It("sets the SupplyChainRef", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(wl.Status.SupplyChainRef.Kind).To(Equal("ClusterSupplyChain"))
			Expect(wl.Status.SupplyChainRef.Name).To(Equal(supplyChainName))
		})

		It("calls the condition manager to specify the supply chain is ready", func() {
			_, _ = reconciler.Reconcile(ctx, req)
			Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(workload.SupplyChainReadyCondition()))
		})

		It("calls the condition manager to report the resources have been submitted", func() {
			_, _ = reconciler.Reconcile(ctx, req)
			Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(workload.ResourcesSubmittedCondition()))
		})

		It("watches the stampedObjects kinds", func() {
			_, _ = reconciler.Reconcile(ctx, req)
			Expect(stampedTracker.WatchCallCount()).To(Equal(2))
			_, obj, hndl, _ := stampedTracker.WatchArgsForCall(0)

			Expect(obj.GetObjectKind().GroupVersionKind().Kind).To(Equal(realizedResources[0].StampedRef.GetObjectKind().GroupVersionKind().Kind))
			Expect(obj.GetObjectKind().GroupVersionKind().Version).To(Equal(realizedResources[0].StampedRef.GetObjectKind().GroupVersionKind().Version))
			Expect(hndl).To(Equal(&handler.EnqueueRequestForOwner{OwnerType: &v1alpha1.Workload{}}))

			_, obj, hndl, _ = stampedTracker.WatchArgsForCall(1)

			Expect(obj.GetObjectKind().GroupVersionKind().Kind).To(Equal(realizedResources[1].StampedRef.GetObjectKind().GroupVersionKind().Kind))
			Expect(obj.GetObjectKind().GroupVersionKind().Version).To(Equal(realizedResources[1].StampedRef.GetObjectKind().GroupVersionKind().Version))
			Expect(hndl).To(Equal(&handler.EnqueueRequestForOwner{OwnerType: &v1alpha1.Workload{}}))
		})

		It("clears the previously tracked objects for the workload", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(dependencyTracker.ClearTrackedCallCount()).To(Equal(1))
			obj := dependencyTracker.ClearTrackedArgsForCall(0)
			Expect(obj.Name).To(Equal("my-workload-name"))
			Expect(obj.Namespace).To(Equal("my-namespace"))
		})

		It("watches the templates and service account", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(dependencyTracker.TrackCallCount()).To(Equal(3))
			serviceAccountKey, _ := dependencyTracker.TrackArgsForCall(0)
			Expect(serviceAccountKey.String()).To(Equal("ServiceAccount/my-namespace/workload-service-account-name"))

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

				Expect(err.Error()).To(ContainSubstring("failed to get object gvk for supply chain [some-supply-chain]: "))
			})
		})

		Context("but the supply chain is not in a ready state", func() {
			BeforeEach(func() {
				supplyChain.Status.Conditions = []metav1.Condition{
					{
						Type:    "Ready",
						Status:  "False",
						Reason:  "SomeReason",
						Message: "some informative message",
					},
				}
				repo.GetSupplyChainsForWorkloadReturns([]*v1alpha1.ClusterSupplyChain{&supplyChain}, nil)
			})

			It("does not return an error", func() {
				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
			})

			It("calls the condition manager to report supply chain not ready", func() {
				_, _ = reconciler.Reconcile(ctx, req)
				expectedCondition := metav1.Condition{
					Type:               v1alpha1.WorkloadSupplyChainReady,
					Status:             metav1.ConditionFalse,
					ObservedGeneration: 0,
					LastTransitionTime: metav1.Time{},
					Reason:             "SomeReason",
					Message:            "some informative message",
				}
				Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(workload.MissingReadyInSupplyChainCondition(expectedCondition)))
			})

			It("logs the handled error message", func() {
				_, _ = reconciler.Reconcile(ctx, req)

				Expect(out).To(Say(`"level":"info"`))
				Expect(out).To(Say(`"msg":"handled error reconciling workload"`))
				Expect(out).To(Say(`"handled error":"supply chain \[some-supply-chain\] is not in ready state"`))
			})
		})

		Context("but the realizer returns an error", func() {
			Context("of type GetTemplateError", func() {
				var templateError error
				BeforeEach(func() {
					templateError = realizer.GetTemplateError{
						Err:      errors.New("some error"),
						Resource: &v1alpha1.SupplyChainResource{Name: "some-name"},
					}
					rlzr.RealizeReturns(nil, templateError)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(workload.TemplateObjectRetrievalFailureCondition(templateError)))
				})

				It("returns an unhandled error and requeues", func() {
					_, err := reconciler.Reconcile(ctx, req)

					Expect(err.Error()).To(ContainSubstring("unable to get template"))
				})

				It("does not track the template", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(dependencyTracker.TrackCallCount()).To(Equal(1))
					key, obj := dependencyTracker.TrackArgsForCall(0)
					Expect(key.String()).To(Equal("ServiceAccount/my-namespace/workload-service-account-name"))
					Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))
				})
			})

			Context("of type StampError", func() {
				var stampError realizer.StampError
				BeforeEach(func() {
					stampError = realizer.StampError{
						Err:             errors.New("some error"),
						Resource:        &v1alpha1.SupplyChainResource{Name: "some-name"},
						SupplyChainName: supplyChainName,
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
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(workload.TemplateStampFailureCondition(stampError)))
				})

				It("does not return an error", func() {
					_, err := reconciler.Reconcile(ctx, req)
					Expect(err).NotTo(HaveOccurred())
				})

				It("logs the handled error message", func() {
					_, _ = reconciler.Reconcile(ctx, req)

					Expect(out).To(Say(`"level":"info"`))
					Expect(out).To(Say(`"msg":"handled error reconciling workload"`))
					Expect(out).To(Say(`"handled error":"unable to stamp object for resource \[some-name\] in supply chain \[some-supply-chain\]: some error"`))
				})

				It("does track the template", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(dependencyTracker.TrackCallCount()).To(Equal(3))
					key, obj := dependencyTracker.TrackArgsForCall(0)
					Expect(key.String()).To(Equal("ServiceAccount/my-namespace/workload-service-account-name"))
					Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))
					key, obj = dependencyTracker.TrackArgsForCall(1)
					Expect(key.String()).To(Equal("my-image-kind.carto.run//my-image-template"))
					Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))
					key, obj = dependencyTracker.TrackArgsForCall(2)
					Expect(key.String()).To(Equal("my-config-kind.carto.run//my-config-template"))
					Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))
				})
			})

			Context("of type ApplyStampedObjectError", func() {
				var stampedObjectError realizer.ApplyStampedObjectError
				BeforeEach(func() {
					stampedObjectError = realizer.ApplyStampedObjectError{
						Err:           errors.New("some error"),
						StampedObject: &unstructured.Unstructured{},
						Resource:      &v1alpha1.SupplyChainResource{Name: "some-name"},
					}
					rlzr.RealizeReturns(realizedResources, stampedObjectError)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(workload.TemplateRejectedByAPIServerCondition(stampedObjectError)))
				})

				It("returns an unhandled error and requeues", func() {
					_, err := reconciler.Reconcile(ctx, req)

					Expect(err.Error()).To(ContainSubstring("unable to apply object"))
				})

				It("does track the template", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(dependencyTracker.TrackCallCount()).To(Equal(3))
					key, obj := dependencyTracker.TrackArgsForCall(0)
					Expect(key.String()).To(Equal("ServiceAccount/my-namespace/workload-service-account-name"))
					Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))
					key, obj = dependencyTracker.TrackArgsForCall(1)
					Expect(key.String()).To(Equal("my-image-kind.carto.run//my-image-template"))
					Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))
					key, obj = dependencyTracker.TrackArgsForCall(2)
					Expect(key.String()).To(Equal("my-config-kind.carto.run//my-config-template"))
					Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))
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
						Err:             kerrors.FromObject(status),
						StampedObject:   stampedObject1,
						Resource:        &v1alpha1.SupplyChainResource{Name: "some-name"},
						SupplyChainName: supplyChainName,
					}

					rlzr.RealizeReturns(realizedResources, stampedObjectError)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(workload.TemplateRejectedByAPIServerCondition(stampedObjectError)))
				})

				It("handles the error and logs it", func() {
					_, err := reconciler.Reconcile(ctx, req)
					Expect(err).NotTo(HaveOccurred())

					Expect(out).To(Say(`"level":"info"`))
					Expect(out).To(Say(`"handled error":"unable to apply object \[a-namespace/a-name\] for resource \[some-name\] in supply chain \[some-supply-chain\]: fantastic error"`))
				})

				It("does track the template", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(dependencyTracker.TrackCallCount()).To(Equal(3))
					key, obj := dependencyTracker.TrackArgsForCall(0)
					Expect(key.String()).To(Equal("ServiceAccount/my-namespace/workload-service-account-name"))
					Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))
					key, obj = dependencyTracker.TrackArgsForCall(1)
					Expect(key.String()).To(Equal("my-image-kind.carto.run//my-image-template"))
					Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))
					key, obj = dependencyTracker.TrackArgsForCall(2)
					Expect(key.String()).To(Equal("my-config-kind.carto.run//my-config-template"))
					Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))
				})
			})

			Context("of type RetrieveOutputError", func() {
				var retrieveError realizer.RetrieveOutputError
				var stampedObject *unstructured.Unstructured
				BeforeEach(func() {
					stampedObject = &unstructured.Unstructured{}
					stampedObject.SetGroupVersionKind(schema.GroupVersionKind{
						Group:   "thing.io",
						Version: "alphabeta1",
						Kind:    "MyThing",
					})
					stampedObject.SetName("my-obj")
					stampedObject.SetNamespace("my-ns")
					jsonPathError := templates.NewJsonPathError("this.wont.find.anything", errors.New("some error"))
					retrieveError = realizer.RetrieveOutputError{
						Err:             jsonPathError,
						Resource:        &v1alpha1.SupplyChainResource{Name: "some-resource"},
						StampedObject:   stampedObject,
						SupplyChainName: supplyChainName,
					}
					rlzr.RealizeReturns(realizedResources, retrieveError)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(
						Equal(workload.MissingValueAtPathCondition(stampedObject, "this.wont.find.anything")))
				})

				It("does not return an error", func() {
					_, err := reconciler.Reconcile(ctx, req)
					Expect(err).NotTo(HaveOccurred())
				})

				It("logs the handled error message", func() {
					_, _ = reconciler.Reconcile(ctx, req)

					Expect(out).To(Say(`"level":"info"`))
					Expect(out).To(Say(`"msg":"handled error reconciling workload"`))
					Expect(out).To(Say(`"handled error":"unable to retrieve outputs \[this.wont.find.anything\] from stamped object \[my-ns/my-obj\] of type \[mything.thing.io\] for resource \[some-resource\] in supply chain \[some-supply-chain\]: failed to evaluate json path 'this.wont.find.anything': some error"`))
				})

				It("does track the template", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(dependencyTracker.TrackCallCount()).To(Equal(3))
					key, obj := dependencyTracker.TrackArgsForCall(0)
					Expect(key.String()).To(Equal("ServiceAccount/my-namespace/workload-service-account-name"))
					Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))
					key, obj = dependencyTracker.TrackArgsForCall(1)
					Expect(key.String()).To(Equal("my-image-kind.carto.run//my-image-template"))
					Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))
					key, obj = dependencyTracker.TrackArgsForCall(2)
					Expect(key.String()).To(Equal("my-config-kind.carto.run//my-config-template"))
					Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))
				})
			})

			Context("of type ResolveTemplateOptionError", func() {
				var resolveOptionErr realizer.ResolveTemplateOptionError
				BeforeEach(func() {
					jsonPathError := templates.NewJsonPathError("this.wont.find.anything", errors.New("some error"))
					resolveOptionErr = realizer.ResolveTemplateOptionError{
						Err:             jsonPathError,
						SupplyChainName: supplyChainName,
						Resource:        &v1alpha1.SupplyChainResource{Name: "some-resource"},
						OptionName:      "some-option",
					}
					rlzr.RealizeReturns(nil, resolveOptionErr)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(
						Equal(workload.ResolveTemplateOptionsErrorCondition(resolveOptionErr)))
				})

				It("does not return an error", func() {
					_, err := reconciler.Reconcile(ctx, req)
					Expect(err).NotTo(HaveOccurred())
				})

				It("logs the handled error message", func() {
					_, _ = reconciler.Reconcile(ctx, req)

					Expect(out).To(Say(`"level":"info"`))
					Expect(out).To(Say(`"msg":"handled error reconciling workload"`))
					Expect(out).To(Say(`"handled error":"error matching against template option \[some-option\] for resource \[some-resource\] in supply chain \[some-supply-chain\]: failed to evaluate json path 'this.wont.find.anything': some error"`))
				})

				It("does not track the template", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(dependencyTracker.TrackCallCount()).To(Equal(1))
					key, obj := dependencyTracker.TrackArgsForCall(0)
					Expect(key.String()).To(Equal("ServiceAccount/my-namespace/workload-service-account-name"))
					Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))
				})
			})

			Context("of type TemplateOptionsMatchError", func() {
				var templateOptionsMatchErr realizer.TemplateOptionsMatchError
				BeforeEach(func() {
					templateOptionsMatchErr = realizer.TemplateOptionsMatchError{
						SupplyChainName: supplyChainName,
						Resource:        &v1alpha1.SupplyChainResource{Name: "some-resource"},
						OptionNames:     []string{"option1", "option2"},
					}
					rlzr.RealizeReturns(nil, templateOptionsMatchErr)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(
						Equal(workload.TemplateOptionsMatchErrorCondition(templateOptionsMatchErr)))
				})

				It("does not return an error", func() {
					_, err := reconciler.Reconcile(ctx, req)
					Expect(err).NotTo(HaveOccurred())
				})

				It("logs the handled error message", func() {
					_, _ = reconciler.Reconcile(ctx, req)

					Expect(out).To(Say(`"level":"info"`))
					Expect(out).To(Say(`"msg":"handled error reconciling workload"`))
					Expect(out).To(Say(`"handled error":"expected exactly 1 option to match, found \[2\] matching options \[option1, option2\] for resource \[some-resource\] in supply chain \[some-supply-chain\]"`))
				})

				It("does not track the template", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(dependencyTracker.TrackCallCount()).To(Equal(1))
					key, obj := dependencyTracker.TrackArgsForCall(0)
					Expect(key.String()).To(Equal("ServiceAccount/my-namespace/workload-service-account-name"))
					Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))
				})

				Context("there are no matching options", func() {
					It("logs the handled error message", func() {
						templateOptionsMatchErr.OptionNames = []string{}
						rlzr.RealizeReturns(nil, templateOptionsMatchErr)

						_, _ = reconciler.Reconcile(ctx, req)

						Expect(out).To(Say(`"level":"info"`))
						Expect(out).To(Say(`"msg":"handled error reconciling workload"`))
						Expect(out).To(Say(`"handled error":"expected exactly 1 option to match, found \[0\] matching options for resource \[some-resource\] in supply chain \[some-supply-chain\]"`))
					})

					It("does not track the template", func() {
						_, _ = reconciler.Reconcile(ctx, req)
						Expect(dependencyTracker.TrackCallCount()).To(Equal(1))
						key, obj := dependencyTracker.TrackArgsForCall(0)
						Expect(key.String()).To(Equal("ServiceAccount/my-namespace/workload-service-account-name"))
						Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))
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
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(workload.UnknownResourceErrorCondition(realizerError)))
				})

				It("returns an unhandled error and requeues", func() {
					_, err := reconciler.Reconcile(ctx, req)

					Expect(err.Error()).To(ContainSubstring("some error"))
				})

				It("does track the template", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(dependencyTracker.TrackCallCount()).To(Equal(3))
					key, obj := dependencyTracker.TrackArgsForCall(0)
					Expect(key.String()).To(Equal("ServiceAccount/my-namespace/workload-service-account-name"))
					Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))
					key, obj = dependencyTracker.TrackArgsForCall(1)
					Expect(key.String()).To(Equal("my-image-kind.carto.run//my-image-template"))
					Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))
					key, obj = dependencyTracker.TrackArgsForCall(2)
					Expect(key.String()).To(Equal("my-config-kind.carto.run//my-config-template"))
					Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))
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
				Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(workload.ServiceAccountSecretNotFoundCondition(repoError)))
			})

			It("handles the error and logs it", func() {
				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				Expect(out).To(Say(`"level":"info"`))
				Expect(out).To(Say(`"handled error":"failed to get service account secret \[my-namespace/workload-service-account-name\]: some error"`))
			})
		})

		Context("but the resource realizer builder fails", func() {
			BeforeEach(func() {
				resourceRealizerBuilderError = errors.New("some error")
			})

			It("calls the condition manager to add a resource realizer builder error condition", func() {
				_, _ = reconciler.Reconcile(ctx, req)
				Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(workload.ResourceRealizerBuilderErrorCondition(resourceRealizerBuilderError)))
			})

			It("returns an unhandled error", func() {
				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("build resource realizer: some error"))
			})
		})
	})

	Context("but the workload has no label to match with the supply chain", func() {
		BeforeEach(func() {
			wl.Labels = nil
		})
		It("calls the condition manager to report the bad state", func() {
			_, _ = reconciler.Reconcile(ctx, req)
			Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(workload.WorkloadMissingLabelsCondition()))
		})

		It("does not return an error", func() {
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
		})

		It("logs the handled error message", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(out).To(Say(`"level":"info"`))
			Expect(out).To(Say(`"msg":"handled error reconciling workload"`))
			Expect(out).To(Say(`"handled error":"workload \[my-namespace/my-workload-name\] is missing required labels"`))
		})
	})

	Context("and repo returns an empty list of supply chains", func() {
		It("calls the condition manager to add a supply chain not found condition", func() {
			_, _ = reconciler.Reconcile(ctx, req)
			Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(workload.SupplyChainNotFoundCondition(workloadLabels)))
		})

		It("does not return an error", func() {
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
		})

		It("logs the handled error message", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(out).To(Say(`"level":"info"`))
			Expect(out).To(Say(`"msg":"handled error reconciling workload"`))
			Expect(out).To(Say(`"handled error":"no supply chain \[my-namespace/my-workload-name\] found where full selector is satisfied by labels: map\[some-key:some-val\]"`))
		})
	})

	Context("and repo returns an an error when requesting supply chains", func() {
		BeforeEach(func() {
			repo.GetSupplyChainsForWorkloadReturns(nil, errors.New("some error"))
		})

		It("returns an unhandled error and requeues", func() {
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err.Error()).To(ContainSubstring("failed to get supply chains for workload [my-namespace/my-workload-name]: some error"))
		})
	})

	Context("and the repo returns multiple supply chains", func() {
		BeforeEach(func() {
			supplyChain := v1alpha1.ClusterSupplyChain{
				ObjectMeta: metav1.ObjectMeta{Name: "my-supply-chain"},
			}
			repo.GetSupplyChainsForWorkloadReturns([]*v1alpha1.ClusterSupplyChain{&supplyChain, &supplyChain}, nil)
		})

		It("calls the condition manager to report too mane supply chains matched", func() {
			_, _ = reconciler.Reconcile(ctx, req)
			Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(workload.TooManySupplyChainMatchesCondition()))
		})

		It("does not return an error", func() {
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
		})

		It("logs the handled error message", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(out).To(Say(`"level":"info"`))
			Expect(out).To(Say(`"msg":"handled error reconciling workload"`))
			Expect(out).To(Say(`"handled error":"more than one supply chain selected for workload \[my-namespace/my-workload-name\]: \[my-supply-chain my-supply-chain\]"`))
		})
	})

	Context("but status update fails", func() {
		BeforeEach(func() {
			repo.StatusUpdateReturns(errors.New("some error"))
		})

		It("returns an unhandled error and requeues", func() {
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).To(MatchError(ContainSubstring("failed to update status for workload: ")))
		})
	})

	Context("getting the workload returns an error", func() {
		BeforeEach(func() {
			repositoryError := errors.New("RepositoryError")
			repo.GetWorkloadReturns(nil, repositoryError)
		})
		It("returns the error from the repository", func() {
			_, err := reconciler.Reconcile(ctx, req)

			Expect(err).To(MatchError(ContainSubstring("RepositoryError")))
		})
	})

	Context("workload is deleted", func() {
		BeforeEach(func() {
			repo.GetWorkloadReturns(nil, nil)
		})
		It("finishes the reconcile and does not requeue", func() {
			_, err := reconciler.Reconcile(ctx, req)

			Expect(err).NotTo(HaveOccurred())
		})
		It("clears the previously tracked objects for the workload", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(dependencyTracker.ClearTrackedCallCount()).To(Equal(1))
			obj := dependencyTracker.ClearTrackedArgsForCall(0)
			Expect(obj.Name).To(Equal("my-workload-name"))
			Expect(obj.Namespace).To(Equal("my-namespace"))
		})
	})

	Describe("cleaning up orphaned objects", func() {
		BeforeEach(func() {
			supplyChain := v1alpha1.ClusterSupplyChain{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-supply-chain",
				},
				Status: v1alpha1.SupplyChainStatus{
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
			repo.GetSupplyChainsForWorkloadReturns([]*v1alpha1.ClusterSupplyChain{&supplyChain}, nil)

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
				wl.Status.Resources = []v1alpha1.RealizedResource{
					{
						Name: "some-resource",
						StampedRef: &corev1.ObjectReference{
							APIVersion: "some-api-version",
							Kind:       "some-kind",
							Name:       "some-new-stamped-obj-name",
						},
					},
				}
				repo.GetWorkloadReturns(wl, nil)
			})

			It("does not attempt to delete any objects", func() {
				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				Expect(repo.DeleteCallCount()).To(Equal(0))
			})
		})

		Context("a template changes so there are orphaned objects", func() {
			BeforeEach(func() {
				wl.Status.Resources = []v1alpha1.RealizedResource{
					{
						Name: "some-resource",
						StampedRef: &corev1.ObjectReference{
							APIVersion: "some-api-version",
							Kind:       "some-kind",
							Name:       "some-old-stamped-obj-name",
						},
					},
				}
				repo.GetWorkloadReturns(wl, nil)
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

					Expect(out).To(Say(`"msg":"failed to cleanup orphaned objects","workload":"my-namespace/my-workload-name"`))
				})
			})
		})
	})
})
