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

package supplychain_test

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/conditions/conditionsfakes"
	"github.com/vmware-tanzu/cartographer/pkg/controller/supplychain"
	"github.com/vmware-tanzu/cartographer/pkg/registrar"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
)

var _ = Describe("Reconciler", func() {
	var (
		out                *Buffer
		reconciler         supplychain.Reconciler
		ctx                context.Context
		req                reconcile.Request
		conditionManager   *conditionsfakes.FakeConditionManager
		repo               *repositoryfakes.FakeRepository
		sc                 *v1alpha1.ClusterSupplyChain
		expectedConditions []metav1.Condition
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

		expectedConditions = []metav1.Condition{{
			Type:               "Happy",
			Status:             "True",
			ObservedGeneration: 1,
			LastTransitionTime: metav1.Time{},
			Reason:             "Because I'm",
			Message:            "Clap Along If you Feel",
		}}
		conditionManager.FinalizeReturns(expectedConditions, true)

		repo = &repositoryfakes.FakeRepository{}

		sc = &v1alpha1.ClusterSupplyChain{
			ObjectMeta: metav1.ObjectMeta{
				Generation: 1,
			},
			Spec: v1alpha1.SupplyChainSpec{
				Resources: []v1alpha1.SupplyChainResource{
					{
						Name: "first name",
						TemplateRef: v1alpha1.ClusterTemplateReference{
							Kind: "some-kind",
							Name: "some-name",
						},
					},
					{
						Name: "second name",
						TemplateRef: v1alpha1.ClusterTemplateReference{
							Kind: "another-kind",
							Name: "another-name",
						},
					},
				},
				Selector: map[string]string{},
			},
		}

		repo.GetSupplyChainReturns(sc, nil)

		scheme := runtime.NewScheme()
		err := registrar.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())
		repo.GetSchemeReturns(scheme)

		reconciler = supplychain.Reconciler{
			Repo:                    repo,
			ConditionManagerBuilder: fakeConditionManagerBuilder,
		}

		req = reconcile.Request{
			NamespacedName: types.NamespacedName{Name: "my-supply-chain", Namespace: "my-namespace"},
		}

		repo.GetClusterTemplateReturns(&v1alpha1.ClusterTemplate{}, nil)
	})

	It("logs that it's begun", func() {
		_, _ = reconciler.Reconcile(ctx, req)

		Expect(out).To(Say(`"msg":"started"`))
	})

	It("logs that it's finished", func() {
		_, _ = reconciler.Reconcile(ctx, req)

		Expect(out).To(Say(`"msg":"finished"`))
	})

	It("updates the status of the supply chain", func() {
		_, _ = reconciler.Reconcile(ctx, req)

		Expect(repo.StatusUpdateCallCount()).To(Equal(1))
	})

	It("updates the status.observedGeneration to equal metadata.generation", func() {
		_, _ = reconciler.Reconcile(ctx, req)

		_, updatedSupplyChain := repo.StatusUpdateArgsForCall(0)

		Expect(*updatedSupplyChain.(*v1alpha1.ClusterSupplyChain)).To(MatchFields(IgnoreExtras, Fields{
			"Status": MatchFields(IgnoreExtras, Fields{
				"ObservedGeneration": BeEquivalentTo(1),
			}),
		}))
	})

	It("updates the conditions based on the output of the conditionManager", func() {
		_, _ = reconciler.Reconcile(ctx, req)

		_, updatedSupplyChain := repo.StatusUpdateArgsForCall(0)

		Expect(*updatedSupplyChain.(*v1alpha1.ClusterSupplyChain)).To(MatchFields(IgnoreExtras, Fields{
			"Status": MatchFields(IgnoreExtras, Fields{
				"Conditions": Equal(expectedConditions),
			}),
		}))
	})

	It("adds a positive templates found condition", func() {
		_, _ = reconciler.Reconcile(ctx, req)

		Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(supplychain.TemplatesFoundCondition()))
	})

	It("does not return an error", func() {
		_, err := reconciler.Reconcile(ctx, req)

		Expect(err).NotTo(HaveOccurred())
	})

	Context("get cluster template fails", func() {
		BeforeEach(func() {
			repo.GetClusterTemplateReturnsOnCall(0, nil, errors.New("getting templates is hard"))
		})

		It("returns an unhandled error and requeues", func() {
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).To(MatchError(ContainSubstring("getting templates is hard")))
		})
	})

	Context("cannot find cluster template", func() {
		BeforeEach(func() {
			repo.GetClusterTemplateReturnsOnCall(0, &v1alpha1.ClusterTemplate{}, nil)
			repo.GetClusterTemplateReturnsOnCall(1, nil, nil)
		})

		It("adds a positive templates NOT found condition", func() {
			_, _ = reconciler.Reconcile(ctx, req)
			Expect(conditionManager.AddPositiveArgsForCall(0)).To(Equal(supplychain.TemplatesNotFoundCondition([]string{"second name"})))
		})

		It("does not return an error", func() {
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when the update fails", func() {
		BeforeEach(func() {
			repo.StatusUpdateReturns(errors.New("updating is hard"))
		})

		It("returns an error and requeues", func() {
			_, err := reconciler.Reconcile(ctx, req)

			Expect(err).To(MatchError(ContainSubstring("failed to update status for supply chain")))
			Expect(err).To(MatchError(ContainSubstring("updating is hard")))
		})
	})

	Context("when the supply chain has been deleted from the apiServer", func() {
		BeforeEach(func() {
			repo.GetSupplyChainReturns(nil, nil)
		})

		It("does not return an error", func() {
			_, err := reconciler.Reconcile(ctx, req)

			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when the client errors", func() {
		BeforeEach(func() {
			repo.GetSupplyChainReturns(nil, errors.New("some error"))
		})

		It("returns an error and requeues", func() {
			_, err := reconciler.Reconcile(ctx, req)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to get supply chain [my-namespace/my-supply-chain]: some error"))
			Expect(err.Error()).To(ContainSubstring("some error"))
		})
	})

})
