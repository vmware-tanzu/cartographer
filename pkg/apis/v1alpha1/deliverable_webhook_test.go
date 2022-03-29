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

package v1alpha1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

var _ = Describe("Deliverable Webhook Validation", func() {
	var deliverable *v1alpha1.Deliverable
	BeforeEach(func() {
		deliverable = &v1alpha1.Deliverable{}
	})

	Context("Deliverable has a name", func() {
		Context("the name is bad", func() {
			BeforeEach(func() {
				deliverable.Name = "java-web-app-2.6"
			})
			It("rejects the deliverable", func() {
				Expect(deliverable.ValidateCreate()).To(MatchError(ContainSubstring("name is not a DNS 1035 label")))
			})
		})
		Context("the name is good", func() {
			BeforeEach(func() {
				deliverable.Name = "java-web-app-2-6"
			})
			It("accepts the deliverable", func() {
				Expect(deliverable.ValidateCreate()).NotTo(HaveOccurred())
			})
		})

	})

	Context("Deliverable has a generateName", func() {
		Context("the generateName is bad", func() {
			BeforeEach(func() {
				deliverable.GenerateName = "java-web-app-2.6"
			})
			It("rejects the deliverable", func() {
				Expect(deliverable.ValidateCreate()).To(MatchError(ContainSubstring("generateName is not a DNS 1035 label prefix")))
			})
		})
		Context("the generateName is good", func() {
			BeforeEach(func() {
				deliverable.GenerateName = "java-web-app-2-6"
			})
			It("accepts the deliverable", func() {
				Expect(deliverable.ValidateCreate()).NotTo(HaveOccurred())
			})
		})
	})

	Context("Deliverable does not have a name or a generateName", func() {
		It("rejects the deliverable", func() {
			Expect(deliverable.ValidateCreate()).To(MatchError(ContainSubstring("name or generateName is required")))
		})
	})
})
