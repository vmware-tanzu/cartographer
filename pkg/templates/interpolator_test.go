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
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/valyala/fasttemplate"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/vmware-tanzu/cartographer/pkg/eval"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	"github.com/vmware-tanzu/cartographer/pkg/templates/templatesfakes"
	. "github.com/vmware-tanzu/cartographer/pkg/utils/matchers"
)

type GenericType struct {
	Name  string      `json:"name"`
	Count int         `json:"count"`
	Empty interface{} `json:"empty"`
	List  []string    `json:"list"`
}

var _ = Describe("Interpolator", func() {
	Describe("InterpolateLeafNode Stubbing executor", func() {
		var (
			template        []byte
			tagInterpolator templatesfakes.FakeTagInterpolator
			executor        templates.TemplateExecutor
		)

		BeforeEach(func() {
			template = []byte("some-template")
		})

		Context("when executor returns an error", func() {
			BeforeEach(func() {
				executor = func(template, startTag, endTag string, f fasttemplate.TagFunc) (string, error) {
					return "", fmt.Errorf("some error")
				}
			})

			It("returns an error", func() {
				_, err := templates.InterpolateLeafNode(executor, template, &tagInterpolator)
				Expect(err).To(BeMeaningful("interpolate tag:"))
			})
		})

		Context("when executor returns a stable result", func() {
			BeforeEach(func() {
				executor = func(template, startTag, endTag string, f fasttemplate.TagFunc) (string, error) {
					return "some result", nil
				}
			})

			It("returns the result as a byte array", func() {
				result, err := templates.InterpolateLeafNode(executor, template, &tagInterpolator)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal("some result"))
			})
		})
	})

	Describe("InterpolateLeafNode with fasttemplate and Stubbing TagInterpolator", func() {
		var (
			template        []byte
			tagInterpolator templatesfakes.FakeTagInterpolator
		)

		BeforeEach(func() {
			tagInterpolator = templatesfakes.FakeTagInterpolator{}
			template = []byte("some-template-with-$(some-tag)$")
		})

		Context("When the tag interpolator returns an error", func() {
			BeforeEach(func() {
				tagInterpolator.InterpolateTagReturns(0, fmt.Errorf("some error"))
			})
			It("returns an error", func() {
				_, err := templates.InterpolateLeafNode(fasttemplate.ExecuteFuncStringWithErr, template, &tagInterpolator)
				Expect(err).To(BeMeaningful("interpolate tag: "))
			})
		})
	})

	Describe("InterpolateLeafNode with fasttemplate and StandardTagInterpolator and eval.Evaluator", func() {
		var (
			template        []byte
			tagInterpolator templates.StandardTagInterpolator
		)

		BeforeEach(func() {
			tagInterpolator = templates.StandardTagInterpolator{
				Evaluator: eval.EvaluatorBuilder(),
			}

			tagInterpolator.Context = struct {
				Params  templates.Params `json:"params"`
				Generic GenericType      `json:"generic"`
			}{

				Params: templates.Params{
					"an-amazing-param": apiextensionsv1.JSON{Raw: []byte(`"exactly what you want"`)},
					"another_param":    apiextensionsv1.JSON{Raw: []byte(`"everything you need"`)},
				},
				Generic: GenericType{
					Name:  "generic-name",
					Count: 99,
					List:  []string{"one", "two"},
				},
			}
		})

		Context("given a template with no tags to interpolate", func() {
			BeforeEach(func() {
				template = []byte("hello, this is dog")
			})

			It("returns the same byte array", func() {
				returnedInterpolatedTemplate, err := templates.InterpolateLeafNode(fasttemplate.ExecuteFuncStringWithErr, template, tagInterpolator)
				Expect(err).NotTo(HaveOccurred())
				Expect(returnedInterpolatedTemplate).To(Equal(string(template)))
			})
		})

		Context("given a template with an empty tag", func() {
			BeforeEach(func() {
				template = []byte("Look at this empty tag ---> $()$")
			})

			It("Returns an error explaining that empty jsonpath is not allowed", func() {
				_, err := templates.InterpolateLeafNode(fasttemplate.ExecuteFuncStringWithErr, template, tagInterpolator)
				Expect(err).To(BeMeaningful("interpolate tag: "))
				Expect(err).To(BeMeaningful("empty jsonpath not allowed"))
			})
		})

		Context("given a template with a tag for an unknown field in the stamp context", func() {
			BeforeEach(func() {
				template = []byte("I've never heard of a $(snarfblatt.name)$")
			})

			It("Returns an error that something went wrong in evaluating jsonpath", func() {
				_, err := templates.InterpolateLeafNode(fasttemplate.ExecuteFuncStringWithErr, template, tagInterpolator)
				Expect(err).To(BeMeaningful("interpolate tag: "))
				Expect(err).To(BeMeaningful("evaluate jsonpath: "))
			})
		})

		Context("given a template with a tag for an unknown subfield in the stamp context", func() {
			BeforeEach(func() {
				template = []byte("Generic doesn't have $(generic.vacationLoad)$")
			})

			It("Returns an error that something went wrong in evaluating jsonpath", func() {
				_, err := templates.InterpolateLeafNode(fasttemplate.ExecuteFuncStringWithErr, template, tagInterpolator)
				Expect(err).To(BeMeaningful("interpolate tag: "))
				Expect(err).To(BeMeaningful("evaluate jsonpath: "))
			})
		})

		Context("given a tag pointing to a field that is empty", func() {
			BeforeEach(func() {
				template = []byte("this generic does not have an env: $(generic.empty)$ <-- so this shouldn't work")
			})

			It("Returns an error that a tag points to a nil value", func() {
				_, err := templates.InterpolateLeafNode(fasttemplate.ExecuteFuncStringWithErr, template, tagInterpolator)
				Expect(err).To(BeMeaningful("interpolate tag: "))
				Expect(err).To(BeMeaningful("tag must not point to nil value: generic.empty"))
			})
		})

		Context("given a tag pointing to a string field that can be interpolated", func() {
			BeforeEach(func() {
				template = []byte("this is the place to put the name: $(generic.name)$ <-- see it there?")
			})

			It("returns the proper string", func() {
				interpolatedTemplate, err := templates.InterpolateLeafNode(fasttemplate.ExecuteFuncStringWithErr, template, tagInterpolator)

				Expect(err).NotTo(HaveOccurred())
				Expect(interpolatedTemplate).To(Equal("this is the place to put the name: generic-name <-- see it there?"))
			})
		})

		Context("given there are multiple tags", func() {
			BeforeEach(func() {
				template = []byte("this is the place to put the name: $(generic.name)$ and the count: $(generic.count)$")
			})

			It("returns the proper string", func() {
				interpolatedTemplate, err := templates.InterpolateLeafNode(fasttemplate.ExecuteFuncStringWithErr, template, tagInterpolator)

				Expect(err).NotTo(HaveOccurred())
				Expect(interpolatedTemplate).To(Equal("this is the place to put the name: generic-name and the count: 99"))
			})
		})

		Context("and a tag that refers to a list", func() {
			BeforeEach(func() {
				template = []byte("this is the place to put the list: $(generic.list)$ <-- see it there?")
			})

			It("returns the proper string", func() {
				interpolatedTemplate, err := templates.InterpolateLeafNode(fasttemplate.ExecuteFuncStringWithErr, template, tagInterpolator)

				Expect(err).NotTo(HaveOccurred())
				Expect(interpolatedTemplate).To(Equal(
					`this is the place to put the list: ["one","two"] <-- see it there?`,
				))
			})
		})

		Context("given a template referencing a missing list element", func() {
			BeforeEach(func() {
				template = []byte("in an empty input, you won't find $(params[0])$")
			})

			It("Returns an error that it cannot evaluate the path", func() {
				_, err := templates.InterpolateLeafNode(fasttemplate.ExecuteFuncStringWithErr, template, tagInterpolator)
				Expect(err).To(BeMeaningful("interpolate tag: "))
				Expect(err).To(BeMeaningful("evaluate jsonpath: "))
			})
		})

		Context("given a template referencing a list element by index", func() {
			BeforeEach(func() {
				template = []byte("with the populated input, you can find $(params.another_param)$")
			})

			It("returns the proper string", func() {
				interpolatedTemplate, err := templates.InterpolateLeafNode(fasttemplate.ExecuteFuncStringWithErr, template, tagInterpolator)

				Expect(err).NotTo(HaveOccurred())
				Expect(interpolatedTemplate).To(Equal(
					"with the populated input, you can find everything you need",
				))
			})
		})

		Context("given a template referencing a list element by attribute", func() {
			BeforeEach(func() {
				template = []byte(`with the populated input, you can find $(params['an-amazing-param'])$`)
			})

			It("returns the proper string", func() {
				interpolatedTemplate, err := templates.InterpolateLeafNode(fasttemplate.ExecuteFuncStringWithErr, template, tagInterpolator)

				Expect(err).NotTo(HaveOccurred())
				Expect(interpolatedTemplate).To(Equal(
					"with the populated input, you can find exactly what you want",
				))
			})
		})

		Context("given a template referencing a list element by attribute but an unknown key", func() {
			BeforeEach(func() {
				template = []byte(`with the populated input, you can find $(params['notknown])$`)
			})

			It("Returns an error that it cannot find the value notknown", func() {
				_, err := templates.InterpolateLeafNode(fasttemplate.ExecuteFuncStringWithErr, template, tagInterpolator)
				Expect(err).To(BeMeaningful("interpolate tag: "))
				Expect(err).To(BeMeaningful("evaluate jsonpath: evaluate: failed to parse jsonpath '{.params['notknown]}': invalid array index 'notknown"))
			})
		})
	})
})

