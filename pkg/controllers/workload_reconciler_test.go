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
	"github.com/vmware-tanzu/cartographer/pkg/controllers/controllersfakes"
	cerrors "github.com/vmware-tanzu/cartographer/pkg/errors"
	"github.com/vmware-tanzu/cartographer/pkg/realizer"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/realizerfakes"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/statuses"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
	"github.com/vmware-tanzu/cartographer/pkg/satoken/satokenfakes"
	"github.com/vmware-tanzu/cartographer/pkg/stamp"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/dependency/dependencyfakes"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/stamped/stampedfakes"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

var _ = Describe("WorkloadReconciler", func() {
	var (
		out                             *Buffer
		reconciler                      controllers.WorkloadReconciler
		ctx                             context.Context
		req                             ctrl.Request
		repo                            *repositoryfakes.FakeRepository
		tokenManager                    *satokenfakes.FakeTokenManager
		conditionManager                *conditionsfakes.FakeConditionManager
		rlzr                            *controllersfakes.FakeRealizer
		wl                              *v1alpha1.Workload
		workloadLabels                  map[string]string
		stampedTracker                  *stampedfakes.FakeStampedTracker
		dependencyTracker               *dependencyfakes.FakeDependencyTracker
		builtResourceRealizer           *realizerfakes.FakeResourceRealizer
		labelerForBuiltResourceRealizer realizer.ResourceLabeler
		resourceRealizerAuthToken       string
		workloadServiceAccount          *corev1.ServiceAccount
		workloadServiceAccountName      = "workload-service-account-name"
		workloadServiceAccountToken     = "workload-sa-token"
		resourceRealizerBuilderError    error
	)

	BeforeEach(func() {
		out = NewBuffer()
		logger := zap.New(zap.WriteTo(out))
		ctx = logr.NewContext(context.Background(), logger)

		conditionManager = &conditionsfakes.FakeConditionManager{}

		fakeConditionManagerBuilder := func(string, []metav1.Condition) conditions.ConditionManager {
			return conditionManager
		}

		rlzr = &controllersfakes.FakeRealizer{}
		rlzr.RealizeReturns(nil)

		stampedTracker = &stampedfakes.FakeStampedTracker{}
		dependencyTracker = &dependencyfakes.FakeDependencyTracker{}

		repo = &repositoryfakes.FakeRepository{}
		scheme := runtime.NewScheme()
		err := utils.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())
		repo.GetSchemeReturns(scheme)
		fakeRESTMapper := controllersfakes.FakeRESTMapper{}
		repo.GetRESTMapperReturns(&fakeRESTMapper)

		tokenManager = &satokenfakes.FakeTokenManager{}

		workloadServiceAccountName = "workload-service-account-name"

		workloadServiceAccount = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Name: workloadServiceAccountName},
		}

		repo.GetServiceAccountReturns(workloadServiceAccount, nil)
		tokenManager.GetServiceAccountTokenReturns(workloadServiceAccountToken, nil)

		resourceRealizerBuilderError = nil

		resourceRealizerBuilder := func(authToken string, owner client.Object, templatingContext realizer.ContextGenerator, systemRepo repository.Repository, resourceLabeler realizer.ResourceLabeler) (realizer.ResourceRealizer, error) {
			labelerForBuiltResourceRealizer = resourceLabeler
			if resourceRealizerBuilderError != nil {
				return nil, resourceRealizerBuilderError
			}
			resourceRealizerAuthToken = authToken
			builtResourceRealizer = &realizerfakes.FakeResourceRealizer{}
			return builtResourceRealizer, nil
		}

		reconciler = controllers.WorkloadReconciler{
			Repo:                    repo,
			TokenManager:            tokenManager,
			ConditionManagerBuilder: fakeConditionManagerBuilder,
			ResourceRealizerBuilder: resourceRealizerBuilder,
			Realizer:                rlzr,
			StampedTracker:          stampedTracker,
			DependencyTracker:       dependencyTracker,
			RESTMapper:              &fakeRESTMapper,
			Scheme:                  scheme,
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
				ServiceAccountName: workloadServiceAccountName,
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
			supplyChainName  string
			supplyChain      v1alpha1.ClusterSupplyChain
			resourceStatuses statuses.ResourceStatuses
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

			resourceStatuses = statuses.NewResourceStatuses(nil, conditions.AddConditionForResourceSubmittedWorkload)
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
				}, nil, false,
			)
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
				}, nil, false,
			)

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

			lifecycleReader := lifecycleReader{lifecycle: templates.Immutable}

			labels := labelerForBuiltResourceRealizer(resource, &lifecycleReader)
			Expect(labels).To(Equal(templates.Labels{
				"carto.run/workload-name":         "my-workload-name",
				"carto.run/workload-namespace":    "my-namespace",
				"carto.run/supply-chain-name":     "some-supply-chain",
				"carto.run/resource-name":         resource.Name,
				"carto.run/template-kind":         resource.TemplateRef.Kind,
				"carto.run/cluster-template-name": resource.TemplateRef.Name,
				"carto.run/template-lifecycle":    "immutable",
			}))
		})

		It("updates the status of the workload with the realizedResources", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(repo.StatusUpdateCallCount()).To(Equal(1))
			_, wk := repo.StatusUpdateArgsForCall(0)
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

			var workloadResource1Status v1alpha1.ResourceStatus
			var workloadResource2Status v1alpha1.ResourceStatus

			for i := range wk.(*v1alpha1.Workload).Status.Resources {
				switch wk.(*v1alpha1.Workload).Status.Resources[i].Name {
				case "resource1":
					workloadResource1Status = wk.(*v1alpha1.Workload).Status.Resources[i]
				case "resource2":
					workloadResource2Status = wk.(*v1alpha1.Workload).Status.Resources[i]
				}
			}

			Expect(workloadResource1Status.RealizedResource).To(Equal(resource1Status.RealizedResource))
			Expect(workloadResource2Status.RealizedResource).To(Equal(resource2Status.RealizedResource))
		})

		It("dynamically creates a resource realizer", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(rlzr.RealizeCallCount()).To(Equal(1))
			_, resourceRealizer, _, _, _ := rlzr.RealizeArgsForCall(0)
			Expect(resourceRealizer).To(Equal(builtResourceRealizer))
		})

		It("uses the service account specified by the workload for realizing resources", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(repo.GetServiceAccountCallCount()).To(Equal(1))
			_, serviceAccountNameArg, serviceAccountNS := repo.GetServiceAccountArgsForCall(0)
			Expect(serviceAccountNameArg).To(Equal(workloadServiceAccountName))
			Expect(serviceAccountNS).To(Equal("my-namespace"))

			Expect(tokenManager.GetServiceAccountTokenArgsForCall(0)).To(Equal(workloadServiceAccount))
			Expect(resourceRealizerAuthToken).To(Equal(workloadServiceAccountToken))
		})

		Context("the workload does not specify a service account", func() {
			BeforeEach(func() {
				wl.Spec.ServiceAccountName = ""
			})

			Context("the supply chain provides a service account", func() {
				supplyChainServiceAccountName := "some-supply-chain-service-account"
				supplyChainServiceAccountToken := "supply-chain-service-account-token"
				supplyChainServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{Name: supplyChainServiceAccountName},
				}

				BeforeEach(func() {
					supplyChain.Spec.ServiceAccountRef.Name = "some-supply-chain-service-account"

					repo.GetServiceAccountReturns(supplyChainServiceAccount, nil)
					tokenManager.GetServiceAccountTokenReturns(supplyChainServiceAccountToken, nil)
				})

				It("uses the supply chain service account in the workload's namespace", func() {
					_, _ = reconciler.Reconcile(ctx, req)

					Expect(repo.GetServiceAccountCallCount()).To(Equal(1))
					_, serviceAccountNameArg, serviceAccountNS := repo.GetServiceAccountArgsForCall(0)
					Expect(serviceAccountNameArg).To(Equal(supplyChainServiceAccountName))
					Expect(serviceAccountNS).To(Equal("my-namespace"))

					Expect(tokenManager.GetServiceAccountTokenArgsForCall(0)).To(Equal(supplyChainServiceAccount))
					Expect(resourceRealizerAuthToken).To(Equal(supplyChainServiceAccountToken))
				})

				Context("the supply chain specifies a namespace", func() {
					BeforeEach(func() {
						supplyChain.Spec.ServiceAccountRef.Namespace = "some-supply-chain-namespace"
					})

					It("uses the supply chain service account in the specified namespace", func() {
						_, _ = reconciler.Reconcile(ctx, req)

						Expect(repo.GetServiceAccountCallCount()).To(Equal(1))
						_, serviceAccountNameArg, serviceAccountNS := repo.GetServiceAccountArgsForCall(0)
						Expect(serviceAccountNameArg).To(Equal(supplyChainServiceAccountName))
						Expect(serviceAccountNS).To(Equal("some-supply-chain-namespace"))

						Expect(tokenManager.GetServiceAccountTokenArgsForCall(0)).To(Equal(supplyChainServiceAccount))
						Expect(resourceRealizerAuthToken).To(Equal(supplyChainServiceAccountToken))
					})
				})
			})

			Context("the supply chain does not provide a service account", func() {
				defaultServiceAccountToken := "default-service-account-token"
				defaultServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{Name: "default"},
				}

				BeforeEach(func() {
					repo.GetServiceAccountReturns(defaultServiceAccount, nil)
					tokenManager.GetServiceAccountTokenReturns(defaultServiceAccountToken, nil)
				})

				It("defaults to the default service account in the workloads namespace", func() {
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

		It("sets the SupplyChainRef", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(wl.Status.SupplyChainRef.Kind).To(Equal("ClusterSupplyChain"))
			Expect(wl.Status.SupplyChainRef.Name).To(Equal(supplyChainName))
		})

		It("calls the condition manager to specify the supply chain is ready", func() {
			_, _ = reconciler.Reconcile(ctx, req)
			Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(conditions.SupplyChainReadyCondition()))
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
			Expect(hndl).To(Equal(handler.EnqueueRequestForOwner(repo.GetScheme(), repo.GetRESTMapper(), &v1alpha1.Workload{})))

			_, obj, hndl, _ = stampedTracker.WatchArgsForCall(1)
			gvks = append(gvks, obj.GetObjectKind().GroupVersionKind())
			Expect(hndl).To(Equal(handler.EnqueueRequestForOwner(repo.GetScheme(), repo.GetRESTMapper(), &v1alpha1.Workload{})))

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
				Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(conditions.MissingReadyInSupplyChainCondition(expectedCondition)))
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
					templateError = cerrors.GetTemplateError{
						Err:          errors.New("some error"),
						ResourceName: "some-name",
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
					Expect(key.String()).To(Equal("ServiceAccount/my-namespace/workload-service-account-name"))
					Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))
				})
			})

			Context("of type StampError", func() {
				var stampError cerrors.StampError
				BeforeEach(func() {
					stampError = cerrors.StampError{
						Err:           errors.New("some error"),
						ResourceName:  "some-name",
						BlueprintName: supplyChainName,
						BlueprintType: cerrors.SupplyChain,
						TemplateName:  "cool-template",
						TemplateKind:  "FrozenTemplate",
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
						}, nil, false,
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
						}, nil, false,
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
					Expect(out).To(Say(`"msg":"handled error reconciling workload"`))
					Expect(out).To(Say(`"handled error":"unable to stamp object for resource \[some-name\] for template \[FrozenTemplate/cool-template\] in supply chain \[some-supply-chain\]: some error"`))
				})

				It("does track the template", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(dependencyTracker.TrackCallCount()).To(Equal(3))
					key, obj := dependencyTracker.TrackArgsForCall(0)
					Expect(key.String()).To(Equal("ServiceAccount/my-namespace/workload-service-account-name"))
					Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))

					var keys []string
					var objs []string
					key, obj = dependencyTracker.TrackArgsForCall(1)
					keys = append(keys, key.String())
					objs = append(objs, obj.String())

					key, obj = dependencyTracker.TrackArgsForCall(2)
					keys = append(keys, key.String())
					objs = append(objs, obj.String())

					Expect(keys).To(ContainElements("my-image-kind.carto.run//my-image-template",
						"my-config-kind.carto.run//my-config-template"))
					Expect(objs).To(ContainElements("my-namespace/my-workload-name",
						"my-namespace/my-workload-name"))
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
					Expect(key.String()).To(Equal("ServiceAccount/my-namespace/workload-service-account-name"))
					Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))

					var keys []string
					var objs []string
					key, obj = dependencyTracker.TrackArgsForCall(1)
					keys = append(keys, key.String())
					objs = append(objs, obj.String())

					key, obj = dependencyTracker.TrackArgsForCall(2)
					keys = append(keys, key.String())
					objs = append(objs, obj.String())

					Expect(keys).To(ContainElements("my-image-kind.carto.run//my-image-template",
						"my-config-kind.carto.run//my-config-template"))
					Expect(objs).To(ContainElements("my-namespace/my-workload-name",
						"my-namespace/my-workload-name"))
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
						BlueprintName: supplyChainName,
						BlueprintType: cerrors.SupplyChain,
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
					Expect(out).To(Say(`"handled error":"unable to apply object \[a-namespace/a-name\] for resource \[some-name\] in supply chain \[some-supply-chain\]: fantastic error"`))
				})

				It("does track the template", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(dependencyTracker.TrackCallCount()).To(Equal(3))
					key, obj := dependencyTracker.TrackArgsForCall(0)
					Expect(key.String()).To(Equal("ServiceAccount/my-namespace/workload-service-account-name"))
					Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))

					var keys []string
					var objs []string
					key, obj = dependencyTracker.TrackArgsForCall(1)
					keys = append(keys, key.String())
					objs = append(objs, obj.String())

					key, obj = dependencyTracker.TrackArgsForCall(2)
					keys = append(keys, key.String())
					objs = append(objs, obj.String())

					Expect(keys).To(ContainElements("my-image-kind.carto.run//my-image-template",
						"my-config-kind.carto.run//my-config-template"))
					Expect(objs).To(ContainElements("my-namespace/my-workload-name",
						"my-namespace/my-workload-name"))
				})
			})

			Context("of type RetrieveOutputError with jsonpath", func() {
				var retrieveError cerrors.RetrieveOutputError
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
					jsonPathError := stamp.NewJsonPathError("this.wont.find.anything", errors.New("some error"))
					retrieveError = cerrors.RetrieveOutputError{
						Err:               jsonPathError,
						ResourceName:      "some-resource",
						StampedObject:     stampedObject,
						BlueprintName:     supplyChainName,
						BlueprintType:     cerrors.SupplyChain,
						QualifiedResource: "mything.thing.io",
					}
					rlzr.RealizeStub = func(ctx context.Context, resourceRealizer realizer.ResourceRealizer, deliveryName string, resources []realizer.OwnerResource, statuses statuses.ResourceStatuses) error {
						statusesVal := reflect.ValueOf(statuses)
						existingVal := reflect.ValueOf(resourceStatuses)

						reflect.Indirect(statusesVal).Set(reflect.Indirect(existingVal))
						return retrieveError
					}
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					var emptyConditionStatus metav1.ConditionStatus
					Expect(conditionManager.AddPositiveArgsForCall(1)).
						To(Equal(conditions.MissingValueAtPathCondition(true, stampedObject, "this.wont.find.anything", "mything.thing.io", emptyConditionStatus)))
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

					var keys []string
					var objs []string
					key, obj = dependencyTracker.TrackArgsForCall(1)
					keys = append(keys, key.String())
					objs = append(objs, obj.String())

					key, obj = dependencyTracker.TrackArgsForCall(2)
					keys = append(keys, key.String())
					objs = append(objs, obj.String())

					Expect(keys).To(ContainElements("my-image-kind.carto.run//my-image-template",
						"my-config-kind.carto.run//my-config-template"))
					Expect(objs).To(ContainElements("my-namespace/my-workload-name",
						"my-namespace/my-workload-name"))
				})
			})
			Context("of type RetrieveOutputError without stampedobject", func() {
				var retrieveError cerrors.RetrieveOutputError
				BeforeEach(func() {
					jsonPathError := stamp.NewJsonPathError("this.wont.find.anything", errors.New("some error"))
					retrieveError = cerrors.RetrieveOutputError{
						Err:               jsonPathError,
						ResourceName:      "some-resource",
						StampedObject:     nil,
						BlueprintName:     supplyChainName,
						BlueprintType:     cerrors.SupplyChain,
						QualifiedResource: "input",
						PassThroughInput:  "configs",
					}
					rlzr.RealizeStub = func(ctx context.Context, resourceRealizer realizer.ResourceRealizer, deliveryName string, resources []realizer.OwnerResource, statuses statuses.ResourceStatuses) error {
						statusesVal := reflect.ValueOf(statuses)
						existingVal := reflect.ValueOf(resourceStatuses)

						reflect.Indirect(statusesVal).Set(reflect.Indirect(existingVal))
						return retrieveError
					}
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(
						Equal(conditions.MissingPassThroughInputCondition("configs", "input")))
				})

				It("does not return an error", func() {
					_, err := reconciler.Reconcile(ctx, req)
					Expect(err).NotTo(HaveOccurred())
				})

				It("logs the handled error message", func() {
					_, _ = reconciler.Reconcile(ctx, req)

					Expect(out).To(Say(`"level":"info"`))
					Expect(out).To(Say(`"msg":"handled error reconciling workload"`))
					Expect(out).To(Say(`"handled error":"unable to retrieve outputs from pass through \[configs\] for resource \[some-resource\] in supply chain \[some-supply-chain\]: failed to evaluate json path 'this.wont.find.anything': some error"`))
				})

				It("does track the template", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(dependencyTracker.TrackCallCount()).To(Equal(3))
					key, obj := dependencyTracker.TrackArgsForCall(0)
					Expect(key.String()).To(Equal("ServiceAccount/my-namespace/workload-service-account-name"))
					Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))

					var keys []string
					var objs []string
					key, obj = dependencyTracker.TrackArgsForCall(1)
					keys = append(keys, key.String())
					objs = append(objs, obj.String())

					key, obj = dependencyTracker.TrackArgsForCall(2)
					keys = append(keys, key.String())
					objs = append(objs, obj.String())

					Expect(keys).To(ContainElements("my-image-kind.carto.run//my-image-template",
						"my-config-kind.carto.run//my-config-template"))
					Expect(objs).To(ContainElements("my-namespace/my-workload-name",
						"my-namespace/my-workload-name"))
				})
			})

			Context("of type ResolveTemplateOptionError", func() {
				var resolveOptionErr cerrors.ResolveTemplateOptionError
				BeforeEach(func() {
					jsonPathError := stamp.NewJsonPathError("this.wont.find.anything", errors.New("some error"))
					resolveOptionErr = cerrors.ResolveTemplateOptionError{
						Err:           jsonPathError,
						BlueprintName: supplyChainName,
						BlueprintType: cerrors.SupplyChain,
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
				var templateOptionsMatchErr cerrors.TemplateOptionsMatchError
				BeforeEach(func() {
					templateOptionsMatchErr = cerrors.TemplateOptionsMatchError{
						BlueprintName: supplyChainName,
						BlueprintType: cerrors.SupplyChain,
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
						rlzr.RealizeReturns(templateOptionsMatchErr)

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

			Context("of type ListCreatedObjectsError", func() {
				var (
					listCreatedObjectsError cerrors.ListCreatedObjectsError
					err                     error
				)

				BeforeEach(func() {
					listCreatedObjectsError = cerrors.ListCreatedObjectsError{
						Err:       fmt.Errorf("some-error"),
						Namespace: "some-namespace",
						Labels:    map[string]string{"label?": "label!"},
					}
					rlzr.RealizeReturns(listCreatedObjectsError)
					_, err = reconciler.Reconcile(ctx, req)
				})

				It("calls the condition manager to report", func() {
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(
						Equal(conditions.BlueprintsFailedToListCreatedObjectsCondition(true, listCreatedObjectsError)))
				})

				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
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
					Expect(key.String()).To(Equal("ServiceAccount/my-namespace/workload-service-account-name"))
					Expect(obj.String()).To(Equal("my-namespace/my-workload-name"))

					var keys []string
					var objs []string
					key, obj = dependencyTracker.TrackArgsForCall(1)
					keys = append(keys, key.String())
					objs = append(objs, obj.String())

					key, obj = dependencyTracker.TrackArgsForCall(2)
					keys = append(keys, key.String())
					objs = append(objs, obj.String())

					Expect(keys).To(ContainElements("my-image-kind.carto.run//my-image-template",
						"my-config-kind.carto.run//my-config-template"))
					Expect(objs).To(ContainElements("my-namespace/my-workload-name",
						"my-namespace/my-workload-name"))
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
				Expect(out).To(Say(`"handled error":"failed to get service account \[my-namespace/workload-service-account-name\]: some error"`))
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
				Expect(out).To(Say(`"handled error":"failed to get token for service account \[my-namespace/workload-service-account-name\]: some error"`))
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
			Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(conditions.WorkloadMissingLabelsCondition()))
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
			Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(conditions.SupplyChainNotFoundCondition(workloadLabels)))
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
			Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(conditions.TooManySupplyChainMatchesCondition()))
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

			rlzr.RealizeReturns(nil)

			resourceStatuses := statuses.NewResourceStatuses(nil, conditions.AddConditionForResourceSubmittedWorkload)
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
					TemplateRef: &corev1.ObjectReference{
						Kind: "some-template-kind",
						Name: "some-template-name",
					},
				}, nil, false,
			)
			rlzr.RealizeStub = func(ctx context.Context, resourceRealizer realizer.ResourceRealizer, deliveryName string, resources []realizer.OwnerResource, statuses statuses.ResourceStatuses) error {
				statusesVal := reflect.ValueOf(statuses)
				existingVal := reflect.ValueOf(resourceStatuses)

				reflect.Indirect(statusesVal).Set(reflect.Indirect(existingVal))
				return nil
			}
		})
		Context("when current resource stamped from template with mutable lifecycle", func() {
			BeforeEach(func() {
				someTemplate := v1alpha1.ClusterTemplate{}
				repo.GetTemplateReturns(&someTemplate, nil)
			})
			Context("template does not change so there are no orphaned objects", func() {
				BeforeEach(func() {
					wl.Status.Resources = []v1alpha1.ResourceStatus{
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
								TemplateRef: &corev1.ObjectReference{
									Name: "some-template-name",
									Kind: "some-template-kind",
								},
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
					wl.Status.Resources = []v1alpha1.ResourceStatus{
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
								TemplateRef: &corev1.ObjectReference{
									Name: "some-template-name",
									Kind: "some-template-kind",
								},
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

						Expect(out).To(Say(`"msg":"failed to cleanup orphaned objects","workload":{"name":"my-workload-name","namespace":"my-namespace"}`))
					})
				})
			})
		})

		Context("when current resource stamped from immutable template", func() {
			BeforeEach(func() {
				someTemplate := v1alpha1.ClusterTemplate{Spec: v1alpha1.TemplateSpec{Lifecycle: "immutable"}}
				repo.GetTemplateReturns(&someTemplate, nil)
			})

			Context("when previous resource stamped with immutable template", func() {
				BeforeEach(func() {
					unstructuredWithImmutableLabel := unstructured.Unstructured{}
					unstructuredWithImmutableLabel.SetLabels(map[string]string{"carto.run/template-lifecycle": "immutable"})
					repo.GetUnstructuredReturnsOnCall(0, &unstructuredWithImmutableLabel, nil)
				})
				Context("and the previous resource and a newly created resource are from the same step and template", func() {
					BeforeEach(func() {
						wl.Status.Resources = []v1alpha1.ResourceStatus{
							{
								RealizedResource: v1alpha1.RealizedResource{
									Name: "some-resource",
									StampedRef: &v1alpha1.StampedRef{
										ObjectReference: &corev1.ObjectReference{
											APIVersion: "some-api-version",
											Kind:       "some-kind",
											Name:       "some-obj-name",
										},
										Resource: "some-kind",
									},
									TemplateRef: &corev1.ObjectReference{
										Name: "some-template-name",
										Kind: "some-template-kind",
									},
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

				Context("and the previous resource shares a template with a newly created resource but shares no step", func() {
					BeforeEach(func() {
						wl.Status.Resources = []v1alpha1.ResourceStatus{
							{
								RealizedResource: v1alpha1.RealizedResource{
									Name: "some-other-resource",
									StampedRef: &v1alpha1.StampedRef{
										ObjectReference: &corev1.ObjectReference{
											APIVersion: "some-api-version",
											Kind:       "some-kind",
											Name:       "some-obj-name",
										},
										Resource: "some-kind",
									},
									TemplateRef: &corev1.ObjectReference{
										Name: "some-template-name",
										Kind: "some-template-kind",
									},
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
						Expect(obj.GetName()).To(Equal("some-obj-name"))
						Expect(obj.GetKind()).To(Equal("some-kind"))
					})
				})

				Context("and the previous resource shares a step with a newly created resource but is from a different template", func() {
					BeforeEach(func() {
						wl.Status.Resources = []v1alpha1.ResourceStatus{
							{
								RealizedResource: v1alpha1.RealizedResource{
									Name: "some-resource",
									StampedRef: &v1alpha1.StampedRef{
										ObjectReference: &corev1.ObjectReference{
											APIVersion: "some-api-version",
											Kind:       "some-kind",
											Name:       "some-obj-name",
										},
										Resource: "some-kind",
									},
									TemplateRef: &corev1.ObjectReference{
										Name: "some-other-template-name",
										Kind: "some-template-kind",
									},
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
						Expect(obj.GetName()).To(Equal("some-obj-name"))
						Expect(obj.GetKind()).To(Equal("some-kind"))
					})
				})
			})

			Context("when previous resource stamped from a mutable template", func() {
				BeforeEach(func() {
					unstructuredWithMutableLabel := unstructured.Unstructured{}
					unstructuredWithMutableLabel.SetLabels(map[string]string{"carto.run/template-lifecycle": "mutable"})
					repo.GetUnstructuredReturnsOnCall(0, &unstructuredWithMutableLabel, nil)
				})

				Context("and the previous resource and a newly created resource are from the same step and template", func() {
					BeforeEach(func() {
						wl.Status.Resources = []v1alpha1.ResourceStatus{
							{
								RealizedResource: v1alpha1.RealizedResource{
									Name: "some-resource",
									StampedRef: &v1alpha1.StampedRef{
										ObjectReference: &corev1.ObjectReference{
											APIVersion: "some-api-version",
											Kind:       "some-kind",
											Name:       "some-obj-name",
										},
										Resource: "some-kind",
									},
									TemplateRef: &corev1.ObjectReference{
										Name: "some-template-name",
										Kind: "some-template-kind",
									},
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
						Expect(obj.GetName()).To(Equal("some-obj-name"))
						Expect(obj.GetKind()).To(Equal("some-kind"))
					})
				})

				Context("and the previous resource shares a template with a newly created resource but shares no step", func() {
					BeforeEach(func() {
						wl.Status.Resources = []v1alpha1.ResourceStatus{
							{
								RealizedResource: v1alpha1.RealizedResource{
									Name: "some-other-resource",
									StampedRef: &v1alpha1.StampedRef{
										ObjectReference: &corev1.ObjectReference{
											APIVersion: "some-api-version",
											Kind:       "some-kind",
											Name:       "some-obj-name",
										},
										Resource: "some-kind",
									},
									TemplateRef: &corev1.ObjectReference{
										Name: "some-template-name",
										Kind: "some-template-kind",
									},
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
						Expect(obj.GetName()).To(Equal("some-obj-name"))
						Expect(obj.GetKind()).To(Equal("some-kind"))
					})
				})

				Context("and the previous resource shares a step with a newly created resource but is from a different template", func() {
					BeforeEach(func() {
						wl.Status.Resources = []v1alpha1.ResourceStatus{
							{
								RealizedResource: v1alpha1.RealizedResource{
									Name: "some-resource",
									StampedRef: &v1alpha1.StampedRef{
										ObjectReference: &corev1.ObjectReference{
											APIVersion: "some-api-version",
											Kind:       "some-kind",
											Name:       "some-obj-name",
										},
										Resource: "some-kind",
									},
									TemplateRef: &corev1.ObjectReference{
										Name: "some-other-template-name",
										Kind: "some-template-kind",
									},
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
						Expect(obj.GetName()).To(Equal("some-obj-name"))
						Expect(obj.GetKind()).To(Equal("some-kind"))
					})
				})
			})

			Context("when the previously stamped object no longer exists", func() {
				BeforeEach(func() {
					referenceToNonExistentObject := v1alpha1.RealizedResource{
						Name: "some-resource",
						StampedRef: &v1alpha1.StampedRef{
							ObjectReference: &corev1.ObjectReference{
								APIVersion: "some-api-version",
								Kind:       "some-kind",
								Name:       "some-obj-name",
							},
							Resource: "some-kind",
						},
						TemplateRef: &corev1.ObjectReference{
							Name: "some-template-name",
							Kind: "some-template-kind",
						},
					}

					wl.Status.Resources = []v1alpha1.ResourceStatus{
						{
							RealizedResource: referenceToNonExistentObject,
						},
					}
					repo.GetWorkloadReturns(wl, nil)

					repo.GetUnstructuredReturnsOnCall(0, nil, nil)
					repo.DeleteReturns(fmt.Errorf("some error"))
				})

				It("does not return an error", func() {
					_, err := reconciler.Reconcile(ctx, req)
					Expect(err).NotTo(HaveOccurred())

					Expect(repo.DeleteCallCount()).To(Equal(1))
					Expect(out).To(Say(`"msg":"failed to cleanup orphaned objects","workload":{"name":"my-workload-name","namespace":"my-namespace"}`))
				})
			})
		})

		Context("when previous resource was stamped from a template that is no longer on cluster", func() {
			BeforeEach(func() {
				repo.GetTemplateReturns(nil, kerrors.NewNotFound(schema.GroupResource{}, "somename"))

				wl.Status.Resources = []v1alpha1.ResourceStatus{
					{
						RealizedResource: v1alpha1.RealizedResource{
							Name: "some-resource",
							StampedRef: &v1alpha1.StampedRef{
								ObjectReference: &corev1.ObjectReference{
									APIVersion: "some-api-version",
									Kind:       "some-kind",
									Name:       "some-obj-name",
								},
								Resource: "some-kind",
							},
							TemplateRef: &corev1.ObjectReference{
								Name: "some-template-name",
								Kind: "some-template-kind",
							},
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
				Expect(obj.GetName()).To(Equal("some-obj-name"))
				Expect(obj.GetKind()).To(Equal("some-kind"))
			})
		})
	})
})
