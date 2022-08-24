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
	"context"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	v1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
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
				template := v1alpha1.TemplateSpec{
					Template: &runtime.RawExtension{
						Raw: []byte(`{ "kind": "Silly", "apiVersion": "silly.io/v1"}`),
					},
				}
				stamped, err := stamper.Stamp(context.TODO(), template)

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
				var template v1alpha1.TemplateSpec
				BeforeEach(func() {
					template = v1alpha1.TemplateSpec{
						Template: &runtime.RawExtension{
							Raw: []byte(`{ "kind": "Silly", "apiVersion": "silly.io/v1"}`),
						},
					}
				})

				It("sets the namespace to match the owner", func() {
					stamped, err := stamper.Stamp(context.TODO(), template)

					Expect(err).NotTo(HaveOccurred())
					Expect(stamped.GetNamespace()).To(Equal("owner-ns"))
				})
			})

			Context("template does specify a namespace", func() {
				var template v1alpha1.TemplateSpec

				Context("it is different than the owner namespace", func() {
					BeforeEach(func() {
						template = v1alpha1.TemplateSpec{
							Template: &runtime.RawExtension{
								Raw: []byte(`{
								"kind": "Silly",
								"apiVersion": "silly.io/v1",
								"metadata": { "namespace": "template-ns" }
							}`),
							},
						}
					})

					It("returns an error", func() {
						_, err := stamper.Stamp(context.TODO(), template)

						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("cannot set namespace in resource template"))
					})
				})

				Context("it is same as owner namespace", func() {
					BeforeEach(func() {
						template = v1alpha1.TemplateSpec{
							Template: &runtime.RawExtension{
								Raw: []byte(`{
								"kind": "Silly",
								"apiVersion": "silly.io/v1",
								"metadata": { "namespace": "owner-ns" }
							}`),
							},
						}
					})

					It("does not error", func() {
						stamped, err := stamper.Stamp(context.TODO(), template)

						Expect(err).NotTo(HaveOccurred())
						Expect(stamped.GetNamespace()).To(Equal("owner-ns"))
					})
				})
			})
		})

		DescribeTable("tag evaluation of template",
			func(tmpl string, subJSON string, expected interface{}, expectedErr string) {
				template := v1alpha1.TemplateSpec{
					Template: &runtime.RawExtension{
						Raw: []byte(`{ "kind": "Silly", "key": "` + tmpl + `"}`),
					},
				}
				params := templates.Params{
					"sub": {
						Raw: []byte(subJSON),
					},
					"extra-for-nested": {
						Raw: []byte(`"nested"`),
					},
					"infinite-recurse": {
						Raw: []byte(`"$(params.sub)$"`),
					},
					"bigger-infinite-recurse": {
						Raw: []byte(`"$(params.infinite-recurse)$"`),
					},
				}

				owner := &v1.ConfigMap{}

				templatingContext := struct {
					Params templates.Params `json:"params"`
				}{
					Params: params,
				}

				stamper := templates.StamperBuilder(owner, templatingContext, templates.Labels{})
				stampedUnstructured, err := stamper.Stamp(context.TODO(), template)
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
				`$(params.sub)$`, `"5"`, "5", ""),

			Entry(`Single tag, string value with nested tag`,
				`$(params.sub)$`, `"$(params.extra-for-nested)$"`, "nested", ""),

			Entry(`Single tag, number value and type preserved`,
				`$(params.sub)$`, `5`, float64(5), ""),

			Entry(`Single tag, map value and type preserved, nested tags evaluated`,
				`$(params.sub)$`, `{"foo": "$(params.extra-for-nested)$"}`, map[string]interface{}{"foo": "nested"}, ""),

			Entry(`Single tag, array value and type preserved, nested tags evaluated`,
				`$(params.sub)$`, `["foo", "$(params['extra-for-nested'])$"]`, []interface{}{"foo", "nested"}, ""),

			Entry(`Single tag, empty array value preserved`,
				`$(params.sub)$`, `["foo", []]`, []interface{}{"foo", []interface{}{}}, ""),

			Entry(`Multiple tags, result becomes a string`,
				`$(params.sub)$$(params.sub)$`, `5`, "55", ""),

			Entry(`Adjacent non-tag (letter), result becomes a string`,
				`b$(params.sub)$`, `5`, "b5", ""),

			Entry(`Adjacent non-tag (number), result still becomes a string`,
				`5$(params.sub)$`, `5`, "55", ""),

			Entry(`Adjacent non-tag, string value with nested tag`,
				`HI:$(params.sub)$`, `"$(params.extra-for-nested)$"`, "HI:nested", ""),

			Entry(`Looks like an array, but result must be preserved as string`,
				`[$(params.sub)$]`, `5`, "[5]", ""),

			Entry(`Looks like a map, but result must be preserved as string`,
				`{\"foo\": $(params.sub)$}`, `5`, `{"foo": 5}`, ""),

			Entry(`Infinite recursion should error`,
				`$(params.sub)$`, `"$(params.infinite-recurse)$"`, nil, "infinite tag loop detected: $(params.sub)$ -> $(params.infinite-recurse)$ -> $(params.sub)$"),

			Entry(`Infinite recursion should error`,
				`$(params.sub)$`, `"$(params.bigger-infinite-recurse)$"`, nil, "infinite tag loop detected: $(params.sub)$ -> $(params.bigger-infinite-recurse)$ -> $(params.infinite-recurse)$ -> $(params.sub)$"),

			Entry(`Infinite recursion with a map should error`,
				`$(params.sub)$`, `{"foo": "$(params.infinite-recurse)$"}`, nil, "infinite tag loop detected: $(params.sub)$ -> $(params.infinite-recurse)$ -> $(params.sub)$"),

			Entry(`Infinite recursion with an array should error`,
				`$(params.sub)$`, `["foo", "$(params.infinite-recurse)$"]`, nil, "infinite tag loop detected: $(params.sub)$ -> $(params.infinite-recurse)$ -> $(params.sub)$"),
		)

		DescribeTable("tag evaluation of ytt template",
			func(tmpl string, subJSON string, expected interface{}, koDataPath string, expectedErr string) {
				template := v1alpha1.TemplateSpec{
					Ytt: `
#@ load("@ytt:data", "data")


apiVersion: v1
kind: TestResource
key: ` + tmpl + `
`,
				}
				params := templates.Params{
					"sub": apiextensionsv1.JSON{Raw: []byte(subJSON)},
				}

				owner := &v1.ConfigMap{}

				templatingContext := struct {
					Params templates.Params `json:"params"`
				}{
					Params: params,
				}

				if koDataPath != "" {
					// set KO_DATA_PATH for this test, and then restore the previous value
					previous, set := os.LookupEnv("KO_DATA_PATH")
					defer func() {
						if set {
							os.Setenv("KO_DATA_PATH", previous)
						} else {
							os.Unsetenv("KO_DATA_PATH")
						}
					}()
					os.Setenv("KO_DATA_PATH", koDataPath)
				}

				stamper := templates.StamperBuilder(owner, templatingContext, templates.Labels{})
				stampedUnstructured, err := stamper.Stamp(context.TODO(), template)
				if expectedErr != "" {
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(ContainSubstring(expectedErr)))
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(stampedUnstructured.Object["key"]).To(Equal(expected))
				}
			},

			Entry(`String value and type preserved`,
				`#@ data.values.params.sub`, `"5"`, "5", "", ""),
			Entry(`Number value and type preserved`,
				`#@ data.values.params.sub`, `5`, int64(5), "", ""),
			Entry(`Map value and type preserved`,
				`#@ data.values.params.sub`, `{"foo": "bar"}`, map[string]interface{}{"foo": "bar"}, "", ""),

			Entry(`Invalid template`,
				"#@ data.values.invalid", `""`, nil, "", "unable to apply ytt template:"),
			Entry(`Invalid context`,
				"#@ data.values.params['sub']", `"`, nil, "", "unable to marshal template context:"),
			Entry(`Invalid ytt`,
				"#@ data.values.params['sub']", `""`, nil, "/not/a/path/to/ytt", "unable to apply ytt template: fork/exec"),
		)
	})
})
