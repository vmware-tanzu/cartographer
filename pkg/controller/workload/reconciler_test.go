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

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	workloadfakes2 "github.com/vmware-tanzu/cartographer/pkg/controller/workload/workloadfakes"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/conditions/conditionsfakes"
	"github.com/vmware-tanzu/cartographer/pkg/controller/workload"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/workload"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/workload/workloadfakes"
	"github.com/vmware-tanzu/cartographer/pkg/registrar"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

var _ = Describe("Reconciler", func() {
	var (
		out              *Buffer
		reconciler       workload.Reconciler
		ctx              context.Context
		req              ctrl.Request
		repo             *repositoryfakes.FakeRepository
		conditionManager *conditionsfakes.FakeConditionManager
		rlzr             *workloadfakes.FakeRealizer
		wl               *v1alpha1.Workload
		workloadLabels   map[string]string
		dynamicTracker   *workloadfakes2.FakeDynamicTracker
	)

	BeforeEach(func() {
		out = NewBuffer()
		logger := zap.New(zap.WriteTo(out))
		ctx = logr.NewContext(context.Background(), logger)

		conditionManager = &conditionsfakes.FakeConditionManager{}

		fakeConditionManagerBuilder := func(string, []metav1.Condition) conditions.ConditionManager {
			return conditionManager
		}

		conditionManager.IsSuccessfulReturns(true)

		rlzr = &workloadfakes.FakeRealizer{}
		rlzr.RealizeReturns(nil, nil)

		dynamicTracker = &workloadfakes2.FakeDynamicTracker{}

		repo = &repositoryfakes.FakeRepository{}
		scheme := runtime.NewScheme()
		err := registrar.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())
		repo.GetSchemeReturns(scheme)

		reconciler = workload.NewReconciler(repo, fakeConditionManagerBuilder, rlzr)
		reconciler.AddTracking(dynamicTracker)

		req = ctrl.Request{
			NamespacedName: types.NamespacedName{Name: "my-workload-name", Namespace: "my-namespace"},
		}

		workloadLabels = map[string]string{"some-key": "some-val"}

		wl = &v1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Generation: 1,
				Labels:     workloadLabels,
			},
		}
		repo.GetWorkloadReturns(wl, nil)
	})

	It("logs that it's begun", func() {
		_, _ = reconciler.Reconcile(ctx, req)

		Expect(out).To(Say(`"msg":"started"`))
		Expect(out).To(Say(`"name":"my-workload-name"`))
		Expect(out).To(Say(`"namespace":"my-namespace"`))
	})

	It("logs that it's finished", func() {
		_, _ = reconciler.Reconcile(ctx, req)

		Expect(out).To(Say(`"msg":"finished"`))
		Expect(out).To(Say(`"name":"my-workload-name"`))
		Expect(out).To(Say(`"namespace":"my-namespace"`))
	})

	It("updates the status of the workload", func() {
		_, _ = reconciler.Reconcile(ctx, req)

		Expect(repo.StatusUpdateCallCount()).To(Equal(1))
	})

	It("updates the status.observedGeneration to equal metadata.generation", func() {
		_, _ = reconciler.Reconcile(ctx, req)

		updatedWorkload := repo.StatusUpdateArgsForCall(0)

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

		updatedWorkload := repo.StatusUpdateArgsForCall(0)

		Expect(*updatedWorkload.(*v1alpha1.Workload)).To(MatchFields(IgnoreExtras, Fields{
			"Status": MatchFields(IgnoreExtras, Fields{
				"Conditions": Equal(someConditions),
			}),
		}))
	})

	It("requests supply chains from the repo", func() {
		_, _ = reconciler.Reconcile(ctx, req)

		Expect(repo.GetSupplyChainsForWorkloadArgsForCall(0)).To(Equal(wl))
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
			It("returns a helpful error", func() {
				_, err := reconciler.Reconcile(ctx, req)

				Expect(err.Error()).To(ContainSubstring("get object gvk: "))
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
			It("returns a helpful error", func() {
				_, err := reconciler.Reconcile(ctx, req)

				Expect(err.Error()).To(ContainSubstring("supply-chain is not in ready condition"))
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
			})

			Context("of type RetrieveOutputError", func() {
				var retrieveError realizer.RetrieveOutputError
				BeforeEach(func() {
					jsonPathError := templates.NewJsonPathError("this.wont.find.anything", errors.New("some error"))
					retrieveError = realizer.NewRetrieveOutputError(
						&v1alpha1.SupplyChainResource{Name: "some-resource"},
						&jsonPathError)
					rlzr.RealizeReturns(nil, retrieveError)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, req)
					Expect(conditionManager.AddPositiveArgsForCall(1)).To(Equal(workload.MissingValueAtPathCondition("some-resource", "this.wont.find.anything")))
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
			})
		})

		Context("but the watcher returns an error", func() {
			It("logs the error message", func() {
				dynamicTracker.WatchReturns(errors.New("could not watch"))

				_, _ = reconciler.Reconcile(ctx, req)

				Expect(out).To(Say(`"level":"error"`))
				Expect(out).To(Say(`"msg":"dynamic tracker watch"`))
			})
		})

		Context("but status update fails", func() {
			BeforeEach(func() {
				repo.StatusUpdateReturns(errors.New("some error"))
			})

			It("returns a helpful error", func() {
				_, err := reconciler.Reconcile(ctx, req)

				Expect(err).To(MatchError(ContainSubstring("update workload status: ")))
			})

			It("does not requeue", func() { // TODO: Discuss, is this the proper behavior?
				result, _ := reconciler.Reconcile(ctx, req)

				Expect(result).To(Equal(ctrl.Result{Requeue: false}))
			})

			It("logs that an error in updating", func() {
				_, _ = reconciler.Reconcile(ctx, req)

				Expect(out).To(Say(`"msg":"update error"`))
				Expect(out).To(Say(`"name":"my-workload-name"`))
				Expect(out).To(Say(`"namespace":"my-namespace"`))
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

		It("returns a helpful error", func() {
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err.Error()).To(ContainSubstring("workload is missing required labels"))
		})
	})

	Context("and repo returns an empty list of supply chains", func() {
		It("calls the condition manager to add a supply chain not found condition", func() {
			_, _ = reconciler.Reconcile(ctx, req)
			Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(workload.SupplyChainNotFoundCondition(workloadLabels)))
		})

		It("returns a helpful error", func() {
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err.Error()).To(ContainSubstring("no supply chain found where full selector is satisfied by labels: "))
		})
	})

	Context("and repo returns an an error when requesting supply chains", func() {
		BeforeEach(func() {
			repo.GetSupplyChainsForWorkloadReturns(nil, errors.New("some error"))
		})
		It("calls the condition manager to add a supply chain not found condition", func() {
			_, _ = reconciler.Reconcile(ctx, req)
			Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(workload.SupplyChainNotFoundCondition(workloadLabels)))
		})

		It("returns a helpful error", func() {
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err.Error()).To(ContainSubstring("get supply chain by label: some error"))
		})
	})

	Context("and the repo returns multiple supply chains", func() {
		BeforeEach(func() {
			supplyChain := v1alpha1.ClusterSupplyChain{}
			repo.GetSupplyChainsForWorkloadReturns([]v1alpha1.ClusterSupplyChain{supplyChain, supplyChain}, nil)
		})

		It("calls the condition manager to report too mane supply chains matched", func() {
			_, _ = reconciler.Reconcile(ctx, req)
			Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(workload.TooManySupplyChainMatchesCondition()))
		})

		It("returns a helpful error", func() {
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err.Error()).To(ContainSubstring("too many supply chains match the workload selector"))
		})
	})

	Context("but status update fails", func() {
		BeforeEach(func() {
			repo.StatusUpdateReturns(errors.New("some error"))
		})

		It("returns the reconciliation error rather than the update error", func() {
			_, err := reconciler.Reconcile(ctx, req)

			Expect(err).NotTo(MatchError(ContainSubstring("update workload status: ")))
		})

		It("does not requeue", func() { // TODO: Discuss, is this the proper behavior?
			result, _ := reconciler.Reconcile(ctx, req)

			Expect(result).To(Equal(ctrl.Result{Requeue: false}))
		})

		It("logs that an error in updating", func() {
			_, _ = reconciler.Reconcile(ctx, req)

			Expect(out).To(Say(`"msg":"update error"`))
			Expect(out).To(Say(`"name":"my-workload-name"`))
			Expect(out).To(Say(`"namespace":"my-namespace"`))
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

	Context("workload is deleted", func() { // Todo: can we move error handling out of repo to make this more obvious?
		BeforeEach(func() {
			repo.GetWorkloadReturns(nil, kerrors.NewNotFound(schema.GroupResource{
				Group:    "carto.run",
				Resource: "workload",
			}, "some-workload"))
		})
		It("finishes the reconcile and does not requeue", func() {
			result, err := reconciler.Reconcile(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{Requeue: false}))
		})
	})
})
