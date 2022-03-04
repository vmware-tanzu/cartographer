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

package repository_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels2 "k8s.io/apimachinery/pkg/labels"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
)

var _ = Describe("BestSelectorMatch", func() {

	type testcase struct {
		selectable        repository.Selectable
		selectors         []repository.SelectingObject
		expectedSelectors []repository.SelectingObject
	}

	DescribeTable("cases",
		func(tc testcase) {
			actual, _ := repository.BestSelectorMatch(
				tc.selectable, tc.selectors,
			)

			if tc.expectedSelectors == nil {
				Expect(actual).To(BeNil())
				return
			}

			Expect(actual).To(Equal(tc.expectedSelectors))
		},

		// ---------- Label Selectors

		Entry("empty selectors", testcase{
			selectable: selectable{
				labels: labels2.Set{}},
			expectedSelectors: nil,
		}),

		Entry("complete mismatched selectors & selectors", testcase{
			selectable: selectable{
				labels: labels2.Set{
					"type": "web",
				},
			},
			selectors: []repository.SelectingObject{
				newSelector(
					labels2.Set{
						"my": "label",
					},
					nil,
					fields{
						v1alpha1.FieldSelectorRequirement{
							Key:      "Spec.libertyGibbet",
							Operator: "Exists",
							Values:   nil,
						},
					},
				),
			},
			expectedSelectors: nil,
		}),

		Entry("partial match; selectors with less labels than selectors", testcase{
			selectable: selectable{
				labels: labels2.Set{
					"type": "web",
					"test": "tekton",
				}},
			selectors: []repository.SelectingObject{
				newSelector(
					labels2.Set{
						"type": "web",
					},
					nil,
					nil,
				),
			},
			expectedSelectors: []repository.SelectingObject{
				newSelector(
					labels2.Set{
						"type": "web",
					},
					nil,
					nil,
				),
			},
		}),

		Entry("partial match; selectors with less labels than target", testcase{
			selectable: selectable{
				labels: labels2.Set{
					"type": "web",
				}},
			selectors: []repository.SelectingObject{
				newSelector(
					labels2.Set{
						"type": "web",
						"test": "tekton",
					},
					nil,
					nil,
				),
			},
			expectedSelectors: nil,
		}),

		Entry("absolute match", testcase{
			selectable: selectable{
				labels: labels2.Set{
					"type": "web",
					"test": "tekton",
				}},
			selectors: []repository.SelectingObject{
				newSelector(
					labels2.Set{
						"type": "web",
						"test": "webvalue",
					},
					nil,
					nil,
				),
				newSelector(
					labels2.Set{
						"type": "web",
						"test": "tekton",
					},
					nil,
					nil,
				),
				newSelector(
					labels2.Set{
						"type": "mobile",
						"test": "mobilevalue",
					},
					nil,
					nil,
				),
			},
			expectedSelectors: []repository.SelectingObject{
				newSelector(
					labels2.Set{
						"type": "web",
						"test": "tekton",
					},
					nil,
					nil,
				),
			},
		}),

		Entry("exact partial match", testcase{
			selectable: selectable{
				labels: labels2.Set{
					"type":  "web",
					"test":  "tekton",
					"scan":  "security",
					"input": "image",
				}},
			selectors: []repository.SelectingObject{
				newSelector(
					labels2.Set{
						"type": "atype",
						"test": "tekton",
						"scan": "ascan",
					},
					nil,
					nil,
				),
				newSelector(
					labels2.Set{
						"type": "web",
						"test": "tekton",
						"scan": "security",
					},
					nil,
					nil,
				),
				newSelector(
					labels2.Set{
						"type":  "web",
						"test":  "tekton",
						"input": "image",
					},
					nil,
					nil,
				),
			},
			expectedSelectors: []repository.SelectingObject{
				newSelector(
					labels2.Set{
						"type": "web",
						"test": "tekton",
						"scan": "security",
					},
					nil,
					nil,
				),
				newSelector(
					labels2.Set{
						"type":  "web",
						"test":  "tekton",
						"input": "image",
					},
					nil,
					nil,
				),
			},
		}),

		Entry("exact match with no extras", testcase{
			selectable: selectable{
				labels: labels2.Set{
					"type": "web",
					"test": "tekton",
					"scan": "security",
				}},
			selectors: []repository.SelectingObject{
				newSelector(
					labels2.Set{
						"type": "atype",
						"test": "tekton",
						"scan": "ascan",
					},
					nil,
					nil,
				),
				newSelector(
					labels2.Set{
						"type": "web",
						"test": "tekton",
						"scan": "security",
					},
					nil,
					nil,
				),
				newSelector(
					labels2.Set{
						"type":  "web",
						"test":  "tekton",
						"scan":  "security",
						"input": "image",
					},
					nil,
					nil,
				),
			},
			expectedSelectors: []repository.SelectingObject{
				newSelector(
					labels2.Set{
						"type": "web",
						"test": "tekton",
						"scan": "security",
					},
					nil,
					nil,
				),
			},
		}),

		Entry("match selectors with many fields in selectors", testcase{
			selectable: selectable{
				Spec: Spec{
					Color: "red",
					Age:   4,
				},
			},
			selectors: []repository.SelectingObject{
				newSelector(
					nil,
					nil,
					fields{
						{
							Key:      "Spec.Color",
							Operator: "NotIn",
							Values:   []string{"green", "blue"},
						},
					},
				),
			},
			expectedSelectors: []repository.SelectingObject{
				newSelector(
					nil,
					nil,
					fields{
						{
							Key:      "Spec.Color",
							Operator: "NotIn",
							Values:   []string{"green", "blue"},
						},
					},
				),
			},
		}),
	)

	Describe("malformed selectors", func() {
		Context("label selector invalid", func() {
			var sel []repository.SelectingObject
			BeforeEach(func() {
				sel = []repository.SelectingObject{
					newSelectorWithID(
						"my-selector",
						"Special",
						labels2.Set{
							"fred-": "derf-",
						},
						nil,
						nil,
					),
				}
			})

			It("returns an error", func() {
				_, err := repository.BestSelectorMatch(selectable{}, sel)
				Expect(err).To(MatchError(ContainSubstring("selectorMatchExpressions or selectors of [Special/my-selector] are not valid")))
				Expect(err).To(MatchError(ContainSubstring("key: Invalid value")))
			})
		})

		Context("expression selector invalid", func() {
			var sel []repository.SelectingObject
			BeforeEach(func() {
				sel = []repository.SelectingObject{
					newSelectorWithID(
						"my-selector",
						"Special",
						nil,
						[]metav1.LabelSelectorRequirement{
							{
								Key:      "fred",
								Operator: "Matchingest",
								Values:   nil,
							},
						},
						nil,
					),
				}
			})

			It("returns an error", func() {
				_, err := repository.BestSelectorMatch(selectable{}, sel)
				Expect(err).To(MatchError(ContainSubstring("selectorMatchExpressions or selectors of [Special/my-selector] are not valid")))
				// TODO: 'pod' - Hmmmmm - perhaps we shouldn't be using v1 code?
				Expect(err).To(MatchError(ContainSubstring("\"Matchingest\" is not a valid pod selector operator")))
			})
		})
	})
})

type fields []v1alpha1.FieldSelectorRequirement

type Spec struct {
	Color string `json:"color"`
	Age   int    `json:"age"`
}

type selectable struct {
	labels map[string]string
	Spec   `json:"spec"`
}

func (o selectable) GetLabels() map[string]string {
	return o.labels
}

type selector struct {
	metav1.TypeMeta
	metav1.ObjectMeta
	v1alpha1.Selectors
}

func newSelector(labels labels2.Set, expressions []metav1.LabelSelectorRequirement, fields []v1alpha1.FieldSelectorRequirement) *selector {
	return newSelectorWithID("testSelectingObject", "Test", labels, expressions, fields)
}

func newSelectorWithID(name, kind string, labels labels2.Set, expressions []metav1.LabelSelectorRequirement, fields []v1alpha1.FieldSelectorRequirement) *selector {
	return &selector{
		TypeMeta: metav1.TypeMeta{
			Kind:       kind,
			APIVersion: "testv1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Selectors: v1alpha1.Selectors{
			Selector:                 labels,
			SelectorMatchExpressions: expressions,
			SelectorMatchFields:      fields,
		},
	}
}

func (b *selector) GetSelectors() v1alpha1.Selectors {
	return b.Selectors
}
