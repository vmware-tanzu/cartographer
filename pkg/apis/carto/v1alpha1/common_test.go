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
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/cartographer/pkg/apis/carto/v1alpha1"
)

var _ = Describe("common", func() {
	Describe("GetAPITemplate", func() {

		DescribeTable("valid template kinds",
			func(templateKind string, expectedTemplate interface{}) {
				actualTemplate, err := v1alpha1.GetAPITemplate(templateKind)
				Expect(err).NotTo(HaveOccurred())
				Expect(reflect.TypeOf(actualTemplate)).To(Equal(reflect.TypeOf(expectedTemplate)))
			},
			Entry("ClusterSourceTemplate", "ClusterSourceTemplate", &v1alpha1.ClusterSourceTemplate{}),
			Entry("ClusterImageTemplate", "ClusterImageTemplate", &v1alpha1.ClusterImageTemplate{}),
			Entry("ClusterConfigTemplate", "ClusterConfigTemplate", &v1alpha1.ClusterConfigTemplate{}),
			Entry("ClusterTemplate", "ClusterTemplate", &v1alpha1.ClusterTemplate{}),
		)

		Context("unknown template kind", func() {
			It("returns an error", func() {
				_, err := v1alpha1.GetAPITemplate("Rubbish")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("resource does not have valid kind: Rubbish"))
			})

		})

	})
})