var _ = Describe("StandardTagInterpolator", func() {
	var (
		evaluator               templatesfakes.FakeEvaluator
		standardTagInterpolator templates.StandardTagInterpolator
	)

	BeforeEach(func() {
		evaluator = templatesfakes.FakeEvaluator{}

		standardTagInterpolator = templates.StandardTagInterpolator{
			Context:   struct{}{},
			Evaluator: &evaluator,
		}
	})

	Describe("InterpolateTag", func() {
		var (
			writer templatesfakes.FakeWriter
			tag    string
		)

		Context("with a tag", func() {
			BeforeEach(func() {
				tag = "some tag"
			})

			Context("when the evaluator returns an error", func() {
				BeforeEach(func() {
					evaluator.EvaluateJsonPathReturns("", fmt.Errorf("some error"))
				})

				It("Returns a jsonpath error", func() {
					_, err := standardTagInterpolator.InterpolateTag(&writer, tag)
					Expect(err).To(BeMeaningful("evaluate jsonpath: "))
				})
			})

			Context("when the evaluator returns a nil value", func() {
				BeforeEach(func() {
					evaluator.EvaluateJsonPathReturns(nil, nil)
				})

				It("Returns a jsonpath error", func() {
					_, err := standardTagInterpolator.InterpolateTag(&writer, tag)
					Expect(err).To(BeMeaningful("tag must not point to nil value: "))
				})
			})

			Context("when the evaluator returns a string", func() {
				BeforeEach(func() {
					writer = templatesfakes.FakeWriter{}
					mockWriteLen := 123
					writer.WriteReturns(mockWriteLen, nil)

					evaluator.EvaluateJsonPathReturns("some value", nil)
				})

				It("writes the value to the writer", func() {
					_, err := standardTagInterpolator.InterpolateTag(&writer, tag)
					Expect(err).NotTo(HaveOccurred())

					byteArray := writer.WriteArgsForCall(0)
					Expect(byteArray).To(Equal([]byte("some value")))
				})

				It("passes back the length from the writer", func() {
					writeLen, err := standardTagInterpolator.InterpolateTag(&writer, tag)
					Expect(err).NotTo(HaveOccurred())

					Expect(writeLen).To(Equal(123))
				})

				Context("and the writer fails to write", func() {
					BeforeEach(func() {
						writer.WriteReturns(0, fmt.Errorf("some error"))
					})

					It("Returns a writer error", func() {
						_, err := standardTagInterpolator.InterpolateTag(&writer, tag)
						Expect(err).To(BeMeaningful("writer write: "))
					})
				})
			})

			Context("when the evaluator returns a non string object", func() {
				BeforeEach(func() {
					writer = templatesfakes.FakeWriter{}
					mockWriteLen := 123
					writer.WriteReturns(mockWriteLen, nil)

					evaluator.EvaluateJsonPathReturns(3, nil)
				})

				It("calls the writer with a json representation of the object", func() {
					_, err := standardTagInterpolator.InterpolateTag(&writer, tag)
					Expect(err).NotTo(HaveOccurred())

					byteArray := writer.WriteArgsForCall(0)
					Expect(byteArray).To(Equal([]byte("3")))
				})

				It("passes back the length from the writer", func() {
					writeLen, err := standardTagInterpolator.InterpolateTag(&writer, tag)
					Expect(err).NotTo(HaveOccurred())

					Expect(writeLen).To(Equal(123))
				})
			})
		})
	})
})
