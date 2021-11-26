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
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/cartographer/pkg/apis/carto/v1alpha1"
)

var _ = Describe("Runnable", func() {
	Describe("RunnableSpec", func() {
		var (
			runnableSpec     v1alpha1.RunnableSpec
			runnableSpecType reflect.Type
		)

		BeforeEach(func() {
			runnableSpecType = reflect.TypeOf(runnableSpec)
		})

		It("requires runTemplate", func() {
			resourcesField, found := runnableSpecType.FieldByName("RunTemplateRef")
			Expect(found).To(BeTrue())
			jsonValue := resourcesField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("runTemplate"))
			Expect(jsonValue).NotTo(ContainSubstring("omitempty"))
		})
	})

	Describe("TemplateReference", func() {
		var (
			templateReference     v1alpha1.TemplateReference
			templateReferenceType reflect.Type
		)

		BeforeEach(func() {
			templateReferenceType = reflect.TypeOf(templateReference)
		})

		It("requires a name", func() {
			nameField, found := templateReferenceType.FieldByName("Name")
			Expect(found).To(BeTrue())
			jsonValue := nameField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("name"))
			Expect(jsonValue).NotTo(ContainSubstring("omitempty"))
		})

		It("requires a kind", func() {
			kindField, found := templateReferenceType.FieldByName("Kind")
			Expect(found).To(BeTrue())
			jsonValue := kindField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("kind"))
			Expect(jsonValue).To(ContainSubstring("omitempty"))
		})
	})
})
