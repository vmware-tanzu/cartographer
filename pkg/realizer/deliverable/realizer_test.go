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

package deliverable_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/deliverable"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/deliverable/deliverablefakes"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

var _ = Describe("Realize", func() {
	var (
		resourceRealizer *deliverablefakes.FakeResourceRealizer
		delivery         *v1alpha1.ClusterDelivery
		resource1        v1alpha1.DeliveryResource
		resource2        v1alpha1.DeliveryResource
		rlzr             realizer.Realizer
		ctx              context.Context
	)
	BeforeEach(func() {
		ctx = context.Background()

		rlzr = realizer.NewRealizer()

		resourceRealizer = &deliverablefakes.FakeResourceRealizer{}
		resource1 = v1alpha1.DeliveryResource{
			Name: "resource1",
		}
		resource2 = v1alpha1.DeliveryResource{
			Name: "resource2",
		}
		delivery = &v1alpha1.ClusterDelivery{
			ObjectMeta: metav1.ObjectMeta{Name: "greatest-delivery"},
			Spec: v1alpha1.DeliverySpec{
				Resources: []v1alpha1.DeliveryResource{resource1, resource2},
			},
		}
	})

	It("realizes each resource in delivery order, accumulating output for each subsequent resource", func() {
		outputFromFirstResource := &templates.Output{Config: "whatever"}

		var executedResourceOrder []string

		resourceRealizer.DoCalls(func(ctx context.Context, resource *v1alpha1.DeliveryResource, deliveryName string, outputs realizer.Outputs) (*unstructured.Unstructured, *templates.Output, error) {
			executedResourceOrder = append(executedResourceOrder, resource.Name)
			Expect(deliveryName).To(Equal("greatest-delivery"))
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

		stampedObjects, err := rlzr.Realize(ctx, resourceRealizer, delivery)
		Expect(err).To(Succeed())

		Expect(executedResourceOrder).To(Equal([]string{"resource1", "resource2"}))

		Expect(stampedObjects).To(HaveLen(2))
	})

	It("returns the first error encountered realizing a resource and continues to realize", func() {
		resourceRealizer.DoReturnsOnCall(0, nil, nil, errors.New("realizing is hard"))
		resourceRealizer.DoReturnsOnCall(1, &unstructured.Unstructured{}, nil, nil)

		stampedObjects, err := rlzr.Realize(ctx, resourceRealizer, delivery)
		Expect(err).To(MatchError("realizing is hard"))
		Expect(stampedObjects).To(HaveLen(1))
	})
})
