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

package realizer_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/events"
	"github.com/vmware-tanzu/cartographer/pkg/events/eventsfakes"
	"github.com/vmware-tanzu/cartographer/pkg/realizer"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/realizerfakes"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/statuses"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

type event struct {
	EventType    string
	Reason       string
	Message      string
	ResourceName string
	FmtArgs      []interface{}
}

var _ = Describe("Realize", func() {
	var (
		resourceRealizer               *realizerfakes.FakeResourceRealizer
		rlzr                           realizer.Realizer
		rec                            *eventsfakes.FakeOwnerEventRecorder
		healthyConditionEvaluator      realizer.HealthyConditionEvaluator
		evaluatedHealthRules           []*v1alpha1.HealthRule
		evaluatedRealizedResourceNames []string
		evaluatedStampedObjectNames    []string
		ctx                            context.Context
		recordedEvents                 []event
		fakeMapper                     *realizerfakes.FakeRESTMapper
	)

	BeforeEach(func() {
		ctx = context.TODO()
		rec = &eventsfakes.FakeOwnerEventRecorder{}
		ctx = events.NewContext(ctx, rec)

		recordedEvents = nil
		rec.ResourceEventfCalls(func(eventtype, reason, messageFmt string, resource *unstructured.Unstructured, i ...interface{}) {
			recordedEvents = append(recordedEvents, event{eventtype, reason, messageFmt, resource.GetName(), i})
		})
		evaluatedHealthRules = []*v1alpha1.HealthRule{}
		evaluatedRealizedResourceNames = []string{}
		evaluatedStampedObjectNames = []string{}
		healthyConditionEvaluator = func(rule *v1alpha1.HealthRule, realizedResource *v1alpha1.RealizedResource, stampedObject *unstructured.Unstructured) metav1.Condition {
			evaluatedHealthRules = append(evaluatedHealthRules, rule)
			evaluatedRealizedResourceNames = append(evaluatedRealizedResourceNames, realizedResource.Name)
			evaluatedStampedObjectNames = append(evaluatedStampedObjectNames, stampedObject.GetName())
			return metav1.Condition{
				Type:   "Healthy",
				Status: "True",
				Reason: "EvaluatorSaysSo",
			}
		}
		fakeMapper = &realizerfakes.FakeRESTMapper{}
		rlzr = realizer.NewRealizer(healthyConditionEvaluator, fakeMapper)
		resourceRealizer = &realizerfakes.FakeResourceRealizer{}
	})

	Context("there are no previous resources", func() {
		var (
			template1             *v1alpha1.ClusterImageTemplate
			template2             *v1alpha1.ClusterTemplate
			executedResourceOrder []string
			supplyChain           *v1alpha1.ClusterSupplyChain
			resource1             v1alpha1.SupplyChainResource
			resource2             v1alpha1.SupplyChainResource
		)
		BeforeEach(func() {
			resource1 = v1alpha1.SupplyChainResource{
				Name: "resource1",
			}
			resource2 = v1alpha1.SupplyChainResource{
				Name: "resource2",
				Images: []v1alpha1.ResourceReference{
					{
						Name:     "my-image",
						Resource: "resource1",
					},
				},
			}
			template1 = &v1alpha1.ClusterImageTemplate{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-image-template",
				},
				Spec: v1alpha1.ImageTemplateSpec{
					TemplateSpec: v1alpha1.TemplateSpec{
						HealthRule: &v1alpha1.HealthRule{
							SingleConditionType: "Happy",
						},
					},
				},
			}
			template2 = &v1alpha1.ClusterTemplate{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-cluster-template",
				},
				Spec: v1alpha1.TemplateSpec{
					HealthRule: &v1alpha1.HealthRule{
						AlwaysHealthy: &runtime.RawExtension{Raw: []byte(`{}`)},
					},
				},
			}
			supplyChain = &v1alpha1.ClusterSupplyChain{
				ObjectMeta: metav1.ObjectMeta{Name: "greatest-supply-chain"},
				Spec: v1alpha1.SupplyChainSpec{
					Resources: []v1alpha1.SupplyChainResource{resource1, resource2},
				},
			}

			outputFromFirstResource := &templates.Output{Image: "whatever"}

			resourceRealizer.DoCalls(func(ctx context.Context, resource realizer.OwnerResource, blueprintName string, outputs realizer.Outputs, mapper meta.RESTMapper) (templates.Template, *unstructured.Unstructured, *templates.Output, error) {
				executedResourceOrder = append(executedResourceOrder, resource.Name)
				Expect(blueprintName).To(Equal("greatest-supply-chain"))
				if resource.Name == "resource1" {
					Expect(outputs).To(Equal(realizer.NewOutputs()))
					template, err := templates.NewModelFromAPI(template1)
					Expect(err).NotTo(HaveOccurred())
					stampedObj := &unstructured.Unstructured{}
					stampedObj.SetName("obj1")
					return template, stampedObj, outputFromFirstResource, nil
				}

				if resource.Name == "resource2" {
					expectedSecondResourceOutputs := realizer.NewOutputs()
					expectedSecondResourceOutputs.AddOutput("resource1", outputFromFirstResource)
					Expect(outputs).To(Equal(expectedSecondResourceOutputs))
				}
				template, err := templates.NewModelFromAPI(template2)
				Expect(err).NotTo(HaveOccurred())
				stampedObj := &unstructured.Unstructured{}
				stampedObj.SetName("obj2")
				return template, stampedObj, &templates.Output{}, nil
			})

			// TODO
			fakeMapper.RESTMappingReturns(&meta.RESTMapping{
				Resource: schema.GroupVersionResource{
					Group:    "EXAMPLE.COM",
					Version:  "v1",
					Resource: "FOO",
				},
				GroupVersionKind: schema.GroupVersionKind{
					Group:   "",
					Version: "",
					Kind:    "",
				},
				Scope: nil,
			}, nil)

		})

		It("realizes each resource in supply chain order, accumulating output for each subsequent resource", func() {
			resourceStatuses := statuses.NewResourceStatuses(nil, conditions.AddConditionForResourceSubmittedWorkload)
			err := rlzr.Realize(ctx, resourceRealizer, supplyChain.Name, realizer.MakeSupplychainOwnerResources(supplyChain), resourceStatuses)
			Expect(err).ToNot(HaveOccurred())

			currentResourceStatuses := resourceStatuses.GetCurrent()
			Expect(executedResourceOrder).To(Equal([]string{"resource1", "resource2"}))

			Expect(evaluatedHealthRules).To(Equal([]*v1alpha1.HealthRule{template1.Spec.HealthRule, template2.Spec.HealthRule}))

			Expect(evaluatedRealizedResourceNames).To(Equal([]string{"resource1", "resource2"}))
			Expect(evaluatedStampedObjectNames).To(Equal([]string{"obj1", "obj2"}))

			Expect(currentResourceStatuses).To(HaveLen(2))

			Expect(currentResourceStatuses[0].Name).To(Equal(resource1.Name))
			Expect(currentResourceStatuses[0].TemplateRef.Name).To(Equal(template1.Name))
			Expect(currentResourceStatuses[0].StampedRef.Name).To(Equal("obj1"))
			Expect(currentResourceStatuses[0].Inputs).To(BeNil())
			Expect(len(currentResourceStatuses[0].Outputs)).To(Equal(1))
			Expect(currentResourceStatuses[0].Outputs[0]).To(MatchFields(IgnoreExtras,
				Fields{
					"Name":    Equal("image"),
					"Preview": Equal("whatever\n"),
					"Digest":  HavePrefix("sha256"),
				},
			))
			Expect(time.Since(currentResourceStatuses[0].Outputs[0].LastTransitionTime.Time)).To(BeNumerically("<", time.Second))
			Expect(len(currentResourceStatuses[0].Conditions)).To(Equal(3))

			Expect(currentResourceStatuses[0].Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal("ResourceSubmitted"),
				"Status": Equal(metav1.ConditionTrue),
			})))
			Expect(currentResourceStatuses[0].Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal("Healthy"),
				"Status": Equal(metav1.ConditionTrue),
			})))
			Expect(currentResourceStatuses[0].Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal("Ready"),
				"Status": Equal(metav1.ConditionTrue),
			})))

			Expect(currentResourceStatuses[1].Name).To(Equal(resource2.Name))
			Expect(currentResourceStatuses[1].TemplateRef.Name).To(Equal(template2.Name))
			Expect(currentResourceStatuses[1].StampedRef.Name).To(Equal("obj2"))
			Expect(len(currentResourceStatuses[1].Inputs)).To(Equal(1))
			Expect(currentResourceStatuses[1].Inputs).To(Equal([]v1alpha1.Input{{Name: "resource1"}}))
			Expect(currentResourceStatuses[1].Outputs).To(BeNil())
			Expect(len(currentResourceStatuses[0].Conditions)).To(Equal(3))

			Expect(currentResourceStatuses[1].Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal("ResourceSubmitted"),
				"Status": Equal(metav1.ConditionTrue),
			})))
			Expect(currentResourceStatuses[1].Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal("Healthy"),
				"Status": Equal(metav1.ConditionTrue),
			})))
			Expect(currentResourceStatuses[1].Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal("Ready"),
				"Status": Equal(metav1.ConditionTrue),
			})))
		})

		It("records an event for resource output changes and health status", func() {
			resourceStatuses := statuses.NewResourceStatuses(nil, conditions.AddConditionForResourceSubmittedWorkload)
			Expect(rlzr.Realize(ctx, resourceRealizer, supplyChain.Name, realizer.MakeSupplychainOwnerResources(supplyChain), resourceStatuses)).To(Succeed())

			Expect(recordedEvents).To(ConsistOf(
				event{"Normal", events.ResourceOutputChangedReason, "[%s] found a new output in [%Q]", "obj1", []interface{}{"resource1"}},
				event{"Normal", events.ResourceHealthyStatusChangedReason, "[%s] found healthy status in [%Q] changed to [%s]", "obj1", []interface{}{"resource1", metav1.ConditionTrue}},
				event{"Normal", events.ResourceHealthyStatusChangedReason, "[%s] found healthy status in [%Q] changed to [%s]", "obj2", []interface{}{"resource2", metav1.ConditionTrue}},
			))
		})

		It("does not record an ResourceOutputChanged event if there was no resource output change", func() {
			previousResources := []v1alpha1.ResourceStatus{
				{
					RealizedResource: v1alpha1.RealizedResource{
						Name: "resource1",
						Outputs: []v1alpha1.Output{
							{
								Name:               "image",
								Preview:            "whatever\n",
								Digest:             fmt.Sprintf("sha256:%x", sha256.Sum256([]byte("whatever\n"))),
								LastTransitionTime: metav1.Now(),
							},
						},
					},
				},
			}

			resourceStatuses := statuses.NewResourceStatuses(previousResources, conditions.AddConditionForResourceSubmittedWorkload)
			Expect(rlzr.Realize(ctx, resourceRealizer, supplyChain.Name, realizer.MakeSupplychainOwnerResources(supplyChain), resourceStatuses)).To(Succeed())

			Expect(recordedEvents).NotTo(ContainElement(MatchFields(IgnoreExtras, Fields{"Reason": Equal(events.ResourceOutputChangedReason)})))
		})

		It("returns the first error encountered realizing a resource and continues to realize", func() {
			template, err := templates.NewModelFromAPI(template2)
			Expect(err).NotTo(HaveOccurred())
			resourceRealizer.DoReturnsOnCall(0, nil, nil, nil, errors.New("realizing is hard"))
			resourceRealizer.DoReturnsOnCall(1, template, &unstructured.Unstructured{}, nil, nil)

			resourceStatuses := statuses.NewResourceStatuses(nil, conditions.AddConditionForResourceSubmittedWorkload)
			err = rlzr.Realize(ctx, resourceRealizer, supplyChain.Name, realizer.MakeSupplychainOwnerResources(supplyChain), resourceStatuses)
			Expect(err).To(MatchError("realizing is hard"))

			currentResourceStatuses := resourceStatuses.GetCurrent()
			Expect(currentResourceStatuses).To(HaveLen(2))

			Expect(currentResourceStatuses[0].Name).To(Equal("resource1"))
			Expect(currentResourceStatuses[0].TemplateRef).To(BeNil())
			Expect(currentResourceStatuses[0].StampedRef).To(BeNil())
			Expect(currentResourceStatuses[1].TemplateRef.Name).To(Equal(template2.Name))
		})
	})

	Context("there are previous resources", func() {
		var (
			templateModel1    templates.Template
			templateModel2    templates.Template
			templateModel3    templates.Template
			obj               *unstructured.Unstructured
			previousResources []v1alpha1.ResourceStatus
			previousTime      metav1.Time
			supplyChain       *v1alpha1.ClusterSupplyChain
		)
		BeforeEach(func() {
			previousTime = metav1.NewTime(time.Now())
			previousResources = []v1alpha1.ResourceStatus{
				{
					RealizedResource: v1alpha1.RealizedResource{
						Name: "resource2",
						StampedRef: &v1alpha1.StampedRef{
							ObjectReference: &corev1.ObjectReference{
								Kind:       "Image",
								Namespace:  "",
								Name:       "",
								APIVersion: "",
							},
							Resource: "",
						},
						TemplateRef: &corev1.ObjectReference{
							Kind:       "ClusterImageTemplate",
							Name:       "my-image-template",
							APIVersion: "",
						},
						Inputs: []v1alpha1.Input{
							{
								Name: "resource1",
							},
						},
						Outputs: []v1alpha1.Output{
							{
								Name:               "image",
								Preview:            "whatever\n",
								Digest:             fmt.Sprintf("sha256:%x", sha256.Sum256([]byte("whatever\n"))),
								LastTransitionTime: previousTime,
							},
						},
					},
					Conditions: []metav1.Condition{
						{
							Type:   "Healthy",
							Status: metav1.ConditionTrue,
							Reason: "HealthyReasonFromBefore",
						},
					},
				},
				{
					RealizedResource: v1alpha1.RealizedResource{
						Name: "resource3",
						StampedRef: &v1alpha1.StampedRef{
							ObjectReference: &corev1.ObjectReference{
								Kind:       "Config",
								Namespace:  "",
								Name:       "PreviousStampedObj",
								APIVersion: "",
							},
							Resource: "",
						},
						TemplateRef: &corev1.ObjectReference{
							Kind:       "ClusterConfigTemplate",
							Name:       "my-config-template",
							APIVersion: "",
						},
						Outputs: []v1alpha1.Output{
							{
								Name:               "config",
								Preview:            "whatever\n",
								Digest:             fmt.Sprintf("sha256:%x", sha256.Sum256([]byte("whatever\n"))),
								LastTransitionTime: previousTime,
							},
						},
					},
				},
			}

			resource1 := v1alpha1.SupplyChainResource{
				Name: "resource1",
			}
			resource2 := v1alpha1.SupplyChainResource{
				Name: "resource2",
				Sources: []v1alpha1.ResourceReference{
					{
						Name:     "my-source",
						Resource: "resource1",
					},
				},
			}
			resource3 := v1alpha1.SupplyChainResource{
				Name: "resource3",
			}

			supplyChain = &v1alpha1.ClusterSupplyChain{
				ObjectMeta: metav1.ObjectMeta{Name: "greatest-supply-chain"},
				Spec: v1alpha1.SupplyChainSpec{
					Resources: []v1alpha1.SupplyChainResource{resource1, resource2, resource3},
				},
			}

			var err error

			template1 := &v1alpha1.ClusterSourceTemplate{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-source-2-template",
				},
			}
			templateModel1, err = templates.NewModelFromAPI(template1)
			Expect(err).NotTo(HaveOccurred())

			template2 := &v1alpha1.ClusterImageTemplate{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-image-template",
				},
			}
			templateModel2, err = templates.NewModelFromAPI(template2)
			Expect(err).NotTo(HaveOccurred())

			template3 := &v1alpha1.ClusterConfigTemplate{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-config-template",
				},
			}
			templateModel3, err = templates.NewModelFromAPI(template3)
			Expect(err).NotTo(HaveOccurred())

			resourceRealizer.DoReturnsOnCall(0, templateModel1, &unstructured.Unstructured{}, nil, nil)
			resourceRealizer.DoReturnsOnCall(1, templateModel2, &unstructured.Unstructured{}, nil, nil)
			resourceRealizer.DoReturnsOnCall(2, templateModel3, &unstructured.Unstructured{}, nil, nil)

			// TODO
			fakeMapper.RESTMappingReturns(&meta.RESTMapping{
				Resource: schema.GroupVersionResource{
					Group:    "EXAMPLE.COM",
					Version:  "v1",
					Resource: "FOO",
				},
				GroupVersionKind: schema.GroupVersionKind{
					Group:   "",
					Version: "",
					Kind:    "",
				},
				Scope: nil,
			}, nil)
		})

		It("realizes each resource and does not update last transition time on resources that have not changed", func() {
			newOutput := &templates.Output{
				Source: &templates.Source{
					URL:      "hi",
					Revision: "bye",
				},
			}
			stampedObj1 := &unstructured.Unstructured{}
			stampedObj1.SetName("obj1")
			resourceRealizer.DoReturnsOnCall(0, templateModel1, stampedObj1, newOutput, nil)

			oldOutput := &templates.Output{
				Image: "whatever",
			}
			resourceRealizer.DoReturnsOnCall(1, templateModel2, &unstructured.Unstructured{}, oldOutput, nil)

			oldOutput2 := &templates.Output{
				Config: "whatever",
			}
			resourceRealizer.DoReturnsOnCall(2, templateModel3, obj, oldOutput2, nil)

			resourceStatuses := statuses.NewResourceStatuses(previousResources, conditions.AddConditionForResourceSubmittedWorkload)
			err := rlzr.Realize(ctx, resourceRealizer, supplyChain.Name, realizer.MakeSupplychainOwnerResources(supplyChain), resourceStatuses)
			Expect(err).ToNot(HaveOccurred())

			currentStatuses := resourceStatuses.GetCurrent()
			var resource1Status v1alpha1.ResourceStatus
			var resource2Status v1alpha1.ResourceStatus
			var resource3Status v1alpha1.ResourceStatus

			for i := range currentStatuses {
				switch currentStatuses[i].Name {
				case "resource1":
					resource1Status = currentStatuses[i]
				case "resource2":
					resource2Status = currentStatuses[i]
				case "resource3":
					resource3Status = currentStatuses[i]
				}
			}

			Expect(len(resource1Status.Outputs)).To(Equal(2))
			Expect(resource1Status.Outputs[0]).To(MatchFields(IgnoreExtras,
				Fields{
					"Name":    Equal("url"),
					"Preview": Equal("hi\n"),
					"Digest":  HavePrefix("sha256"),
				},
			))
			Expect(resource1Status.Outputs[1]).To(MatchFields(IgnoreExtras,
				Fields{
					"Name":    Equal("revision"),
					"Preview": Equal("bye\n"),
					"Digest":  HavePrefix("sha256"),
				},
			))
			Expect(resource1Status.Outputs[0].LastTransitionTime).ToNot(Equal(previousTime))
			Expect(resource1Status.Outputs[1].LastTransitionTime).ToNot(Equal(previousTime))

			Expect(len(resource2Status.Outputs)).To(Equal(1))
			Expect(resource2Status.Outputs[0].LastTransitionTime).To(Equal(previousTime))

			Expect(len(resource3Status.Outputs)).To(Equal(1))
			Expect(resource3Status.Outputs[0].LastTransitionTime).To(Equal(previousTime))

			Expect(rec.ResourceEventfCallCount()).To(Equal(2))

			evType, reason, messageFmt, resourceObj, fmtArgs := rec.ResourceEventfArgsForCall(0)
			Expect(evType).To(Equal("Normal"))
			Expect(reason).To(Equal(events.ResourceOutputChangedReason))
			Expect(messageFmt).To(Equal("[%s] found a new output in [%Q]"))
			Expect(fmtArgs).To(Equal([]interface{}{"resource1"}))
			Expect(resourceObj).To(Equal(stampedObj1))

			evType, reason, messageFmt, resourceObj, fmtArgs = rec.ResourceEventfArgsForCall(1)
			Expect(evType).To(Equal("Normal"))
			Expect(reason).To(Equal(events.ResourceHealthyStatusChangedReason))
			Expect(messageFmt).To(Equal("[%s] found healthy status in [%Q] changed to [%s]"))
			Expect(fmtArgs).To(Equal([]interface{}{"resource1", metav1.ConditionTrue}))
			Expect(resourceObj).To(Equal(stampedObj1))
		})

		Context("there is an error realizing resource 1 and resource 2", func() {
			BeforeEach(func() {
				resourceRealizer.DoReturnsOnCall(0, nil, nil, nil, errors.New("im in a bad state"))
				resourceRealizer.DoReturnsOnCall(1, nil, nil, nil, errors.New("im missing inputs"))

				obj = &unstructured.Unstructured{}
				obj.SetName("StampedObj")

				resourceRealizer.DoReturnsOnCall(2, templateModel3, obj, nil, nil)
			})

			It("the status uses the previous resource for resource 2", func() {
				resourceStatuses := statuses.NewResourceStatuses(previousResources, conditions.AddConditionForResourceSubmittedWorkload)

				err := rlzr.Realize(ctx, resourceRealizer, supplyChain.Name, realizer.MakeSupplychainOwnerResources(supplyChain), resourceStatuses)
				Expect(err).To(MatchError("im in a bad state"))

				Expect(evaluatedRealizedResourceNames).To(Equal([]string{"resource3"}))

				currentStatuses := resourceStatuses.GetCurrent()
				Expect(currentStatuses).To(HaveLen(3))

				var resource1Status v1alpha1.ResourceStatus
				var resource2Status v1alpha1.ResourceStatus
				var resource3Status v1alpha1.ResourceStatus

				for i := range currentStatuses {
					switch currentStatuses[i].Name {
					case "resource1":
						resource1Status = currentStatuses[i]
					case "resource2":
						resource2Status = currentStatuses[i]
					case "resource3":
						resource3Status = currentStatuses[i]
					}
				}

				// Error realizing resource1, no previous resource, realizedResource should be nil
				Expect(resource1Status.RealizedResource.Name).To(Equal("resource1"))
				Expect(resource1Status.RealizedResource.StampedRef).To(BeNil())
				Expect(resource1Status.RealizedResource.TemplateRef).To(BeNil())
				Expect(resource1Status.RealizedResource.Inputs).To(BeNil())
				Expect(resource1Status.RealizedResource.Outputs).To(BeNil())

				// Error realizing resource2, realizedResource should be previous resource
				Expect(resource2Status.Name).To(Equal(previousResources[0].Name))
				Expect(resource2Status.StampedRef).To(Equal(previousResources[0].StampedRef))
				Expect(resource2Status.TemplateRef).To(Equal(previousResources[0].TemplateRef))
				Expect(resource2Status.Inputs).To(Equal(previousResources[0].Inputs))
				Expect(resource2Status.Outputs).To(Equal(previousResources[0].Outputs))
				Expect(resource2Status.Conditions).ToNot(Equal(previousResources[0].Conditions))
				Expect(len(resource2Status.Conditions)).To(Equal(3))

				Expect(resource2Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
					"Type":   Equal("ResourceSubmitted"),
					"Status": Equal(metav1.ConditionFalse),
				})))
				Expect(resource2Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
					"Type":   Equal("Healthy"),
					"Status": Equal(metav1.ConditionTrue),
					"Reason": Equal("HealthyReasonFromBefore"),
				})))
				Expect(resource2Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
					"Type":   Equal("Ready"),
					"Status": Equal(metav1.ConditionFalse),
				})))

				// No error realizing resource3, realizedResource should be a new resource
				Expect(resource3Status.Name).To(Equal(previousResources[1].Name))
				Expect(resource3Status.StampedRef).ToNot(Equal(previousResources[1].StampedRef))
				Expect(resource3Status.TemplateRef).ToNot(Equal(previousResources[1].TemplateRef))
				Expect(len(resource3Status.Conditions)).To(Equal(3))

				Expect(resource3Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
					"Type":   Equal("ResourceSubmitted"),
					"Status": Equal(metav1.ConditionTrue),
				})))
				Expect(resource3Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
					"Type":   Equal("Healthy"),
					"Status": Equal(metav1.ConditionTrue),
					"Reason": Equal("EvaluatorSaysSo"),
				})))
				Expect(resource3Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
					"Type":   Equal("Ready"),
					"Status": Equal(metav1.ConditionTrue),
				})))
			})
		})
	})
})
