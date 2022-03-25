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

// TODO: this test looks exactly like the Supply Chain delivery test (Coz it is)
//       Make it a little more realistic for delivery/deliverable

package delivery_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

var _ = Describe("Delivery selection for deliverables", func() {
	var (
		ctx      context.Context
		cleanups []client.Object
	)

	var apply = func(resourceYaml string) {
		resource := &unstructured.Unstructured{}
		err := yaml.Unmarshal([]byte(resourceYaml), resource)
		Expect(err).NotTo(HaveOccurred())

		err = c.Create(ctx, resource, &client.CreateOptions{})
		cleanups = append(cleanups, resource)
		Expect(err).NotTo(HaveOccurred())
	}

	BeforeEach(func() {
		ctx = context.Background()

		apply(utils.HereYaml(`
			---
			apiVersion: carto.run/v1alpha1
			kind: ClusterTemplate
			metadata:
			  name: my-template
			spec:
			  template:
                apiVersion: v1
                kind: ConfigMap
                metadata:
                  name: my-config-map
                data:
                  foo: "bar"
		`))

	})

	AfterEach(func() {
		for _, obj := range cleanups {
			_ = c.Delete(ctx, obj, &client.DeleteOptions{})
		}
	})

	// Scenario: Deliverable matches label on one Delivery
	// Scenario: Deliverable matches expression on one Delivery
	// Scenario: Deliverable matches field on one Delivery

	Describe("Deliverable does not match any of many deliveries", func() {
		BeforeEach(func() {
			// web-on-main-delivery: delivery that matches on type:web and is git and is on main [should match, most specific]
			apply(utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterDelivery
					metadata:
					  name: web-on-main-delivery
					spec:
					  selector:
						"type": "web"
					  selectorMatchFields:
						- { key: "spec.source.git", operator: "Exists" }
						- { key: "spec.source.git.ref.branch", operator: "In", values: ["main"] }
					  resources:
						- name: my-first-resource
						  templateRef:
							kind: ClusterTemplate
							name: my-template
				`))

			// web-with-git-delivery: delivery that matches on type:web and is git [shouldn't match, less specific]
			apply(utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterDelivery
					metadata:
					  name: web-with-git-delivery
					spec:
					  selector:
						"type": "web"
					  selectorMatchFields:
						- { key: "spec.source.git", operator: "Exists" }
					  resources:
						- name: my-first-resource
						  templateRef:
							kind: ClusterTemplate
							name: my-template
				`))

			// job-delivery: delivery that matches on type:job [shouldn't match]
			apply(utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterDelivery
					metadata:
					  name: job-delivery
					spec:
					  selector:
						"type": "job"
					  resources:
						- name: my-first-resource
						  templateRef:
							kind: ClusterTemplate
							name: my-template
				`))
		})

		When("I apply a deliverable with an image and label type:edgeIOT", func() {
			BeforeEach(func() {
				apply(utils.HereYamlF(`
						---
						apiVersion: carto.run/v1alpha1
						kind: Deliverable
						metadata:
						  labels:
						    type: edgeIOT
						  name: deliverable-with-image-for-edge-iot
					      namespace: %s
						spec:
						  serviceAccountName: default
				          image: https://docker.io/samsplace/my-iot-image
					`, testNS))
			})

			It("does not match", func() {
				Eventually(func() ([]metav1.Condition, error) {
					deliverable := &v1alpha1.Deliverable{}
					err := c.Get(ctx, client.ObjectKey{Name: "deliverable-with-image-for-edge-iot", Namespace: testNS}, deliverable)
					return deliverable.Status.Conditions, err
				}).Should(
					ContainElements(
						MatchFields(IgnoreExtras, Fields{
							"Type":    Equal("DeliveryReady"),
							"Status":  Equal(metav1.ConditionFalse),
							"Reason":  Equal("DeliveryNotFound"),
							"Message": MatchRegexp("^no delivery found where"), // Fixme: The error we emit is specific to fields and only one delivery. Needs review (post dev release)
						}),
					),
				)
			})
		})

	})

	Describe("Deliverable matches most specific delivery", func() {
		BeforeEach(func() {
			// web-on-main-delivery: delivery that matches on type:web and is git and is on main [should match, most specific]
			apply(utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterDelivery
					metadata:
					  name: web-on-main-delivery
					spec:
					  selector:
						"type": "web"
					  selectorMatchFields:
						- { key: "spec.source.git", operator: "Exists" }
						- { key: "spec.source.git.ref.branch", operator: "In", values: ["main"] }
					  resources:
						- name: my-first-resource
						  templateRef:
							kind: ClusterTemplate
							name: my-template
				`))

			// web-with-git-delivery: delivery that matches on type:web and is git [shouldn't match, less specific]
			apply(utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterDelivery
					metadata:
					  name: web-with-git-delivery
					spec:
					  selector:
						"type": "web"
					  selectorMatchFields:
						- { key: "spec.source.git", operator: "Exists" }
					  resources:
						- name: my-first-resource
						  templateRef:
							kind: ClusterTemplate
							name: my-template
				`))

			// job-delivery: delivery that matches on type:job [shouldn't match]
			apply(utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterDelivery
					metadata:
					  name: job-delivery
					spec:
					  selector:
						"type": "job"
					  resources:
						- name: my-first-resource
						  templateRef:
							kind: ClusterTemplate
							name: my-template
				`))
		})

		When("A deliverable with a git repository on main and a web label is applied", func() {
			BeforeEach(func() {
				apply(utils.HereYamlF(`
						---
						apiVersion: carto.run/v1alpha1
						kind: Deliverable
						metadata:
						  labels:
						    type: web
						  name: deliverable-on-main-for-web
					      namespace: %s
						spec:
						  serviceAccountName: default
						  source:
						    git:
						      url: https://github.com/my-app.git
						      ref:
						        branch: main
					`, testNS))
			})

			It("matches the web-on-main-delivery delivery", func() {
				Eventually(func() (v1alpha1.DeliverableStatus, error) {
					deliverable := &v1alpha1.Deliverable{}
					err := c.Get(ctx, client.ObjectKey{Name: "deliverable-on-main-for-web", Namespace: testNS}, deliverable)
					return deliverable.Status, err
				}).Should(
					MatchFields(IgnoreExtras, Fields{
						"DeliveryRef": MatchFields(IgnoreExtras, Fields{
							"Kind": Equal("ClusterDelivery"),
							"Name": Equal("web-on-main-delivery"),
						}),
					}),
				)
			})
		})
	})

	Describe("Deliverable matches two 'most specific' deliveries", func() { // suggested domain term: `Most Machingest`
		BeforeEach(func() {
			// web-on-main-delivery: delivery that matches on type:web and is git and is on main [should match, most specific]
			apply(utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterDelivery
					metadata:
					  name: web-on-main-delivery
					spec:
					  selector:
						"type": "web"
					  selectorMatchFields:
						- { key: "spec.source.git", operator: "Exists" }
						- { key: "spec.source.git.ref.branch", operator: "In", values: ["main"] }
					  resources:
						- name: my-first-resource
						  templateRef:
							kind: ClusterTemplate
							name: my-template
				`))

			// web-on-main-or-master-delivery: delivery that matches on type:web and is git and is on main or master [should match, most specific]
			apply(utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterDelivery
					metadata:
					  name: web-on-main-or-master-delivery
					spec:
					  selector:
						"type": "web"
					  selectorMatchFields:
						- { key: "spec.source.git", operator: "Exists" }
						- { key: "spec.source.git.ref.branch", operator: "In", values: ["main", "master"] }
					  resources:
						- name: my-first-resource
						  templateRef:
							kind: ClusterTemplate
							name: my-template
				`))

		})
		When("A deliverable with a git repository on main and a web label is applied", func() {
			BeforeEach(func() {
				apply(utils.HereYamlF(`
						---
						apiVersion: carto.run/v1alpha1
						kind: Deliverable
						metadata:
						  labels:
						    type: web
						  name: deliverable-on-main-for-web
					      namespace: %s
						spec:
						  serviceAccountName: default
						  source:
						    git:
						      url: https://github.com/my-app.git
						      ref:
						        branch: main
					`, testNS))
			})

			It("matches the web-on-main-delivery delivery", func() {
				Eventually(func() ([]metav1.Condition, error) {
					deliverable := &v1alpha1.Deliverable{}
					err := c.Get(ctx, client.ObjectKey{Name: "deliverable-on-main-for-web", Namespace: testNS}, deliverable)
					return deliverable.Status.Conditions, err
				}).Should(
					ContainElements(
						MatchFields(IgnoreExtras, Fields{
							"Type":    Equal("DeliveryReady"),
							"Status":  Equal(metav1.ConditionFalse),
							"Reason":  Equal("MultipleDeliveryMatches"),
							"Message": Equal("deliverable may only match a single delivery's selector"),
						}),
					),
				)
			})
		})
	})
})
