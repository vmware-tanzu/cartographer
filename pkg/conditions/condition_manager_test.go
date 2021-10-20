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

package conditions_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/cartographer/pkg/conditions"
)

var _ = Describe("conditionManager", func() {
	var manager conditions.ConditionManager
	Context("without any conditions added", func() {
		BeforeEach(func() {
			manager = conditions.NewConditionManager("HappyParent", []metav1.Condition{})
		})

		It("returns a top level unknown", func() {
			result := manager.Finalize()

			Expect(manager.IsSuccessful()).To(BeFalse())
			Expect(result).To(HaveLen(1))
			Expect(result).To(ContainElement(MatchFields(IgnoreExtras,
				Fields{
					"Type":   Equal("HappyParent"),
					"Status": Equal(metav1.ConditionUnknown),
				},
			)))
		})
	})

	Context("with positive polarity conditions", func() {
		Context("with successful conditions", func() {
			BeforeEach(func() {
				manager = conditions.NewConditionManager("HappyParent", []metav1.Condition{})
				goodnessCondition := metav1.Condition{
					Type:   "Goodness",
					Status: metav1.ConditionTrue,
				}
				manager.AddPositive(goodnessCondition)

				greatnessCondition := metav1.Condition{
					Type:   "Greatness",
					Status: metav1.ConditionTrue,
				}

				manager.AddPositive(greatnessCondition)
			})

			It("returns the conditions and a successful parent", func() {
				result := manager.Finalize()

				Expect(manager.IsSuccessful()).To(BeTrue())
				Expect(result).To(HaveLen(3))
				Expect(result).To(ContainElements(
					MatchFields(IgnoreExtras,
						Fields{
							"Type":   Equal("Goodness"),
							"Status": Equal(metav1.ConditionTrue),
						},
					),
					MatchFields(IgnoreExtras,
						Fields{
							"Type":   Equal("Greatness"),
							"Status": Equal(metav1.ConditionTrue),
						},
					),
					MatchFields(IgnoreExtras,
						Fields{
							"Type":   Equal("HappyParent"),
							"Status": Equal(metav1.ConditionTrue),
						},
					),
				))
			})
		})

		Context("with a failing condition", func() {
			BeforeEach(func() {
				manager = conditions.NewConditionManager("HappyParent", []metav1.Condition{})
				goodnessCondition := metav1.Condition{
					Type:   "Goodness",
					Status: metav1.ConditionTrue,
				}
				manager.AddPositive(goodnessCondition)

				greatnessCondition := metav1.Condition{
					Type:    "Greatness",
					Status:  metav1.ConditionFalse,
					Reason:  "SomeReason",
					Message: "some verbose message",
				}

				manager.AddPositive(greatnessCondition)
			})

			It("returns the conditions and a failing parent", func() {
				result := manager.Finalize()

				Expect(manager.IsSuccessful()).To(BeFalse())
				Expect(result).To(HaveLen(3))
				Expect(result).To(ContainElements(
					MatchFields(IgnoreExtras,
						Fields{
							"Type":   Equal("Goodness"),
							"Status": Equal(metav1.ConditionTrue),
						},
					),
					MatchFields(IgnoreExtras,
						Fields{
							"Type":   Equal("Greatness"),
							"Status": Equal(metav1.ConditionFalse),
						},
					),
					MatchFields(IgnoreExtras,
						Fields{
							"Type":    Equal("HappyParent"),
							"Status":  Equal(metav1.ConditionFalse),
							"Reason":  Equal("SomeReason"),
							"Message": Equal("some verbose message"),
						},
					),
				))
			})
		})
	})

	Context("with a negative polarity condition", func() {
		Context("with successful conditions", func() {
			BeforeEach(func() {
				manager = conditions.NewConditionManager("HappyParent", []metav1.Condition{})
				goodnessCondition := metav1.Condition{
					Type:   "Goodness",
					Status: metav1.ConditionTrue,
				}
				manager.AddPositive(goodnessCondition)

				badnessCondition := metav1.Condition{
					Type:   "Badness",
					Status: metav1.ConditionFalse,
				}

				manager.AddNegative(badnessCondition)
			})

			It("returns the conditions and a successful parent", func() {
				result := manager.Finalize()

				Expect(manager.IsSuccessful()).To(BeTrue())
				Expect(result).To(HaveLen(3))
				Expect(result).To(ContainElements(
					MatchFields(IgnoreExtras,
						Fields{
							"Type":   Equal("Goodness"),
							"Status": Equal(metav1.ConditionTrue),
						},
					),
					MatchFields(IgnoreExtras,
						Fields{
							"Type":   Equal("Badness"),
							"Status": Equal(metav1.ConditionFalse),
						},
					),
					MatchFields(IgnoreExtras,
						Fields{
							"Type":   Equal("HappyParent"),
							"Status": Equal(metav1.ConditionTrue),
						},
					),
				))
			})
		})

		Context("with a failing negative polarity condition", func() {
			BeforeEach(func() {
				manager = conditions.NewConditionManager("HappyParent", []metav1.Condition{})
				goodnessCondition := metav1.Condition{
					Type:   "Goodness",
					Status: metav1.ConditionTrue,
				}
				manager.AddPositive(goodnessCondition)

				badnessCondition := metav1.Condition{
					Type:    "Badness",
					Status:  metav1.ConditionTrue,
					Reason:  "SomeReason",
					Message: "some verbose message",
				}

				manager.AddNegative(badnessCondition)
			})

			It("returns the conditions and a failing parent", func() {
				result := manager.Finalize()

				Expect(manager.IsSuccessful()).To(BeFalse())
				Expect(result).To(HaveLen(3))
				Expect(result).To(ContainElements(
					MatchFields(IgnoreExtras,
						Fields{
							"Type":   Equal("Goodness"),
							"Status": Equal(metav1.ConditionTrue),
						},
					),
					MatchFields(IgnoreExtras,
						Fields{
							"Type":   Equal("Badness"),
							"Status": Equal(metav1.ConditionTrue),
						},
					),
					MatchFields(IgnoreExtras,
						Fields{
							"Type":    Equal("HappyParent"),
							"Status":  Equal(metav1.ConditionFalse),
							"Reason":  Equal("SomeReason"),
							"Message": Equal("some verbose message"),
						},
					),
				))
			})
		})
	})

	Context("when there are multiple conditions", func() {
		Context("and any are in a bad state", func() {
			BeforeEach(func() {
				manager = conditions.NewConditionManager("HappyParent", []metav1.Condition{})
				goodCondition := metav1.Condition{
					Type:   "some type",
					Status: metav1.ConditionTrue,
				}
				manager.AddPositive(goodCondition)

				badCondition := metav1.Condition{
					Type:    "another type",
					Status:  metav1.ConditionFalse,
					Reason:  "FirstReason",
					Message: "first verbose message",
				}

				manager.AddPositive(badCondition)

				secondBadCondition := metav1.Condition{
					Type:    "yet another type",
					Status:  metav1.ConditionFalse,
					Reason:  "SecondReason",
					Message: "second verbose message",
				}

				manager.AddPositive(secondBadCondition)

				unknownCondition := metav1.Condition{
					Type:   "additional type",
					Status: metav1.ConditionUnknown,
				}

				manager.AddPositive(unknownCondition)
			})

			It("returns a parent in a bad state", func() {
				result := manager.Finalize()

				Expect(manager.IsSuccessful()).To(BeFalse())
				Expect(result).To(ContainElement(
					MatchFields(IgnoreExtras,
						Fields{
							"Type":   Equal("HappyParent"),
							"Status": Equal(metav1.ConditionFalse),
						},
					),
				))
			})

			It("sets the parent reason and message to the last added bad condition", func() {
				result := manager.Finalize()

				Expect(result).To(ContainElement(
					MatchFields(IgnoreExtras,
						Fields{
							"Type":    Equal("HappyParent"),
							"Reason":  Equal("SecondReason"),
							"Message": Equal("second verbose message"),
						},
					),
				))
			})
		})

		Context("and some are in an unknown state, with none in a bad state", func() {
			BeforeEach(func() {
				manager = conditions.NewConditionManager("HappyParent", []metav1.Condition{})
				goodCondition := metav1.Condition{
					Type:   "some type",
					Status: metav1.ConditionTrue,
				}
				manager.AddPositive(goodCondition)

				gooderCondition := metav1.Condition{
					Type:   "another type",
					Status: metav1.ConditionTrue,
				}

				manager.AddPositive(gooderCondition)

				unknownCondition := metav1.Condition{
					Type:    "additional type",
					Status:  metav1.ConditionUnknown,
					Reason:  "NotKnown",
					Message: "some curious thing",
				}

				manager.AddPositive(unknownCondition)
			})

			It("returns a parent in an unknown state", func() {
				result := manager.Finalize()

				Expect(manager.IsSuccessful()).To(BeTrue())
				Expect(result).To(ContainElement(
					MatchFields(IgnoreExtras,
						Fields{
							"Type":    Equal("HappyParent"),
							"Status":  Equal(metav1.ConditionUnknown),
							"Reason":  Equal("NotKnown"),
							"Message": Equal("some curious thing"),
						},
					),
				))
			})
		})

		Context("and all are in a good state", func() {
			BeforeEach(func() {
				manager = conditions.NewConditionManager("HappyParent", []metav1.Condition{})
				goodCondition := metav1.Condition{
					Type:   "some type",
					Status: metav1.ConditionTrue,
				}
				manager.AddPositive(goodCondition)

				gooderCondition := metav1.Condition{
					Type:   "another type",
					Status: metav1.ConditionTrue,
				}

				manager.AddPositive(gooderCondition)

				goodestCondition := metav1.Condition{
					Type:   "additional type",
					Status: metav1.ConditionTrue,
				}

				manager.AddPositive(goodestCondition)
			})

			It("returns a parent in a good state", func() {
				result := manager.Finalize()

				Expect(manager.IsSuccessful()).To(BeTrue())
				Expect(result).To(ContainElement(
					MatchFields(IgnoreExtras,
						Fields{
							"Type":   Equal("HappyParent"),
							"Status": Equal(metav1.ConditionTrue),
							"Reason": Equal("Ready"),
						},
					),
				))
			})
		})
	})

	Context("with a plain string status", func() {
		BeforeEach(func() {
			manager = conditions.NewConditionManager("HappyParent", []metav1.Condition{})
			goodnessCondition := metav1.Condition{
				Type:   "Goodness",
				Status: "False",
			}
			manager.AddPositive(goodnessCondition)
		})

		It("returns the conditions and a failing parent", func() {
			result := manager.Finalize()

			Expect(manager.IsSuccessful()).To(BeFalse())
			Expect(result).To(HaveLen(2))
			Expect(result).To(ContainElements(
				MatchFields(IgnoreExtras,
					Fields{
						"Type":   Equal("Goodness"),
						"Status": Equal(metav1.ConditionFalse),
					},
				),
				MatchFields(IgnoreExtras,
					Fields{
						"Type":   Equal("HappyParent"),
						"Status": Equal(metav1.ConditionFalse),
					},
				),
			))
		})

	})

	Context("with previous conditions", func() {
		var (
			firstConditions   []metav1.Condition
			goodnessCondition metav1.Condition
		)

		BeforeEach(func() {
			manager = conditions.NewConditionManager("HappyParent", []metav1.Condition{})
			goodnessCondition = metav1.Condition{
				Type:   "Goodness",
				Status: metav1.ConditionTrue,
			}
			manager.AddPositive(goodnessCondition)
			firstConditions = manager.Finalize()
		})

		Context("when one of our conditions has changed", func() {
			BeforeEach(func() {
				manager = conditions.NewConditionManager("HappyParent", firstConditions)
				goodnessCondition.Reason = "Dog ate homework"
				manager.AddPositive(goodnessCondition)
			})

			It("does not affect other conditions", func() {
				newConditions := manager.Finalize()

				Expect(newConditions).To(ConsistOf(
					Not(Equal(firstConditions[0])),
					Equal(firstConditions[1]),
				))
			})
		})

		Context("when none of our conditions have changed", func() {
			BeforeEach(func() {
				manager = conditions.NewConditionManager("HappyParent", firstConditions)
				manager.AddPositive(goodnessCondition)
			})

			It("nothing is changed", func() {
				newConditions := manager.Finalize()

				Expect(newConditions).To(ConsistOf(
					Equal(firstConditions[0]),
					Equal(firstConditions[1]),
				))
			})
		})

	})
})
