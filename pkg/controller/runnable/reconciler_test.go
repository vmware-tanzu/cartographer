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

package runnable_test

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/conditions/conditionsfakes"
	"github.com/vmware-tanzu/cartographer/pkg/controller/runnable"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/runnable"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/runnable/runnablefakes"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/trackerfakes"
)

var _ = Describe("Reconcile", func() {
	var (
		out              *Buffer
		ctx              context.Context
		reconciler       runnable.Reconciler
		request          controllerruntime.Request
		repository       *repositoryfakes.FakeRepository
		rlzr             *runnablefakes.FakeRealizer
		dynamicTracker   *trackerfakes.FakeDynamicTracker
		conditionManager *conditionsfakes.FakeConditionManager
	)

	BeforeEach(func() {
		out = NewBuffer()
		logger := zap.New(zap.WriteTo(out))
		ctx = logr.NewContext(context.Background(), logger)
		repository = &repositoryfakes.FakeRepository{}
		rlzr = &runnablefakes.FakeRealizer{}
		dynamicTracker = &trackerfakes.FakeDynamicTracker{}
		conditionManager = &conditionsfakes.FakeConditionManager{}

		fakeConditionManagerBuilder := func(string, []metav1.Condition) conditions.ConditionManager {
			return conditionManager
		}

		reconciler = runnable.Reconciler{
			Repo:                    repository,
			Realizer:                rlzr,
			DynamicTracker:          dynamicTracker,
			ConditionManagerBuilder: fakeConditionManagerBuilder,
		}

		request = controllerruntime.Request{
			NamespacedName: types.NamespacedName{
				Name:      "my-runnable",
				Namespace: "my-namespace",
			},
		}
	})

	Context("reconcile a new valid Runnable", func() {
		BeforeEach(func() {
			repository.GetRunnableReturns(&v1alpha1.Runnable{
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
				},
			}, nil)

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

			updatedRunnable := repository.StatusUpdateArgsForCall(0)

			Expect(*updatedRunnable.(*v1alpha1.Runnable)).To(MatchFields(IgnoreExtras, Fields{
				"Status": MatchFields(IgnoreExtras, Fields{
					"Conditions": Equal(someConditions),
				}),
			}))
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
				Expect(dynamicTracker.WatchCallCount()).To(Equal(1))
				_, obj, hndl := dynamicTracker.WatchArgsForCall(0)

				Expect(obj).To(Equal(stampedObject))
				Expect(hndl).To(Equal(&handler.EnqueueRequestForOwner{OwnerType: &v1alpha1.Runnable{}}))
			})
		})

		Context("watching causes an error", func() {
			BeforeEach(func() {
				stampedObject := &unstructured.Unstructured{}
				rlzr.RealizeReturns(stampedObject, nil, nil)

				dynamicTracker.WatchReturns(errors.New("could not watch"))
			})

			It("logs the error message", func() {
				_, _ = reconciler.Reconcile(ctx, request)

				Expect(out).To(Say(`"level":"error"`))
				Expect(out).To(Say(`"msg":"dynamic tracker watch"`))
			})

			It("returns an unhandled error and requeues", func() {
				_, err := reconciler.Reconcile(ctx, request)

				Expect(err.Error()).To(ContainSubstring("could not watch"))
			})
		})

		Context("no outputs were returned from the realizer", func() {
			BeforeEach(func() {
				rlzr.RealizeReturns(nil, nil, nil)
			})

			It("fetches the runnable", func() {
				_, _ = reconciler.Reconcile(ctx, request)

				Expect(repository.GetRunnableCallCount()).To(Equal(1))
				actualName, actualNamespace := repository.GetRunnableArgsForCall(0)
				Expect(actualName).To(Equal("my-runnable"))
				Expect(actualNamespace).To(Equal("my-namespace"))
			})

			It("Starts and Finishes cleanly", func() {
				_, err := reconciler.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())

				Expect(out).To(Say(`"msg":"started","name":"my-runnable","namespace":"my-namespace"`))
				Expect(out).To(Say(`"msg":"finished","name":"my-runnable","namespace":"my-namespace"`))
			})

			It("Updates the status with no outputs", func() {
				_, err := reconciler.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())

				Expect(repository.StatusUpdateCallCount()).To(Equal(1))
				statusObject, ok := repository.StatusUpdateArgsForCall(0).(*v1alpha1.Runnable)
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

				Expect(repository.StatusUpdateCallCount()).To(Equal(1))
				statusObject, ok := repository.StatusUpdateArgsForCall(0).(*v1alpha1.Runnable)
				Expect(ok).To(BeTrue())

				Expect(statusObject.Status.Outputs).To(HaveLen(1))
				Expect(statusObject.Status.Outputs["an-output"]).To(Equal(apiextensionsv1.JSON{Raw: []byte(`"the value"`)}))

			})
		})

		Context("updating the status fails", func() {
			BeforeEach(func() {
				rlzr.RealizeReturns(nil, nil, nil)
				repository.StatusUpdateReturns(errors.New("bad status update error"))
			})

			It("Starts and Finishes cleanly", func() {
				_, _ = reconciler.Reconcile(ctx, request)
				Expect(out).To(Say(`"msg":"started","name":"my-runnable","namespace":"my-namespace"`))
				Expect(out).To(Say(`"msg":"finished","name":"my-runnable","namespace":"my-namespace"`))
			})

			It("returns a status error", func() {
				result, err := reconciler.Reconcile(ctx, request)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("update runnable status"))
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

				Expect(out).To(Say(`"msg":"started","name":"my-runnable","namespace":"my-namespace"`))
				Expect(out).To(Say(`"msg":"finished","name":"my-runnable","namespace":"my-namespace"`))
			})

			It("does not try to watch stampedObject", func() {
				_, err := reconciler.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())

				Expect(dynamicTracker.WatchCallCount()).To(Equal(0))
			})

			Context("of type GetRunTemplateError", func() {
				var err error
				BeforeEach(func() {
					err = realizer.GetRunTemplateError{
						Err:      errors.New("some error"),
						Runnable: &v1alpha1.Runnable{ObjectMeta: metav1.ObjectMeta{Name: "my-runnable", Namespace: "my-ns"}},
					}
					rlzr.RealizeReturns(nil, nil, err)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, request)
					Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(runnable.RunTemplateMissingCondition(err)))
				})

				It("returns an unhandled error and requeues", func() {
					_, err := reconciler.Reconcile(ctx, request)

					Expect(err.Error()).To(ContainSubstring("unable to get runnable 'my-ns/my-runnable': 'some error'"))
				})
			})

			Context("of type ResolveSelectorError", func() {
				var err error
				BeforeEach(func() {
					err = realizer.ResolveSelectorError{
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
					Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(runnable.TemplateStampFailureCondition(err)))
				})

				It("does not return an error", func() {
					_, err := reconciler.Reconcile(ctx, request)
					Expect(err).NotTo(HaveOccurred())
				})

				It("logs the handled error message", func() {
					_, _ = reconciler.Reconcile(ctx, request)

					Expect(out).To(Say(`"level":"info"`))
					Expect(out).To(Say(`"msg":"handled error"`))
					Expect(out).To(Say(`"error":"unable to resolve selector '\(apiVersion:my-api-version kind:my-kind labels:map\[foo:bar moo:cow\]\)': 'some error'"`))
				})
			})

			Context("of type StampError", func() {
				var err error
				BeforeEach(func() {
					err = realizer.StampError{
						Err:      errors.New("some error"),
						Runnable: &v1alpha1.Runnable{ObjectMeta: metav1.ObjectMeta{Name: "my-runnable", Namespace: "my-ns"}},
					}
					rlzr.RealizeReturns(nil, nil, err)
				})

				It("does not try to watch the stampedObjects", func() {
					_, err := reconciler.Reconcile(ctx, request)
					Expect(err).NotTo(HaveOccurred())

					Expect(dynamicTracker.WatchCallCount()).To(Equal(0))
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, request)
					Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(runnable.TemplateStampFailureCondition(err)))
				})

				It("does not return an error", func() {
					_, err := reconciler.Reconcile(ctx, request)
					Expect(err).NotTo(HaveOccurred())
				})

				It("logs the handled error message", func() {
					_, _ = reconciler.Reconcile(ctx, request)

					Expect(out).To(Say(`"level":"info"`))
					Expect(out).To(Say(`"msg":"handled error"`))
					Expect(out).To(Say(`"error":"unable to stamp object 'my-ns/my-runnable': 'some error'"`))
				})
			})

			Context("of type ApplyStampedObjectError", func() {
				var err error
				BeforeEach(func() {
					err = realizer.ApplyStampedObjectError{
						Err:           errors.New("some error"),
						StampedObject: &unstructured.Unstructured{},
					}
					rlzr.RealizeReturns(nil, nil, err)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, request)
					Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(runnable.StampedObjectRejectedByAPIServerCondition(err)))
				})

				It("returns an unhandled error and requeues", func() {
					_, err := reconciler.Reconcile(ctx, request)

					Expect(err.Error()).To(ContainSubstring("unable to apply stamped object"))
				})
			})

			Context("of type ListCreatedObjectsError", func() {
				var err error
				BeforeEach(func() {
					err = realizer.ListCreatedObjectsError{
						Err:       errors.New("some error"),
						Namespace: "some-ns",
						Labels:    map[string]string{"hi": "bye"},
					}
					rlzr.RealizeReturns(nil, nil, err)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, request)
					Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(runnable.FailedToListCreatedObjectsCondition(err)))
				})

				It("returns an unhandled error and requeues", func() {
					_, err := reconciler.Reconcile(ctx, request)

					Expect(err.Error()).To(ContainSubstring("unable to list objects in namespace 'some-ns' with labels 'map[hi:bye]': 'some error'"))
				})
			})

			Context("of type RetrieveOutputError", func() {
				var err error
				BeforeEach(func() {
					err = realizer.RetrieveOutputError{
						Err:      errors.New("some error"),
						Runnable: &v1alpha1.Runnable{ObjectMeta: metav1.ObjectMeta{Name: "my-runnable", Namespace: "my-ns"}},
					}
					rlzr.RealizeReturns(nil, nil, err)
				})

				It("calls the condition manager to report", func() {
					_, _ = reconciler.Reconcile(ctx, request)
					Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(runnable.OutputPathNotSatisfiedCondition(err)))
				})

				It("does not return an error", func() {
					_, err := reconciler.Reconcile(ctx, request)
					Expect(err).NotTo(HaveOccurred())
				})

				It("logs the handled error message", func() {
					_, _ = reconciler.Reconcile(ctx, request)

					Expect(out).To(Say(`"level":"info"`))
					Expect(out).To(Say(`"msg":"handled error"`))
					Expect(out).To(Say(`"error":"unable to retrieve outputs from stamped object for runnable 'my-ns/my-runnable': some error"`))
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
					Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(runnable.UnknownErrorCondition(err)))
				})

				It("returns an unhandled error and requeues", func() {
					_, err := reconciler.Reconcile(ctx, request)

					Expect(err.Error()).To(ContainSubstring("some error"))
				})
			})
		})
	})

	Context("the runnable goes away", func() {
		BeforeEach(func() {
			repository.GetRunnableReturns(nil, nil)
		})

		It("considers the reconcile complete", func() {
			result, err := reconciler.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(controllerruntime.Result{}))
		})

		It("logs that we saw the runnable go away", func() {
			_, err := reconciler.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())

			Expect(out).To(Say(`"msg":"runnable no longer exists","name":"my-runnable","namespace":"my-namespace"`))
		})

		It("Starts and Finishes cleanly", func() {
			_, err := reconciler.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())

			Eventually(out).Should(Say(`"msg":"started","name":"my-runnable","namespace":"my-namespace"`))
			Eventually(out).Should(Say(`"msg":"finished","name":"my-runnable","namespace":"my-namespace"`))
		})
	})

	Context("the runnable fetch is in error", func() {
		BeforeEach(func() {
			repository.GetRunnableReturns(nil, errors.New("very bad runnable"))
		})

		It("returns an error and requeues", func() {
			_, err := reconciler.Reconcile(ctx, request)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("very bad runnable"))
		})
	})
})
