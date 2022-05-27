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

package selector_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/selector"
)

var _ = Describe("Selector", func() {

	Context("when the key exists in the context", func() {
		var context map[string]interface{}

		BeforeEach(func() {
			context = map[string]interface{}{
				"hello": "world",
				"empty": nil,
			}
		})

		Context("when the operator is Exists", func() {
			Context("when the key is valid and returns a value", func() {
				It("returns true and does not error", func() {
					req := v1alpha1.FieldSelectorRequirement{
						Key:      "hello",
						Operator: v1alpha1.FieldSelectorOpExists,
					}
					ret, err := selector.Matches(req, context)
					Expect(err).ToNot(HaveOccurred())
					Expect(ret).To(BeTrue())
				})
			})

			Context("when the key is valid but returns a nil value", func() {
				It("returns false and does not error", func() {
					req := v1alpha1.FieldSelectorRequirement{
						Key:      "empty",
						Operator: v1alpha1.FieldSelectorOpExists,
					}
					ret, err := selector.Matches(req, context)
					Expect(err).ToNot(HaveOccurred())
					Expect(ret).To(BeFalse())
				})
			})
		})

		Context("when the operator is DoesNotExist", func() {
			Context("when the key is valid and returns a value", func() {
				It("returns true and does not error", func() {
					req := v1alpha1.FieldSelectorRequirement{
						Key:      "bad",
						Operator: v1alpha1.FieldSelectorOpDoesNotExist,
					}
					ret, err := selector.Matches(req, context)
					Expect(err).ToNot(HaveOccurred())
					Expect(ret).To(BeTrue())
				})
			})

			Context("when the key is valid but returns a nil value", func() {
				It("returns false and does not error", func() {
					req := v1alpha1.FieldSelectorRequirement{
						Key:      "hello",
						Operator: v1alpha1.FieldSelectorOpDoesNotExist,
					}
					ret, err := selector.Matches(req, context)
					Expect(err).ToNot(HaveOccurred())
					Expect(ret).To(BeFalse())
				})
			})
		})

		Context("when the operator is In", func() {
			Context("when the key is valid and returns a value", func() {
				It("returns true and does not error", func() {
					req := v1alpha1.FieldSelectorRequirement{
						Key:      "hello",
						Operator: v1alpha1.FieldSelectorOpIn,
						Values:   []string{"world"},
					}
					ret, err := selector.Matches(req, context)
					Expect(err).ToNot(HaveOccurred())
					Expect(ret).To(BeTrue())
				})
			})

			Context("when the key is valid but not in values", func() {
				It("returns false and does not error", func() {
					req := v1alpha1.FieldSelectorRequirement{
						Key:      "hello",
						Operator: v1alpha1.FieldSelectorOpIn,
						Values:   []string{"planet"},
					}
					ret, err := selector.Matches(req, context)
					Expect(err).ToNot(HaveOccurred())
					Expect(ret).To(BeFalse())
				})
			})
		})

		Context("when the operator is NotIn", func() {
			Context("when the key is valid and returns a value", func() {
				It("returns true and does not error", func() {
					req := v1alpha1.FieldSelectorRequirement{
						Key:      "hello",
						Operator: v1alpha1.FieldSelectorOpNotIn,
						Values:   []string{"planet"},
					}
					ret, err := selector.Matches(req, context)
					Expect(err).ToNot(HaveOccurred())
					Expect(ret).To(BeTrue())
				})
			})

			Context("when the key is valid but not in values", func() {
				It("returns false and does not error", func() {
					req := v1alpha1.FieldSelectorRequirement{
						Key:      "hello",
						Operator: v1alpha1.FieldSelectorOpNotIn,
						Values:   []string{"world"},
					}
					ret, err := selector.Matches(req, context)
					Expect(err).ToNot(HaveOccurred())
					Expect(ret).To(BeFalse())
				})
			})
		})

		Context("when the operator is not valid", func() {
			It("returns an error", func() {
				req := v1alpha1.FieldSelectorRequirement{
					Key:      "hello",
					Operator: "NotValid",
				}
				ret, err := selector.Matches(req, context)
				Expect(err).To(HaveOccurred())
				Expect(ret).To(BeFalse())
			})
		})

	})

	Context("when the key is empty in the context", func() {
		var context map[string]interface{}

		BeforeEach(func() {
			context = map[string]interface{}{
				"hello": "",
			}
		})

		Context("when the operator is Exists", func() {
			Context("when the key is valid and returns a value", func() {
				It("returns true and does not error", func() {
					req := v1alpha1.FieldSelectorRequirement{
						Key:      "hello",
						Operator: v1alpha1.FieldSelectorOpExists,
					}
					ret, err := selector.Matches(req, context)
					Expect(err).ToNot(HaveOccurred())
					Expect(ret).To(BeTrue())
				})
			})
		})
	})

	Context("when the key does not exist in the context", func() {
		var context map[string]interface{}

		BeforeEach(func() {
			context = map[string]interface{}{
				"goodbye": "world",
			}
		})

		It("returns false and returns an error", func() {
			req := v1alpha1.FieldSelectorRequirement{
				Key: "hello",
			}

			ret, err := selector.Matches(req, context)
			Expect(err).To(HaveOccurred())
			Expect(ret).To(BeFalse())
		})

	})

})
