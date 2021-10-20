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

package supply_chain_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

var _ = Describe("SupplyChainReconciler", func() {
	var newClusterSupplyChain = func(name string, selector map[string]string) *v1alpha1.ClusterSupplyChain {
		return &v1alpha1.ClusterSupplyChain{
			TypeMeta: v1.TypeMeta{},
			ObjectMeta: v1.ObjectMeta{
				Name: name,
			},
			Spec: v1alpha1.SupplyChainSpec{
				Resources: []v1alpha1.SupplyChainResource{},
				Selector:  selector,
			},
		}
	}

	var reconcileAgain = func() {
		time.Sleep(1 * time.Second) //metav1.Time unmarshals with 1 second accuracy so this sleep avoids a race condition

		supplyChain := &v1alpha1.ClusterSupplyChain{}
		err := c.Get(context.Background(), client.ObjectKey{Name: "supplychain-bob"}, supplyChain)
		Expect(err).NotTo(HaveOccurred())

		supplyChain.Spec.Selector = map[string]string{"blah": "blah"}
		err = c.Update(context.Background(), supplyChain)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() int64 {
			supplyChain := &v1alpha1.ClusterSupplyChain{}
			err := c.Get(context.Background(), client.ObjectKey{Name: "supplychain-bob"}, supplyChain)
			Expect(err).NotTo(HaveOccurred())
			return supplyChain.Status.ObservedGeneration
		}).Should(Equal(supplyChain.Generation))
	}

	var (
		ctx      context.Context
		cleanups []client.Object
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	AfterEach(func() {
		for _, obj := range cleanups {
			_ = c.Delete(ctx, obj, &client.DeleteOptions{})
		}
	})

	Context("when reconciling a supply chain", func() {
		var (
			lastConditions []v1.Condition
		)
		BeforeEach(func() {
			supplyChain := newClusterSupplyChain("supplychain-bob", map[string]string{"name": "webapp"})
			err := c.Create(ctx, supplyChain, &client.CreateOptions{})
			cleanups = append(cleanups, supplyChain)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() bool {
				supplyChain := &v1alpha1.ClusterSupplyChain{}
				err := c.Get(ctx, client.ObjectKey{Name: "supplychain-bob"}, supplyChain)
				Expect(err).NotTo(HaveOccurred())
				lastConditions = supplyChain.Status.Conditions

				return supplyChain.Status.ObservedGeneration == supplyChain.Generation
			}, 5*time.Second).Should(BeTrue())
		})

		It("does not update the lastTransitionTime on subsequent reconciliation if the status does not change", func() {
			reconcileAgain()

			supplyChain := &v1alpha1.ClusterSupplyChain{}
			err := c.Get(ctx, client.ObjectKey{Name: "supplychain-bob"}, supplyChain)
			Expect(err).NotTo(HaveOccurred())
			Expect(supplyChain.Status.Conditions).To(Equal(lastConditions))
		})
	})
})
