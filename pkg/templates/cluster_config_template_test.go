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
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	"github.com/vmware-tanzu/cartographer/pkg/templates/templatesfakes"
)

var _ = Describe("ClusterConfigTemplate", func() {
	var (
		err            error
		configTemplate *v1alpha1.ClusterConfigTemplate
	)

	ItReturnsAHelpfulError := func(expectedErrorSubstring string) {
		It("returns a helpful error", func() {
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(expectedErrorSubstring))
		})
	}

	BeforeEach(func() {
		configTemplate = &v1alpha1.ClusterConfigTemplate{
			Spec: v1alpha1.ConfigTemplateSpec{
				ConfigPath: "some.path",
			},
		}
	})

	Describe("GetOutput", func() {
		var (
			output        *templates.Output
			stampedObject *unstructured.Unstructured
			evaluator     *templatesfakes.FakeEvaluator
		)

		BeforeEach(func() {
			stampedObject = &unstructured.Unstructured{}
			evaluator = &templatesfakes.FakeEvaluator{}
		})

		JustBeforeEach(func() {
			clusterConfigTemplateModel := templates.NewClusterConfigTemplateModel(configTemplate, evaluator)
			output, err = clusterConfigTemplateModel.GetOutput(stampedObject, nil)
		})

		When("passed a stamped object for which the evaluator can return a value at the configPath", func() {
			BeforeEach(func() {
				evaluator.EvaluateJsonPathReturns("some value", nil)
			})
			It("returns an appropriate output", func() {
				Expect(evaluator.EvaluateJsonPathCallCount()).To(Equal(1))
				path, obj := evaluator.EvaluateJsonPathArgsForCall(0)
				Expect(path).To(Equal("some.path"))
				Expect(obj).To(Equal(stampedObject.UnstructuredContent()))

				Expect(output.Config).To(Equal("some value"))
			})
		})
		When("passed a stamped object for which the evaluator cannot return a value at the configPath", func() {
			BeforeEach(func() {
				evaluator.EvaluateJsonPathReturns("", fmt.Errorf("some error"))
			})
			It("does return an output", func() {
				Expect(output).To(BeNil())
			})
			It("returns an error which identifies the failing json path expression", func() {
				jsonPathErr, ok := err.(templates.JsonPathError)
				Expect(ok).To(BeTrue())
				Expect(jsonPathErr.JsonPathExpression()).To(Equal("some.path"))
			})
			ItReturnsAHelpfulError("some error")
		})
	})
})
