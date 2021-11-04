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

var _ = Describe("ClusterSourceTemplate", func() {
	var (
		err                   error
		urlPath, revisionPath string
		sourceTemplate        *v1alpha1.ClusterSourceTemplate
	)

	ItReturnsAHelpfulError := func(expectedErrorSubstring string) {
		It("returns a helpful error", func() {
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(expectedErrorSubstring))
		})
	}

	BeforeEach(func() {
		urlPath = "some.url.path"
		revisionPath = "some.revision.path"

		sourceTemplate = &v1alpha1.ClusterSourceTemplate{
			Spec: v1alpha1.SourceTemplateSpec{
				URLPath:      urlPath,
				RevisionPath: revisionPath,
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
			clusterSourceTemplateModel := templates.NewClusterSourceTemplateModel(sourceTemplate, evaluator)
			output, err = clusterSourceTemplateModel.GetOutput(stampedObject, nil)
		})

		When("passed a stamped object for which the evaluator can return a value at the urlPath and revisionPath", func() {
			BeforeEach(func() {
				evaluator.EvaluateJsonPathStub = func(path string, obj interface{}) (interface{}, error) {
					if path == urlPath {
						return "some value", nil
					} else if path == revisionPath {
						return "some other value", nil
					} else {
						return "", fmt.Errorf("unexpected error")
					}
				}
			})
			It("returns an appropriate output", func() {
				Expect(evaluator.EvaluateJsonPathCallCount()).To(Equal(2))

				path, obj := evaluator.EvaluateJsonPathArgsForCall(0)
				Expect(path).To(Equal(urlPath))
				Expect(obj).To(Equal(stampedObject.UnstructuredContent()))

				path, obj = evaluator.EvaluateJsonPathArgsForCall(1)
				Expect(path).To(Equal(revisionPath))
				Expect(obj).To(Equal(stampedObject.UnstructuredContent()))

				Expect(*output.Source).To(Equal(templates.Source{
					URL:      "some value",
					Revision: "some other value",
				}))
				Expect(err).To(BeNil())
			})
		})
		When("passed a stamped object for which the evaluator cannot return a value at the urlPath and revisionPath", func() {
			BeforeEach(func() {
				evaluator.EvaluateJsonPathReturns("", fmt.Errorf("some error"))
			})
			It("does not return an output", func() {
				Expect(output).To(BeNil())
			})
			It("returns an error which identifies the failing json path expression", func() {
				jsonPathErr, ok := err.(templates.JsonPathError)
				Expect(ok).To(BeTrue())
				Expect(jsonPathErr.JsonPathExpression()).To(Equal("some.url.path"))
			})
			ItReturnsAHelpfulError("some error")
		})
	})
})
