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
	"github.com/vmware-tanzu/cartographer/pkg/realizer/workload"
)

var _ = FDescribe("ResourceStatuses", func() {
	var statuses workload.ResourceStatuses

	Context("a resourceStatuses with previous statuses", func() {
		BeforeEach(func() {
			previous := []v1alpha1.ResourceStatus{
				{
					RealizedResource: v1alpha1.RealizedResource{
						Name:        "resource1",
						StampedRef:  nil,
						TemplateRef: nil,
						Inputs:      nil,
						Outputs:     nil,
					},
					Conditions: nil,
				},
			}
			statuses = workload.NewResourceStatuses(previous)
		})

		Context("there are no new realized resources", func() {
			It("the resourceStatuses reports IsChanged is false", func() {
				Expect(statuses.IsChanged()).To(BeFalse())
			})
		})

		Context("there are new realized resources", func() {

		})
	})

	Context("a resourceStatuses with no previous statuses", func() {
		BeforeEach(func() {
			statuses = workload.NewResourceStatuses(nil)
		})

		Context("there are new realized resources", func() {
			BeforeEach(func() {
				statuses.Add(&v1alpha1.RealizedResource{
					Name:        "resource1",
					StampedRef:  nil,
					TemplateRef: nil,
					Inputs:      nil,
					Outputs:     nil,
				}, nil)
			})

			It("the resourceStatuses reports IsChanged is true", func() {
				Expect(statuses.IsChanged()).To(BeTrue())
			})
		})
	})
})
