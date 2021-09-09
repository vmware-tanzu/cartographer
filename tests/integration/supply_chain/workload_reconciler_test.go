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
	"bufio"
	"context"
	"encoding/json"
	"io"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

type LogLine struct {
	Timestamp float64 `json:"ts"`
	Message   string  `json:"msg"`
	Name      string  `json:"name"`
	Namespace string  `json:"namespace"`
}

var _ = Describe("WorkloadReconciler", func() {
	var templateBytes = func() []byte {
		configMap := &corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example-config-map",
				Namespace: testNS,
			},
			Data: map[string]string{},
		}

		templateBytes, err := json.Marshal(configMap)
		Expect(err).ToNot(HaveOccurred())
		return templateBytes
	}

	var newClusterSupplyChain = func(name string, selector map[string]string) *v1alpha1.ClusterSupplyChain {
		return &v1alpha1.ClusterSupplyChain{
			TypeMeta: v1.TypeMeta{},
			ObjectMeta: v1.ObjectMeta{
				Name: name,
			},
			Spec: v1alpha1.SupplyChainSpec{
				Components: []v1alpha1.SupplyChainComponent{},
				Selector:   selector,
			},
		}
	}

	var reconcileAgain = func() {
		time.Sleep(1 * time.Second) //metav1.Time unmarshals with 1 second accuracy so this sleep avoids a race condition

		workload := &v1alpha1.Workload{}
		err := c.Get(context.Background(), client.ObjectKey{Name: "workload-bob", Namespace: testNS}, workload)
		Expect(err).NotTo(HaveOccurred())

		workload.Spec.Params = []v1alpha1.WorkloadParam{{Name: "foo", Value: apiextensionsv1.JSON{
			Raw: []byte(`"definitelybar"`),
		}}}
		err = c.Update(context.Background(), workload)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() bool {
			workload := &v1alpha1.Workload{}
			err := c.Get(context.Background(), client.ObjectKey{Name: "workload-bob", Namespace: testNS}, workload)
			Expect(err).NotTo(HaveOccurred())
			return workload.Status.ObservedGeneration == workload.Generation
		}).Should(BeTrue())
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

	Context("Has the source template and workload installed", func() {
		BeforeEach(func() {
			workload := &v1alpha1.Workload{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					Name:      "workload-bob",
					Namespace: testNS,
					Labels: map[string]string{
						"name": "webapp",
					},
				},
				Spec: v1alpha1.WorkloadSpec{},
			}

			cleanups = append(cleanups, workload)
			err := c.Create(ctx, workload, &client.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not update the lastTransitionTime on subsequent reconciliation if the status does not change", func() {
			var lastConditions []v1.Condition

			Eventually(func() bool {
				workload := &v1alpha1.Workload{}
				err := c.Get(context.Background(), client.ObjectKey{Name: "workload-bob", Namespace: testNS}, workload)
				Expect(err).NotTo(HaveOccurred())
				lastConditions = workload.Status.Conditions
				return workload.Status.ObservedGeneration == workload.Generation
			}, 5*time.Second).Should(BeTrue())

			reconcileAgain()

			workload := &v1alpha1.Workload{}
			err := c.Get(ctx, client.ObjectKey{Name: "workload-bob", Namespace: testNS}, workload)
			Expect(err).NotTo(HaveOccurred())

			Expect(workload.Status.Conditions).To(Equal(lastConditions))
		})

		Context("when reconciliation will end in an unknown status", func() {
			BeforeEach(func() {
				template := &v1alpha1.ClusterSourceTemplate{
					TypeMeta: v1.TypeMeta{},
					ObjectMeta: v1.ObjectMeta{
						Name: "proper-template-bob",
					},
					Spec: v1alpha1.SourceTemplateSpec{
						Template: runtime.RawExtension{Raw: templateBytes()},
						URLPath:  "nonexistant.path",
					},
				}

				cleanups = append(cleanups, template)
				err := c.Create(ctx, template, &client.CreateOptions{})
				Expect(err).NotTo(HaveOccurred())

				supplyChain := newClusterSupplyChain("supplychain-bob", map[string]string{"name": "webapp"})
				supplyChain.Spec.Components = []v1alpha1.SupplyChainComponent{
					{
						Name: "fred-component",
						TemplateRef: v1alpha1.ClusterTemplateReference{
							Kind: "ClusterSourceTemplate",
							Name: "proper-template-bob",
						},
					},
				}
				cleanups = append(cleanups, supplyChain)

				err = c.Create(ctx, supplyChain, &client.CreateOptions{})
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not error if the reconciliation ends in an unknown status", func() {
				Eventually(func() []v1.Condition {
					obj := &v1alpha1.Workload{}
					err := c.Get(ctx, client.ObjectKey{Name: "workload-bob", Namespace: testNS}, obj)
					Expect(err).NotTo(HaveOccurred())

					return obj.Status.Conditions
				}, 5*time.Second).Should(ContainElements(
					MatchFields(IgnoreExtras, Fields{
						"Type":   Equal("ComponentsSubmitted"),
						"Reason": Equal("MissingValueAtPath"),
						"Status": Equal(v1.ConditionStatus("Unknown")),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Type":   Equal("Ready"),
						"Reason": Equal("MissingValueAtPath"),
						"Status": Equal(v1.ConditionStatus("Unknown")),
					}),
				))
				Expect(controllerBuffer).NotTo(gbytes.Say("Reconciler error.*unable to retrieve outputs from stamped object for component"))
			})
		})

		It("shortcuts backoff when a supply chain is provided", func() {
			By("expecting a supply chain")
			Eventually(func() []v1.Condition {
				obj := &v1alpha1.Workload{}
				err := c.Get(ctx, client.ObjectKey{Name: "workload-bob", Namespace: testNS}, obj)
				Expect(err).NotTo(HaveOccurred())

				return obj.Status.Conditions
			}, 5*time.Second).Should(ContainElement(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal("SupplyChainReady"),
				"Reason": Equal("SupplyChainNotFound"),
				"Status": Equal(v1.ConditionStatus("False")),
			})))

			// todo: this test is flakey
			reader := bufio.NewReader(controllerBuffer)
			var previousSeconds float64
			Eventually(func() float64 {
				line, _, err := reader.ReadLine()
				if err == io.EOF {
					return 0
				}
				Expect(err).NotTo(HaveOccurred())
				var logLine LogLine
				err = json.Unmarshal(line, &logLine)
				if err != nil {
					return 0
				}
				if logLine.Message != "Reconciler error" || logLine.Namespace != testNS || logLine.Name != "workload-bob" {
					return 0
				}

				if previousSeconds == 0 {
					previousSeconds = logLine.Timestamp
					return 0
				}

				return logLine.Timestamp - previousSeconds
			}, 2*time.Second).Should(BeNumerically(">", 1.0))

			By("accepting a supply chain")
			supplyChain := newClusterSupplyChain("supplychain-bob", map[string]string{"name": "webapp"})
			cleanups = append(cleanups, supplyChain)

			err := c.Create(ctx, supplyChain, &client.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			obj := &v1alpha1.ClusterSupplyChain{}
			Eventually(func() ([]v1.Condition, error) {
				err = c.Get(ctx, client.ObjectKey{Name: "supplychain-bob"}, obj)
				return obj.Status.Conditions, err
			}, 5*time.Second).Should(ContainElement(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal("Ready"),
				"Reason": Equal("Ready"),
				"Status": Equal(v1.ConditionStatus("True")),
			})))

			By("reconcile in less than a second")
			Eventually(func() ([]v1.Condition, error) {
				err = c.Get(ctx, client.ObjectKey{Name: "supplychain-bob"}, obj)
				return obj.Status.Conditions, err
			}, 500*time.Millisecond).Should(ContainElement(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal("Ready"),
				"Reason": Equal("Ready"),
				"Status": Equal(v1.ConditionStatus("True")),
			})))
		})
	})
})
