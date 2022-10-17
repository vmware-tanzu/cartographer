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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

var _ = Describe("Templates", func() {
	var (
		err           error
		apiTemplate   client.Object
		templateModel templates.Reader
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

	JustBeforeEach(func() {
		templateModel, err = templates.NewReaderFromAPI(apiTemplate)
	})

	Describe("NewReaderFromAPI", func() {
		Context("when passed a ClusterSourceTemplate", func() {
			BeforeEach(func() {
				apiTemplate = &v1alpha1.ClusterSourceTemplate{}
			})

			ItDoesNotReturnAnError()

			It("returns a template", func() {
				Expect(templateModel).NotTo(BeNil())
			})
		})
		Context("when passed a ClusterImageTemplate", func() {
			BeforeEach(func() {
				apiTemplate = &v1alpha1.ClusterImageTemplate{}
			})

			ItDoesNotReturnAnError()

			It("returns a template", func() {
				Expect(templateModel).NotTo(BeNil())
			})
		})
		Context("when passed a ClusterConfigTemplate", func() {
			BeforeEach(func() {
				apiTemplate = &v1alpha1.ClusterConfigTemplate{}
			})

			ItDoesNotReturnAnError()

			It("returns a template", func() {
				Expect(templateModel).NotTo(BeNil())
			})
		})

		Context("when passed a ClusterTemplate", func() {
			BeforeEach(func() {
				apiTemplate = &v1alpha1.ClusterTemplate{}
			})

			ItDoesNotReturnAnError()

			It("returns a template", func() {
				Expect(templateModel).NotTo(BeNil())
			})
		})

		Context("when passed an unsupported object", func() {
			BeforeEach(func() {
				apiTemplate = &v1alpha1.Workload{}
			})

			ItReturnsAHelpfulError("resource does not match a known template")
		})
	})
})
