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

package eval_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/cartographer/pkg/eval"
	"github.com/vmware-tanzu/cartographer/pkg/eval/evalfakes"
)

var _ = Describe("JsonPath", func() {
	var (
		err       error
		evaluator eval.Evaluator
		evaluate  evalfakes.FakeEvaluate
		path      string
		result    interface{}
	)

	ItDoesNotReturnAnError := func() {
		It("does not return an error", func() {
			Expect(err).ShouldNot(HaveOccurred())
		})
	}

	ItReturnsAHelpfulError := func(expectedErrorSubstring string) {
		It("returns a helpful error", func() {
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(expectedErrorSubstring))
		})
	}

	BeforeEach(func() {
		path = "some.path"
		evaluate = evalfakes.FakeEvaluate{}
		evaluator = eval.Evaluator{
			Evaluate: evaluate.Spy,
		}
	})

	Describe("EvaluateJsonPath", func() {
		var (
			obj interface{}
		)

		DescribeTable("ensures valid wrapping",
			func(tablePath, expectedCall string) {
				obj = "some object"
				evaluate.Returns(nil, fmt.Errorf("some short circuiting error"))
				_, _ = evaluator.EvaluateJsonPath(tablePath, obj)
				pathCallValue, objCallValue := evaluate.ArgsForCall(0)
				Expect(pathCallValue).To(Equal(expectedCall))
				Expect(objCallValue).To(Equal(obj))
			},
			Entry("when wrapping characters are missing", "validity", "{.validity}"),
			Entry("when leading characters are missing", "validity}", "{.validity}"),
			Entry("when braces are missing", ".validity", "{.validity}"),
			Entry("when trailing character is missing", "{.validity", "{.validity}"),
			Entry("when input is complete", "{.validity}", "{.validity}"),
		)

		Context("when evaluate returns an error", func() {
			BeforeEach(func() {
				evaluate.Returns(nil, fmt.Errorf("some error"))
				result, err = evaluator.EvaluateJsonPath(path, obj)
			})

			ItReturnsAHelpfulError("evaluate: ")
		})

		Context("when evaluate returns a list of multiple items", func() {
			BeforeEach(func() {
				evaluate.Returns([]interface{}{"some error"}, nil)
				result, err = evaluator.EvaluateJsonPath(path, obj)
			})

			ItDoesNotReturnAnError()

			It("returns that single item", func() {
				Expect(result).To(Equal("some error"))
			})
		})

		Context("when evaluate returns a list of no items", func() {
			BeforeEach(func() {
				evaluate.Returns([]interface{}{}, nil)
				result, err = evaluator.EvaluateJsonPath(path, obj)
			})

			ItReturnsAHelpfulError("jsonpath returned empty list")
		})

		Context("when evaluate returns a list with a single item", func() {
			BeforeEach(func() {
				evaluate.Returns([]interface{}{"some error"}, nil)
				result, err = evaluator.EvaluateJsonPath(path, obj)
			})

			ItDoesNotReturnAnError()

			It("returns that single item", func() {
				Expect(result).To(Equal("some error"))
			})
		})

		Context("when path is empty", func() {
			BeforeEach(func() {
				path = ""
				result, err = evaluator.EvaluateJsonPath(path, obj)
			})

			ItReturnsAHelpfulError("empty jsonpath not allowed")
		})
	})
})
