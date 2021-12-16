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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/vmware-tanzu/cartographer/pkg/controller/runnable"
)

var _ = Describe("Conditions", func() {

	Describe("MissingValueAtPath", func() {
		var obj *unstructured.Unstructured
		BeforeEach(func() {
			obj = &unstructured.Unstructured{}
			obj.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "thing.io",
				Version: "alphabeta1",
				Kind:    "Widget",
			})
			obj.SetName("my-widget")
		})

		Context("stamped object has a namespace", func() {
			It("has the correct message", func() {
				obj.SetNamespace("my-ns")

				condition := runnable.OutputPathNotSatisfiedCondition(obj, "problem at spec.foo")
				Expect(condition.Message).To(Equal("waiting to read value from resource [widget.thing.io/my-widget] in namespace [my-ns]: problem at spec.foo"))
			})
		})

		Context("stamped object does not have a namespace", func() {
			It("has the correct message", func() {
				condition := runnable.OutputPathNotSatisfiedCondition(obj, "problem at spec.foo")
				Expect(condition.Message).To(Equal("waiting to read value from resource [widget.thing.io/my-widget]: problem at spec.foo"))
			})
		})
	})
})
