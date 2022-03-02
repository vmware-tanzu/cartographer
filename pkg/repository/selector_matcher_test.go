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
		selectors         []repository.Selector
		expectedSelectors []repository.Selector
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
			selectors: []repository.Selector{
				&selector{
					labels: labels2.Set{
						"my": "label",
					},
					fields: fields{
						v1alpha1.FieldSelectorRequirement{
							Key:      "Spec.libertyGibbet",
							Operator: "Exists",
							Values:   nil,
						},
					},
				},
			},
			expectedSelectors: nil,
		}),

		Entry("partial match; selectors with less labels than selectors", testcase{
			selectable: selectable{
				labels: labels2.Set{
					"type": "web",
					"test": "tekton",
				}},
			selectors: []repository.Selector{
				&selector{
					labels: labels2.Set{
						"type": "web",
					},
					fields: fields{},
				},
			},
			expectedSelectors: []repository.Selector{
				&selector{
					labels: labels2.Set{
						"type": "web",
					},
					fields: fields{},
				},
			},
		}),

		Entry("partial match; selectors with less labels than target", testcase{
			selectable: selectable{
				labels: labels2.Set{
					"type": "web",
				}},
			selectors: []repository.Selector{
				&selector{
					labels: labels2.Set{
						"type": "web",
						"test": "tekton",
					},
					fields: fields{},
				},
			},
			expectedSelectors: nil,
		}),

		Entry("absolute match", testcase{
			selectable: selectable{
				labels: labels2.Set{
					"type": "web",
					"test": "tekton",
				}},
			selectors: []repository.Selector{
				&selector{
					labels: labels2.Set{
						"type": "web",
						"test": "webvalue",
					},
					fields: fields{},
				},
				&selector{
					labels: labels2.Set{ // ! this
						"type": "web",
						"test": "tekton",
					},
					fields: fields{},
				},
				&selector{
					labels: labels2.Set{
						"type": "mobile",
						"test": "mobilevalue",
					},
					fields: fields{},
				},
			},
			expectedSelectors: []repository.Selector{
				&selector{
					labels: labels2.Set{
						"type": "web",
						"test": "tekton",
					},
					fields: fields{},
				},
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
			selectors: []repository.Selector{
				&selector{
					labels: labels2.Set{
						"type": "atype",
						"test": "tekton",
						"scan": "ascan",
					},
					fields: fields{},
				},
				&selector{
					labels: labels2.Set{ // ! this
						"type": "web",
						"test": "tekton",
						"scan": "security",
					},
					fields: fields{},
				},
				&selector{
					labels: labels2.Set{ // ! this
						"type":  "web",
						"test":  "tekton",
						"input": "image",
					},
					fields: fields{},
				},
			},
			expectedSelectors: []repository.Selector{
				&selector{
					labels: labels2.Set{
						"type": "web",
						"test": "tekton",
						"scan": "security",
					},
					fields: fields{},
				},
				&selector{
					labels: labels2.Set{
						"type":  "web",
						"test":  "tekton",
						"input": "image",
					},
					fields: fields{},
				},
			},
		}),

		Entry("exact match with no extras", testcase{
			selectable: selectable{
				labels: labels2.Set{
					"type": "web",
					"test": "tekton",
					"scan": "security",
				}},
			selectors: []repository.Selector{
				&selector{
					labels: labels2.Set{
						"type": "atype",
						"test": "tekton",
						"scan": "ascan",
					},
					fields: fields{},
				},
				&selector{
					labels: labels2.Set{ // ! this
						"type": "web",
						"test": "tekton",
						"scan": "security",
					},
					fields: fields{},
				},
				&selector{
					labels: labels2.Set{
						"type":  "web",
						"test":  "tekton",
						"scan":  "security",
						"input": "image",
					},
					fields: fields{},
				},
			},
			expectedSelectors: []repository.Selector{
				&selector{
					labels: labels2.Set{
						"type": "web",
						"test": "tekton",
						"scan": "security",
					},
					fields: fields{},
				},
			},
		}),

		// ---------- Field Selectors
		// TODO: Comprehensive testing!? Eesh?

		Entry("match selectors with many fields in selectors", testcase{
			selectable: selectable{
				Spec: Spec{
					Color: "red",
					Age:   4,
				},
			},
			selectors: []repository.Selector{
				&selector{
					fields: fields{
						{
							Key:      "Spec.Color",
							Operator: "NotIn",
							Values:   []string{"green", "blue"},
						},
					},
				},
			},
			expectedSelectors: []repository.Selector{
				&selector{
					fields: fields{
						{
							Key:      "Spec.Color",
							Operator: "NotIn",
							Values:   []string{"green", "blue"},
						},
					},
				},
			},
		}),

		// TODO: Error cases and handling with field context
	)
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
	labels labels2.Set
	fields
}

func (b *selector) GetMatchExpressions() []metav1.LabelSelectorRequirement {
	return []metav1.LabelSelectorRequirement{}
}

func (b *selector) GetName() string {
	return "fred"
}

func (b *selector) GetMatchFields() []v1alpha1.FieldSelectorRequirement {
	return b.fields
}

func (b *selector) GetMatchLabels() labels2.Set {
	return b.labels
}
