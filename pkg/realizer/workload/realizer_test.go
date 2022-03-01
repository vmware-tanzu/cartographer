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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/workload"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/workload/workloadfakes"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

var _ = Describe("Realize", func() {
	var (
		resourceRealizer      *workloadfakes.FakeResourceRealizer
		supplyChain           *v1alpha1.ClusterSupplyChain
		resource1             v1alpha1.SupplyChainResource
		resource2             v1alpha1.SupplyChainResource
		rlzr                  realizer.Realizer
		template1             *v1alpha1.ClusterImageTemplate
		template2             *v1alpha1.ClusterTemplate
		executedResourceOrder []string
	)
	BeforeEach(func() {
		rlzr = realizer.NewRealizer()

		resourceRealizer = &workloadfakes.FakeResourceRealizer{}
		resource1 = v1alpha1.SupplyChainResource{
			Name: "resource1",
		}
		resource2 = v1alpha1.SupplyChainResource{
			Name: "resource2",
			Images: []v1alpha1.ResourceReference{
				{
					Name:     "my-image",
					Resource: "resource1",
				},
			},
		}
		template1 = &v1alpha1.ClusterImageTemplate{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: "my-image-template",
			},
		}
		template2 = &v1alpha1.ClusterTemplate{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: "my-cluster-template",
			},
		}
		supplyChain = &v1alpha1.ClusterSupplyChain{
			ObjectMeta: metav1.ObjectMeta{Name: "greatest-supply-chain"},
			Spec: v1alpha1.SupplyChainSpec{
				Resources: []v1alpha1.SupplyChainResource{resource1, resource2},
			},
		}

		outputFromFirstResource := &templates.Output{Image: "whatever"}

		resourceRealizer.DoCalls(func(ctx context.Context, resource *v1alpha1.SupplyChainResource, supplyChainName string, outputs realizer.Outputs) (templates.Template, *unstructured.Unstructured, *templates.Output, error) {
			executedResourceOrder = append(executedResourceOrder, resource.Name)
			Expect(supplyChainName).To(Equal("greatest-supply-chain"))
			if resource.Name == "resource1" {
				Expect(outputs).To(Equal(realizer.NewOutputs()))
				template, err := templates.NewModelFromAPI(template1)
				Expect(err).NotTo(HaveOccurred())
				stampedObj := &unstructured.Unstructured{}
				stampedObj.SetName("obj1")
				return template, stampedObj, outputFromFirstResource, nil
			}

			if resource.Name == "resource2" {
				expectedSecondResourceOutputs := realizer.NewOutputs()
				expectedSecondResourceOutputs.AddOutput("resource1", outputFromFirstResource)
				Expect(outputs).To(Equal(expectedSecondResourceOutputs))
			}
			template, err := templates.NewModelFromAPI(template2)
			Expect(err).NotTo(HaveOccurred())
			stampedObj := &unstructured.Unstructured{}
			stampedObj.SetName("obj2")
			return template, stampedObj, &templates.Output{}, nil
		})

	})

	It("realizes each resource in supply chain order, accumulating output for each subsequent resource", func() {
		realizedResources, err := rlzr.Realize(context.TODO(), resourceRealizer, supplyChain, nil)
		Expect(err).ToNot(HaveOccurred())

		Expect(executedResourceOrder).To(Equal([]string{"resource1", "resource2"}))

		Expect(realizedResources).To(HaveLen(2))

		Expect(realizedResources[0].Name).To(Equal(resource1.Name))
		Expect(realizedResources[0].TemplateRef.Name).To(Equal(template1.Name))
		Expect(realizedResources[0].StampedRef.Name).To(Equal("obj1"))
		Expect(realizedResources[0].Inputs).To(BeNil())
		Expect(len(realizedResources[0].Outputs)).To(Equal(1))
		Expect(realizedResources[0].Outputs[0]).To(MatchFields(IgnoreExtras,
			Fields{
				"Name":    Equal("image"),
				"Preview": Equal("whatever"),
				"Digest":  Equal("sha256:85738f8f9a7f1b04b5329c590ebcb9e425925c6d0984089c43a022de4f19c281"),
			},
		))
		Expect(time.Since(realizedResources[0].Outputs[0].LastTransitionTime.Time)).To(BeNumerically("<", time.Second))

		Expect(realizedResources[1].Name).To(Equal(resource2.Name))
		Expect(realizedResources[1].TemplateRef.Name).To(Equal(template2.Name))
		Expect(realizedResources[1].StampedRef.Name).To(Equal("obj2"))
		Expect(len(realizedResources[1].Inputs)).To(Equal(1))
		Expect(realizedResources[1].Inputs).To(Equal([]v1alpha1.Input{{Name: "resource1"}}))
		Expect(realizedResources[1].Outputs).To(BeNil())
	})

	It("returns the first error encountered realizing a resource and continues to realize", func() {
		template, err := templates.NewModelFromAPI(template2)
		Expect(err).NotTo(HaveOccurred())
		resourceRealizer.DoReturnsOnCall(0, nil, nil, nil, errors.New("realizing is hard"))
		resourceRealizer.DoReturnsOnCall(1, template, &unstructured.Unstructured{}, nil, nil)

		realizedResources, err := rlzr.Realize(context.TODO(), resourceRealizer, supplyChain, nil)
		Expect(err).To(MatchError("realizing is hard"))
		Expect(realizedResources).To(HaveLen(2))

		Expect(realizedResources[0].TemplateRef).To(BeNil())
		Expect(realizedResources[0].StampedRef).To(BeNil())
		Expect(realizedResources[1].TemplateRef.Name).To(Equal(template2.Name))
	})

	It("realizes each resource and does not update last transition time since the resource has not changed", func() {
		previousTime := metav1.NewTime(time.Now())
		previousResources := []v1alpha1.RealizedResource{
			{
				Name:        "resource1",
				StampedRef:  nil,
				TemplateRef: nil,
				Inputs:      nil,
				Outputs: []v1alpha1.Output{
					{
						Name:               "image",
						Preview:            "whatever",
						Digest:             "sha256:85738f8f9a7f1b04b5329c590ebcb9e425925c6d0984089c43a022de4f19c281",
						LastTransitionTime: previousTime,
					},
				},
			},
		}
		realizedResources, err := rlzr.Realize(context.TODO(), resourceRealizer, supplyChain, previousResources)
		Expect(err).ToNot(HaveOccurred())

		Expect(len(realizedResources[0].Outputs)).To(Equal(1))
		Expect(realizedResources[0].Outputs[0]).To(MatchFields(IgnoreExtras,
			Fields{
				"Name":    Equal("image"),
				"Preview": Equal("whatever"),
				"Digest":  Equal("sha256:85738f8f9a7f1b04b5329c590ebcb9e425925c6d0984089c43a022de4f19c281"),
			},
		))
		Expect(realizedResources[0].Outputs[0].LastTransitionTime).To(Equal(previousTime))
	})
})
