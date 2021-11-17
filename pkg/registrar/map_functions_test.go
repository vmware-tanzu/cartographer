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

package registrar_test

import (
	"context"
	"fmt"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/registrar"
	"github.com/vmware-tanzu/cartographer/pkg/registrar/registrarfakes"
)

var _ = Describe("MapFunctions", func() {
	Describe("TemplateToDeliverableRequests", func() {
		var (
			m          *registrar.Mapper
			fakeLogger *registrarfakes.FakeLogger
			fakeClient *registrarfakes.FakeClient
		)

		Context("the template kind can be found", func() {
			BeforeEach(func() {
				fakeLogger = &registrarfakes.FakeLogger{}
				fakeClient = &registrarfakes.FakeClient{}

				m = &registrar.Mapper{
					Client: fakeClient,
					Logger: fakeLogger,
				}

				scheme := runtime.NewScheme()
				err := v1alpha1.AddToScheme(scheme)
				Expect(err).NotTo(HaveOccurred())

				fakeClient.SchemeReturns(scheme)
			})

			Context("client.list does not return errors", func() {

				Context("there are no Deliveries", func() {
					BeforeEach(func() {
						existingList := v1alpha1.ClusterDeliveryList{
							Items: []v1alpha1.ClusterDelivery{},
						}

						fakeClient.ListStub = func(ctx context.Context, list client.ObjectList, option ...client.ListOption) error {
							listVal := reflect.ValueOf(list)
							existingVal := reflect.ValueOf(existingList)

							reflect.Indirect(listVal).Set(reflect.Indirect(existingVal))
							return nil
						}
					})

					It("returns an empty request list", func() {
						t := &v1alpha1.ClusterTemplate{}
						reqs := m.TemplateToDeliverableRequests(t)

						Expect(reqs).To(HaveLen(0))
					})
				})

				Context("there are multiple Deliveries", func() {
					BeforeEach(func() {
						existingDelivery1 := v1alpha1.ClusterDelivery{
							Spec: v1alpha1.ClusterDeliverySpec{
								Resources: []v1alpha1.ClusterDeliveryResource{
									{
										TemplateRef: v1alpha1.DeliveryClusterTemplateReference{
											Kind: "ClusterTemplate",
											Name: "my-template-foo",
										},
									},
								},
							},
						}
						existingDelivery2 := v1alpha1.ClusterDelivery{
							ObjectMeta: metav1.ObjectMeta{Name: "good-supply-chain"},
							Spec: v1alpha1.ClusterDeliverySpec{
								Resources: []v1alpha1.ClusterDeliveryResource{
									{
										TemplateRef: v1alpha1.DeliveryClusterTemplateReference{
											Kind: "ClusterTemplate",
											Name: "my-template",
										},
									},
								},
							},
						}
						existingDeliveryList := v1alpha1.ClusterDeliveryList{
							Items: []v1alpha1.ClusterDelivery{existingDelivery1, existingDelivery2},
						}

						existingDeliverable := v1alpha1.Deliverable{
							ObjectMeta: metav1.ObjectMeta{
								Name: "my-deliverable",
							},
						}
						existingDeliverableList := v1alpha1.DeliverableList{
							Items: []v1alpha1.Deliverable{existingDeliverable},
						}

						fakeClient.ListStub = func(ctx context.Context, list client.ObjectList, option ...client.ListOption) error {
							listVal := reflect.Indirect(reflect.ValueOf(list))
							switch list.(type) {
							case *v1alpha1.ClusterDeliveryList:
								existingVal := reflect.Indirect(reflect.ValueOf(existingDeliveryList))
								listVal.Set(existingVal)
							case *v1alpha1.DeliverableList:
								existingVal := reflect.Indirect(reflect.ValueOf(existingDeliverableList))
								listVal.Set(existingVal)
							default:
								panic("list type not stubbed")
							}

							return nil
						}
					})

					Describe("The template refers to some Deliveries", func() {
						It("returns requests for only the matching Deliverables", func() {
							t := &v1alpha1.ClusterTemplate{
								ObjectMeta: metav1.ObjectMeta{
									Name: "my-template",
								},
							}
							reqs := m.TemplateToDeliverableRequests(t)

							Expect(reqs).To(HaveLen(1))
							Expect(reqs[0].Name).To(Equal("my-deliverable"))
						})
					})

					Describe("The template does not reference a Delivery", func() {
						It("returns an empty request list", func() {
							t := &v1alpha1.ClusterTemplate{
								ObjectMeta: metav1.ObjectMeta{
									Name: "my-template-bar",
								},
							}
							reqs := m.TemplateToDeliverableRequests(t)

							Expect(reqs).To(HaveLen(0))
						})
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
					t := &v1alpha1.ClusterTemplate{}
					reqs := m.TemplateToDeliverableRequests(t)

					Expect(reqs).To(HaveLen(0))
					Expect(fakeLogger.ErrorCallCount()).To(Equal(1))

					err, msg, _ := fakeLogger.ErrorArgsForCall(0)
					Expect(err).To(Equal(listErr))
					Expect(msg).To(Equal("list ClusterDeliveries"))
				})
			})
		})

	})

	Describe("TemplateToWorkloadRequests", func() {
		var (
			m          *registrar.Mapper
			fakeLogger *registrarfakes.FakeLogger
			fakeClient *registrarfakes.FakeClient
		)

		Context("the template kind can be found", func() {
			BeforeEach(func() {
				fakeLogger = &registrarfakes.FakeLogger{}
				fakeClient = &registrarfakes.FakeClient{}

				m = &registrar.Mapper{
					Client: fakeClient,
					Logger: fakeLogger,
				}

				scheme := runtime.NewScheme()
				err := v1alpha1.AddToScheme(scheme)
				Expect(err).NotTo(HaveOccurred())

				fakeClient.SchemeReturns(scheme)
			})

			Context("client.list does not return errors", func() {

				Context("there are no SupplyChains", func() {
					BeforeEach(func() {
						existingList := v1alpha1.ClusterSupplyChainList{
							Items: []v1alpha1.ClusterSupplyChain{},
						}

						fakeClient.ListStub = func(ctx context.Context, list client.ObjectList, option ...client.ListOption) error {
							listVal := reflect.ValueOf(list)
							existingVal := reflect.ValueOf(existingList)

							reflect.Indirect(listVal).Set(reflect.Indirect(existingVal))
							return nil
						}
					})

					It("returns an empty request list", func() {
						t := &v1alpha1.ClusterTemplate{}
						reqs := m.TemplateToWorkloadRequests(t)

						Expect(reqs).To(HaveLen(0))
					})
				})

				Context("there are multiple supply chains", func() {
					BeforeEach(func() {
						existingSupplyChain1 := v1alpha1.ClusterSupplyChain{
							Spec: v1alpha1.SupplyChainSpec{
								Resources: []v1alpha1.SupplyChainResource{
									{
										TemplateRef: v1alpha1.ClusterTemplateReference{
											Kind: "ClusterTemplate",
											Name: "my-template-foo",
										},
									},
								},
							},
						}
						existingSupplyChain2 := v1alpha1.ClusterSupplyChain{
							ObjectMeta: metav1.ObjectMeta{Name: "good-supply-chain"},
							Spec: v1alpha1.SupplyChainSpec{
								Resources: []v1alpha1.SupplyChainResource{
									{
										TemplateRef: v1alpha1.ClusterTemplateReference{
											Kind: "ClusterTemplate",
											Name: "my-template",
										},
									},
								},
							},
						}
						existingSupplyChainList := v1alpha1.ClusterSupplyChainList{
							Items: []v1alpha1.ClusterSupplyChain{existingSupplyChain1, existingSupplyChain2},
						}

						existingWorkload := v1alpha1.Workload{
							ObjectMeta: metav1.ObjectMeta{
								Name: "my-workload",
							},
						}
						existingWorkloadList := v1alpha1.WorkloadList{
							Items: []v1alpha1.Workload{existingWorkload},
						}

						fakeClient.ListStub = func(ctx context.Context, list client.ObjectList, option ...client.ListOption) error {
							listVal := reflect.Indirect(reflect.ValueOf(list))
							switch list.(type) {
							case *v1alpha1.ClusterSupplyChainList:
								existingVal := reflect.Indirect(reflect.ValueOf(existingSupplyChainList))
								listVal.Set(existingVal)
							case *v1alpha1.WorkloadList:
								existingVal := reflect.Indirect(reflect.ValueOf(existingWorkloadList))
								listVal.Set(existingVal)
							default:
								panic("list type not stubbed")
							}

							return nil
						}
					})

					Describe("The template refers to some supply chains", func() {
						It("returns requests for only the matching workloads", func() {
							t := &v1alpha1.ClusterTemplate{
								ObjectMeta: metav1.ObjectMeta{
									Name: "my-template",
								},
							}
							reqs := m.TemplateToWorkloadRequests(t)

							Expect(reqs).To(HaveLen(1))
							Expect(reqs[0].Name).To(Equal("my-workload"))
						})
					})

					Describe("The template does not reference a supply chain", func() {
						It("returns an empty request list", func() {
							t := &v1alpha1.ClusterTemplate{
								ObjectMeta: metav1.ObjectMeta{
									Name: "my-template-bar",
								},
							}
							reqs := m.TemplateToWorkloadRequests(t)

							Expect(reqs).To(HaveLen(0))
						})
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
					t := &v1alpha1.ClusterTemplate{}
					reqs := m.TemplateToWorkloadRequests(t)

					Expect(reqs).To(HaveLen(0))
					Expect(fakeLogger.ErrorCallCount()).To(Equal(1))

					err, msg, _ := fakeLogger.ErrorArgsForCall(0)
					Expect(err).To(Equal(listErr))
					Expect(msg).To(Equal("list ClusterSupplyChains"))
				})
			})
		})

	})

	Describe("ClusterSupplyChainToWorkloadRequests", func() {
		var (
			clientObjects      []client.Object
			mapper             *registrar.Mapper
			fakeClientBuilder  *fake.ClientBuilder
			scheme             *runtime.Scheme
			fakeLogger         *registrarfakes.FakeLogger
			clusterSupplyChain client.Object
			result             []reconcile.Request
		)

		BeforeEach(func() {
			scheme = runtime.NewScheme()
			fakeClientBuilder = fake.NewClientBuilder()
			fakeLogger = &registrarfakes.FakeLogger{}

			clusterSupplyChain = &v1alpha1.ClusterSupplyChain{
				Spec: v1alpha1.SupplyChainSpec{
					Selector: map[string]string{
						"myLabel": "myLabelsValue",
					},
				},
			}
		})

		JustBeforeEach(func() {
			fakeClientBuilder.
				WithScheme(scheme).
				WithObjects(clientObjects...)

			fakeClient := fakeClientBuilder.Build()

			mapper = &registrar.Mapper{
				Client: fakeClient,
				Logger: fakeLogger,
			}

			result = mapper.ClusterSupplyChainToWorkloadRequests(clusterSupplyChain)
		})

		Context("client.List returns an error", func() {
			// By using a scheme without v1alpha1, the client will error when handed our Objects
			It("logs an error to the client", func() {
				Expect(result).To(BeEmpty())

				Expect(fakeLogger.ErrorCallCount()).To(Equal(1))
				firstArg, secondArg, _ := fakeLogger.ErrorArgsForCall(0)
				Expect(firstArg).NotTo(BeNil())
				Expect(secondArg).To(Equal("cluster supply chain to workload requests: client list"))
			})
		})

		Context("client does not return errors", func() {
			BeforeEach(func() {
				// By including the scheme, the client will not error when handed our Objects
				err := v1alpha1.AddToScheme(scheme)
				Expect(err).ToNot(HaveOccurred())
			})
			Context("no workloads", func() {
				BeforeEach(func() {
					clusterSupplyChain = &v1alpha1.ClusterSupplyChain{
						Spec: v1alpha1.SupplyChainSpec{
							Selector: map[string]string{
								"myLabel": "myLabelsValue",
							},
						},
					}
				})
				It("returns an empty list of requests", func() {
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
				})

				Context("supply chain with one matching workload", func() {
					BeforeEach(func() {
						workload.Labels = map[string]string{
							"myLabel": "myLabelsValue",
						}
						clientObjects = []client.Object{workload}
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

						Expect(result).To(Equal(expected))
					})
				})
				Context("supply chain without matching workload", func() {
					BeforeEach(func() {
						workload.Labels = map[string]string{
							"myLabel": "otherLabel",
						}
						clientObjects = []client.Object{workload}
					})
					It("returns an empty list of requests", func() {
						Expect(result).To(BeEmpty())
					})
				})
			})

			Context("when function is passed an object that is not a supplyChain", func() {
				BeforeEach(func() {
					clusterSupplyChain = &v1alpha1.Workload{}
				})
				It("logs a helpful error", func() {
					Expect(result).To(BeEmpty())

					Expect(fakeLogger.ErrorCallCount()).To(Equal(1))
					firstArg, secondArg, _ := fakeLogger.ErrorArgsForCall(0)
					Expect(firstArg).To(BeNil())
					Expect(secondArg).To(Equal("cluster supply chain to workload requests: cast to ClusterSupplyChain failed"))
				})
			})
		})
	})

	Describe("ClusterDeliveryToDeliverableRequests", func() {
		var (
			clientObjects     []client.Object
			mapper            *registrar.Mapper
			fakeClientBuilder *fake.ClientBuilder
			scheme            *runtime.Scheme
			fakeLogger        *registrarfakes.FakeLogger
			clusterDelivery   client.Object
			result            []reconcile.Request
		)

		BeforeEach(func() {
			scheme = runtime.NewScheme()
			fakeClientBuilder = fake.NewClientBuilder()
			fakeLogger = &registrarfakes.FakeLogger{}

			clusterDelivery = &v1alpha1.ClusterDelivery{
				Spec: v1alpha1.ClusterDeliverySpec{
					Selector: map[string]string{
						"myLabel": "myLabelsValue",
					},
				},
			}
		})

		JustBeforeEach(func() {
			fakeClientBuilder.
				WithScheme(scheme).
				WithObjects(clientObjects...)

			fakeClient := fakeClientBuilder.Build()

			mapper = &registrar.Mapper{
				Client: fakeClient,
				Logger: fakeLogger,
			}

			result = mapper.ClusterDeliveryToDeliverableRequests(clusterDelivery)
		})

		Context("client.List returns an error", func() {
			// By using a scheme without v1alpha1, the client will error when handed our Objects
			It("logs an error to the client", func() {
				Expect(result).To(BeEmpty())

				Expect(fakeLogger.ErrorCallCount()).To(Equal(1))
				firstArg, secondArg, _ := fakeLogger.ErrorArgsForCall(0)
				Expect(firstArg).NotTo(BeNil())
				Expect(secondArg).To(Equal("cluster delivery to deliverable requests: client list"))
			})
		})

		Context("client does not return errors", func() {
			BeforeEach(func() {
				// By including the scheme, the client will not error when handed our Objects
				err := v1alpha1.AddToScheme(scheme)
				Expect(err).ToNot(HaveOccurred())
			})
			Context("no deliverables", func() {
				BeforeEach(func() {
					clusterDelivery = &v1alpha1.ClusterDelivery{
						Spec: v1alpha1.ClusterDeliverySpec{
							Selector: map[string]string{
								"myLabel": "myLabelsValue",
							},
						},
					}
				})
				It("returns an empty list of requests", func() {
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
				})

				Context("delivery with one matching deliverable", func() {
					BeforeEach(func() {
						deliverable.Labels = map[string]string{
							"myLabel": "myLabelsValue",
						}
						clientObjects = []client.Object{deliverable}
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

						Expect(result).To(Equal(expected))
					})
				})
				Context("delivery without matching deliverable", func() {
					BeforeEach(func() {
						deliverable.Labels = map[string]string{
							"myLabel": "otherLabel",
						}
						clientObjects = []client.Object{deliverable}
					})
					It("returns an empty list of requests", func() {
						Expect(result).To(BeEmpty())
					})
				})
			})

			Context("when function is passed an object that is not a supplyChain", func() {
				BeforeEach(func() {
					clusterDelivery = &v1alpha1.Workload{}
				})
				It("logs a helpful error", func() {
					Expect(result).To(BeEmpty())

					Expect(fakeLogger.ErrorCallCount()).To(Equal(1))
					firstArg, secondArg, _ := fakeLogger.ErrorArgsForCall(0)
					Expect(firstArg).To(BeNil())
					Expect(secondArg).To(Equal("cluster delivery to deliverable requests: cast to ClusterDelivery failed"))
				})
			})
		})
	})

	Describe("RunTemplateToRunnableRequests", func() {
		var (
			clientObjects     []client.Object
			mapper            *registrar.Mapper
			fakeClientBuilder *fake.ClientBuilder
			scheme            *runtime.Scheme
			fakeLogger        *registrarfakes.FakeLogger
			runTemplate       client.Object
			result            []reconcile.Request
		)

		BeforeEach(func() {
			scheme = runtime.NewScheme()
			fakeClientBuilder = fake.NewClientBuilder()
			fakeLogger = &registrarfakes.FakeLogger{}

			runTemplate = &v1alpha1.ClusterRunTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "match",
					Namespace: "match",
				},
			}
		})

		JustBeforeEach(func() {
			fakeClientBuilder.
				WithScheme(scheme).
				WithObjects(clientObjects...)

			fakeClient := fakeClientBuilder.Build()

			mapper = &registrar.Mapper{
				Client: fakeClient,
				Logger: fakeLogger,
			}

			result = mapper.RunTemplateToRunnableRequests(runTemplate)
		})

		Context("client.List returns an error", func() {
			// By using a scheme without v1alpha1, the client will error when handed our Objects
			It("logs an error to the client", func() {
				Expect(result).To(BeEmpty())

				Expect(fakeLogger.ErrorCallCount()).To(Equal(1))
				firstArg, secondArg, _ := fakeLogger.ErrorArgsForCall(0)
				Expect(firstArg).NotTo(BeNil())
				Expect(secondArg).To(Equal("run template to runnable requests: client list"))
			})
		})

		Context("client does not return errors", func() {
			BeforeEach(func() {
				// By including the scheme, the client will not error when handed our Objects
				err := v1alpha1.AddToScheme(scheme)
				Expect(err).ToNot(HaveOccurred())
			})
			Context("but there exist no runnables", func() {
				It("returns an empty list of requests", func() {
					Expect(result).To(BeEmpty())
				})
			})
			Context("and there are runnables", func() {
				var runnable *v1alpha1.Runnable
				BeforeEach(func() {
					runnable = &v1alpha1.Runnable{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "my-runnable",
							Namespace: "my-namespace",
						},
						TypeMeta: metav1.TypeMeta{},
					}
				})

				Context("a runnable matches the runTemplate", func() {
					Context("with a templateRef that specifies a namespace", func() {
						BeforeEach(func() {
							runnable.Spec.RunTemplateRef = v1alpha1.TemplateReference{
								Name: "match",
							}
							clientObjects = []client.Object{runnable}
						})

						It("returns a list of requests with the runnable present", func() {
							expected := []reconcile.Request{
								{
									types.NamespacedName{
										Namespace: "my-namespace",
										Name:      "my-runnable",
									},
								},
							}

							Expect(result).To(Equal(expected))
						})
					})

					Context("with a templateRef that specifies a namespace", func() {
						BeforeEach(func() {
							runnable.Spec.RunTemplateRef = v1alpha1.TemplateReference{
								Name: "match",
							}
							runnable.Namespace = "match"
							clientObjects = []client.Object{runnable}
						})

						It("returns a list of requests with the runnable present", func() {
							expected := []reconcile.Request{
								{
									types.NamespacedName{
										Namespace: "match",
										Name:      "my-runnable",
									},
								},
							}

							Expect(result).To(Equal(expected))
						})
					})
				})
				Context("no runnable matches the runTemplate", func() {
					Context("because the name in the templateRef is different", func() {
						BeforeEach(func() {
							runnable.Spec.RunTemplateRef = v1alpha1.TemplateReference{
								Name: "non-existent-name",
							}
							clientObjects = []client.Object{runnable}
						})

						It("returns an empty list of requests", func() {
							Expect(result).To(BeEmpty())
						})
					})

					Context("because the templateRef is the wrong Kind", func() {
						BeforeEach(func() {
							runnable.Spec.RunTemplateRef = v1alpha1.TemplateReference{
								Name: "match",
								Kind: "some-kind",
							}
							clientObjects = []client.Object{runnable}
						})

						It("returns an empty list of requests", func() {
							Expect(result).To(BeEmpty())
						})
					})
				})
			})

			Context("when function is passed an object that is not a supplyChain", func() {
				BeforeEach(func() {
					runTemplate = &v1alpha1.Workload{}
				})
				It("logs a helpful error", func() {
					Expect(result).To(BeEmpty())

					Expect(fakeLogger.ErrorCallCount()).To(Equal(1))
					firstArg, secondArg, _ := fakeLogger.ErrorArgsForCall(0)
					Expect(firstArg).To(BeNil())
					Expect(secondArg).To(Equal("run template to runnable requests: cast to run template failed"))
				})
			})
		})
	})

	Describe("TemplateToSupplyChainRequests", func() {

		var (
			m          *registrar.Mapper
			fakeLogger *registrarfakes.FakeLogger
			fakeClient *registrarfakes.FakeClient
		)

		Context("the template kind can be found", func() {
			BeforeEach(func() {
				fakeLogger = &registrarfakes.FakeLogger{}
				fakeClient = &registrarfakes.FakeClient{}

				m = &registrar.Mapper{
					Client: fakeClient,
					Logger: fakeLogger,
				}

				scheme := runtime.NewScheme()
				err := v1alpha1.AddToScheme(scheme)
				Expect(err).NotTo(HaveOccurred())

				fakeClient.SchemeReturns(scheme)
			})

			Context("client.list does not return errors", func() {

				Context("there are no SupplyChains", func() {
					BeforeEach(func() {
						existingList := v1alpha1.ClusterSupplyChainList{
							Items: []v1alpha1.ClusterSupplyChain{},
						}

						fakeClient.ListStub = func(ctx context.Context, list client.ObjectList, option ...client.ListOption) error {
							listVal := reflect.ValueOf(list)
							existingVal := reflect.ValueOf(existingList)

							reflect.Indirect(listVal).Set(reflect.Indirect(existingVal))
							return nil
						}
					})

					It("returns an empty request list", func() {
						t := &v1alpha1.ClusterTemplate{}
						reqs := m.TemplateToSupplyChainRequests(t)

						Expect(reqs).To(HaveLen(0))
					})
				})

				Context("there are multiple supply chains", func() {
					BeforeEach(func() {
						existingSupplyChain1 := &v1alpha1.ClusterSupplyChain{
							Spec: v1alpha1.SupplyChainSpec{
								Resources: []v1alpha1.SupplyChainResource{
									{
										TemplateRef: v1alpha1.ClusterTemplateReference{
											Kind: "ClusterTemplate",
											Name: "my-template-foo",
										},
									},
								},
							},
						}
						existingSupplyChain2 := &v1alpha1.ClusterSupplyChain{
							ObjectMeta: metav1.ObjectMeta{Name: "good-supply-chain"},
							Spec: v1alpha1.SupplyChainSpec{
								Resources: []v1alpha1.SupplyChainResource{
									{
										TemplateRef: v1alpha1.ClusterTemplateReference{
											Kind: "ClusterTemplate",
											Name: "my-template",
										},
									},
								},
							},
						}
						existingList := v1alpha1.ClusterSupplyChainList{
							Items: []v1alpha1.ClusterSupplyChain{*existingSupplyChain1, *existingSupplyChain2},
						}

						fakeClient.ListStub = func(ctx context.Context, list client.ObjectList, option ...client.ListOption) error {
							listVal := reflect.ValueOf(list)
							existingVal := reflect.ValueOf(existingList)

							reflect.Indirect(listVal).Set(reflect.Indirect(existingVal))
							return nil
						}
					})

					Describe("The template refers to some supply chains", func() {
						It("returns requests for only the matching supply chains", func() {
							t := &v1alpha1.ClusterTemplate{
								ObjectMeta: metav1.ObjectMeta{
									Name: "my-template",
								},
							}
							reqs := m.TemplateToSupplyChainRequests(t)

							Expect(reqs).To(HaveLen(1))
							Expect(reqs[0].Name).To(Equal("good-supply-chain"))
						})
					})

					Describe("The template does not reference a supply chain", func() {
						It("returns an empty request list", func() {
							t := &v1alpha1.ClusterTemplate{
								ObjectMeta: metav1.ObjectMeta{
									Name: "my-template-bar",
								},
							}
							reqs := m.TemplateToSupplyChainRequests(t)

							Expect(reqs).To(HaveLen(0))
						})
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
					t := &v1alpha1.ClusterTemplate{}
					reqs := m.TemplateToSupplyChainRequests(t)

					Expect(reqs).To(HaveLen(0))
					Expect(fakeLogger.ErrorCallCount()).To(Equal(1))

					err, msg, _ := fakeLogger.ErrorArgsForCall(0)
					Expect(err).To(Equal(listErr))
					Expect(msg).To(Equal("list ClusterSupplyChains"))
				})
			})
		})

		Context("the template kind cannot be found", func() {
			BeforeEach(func() {
				fakeLogger = &registrarfakes.FakeLogger{}
				fakeClient = &registrarfakes.FakeClient{}

				m = &registrar.Mapper{
					Client: fakeClient,
					Logger: fakeLogger,
				}

				// empty scheme causes the error
				fakeClient.SchemeReturns(&runtime.Scheme{})
			})

			It("returns an error", func() {
				t := &v1alpha1.ClusterTemplate{
					ObjectMeta: metav1.ObjectMeta{
						Name: "my-template",
					},
				}
				reqs := m.TemplateToSupplyChainRequests(t)

				Expect(reqs).To(HaveLen(0))

				Expect(fakeLogger.ErrorCallCount()).To(Equal(1))
				err, msg, _ := fakeLogger.ErrorArgsForCall(0)
				Expect(err).To(MatchError(ContainSubstring("missing apiVersion or kind: my-template")))
				Expect(msg).To(Equal("could not get GVK for template: my-template"))
			})
		})
	})

	Describe("TemplateToDeliveryRequests", func() {

		var (
			m          *registrar.Mapper
			fakeLogger *registrarfakes.FakeLogger
			fakeClient *registrarfakes.FakeClient
		)

		Context("the template kind can be found", func() {
			BeforeEach(func() {
				fakeLogger = &registrarfakes.FakeLogger{}
				fakeClient = &registrarfakes.FakeClient{}

				m = &registrar.Mapper{
					Client: fakeClient,
					Logger: fakeLogger,
				}

				scheme := runtime.NewScheme()
				err := v1alpha1.AddToScheme(scheme)
				Expect(err).NotTo(HaveOccurred())

				fakeClient.SchemeReturns(scheme)
			})

			Context("client.list does not return errors", func() {

				Context("there are no Deliveries", func() {
					BeforeEach(func() {
						existingList := v1alpha1.ClusterDeliveryList{
							Items: []v1alpha1.ClusterDelivery{},
						}

						fakeClient.ListStub = func(ctx context.Context, list client.ObjectList, option ...client.ListOption) error {
							listVal := reflect.ValueOf(list)
							existingVal := reflect.ValueOf(existingList)

							reflect.Indirect(listVal).Set(reflect.Indirect(existingVal))
							return nil
						}
					})

					It("returns an empty request list", func() {
						t := &v1alpha1.ClusterTemplate{}
						reqs := m.TemplateToDeliveryRequests(t)

						Expect(reqs).To(HaveLen(0))
					})
				})

				Context("there are multiple deliveries", func() {
					BeforeEach(func() {
						existingDelivery1 := &v1alpha1.ClusterDelivery{
							Spec: v1alpha1.ClusterDeliverySpec{
								Resources: []v1alpha1.ClusterDeliveryResource{
									{
										TemplateRef: v1alpha1.DeliveryClusterTemplateReference{
											Kind: "ClusterTemplate",
											Name: "my-template-foo",
										},
									},
								},
							},
						}
						existingDelivery2 := &v1alpha1.ClusterDelivery{
							ObjectMeta: metav1.ObjectMeta{Name: "good-delivery"},
							Spec: v1alpha1.ClusterDeliverySpec{
								Resources: []v1alpha1.ClusterDeliveryResource{
									{
										TemplateRef: v1alpha1.DeliveryClusterTemplateReference{
											Kind: "ClusterTemplate",
											Name: "my-template",
										},
									},
								},
							},
						}
						existingList := v1alpha1.ClusterDeliveryList{
							Items: []v1alpha1.ClusterDelivery{*existingDelivery1, *existingDelivery2},
						}

						fakeClient.ListStub = func(ctx context.Context, list client.ObjectList, option ...client.ListOption) error {
							listVal := reflect.ValueOf(list)
							existingVal := reflect.ValueOf(existingList)

							reflect.Indirect(listVal).Set(reflect.Indirect(existingVal))
							return nil
						}
					})

					Describe("The template refers to some deliveries", func() {
						It("returns requests for only the matching deliveries", func() {
							t := &v1alpha1.ClusterTemplate{
								ObjectMeta: metav1.ObjectMeta{
									Name: "my-template",
								},
							}
							reqs := m.TemplateToDeliveryRequests(t)

							Expect(reqs).To(HaveLen(1))
							Expect(reqs[0].Name).To(Equal("good-delivery"))
						})
					})

					Describe("The template does not reference a delivery", func() {
						It("returns an empty request list", func() {
							t := &v1alpha1.ClusterTemplate{
								ObjectMeta: metav1.ObjectMeta{
									Name: "my-template-bar",
								},
							}
							reqs := m.TemplateToDeliveryRequests(t)

							Expect(reqs).To(HaveLen(0))
						})
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
					t := &v1alpha1.ClusterTemplate{}
					reqs := m.TemplateToDeliveryRequests(t)

					Expect(reqs).To(HaveLen(0))
					Expect(fakeLogger.ErrorCallCount()).To(Equal(1))

					err, msg, _ := fakeLogger.ErrorArgsForCall(0)
					Expect(err).To(Equal(listErr))
					Expect(msg).To(Equal("list ClusterDeliveries"))
				})
			})
		})

		Context("the template kind cannot be found", func() {
			BeforeEach(func() {
				fakeLogger = &registrarfakes.FakeLogger{}
				fakeClient = &registrarfakes.FakeClient{}

				m = &registrar.Mapper{
					Client: fakeClient,
					Logger: fakeLogger,
				}

				// empty scheme causes the error
				fakeClient.SchemeReturns(&runtime.Scheme{})
			})

			It("returns an error", func() {
				t := &v1alpha1.ClusterTemplate{
					ObjectMeta: metav1.ObjectMeta{
						Name: "my-template",
					},
				}
				reqs := m.TemplateToDeliveryRequests(t)

				Expect(reqs).To(HaveLen(0))

				Expect(fakeLogger.ErrorCallCount()).To(Equal(1))
				err, msg, _ := fakeLogger.ErrorArgsForCall(0)
				Expect(err).To(MatchError(ContainSubstring("missing apiVersion or kind: my-template")))
				Expect(msg).To(Equal("could not get GVK for template: my-template"))
			})
		})
	})
})
