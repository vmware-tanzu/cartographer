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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/workload"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/workload/workloadfakes"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

var _ = Describe("Realize", func() {
	var (
		resourceRealizer *workloadfakes.FakeResourceRealizer
		supplyChain      *v1alpha1.ClusterSupplyChain
		resource1        v1alpha1.SupplyChainResource
		resource2        v1alpha1.SupplyChainResource
		rlzr             realizer.Realizer
	)
	BeforeEach(func() {
		rlzr = realizer.NewRealizer()

		resourceRealizer = &workloadfakes.FakeResourceRealizer{}
		resource1 = v1alpha1.SupplyChainResource{
			Name: "resource1",
		}
		resource2 = v1alpha1.SupplyChainResource{
			Name: "resource2",
		}
		supplyChain = &v1alpha1.ClusterSupplyChain{
			ObjectMeta: metav1.ObjectMeta{Name: "greatest-supply-chain"},
			Spec: v1alpha1.SupplyChainSpec{
				Resources: []v1alpha1.SupplyChainResource{resource1, resource2},
			},
		}
	})

	It("realizes each resource in supply chain order, accumulating output for each subsequent resource", func() {
		outputFromFirstResource := &templates.Output{Image: "whatever"}

		var executedResourceOrder []string

		resourceRealizer.DoCalls(func(ctx context.Context, resource *v1alpha1.SupplyChainResource, supplyChainName string, outputs realizer.Outputs) (*unstructured.Unstructured, *templates.Output, error) {
			executedResourceOrder = append(executedResourceOrder, resource.Name)
			Expect(supplyChainName).To(Equal("greatest-supply-chain"))
			if resource.Name == "resource1" {
				Expect(outputs).To(Equal(realizer.NewOutputs()))
				return &unstructured.Unstructured{}, outputFromFirstResource, nil
			}

			if resource.Name == "resource2" {
				expectedSecondResourceOutputs := realizer.NewOutputs()
				expectedSecondResourceOutputs.AddOutput("resource1", outputFromFirstResource)
				Expect(outputs).To(Equal(expectedSecondResourceOutputs))
			}

			return &unstructured.Unstructured{}, &templates.Output{}, nil
		})

		stampedObjects, err := rlzr.Realize(context.TODO(), resourceRealizer, supplyChain)
		Expect(err).ToNot(HaveOccurred())

		Expect(executedResourceOrder).To(Equal([]string{"resource1", "resource2"}))

		Expect(stampedObjects).To(HaveLen(2))
	})

	It("returns any error encountered realizing a resource", func() {
		resourceRealizer.DoReturns(nil, nil, errors.New("realizing is hard"))
		stampedObjects, err := rlzr.Realize(context.TODO(), resourceRealizer, supplyChain)
		Expect(err).To(MatchError("realizing is hard"))
		Expect(stampedObjects).To(HaveLen(0))
	})
})
