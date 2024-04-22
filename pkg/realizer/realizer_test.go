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
	"github.com/vmware-tanzu/cartographer/pkg/controllers"
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
		rlzr                           controllers.Realizer
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
		rec.EventfCalls(func(eventtype, reason, messageFmt string, i ...interface{}) {
			recordedEvents = append(recordedEvents, event{eventtype, reason, messageFmt, "", i})
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
			resource1 = v1alpha1.SupplyChainResource{
				Name: "resource1",
				TemplateRef: v1alpha1.SupplyChainTemplateReference{
					Kind: template1.Kind,
					Name: template1.Name,
				},
			}
			resource2 = v1alpha1.SupplyChainResource{
				Name: "resource2",
				TemplateRef: v1alpha1.SupplyChainTemplateReference{
					Kind: template2.Kind,
					Name: template2.Name,
				},
				Images: []v1alpha1.ResourceReference{
					{
						Name:     "my-image",
						Resource: "resource1",
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

			resourceRealizer.DoCalls(func(ctx context.Context, resource realizer.OwnerResource, blueprintName string, outputs realizer.Outputs, mapper meta.RESTMapper) (templates.Reader, *unstructured.Unstructured, *templates.Output, bool, string, error) {
				executedResourceOrder = append(executedResourceOrder, resource.Name)
				Expect(blueprintName).To(Equal("greatest-supply-chain"))
				if resource.Name == "resource1" {
					Expect(outputs).To(Equal(realizer.NewOutputs()))
					reader, err := templates.NewReaderFromAPI(template1)
					Expect(err).NotTo(HaveOccurred())
					stampedObj := &unstructured.Unstructured{}
					stampedObj.SetName("obj1")
					return reader, stampedObj, outputFromFirstResource, false, "returned val that would generally equal template 1 name", nil
				}

				if resource.Name == "resource2" {
					expectedSecondResourceOutputs := realizer.NewOutputs()
					expectedSecondResourceOutputs.AddOutput("resource1", outputFromFirstResource)
					Expect(outputs).To(Equal(expectedSecondResourceOutputs))
				}
				reader, err := templates.NewReaderFromAPI(template2)
				Expect(err).NotTo(HaveOccurred())
				stampedObj := &unstructured.Unstructured{}
				stampedObj.SetName("obj2")
				return reader, stampedObj, &templates.Output{}, false, "returned val that would generally equal template 2 name", nil
			})

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
			Expect(currentResourceStatuses[0].TemplateRef.Name).To(Equal("returned val that would generally equal template 1 name"))
			Expect(currentResourceStatuses[0].StampedRef.Name).To(Equal("obj1"))
			Expect(currentResourceStatuses[0].StampedRef.Resource).To(Equal("FOO.EXAMPLE.COM"))
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
			Expect(currentResourceStatuses[1].TemplateRef.Name).To(Equal("returned val that would generally equal template 2 name"))
			Expect(currentResourceStatuses[1].StampedRef.Name).To(Equal("obj2"))
			Expect(currentResourceStatuses[1].StampedRef.Resource).To(Equal("FOO.EXAMPLE.COM"))
			Expect(len(currentResourceStatuses[1].Inputs)).To(Equal(1))
			Expect(currentResourceStatuses[1].Inputs).To(Equal([]v1alpha1.Input{{Name: "resource1"}}))
			Expect(currentResourceStatuses[1].Outputs).To(BeNil())
			Expect(len(currentResourceStatuses[1].Conditions)).To(Equal(3))

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

		Context("the first resource returns an error and the second does not", func() {
			var (
				err error
			)
			BeforeEach(func() {
				var reader templates.Reader
				reader, err = templates.NewReaderFromAPI(template2)
				Expect(err).NotTo(HaveOccurred())
				resourceRealizer.DoReturnsOnCall(0, nil, nil, nil, false, "", errors.New("realizing is hard"))
				resourceRealizer.DoReturnsOnCall(1, reader, &unstructured.Unstructured{}, nil, false, resource2.TemplateRef.Name, nil)
			})

			It("returns the first error encountered and continues to realize", func() {
				resourceStatuses := statuses.NewResourceStatuses(nil, conditions.AddConditionForResourceSubmittedWorkload)
				err = rlzr.Realize(ctx, resourceRealizer, supplyChain.Name, realizer.MakeSupplychainOwnerResources(supplyChain), resourceStatuses)

				Expect(err).To(MatchError("realizing is hard"))
				rs := *resourceStatuses
				currentResourceStatuses := rs.GetCurrent()
				Expect(currentResourceStatuses).To(HaveLen(2))

				Expect(currentResourceStatuses[0].Name).To(Equal("resource1"))
				Expect(currentResourceStatuses[0].TemplateRef).To(BeNil())
				Expect(currentResourceStatuses[0].StampedRef).To(BeNil())
				Expect(currentResourceStatuses[1].TemplateRef.Name).To(Equal(template2.Name))
			})
		})
	})

	Context("one of the resources is passed through", func() {
		var (
			template1             *v1alpha1.ClusterImageTemplate
			executedResourceOrder []string
			supplyChain           *v1alpha1.ClusterSupplyChain
			resource1             v1alpha1.SupplyChainResource
			resource2             v1alpha1.SupplyChainResource
		)
		BeforeEach(func() {
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
			resource1 = v1alpha1.SupplyChainResource{
				Name: "resource1",
				TemplateRef: v1alpha1.SupplyChainTemplateReference{
					Kind: template1.Kind,
					Name: template1.Name,
				},
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
			supplyChain = &v1alpha1.ClusterSupplyChain{
				ObjectMeta: metav1.ObjectMeta{Name: "greatest-supply-chain"},
				Spec: v1alpha1.SupplyChainSpec{
					Resources: []v1alpha1.SupplyChainResource{resource1, resource2},
				},
			}

			outputFromFirstResource := &templates.Output{Image: "whatever"}

			resourceRealizer.DoCalls(func(ctx context.Context, resource realizer.OwnerResource, blueprintName string, outputs realizer.Outputs, mapper meta.RESTMapper) (templates.Reader, *unstructured.Unstructured, *templates.Output, bool, string, error) {
				executedResourceOrder = append(executedResourceOrder, resource.Name)
				Expect(blueprintName).To(Equal("greatest-supply-chain"))
				if resource.Name == "resource1" {
					Expect(outputs).To(Equal(realizer.NewOutputs()))
					reader, err := templates.NewReaderFromAPI(template1)
					Expect(err).NotTo(HaveOccurred())
					stampedObj := &unstructured.Unstructured{}
					stampedObj.SetName("obj1")
					return reader, stampedObj, outputFromFirstResource, false, resource.TemplateRef.Name, nil
				}

				if resource.Name == "resource2" {
					expectedSecondResourceOutputs := realizer.NewOutputs()
					expectedSecondResourceOutputs.AddOutput("resource1", outputFromFirstResource)
					Expect(outputs).To(Equal(expectedSecondResourceOutputs))
				}

				return nil, nil, outputFromFirstResource, true, "field not leveraged when pass-through", nil
			})

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

		It("records an event for resource output changes", func() {
			resourceStatuses := statuses.NewResourceStatuses(nil, conditions.AddConditionForResourceSubmittedWorkload)
			Expect(rlzr.Realize(ctx, resourceRealizer, supplyChain.Name, realizer.MakeSupplychainOwnerResources(supplyChain), resourceStatuses)).To(Succeed())

			Expect(recordedEvents).To(ConsistOf(
				event{"Normal", events.ResourceOutputChangedReason, "[%s] found a new output in [%Q]", "obj1", []interface{}{"resource1"}},
				event{"Normal", events.ResourceHealthyStatusChangedReason, "[%s] found healthy status in [%Q] changed to [%s]", "obj1", []interface{}{"resource1", metav1.ConditionTrue}},
				event{"Normal", events.ResourceOutputChangedReason, "[%s] passed through a new output", "", []interface{}{"resource2"}},
			))
		})

		It("does not record an event if there was no resource output change", func() {
			previousResources := []v1alpha1.ResourceStatus{
				{
					RealizedResource: v1alpha1.RealizedResource{
						Name: "resource2",
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

			Expect(recordedEvents).NotTo(ContainElement(MatchFields(IgnoreExtras, Fields{"Message": ContainSubstring("passed through")})))
		})

		It("generates the correct realized resource", func() {
			resourceStatuses := statuses.NewResourceStatuses(nil, conditions.AddConditionForResourceSubmittedWorkload)
			err := rlzr.Realize(ctx, resourceRealizer, supplyChain.Name, realizer.MakeSupplychainOwnerResources(supplyChain), resourceStatuses)
			Expect(err).ToNot(HaveOccurred())

			currentResourceStatuses := resourceStatuses.GetCurrent()

			Expect(currentResourceStatuses).To(HaveLen(2))

			Expect(currentResourceStatuses[0].Name).To(Equal(resource1.Name))
			Expect(currentResourceStatuses[0].TemplateRef.Name).To(Equal(template1.Name))
			Expect(currentResourceStatuses[0].StampedRef.Name).To(Equal("obj1"))
			Expect(currentResourceStatuses[0].StampedRef.Resource).To(Equal("FOO.EXAMPLE.COM"))
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
			Expect(currentResourceStatuses[1].TemplateRef).To(BeNil())
			Expect(currentResourceStatuses[1].StampedRef).To(BeNil())
			Expect(len(currentResourceStatuses[1].Inputs)).To(Equal(1))
			Expect(currentResourceStatuses[1].Inputs).To(Equal([]v1alpha1.Input{{Name: "resource1"}}))
			Expect(currentResourceStatuses[1].Outputs[0]).To(MatchFields(IgnoreExtras,
				Fields{
					"Name":    Equal("image"),
					"Preview": Equal("whatever\n"),
					"Digest":  HavePrefix("sha256"),
				},
			))
			Expect(len(currentResourceStatuses[1].Conditions)).To(Equal(2))

			Expect(currentResourceStatuses[1].Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal("ResourceSubmitted"),
				"Status": Equal(metav1.ConditionTrue),
				"Reason": Equal("PassThrough"),
			})))
			Expect(currentResourceStatuses[1].Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal("Ready"),
				"Status": Equal(metav1.ConditionTrue),
			})))
		})
	})

	Context("there are previous resources", func() {
		var (
			reader1           templates.Reader
			reader2           templates.Reader
			reader3           templates.Reader
			obj               *unstructured.Unstructured
			previousResources []v1alpha1.ResourceStatus
			previousTime      metav1.Time
			supplyChain       *v1alpha1.ClusterSupplyChain
			resource2         v1alpha1.SupplyChainResource
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
							Resource: "image",
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
							Resource: "config",
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
			resource2 = v1alpha1.SupplyChainResource{
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
			reader1, err = templates.NewReaderFromAPI(template1)
			Expect(err).NotTo(HaveOccurred())

			template2 := &v1alpha1.ClusterImageTemplate{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-image-template",
				},
			}
			reader2, err = templates.NewReaderFromAPI(template2)
			Expect(err).NotTo(HaveOccurred())

			template3 := &v1alpha1.ClusterConfigTemplate{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-config-template",
				},
			}
			reader3, err = templates.NewReaderFromAPI(template3)
			Expect(err).NotTo(HaveOccurred())

			resourceRealizer.DoReturnsOnCall(0, reader1, &unstructured.Unstructured{}, nil, false, "first expected name", nil)
			resourceRealizer.DoReturnsOnCall(1, reader2, &unstructured.Unstructured{}, nil, false, resource2.Name, nil)
			resourceRealizer.DoReturnsOnCall(2, reader3, &unstructured.Unstructured{}, nil, false, resource3.Name, nil)

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
			resourceRealizer.DoReturnsOnCall(0, reader1, stampedObj1, newOutput, false, "", nil)

			oldOutput := &templates.Output{
				Image: "whatever",
			}
			resourceRealizer.DoReturnsOnCall(1, reader2, &unstructured.Unstructured{}, oldOutput, false, "", nil)

			oldOutput2 := &templates.Output{
				Config: "whatever",
			}
			resourceRealizer.DoReturnsOnCall(2, reader3, obj, oldOutput2, false, "", nil)

			resourceStatuses := statuses.NewResourceStatuses(previousResources, conditions.AddConditionForResourceSubmittedWorkload)
			err := rlzr.Realize(ctx, resourceRealizer, supplyChain.Name, realizer.MakeSupplychainOwnerResources(supplyChain), resourceStatuses)
			Expect(err).ToNot(HaveOccurred())

			currentStatuses := resourceStatuses.GetCurrent()
			var resource1Status v1alpha1.ResourceStatus
			var resource2Status v1alpha1.ResourceStatus
			var resource3Status v1alpha1.ResourceStatus

			for i := range currentStatuses {
				Expect(currentStatuses[i].Name).To(Equal(fmt.Sprintf("resource%d", i+1)))
			}

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

		Context("the supply chain has changed to have fewer resources", func() {
			var currentStatuses []v1alpha1.ResourceStatus
			BeforeEach(func() {
				supplyChain = &v1alpha1.ClusterSupplyChain{
					ObjectMeta: metav1.ObjectMeta{Name: "greatest-supply-chain"},
					Spec: v1alpha1.SupplyChainSpec{
						Resources: []v1alpha1.SupplyChainResource{resource2},
					},
				}

				template2 := &v1alpha1.ClusterImageTemplate{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "my-image-template",
					},
				}
				reader2, err := templates.NewReaderFromAPI(template2)
				Expect(err).NotTo(HaveOccurred())

				resourceRealizer.DoReturnsOnCall(0, reader2, &unstructured.Unstructured{}, nil, false, resource2.Name, nil)

				oldOutput := &templates.Output{
					Image: "whatever",
				}
				resourceRealizer.DoReturnsOnCall(0, reader2, &unstructured.Unstructured{}, oldOutput, false, "", nil)

				resourceStatuses := statuses.NewResourceStatuses(previousResources, conditions.AddConditionForResourceSubmittedWorkload)
				err = rlzr.Realize(ctx, resourceRealizer, supplyChain.Name, realizer.MakeSupplychainOwnerResources(supplyChain), resourceStatuses)
				Expect(err).ToNot(HaveOccurred())

				currentStatuses = resourceStatuses.GetCurrent()
			})
			It("only creates 1 resource status in currentStatuses", func() {
				Expect(currentStatuses[0].Name).To(Equal(fmt.Sprintf("resource2")))
				Expect(len(currentStatuses)).To(Equal(1))
			})
		})

		Context("there is an error realizing resource 1 and resource 2", func() {
			BeforeEach(func() {
				resourceRealizer.DoReturnsOnCall(0, nil, nil, nil, false, "", errors.New("im in a bad state"))
				resourceRealizer.DoReturnsOnCall(1, nil, nil, nil, false, "", errors.New("im missing inputs"))

				obj = &unstructured.Unstructured{}
				obj.SetName("StampedObj")

				resourceRealizer.DoReturnsOnCall(2, reader3, obj, nil, false, "expected name for resource 3", nil)
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
				Expect(resource3Status.RealizedResource.TemplateRef.Name).To(Equal("expected name for resource 3"))
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
