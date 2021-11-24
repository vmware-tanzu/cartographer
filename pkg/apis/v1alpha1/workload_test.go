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

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

var _ = Describe("Workload", func() {
	Describe("Workload Spec", func() {
		var (
			workloadSpec     v1alpha1.WorkloadSpec
			workloadSpecType reflect.Type
		)

		BeforeEach(func() {
			workloadSpecType = reflect.TypeOf(workloadSpec)
		})

		It("allows but does not require service account name", func() {
			metadataField, found := workloadSpecType.FieldByName("ServiceAccountName")
			Expect(found).To(BeTrue())
			jsonValue := metadataField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("serviceAccountName"))
			Expect(jsonValue).To(ContainSubstring("omitempty"))
		})

		It("allows but does not require params", func() {
			metadataField, found := workloadSpecType.FieldByName("Params")
			Expect(found).To(BeTrue())
			jsonValue := metadataField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("params"))
			Expect(jsonValue).To(ContainSubstring("omitempty"))
		})

		It("allows but does not require source", func() {
			sourceField, found := workloadSpecType.FieldByName("Source")
			Expect(found).To(BeTrue())
			jsonValue := sourceField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("source"))
			Expect(jsonValue).To(ContainSubstring("omitempty"))
		})

		It("allows but does not require serviceclaims", func() {
			serviceClaimsField, found := workloadSpecType.FieldByName("ServiceClaims")
			Expect(found).To(BeTrue())
			jsonValue := serviceClaimsField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("serviceClaims"))
			Expect(jsonValue).To(ContainSubstring("omitempty"))
		})

		It("allows but does not require env", func() {
			envField, found := workloadSpecType.FieldByName("Env")
			Expect(found).To(BeTrue())
			jsonValue := envField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("env"))
			Expect(jsonValue).To(ContainSubstring("omitempty"))
		})

		It("allows but does not require resources", func() {
			resourcesField, found := workloadSpecType.FieldByName("Resources")
			Expect(found).To(BeTrue())
			jsonValue := resourcesField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("resources"))
			Expect(jsonValue).To(ContainSubstring("omitempty"))
		})
	})

	Describe("Workload Source", func() {
		var (
			workloadSource     v1alpha1.Source
			workloadSourceType reflect.Type
		)

		BeforeEach(func() {
			workloadSourceType = reflect.TypeOf(workloadSource)
		})

		It("allows but does not require git", func() {
			gitField, found := workloadSourceType.FieldByName("Git")
			Expect(found).To(BeTrue())
			jsonValue := gitField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("git"))
			Expect(jsonValue).To(ContainSubstring("omitempty"))
		})

		It("allows but does not require image", func() {
			imageField, found := workloadSourceType.FieldByName("Image")
			Expect(found).To(BeTrue())
			jsonValue := imageField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("image"))
			Expect(jsonValue).To(ContainSubstring("omitempty"))
		})

		It("allows but does not require a subpath", func() {
			subpathField, found := workloadSourceType.FieldByName("Subpath")
			Expect(found).To(BeTrue())
			jsonValue := subpathField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("subPath"))
			Expect(jsonValue).To(ContainSubstring("omitempty"))
		})
	})

	Describe("Workload Git", func() {
		var (
			workloadGit     v1alpha1.GitSource
			workloadGitType reflect.Type
		)

		BeforeEach(func() {
			workloadGitType = reflect.TypeOf(workloadGit)
		})

		It("allows but does not require url", func() {
			urlField, found := workloadGitType.FieldByName("URL")
			Expect(found).To(BeTrue())
			jsonValue := urlField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("url"))
			Expect(jsonValue).To(ContainSubstring("omitempty"))
		})

		It("allows but does not require ref", func() {
			refField, found := workloadGitType.FieldByName("Ref")
			Expect(found).To(BeTrue())
			jsonValue := refField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("ref"))
			Expect(jsonValue).To(ContainSubstring("omitempty"))
		})
	})

	Describe("Workload Git Ref", func() {
		var (
			workloadGitRef     v1alpha1.GitRef
			workloadGitRefType reflect.Type
		)

		BeforeEach(func() {
			workloadGitRefType = reflect.TypeOf(workloadGitRef)
		})

		It("allows but does not require branch", func() {
			branchField, found := workloadGitRefType.FieldByName("Branch")
			Expect(found).To(BeTrue())
			jsonValue := branchField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("branch"))
			Expect(jsonValue).To(ContainSubstring("omitempty"))
		})

		It("allows but does not require tag", func() {
			tagField, found := workloadGitRefType.FieldByName("Tag")
			Expect(found).To(BeTrue())
			jsonValue := tagField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("tag"))
			Expect(jsonValue).To(ContainSubstring("omitempty"))
		})

		It("allows but does not require commit", func() {
			commitField, found := workloadGitRefType.FieldByName("Commit")
			Expect(found).To(BeTrue())
			jsonValue := commitField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("commit"))
			Expect(jsonValue).To(ContainSubstring("omitempty"))
		})
	})

	Describe("Workload Param", func() {
		var (
			workloadParam     v1alpha1.Param
			workloadParamType reflect.Type
		)

		BeforeEach(func() {
			workloadParamType = reflect.TypeOf(workloadParam)
		})

		It("requires name", func() {
			nameField, found := workloadParamType.FieldByName("Name")
			Expect(found).To(BeTrue())
			jsonValue := nameField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("name"))
			Expect(jsonValue).NotTo(ContainSubstring("omitempty"))
		})

		It("requires value", func() {
			valueField, found := workloadParamType.FieldByName("Value")
			Expect(found).To(BeTrue())
			jsonValue := valueField.Tag.Get("json")
			Expect(jsonValue).To(ContainSubstring("value"))
			Expect(jsonValue).NotTo(ContainSubstring("omitempty"))
		})
	})
})
