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

package selector_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels2 "k8s.io/apimachinery/pkg/labels"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/selector"
)

var _ = Describe("BestSelectorMatchIndices", func() {

	type testcase struct {
		selectable              selector.Selectable
		selectors               []v1alpha1.Selector
		expectedSelectorIndices []int
	}

	DescribeTable(
		"non error-cases",
		func(tc testcase) {
			actual, err := selector.BestSelectorMatchIndices(
				tc.selectable, tc.selectors,
			)

			if tc.expectedSelectorIndices == nil {
				Expect(actual).To(BeNil())
				return
			}

			Expect(err).NotTo(HaveOccurred())

			Expect(actual).To(Equal(tc.expectedSelectorIndices))
		},

		Entry("no selectors", testcase{
			selectable: selectable{
				labels: labels2.Set{}},
			expectedSelectorIndices: nil,
		}),

		Entry("complete mismatched label & field selectors", testcase{
			selectable: selectable{
				labels: labels2.Set{
					"type": "web",
				},
			},
			selectors: []v1alpha1.Selector{
				{
					metav1.LabelSelector{
						MatchLabels: labels2.Set{
							"my": "label",
						},
						MatchExpressions: nil,
					},
					[]v1alpha1.FieldSelectorRequirement{
						{
							Key:      "Spec.libertyGibbet",
							Operator: "Exists",
							Values:   nil,
						},
					},
				},
			},
			expectedSelectorIndices: nil,
		}),

		Entry("partial match; selectors with less labels than selectors", testcase{
			selectable: selectable{
				labels: labels2.Set{
					"type": "web",
					"test": "tekton",
				}},
			selectors: []v1alpha1.Selector{
				{
					metav1.LabelSelector{
						MatchLabels: labels2.Set{
							"type": "web",
						},
						MatchExpressions: nil,
					},
					nil,
				},
			},
			expectedSelectorIndices: []int{0},
		}),

		Entry("partial match; selectors with less labels than target", testcase{
			selectable: selectable{
				labels: labels2.Set{
					"type": "web",
				}},
			selectors: []v1alpha1.Selector{
				{
					metav1.LabelSelector{
						MatchLabels: labels2.Set{
							"type": "web",
							"test": "tekton",
						},
					},
					nil,
				},
			},
			expectedSelectorIndices: nil,
		}),

		Entry("absolute match", testcase{
			selectable: selectable{
				labels: labels2.Set{
					"type": "web",
					"test": "tekton",
				}},
			selectors: []v1alpha1.Selector{
				{
					metav1.LabelSelector{
						MatchLabels: labels2.Set{
							"type": "web",
							"test": "webvalue",
						},
					},
					nil,
				},
				{
					metav1.LabelSelector{
						MatchLabels: labels2.Set{
							"type": "web",
							"test": "tekton",
						},
					},
					nil,
				},
				{
					metav1.LabelSelector{
						MatchLabels: labels2.Set{
							"type": "mobile",
							"test": "mobilevalue",
						},
					},
					nil,
				},
			},
			expectedSelectorIndices: []int{1},
		}),

		Entry("exact partial match", testcase{
			selectable: selectable{
				labels: labels2.Set{
					"type":  "web",
					"test":  "tekton",
					"scan":  "security",
					"input": "image",
				}},
			selectors: []v1alpha1.Selector{
				{
					metav1.LabelSelector{
						MatchLabels: labels2.Set{
							"type": "atype",
							"test": "tekton",
							"scan": "ascan",
						},
					},
					nil,
				},
				{
					metav1.LabelSelector{
						MatchLabels: labels2.Set{
							"type": "web",
							"test": "tekton",
							"scan": "security",
						},
					},
					nil,
				},
				{
					metav1.LabelSelector{
						MatchLabels: labels2.Set{
							"type":  "web",
							"test":  "tekton",
							"input": "image",
						},
					},
					nil,
				},
			},
			expectedSelectorIndices: []int{1, 2},
		}),

		Entry("exact match with no extras", testcase{
			selectable: selectable{
				labels: labels2.Set{
					"type": "web",
					"test": "tekton",
					"scan": "security",
				}},
			selectors: []v1alpha1.Selector{
				{
					metav1.LabelSelector{
						MatchLabels: labels2.Set{
							"type": "atype",
							"test": "tekton",
							"scan": "ascan",
						},
					},
					nil,
				},
				{
					metav1.LabelSelector{
						MatchLabels: labels2.Set{
							"type": "web",
							"test": "tekton",
							"scan": "security",
						},
					},
					nil,
				},
				{
					metav1.LabelSelector{
						MatchLabels: labels2.Set{
							"type":  "web",
							"test":  "tekton",
							"scan":  "security",
							"input": "image",
						},
					},
					nil,
				},
			},
			expectedSelectorIndices: []int{1},
		}),

		Entry("match selectors with many fields in selectors", testcase{
			selectable: selectable{
				Spec: Spec{
					Color: "red",
					Age:   4,
				},
			},
			selectors: []v1alpha1.Selector{
				{
					MatchFields: []v1alpha1.FieldSelectorRequirement{
						{
							Key:      "Spec.Color",
							Operator: "NotIn",
							Values:   []string{"green", "blue"},
						},
					},
				},
			},
			expectedSelectorIndices: []int{0},
		}),

		Entry("match selectors when json path is not found, don't error", testcase{
			selectable: selectable{
				Spec: Spec{
					Color:  "red",
					Age:    4,
					Bucket: []Thing{{Name: "morko"}},
				},
			},
			selectors: []v1alpha1.Selector{
				{
					MatchFields: []v1alpha1.FieldSelectorRequirement{
						{
							Key:      "spec.bucket[?(@.name==\"marco\")].name",
							Operator: "NotIn",
							Values:   []string{"green", "blue"},
						},
					},
				},
			},
			expectedSelectorIndices: nil,
		}),
		//FIXME: should this case matter? it's broken
		//FEntry("match selectors when json path is not found, don't error", testcase{
		//	selectable: selectable{
		//		Spec: Spec{
		//			Color: "red",
		//			Age:   4,
		//			Map:   map[string]string{"name": "morko"},
		//		},
		//	},
		//
		//
		//	selectors: []v1alpha1.Selector{
		//		{
		//			MatchFields: []v1alpha1.FieldSelectorRequirement{
		//				{
		//					Key:      "spec.map.marco",
		//					Operator: "NotIn",
		//					Values:   []string{"green", "blue"},
		//				},
		//			},
		//		},
		//	},
		//	expectedSelectorIndices: []int{},
		//}),
	)

	Describe("malformed selectors", func() {
		Context("label selector invalid", func() {
			var sel []v1alpha1.Selector
			BeforeEach(func() {
				sel = []v1alpha1.Selector{
					{
						LabelSelector: metav1.LabelSelector{
							MatchLabels: labels2.Set{
								"this-label": "is-valid",
							},
							MatchExpressions: nil,
						},
					},
					{
						LabelSelector: metav1.LabelSelector{
							MatchLabels: labels2.Set{
								"this-one-": "sure-is-not",
							},
							MatchExpressions: nil,
						},
					},
				}
			})

			It("returns an error", func() {
				_, err := selector.BestSelectorMatchIndices(selectable{}, sel)
				Expect(err).To(MatchError(ContainSubstring("selector labels or matchExpressions are not valid")))
				Expect(err).To(MatchError(ContainSubstring("key: Invalid value")))
				Expect(err.SelectorIndex()).To(Equal(1))
			})
		})

		Context("expression selector invalid", func() {
			var sel []v1alpha1.Selector
			BeforeEach(func() {
				sel = []v1alpha1.Selector{
					{
						LabelSelector: metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "valid",
									Operator: "Exists",
									Values:   nil,
								},
							},
						},
					},
					{
						LabelSelector: metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "not-valid",
									Operator: "Matchingest",
									Values:   nil,
								},
							},
						},
					},
				}
			})

			It("returns an error", func() {
				_, err := selector.BestSelectorMatchIndices(selectable{}, sel)
				Expect(err).To(MatchError(ContainSubstring("selector labels or matchExpressions are not valid")))
				Expect(err).To(MatchError(ContainSubstring("\"Matchingest\" is not a valid label selector operator")))
				Expect(err.SelectorIndex()).To(Equal(1))
			})
		})

		Context("match fields invalid", func() {
			var sel []v1alpha1.Selector
			BeforeEach(func() {
				sel = []v1alpha1.Selector{
					{
						MatchFields: []v1alpha1.FieldSelectorRequirement{
							{
								Key:      "spec",
								Operator: "Exists",
								Values:   nil,
							},
						},
					},
					{
						MatchFields: []v1alpha1.FieldSelectorRequirement{
							{
								Key:      "spec",
								Operator: "MostAwesomeNonExistentOperator",
								Values:   nil,
							},
						},
					},
				}
			})

			It("returns an error", func() {
				_, err := selector.BestSelectorMatchIndices(selectable{}, sel)
				Expect(err).To(MatchError(ContainSubstring("failed to evaluate selector matchFields")))
				Expect(err).To(MatchError(ContainSubstring("unable to match field requirement with key [spec] operator [MostAwesomeNonExistentOperator] values [[]]: invalid operator MostAwesomeNonExistentOperator")))
				Expect(err.SelectorIndex()).To(Equal(1))
			})
		})
	})
})

type Spec struct {
	Color  string            `json:"color"`
	Age    int               `json:"age"`
	Bucket []Thing           `json:"bucket"`
	Map    map[string]string `json:"map"`
}

type Thing struct {
	Name string `json:"name"`
}

type selectable struct {
	labels map[string]string
	Spec   `json:"spec"`
}

func (o selectable) GetLabels() map[string]string {
	return o.labels
}
