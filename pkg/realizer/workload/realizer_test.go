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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/workload"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/workload/workloadfakes"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

var _ = Describe("Realize", func() {
	var (
		componentRealizer *workloadfakes.FakeComponentRealizer
		supplyChain       *v1alpha1.ClusterSupplyChain
		component1        v1alpha1.SupplyChainComponent
		component2        v1alpha1.SupplyChainComponent
		rlzr              realizer.Realizer
	)
	BeforeEach(func() {
		rlzr = realizer.NewRealizer()

		componentRealizer = &workloadfakes.FakeComponentRealizer{}
		component1 = v1alpha1.SupplyChainComponent{
			Name: "component1",
		}
		component2 = v1alpha1.SupplyChainComponent{
			Name: "component2",
		}
		supplyChain = &v1alpha1.ClusterSupplyChain{
			ObjectMeta: metav1.ObjectMeta{Name: "greatest-supply-chain"},
			Spec: v1alpha1.SupplyChainSpec{
				Components: []v1alpha1.SupplyChainComponent{component1, component2},
			},
		}
	})

	It("realizes each component in supply chain order, accumulating output for each subsequent component", func() {
		outputFromFirstComponent := &templates.Output{Image: "whatever"}

		var executedComponentOrder []string

		componentRealizer.DoCalls(func(ctx context.Context, component *v1alpha1.SupplyChainComponent, supplyChainName string, outputs realizer.Outputs) (*templates.Output, error) {
			executedComponentOrder = append(executedComponentOrder, component.Name)
			Expect(supplyChainName).To(Equal("greatest-supply-chain"))
			if component.Name == "component1" {
				Expect(outputs).To(Equal(realizer.NewOutputs()))
				return outputFromFirstComponent, nil
			}

			if component.Name == "component2" {
				expectedSecondComponentOutputs := realizer.NewOutputs()
				expectedSecondComponentOutputs.AddOutput("component1", outputFromFirstComponent)
				Expect(outputs).To(Equal(expectedSecondComponentOutputs))
			}

			return &templates.Output{}, nil
		})

		Expect(rlzr.Realize(context.TODO(), componentRealizer, supplyChain)).To(Succeed())

		Expect(executedComponentOrder).To(Equal([]string{"component1", "component2"}))
	})

	It("returns any error encountered realizing a component", func() {
		componentRealizer.DoReturns(nil, errors.New("realizing is hard"))
		Expect(rlzr.Realize(context.TODO(), componentRealizer, supplyChain)).To(MatchError("realizing is hard"))
	})
})
