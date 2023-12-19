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

package mapper_test

import (
	"context"
	"fmt"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/mapper"
	"github.com/vmware-tanzu/cartographer/pkg/mapper/mapperfakes"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/dependency/dependencyfakes"
)

var _ = Describe("MapFunctions", func() {
	var (
		m           *mapper.Mapper
		fakeLogger  *mapperfakes.FakeLogger
		fakeClient  *mapperfakes.FakeClient
		fakeTracker *dependencyfakes.FakeDependencyTracker
	)

	BeforeEach(func() {
		fakeLogger = &mapperfakes.FakeLogger{}
		fakeClient = &mapperfakes.FakeClient{}
		fakeTracker = &dependencyfakes.FakeDependencyTracker{}

		m = &mapper.Mapper{
			Client:  fakeClient,
			Logger:  fakeLogger,
			Tracker: fakeTracker,
		}

		scheme := runtime.NewScheme()
		err := v1alpha1.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())

		err = corev1.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())

		fakeClient.SchemeReturns(scheme)
	})

	// Workload

	Describe("ClusterSupplyChainToWorkloadRequests", func() {
		var supplyChain *v1alpha1.ClusterSupplyChain
		BeforeEach(func() {
			supplyChain = &v1alpha1.ClusterSupplyChain{}
		})

		Context("no workloads", func() {
			BeforeEach(func() {
				fakeClient.ListReturns(nil)
			})
			It("returns an empty list of requests", func() {
				result := m.ClusterSupplyChainToWorkloadRequests(context.Background(), supplyChain)
				Expect(result).To(BeEmpty())
			})
		})
		Context("workloads", func() {
			var workload *v1alpha1.Workload
			BeforeEach(func() {
				workload = &v1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "first-workload",
						Namespace: "first-namespace",
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       "Workload",
						APIVersion: "carto.run/v1alpha1",
					},
				}

				fakeClient.ListStub = func(ctx context.Context, list client.ObjectList, options ...client.ListOption) error {
					listVal := reflect.Indirect(reflect.ValueOf(list))

					existingVal := reflect.Indirect(reflect.ValueOf(v1alpha1.WorkloadList{Items: []v1alpha1.Workload{*workload}}))
					listVal.Set(existingVal)
					return nil
				}
			})

			It("returns a list of requests that includes the workload", func() {
				expected := []reconcile.Request{
					{
						types.NamespacedName{
							Namespace: "first-namespace",
							Name:      "first-workload",
						},
					},
				}

				result := m.ClusterSupplyChainToWorkloadRequests(context.Background(), supplyChain)
				Expect(result).To(Equal(expected))
			})
		})

		Context("client returns an error", func() {
			BeforeEach(func() {
				fakeClient.ListReturns(fmt.Errorf("some-error"))
			})
			It("logs an error to the client", func() {
				result := m.ClusterSupplyChainToWorkloadRequests(context.Background(), supplyChain)
				Expect(result).To(BeEmpty())

				Expect(fakeLogger.ErrorCallCount()).To(Equal(1))
				firstArg, secondArg, _ := fakeLogger.ErrorArgsForCall(0)
				Expect(firstArg).NotTo(BeNil())
				Expect(secondArg).To(Equal("cluster supply chain to workload requests: client list workloads"))
			})
		})
	})

	Describe("ServiceAccountToWorkloadRequests", func() {
		var serviceAccount *corev1.ServiceAccount
		BeforeEach(func() {
			serviceAccount = &corev1.ServiceAccount{}
		})

		Context("there are no workloads tracking the service account", func() {
			BeforeEach(func() {
				fakeTracker.LookupReturns(nil)
			})

			It("returns an empty request list", func() {
				reqs := m.ServiceAccountToWorkloadRequests(context.Background(), serviceAccount)

				Expect(reqs).To(HaveLen(0))
			})
		})

		Context("there is a workload tracking the service account", func() {
			var (
				existingWorkload *v1alpha1.Workload
			)
			BeforeEach(func() {
				existingWorkload = &v1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-workload",
						Namespace: "some-namespace",
					},
					Spec: v1alpha1.WorkloadSpec{
						ServiceAccountName: "some-service-account",
					},
				}

				fakeTracker.LookupReturns([]types.NamespacedName{
					{
						Namespace: existingWorkload.Namespace,
						Name:      existingWorkload.Name,
					},
				})
			})

			It("returns a request for the workload", func() {
				reqs := m.ServiceAccountToWorkloadRequests(context.Background(), serviceAccount)

				Expect(reqs).To(HaveLen(1))
				Expect(reqs[0].Name).To(Equal("some-workload"))
			})
		})
	})

	Describe("RoleBindingToWorkloadRequests", func() {
		var roleBinding *rbacv1.RoleBinding
		BeforeEach(func() {
			roleBinding = &rbacv1.RoleBinding{
				Subjects: []rbacv1.Subject{
					{
						Kind:      "ServiceAccount",
						Name:      "some-service-account",
						Namespace: "some-namespace",
					},
				},
			}
		})

		Context("there are no workloads tracking the service account subject", func() {
			BeforeEach(func() {
				fakeTracker.LookupReturns(nil)
			})

			It("returns an empty request list", func() {
				reqs := m.RoleBindingToWorkloadRequests(context.Background(), roleBinding)

				Expect(reqs).To(HaveLen(0))
			})
		})

		Context("there is a workload tracking the service account subject", func() {
			var existingWorkload *v1alpha1.Workload
			BeforeEach(func() {
				existingWorkload = &v1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-workload",
						Namespace: "some-namespace",
					},
				}

				fakeTracker.LookupReturns([]types.NamespacedName{
					{
						Namespace: existingWorkload.Namespace,
						Name:      existingWorkload.Name,
					},
				})
			})

			It("returns a request for the workload", func() {
				reqs := m.RoleBindingToWorkloadRequests(context.Background(), roleBinding)

				Expect(reqs).To(HaveLen(1))
				Expect(reqs[0].Name).To(Equal("some-workload"))
			})
		})
	})

	Describe("ClusterRoleBindingToWorkloadRequests", func() {
		var (
			clusterRoleBinding *rbacv1.ClusterRoleBinding
		)

		BeforeEach(func() {
			clusterRoleBinding = &rbacv1.ClusterRoleBinding{
				Subjects: []rbacv1.Subject{
					{
						Kind:      "ServiceAccount",
						Name:      "some-service-account",
						Namespace: "some-namespace",
					},
				},
			}
		})

		Context("there are no workloads tracking the service account subject", func() {
			BeforeEach(func() {
				fakeTracker.LookupReturns(nil)
			})

			It("returns an empty request list", func() {
				reqs := m.ClusterRoleBindingToWorkloadRequests(context.Background(), clusterRoleBinding)

				Expect(reqs).To(HaveLen(0))
			})
		})

		Context("there is a workload tracking the service account subject", func() {
			var existingWorkload *v1alpha1.Workload
			BeforeEach(func() {
				existingWorkload = &v1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-workload",
						Namespace: "some-namespace",
					},
				}

				fakeTracker.LookupReturns([]types.NamespacedName{
					{
						Namespace: existingWorkload.Namespace,
						Name:      existingWorkload.Name,
					},
				})
			})

			It("returns a request for the workload", func() {
				reqs := m.ClusterRoleBindingToWorkloadRequests(context.Background(), clusterRoleBinding)

				Expect(reqs).To(HaveLen(1))
				Expect(reqs[0].Name).To(Equal("some-workload"))
			})
		})
	})

	Describe("RoleToWorkloadRequests", func() {
		var (
			role *rbacv1.Role
		)

		BeforeEach(func() {
			role = &rbacv1.Role{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-role",
				},
			}
		})

		Context("there is a workload tracking the service account subject", func() {
			var existingWorkload *v1alpha1.Workload
			BeforeEach(func() {
				existingWorkload = &v1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-workload",
						Namespace: "some-namespace",
					},
				}

				fakeTracker.LookupReturns([]types.NamespacedName{
					{
						Namespace: existingWorkload.Namespace,
						Name:      existingWorkload.Name,
					},
				})
			})

			Context("there is a role binding containing the role", func() {
				BeforeEach(func() {
					fakeClient.ListStub = func(ctx context.Context, list client.ObjectList, options ...client.ListOption) error {
						listVal := reflect.Indirect(reflect.ValueOf(list))
						roleBinding := &rbacv1.RoleBinding{
							Subjects: []rbacv1.Subject{
								{
									Kind:      "ServiceAccount",
									Name:      "some-service-account",
									Namespace: "some-namespace",
								},
							},
							RoleRef: rbacv1.RoleRef{
								APIGroup: "",
								Kind:     "Role",
								Name:     role.Name,
							},
						}

						existingVal := reflect.Indirect(reflect.ValueOf(rbacv1.RoleBindingList{Items: []rbacv1.RoleBinding{*roleBinding}}))
						listVal.Set(existingVal)
						return nil
					}
				})

				It("returns a request for the workload", func() {
					reqs := m.RoleToWorkloadRequests(context.Background(), role)

					Expect(reqs).To(HaveLen(1))
					Expect(reqs[0].Name).To(Equal("some-workload"))
				})
			})
		})

		Context("client.list errors", func() {
			var (
				listErr error
			)
			BeforeEach(func() {
				listErr = fmt.Errorf("some error")

				fakeClient.ListReturns(listErr)
			})

			It("returns the error", func() {
				reqs := m.RoleToWorkloadRequests(context.Background(), role)

				Expect(reqs).To(HaveLen(0))
				Expect(fakeLogger.ErrorCallCount()).To(Equal(1))

				err, msg, _ := fakeLogger.ErrorArgsForCall(0)
				Expect(err).To(MatchError(listErr))
				Expect(msg).To(Equal("role to workload requests: list role bindings"))
			})
		})
	})

	Describe("ClusterRoleToWorkloadRequests", func() {
		var (
			clusterRole *rbacv1.ClusterRole
		)

		BeforeEach(func() {
			clusterRole = &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-cluster-role",
				},
			}
		})

		Context("there is a workload tracking a service account", func() {
			var existingWorkload *v1alpha1.Workload
			BeforeEach(func() {
				existingWorkload = &v1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-workload",
						Namespace: "some-namespace",
					},
				}

				fakeTracker.LookupReturns([]types.NamespacedName{
					{
						Namespace: existingWorkload.Namespace,
						Name:      existingWorkload.Name,
					},
				})
			})

			Context("there is a cluster role binding containing the role and service account", func() {
				BeforeEach(func() {
					fakeClient.ListStub = func(ctx context.Context, list client.ObjectList, options ...client.ListOption) error {
						switch list.(type) {
						case *rbacv1.ClusterRoleBindingList:
							listVal := reflect.Indirect(reflect.ValueOf(list))
							clusterRoleBinding := &rbacv1.ClusterRoleBinding{
								Subjects: []rbacv1.Subject{
									{
										Kind:      "ServiceAccount",
										Name:      "some-service-account",
										Namespace: "some-namespace",
									},
								},
								RoleRef: rbacv1.RoleRef{
									APIGroup: "",
									Kind:     "ClusterRole",
									Name:     clusterRole.Name,
								},
							}

							existingVal := reflect.Indirect(reflect.ValueOf(rbacv1.ClusterRoleBindingList{Items: []rbacv1.ClusterRoleBinding{*clusterRoleBinding}}))
							listVal.Set(existingVal)
							return nil
						default:
							return nil
						}
					}
				})

				It("returns a request for the workload", func() {
					reqs := m.ClusterRoleToWorkloadRequests(context.Background(), clusterRole)

					Expect(reqs).To(HaveLen(1))
					Expect(reqs[0].Name).To(Equal("some-workload"))
				})
			})

			Context("there is a role binding containing the role and service account", func() {
				BeforeEach(func() {
					fakeClient.ListStub = func(ctx context.Context, list client.ObjectList, options ...client.ListOption) error {
						switch list.(type) {
						case *rbacv1.RoleBindingList:
							listVal := reflect.Indirect(reflect.ValueOf(list))
							roleBinding := &rbacv1.RoleBinding{
								Subjects: []rbacv1.Subject{
									{
										Kind:      "ServiceAccount",
										Name:      "some-service-account",
										Namespace: "some-namespace",
									},
								},
								RoleRef: rbacv1.RoleRef{
									APIGroup: "",
									Kind:     "ClusterRole",
									Name:     clusterRole.Name,
								},
							}

							existingVal := reflect.Indirect(reflect.ValueOf(rbacv1.RoleBindingList{Items: []rbacv1.RoleBinding{*roleBinding}}))
							listVal.Set(existingVal)
							return nil
						default:
							return nil
						}
					}
				})

				It("returns a request for the workload", func() {
					reqs := m.ClusterRoleToWorkloadRequests(context.Background(), clusterRole)

					Expect(reqs).To(HaveLen(1))
					Expect(reqs[0].Name).To(Equal("some-workload"))
				})
			})
		})

		Context("client.list errors", func() {
			var (
				listErr error
			)
			BeforeEach(func() {
				listErr = fmt.Errorf("some error")

				fakeClient.ListReturns(listErr)
			})

			It("returns the error", func() {
				reqs := m.ClusterRoleToWorkloadRequests(context.Background(), clusterRole)

				Expect(reqs).To(HaveLen(0))
				Expect(fakeLogger.ErrorCallCount()).To(Equal(1))

				err, msg, _ := fakeLogger.ErrorArgsForCall(0)
				Expect(err).To(MatchError(listErr))
				Expect(msg).To(Equal("cluster role to workload requests: list cluster role bindings"))
			})
		})
	})

	// Deliverable

	Describe("ClusterDeliveryToDeliverableRequests", func() {
		var delivery *v1alpha1.ClusterDelivery
		BeforeEach(func() {
			delivery = &v1alpha1.ClusterDelivery{}
		})

		Context("no deliverables", func() {
			BeforeEach(func() {
				fakeClient.ListReturns(nil)
			})
			It("returns an empty list of requests", func() {
				result := m.ClusterDeliveryToDeliverableRequests(context.Background(), delivery)
				Expect(result).To(BeEmpty())
			})
		})
		Context("deliverables", func() {
			var deliverable *v1alpha1.Deliverable
			BeforeEach(func() {
				deliverable = &v1alpha1.Deliverable{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "first-deliverable",
						Namespace: "first-namespace",
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       "Deliverable",
						APIVersion: "carto.run/v1alpha1",
					},
				}

				fakeClient.ListStub = func(ctx context.Context, list client.ObjectList, options ...client.ListOption) error {
					listVal := reflect.Indirect(reflect.ValueOf(list))

					existingVal := reflect.Indirect(reflect.ValueOf(v1alpha1.DeliverableList{Items: []v1alpha1.Deliverable{*deliverable}}))
					listVal.Set(existingVal)
					return nil
				}
			})

			It("returns a list of requests that includes the deliverable", func() {
				expected := []reconcile.Request{
					{
						types.NamespacedName{
							Namespace: "first-namespace",
							Name:      "first-deliverable",
						},
					},
				}

				result := m.ClusterDeliveryToDeliverableRequests(context.Background(), delivery)
				Expect(result).To(Equal(expected))
			})
		})

		Context("client returns an error", func() {
			BeforeEach(func() {
				fakeClient.ListReturns(fmt.Errorf("some-error"))
			})
			It("logs an error to the client", func() {
				result := m.ClusterDeliveryToDeliverableRequests(context.Background(), delivery)
				Expect(result).To(BeEmpty())

				Expect(fakeLogger.ErrorCallCount()).To(Equal(1))
				firstArg, secondArg, _ := fakeLogger.ErrorArgsForCall(0)
				Expect(firstArg).NotTo(BeNil())
				Expect(secondArg).To(Equal("cluster delivery to deliverable requests: client list deliverables"))
			})
		})
	})

	Describe("ServiceAccountToDeliverableRequests", func() {
		var serviceAccount *corev1.ServiceAccount
		BeforeEach(func() {
			serviceAccount = &corev1.ServiceAccount{}
		})

		Context("there are no deliverables tracking the service account", func() {
			BeforeEach(func() {
				fakeTracker.LookupReturns(nil)
			})

			It("returns an empty request list", func() {
				reqs := m.ServiceAccountToDeliverableRequests(context.Background(), serviceAccount)

				Expect(reqs).To(HaveLen(0))
			})
		})

		Context("there is a deliverable tracking the service account", func() {
			var (
				existingDeliverable *v1alpha1.Deliverable
			)
			BeforeEach(func() {
				existingDeliverable = &v1alpha1.Deliverable{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-deliverable",
						Namespace: "some-namespace",
					},
					Spec: v1alpha1.DeliverableSpec{
						ServiceAccountName: "some-service-account",
					},
				}

				fakeTracker.LookupReturns([]types.NamespacedName{
					{
						Namespace: existingDeliverable.Namespace,
						Name:      existingDeliverable.Name,
					},
				})
			})

			It("returns a request for the deliverable", func() {
				reqs := m.ServiceAccountToDeliverableRequests(context.Background(), serviceAccount)

				Expect(reqs).To(HaveLen(1))
				Expect(reqs[0].Name).To(Equal("some-deliverable"))
			})
		})
	})

	Describe("RoleBindingToDeliverableRequests", func() {
		var roleBinding *rbacv1.RoleBinding
		BeforeEach(func() {
			roleBinding = &rbacv1.RoleBinding{
				Subjects: []rbacv1.Subject{
					{
						Kind:      "ServiceAccount",
						Name:      "some-service-account",
						Namespace: "some-namespace",
					},
				},
			}
		})

		Context("there are no deliverables tracking the service account subject", func() {
			BeforeEach(func() {
				fakeTracker.LookupReturns(nil)
			})

			It("returns an empty request list", func() {
				reqs := m.RoleBindingToDeliverableRequests(context.Background(), roleBinding)

				Expect(reqs).To(HaveLen(0))
			})
		})

		Context("there is a deliverable tracking the service account subject", func() {
			var existingDeliverable *v1alpha1.Deliverable
			BeforeEach(func() {
				existingDeliverable = &v1alpha1.Deliverable{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-deliverable",
						Namespace: "some-namespace",
					},
				}

				fakeTracker.LookupReturns([]types.NamespacedName{
					{
						Namespace: existingDeliverable.Namespace,
						Name:      existingDeliverable.Name,
					},
				})
			})

			It("returns a request for the deliverable", func() {
				reqs := m.RoleBindingToDeliverableRequests(context.Background(), roleBinding)

				Expect(reqs).To(HaveLen(1))
				Expect(reqs[0].Name).To(Equal("some-deliverable"))
			})
		})
	})

	Describe("ClusterRoleBindingToDeliverableRequests", func() {
		var (
			clusterRoleBinding *rbacv1.ClusterRoleBinding
		)

		BeforeEach(func() {
			clusterRoleBinding = &rbacv1.ClusterRoleBinding{
				Subjects: []rbacv1.Subject{
					{
						Kind:      "ServiceAccount",
						Name:      "some-service-account",
						Namespace: "some-namespace",
					},
				},
			}
		})

		Context("there are no deliverables tracking the service account subject", func() {
			BeforeEach(func() {
				fakeTracker.LookupReturns(nil)
			})

			It("returns an empty request list", func() {
				reqs := m.ClusterRoleBindingToDeliverableRequests(context.Background(), clusterRoleBinding)

				Expect(reqs).To(HaveLen(0))
			})
		})

		Context("there is a deliverable tracking the service account subject", func() {
			var existingDeliverable *v1alpha1.Deliverable
			BeforeEach(func() {
				existingDeliverable = &v1alpha1.Deliverable{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-deliverable",
						Namespace: "some-namespace",
					},
				}

				fakeTracker.LookupReturns([]types.NamespacedName{
					{
						Namespace: existingDeliverable.Namespace,
						Name:      existingDeliverable.Name,
					},
				})
			})

			It("returns a request for the deliverable", func() {
				reqs := m.ClusterRoleBindingToDeliverableRequests(context.Background(), clusterRoleBinding)

				Expect(reqs).To(HaveLen(1))
				Expect(reqs[0].Name).To(Equal("some-deliverable"))
			})
		})
	})

	Describe("RoleToDeliverableRequests", func() {
		var (
			role *rbacv1.Role
		)

		BeforeEach(func() {
			role = &rbacv1.Role{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-role",
				},
			}
		})

		Context("there is a deliverable tracking a service account", func() {
			var existingDeliverable *v1alpha1.Deliverable
			BeforeEach(func() {
				existingDeliverable = &v1alpha1.Deliverable{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-deliverable",
						Namespace: "some-namespace",
					},
				}

				fakeTracker.LookupReturns([]types.NamespacedName{
					{
						Namespace: existingDeliverable.Namespace,
						Name:      existingDeliverable.Name,
					},
				})
			})

			Context("there is a role binding containing the role and service account", func() {
				BeforeEach(func() {
					fakeClient.ListStub = func(ctx context.Context, list client.ObjectList, options ...client.ListOption) error {
						listVal := reflect.Indirect(reflect.ValueOf(list))
						roleBinding := &rbacv1.RoleBinding{
							Subjects: []rbacv1.Subject{
								{
									Kind:      "ServiceAccount",
									Name:      "some-service-account",
									Namespace: "some-namespace",
								},
							},
							RoleRef: rbacv1.RoleRef{
								APIGroup: "",
								Kind:     "Role",
								Name:     role.Name,
							},
						}

						existingVal := reflect.Indirect(reflect.ValueOf(rbacv1.RoleBindingList{Items: []rbacv1.RoleBinding{*roleBinding}}))
						listVal.Set(existingVal)
						return nil
					}
				})

				It("returns a request for the deliverable", func() {
					reqs := m.RoleToDeliverableRequests(context.Background(), role)

					Expect(reqs).To(HaveLen(1))
					Expect(reqs[0].Name).To(Equal("some-deliverable"))
				})
			})
		})

		Context("client.list errors", func() {
			var (
				listErr error
			)
			BeforeEach(func() {
				listErr = fmt.Errorf("some error")

				fakeClient.ListReturns(listErr)
			})

			It("returns the error", func() {
				reqs := m.RoleToDeliverableRequests(context.Background(), role)

				Expect(reqs).To(HaveLen(0))
				Expect(fakeLogger.ErrorCallCount()).To(Equal(1))

				err, msg, _ := fakeLogger.ErrorArgsForCall(0)
				Expect(err).To(MatchError(listErr))
				Expect(msg).To(Equal("role to deliverable requests: list role bindings"))
			})
		})
	})

	Describe("ClusterRoleToDeliverableRequests", func() {
		var (
			clusterRole *rbacv1.ClusterRole
		)

		BeforeEach(func() {
			clusterRole = &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-cluster-role",
				},
			}
		})

		Context("there is a deliverable tracking a service account", func() {
			var existingDeliverable *v1alpha1.Deliverable
			BeforeEach(func() {
				existingDeliverable = &v1alpha1.Deliverable{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-deliverable",
						Namespace: "some-namespace",
					},
				}

				fakeTracker.LookupReturns([]types.NamespacedName{
					{
						Namespace: existingDeliverable.Namespace,
						Name:      existingDeliverable.Name,
					},
				})
			})

			Context("there is a cluster role binding containing the role and service account", func() {
				BeforeEach(func() {
					fakeClient.ListStub = func(ctx context.Context, list client.ObjectList, options ...client.ListOption) error {
						switch list.(type) {
						case *rbacv1.ClusterRoleBindingList:
							listVal := reflect.Indirect(reflect.ValueOf(list))
							clusterRoleBinding := &rbacv1.ClusterRoleBinding{
								Subjects: []rbacv1.Subject{
									{
										Kind:      "ServiceAccount",
										Name:      "some-service-account",
										Namespace: "some-namespace",
									},
								},
								RoleRef: rbacv1.RoleRef{
									APIGroup: "",
									Kind:     "ClusterRole",
									Name:     clusterRole.Name,
								},
							}

							existingVal := reflect.Indirect(reflect.ValueOf(rbacv1.ClusterRoleBindingList{Items: []rbacv1.ClusterRoleBinding{*clusterRoleBinding}}))
							listVal.Set(existingVal)
							return nil
						default:
							return nil
						}
					}
				})

				It("returns a request for the deliverable", func() {
					reqs := m.ClusterRoleToDeliverableRequests(context.Background(), clusterRole)

					Expect(reqs).To(HaveLen(1))
					Expect(reqs[0].Name).To(Equal("some-deliverable"))
				})
			})

			Context("there is a role binding containing the role and service account", func() {
				BeforeEach(func() {
					fakeClient.ListStub = func(ctx context.Context, list client.ObjectList, options ...client.ListOption) error {
						switch list.(type) {
						case *rbacv1.RoleBindingList:
							listVal := reflect.Indirect(reflect.ValueOf(list))
							roleBinding := &rbacv1.RoleBinding{
								Subjects: []rbacv1.Subject{
									{
										Kind:      "ServiceAccount",
										Name:      "some-service-account",
										Namespace: "some-namespace",
									},
								},
								RoleRef: rbacv1.RoleRef{
									APIGroup: "",
									Kind:     "ClusterRole",
									Name:     clusterRole.Name,
								},
							}

							existingVal := reflect.Indirect(reflect.ValueOf(rbacv1.RoleBindingList{Items: []rbacv1.RoleBinding{*roleBinding}}))
							listVal.Set(existingVal)
							return nil
						default:
							return nil
						}
					}
				})

				It("returns a request for the deliverable", func() {
					reqs := m.ClusterRoleToDeliverableRequests(context.Background(), clusterRole)

					Expect(reqs).To(HaveLen(1))
					Expect(reqs[0].Name).To(Equal("some-deliverable"))
				})
			})
		})

		Context("client.list errors", func() {
			var (
				listErr error
			)
			BeforeEach(func() {
				listErr = fmt.Errorf("some error")

				fakeClient.ListReturns(listErr)
			})

			It("returns the error", func() {
				reqs := m.ClusterRoleToDeliverableRequests(context.Background(), clusterRole)

				Expect(reqs).To(HaveLen(0))
				Expect(fakeLogger.ErrorCallCount()).To(Equal(1))

				err, msg, _ := fakeLogger.ErrorArgsForCall(0)
				Expect(err).To(MatchError(listErr))
				Expect(msg).To(Equal("cluster role to deliverable requests: list cluster role bindings"))
			})
		})
	})

	// Runnable

	Describe("ServiceAccountToRunnableRequests", func() {
		var serviceAccount *corev1.ServiceAccount
		BeforeEach(func() {
			serviceAccount = &corev1.ServiceAccount{}
		})

		Context("there are no runnables tracking the service account", func() {
			BeforeEach(func() {
				fakeTracker.LookupReturns(nil)
			})

			It("returns an empty request list", func() {
				reqs := m.ServiceAccountToRunnableRequests(context.Background(), serviceAccount)

				Expect(reqs).To(HaveLen(0))
			})
		})

		Context("there is a runnable tracking the service account", func() {
			var (
				existingRunnable *v1alpha1.Runnable
			)
			BeforeEach(func() {
				existingRunnable = &v1alpha1.Runnable{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-runnable",
						Namespace: "some-namespace",
					},
					Spec: v1alpha1.RunnableSpec{
						ServiceAccountName: "some-service-account",
					},
				}

				fakeTracker.LookupReturns([]types.NamespacedName{
					{
						Namespace: existingRunnable.Namespace,
						Name:      existingRunnable.Name,
					},
				})
			})

			It("returns a request for the runnable", func() {
				reqs := m.ServiceAccountToRunnableRequests(context.Background(), serviceAccount)

				Expect(reqs).To(HaveLen(1))
				Expect(reqs[0].Name).To(Equal("some-runnable"))
			})
		})
	})

	Describe("RoleBindingToRunnableRequests", func() {
		var roleBinding *rbacv1.RoleBinding
		BeforeEach(func() {
			roleBinding = &rbacv1.RoleBinding{
				Subjects: []rbacv1.Subject{
					{
						Kind:      "ServiceAccount",
						Name:      "some-service-account",
						Namespace: "some-namespace",
					},
				},
			}
		})

		Context("there are no runnables tracking the service account subject", func() {
			BeforeEach(func() {
				fakeTracker.LookupReturns(nil)
			})

			It("returns an empty request list", func() {
				reqs := m.RoleBindingToRunnableRequests(context.Background(), roleBinding)

				Expect(reqs).To(HaveLen(0))
			})
		})

		Context("there is a runnable tracking the service account subject", func() {
			var existingRunnable *v1alpha1.Runnable
			BeforeEach(func() {
				existingRunnable = &v1alpha1.Runnable{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-runnable",
						Namespace: "some-namespace",
					},
				}

				fakeTracker.LookupReturns([]types.NamespacedName{
					{
						Namespace: existingRunnable.Namespace,
						Name:      existingRunnable.Name,
					},
				})
			})

			It("returns a request for the runnable", func() {
				reqs := m.RoleBindingToRunnableRequests(context.Background(), roleBinding)

				Expect(reqs).To(HaveLen(1))
				Expect(reqs[0].Name).To(Equal("some-runnable"))
			})
		})
	})

	Describe("ClusterRoleBindingToRunnableRequests", func() {
		var (
			clusterRoleBinding *rbacv1.ClusterRoleBinding
		)

		BeforeEach(func() {
			clusterRoleBinding = &rbacv1.ClusterRoleBinding{
				Subjects: []rbacv1.Subject{
					{
						Kind:      "ServiceAccount",
						Name:      "some-service-account",
						Namespace: "some-namespace",
					},
				},
			}
		})

		Context("there are no runnables tracking the service account subject", func() {
			BeforeEach(func() {
				fakeTracker.LookupReturns(nil)
			})

			It("returns an empty request list", func() {
				reqs := m.ClusterRoleBindingToRunnableRequests(context.Background(), clusterRoleBinding)

				Expect(reqs).To(HaveLen(0))
			})
		})

		Context("there is a runnable tracking the service account subject", func() {
			var existingRunnable *v1alpha1.Runnable
			BeforeEach(func() {
				existingRunnable = &v1alpha1.Runnable{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-runnable",
						Namespace: "some-namespace",
					},
				}

				fakeTracker.LookupReturns([]types.NamespacedName{
					{
						Namespace: existingRunnable.Namespace,
						Name:      existingRunnable.Name,
					},
				})
			})

			It("returns a request for the runnable", func() {
				reqs := m.ClusterRoleBindingToRunnableRequests(context.Background(), clusterRoleBinding)

				Expect(reqs).To(HaveLen(1))
				Expect(reqs[0].Name).To(Equal("some-runnable"))
			})
		})
	})

	Describe("RoleToRunnableRequests", func() {
		var (
			role *rbacv1.Role
		)

		BeforeEach(func() {
			role = &rbacv1.Role{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-role",
				},
			}
		})

		Context("there is a runnable tracking the service account subject", func() {
			var existingRunnable *v1alpha1.Runnable
			BeforeEach(func() {
				existingRunnable = &v1alpha1.Runnable{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-runnable",
						Namespace: "some-namespace",
					},
				}

				fakeTracker.LookupReturns([]types.NamespacedName{
					{
						Namespace: existingRunnable.Namespace,
						Name:      existingRunnable.Name,
					},
				})
			})

			Context("there is a role binding containing the role", func() {
				BeforeEach(func() {
					fakeClient.ListStub = func(ctx context.Context, list client.ObjectList, options ...client.ListOption) error {
						listVal := reflect.Indirect(reflect.ValueOf(list))
						roleBinding := &rbacv1.RoleBinding{
							Subjects: []rbacv1.Subject{
								{
									Kind:      "ServiceAccount",
									Name:      "some-service-account",
									Namespace: "some-namespace",
								},
							},
							RoleRef: rbacv1.RoleRef{
								APIGroup: "",
								Kind:     "Role",
								Name:     role.Name,
							},
						}

						existingVal := reflect.Indirect(reflect.ValueOf(rbacv1.RoleBindingList{Items: []rbacv1.RoleBinding{*roleBinding}}))
						listVal.Set(existingVal)
						return nil
					}
				})

				It("returns a request for the runnable", func() {
					reqs := m.RoleToRunnableRequests(context.Background(), role)

					Expect(reqs).To(HaveLen(1))
					Expect(reqs[0].Name).To(Equal("some-runnable"))
				})
			})
		})

		Context("client.list errors", func() {
			var (
				listErr error
			)
			BeforeEach(func() {
				listErr = fmt.Errorf("some error")

				fakeClient.ListReturns(listErr)
			})

			It("returns the error", func() {
				reqs := m.RoleToRunnableRequests(context.Background(), role)

				Expect(reqs).To(HaveLen(0))
				Expect(fakeLogger.ErrorCallCount()).To(Equal(1))

				err, msg, _ := fakeLogger.ErrorArgsForCall(0)
				Expect(err).To(MatchError(listErr))
				Expect(msg).To(Equal("role to runnable requests: list role bindings"))
			})
		})
	})

	Describe("ClusterRoleToRunnableRequests", func() {
		var (
			clusterRole *rbacv1.ClusterRole
		)

		BeforeEach(func() {
			clusterRole = &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-cluster-role",
				},
			}
		})

		Context("there is a runnable tracking a service account", func() {
			var existingRunnable *v1alpha1.Runnable
			BeforeEach(func() {
				existingRunnable = &v1alpha1.Runnable{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-runnable",
						Namespace: "some-namespace",
					},
				}

				fakeTracker.LookupReturns([]types.NamespacedName{
					{
						Namespace: existingRunnable.Namespace,
						Name:      existingRunnable.Name,
					},
				})
			})

			Context("there is a cluster role binding containing the role and service account", func() {
				BeforeEach(func() {
					fakeClient.ListStub = func(ctx context.Context, list client.ObjectList, options ...client.ListOption) error {
						switch list.(type) {
						case *rbacv1.ClusterRoleBindingList:
							listVal := reflect.Indirect(reflect.ValueOf(list))
							clusterRoleBinding := &rbacv1.ClusterRoleBinding{
								Subjects: []rbacv1.Subject{
									{
										Kind:      "ServiceAccount",
										Name:      "some-service-account",
										Namespace: "some-namespace",
									},
								},
								RoleRef: rbacv1.RoleRef{
									APIGroup: "",
									Kind:     "ClusterRole",
									Name:     clusterRole.Name,
								},
							}

							existingVal := reflect.Indirect(reflect.ValueOf(rbacv1.ClusterRoleBindingList{Items: []rbacv1.ClusterRoleBinding{*clusterRoleBinding}}))
							listVal.Set(existingVal)
							return nil
						default:
							return nil
						}
					}
				})

				It("returns a request for the runnable", func() {
					reqs := m.ClusterRoleToRunnableRequests(context.Background(), clusterRole)

					Expect(reqs).To(HaveLen(1))
					Expect(reqs[0].Name).To(Equal("some-runnable"))
				})
			})

			Context("there is a role binding containing the role and service account", func() {
				BeforeEach(func() {
					fakeClient.ListStub = func(ctx context.Context, list client.ObjectList, options ...client.ListOption) error {
						switch list.(type) {
						case *rbacv1.RoleBindingList:
							listVal := reflect.Indirect(reflect.ValueOf(list))
							roleBinding := &rbacv1.RoleBinding{
								Subjects: []rbacv1.Subject{
									{
										Kind:      "ServiceAccount",
										Name:      "some-service-account",
										Namespace: "some-namespace",
									},
								},
								RoleRef: rbacv1.RoleRef{
									APIGroup: "",
									Kind:     "ClusterRole",
									Name:     clusterRole.Name,
								},
							}

							existingVal := reflect.Indirect(reflect.ValueOf(rbacv1.RoleBindingList{Items: []rbacv1.RoleBinding{*roleBinding}}))
							listVal.Set(existingVal)
							return nil
						default:
							return nil
						}
					}
				})

				It("returns a request for the runnable", func() {
					reqs := m.ClusterRoleToRunnableRequests(context.Background(), clusterRole)

					Expect(reqs).To(HaveLen(1))
					Expect(reqs[0].Name).To(Equal("some-runnable"))
				})
			})
		})

		Context("client.list errors", func() {
			var (
				listErr error
			)
			BeforeEach(func() {
				listErr = fmt.Errorf("some error")

				fakeClient.ListReturns(listErr)
			})

			It("returns the error", func() {
				reqs := m.ClusterRoleToRunnableRequests(context.Background(), clusterRole)

				Expect(reqs).To(HaveLen(0))
				Expect(fakeLogger.ErrorCallCount()).To(Equal(1))

				err, msg, _ := fakeLogger.ErrorArgsForCall(0)
				Expect(err).To(MatchError(listErr))
				Expect(msg).To(Equal("cluster role to runnable requests: list cluster role bindings"))
			})
		})
	})
})
