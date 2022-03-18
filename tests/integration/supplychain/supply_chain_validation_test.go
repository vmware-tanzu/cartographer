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
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
				LegacySelector: v1alpha1.LegacySelector{
					Selector: map[string]string{"integration-test": "workload-no-supply-chain"},
				},
			},
		}
	}
	var SupplyChainWithTemplateReference = func(templateName string, templateKind string) *v1alpha1.ClusterSupplyChain {
		result := BaseSupplyChain("my-supply-chain")
		result.Spec.Resources = []v1alpha1.SupplyChainResource{
			{
				Name: "funky-resource",
				TemplateRef: v1alpha1.SupplyChainTemplateReference{
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

	Describe("option doesn't include selector", func() {
		var err error
		Context("Submitting a new supply chain", func() {
			BeforeEach(func() {
				supplyChain.Spec.Resources = []v1alpha1.SupplyChainResource{
					{
						Name: "source-provider",
						TemplateRef: v1alpha1.SupplyChainTemplateReference{
							Kind: "ClusterSourceTemplate",
							Options: []v1alpha1.TemplateOption{
								{
									Name: "git-template---default-params",
								},
								{
									Name: "some-other-template",
									Selector: v1alpha1.Selector{
										MatchFields: []v1alpha1.FieldSelectorRequirement{
											{
												Key:      "spec.source.git.url",
												Operator: v1alpha1.FieldSelectorOpExists,
											},
										},
									},
								},
							},
						},
					},
				}
			})
			It("Rejects the supply chain", func() {
				err = c.Create(ctx, supplyChain)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(`error validating clustersupplychain [responsible-ops---default-params]: error validating resource [source-provider]: error validating option [git-template---default-params] selector: at least one of matchLabels, matchExpressions or MatchFields must be specified`))
			})
		})
	})
})
