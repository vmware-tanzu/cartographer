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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
	kerrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/vmware-tanzu/cartographer/pkg/repository"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/conditions/conditionsfakes"
	"github.com/vmware-tanzu/cartographer/pkg/controller/workload"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/workload"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/workload/workloadfakes"
	"github.com/vmware-tanzu/cartographer/pkg/registrar"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/trackerfakes"
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
		dynamicTracker               *trackerfakes.FakeDynamicTracker
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

		dynamicTracker = &trackerfakes.FakeDynamicTracker{}

		repo = &repositoryfakes.FakeRepository{}
		scheme := runtime.NewScheme()
		err := registrar.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())
		repo.GetSchemeReturns(scheme)

		repo.GetServiceAccountSecretReturns(serviceAccountSecret, nil)

		resourceRealizerBuilderError = nil
		resourceRealizerBuilder := func(ctx context.Context, secret *corev1.Secret, workload *v1alpha1.Workload, systemRepo repository.Repository) (realizer.ResourceRealizer, error) {
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
			DynamicTracker:          dynamicTracker,
		}

		req = ctrl.Request{
			NamespacedName: types.NamespacedName{Name: "my-workload-name", Namespace: "my-namespace"},
		}

		workloadLabels = map[string]string{"some-key": "some-val"}

		serviceAccountName = "alternate-service-account-name"

		wl = &v1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Generation: 1,
				Labels:     workloadLabels,
				Name:       "my-workload",
				Namespace:  "my-ns",
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

		_, updatedWorkload := repo.StatusUpdateArgsForCall(0)
		Expect(*updatedWorkload.(*v1alpha1.Workload)).To(MatchFields(IgnoreExtras, Fields{
			"Status": MatchFields(IgnoreExtras, Fields{
				"Conditions": Equal(someConditions),
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
			supplyChainName string
			supplyChain     v1alpha1.ClusterSupplyChain
			stampedObject1  *unstructured.Unstructured
			stampedObject2  *unstructured.Unstructured
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
			repo.GetSupplyChainsForWorkloadReturns([]v1alpha1.ClusterSupplyChain{supplyChain}, nil)
			stampedObject1 = &unstructured.Unstructured{}
			stampedObject1.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "thing.io",
				Version: "alphabeta1",
				Kind:    "MyThing",
			})
			stampedObject2 = &unstructured.Unstructured{}
			stampedObject2.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "hello.io",
				Version: "goodbye",
				Kind:    "NiceToSeeYou",
			})
			rlzr.RealizeReturns([]*unstructured.Unstructured{stampedObject1, stampedObject2}, nil)
		})

		It("uses the service account specified by the workload for realizing resources", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(repo.GetServiceAccountSecretCallCount()).To(Equal(1))
			_, serviceAccountName, serviceAccountNS := repo.GetServiceAccountSecretArgsForCall(0)
			Expect(serviceAccountName).To(Equal(serviceAccountName))
			Expect(serviceAccountNS).To(Equal("my-namespace"))
			Expect(resourceRealizerSecret).To(Equal(serviceAccountSecret))

			Expect(rlzr.RealizeCallCount()).To(Equal(1))
			_, resourceRealizer, _ := rlzr.RealizeArgsForCall(0)
			Expect(resourceRealizer).To(Equal(builtResourceRealizer))
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
			Expect(dynamicTracker.WatchCallCount()).To(Equal(2))
			_, obj, hndl := dynamicTracker.WatchArgsForCall(0)

			Expect(obj).To(Equal(stampedObject1))
			Expect(hndl).To(Equal(&handler.EnqueueRequestForOwner{OwnerType: &v1alpha1.Workload{}}))

			_, obj, hndl = dynamicTracker.WatchArgsForCall(1)

			Expect(obj).To(Equal(stampedObject2))
			Expect(hndl).To(Equal(&handler.EnqueueRequestForOwner{OwnerType: &v1alpha1.Workload{}}))
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
				repo.GetSupplyChainsForWorkloadReturns([]v1alpha1.ClusterSupplyChain{supplyChain}, nil)
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
			Context("of type GetClusterTemplateError", func() {
				var templateError error
				BeforeEach(func() {
					templateError = realizer.GetClusterTemplateError{
						Err: errors.New("some error"),
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
			})

			Context("of type StampError", func() {
				var stampError realizer.StampError
				BeforeEach(func() {
					stampError = realizer.StampError{
						Err:      errors.New("some error"),
						Resource: &v1alpha1.SupplyChainResource{Name: "some-name"},
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
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(workload.TemplateRejectedByAPIServerCondition(stampedObjectError)))
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
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(workload.TemplateRejectedByAPIServerCondition(stampedObjectError)))
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
				BeforeEach(func() {
					jsonPathError := templates.NewJsonPathError("this.wont.find.anything", errors.New("some error"))
					retrieveError = realizer.RetrieveOutputError{
						Err:      jsonPathError,
						Resource: &v1alpha1.SupplyChainResource{Name: "some-resource"},
					}
					rlzr.RealizeReturns(nil, retrieveError)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(workload.MissingValueAtPathCondition("some-resource", "this.wont.find.anything")))
				})

				It("does not return an error", func() {
					_, err := reconciler.Reconcile(ctx, req)
					Expect(err).NotTo(HaveOccurred())
				})

				It("logs the handled error message", func() {
					_, _ = reconciler.Reconcile(ctx, req)

					Expect(out).To(Say(`"level":"info"`))
					Expect(out).To(Say(`"msg":"handled error reconciling workload"`))
					Expect(out).To(Say(`"handled error":"unable to retrieve outputs from stamped object for resource \[some-resource\]: evaluate json path 'this.wont.find.anything': some error"`))
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
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(workload.UnknownResourceErrorCondition(realizerError)))
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
				Expect(out).To(Say(`"handled error":"failed to get service account secret \[alternate-service-account-name\]: some error"`))
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
			Expect(out).To(Say(`"handled error":"workload \[my-ns/my-workload\] is missing required labels"`))
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
			Expect(out).To(Say(`"handled error":"no supply chain \[my-ns/my-workload\] found where full selector is satisfied by labels: map\[some-key:some-val\]"`))
		})
	})

	Context("and repo returns an an error when requesting supply chains", func() {
		BeforeEach(func() {
			repo.GetSupplyChainsForWorkloadReturns(nil, errors.New("some error"))
		})

		It("returns an unhandled error and requeues", func() {
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err.Error()).To(ContainSubstring("failed to get supply chains for workload [my-ns/my-workload]: some error"))
		})
	})

	Context("and the repo returns multiple supply chains", func() {
		BeforeEach(func() {
			supplyChain := v1alpha1.ClusterSupplyChain{
				ObjectMeta: metav1.ObjectMeta{Name: "my-supply-chain"},
			}
			repo.GetSupplyChainsForWorkloadReturns([]v1alpha1.ClusterSupplyChain{supplyChain, supplyChain}, nil)
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
			Expect(out).To(Say(`"handled error":"more than one supply chain selected for workload \[my-ns/my-workload\]: \[my-supply-chain my-supply-chain\]"`))
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
	})
})
