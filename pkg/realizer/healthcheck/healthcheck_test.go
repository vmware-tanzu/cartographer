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

package healthcheck_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/healthcheck"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

var _ = Describe("DetermineHealthCondition", func() {
	It("is always healthy for AlwaysHealthy health rule", func() {
		healthRule := &v1alpha1.HealthRule{AlwaysHealthy: &runtime.RawExtension{Raw: []byte{}}}
		Expect(healthcheck.DetermineHealthCondition(healthRule, nil, nil)).To(MatchFields(IgnoreExtras,
			Fields{
				"Type":   Equal("Healthy"),
				"Status": Equal(metav1.ConditionTrue),
				"Reason": Equal(v1alpha1.AlwaysHealthyResourcesHealthyReason),
			},
		))
	})

	It("is always healthy for no rule on ClusterTemplates if a stamped object exists", func() {
		realizedResource := &v1alpha1.RealizedResource{
			StampedRef: &v1alpha1.StampedRef{
				ObjectReference: &corev1.ObjectReference{
					Kind:       "Anything",
					APIVersion: "of-any/kind",
				},
				Resource: "",
			},
			TemplateRef: &corev1.ObjectReference{
				Kind:       "ClusterTemplate",
				APIVersion: "carto.run/v1alpha1",
			},
		}

		Expect(healthcheck.DetermineHealthCondition(nil, realizedResource, nil)).To(MatchFields(IgnoreExtras,
			Fields{
				"Type":   Equal("Healthy"),
				"Status": Equal(metav1.ConditionTrue),
				"Reason": Equal(v1alpha1.AlwaysHealthyResourcesHealthyReason),
			},
		))
	})

	It("is always unknown for no rule on ClusterTemplates if no stamped object exists", func() {
		realizedResource := &v1alpha1.RealizedResource{
			TemplateRef: &corev1.ObjectReference{
				Kind:       "ClusterTemplate",
				APIVersion: "carto.run/v1alpha1",
			},
		}

		Expect(healthcheck.DetermineHealthCondition(nil, realizedResource, nil)).To(MatchFields(IgnoreExtras,
			Fields{
				"Type":   Equal("Healthy"),
				"Status": Equal(metav1.ConditionUnknown),
				"Reason": Equal(v1alpha1.NoStampedObjectHealthyReason),
			},
		))
	})

	It("is Unknown when no health rule and no resource", func() {
		Expect(healthcheck.DetermineHealthCondition(nil, nil, nil)).To(MatchFields(IgnoreExtras,
			Fields{
				"Type":   Equal("Healthy"),
				"Status": Equal(metav1.ConditionUnknown),
				"Reason": Equal(v1alpha1.NoResourceResourcesHealthyReason),
			},
		))
	})

	It("is Unknown when no health rule and resource has no outputs", func() {
		realizedResource := &v1alpha1.RealizedResource{
			Outputs: []v1alpha1.Output{},
		}
		Expect(healthcheck.DetermineHealthCondition(nil, realizedResource, nil)).To(MatchFields(IgnoreExtras,
			Fields{
				"Type":   Equal("Healthy"),
				"Status": Equal(metav1.ConditionUnknown),
				"Reason": Equal(v1alpha1.OutputNotAvailableResourcesHealthyReason),
			},
		))
	})

	It("is Healthy when no health rule and resource has outputs", func() {
		realizedResource := &v1alpha1.RealizedResource{
			Outputs: []v1alpha1.Output{
				{
					Name: "cool-output",
				},
			},
		}
		Expect(healthcheck.DetermineHealthCondition(nil, realizedResource, nil)).To(MatchFields(IgnoreExtras,
			Fields{
				"Type":   Equal("Healthy"),
				"Status": Equal(metav1.ConditionTrue),
				"Reason": Equal(v1alpha1.OutputAvailableResourcesHealthyReason),
			},
		))
	})

	Context("HealthRule is singleConditionType", func() {
		var healthRule *v1alpha1.HealthRule

		BeforeEach(func() {
			healthRule = &v1alpha1.HealthRule{
				SingleConditionType: "OhSoWonderful",
			}
		})

		It("returns unknown if there is no stamped object", func() {
			Expect(healthcheck.DetermineHealthCondition(healthRule, nil, nil)).To(MatchFields(IgnoreExtras,
				Fields{
					"Type":   Equal("Healthy"),
					"Status": Equal(metav1.ConditionUnknown),
				},
			))
		})

		It("returns the status of the condition on the stamped object if it is True", func() {
			stampedObject := &unstructured.Unstructured{}
			stampedObjectYaml := utils.HereYamlF(`
				apiVersion: thing/v1
				kind: Thing
				metadata:
				  name: named-thing
				  namespace: somens
				spec:
				status:
				  conditions:
				    - type: OhSoWonderful
				      status: "True"
				      message: "congratulations on your clean bill of health!"
			`)

			dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
			_, _, err := dec.Decode([]byte(stampedObjectYaml), nil, stampedObject)
			Expect(err).NotTo(HaveOccurred())
			Expect(healthcheck.DetermineHealthCondition(healthRule, nil, stampedObject)).To(MatchFields(IgnoreExtras,
				Fields{
					"Type":    Equal("Healthy"),
					"Status":  Equal(metav1.ConditionTrue),
					"Reason":  Equal("OhSoWonderfulCondition"),
					"Message": Equal("congratulations on your clean bill of health!"),
				},
			))
		})

		It("returns the status of the condition on the stamped object if it is False", func() {
			stampedObject := &unstructured.Unstructured{}
			stampedObjectYaml := utils.HereYamlF(`
				apiVersion: thing/v1
				kind: Thing
				metadata:
				  name: named-thing
				  namespace: somens
				spec:
				status:
				  conditions:
				    - type: OhSoWonderful
				      status: "False"
				      message: "sorry we're all our of wonderful. please check back later."
			`)

			dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
			_, _, err := dec.Decode([]byte(stampedObjectYaml), nil, stampedObject)
			Expect(err).NotTo(HaveOccurred())
			Expect(healthcheck.DetermineHealthCondition(healthRule, nil, stampedObject)).To(MatchFields(IgnoreExtras,
				Fields{
					"Type":    Equal("Healthy"),
					"Status":  Equal(metav1.ConditionFalse),
					"Reason":  Equal("OhSoWonderfulCondition"),
					"Message": Equal("sorry we're all our of wonderful. please check back later."),
				},
			))
		})

		It("returns Unknown status if the condition status on the stamped object is not True or False", func() {
			stampedObject := &unstructured.Unstructured{}
			stampedObjectYaml := utils.HereYamlF(`
				apiVersion: thing/corev1
				kind: Thing
				metadata:
				  name: named-thing
				  namespace: somens
				spec:
				status:
				  conditions:
				    - type: OhSoWonderful
				      status: "SomethingElse"
				      message: "an exciting message about our condition"
			`)

			dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
			_, _, err := dec.Decode([]byte(stampedObjectYaml), nil, stampedObject)
			Expect(err).NotTo(HaveOccurred())
			Expect(healthcheck.DetermineHealthCondition(healthRule, nil, stampedObject)).To(MatchFields(IgnoreExtras,
				Fields{
					"Type":    Equal("Healthy"),
					"Status":  Equal(metav1.ConditionUnknown),
					"Reason":  Equal("OhSoWonderfulCondition"),
					"Message": Equal("an exciting message about our condition"),
				},
			))
		})

		It("returns Unknown status if the condition is not present on the stamped object", func() {
			stampedObject := &unstructured.Unstructured{}
			stampedObjectYaml := utils.HereYamlF(`
				apiVersion: thing/corev1
				kind: Thing
				metadata:
				  name: named-thing
				  namespace: somens
				spec:
			`)

			dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
			_, _, err := dec.Decode([]byte(stampedObjectYaml), nil, stampedObject)
			Expect(err).NotTo(HaveOccurred())
			Expect(healthcheck.DetermineHealthCondition(healthRule, nil, stampedObject)).To(MatchFields(IgnoreExtras,
				Fields{
					"Type":    Equal("Healthy"),
					"Status":  Equal(metav1.ConditionUnknown),
					"Reason":  Equal("OhSoWonderfulCondition"),
					"Message": Equal("condition with type [OhSoWonderful] not found on resource status"),
				},
			))
		})
	})

	Context("HealthRule is MultiMatch", func() {
		var healthRule *v1alpha1.HealthRule

		BeforeEach(func() {
			healthRule = &v1alpha1.HealthRule{
				MultiMatch: &v1alpha1.MultiMatchHealthRule{
					Healthy: v1alpha1.HealthMatchRule{
						MatchConditions: []v1alpha1.ConditionRequirement{
							{
								Type:   "HealthyCond1",
								Status: "True",
							},
							{
								Type:   "HealthyCond2",
								Status: "VeryHealthy",
							},
						},
						MatchFields: []v1alpha1.HealthMatchFieldSelectorRequirement{
							{
								FieldSelectorRequirement: v1alpha1.FieldSelectorRequirement{
									Key:      `.status.howgood`,
									Operator: "In",
									Values:   []string{"VeryGood"},
								},
							},
							{
								FieldSelectorRequirement: v1alpha1.FieldSelectorRequirement{
									Key:      `.status.greenlight`,
									Operator: "Exists",
								},
							},
						},
					},
					Unhealthy: v1alpha1.HealthMatchRule{
						MatchConditions: []v1alpha1.ConditionRequirement{
							{
								Type:   "UnhealthyCond1",
								Status: "VeryUnhealthy",
							},
							{
								Type:   "UnhealthyCond2",
								Status: "True",
							},
						},
						MatchFields: []v1alpha1.HealthMatchFieldSelectorRequirement{
							{
								FieldSelectorRequirement: v1alpha1.FieldSelectorRequirement{
									Key:      `.status.stopsign`,
									Operator: "Exists",
								},
								MessagePath: "status.usefulErrorMessage",
							},
							{
								FieldSelectorRequirement: v1alpha1.FieldSelectorRequirement{
									Key:      `status.conditions[?(@.type=="DoesntMatterJustPassingOneIsEnoughToBeUnhealthy")].status`,
									Operator: "In",
									Values:   []string{"True"},
								},
								MessagePath: "status.usefulErrorMessage",
							},
						},
					},
				},
			}
		})

		It("returns unknown if there is no stamped object", func() {
			Expect(healthcheck.DetermineHealthCondition(healthRule, nil, nil)).To(MatchFields(IgnoreExtras,
				Fields{
					"Type":   Equal("Healthy"),
					"Status": Equal(metav1.ConditionUnknown),
				},
			))
		})

		It("returns True if all healthy condition and field expectations are met", func() {
			stampedObject := &unstructured.Unstructured{}
			stampedObjectYaml := utils.HereYamlF(`
				apiVersion: thing/v1
				kind: Thing
				metadata:
				  name: named-thing
				  namespace: somens
				spec:
				status:
                  howgood: VeryGood
                  greenlight: true
				  conditions:
				    - type: HealthyCond1
				      status: "True"
				      message: "here is my message"
				    - type: HealthyCond2
				      status: "VeryHealthy"
			`)

			dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
			_, _, err := dec.Decode([]byte(stampedObjectYaml), nil, stampedObject)
			Expect(err).NotTo(HaveOccurred())
			Expect(healthcheck.DetermineHealthCondition(healthRule, nil, stampedObject)).To(MatchFields(IgnoreExtras,
				Fields{
					"Type":    Equal("Healthy"),
					"Status":  Equal(metav1.ConditionTrue),
					"Reason":  Equal("MatchedCondition"),
					"Message": Equal("condition status: True, message: here is my message"),
				},
			))
		})

		It("uses the first matching field messagePath if no condition matches occur", func() {
			healthRule = &v1alpha1.HealthRule{
				MultiMatch: &v1alpha1.MultiMatchHealthRule{
					Healthy: v1alpha1.HealthMatchRule{
						MatchConditions: []v1alpha1.ConditionRequirement{},
						MatchFields: []v1alpha1.HealthMatchFieldSelectorRequirement{
							{
								FieldSelectorRequirement: v1alpha1.FieldSelectorRequirement{
									Key:      `.status.greenlight`,
									Operator: "Exists",
								},
								MessagePath: "status.infoMessage",
							},
						},
					},
					Unhealthy: v1alpha1.HealthMatchRule{
						MatchConditions: []v1alpha1.ConditionRequirement{},
						MatchFields: []v1alpha1.HealthMatchFieldSelectorRequirement{
							{
								FieldSelectorRequirement: v1alpha1.FieldSelectorRequirement{
									Key:      `.status.stopsign`,
									Operator: "Exists",
								},
								MessagePath: "status.usefulErrorMessage",
							},
						},
					},
				},
			}
			stampedObject := &unstructured.Unstructured{}
			stampedObjectYaml := utils.HereYamlF(`
				apiVersion: thing/v1
				kind: Thing
				metadata:
				  name: named-thing
				  namespace: somens
				spec:
				status:
                  greenlight: chartreuse
                  infoMessage: "did everything successfully in 3.14159 seconds"
			`)

			dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
			_, _, err := dec.Decode([]byte(stampedObjectYaml), nil, stampedObject)
			Expect(err).NotTo(HaveOccurred())
			Expect(healthcheck.DetermineHealthCondition(healthRule, nil, stampedObject)).To(MatchFields(IgnoreExtras,
				Fields{
					"Type":    Equal("Healthy"),
					"Status":  Equal(metav1.ConditionTrue),
					"Reason":  Equal("MatchedField"),
					"Message": Equal("field value: chartreuse, message: did everything successfully in 3.14159 seconds"),
				},
			))
		})

		It("returns Unknown if not all condition expectations are met", func() {
			stampedObject := &unstructured.Unstructured{}
			stampedObjectYaml := utils.HereYamlF(`
				apiVersion: thing/v1
				kind: Thing
				metadata:
				  name: named-thing
				  namespace: somens
				spec:
				status:
                  howgood: VeryGood
                  greenlight: true
				  conditions:
				    - type: HealthyCond1
				      status: "True"
				    - type: HealthyCond2
				      status: "NoSoMuch"
			`)

			dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
			_, _, err := dec.Decode([]byte(stampedObjectYaml), nil, stampedObject)
			Expect(err).NotTo(HaveOccurred())
			Expect(healthcheck.DetermineHealthCondition(healthRule, nil, stampedObject)).To(MatchFields(IgnoreExtras,
				Fields{
					"Type":   Equal("Healthy"),
					"Status": Equal(metav1.ConditionUnknown),
					"Reason": Equal("NoMatchesFulfilled"),
				},
			))
		})

		It("returns Unknown if not all field expectations are met", func() {
			stampedObject := &unstructured.Unstructured{}
			stampedObjectYaml := utils.HereYamlF(`
				apiVersion: thing/v1
				kind: Thing
				metadata:
				  name: named-thing
				  namespace: somens
				spec:
				status:
                  howgood: NotThatGreatHonestly
                  greenlight: true
				  conditions:
				    - type: HealthyCond1
				      status: "True"
				    - type: HealthyCond2
				      status: "VeryHealthy"
			`)

			dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
			_, _, err := dec.Decode([]byte(stampedObjectYaml), nil, stampedObject)
			Expect(err).NotTo(HaveOccurred())
			Expect(healthcheck.DetermineHealthCondition(healthRule, nil, stampedObject)).To(MatchFields(IgnoreExtras,
				Fields{
					"Type":   Equal("Healthy"),
					"Status": Equal(metav1.ConditionUnknown),
					"Reason": Equal("NoMatchesFulfilled"),
				},
			))
		})

		It("returns False if any unhealthy condition expectation is met", func() {
			stampedObject := &unstructured.Unstructured{}
			stampedObjectYaml := utils.HereYamlF(`
				apiVersion: thing/v1
				kind: Thing
				metadata:
				  name: named-thing
				  namespace: somens
				spec:
				status:
                  howgood: VeryGood
                  greenlight: true
				  conditions:
				    - type: HealthyCond1
				      status: "True"
				    - type: HealthyCond2
				      status: "VeryHealthy"
				    - type: UnhealthyCond2
                      status: "True"
                      message: "looks like a case of the flu"
			`)

			dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
			_, _, err := dec.Decode([]byte(stampedObjectYaml), nil, stampedObject)
			Expect(err).NotTo(HaveOccurred())
			Expect(healthcheck.DetermineHealthCondition(healthRule, nil, stampedObject)).To(MatchFields(IgnoreExtras,
				Fields{
					"Type":    Equal("Healthy"),
					"Status":  Equal(metav1.ConditionFalse),
					"Reason":  Equal("MatchedCondition"),
					"Message": Equal("condition status: True, message: looks like a case of the flu"),
				},
			))
		})

		It("returns False if any unhealthy field expectations is met", func() {
			stampedObject := &unstructured.Unstructured{}
			stampedObjectYaml := utils.HereYamlF(`
				apiVersion: thing/v1
				kind: Thing
				metadata:
				  name: named-thing
				  namespace: somens
				spec:
				status:
                  howgood: VeryGood
                  greenlight: true
                  stopsign: true
                  usefulErrorMessage: "this would contain some useful context for the reader"
				  conditions:
				    - type: HealthyCond1
				      status: "True"
				    - type: HealthyCond2
				      status: "VeryHealthy"
			`)

			dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
			_, _, err := dec.Decode([]byte(stampedObjectYaml), nil, stampedObject)
			Expect(err).NotTo(HaveOccurred())
			Expect(healthcheck.DetermineHealthCondition(healthRule, nil, stampedObject)).To(MatchFields(IgnoreExtras,
				Fields{
					"Type":    Equal("Healthy"),
					"Status":  Equal(metav1.ConditionFalse),
					"Reason":  Equal("MatchedField"),
					"Message": Equal("field value: true, message: this would contain some useful context for the reader"),
				},
			))
		})

		It("surfaces errors evaluating message path into the message", func() {
			stampedObject := &unstructured.Unstructured{}
			stampedObjectYaml := utils.HereYamlF(`
				apiVersion: thing/v1
				kind: Thing
				metadata:
				  name: named-thing
				  namespace: somens
				spec:
				status:
                  stopsign: true
				  conditions: []
			`)

			dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
			_, _, err := dec.Decode([]byte(stampedObjectYaml), nil, stampedObject)
			Expect(err).NotTo(HaveOccurred())
			Expect(healthcheck.DetermineHealthCondition(healthRule, nil, stampedObject)).To(MatchFields(IgnoreExtras,
				Fields{
					"Type":    Equal("Healthy"),
					"Status":  Equal(metav1.ConditionFalse),
					"Reason":  Equal("MatchedField"),
					"Message": Equal("field value: true, message: unknown, error retrieving message path [status.usefulErrorMessage]"),
				},
			))
		})
	})
})
