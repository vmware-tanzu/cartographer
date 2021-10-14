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

package templates_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

var _ = Describe("Inputs", func() {
	var inputs templates.Inputs

	BeforeEach(func() {
		inputs = templates.Inputs{
			Sources: map[string]templates.SourceInput{},
			Images:  map[string]templates.ImageInput{},
			Configs: map[string]templates.ConfigInput{},
		}
	})

	Describe("OnlySource", func() {
		It("is nil when there are no sources", func() {
			Expect(inputs.OnlySource()).To(BeNil())
		})

		Context("when there is only one source", func() {
			BeforeEach(func() {
				inputs.Sources["one"] = templates.SourceInput{Name: "one"}
			})

			It("returns a pointer to that source", func() {
				Expect(*inputs.OnlySource()).To(Equal(inputs.Sources["one"]))
			})
		})

		Context("when there is more than one source", func() {
			BeforeEach(func() {
				inputs.Sources["one"] = templates.SourceInput{Name: "one"}
				inputs.Sources["deux"] = templates.SourceInput{Name: "deux"}
			})

			It("returns nil", func() {
				Expect(inputs.OnlySource()).To(BeNil())
			})
		})
	})

	Describe("OnlyImage", func() {
		It("is nil when there are no images", func() {
			Expect(inputs.OnlyImage()).To(BeNil())
		})

		Context("when there is only one image", func() {
			var actualImage interface{}
			BeforeEach(func() {
				actualImage = struct{ This string }{This: "actualImage could be *anything*, as this anonymous struct shows"}
				inputs.Images["one"] = templates.ImageInput{Name: "one", Image: actualImage}
			})

			It("returns that image's nested image", func() {
				Expect(inputs.OnlyImage()).To(Equal(actualImage))
			})
		})

		Context("when there is more than one image", func() {
			BeforeEach(func() {
				inputs.Images["one"] = templates.ImageInput{Name: "one", Image: "01"}
				inputs.Images["deux"] = templates.ImageInput{Name: "deux", Image: "10"}
			})

			It("returns nil", func() {
				Expect(inputs.OnlyImage()).To(BeNil())
			})
		})
	})

	Describe("OnlyConfig", func() {
		It("is nil when there are no configs", func() {
			Expect(inputs.OnlyConfig()).To(BeNil())
		})

		Context("when there is only one config", func() {
			var actualConfig interface{}
			BeforeEach(func() {
				actualConfig = struct{ ArbitraryThingVeryMuch string }{ArbitraryThingVeryMuch: "much like images, it could be anything."}
				inputs.Configs["one"] = templates.ConfigInput{Name: "one", Config: actualConfig}
			})

			It("returns that config's nested config", func() {
				Expect(inputs.OnlyConfig()).To(Equal(actualConfig))
			})
		})

		Context("when there is more than one config", func() {
			BeforeEach(func() {
				inputs.Configs["one"] = templates.ConfigInput{Name: "one", Config: "01"}
				inputs.Configs["deux"] = templates.ConfigInput{Name: "deux", Config: "10"}
			})

			It("returns nil", func() {
				Expect(inputs.OnlyConfig()).To(BeNil())
			})
		})
	})
})
