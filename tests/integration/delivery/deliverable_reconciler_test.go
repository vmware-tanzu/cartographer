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

package delivery_test

import (
	"context"
	"encoding/json"
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

var _ = Describe("DeliverableReconciler", func() {
	var templateBytes = func() []byte {
		configMap := &corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "example-config-map",
			},
			Data: map[string]string{},
		}

		templateBytes, err := json.Marshal(configMap)
		Expect(err).ToNot(HaveOccurred())
		return templateBytes
	}

	var newClusterDelivery = func(name string, selector map[string]string) *v1alpha1.ClusterDelivery {
		return &v1alpha1.ClusterDelivery{
			TypeMeta: v1.TypeMeta{},
			ObjectMeta: v1.ObjectMeta{
				Name: name,
			},
			Spec: v1alpha1.ClusterDeliverySpec{
				Resources: []v1alpha1.ClusterDeliveryResource{},
				Selector:  selector,
			},
		}
	}

	var reconcileAgain = func() {
		time.Sleep(1 * time.Second) //metav1.Time unmarshals with 1 second accuracy so this sleep avoids a race condition

		deliverable := &v1alpha1.Deliverable{}
		err := c.Get(context.Background(), client.ObjectKey{Name: "deliverable-bob", Namespace: testNS}, deliverable)
		Expect(err).NotTo(HaveOccurred())

		deliverable.Spec.Params = []v1alpha1.Param{{Name: "foo", Value: apiextensionsv1.JSON{
			Raw: []byte(`"definitelybar"`),
		}}}
		err = c.Update(context.Background(), deliverable)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() bool {
			deliverable := &v1alpha1.Deliverable{}
			err := c.Get(context.Background(), client.ObjectKey{Name: "deliverable-bob", Namespace: testNS}, deliverable)
			Expect(err).NotTo(HaveOccurred())
			return deliverable.Status.ObservedGeneration == deliverable.Generation
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

	Context("Has the source template and deliverable installed", func() {
		BeforeEach(func() {
			deliverable := &v1alpha1.Deliverable{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					Name:      "deliverable-bob",
					Namespace: testNS,
					Labels: map[string]string{
						"name": "webapp",
					},
				},
				Spec: v1alpha1.DeliverableSpec{},
			}

			cleanups = append(cleanups, deliverable)
			err := c.Create(ctx, deliverable, &client.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not update the lastTransitionTime on subsequent reconciliation if the status does not change", func() {
			var lastConditions []v1.Condition

			Eventually(func() bool {
				deliverable := &v1alpha1.Deliverable{}
				err := c.Get(context.Background(), client.ObjectKey{Name: "deliverable-bob", Namespace: testNS}, deliverable)
				Expect(err).NotTo(HaveOccurred())
				lastConditions = deliverable.Status.Conditions
				return deliverable.Status.ObservedGeneration == deliverable.Generation
			}, 5*time.Second).Should(BeTrue())

			reconcileAgain()

			deliverable := &v1alpha1.Deliverable{}
			err := c.Get(ctx, client.ObjectKey{Name: "deliverable-bob", Namespace: testNS}, deliverable)
			Expect(err).NotTo(HaveOccurred())

			Expect(deliverable.Status.Conditions).To(Equal(lastConditions))
		})

		Context("when reconciliation will end in an unknown status", func() {
			BeforeEach(func() {
				template := &v1alpha1.ClusterSourceTemplate{
					TypeMeta: v1.TypeMeta{},
					ObjectMeta: v1.ObjectMeta{
						Name: "proper-template-bob",
					},
					Spec: v1alpha1.SourceTemplateSpec{
						TemplateSpec: v1alpha1.TemplateSpec{
							Template: &runtime.RawExtension{Raw: templateBytes()},
						},
						URLPath: "nonexistant.path",
					},
				}

				cleanups = append(cleanups, template)
				err := c.Create(ctx, template, &client.CreateOptions{})
				Expect(err).NotTo(HaveOccurred())

				delivery := newClusterDelivery("delivery-bob", map[string]string{"name": "webapp"})
				delivery.Spec.Resources = []v1alpha1.ClusterDeliveryResource{
					{
						Name: "fred-resource",
						TemplateRef: v1alpha1.DeliveryClusterTemplateReference{
							Kind: "ClusterSourceTemplate",
							Name: "proper-template-bob",
						},
					},
				}
				cleanups = append(cleanups, delivery)

				err = c.Create(ctx, delivery, &client.CreateOptions{})
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not error if the reconciliation ends in an unknown status", func() {
				Eventually(func() []v1.Condition {
					obj := &v1alpha1.Deliverable{}
					err := c.Get(ctx, client.ObjectKey{Name: "deliverable-bob", Namespace: testNS}, obj)
					Expect(err).NotTo(HaveOccurred())

					return obj.Status.Conditions
				}, 5*time.Second).Should(ContainElements(
					MatchFields(IgnoreExtras, Fields{
						"Type":   Equal("ResourcesSubmitted"),
						"Reason": Equal("MissingValueAtPath"),
						"Status": Equal(v1.ConditionStatus("Unknown")),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Type":   Equal("Ready"),
						"Reason": Equal("MissingValueAtPath"),
						"Status": Equal(v1.ConditionStatus("Unknown")),
					}),
				))
				Expect(controllerBuffer).NotTo(gbytes.Say("Reconciler error.*unable to retrieve outputs from stamped object for resource"))
			})
		})
	})
})
