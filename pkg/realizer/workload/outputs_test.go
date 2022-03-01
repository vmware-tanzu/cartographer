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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/workload"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

var _ = Describe("Outputs", func() {
	Describe("GenerateInputs", func() {
		Context("When resource contains sources", func() {
			var outs realizer.Outputs
			BeforeEach(func() {
				outs = realizer.NewOutputs()
				sourceOutput := &templates.Output{
					Source: &templates.Source{
						URL:      "source-url",
						Revision: "source-revision",
					},
				}
				outs.AddOutput("source-output", sourceOutput)
			})

			Context("And the sources have a match with the outputs", func() {
				It("Adds sources to inputs", func() {
					resource := &v1alpha1.SupplyChainResource{
						ResourceInputs: v1alpha1.ResourceInputs{
							Sources: []v1alpha1.ResourceReference{
								{
									Name:     "source-ref",
									Resource: "source-output",
								},
							},
						},
					}
					inputs := outs.GenerateInputs(resource)
					Expect(inputs.Sources).To(HaveLen(1))
					Expect(inputs.Sources["source-ref"].Name).To(Equal("source-ref"))
					Expect(inputs.Sources["source-ref"].URL).To(Equal("source-url"))
					Expect(inputs.Sources["source-ref"].Revision).To(Equal("source-revision"))
				})
			})

			Context("And the sources do not have a match with the outputs", func() {
				It("Does not add sources to inputs", func() {
					resource := &v1alpha1.SupplyChainResource{
						ResourceInputs: v1alpha1.ResourceInputs{
							Sources: []v1alpha1.ResourceReference{
								{
									Name:     "source-ref",
									Resource: "source-output-does-not-exist",
								},
							},
						},
					}
					inputs := outs.GenerateInputs(resource)
					Expect(len(inputs.Sources)).To(Equal(0))
				})
			})
		})

		Context("When resource contains images", func() {
			var outs realizer.Outputs
			BeforeEach(func() {
				outs = realizer.NewOutputs()
				imageOutput := &templates.Output{
					Image: "image12345",
				}
				outs.AddOutput("image-output", imageOutput)
			})

			Context("And the images have a match with the outputs", func() {
				It("Adds images to inputs", func() {
					resource := &v1alpha1.SupplyChainResource{
						ResourceInputs: v1alpha1.ResourceInputs{
							Images: []v1alpha1.ResourceReference{
								{
									Name:     "image-ref",
									Resource: "image-output",
								},
							},
						},
					}
					inputs := outs.GenerateInputs(resource)
					Expect(inputs.Images).To(HaveLen(1))
					Expect(inputs.Images["image-ref"].Name).To(Equal("image-ref"))
					Expect(inputs.Images["image-ref"].Image).To(Equal("image12345"))
				})
			})

			Context("And the images do not have a match with the outputs", func() {
				It("Does not add images to inputs", func() {
					resource := &v1alpha1.SupplyChainResource{
						ResourceInputs: v1alpha1.ResourceInputs{
							Images: []v1alpha1.ResourceReference{
								{
									Name:     "image-ref",
									Resource: "image-output-does-not-exist",
								},
							},
						},
					}
					inputs := outs.GenerateInputs(resource)
					Expect(inputs.Sources).To(BeEmpty())
				})
			})

		})

		Context("When resource contains configs", func() {
			var outs realizer.Outputs
			BeforeEach(func() {
				outs = realizer.NewOutputs()
				configOutput := &templates.Output{
					Config: "config12345",
				}
				outs.AddOutput("config-output", configOutput)
			})

			Context("And the configs have a match with the outputs", func() {
				It("Adds configs to inputs", func() {
					resource := &v1alpha1.SupplyChainResource{
						ResourceInputs: v1alpha1.ResourceInputs{
							Configs: []v1alpha1.ResourceReference{
								{
									Name:     "config-ref",
									Resource: "config-output",
								},
							},
						},
					}
					inputs := outs.GenerateInputs(resource)
					Expect(inputs.Configs).To(HaveLen(1))
					Expect(inputs.Configs["config-ref"].Name).To(Equal("config-ref"))
					Expect(inputs.Configs["config-ref"].Config).To(Equal("config12345"))
				})
			})

			Context("And the configs do not have a match with the outputs", func() {
				It("Does not add configs to inputs", func() {
					resource := &v1alpha1.SupplyChainResource{
						ResourceInputs: v1alpha1.ResourceInputs{
							Configs: []v1alpha1.ResourceReference{
								{
									Name:     "config-ref",
									Resource: "config-output-does-not-exist",
								},
							},
						},
					}
					inputs := outs.GenerateInputs(resource)
					Expect(inputs.Configs).To(BeEmpty())
				})
			})
		})
	})
})
