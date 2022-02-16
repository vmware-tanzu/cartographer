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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

var _ = Describe("SupplyChainReconciler", func() {
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

	Context("when reconciling a supply chain with template references", func() {
		BeforeEach(func() {
			supplyChainYaml := utils.HereYaml(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterSupplyChain
				metadata:
				  name: my-supply-chain
				spec:
				  selector:
					"some-key": "some-value"
			      resources:
			        - name: my-first-resource
					  templateRef:
				        kind: ClusterTemplate
				        name: my-terminal-template
			`)

			supplyChain := &unstructured.Unstructured{}
			err := yaml.Unmarshal([]byte(supplyChainYaml), supplyChain)
			Expect(err).NotTo(HaveOccurred())

			err = c.Create(ctx, supplyChain, &client.CreateOptions{})
			cleanups = append(cleanups, supplyChain)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() []v1.Condition {
				supplyChain := &v1alpha1.ClusterSupplyChain{}
				err := c.Get(ctx, client.ObjectKey{Name: "my-supply-chain"}, supplyChain)
				Expect(err).NotTo(HaveOccurred())

				return supplyChain.Status.Conditions

			}, 5*time.Second).Should(
				ContainElement(
					MatchFields(IgnoreExtras, Fields{
						"Type":   Equal("TemplatesReady"),
						"Status": Equal(v1.ConditionFalse),
						"Reason": Equal("TemplatesNotFound"),
					}),
				),
			)
		})

		Context("a change to the supply chain occurs that does not cause the status to change", func() {
			var conditionsBeforeMutation []v1.Condition

			BeforeEach(func() {
				// metav1.Time unmarshals with 1 second accuracy so this sleep ensures
				// the transition time is noticeable if it changes
				time.Sleep(1 * time.Second)

				supplyChain := &v1alpha1.ClusterSupplyChain{}
				err := c.Get(context.Background(), client.ObjectKey{Name: "my-supply-chain"}, supplyChain)
				Expect(err).NotTo(HaveOccurred())

				conditionsBeforeMutation = supplyChain.Status.Conditions

				supplyChain.Spec.Selector = map[string]string{"blah": "blah"}
				err = c.Update(context.Background(), supplyChain)
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() int64 {
					supplyChain := &v1alpha1.ClusterSupplyChain{}
					err := c.Get(context.Background(), client.ObjectKey{Name: "my-supply-chain"}, supplyChain)
					Expect(err).NotTo(HaveOccurred())
					return supplyChain.Status.ObservedGeneration
				}).Should(Equal(supplyChain.Generation))
			})

			It("does not update the lastTransitionTime", func() {
				supplyChain := &v1alpha1.ClusterSupplyChain{}
				err := c.Get(ctx, client.ObjectKey{Name: "my-supply-chain"}, supplyChain)
				Expect(err).NotTo(HaveOccurred())
				Expect(supplyChain.Status.Conditions).To(Equal(conditionsBeforeMutation))
			})
		})

		Context("a missing referenced template is created", func() {
			BeforeEach(func() {
				sourceTemplateYaml := utils.HereYaml(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterTemplate
				metadata:
				  name: my-terminal-template
				spec:
					template: {}
				`)

				sourceTemplate := &unstructured.Unstructured{}
				err := yaml.Unmarshal([]byte(sourceTemplateYaml), sourceTemplate)
				Expect(err).NotTo(HaveOccurred())

				err = c.Create(ctx, sourceTemplate, &client.CreateOptions{})
				cleanups = append(cleanups, sourceTemplate)
				Expect(err).NotTo(HaveOccurred())
			})

			It("immediately updates the supply-chain status", func() {
				Eventually(func() []v1.Condition {
					supplyChain := &v1alpha1.ClusterSupplyChain{}
					err := c.Get(ctx, client.ObjectKey{Name: "my-supply-chain"}, supplyChain)
					Expect(err).NotTo(HaveOccurred())

					return supplyChain.Status.Conditions

				}).Should(
					ContainElements(
						MatchFields(IgnoreExtras, Fields{
							"Type":   Equal("Ready"),
							"Status": Equal(v1.ConditionTrue),
						}),
						MatchFields(IgnoreExtras, Fields{
							"Type":   Equal("TemplatesReady"),
							"Status": Equal(v1.ConditionTrue),
						}),
					),
				)
			})
		})
	})
})
