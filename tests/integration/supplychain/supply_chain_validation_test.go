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

	. "github.com/MakeNowJust/heredoc/dot"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

var _ = Describe("SupplyChainValidation", func() {
	var (
		ctx         context.Context
		supplyChain *v1alpha1.ClusterSupplyChain
	)

	var BaseSupplyChain = func(name string) *v1alpha1.ClusterSupplyChain {
		return &v1alpha1.ClusterSupplyChain{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "default",
			},
			Spec: v1alpha1.SupplyChainSpec{
				Resources: []v1alpha1.SupplyChainResource{},
				Selector:  map[string]string{"integration-test": "workload-no-supply-chain"},
			},
		}
	}
	var SupplyChainWithTemplateReference = func(templateName string, templateKind string) *v1alpha1.ClusterSupplyChain {
		result := BaseSupplyChain("my-supply-chain")
		result.Spec.Resources = []v1alpha1.SupplyChainResource{
			{
				Name: "funky-resource",
				TemplateRef: v1alpha1.ClusterTemplateReference{
					Kind: templateKind,
					Name: templateName,
				},
			},
		}

		return result
	}

	BeforeEach(func() {
		ctx = context.Background()
		supplyChain = BaseSupplyChain("responsible-ops---default-params")
	})

	AfterEach(func() {
		_ = c.Delete(ctx, supplyChain)
	})

	Context("Supply Chain resource names are unique", func() {
		Context("Submitting a new supply chain", func() {
			BeforeEach(func() {
				supplyChain.Spec.Resources = []v1alpha1.SupplyChainResource{
					{
						Name: "source-provider",
						TemplateRef: v1alpha1.ClusterTemplateReference{
							Kind: "ClusterSourceTemplate",
							Name: "git-template---default-params",
						},
					},
					{
						Name: "other-source-provider",
						TemplateRef: v1alpha1.ClusterTemplateReference{
							Kind: "ClusterSourceTemplate",
							Name: "git-template---default-params",
						},
					},
				}
			})

			It("Accepts the supply chain", func() {
				err := c.Create(ctx, supplyChain)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("Updating a valid supply chain", func() {
			BeforeEach(func() {
				supplyChain.Spec.Resources = []v1alpha1.SupplyChainResource{
					{
						Name: "source-provider",
						TemplateRef: v1alpha1.ClusterTemplateReference{
							Kind: "ClusterSourceTemplate",
							Name: "git-template---default-params",
						},
					},
					{
						Name: "other-source-provider",
						TemplateRef: v1alpha1.ClusterTemplateReference{
							Kind: "ClusterSourceTemplate",
							Name: "git-template---default-params",
						},
					},
				}
				err := c.Create(ctx, supplyChain)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Accepts the supply chain", func() {
				supplyChain.Spec.Resources = append(
					supplyChain.Spec.Resources,
					v1alpha1.SupplyChainResource{
						Name: "another-other-source-provider",
						TemplateRef: v1alpha1.ClusterTemplateReference{
							Kind: "ClusterSourceTemplate",
							Name: "git-template---default-params",
						},
					},
				)

				err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
					err := c.Get(ctx, client.ObjectKey{Name: supplyChain.Name}, supplyChain)
					Expect(err).ToNot(HaveOccurred())

					resources := []v1alpha1.SupplyChainResource{
						{
							Name: "source-provider",
							TemplateRef: v1alpha1.ClusterTemplateReference{
								Kind: "ClusterSourceTemplate",
								Name: "git-template---default-params",
							},
						},
						{
							Name: "other-source-provider",
							TemplateRef: v1alpha1.ClusterTemplateReference{
								Kind: "ClusterSourceTemplate",
								Name: "git-template---default-params",
							},
						},
						{
							Name: "another-other-source-provider",
							TemplateRef: v1alpha1.ClusterTemplateReference{
								Kind: "ClusterSourceTemplate",
								Name: "git-template---default-params",
							},
						},
					}

					supplyChain.Spec.Resources = resources

					return c.Update(ctx, supplyChain)
				})
				Expect(err).ToNot(HaveOccurred())
			})

		})
	})

	Context("Supply Chain resource names are identical", func() {
		var err error
		Context("Submitting a new supply chain", func() {
			BeforeEach(func() {
				supplyChain.Spec.Resources = []v1alpha1.SupplyChainResource{
					{
						Name: "source-provider",
						TemplateRef: v1alpha1.ClusterTemplateReference{
							Kind: "ClusterSourceTemplate",
							Name: "git-template---default-params",
						},
					},
					{
						Name: "source-provider",
						TemplateRef: v1alpha1.ClusterTemplateReference{
							Kind: "ClusterImageTemplate",
							Name: "some-other-template",
						},
					},
				}
			})
			It("Rejects the supply chain", func() {
				err = c.Create(ctx, supplyChain)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("duplicate resource name 'source-provider' found in clustersupplychain 'responsible-ops---default-params'"))
			})
		})

		Context("Updating a valid supply chain", func() {
			BeforeEach(func() {
				supplyChain.Spec.Resources = []v1alpha1.SupplyChainResource{
					{
						Name: "source-provider",
						TemplateRef: v1alpha1.ClusterTemplateReference{
							Kind: "ClusterSourceTemplate",
							Name: "git-template---default-params",
						},
					},
				}
				err = c.Create(ctx, supplyChain)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Rejects the supply chain", func() {
				err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
					err := c.Get(ctx, client.ObjectKey{Name: supplyChain.Name}, supplyChain)
					Expect(err).ToNot(HaveOccurred())

					resource := v1alpha1.SupplyChainResource{
						Name: "source-provider",
						TemplateRef: v1alpha1.ClusterTemplateReference{
							Kind: "ClusterSourceTemplate",
							Name: "git-template---default-params",
						},
					}
					resources := []v1alpha1.SupplyChainResource{
						resource,
						resource,
					}

					supplyChain.Spec.Resources = resources

					return c.Update(ctx, supplyChain)
				})
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("duplicate resource name 'source-provider' found in clustersupplychain 'responsible-ops---default-params'"))
			})
		})
	})

	Context("resource types are mismatched", func() {
		var err error
		Context("Submitting a new supply chain", func() {
			BeforeEach(func() {
				supplyChain = BaseSupplyChain("my-supply-chain")
				supplyChain.Spec.Resources = []v1alpha1.SupplyChainResource{
					{
						Name: "provider",
						TemplateRef: v1alpha1.ClusterTemplateReference{
							Kind: "ClusterImageTemplate",
							Name: "dockerfile-build",
						},
					},
					{
						Name: "build-image-provider",
						TemplateRef: v1alpha1.ClusterTemplateReference{
							Kind: "ClusterImageTemplate",
							Name: "kpack-battery",
						},
						Sources: []v1alpha1.ResourceReference{
							{
								Resource: "provider",
								Name:     "solo-source-provider",
							},
						},
					},
				}
			})
			It("Rejects the supply chain", func() {
				err = c.Create(ctx, supplyChain)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("resource 'provider' providing 'solo-source-provider' must reference a ClusterSourceTemplate"))
			})
		})
	})

	Describe("Template reference kind field", func() {
		DescribeTable("Reference a valid Template Kind", func(kind string) {
			supplyChain = SupplyChainWithTemplateReference("my-template", kind)
			err := c.Create(ctx, supplyChain)
			Expect(err).NotTo(HaveOccurred())
		},
			Entry("ClusterImageTemplate reference", "ClusterImageTemplate"),
			Entry("ClusterConfigTemplate reference", "ClusterConfigTemplate"),
			Entry("ClusterSourceTemplate reference", "ClusterSourceTemplate"),
			Entry("ClusterTemplate reference", "ClusterTemplate"),
		)

		DescribeTable("Reference an invalid kind", func(kind string) {
			supplyChain = SupplyChainWithTemplateReference("my-template", kind)
			err := c.Create(ctx, supplyChain)
			Expect(err).To(HaveOccurred())
		},
			Entry("Empty kind in reference", ""),
			Entry("Unsupported kind in reference", "ClusterBobTemplate"),
		)
	})

	Describe("Template reference name field", func() {
		Context("with an empty string for name", func() {
			BeforeEach(func() {
				supplyChain = SupplyChainWithTemplateReference("", "ClusterImageTemplate")
			})
			It("Rejects the the supply chain", func() {
				err := c.Create(ctx, supplyChain)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("spec.resources.templateRef.name in body should be at least 1 chars long"))
			})
		})

		Context("with a missing name field", func() {
			var supplyChainUnstructured *unstructured.Unstructured
			var supplyChainJson = D(`
				---
				apiVersion: carto.run/v1alpha1
				kind: ClusterSupplyChain
				metadata:
				  name: my-supply-chain
				spec:
				  selector:
				    integration-test: must-have-selector
				resources:
				- name: source-provider
				  templateRef:
				    kind: ClusterSourceTemplate
			`)
			BeforeEach(func() {
				supplyChainUnstructured = &unstructured.Unstructured{}
				err := yaml.Unmarshal([]byte(supplyChainJson), supplyChainUnstructured)
				Expect(err).NotTo(HaveOccurred())

			})
			It("Rejects the the supply chain", func() {
				err := c.Create(ctx, supplyChainUnstructured)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(`ClusterSupplyChain.carto.run "my-supply-chain" is invalid: spec.resources:`))
			})
		})

		Context("with a non-empty name", func() {
			BeforeEach(func() {
				supplyChain = SupplyChainWithTemplateReference("hey-there", "ClusterImageTemplate")
			})
			It("Accepts the the supply chain", func() {
				err := c.Create(ctx, supplyChain)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
