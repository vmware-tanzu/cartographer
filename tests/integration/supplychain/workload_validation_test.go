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
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

var _ = Describe("WorkloadValidation", func() {
	var (
		ctx      context.Context
		workload *v1alpha1.Workload
	)

	var NamedWorkload = func(name string) *v1alpha1.Workload {
		return &v1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "default",
			},
		}
	}

	BeforeEach(func() {
		ctx = context.Background()
	})

	AfterEach(func() {
		_ = c.Delete(ctx, workload)
	})

	Context("Workload with bad name", func() {
		BeforeEach(func() {
			workload = NamedWorkload("java-web-app-2.6")
		})
		It("Rejects the workload", func() {
			err := c.Create(ctx, workload)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ContainSubstring("workload name is not a DNS 1035 label")))
		})
	})

	Context("Workload with okay name", func() {
		BeforeEach(func() {
			workload = NamedWorkload("java-web-app-2-6")
		})
		It("Accepts the workload", func() {
			err := c.Create(ctx, workload)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
