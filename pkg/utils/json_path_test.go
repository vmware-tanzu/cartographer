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

package utils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

var _ = Describe("JsonPath", func() {
	var (
		err    error
		path   string
		obj    map[string]string
		result []interface{}
	)

	ItReturnsAHelpfulError := func(expectedErrorSubstring string) {
		It("returns a helpful error", func() {
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(expectedErrorSubstring))
		})
	}

	ItDoesNotReturnAnError := func() {
		It("does not return an error", func() {
			Expect(err).ShouldNot(HaveOccurred())
		})
	}

	BeforeEach(func() {
		obj = make(map[string]string)
		obj["hello"] = "there"
	})

	JustBeforeEach(func() {
		result, err = utils.SinglePathEvaluate(path, obj)
	})

	Context("when path is malformed", func() {
		BeforeEach(func() {
			path = "{{"
		})

		ItReturnsAHelpfulError("failed to parse jsonpath '{{': ")
	})

	Context("when there are two queries are in the path", func() {
		BeforeEach(func() {
			path = `{.hello}{.hello}`
		})

		ItReturnsAHelpfulError("more queries than expected: ")
	})

	Context("when find fails", func() {
		BeforeEach(func() {
			path = `{.hello[1]}`
		})

		ItReturnsAHelpfulError("find results: ")
	})

	Context("when find succeeds", func() {
		BeforeEach(func() {
			path = `{.hello}`
		})

		ItDoesNotReturnAnError()

		It("returns the object found", func() {
			Expect(result).To(Equal([]interface{}{"there"}))
		})
	})
})
