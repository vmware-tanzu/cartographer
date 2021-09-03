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
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	v1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"

	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

var _ = Describe("Stamper", func() {
	Describe("Stamp", func() {

		Describe("references to owner fields", func() {
			var stamper templates.Stamper

			BeforeEach(func() {
				owner := &v1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						UID:       "1234567890abcdef",
						Name:      "my-config-map",
						Namespace: "owner-ns",
					},
				}

				templatingContext := struct{}{}

				stamper = templates.StamperBuilder(owner, templatingContext, templates.Labels{})

			})
			It("sets the owner reference in the stamped output", func() {
				template := `{ "kind": "Silly", "apiVersion": "silly.io/v1"}`
				stamped, err := stamper.Stamp([]byte(template))

				Expect(err).NotTo(HaveOccurred())

				Expect(stamped.GetOwnerReferences()).To(HaveLen(1))
				owner := stamped.GetOwnerReferences()[0]

				Expect(owner).To(MatchFields(IgnoreExtras, Fields{
					"Name":               Equal("my-config-map"),
					"Kind":               Equal("ConfigMap"),
					"APIVersion":         Equal("v1"),
					"UID":                Equal(types.UID("1234567890abcdef")),
					"BlockOwnerDeletion": Equal(pointer.BoolPtr(true)),
					"Controller":         Equal(pointer.BoolPtr(true)),
				}))
			})

			Context("template does not specify a namespace", func() {
				var template string
				BeforeEach(func() {
					template = `{ "kind": "Silly", "apiVersion": "silly.io/v1"}`
				})

				It("sets the namespace to match the owner", func() {
					stamped, err := stamper.Stamp([]byte(template))

					Expect(err).NotTo(HaveOccurred())
					Expect(stamped.GetNamespace()).To(Equal("owner-ns"))
				})
			})
			Context("template does specify a namespace", func() {
				var template string
				BeforeEach(func() {
					template = `{
						"kind": "Silly",
						"apiVersion": "silly.io/v1",
						"metadata": { "namespace": "template-ns" }
					}`
				})

				It("does not change the namespace", func() {
					stamped, err := stamper.Stamp([]byte(template))

					Expect(err).NotTo(HaveOccurred())
					Expect(stamped.GetNamespace()).To(Equal("template-ns"))

				})
			})
		})

		DescribeTable("tag evaluation of template",
			func(template, subJSON string, expected interface{}, expectedErr string) {
				template = `{ "kind": "Silly", "key": "` + template + `"}`
				params := templates.Params{
					{
						Name: "sub",
						Value: apiextensionsv1.JSON{
							Raw: []byte(subJSON),
						},
					},
					{
						Name: "extra-for-nested",
						Value: apiextensionsv1.JSON{
							Raw: []byte(`"nested"`),
						},
					},
					{
						Name: "infinite-recurse",
						Value: apiextensionsv1.JSON{
							Raw: []byte(`"$(params[0].value)$"`),
						},
					},
					{
						Name: "bigger-infinite-recurse",
						Value: apiextensionsv1.JSON{
							Raw: []byte(`"$(params[2].value)$"`),
						},
					},
				}

				owner := &v1.ConfigMap{}

				templatingContext := struct {
					Params templates.Params `json:"params"`
				}{
					Params: params,
				}

				stamper := templates.StamperBuilder(owner, templatingContext, templates.Labels{})
				stampedUnstructured, err := stamper.Stamp([]byte(template))
				if expectedErr != "" {
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(ContainSubstring(expectedErr)))
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(stampedUnstructured.Object["key"]).To(Equal(expected))
				}
			},

			Entry(`Single empty tag yields an empty string`,
				`$()$`, `"some-value"`, "", "empty jsonpath not allowed"),

			Entry(`Not parsable tag, jsonpath contract test`,
				`$(,)$`, `"some-value"`, "", "unrecognized character in action"),

			Entry(`Not parsable tag, jsonpath contract test`,
				`$($()$`, `"some-value"`, "", "unrecognized character in action"),

			Entry(`Single tag, string value and type preserved`,
				`$(params[0].value)$`, `"5"`, "5", ""),

			Entry(`Single tag, string value with nested tag`,
				`$(params[0].value)$`, `"$(params[1].value)$"`, "nested", ""),

			Entry(`Single tag, number value and type preserved`,
				`$(params[0].value)$`, `5`, float64(5), ""),

			Entry(`Single tag, map value and type preserved, nested tags evaluated`,
				`$(params[0].value)$`, `{"foo": "$(params[1].value)$"}`, map[string]interface{}{"foo": "nested"}, ""),

			Entry(`Single tag, array value and type preserved, nested tags evaluated`,
				`$(params[0].value)$`, `["foo", "$(params[1].value)$"]`, []interface{}{"foo", "nested"}, ""),

			Entry(`Multiple tags, result becomes a string`,
				`$(params[0].value)$$(params[0].value)$`, `5`, "55", ""),

			Entry(`Adjacent non-tag (letter), result becomes a string`,
				`b$(params[0].value)$`, `5`, "b5", ""),

			Entry(`Adjacent non-tag (number), result still becomes a string`,
				`5$(params[0].value)$`, `5`, "55", ""),

			Entry(`Adjacent non-tag, string value with nested tag`,
				`HI:$(params[0].value)$`, `"$(params[1].value)$"`, "HI:nested", ""),

			Entry(`Looks like an array, but result must be preserved as string`,
				`[$(params[0].value)$]`, `5`, "[5]", ""),

			Entry(`Looks like a map, but result must be preserved as string`,
				`{\"foo\": $(params[0].value)$}`, `5`, `{"foo": 5}`, ""),

			Entry(`Infinite recursion should error`,
				`$(params[0].value)$`, `"$(params[2].value)$"`, nil, "infinite tag loop detected: $(params[0].value)$ -> $(params[2].value)$ -> $(params[0].value)$"),

			Entry(`Infinite recursion should error`,
				`$(params[0].value)$`, `"$(params[3].value)$"`, nil, "infinite tag loop detected: $(params[0].value)$ -> $(params[3].value)$ -> $(params[2].value)$ -> $(params[0].value)$"),

			Entry(`Infinite recursion with a map should error`,
				`$(params[0].value)$`, `{"foo": "$(params[2].value)$"}`, nil, "infinite tag loop detected: $(params[0].value)$ -> $(params[2].value)$ -> $(params[0].value)$"),

			Entry(`Infinite recursion with an array should error`,
				`$(params[0].value)$`, `["foo", "$(params[2].value)$"]`, nil, "infinite tag loop detected: $(params[0].value)$ -> $(params[2].value)$ -> $(params[0].value)$"),
		)
	})
})
