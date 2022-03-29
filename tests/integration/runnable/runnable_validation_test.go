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

package runnable_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

var _ = Describe("RunnableValidation", func() {
	var (
		ctx      context.Context
		runnable *v1alpha1.Runnable
	)

	var NamedRunnable = func(name string) *v1alpha1.Runnable {
		return &v1alpha1.Runnable{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "default",
			},
			Spec: v1alpha1.RunnableSpec{
				RunTemplateRef: v1alpha1.TemplateReference{
					Name: "run-template-name",
				},
				RetentionPolicy: v1alpha1.RetentionPolicy{
					MaxFailedRuns:     1,
					MaxSuccessfulRuns: 1,
				},
			},
		}
	}

	BeforeEach(func() {
		ctx = context.Background()
	})

	AfterEach(func() {
		_ = c.Delete(ctx, runnable)
	})

	Context("Runnable with bad name", func() {
		BeforeEach(func() {
			runnable = NamedRunnable("java-web-app-2.6")
		})
		It("Rejects the runnable", func() {
			err := c.Create(ctx, runnable)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ContainSubstring("name is not a DNS 1035 label")))
		})
	})

	Context("Runnable with okay name", func() {
		BeforeEach(func() {
			runnable = NamedRunnable("java-web-app-2-6")
		})
		It("Accepts the runnable", func() {
			err := c.Create(ctx, runnable)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
