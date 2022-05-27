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

	It("is always healthy for no rule on ClusterTemplates", func() {
		realizedResource := &v1alpha1.RealizedResource{
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
			`)

			dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
			_, _, err := dec.Decode([]byte(stampedObjectYaml), nil, stampedObject)
			Expect(err).NotTo(HaveOccurred())
			Expect(healthcheck.DetermineHealthCondition(healthRule, nil, stampedObject)).To(MatchFields(IgnoreExtras,
				Fields{
					"Type":   Equal("Healthy"),
					"Status": Equal(metav1.ConditionTrue),
					"Reason": Equal("OhSoWonderfulCondition"),
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
			`)

			dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
			_, _, err := dec.Decode([]byte(stampedObjectYaml), nil, stampedObject)
			Expect(err).NotTo(HaveOccurred())
			Expect(healthcheck.DetermineHealthCondition(healthRule, nil, stampedObject)).To(MatchFields(IgnoreExtras,
				Fields{
					"Type":   Equal("Healthy"),
					"Status": Equal(metav1.ConditionFalse),
					"Reason": Equal("OhSoWonderfulCondition"),
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
			`)

			dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
			_, _, err := dec.Decode([]byte(stampedObjectYaml), nil, stampedObject)
			Expect(err).NotTo(HaveOccurred())
			Expect(healthcheck.DetermineHealthCondition(healthRule, nil, stampedObject)).To(MatchFields(IgnoreExtras,
				Fields{
					"Type":   Equal("Healthy"),
					"Status": Equal(metav1.ConditionUnknown),
					"Reason": Equal("OhSoWonderfulCondition"),
				},
			))
		})
	})
})
