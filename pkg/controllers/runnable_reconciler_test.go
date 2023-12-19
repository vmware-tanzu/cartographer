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

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/conditions/conditionsfakes"
	"github.com/vmware-tanzu/cartographer/pkg/controllers"
	"github.com/vmware-tanzu/cartographer/pkg/controllers/controllersfakes"
	cerrors "github.com/vmware-tanzu/cartographer/pkg/errors"
	"github.com/vmware-tanzu/cartographer/pkg/events/eventsfakes"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/runnable/runnablefakes"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
	"github.com/vmware-tanzu/cartographer/pkg/satoken/satokenfakes"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/dependency/dependencyfakes"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/stamped/stampedfakes"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

var _ = Describe("Reconcile", func() {
	var (
		out                      *Buffer
		ctx                      context.Context
		reconciler               controllers.RunnableReconciler
		request                  controllerruntime.Request
		repo                     *repositoryfakes.FakeRepository
		tokenManager             *satokenfakes.FakeTokenManager
		rlzr                     *runnablefakes.FakeRealizer
		stampedTracker           *stampedfakes.FakeStampedTracker
		dependencyTracker        *dependencyfakes.FakeDependencyTracker
		conditionManager         *conditionsfakes.FakeConditionManager
		builtClient              *repositoryfakes.FakeClient
		clientForBuiltRepository *client.Client
		cacheForBuiltRepository  *repository.RepoCache
		fakeCache                *repositoryfakes.FakeRepoCache
		fakeRunnabeRepo          *repositoryfakes.FakeRepository
		fakeEventRecorder        *eventsfakes.FakeEventRecorder
		serviceAccount           *corev1.ServiceAccount
		authTokenForBuiltClient  string
		serviceAccountName       = "alternate-service-account-name"
		serviceAccountToken      = "alternate-service-account-token"
		fakeDiscoveryClient      *runnablefakes.FakeDiscoveryInterface
	)

	BeforeEach(func() {
		out = NewBuffer()
		logger := zap.New(zap.WriteTo(out))
		ctx = logr.NewContext(context.Background(), logger)
		repo = &repositoryfakes.FakeRepository{}
		tokenManager = &satokenfakes.FakeTokenManager{}
		rlzr = &runnablefakes.FakeRealizer{}
		stampedTracker = &stampedfakes.FakeStampedTracker{}
		dependencyTracker = &dependencyfakes.FakeDependencyTracker{}
		conditionManager = &conditionsfakes.FakeConditionManager{}
		fakeCache = &repositoryfakes.FakeRepoCache{}
		fakeDiscoveryClient = &runnablefakes.FakeDiscoveryInterface{}
		fakeEventRecorder = &eventsfakes.FakeEventRecorder{}

		scheme := runtime.NewScheme()
		err := utils.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())
		repo.GetSchemeReturns(scheme)
		fakeRESTMapper := controllersfakes.FakeRESTMapper{}
		repo.GetRESTMapperReturns(&fakeRESTMapper)

		serviceAccount = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Name: serviceAccountName},
		}
		repo.GetServiceAccountReturns(serviceAccount, nil)
		tokenManager.GetServiceAccountTokenReturns(serviceAccountToken, nil)

		fakeConditionManagerBuilder := func(string, []metav1.Condition) conditions.ConditionManager {
			return conditionManager
		}

		repositoryBuilder := func(client client.Client, repoCache repository.RepoCache) repository.Repository {
			clientForBuiltRepository = &client
			cacheForBuiltRepository = &repoCache
			return fakeRunnabeRepo
		}

		builtClient = &repositoryfakes.FakeClient{}
		clientBuilder := func(authToken string, _ bool) (client.Client, discovery.DiscoveryInterface, error) {
			authTokenForBuiltClient = authToken
			return builtClient, fakeDiscoveryClient, nil
		}

		reconciler = controllers.RunnableReconciler{
			Repo:                    repo,
			TokenManager:            tokenManager,
			Realizer:                rlzr,
			StampedTracker:          stampedTracker,
			ConditionManagerBuilder: fakeConditionManagerBuilder,
			RunnableCache:           fakeCache,
			ClientBuilder:           clientBuilder,
			RepositoryBuilder:       repositoryBuilder,
			DependencyTracker:       dependencyTracker,
			EventRecorder:           fakeEventRecorder,
			RESTMapper:              &fakeRESTMapper,
			Scheme:                  scheme,
		}

		request = controllerruntime.Request{
			NamespacedName: types.NamespacedName{
				Name:      "my-runnable",
				Namespace: "my-namespace",
			},
		}
	})

	Context("reconcile a new valid Runnable", func() {
		var rb *v1alpha1.Runnable
		BeforeEach(func() {
			rb = &v1alpha1.Runnable{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Runnable",
					APIVersion: "carto.run/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:       "my-runnable",
					Namespace:  "my-namespace",
					Generation: 1,
				},
				Spec: v1alpha1.RunnableSpec{
					RunTemplateRef: v1alpha1.TemplateReference{
						Kind: "RunTemplateRef",
						Name: "my-run-template",
					},
					ServiceAccountName: serviceAccountName,
				},
			}
			repo.GetRunnableReturns(rb, nil)
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

			_, _ = reconciler.Reconcile(ctx, request)

			_, updatedRunnable := repo.StatusUpdateArgsForCall(0)

			Expect(*updatedRunnable.(*v1alpha1.Runnable)).To(MatchFields(IgnoreExtras, Fields{
				"Status": MatchFields(IgnoreExtras, Fields{
					"Conditions": Equal(someConditions),
				}),
			}))
		})

		It("uses the service account specified by the runnable for realizing resources", func() {
			_, _ = reconciler.Reconcile(ctx, request)

			Expect(repo.GetServiceAccountCallCount()).To(Equal(1))
			_, serviceAccountName, serviceAccountNS := repo.GetServiceAccountArgsForCall(0)
			Expect(serviceAccountName).To(Equal(serviceAccountName))
			Expect(serviceAccountNS).To(Equal("my-namespace"))

			Expect(tokenManager.GetServiceAccountTokenArgsForCall(0)).To(Equal(serviceAccount))

			Expect(*clientForBuiltRepository).To(Equal(builtClient))
			Expect(authTokenForBuiltClient).To(Equal(serviceAccountToken))
			Expect(*cacheForBuiltRepository).To(Equal(reconciler.RunnableCache))

			Expect(rlzr.RealizeCallCount()).To(Equal(1))
			_, _, systemRepo, runnableRepo, discoveryClient := rlzr.RealizeArgsForCall(0)
			Expect(systemRepo).To(Equal(repo))
			Expect(runnableRepo).To(Equal(fakeRunnabeRepo))
			Expect(discoveryClient).To(Equal(fakeDiscoveryClient))
		})

		It("updates the status.observedGeneration to equal metadata.generation", func() {
			_, _ = reconciler.Reconcile(ctx, request)

			_, updatedRunnable := repo.StatusUpdateArgsForCall(0)

			Expect(*updatedRunnable.(*v1alpha1.Runnable)).To(MatchFields(IgnoreExtras, Fields{
				"Status": MatchFields(IgnoreExtras, Fields{
					"ObservedGeneration": BeEquivalentTo(1),
				}),
			}))
		})

		It("clears the previously tracked objects for the runnable", func() {
			_, _ = reconciler.Reconcile(ctx, request)

			Expect(dependencyTracker.ClearTrackedCallCount()).To(Equal(1))
			obj := dependencyTracker.ClearTrackedArgsForCall(0)
			Expect(obj.Name).To(Equal("my-runnable"))
			Expect(obj.Namespace).To(Equal("my-namespace"))
		})

		It("watches the cluster run template and service account", func() {
			_, _ = reconciler.Reconcile(ctx, request)

			Expect(dependencyTracker.TrackCallCount()).To(Equal(2))
			serviceAccountKey, _ := dependencyTracker.TrackArgsForCall(0)
			Expect(serviceAccountKey.String()).To(Equal("ServiceAccount/my-namespace/alternate-service-account-name"))

			runTemplateKey, _ := dependencyTracker.TrackArgsForCall(1)
			Expect(runTemplateKey.String()).To(Equal("ClusterRunTemplate.carto.run//my-run-template"))
		})

		Context("watching does not cause an error", func() {
			It("watches the stampedObject's kind", func() {
				stampedObject := &unstructured.Unstructured{}
				stampedObject.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   "thing.io",
					Version: "alphabeta1",
					Kind:    "MyThing",
				})
				rlzr.RealizeReturns(stampedObject, nil, nil)

				_, _ = reconciler.Reconcile(ctx, request)
				Expect(stampedTracker.WatchCallCount()).To(Equal(1))
				_, obj, hndl, _ := stampedTracker.WatchArgsForCall(0)

				Expect(obj).To(Equal(stampedObject))
				Expect(hndl).To(Equal(handler.EnqueueRequestForOwner(repo.GetScheme(), repo.GetRESTMapper(), &v1alpha1.Runnable{})))
			})
		})

		Context("watching causes an error", func() {
			BeforeEach(func() {
				stampedObject := &unstructured.Unstructured{}
				rlzr.RealizeReturns(stampedObject, nil, nil)

				stampedTracker.WatchReturns(errors.New("could not watch"))
			})

			It("logs the error message", func() {
				_, _ = reconciler.Reconcile(ctx, request)

				Expect(out).To(Say(`"level":"error"`))
				Expect(out).To(Say(`"msg":"failed to add informer for object"`))
			})

			It("returns an unhandled error and requeues", func() {
				_, err := reconciler.Reconcile(ctx, request)

				Expect(err.Error()).To(ContainSubstring("could not watch"))
			})
		})

		Context("requesting the service account causes an error", func() {
			var repoError error
			BeforeEach(func() {
				repoError = errors.New("some error")
				repo.GetServiceAccountReturns(nil, repoError)
			})

			It("calls the condition manager to add a service account not found condition", func() {
				_, _ = reconciler.Reconcile(ctx, request)
				Expect(conditionManager.AddPositiveCallCount()).To(BeNumerically(">", 0))
				Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(conditions.RunnableServiceAccountNotFoundCondition(repoError)))
			})

			It("handles the error and logs it", func() {
				_, err := reconciler.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())

				Expect(out).To(Say(`"level":"info"`))
				Expect(out).To(Say(`"handled error":"failed to get service account \[my-namespace/alternate-service-account-name\]: some error"`))
			})
		})

		Context("requesting a service account token causes an error", func() {
			var tokenError error
			BeforeEach(func() {
				tokenError = errors.New("some error")
				tokenManager.GetServiceAccountTokenReturns("", tokenError)
			})

			It("calls the condition manager to add a service account not found condition", func() {
				_, _ = reconciler.Reconcile(ctx, request)
				Expect(conditionManager.AddPositiveCallCount()).To(BeNumerically(">", 0))
				Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(conditions.RunnableServiceAccountTokenErrorCondition(tokenError)))
			})

			It("handles the error and logs it", func() {
				_, err := reconciler.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())

				Expect(out).To(Say(`"level":"info"`))
				Expect(out).To(Say(`"handled error":"failed to get token for service account \[my-namespace/alternate-service-account-name\]: some error"`))
			})
		})

		Context("no outputs were returned from the realizer", func() {
			BeforeEach(func() {
				rlzr.RealizeReturns(nil, nil, nil)
			})

			It("fetches the runnable", func() {
				_, _ = reconciler.Reconcile(ctx, request)

				Expect(repo.GetRunnableCallCount()).To(Equal(1))
				_, actualName, actualNamespace := repo.GetRunnableArgsForCall(0)
				Expect(actualName).To(Equal("my-runnable"))
				Expect(actualNamespace).To(Equal("my-namespace"))
			})

			It("Starts and Finishes cleanly", func() {
				_, err := reconciler.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())

				Expect(out).To(Say(`"msg":"started"`))
				Expect(out).To(Say(`"msg":"finished"`))
			})

			It("Updates the status with no outputs", func() {
				_, err := reconciler.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())

				Expect(repo.StatusUpdateCallCount()).To(Equal(1))
				_, obj := repo.StatusUpdateArgsForCall(0)
				statusObject, ok := obj.(*v1alpha1.Runnable)
				Expect(ok).To(BeTrue())

				Expect(statusObject.Status.Outputs).To(HaveLen(0))
			})
		})

		Context("outputs are returned from the realizer", func() {
			BeforeEach(func() {
				rlzr.RealizeReturns(nil, templates.Outputs{
					"an-output": apiextensionsv1.JSON{Raw: []byte(`"the value"`)},
				}, nil)
			})

			It("Updates the status with the outputs", func() {
				_, err := reconciler.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())

				Expect(repo.StatusUpdateCallCount()).To(Equal(1))
				_, obj := repo.StatusUpdateArgsForCall(0)
				statusObject, ok := obj.(*v1alpha1.Runnable)
				Expect(ok).To(BeTrue())

				Expect(statusObject.Status.Outputs).To(HaveLen(1))
				Expect(statusObject.Status.Outputs["an-output"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"the value"`)}))

			})

			Context("the runnable already had outputs in the status", func() {
				BeforeEach(func() {
					rb.Status.Outputs = map[string]apiextensionsv1.JSON{
						"old-output": {Raw: []byte(`"old value"`)},
					}
					rb.Status.ObservedGeneration = 1
				})

				It("updates the status with the new outputs", func() {
					_, err := reconciler.Reconcile(ctx, request)
					Expect(err).NotTo(HaveOccurred())

					Expect(repo.StatusUpdateCallCount()).To(Equal(1))
					_, obj := repo.StatusUpdateArgsForCall(0)
					statusObject, ok := obj.(*v1alpha1.Runnable)
					Expect(ok).To(BeTrue())

					Expect(statusObject.Status.Outputs).To(HaveLen(1))
					Expect(statusObject.Status.Outputs["an-output"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"the value"`)}))
				})
			})
		})

		Context("updating the status fails", func() {
			BeforeEach(func() {
				rlzr.RealizeReturns(nil, nil, nil)
				repo.StatusUpdateReturns(errors.New("bad status update error"))
			})

			It("Starts and Finishes cleanly", func() {
				_, _ = reconciler.Reconcile(ctx, request)
				Expect(out).To(Say(`"msg":"started"`))
				Expect(out).To(Say(`"msg":"finished"`))
			})

			It("returns a status error", func() {
				result, err := reconciler.Reconcile(ctx, request)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to update status for runnable: bad status update error"))
				Expect(result).To(Equal(controllerruntime.Result{}))
			})
		})

		Context("the realizer returns an error", func() {
			BeforeEach(func() {
				rlzr.RealizeReturns(nil, nil, nil)
			})

			It("Starts and Finishes cleanly", func() {
				_, err := reconciler.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())

				Expect(out).To(Say(`"msg":"started"`))
				Expect(out).To(Say(`"msg":"finished"`))
			})

			It("does not try to watch stampedObject", func() {
				_, err := reconciler.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())

				Expect(stampedTracker.WatchCallCount()).To(Equal(0))
			})

			Context("of type GetRunTemplateError", func() {
				var err error
				BeforeEach(func() {
					err = cerrors.RunnableGetRunTemplateError{
						Err:         errors.New("some error"),
						TemplateRef: &v1alpha1.TemplateReference{Kind: "ClusterRunTemplate", Name: "my-run-template"},
					}
					rlzr.RealizeReturns(nil, nil, err)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, request)
					Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(conditions.RunTemplateMissingCondition(err)))
				})

				It("returns an unhandled error and requeues", func() {
					_, err := reconciler.Reconcile(ctx, request)

					Expect(err.Error()).To(ContainSubstring("unable to get run template [my-run-template]: some error"))
				})
			})

			Context("of type ResolveSelectorError", func() {
				var err error
				BeforeEach(func() {
					err = cerrors.RunnableResolveSelectorError{
						Err: errors.New("some error"),
						Selector: &v1alpha1.ResourceSelector{
							Resource: v1alpha1.ResourceType{
								APIVersion: "my-api-version",
								Kind:       "my-kind",
							},
							MatchingLabels: map[string]string{"foo": "bar", "moo": "cow"},
						},
					}
					rlzr.RealizeReturns(nil, nil, err)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, request)
					Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(conditions.RunnableTemplateStampFailureCondition(err)))
				})

				It("does not return an error", func() {
					_, err := reconciler.Reconcile(ctx, request)
					Expect(err).NotTo(HaveOccurred())
				})

				It("logs the handled error message", func() {
					_, _ = reconciler.Reconcile(ctx, request)

					Expect(out).To(Say(`"level":"info"`))
					Expect(out).To(Say(`"msg":"handled error reconciling runnable"`))
					Expect(out).To(Say(`"handled error":"unable to resolve selector \[map\[foo:bar moo:cow\]\], apiVersion \[my-api-version\], kind \[my-kind\]: some error"`))
				})
			})

			Context("of type StampError", func() {
				var err error
				BeforeEach(func() {
					err = cerrors.RunnableStampError{
						Err:         errors.New("some error"),
						TemplateRef: &v1alpha1.TemplateReference{Kind: "ClusterRunTemplate", Name: "my-run-template"},
					}
					rlzr.RealizeReturns(nil, nil, err)
				})

				It("does not try to watch the stampedObjects", func() {
					_, err := reconciler.Reconcile(ctx, request)
					Expect(err).NotTo(HaveOccurred())

					Expect(stampedTracker.WatchCallCount()).To(Equal(0))
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, request)
					Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(conditions.RunnableTemplateStampFailureCondition(err)))
				})

				It("does not return an error", func() {
					_, err := reconciler.Reconcile(ctx, request)
					Expect(err).NotTo(HaveOccurred())
				})

				It("logs the handled error message", func() {
					_, _ = reconciler.Reconcile(ctx, request)

					Expect(out).To(Say(`"level":"info"`))
					Expect(out).To(Say(`"msg":"handled error reconciling runnable"`))
					Expect(out).To(Say(`"handled error":"unable to stamp object for run template \[my-run-template\]: some error"`))
				})
			})

			Context("of type ApplyStampedObjectError", func() {
				var err error
				BeforeEach(func() {
					err = cerrors.RunnableApplyStampedObjectError{
						Err:           errors.New("some error"),
						StampedObject: &unstructured.Unstructured{},
						TemplateRef:   &v1alpha1.TemplateReference{Kind: "ClusterRunTemplate", Name: "my-run-template"},
					}
					rlzr.RealizeReturns(nil, nil, err)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, request)
					Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(conditions.StampedObjectRejectedByAPIServerCondition(err)))
				})

				It("returns an unhandled error and requeues", func() {
					_, err := reconciler.Reconcile(ctx, request)

					Expect(err.Error()).To(ContainSubstring("unable to apply object"))
				})
			})

			Context("of type ApplyStampedObjectError where the user did not have proper permissions", func() {
				var stampedObjectError cerrors.RunnableApplyStampedObjectError
				BeforeEach(func() {
					status := &metav1.Status{
						Message: "fantastic error",
						Reason:  metav1.StatusReasonForbidden,
						Code:    403,
					}
					stampedObject := &unstructured.Unstructured{}
					stampedObject.SetGroupVersionKind(schema.GroupVersionKind{
						Group:   "thing.io",
						Version: "alphabeta1",
						Kind:    "MyThing",
					})
					stampedObject.SetNamespace("a-namespace")
					stampedObject.SetName("a-name")

					stampedObjectError = cerrors.RunnableApplyStampedObjectError{
						Err:           kerrors.FromObject(status),
						StampedObject: stampedObject,
						TemplateRef:   &v1alpha1.TemplateReference{Kind: "ClusterRunTemplate", Name: "my-run-template"},
					}

					rlzr.RealizeReturns(nil, nil, stampedObjectError)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, request)
					Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(conditions.StampedObjectRejectedByAPIServerCondition(stampedObjectError)))
				})

				It("handles the error and logs it", func() {
					_, err := reconciler.Reconcile(ctx, request)
					Expect(err).NotTo(HaveOccurred())

					Expect(out).To(Say(`"level":"info"`))
					Expect(out).To(Say(`"handled error":"unable to apply object \[a-namespace/a-name\] for run template \[my-run-template\]: fantastic error"`))
				})
			})

			Context("of type ListCreatedObjectsError", func() {
				var err error
				BeforeEach(func() {
					err = cerrors.ListCreatedObjectsError{
						Err:       errors.New("some error"),
						Namespace: "some-ns",
						Labels:    map[string]string{"hi": "bye"},
					}
					rlzr.RealizeReturns(nil, nil, err)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, request)
					Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(conditions.FailedToListCreatedObjectsCondition(err)))
				})

				It("returns an unhandled error and requeues", func() {
					_, err := reconciler.Reconcile(ctx, request)

					Expect(err.Error()).To(ContainSubstring("unable to list objects in namespace [some-ns] with labels [map[hi:bye]]: some error"))
				})
			})

			Context("of type RetrieveOutputError", func() {
				var err error
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

					err = cerrors.RunnableRetrieveOutputError{
						Err:               errors.New("some error"),
						TemplateRef:       &v1alpha1.TemplateReference{Kind: "ClusterRunTemplate", Name: "my-run-template"},
						StampedObject:     stampedObject,
						QualifiedResource: "mything.thing.io",
					}
					rlzr.RealizeReturns(nil, nil, err)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, request)
					Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(conditions.OutputPathNotSatisfiedCondition(stampedObject, "mything.thing.io", err.Error())))
				})

				It("does not return an error", func() {
					_, err := reconciler.Reconcile(ctx, request)
					Expect(err).NotTo(HaveOccurred())
				})

				It("logs the handled error message", func() {
					_, _ = reconciler.Reconcile(ctx, request)

					Expect(out).To(Say(`"level":"info"`))
					Expect(out).To(Say(`"msg":"handled error reconciling runnable"`))
					Expect(out).To(Say(`"handled error":"unable to retrieve outputs from stamped object \[my-ns/my-obj\] of type \[mything.thing.io\] for run template \[my-run-template\]: some error"`))
				})
			})

			Context("of unknown type", func() {
				var err error
				BeforeEach(func() {
					err = errors.New("some error")
					rlzr.RealizeReturns(nil, nil, err)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, request)
					Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(conditions.UnknownErrorCondition(err)))
				})

				It("returns an unhandled error and requeues", func() {
					_, err := reconciler.Reconcile(ctx, request)

					Expect(err.Error()).To(ContainSubstring("some error"))
				})
			})
		})

		Context("the runnable does not specify a service account", func() {
			defaultServiceAccountToken := "default-service-account-token"
			defaultServiceAccount := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{Name: "default"},
			}

			BeforeEach(func() {
				rb.Spec.ServiceAccountName = ""
				repo.GetRunnableReturns(rb, nil)
				repo.GetServiceAccountReturns(defaultServiceAccount, nil)
				tokenManager.GetServiceAccountTokenReturns(defaultServiceAccountToken, nil)
			})

			It("uses the default service account in the namespace", func() {
				_, _ = reconciler.Reconcile(ctx, request)

				Expect(repo.GetServiceAccountCallCount()).To(Equal(1))
				_, serviceAccountName, serviceAccountNS := repo.GetServiceAccountArgsForCall(0)
				Expect(serviceAccountName).To(Equal("default"))
				Expect(serviceAccountNS).To(Equal("my-namespace"))

				Expect(tokenManager.GetServiceAccountTokenArgsForCall(0)).To(Equal(defaultServiceAccount))
				Expect(authTokenForBuiltClient).To(Equal(defaultServiceAccountToken))
			})
		})
	})

	Context("the runnable goes away", func() {
		BeforeEach(func() {
			repo.GetRunnableReturns(nil, nil)
		})

		It("considers the reconcile complete", func() {
			result, err := reconciler.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(controllerruntime.Result{}))
		})

		It("logs that we saw the runnable go away", func() {
			_, err := reconciler.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())

			Expect(out).To(Say(`"msg":"runnable no longer exists"`))
		})

		It("Starts and Finishes cleanly", func() {
			_, err := reconciler.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())

			Eventually(out).Should(Say(`"msg":"started"`))
			Eventually(out).Should(Say(`"msg":"finished"`))
		})

		It("clears the previously tracked objects for the runnable", func() {
			_, _ = reconciler.Reconcile(ctx, request)

			Expect(dependencyTracker.ClearTrackedCallCount()).To(Equal(1))
			obj := dependencyTracker.ClearTrackedArgsForCall(0)
			Expect(obj.Name).To(Equal("my-runnable"))
			Expect(obj.Namespace).To(Equal("my-namespace"))
		})
	})

	Context("the runnable fetch is in error", func() {
		BeforeEach(func() {
			repo.GetRunnableReturns(nil, errors.New("very bad runnable"))
		})

		It("returns an error and requeues", func() {
			_, err := reconciler.Reconcile(ctx, request)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("very bad runnable"))
		})
	})
})
