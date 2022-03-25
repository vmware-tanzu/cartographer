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

var _ = Describe("Supply Chain selection for workloads", func() {
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

	Describe("Workload matches one SupplyChain based on label", func() {
		BeforeEach(func() {
			// job-sc: supply chain that matches on type:job via selector [should match]
			apply(utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterSupplyChain
					metadata:
					  name: job-sc
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

		When("I apply a workload with label type:job", func() {
			BeforeEach(func() {
				apply(utils.HereYamlF(`
						---
						apiVersion: carto.run/v1alpha1
						kind: Workload
						metadata:
						  labels:
						    type: job
						  name: workload-with-type-job
					      namespace: %s
						spec:
						  serviceAccountName: default
				          image: https://docker.io/samsplace/my-iot-image
					`, testNS))
			})

			It("matches", func() {
				Eventually(func() (v1alpha1.WorkloadStatus, error) {
					workload := &v1alpha1.Workload{}
					err := c.Get(ctx, client.ObjectKey{Name: "workload-with-type-job", Namespace: testNS}, workload)
					return workload.Status, err
				}).Should(
					MatchFields(IgnoreExtras, Fields{
						"SupplyChainRef": MatchFields(IgnoreExtras, Fields{
							"Kind": Equal("ClusterSupplyChain"),
							"Name": Equal("job-sc"),
						}),
					}),
				)
			})
		})
	})

	Describe("Workload matches one SupplyChain based on selectorMatchExpressions", func() {
		BeforeEach(func() {
			// job-via-match-expression-sc: supply chain that matches on type:job via selectorMatchExpressions [should match]
			apply(utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterSupplyChain
					metadata:
					  name: job-via-match-expression-sc
					spec:
					  selectorMatchExpressions:
                        - { key: "type", operator: "In", values: [ "job" ]}
					  resources:
						- name: my-first-resource
						  templateRef:
							kind: ClusterTemplate
							name: my-template
				`))
		})

		When("I apply a workload with label type:job", func() {
			BeforeEach(func() {
				apply(utils.HereYamlF(`
						---
						apiVersion: carto.run/v1alpha1
						kind: Workload
						metadata:
						  labels:
						    type: job
						  name: workload-with-type-job
					      namespace: %s
						spec:
						  serviceAccountName: default
				          image: https://docker.io/samsplace/my-iot-image
					`, testNS))
			})

			It("matches", func() {
				Eventually(func() (v1alpha1.WorkloadStatus, error) {
					workload := &v1alpha1.Workload{}
					err := c.Get(ctx, client.ObjectKey{Name: "workload-with-type-job", Namespace: testNS}, workload)
					return workload.Status, err
				}).Should(
					MatchFields(IgnoreExtras, Fields{
						"SupplyChainRef": MatchFields(IgnoreExtras, Fields{
							"Kind": Equal("ClusterSupplyChain"),
							"Name": Equal("job-via-match-expression-sc"),
						}),
					}),
				)
			})
		})
	})

	Describe("Workload matches one SupplyChain based on selectorMatchFields", func() {
		BeforeEach(func() {
			// job-via-match-fields-sc: supply chain that matches on type:job via selectorMatchFields [should match]
			apply(utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterSupplyChain
					metadata:
					  name: job-via-match-fields-sc
					spec:
					  selectorMatchFields:
                        - { key: "metadata.labels.type", operator: "In", values: [ "job" ]}
					  resources:
						- name: my-first-resource
						  templateRef:
							kind: ClusterTemplate
							name: my-template
				`))
		})

		When("I apply a workload with label type:job", func() {
			BeforeEach(func() {
				apply(utils.HereYamlF(`
						---
						apiVersion: carto.run/v1alpha1
						kind: Workload
						metadata:
						  labels:
						    type: job
						  name: workload-with-type-job
					      namespace: %s
						spec:
						  serviceAccountName: default
				          image: https://docker.io/samsplace/my-iot-image
					`, testNS))
			})

			It("matches", func() {
				Eventually(func() (v1alpha1.WorkloadStatus, error) {
					workload := &v1alpha1.Workload{}
					err := c.Get(ctx, client.ObjectKey{Name: "workload-with-type-job", Namespace: testNS}, workload)
					return workload.Status, err
				}).Should(
					MatchFields(IgnoreExtras, Fields{
						"SupplyChainRef": MatchFields(IgnoreExtras, Fields{
							"Kind": Equal("ClusterSupplyChain"),
							"Name": Equal("job-via-match-fields-sc"),
						}),
					}),
				)
			})
		})
	})

	Describe("Workload does not match any of many supply chains", func() {
		BeforeEach(func() {
			// web-on-main-sc: supply chain that matches on type:web and is git and is on main [should match, most specific]
			apply(utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterSupplyChain
					metadata:
					  name: web-on-main-sc
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

			// web-with-git-sc: supply chain that matches on type:web and is git [shouldn't match, less specific]
			apply(utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterSupplyChain
					metadata:
					  name: web-with-git-sc
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

			// job-sc: supply chain that matches on type:job [shouldn't match]
			apply(utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterSupplyChain
					metadata:
					  name: job-sc
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

		When("I apply a workload with an image and label type:edgeIOT", func() {
			BeforeEach(func() {
				apply(utils.HereYamlF(`
						---
						apiVersion: carto.run/v1alpha1
						kind: Workload
						metadata:
						  labels:
						    type: edgeIOT
						  name: workload-with-image-for-edge-iot
					      namespace: %s
						spec:
						  serviceAccountName: default
				          image: https://docker.io/samsplace/my-iot-image
					`, testNS))
			})

			It("does not match", func() {
				Eventually(func() ([]metav1.Condition, error) {
					workload := &v1alpha1.Workload{}
					err := c.Get(ctx, client.ObjectKey{Name: "workload-with-image-for-edge-iot", Namespace: testNS}, workload)
					return workload.Status.Conditions, err
				}).Should(
					ContainElements(
						MatchFields(IgnoreExtras, Fields{
							"Type":    Equal("SupplyChainReady"),
							"Status":  Equal(metav1.ConditionFalse),
							"Reason":  Equal("SupplyChainNotFound"),
							"Message": MatchRegexp("^no supply chain found where"), // Fixme: The error we emit is specific to fields and only one supply chain. Needs review (post dev release)
						}),
					),
				)
			})
		})

	})

	Describe("Workload matches most specific supply chain", func() {
		BeforeEach(func() {
			// web-on-main-sc: supply chain that matches on type:web and is git and is on main [should match, most specific]
			apply(utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterSupplyChain
					metadata:
					  name: web-on-main-sc
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

			// web-with-git-sc: supply chain that matches on type:web and is git [shouldn't match, less specific]
			apply(utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterSupplyChain
					metadata:
					  name: web-with-git-sc
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

			// job-sc: supply chain that matches on type:job [shouldn't match]
			apply(utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterSupplyChain
					metadata:
					  name: job-sc
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

		When("A workload with a git repository on main and a web label is applied", func() {
			BeforeEach(func() {
				apply(utils.HereYamlF(`
						---
						apiVersion: carto.run/v1alpha1
						kind: Workload
						metadata:
						  labels:
						    type: web
						  name: workload-on-main-for-web
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

			It("matches the web-on-main-sc supply chain", func() {
				Eventually(func() (v1alpha1.WorkloadStatus, error) {
					workload := &v1alpha1.Workload{}
					err := c.Get(ctx, client.ObjectKey{Name: "workload-on-main-for-web", Namespace: testNS}, workload)
					return workload.Status, err
				}).Should(
					MatchFields(IgnoreExtras, Fields{
						"SupplyChainRef": MatchFields(IgnoreExtras, Fields{
							"Kind": Equal("ClusterSupplyChain"),
							"Name": Equal("web-on-main-sc"),
						}),
					}),
				)
			})
		})
	})

	Describe("Workload matches two 'most specific' supply chains", func() { // suggested domain term: `Most Machingest`
		BeforeEach(func() {
			// web-on-main-sc: supply chain that matches on type:web and is git and is on main [should match, most specific]
			apply(utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterSupplyChain
					metadata:
					  name: web-on-main-sc
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

			// web-on-main-or-master-sc: supply chain that matches on type:web and is git and is on main or master [should match, most specific]
			apply(utils.HereYaml(`
					---
					apiVersion: carto.run/v1alpha1
					kind: ClusterSupplyChain
					metadata:
					  name: web-on-main-or-master-sc
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
		When("A workload with a git repository on main and a web label is applied", func() {
			BeforeEach(func() {
				apply(utils.HereYamlF(`
						---
						apiVersion: carto.run/v1alpha1
						kind: Workload
						metadata:
						  labels:
						    type: web
						  name: workload-on-main-for-web
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

			It("matches the web-on-main-sc supply chain", func() {
				Eventually(func() ([]metav1.Condition, error) {
					workload := &v1alpha1.Workload{}
					err := c.Get(ctx, client.ObjectKey{Name: "workload-on-main-for-web", Namespace: testNS}, workload)
					return workload.Status.Conditions, err
				}).Should(
					ContainElements(
						MatchFields(IgnoreExtras, Fields{
							"Type":    Equal("SupplyChainReady"),
							"Status":  Equal(metav1.ConditionFalse),
							"Reason":  Equal("MultipleSupplyChainMatches"),
							"Message": Equal("workload may only match a single supply chain's selector"),
						}),
					),
				)
			})
		})
	})
})
