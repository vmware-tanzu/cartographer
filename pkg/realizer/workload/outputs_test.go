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
		Context("When component contains sources", func() {
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
					component := &v1alpha1.SupplyChainComponent{
						Sources: []v1alpha1.ComponentReference{
							{
								Name:      "source-ref",
								Component: "source-output",
							},
						},
					}
					inputs := outs.GenerateInputs(component)
					Expect(inputs.Sources).To(HaveLen(1))
					Expect(inputs.Sources[0].Name).To(Equal("source-ref"))
					Expect(inputs.Sources[0].URL).To(Equal("source-url"))
					Expect(inputs.Sources[0].Revision).To(Equal("source-revision"))
				})
			})

			Context("And the sources do not have a match with the outputs", func() {
				It("Does not add sources to inputs", func() {
					component := &v1alpha1.SupplyChainComponent{
						Sources: []v1alpha1.ComponentReference{
							{
								Name:      "source-ref",
								Component: "source-output-does-not-exist",
							},
						},
					}
					inputs := outs.GenerateInputs(component)
					Expect(len(inputs.Sources)).To(Equal(0))
				})
			})
		})

		Context("When component contains images", func() {
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
					component := &v1alpha1.SupplyChainComponent{
						Images: []v1alpha1.ComponentReference{
							{
								Name:      "image-ref",
								Component: "image-output",
							},
						},
					}
					inputs := outs.GenerateInputs(component)
					Expect(inputs.Images).To(HaveLen(1))
					Expect(inputs.Images[0].Name).To(Equal("image-ref"))
					Expect(inputs.Images[0].Image).To(Equal("image12345"))
				})
			})

			Context("And the images do not have a match with the outputs", func() {
				It("Does not add images to inputs", func() {
					component := &v1alpha1.SupplyChainComponent{
						Images: []v1alpha1.ComponentReference{
							{
								Name:      "image-ref",
								Component: "image-output-does-not-exist",
							},
						},
					}
					inputs := outs.GenerateInputs(component)
					Expect(inputs.Sources).To(BeEmpty())
				})
			})

		})

		Context("When component contains configs", func() {
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
					component := &v1alpha1.SupplyChainComponent{
						Configs: []v1alpha1.ComponentReference{
							{
								Name:      "config-ref",
								Component: "config-output",
							},
						},
					}
					inputs := outs.GenerateInputs(component)
					Expect(inputs.Configs).To(HaveLen(1))
					Expect(inputs.Configs[0].Name).To(Equal("config-ref"))
					Expect(inputs.Configs[0].Config).To(Equal("config12345"))
				})
			})

			Context("And the configs do not have a match with the outputs", func() {
				It("Does not add configs to inputs", func() {
					component := &v1alpha1.SupplyChainComponent{
						Configs: []v1alpha1.ComponentReference{
							{
								Name:      "config-ref",
								Component: "config-output-does-not-exist",
							},
						},
					}
					inputs := outs.GenerateInputs(component)
					Expect(inputs.Configs).To(BeEmpty())
				})
			})
		})
	})
})
