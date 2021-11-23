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

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
)

var _ = Describe("BestLabelMatches", func() {

	type testcase struct {
		source   repository.LabelsGetter
		targets  []repository.SelectorGetter
		expected []repository.SelectorGetter
	}

	type labels map[string]string

	var lg = func(labelset labels) repository.LabelsGetter {
		return &v1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labelset,
			},
		}
	}

	var sg = func(labelset labels) repository.SelectorGetter {
		return &v1alpha1.ClusterSupplyChain{
			Spec: v1alpha1.SupplyChainSpec{
				Selector: labelset,
			},
		}
	}

	DescribeTable("cases",
		func(tc testcase) {
			actual := repository.BestLabelMatches(
				tc.source, tc.targets,
			)

			if tc.expected == nil {
				Expect(actual).To(BeNil())
				return
			}

			Expect(actual).To(Equal(tc.expected))
		},

		Entry("empty targets", testcase{
			source:   lg(labels{"foo": "bar"}),
			expected: nil,
		}),

		Entry("complete mismatched src & targets", testcase{
			source: lg(labels{
				"type": "web",
			}),
			targets: []repository.SelectorGetter{
				sg(labels{
					"----": "----",
				}),
			},
			expected: nil,
		}),

		Entry("partial match; target with less labels than source", testcase{
			source: lg(labels{
				"type": "web",
				"test": "tekton",
			}),
			targets: []repository.SelectorGetter{
				sg(labels{
					"type": "web",
				}),
			},
			expected: []repository.SelectorGetter{
				sg(labels{
					"type": "web",
				}),
			},
		}),

		Entry("partial match; source with less labels than target", testcase{
			source: lg(labels{
				"type": "web",
			}),
			targets: []repository.SelectorGetter{
				sg(labels{
					"type": "web",
					"test": "tekton",
				}),
			},
			expected: nil,
		}),

		Entry("absolute match", testcase{
			source: lg(labels{
				"type": "web",
				"test": "tekton",
			}),
			targets: []repository.SelectorGetter{
				sg(labels{
					"type": "web",
					"test": "----",
				}),
				sg(labels{ // ! this
					"type": "web",
					"test": "tekton",
				}),
				sg(labels{
					"type": "mobile",
					"test": "----",
				}),
			},
			expected: []repository.SelectorGetter{
				sg(labels{
					"type": "web",
					"test": "tekton",
				}),
			},
		}),

		Entry("exact partial match", testcase{
			source: lg(labels{
				"type":  "web",
				"test":  "tekton",
				"scan":  "security",
				"input": "image",
			}),
			targets: []repository.SelectorGetter{
				sg(labels{
					"type": "----",
					"test": "tekton",
					"scan": "----",
				}),
				sg(labels{ // ! this
					"type": "web",
					"test": "tekton",
					"scan": "security",
				}),
				sg(labels{ // ! this
					"type":  "web",
					"test":  "tekton",
					"input": "image",
				}),
			},
			expected: []repository.SelectorGetter{
				sg(labels{
					"type": "web",
					"test": "tekton",
					"scan": "security",
				}),
				sg(labels{
					"type":  "web",
					"test":  "tekton",
					"input": "image",
				}),
			},
		}),

		Entry("exact match with no extras", testcase{
			source: lg(labels{
				"type": "web",
				"test": "tekton",
				"scan": "security",
			}),
			targets: []repository.SelectorGetter{
				sg(labels{
					"type": "----",
					"test": "tekton",
					"scan": "----",
				}),
				sg(labels{ // ! this
					"type": "web",
					"test": "tekton",
					"scan": "security",
				}),
				sg(labels{
					"type":  "web",
					"test":  "tekton",
					"scan":  "security",
					"input": "image",
				}),
			},
			expected: []repository.SelectorGetter{
				sg(labels{
					"type": "web",
					"test": "tekton",
					"scan": "security",
				}),
			},
		}),
	)
})
