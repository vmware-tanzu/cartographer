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

package deliverable_test

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/deliverable"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/deliverable/deliverablefakes"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

var _ = Describe("Realize", func() {
	var (
		resourceRealizer *deliverablefakes.FakeResourceRealizer
		rlzr             realizer.Realizer
		ctx              context.Context
	)
	BeforeEach(func() {
		ctx = context.Background()

		rlzr = realizer.NewRealizer()

		resourceRealizer = &deliverablefakes.FakeResourceRealizer{}
	})

	Context("there are no previous resources", func() {
		var (
			delivery              *v1alpha1.ClusterDelivery
			resource1             v1alpha1.DeliveryResource
			resource2             v1alpha1.DeliveryResource
			template1             *v1alpha1.ClusterSourceTemplate
			template2             *v1alpha1.ClusterTemplate
			executedResourceOrder []string
		)
		BeforeEach(func() {
			resource1 = v1alpha1.DeliveryResource{
				Name: "resource1",
			}
			resource2 = v1alpha1.DeliveryResource{
				Name: "resource2",
				Configs: []v1alpha1.ResourceReference{
					{
						Name:     "my-config",
						Resource: "resource1",
					},
				},
			}
			template1 = &v1alpha1.ClusterSourceTemplate{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-source-template",
				},
			}
			template2 = &v1alpha1.ClusterTemplate{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-config-template",
				},
			}
			delivery = &v1alpha1.ClusterDelivery{
				ObjectMeta: metav1.ObjectMeta{Name: "greatest-delivery"},
				Spec: v1alpha1.DeliverySpec{
					Resources: []v1alpha1.DeliveryResource{resource1, resource2},
				},
			}

			outputFromFirstResource := &templates.Output{Source: &templates.Source{
				URL:      "whatever",
				Revision: "whatever-rev",
			}}

			resourceRealizer.DoCalls(func(ctx context.Context, resource *v1alpha1.DeliveryResource, deliveryName string, outputs realizer.Outputs) (templates.Template, *unstructured.Unstructured, *templates.Output, error) {
				executedResourceOrder = append(executedResourceOrder, resource.Name)
				Expect(deliveryName).To(Equal("greatest-delivery"))
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
		})

		It("realizes each resource in delivery order, accumulating output for each subsequent resource", func() {
			realizedResources, err := rlzr.Realize(ctx, resourceRealizer, delivery, nil)
			Expect(err).To(Succeed())

			Expect(executedResourceOrder).To(Equal([]string{"resource1", "resource2"}))

			Expect(realizedResources).To(HaveLen(2))

			Expect(realizedResources[0].Name).To(Equal(resource1.Name))
			Expect(realizedResources[0].TemplateRef.Name).To(Equal(template1.Name))
			Expect(realizedResources[0].StampedRef.Name).To(Equal("obj1"))
			Expect(realizedResources[0].Inputs).To(BeNil())
			Expect(len(realizedResources[0].Outputs)).To(Equal(2))
			Expect(realizedResources[0].Outputs[0]).To(MatchFields(IgnoreExtras,
				Fields{
					"Name":    Equal("url"),
					"Preview": Equal("whatever"),
					"Digest":  Equal("sha256:85738f8f9a7f1b04b5329c590ebcb9e425925c6d0984089c43a022de4f19c281"),
				},
			))
			Expect(realizedResources[0].Outputs[1]).To(MatchFields(IgnoreExtras,
				Fields{
					"Name":    Equal("revision"),
					"Preview": Equal("whatever-rev"),
					"Digest":  Equal("sha256:4829ca8682c2089cde1ddd694c5b27f7912e2cdebb6c61b6c5e36c93a534aab1"),
				},
			))
			Expect(time.Since(realizedResources[0].Outputs[0].LastTransitionTime.Time)).To(BeNumerically("<", time.Second))

			Expect(realizedResources[1].Name).To(Equal(resource2.Name))
			Expect(realizedResources[1].TemplateRef.Name).To(Equal(template2.Name))
			Expect(realizedResources[1].StampedRef.Name).To(Equal("obj2"))
			Expect(len(realizedResources[1].Inputs)).To(Equal(1))
			Expect(realizedResources[1].Inputs).To(Equal([]v1alpha1.Input{{Name: "resource1"}}))
			Expect(realizedResources[1].Outputs).To(BeNil())
		})

		It("returns the first error encountered realizing a resource and continues to realize", func() {
			template, err := templates.NewModelFromAPI(template2)
			Expect(err).NotTo(HaveOccurred())
			resourceRealizer.DoReturnsOnCall(0, nil, nil, nil, errors.New("realizing is hard"))
			resourceRealizer.DoReturnsOnCall(1, template, &unstructured.Unstructured{}, nil, nil)

			realizedResources, err := rlzr.Realize(ctx, resourceRealizer, delivery, nil)
			Expect(err).To(MatchError("realizing is hard"))
			Expect(realizedResources).To(HaveLen(2))

			Expect(realizedResources[0].Name).To(Equal("resource1"))
			Expect(realizedResources[0].TemplateRef).To(BeNil())
			Expect(realizedResources[0].StampedRef).To(BeNil())
			Expect(realizedResources[1].TemplateRef.Name).To(Equal(template2.Name))
		})
	})

	Context("there are previous resources", func() {
		var (
			templateModel1    templates.Template
			templateModel2    templates.Template
			templateModel3    templates.Template
			obj               *unstructured.Unstructured
			previousResources []v1alpha1.RealizedResource
			previousTime      metav1.Time
			delivery          *v1alpha1.ClusterDelivery
		)
		BeforeEach(func() {
			previousTime = metav1.NewTime(time.Now())
			previousResources = []v1alpha1.RealizedResource{
				{
					Name: "resource1",
					StampedRef: &corev1.ObjectReference{
						Kind:       "GitRepository",
						Namespace:  "",
						Name:       "",
						APIVersion: "",
					},
					TemplateRef: &corev1.ObjectReference{
						Kind:       "ClusterSourceTemplate",
						Name:       "my-source-template",
						APIVersion: "",
					},
					Inputs: nil,
					Outputs: []v1alpha1.Output{
						{
							Name:               "url",
							Preview:            "http://example.com",
							Digest:             "sha256:85738f8f9a7f1b04b5329c590ebcb9e425925c6d0984089c43a022de4f19c281",
							LastTransitionTime: previousTime,
						},
						{
							Name:               "revision",
							Preview:            "main",
							Digest:             "sha256:85738f8f9a7f1b04b5329c590ebcb9e425925c6d0984089c43a022de4f19c281",
							LastTransitionTime: previousTime,
						},
					},
				},
				{
					Name: "resource2",
					StampedRef: &corev1.ObjectReference{
						Kind:       "Image",
						Namespace:  "",
						Name:       "",
						APIVersion: "",
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
							Name:               "config",
							Preview:            "whatever",
							Digest:             "sha256:85738f8f9a7f1b04b5329c590ebcb9e425925c6d0984089c43a022de4f19c281",
							LastTransitionTime: previousTime,
						},
					},
				},
				{
					Name: "resource3",
					StampedRef: &corev1.ObjectReference{
						Kind:       "Config",
						Namespace:  "",
						Name:       "PreviousStampedObj",
						APIVersion: "",
					},
					TemplateRef: &corev1.ObjectReference{
						Kind:       "ClusterConfigTemplate",
						Name:       "my-config-template",
						APIVersion: "",
					},
					Outputs: []v1alpha1.Output{
						{
							Name:               "config",
							Preview:            "whatever",
							Digest:             "sha256:85738f8f9a7f1b04b5329c590ebcb9e425925c6d0984089c43a022de4f19c281",
							LastTransitionTime: previousTime,
						},
					},
				},
			}

			resource1 := v1alpha1.DeliveryResource{
				Name: "resource1",
			}
			resource2 := v1alpha1.DeliveryResource{
				Name: "resource2",
				Sources: []v1alpha1.ResourceReference{
					{
						Name:     "my-source",
						Resource: "resource1",
					},
				},
			}
			resource3 := v1alpha1.DeliveryResource{
				Name: "resource3",
			}

			delivery = &v1alpha1.ClusterDelivery{
				ObjectMeta: metav1.ObjectMeta{Name: "greatest-supply-chain"},
				Spec: v1alpha1.DeliverySpec{
					Resources: []v1alpha1.DeliveryResource{resource1, resource2, resource3},
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

			template2 := &v1alpha1.ClusterConfigTemplate{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-config-template",
				},
			}
			templateModel2, err = templates.NewModelFromAPI(template2)
			Expect(err).NotTo(HaveOccurred())

			template3 := &v1alpha1.ClusterConfigTemplate{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-config-2-template",
				},
			}
			templateModel3, err = templates.NewModelFromAPI(template3)
			Expect(err).NotTo(HaveOccurred())

			resourceRealizer.DoReturnsOnCall(0, templateModel1, &unstructured.Unstructured{}, nil, nil)
			resourceRealizer.DoReturnsOnCall(1, templateModel2, &unstructured.Unstructured{}, nil, nil)
			resourceRealizer.DoReturnsOnCall(2, templateModel3, &unstructured.Unstructured{}, nil, nil)
		})

		It("realizes each resource and does not update last transition time since the resource has not changed", func() {
			newOutput := &templates.Output{
				Source: &templates.Source{
					URL:      "hi",
					Revision: "bye",
				},
			}
			resourceRealizer.DoReturnsOnCall(0, templateModel1, &unstructured.Unstructured{}, newOutput, nil)

			oldOutput := &templates.Output{
				Config: "whatever",
			}
			resourceRealizer.DoReturnsOnCall(1, templateModel2, &unstructured.Unstructured{}, oldOutput, nil)

			oldOutput2 := &templates.Output{
				Config: "whatever",
			}
			resourceRealizer.DoReturnsOnCall(2, templateModel3, obj, oldOutput2, nil)

			realizedResources, err := rlzr.Realize(context.TODO(), resourceRealizer, delivery, previousResources)
			Expect(err).ToNot(HaveOccurred())

			Expect(len(realizedResources[0].Outputs)).To(Equal(2))
			Expect(realizedResources[0].Outputs[0]).To(MatchFields(IgnoreExtras,
				Fields{
					"Name":    Equal("url"),
					"Preview": Equal("hi"),
					"Digest":  Equal("sha256:8f434346648f6b96df89dda901c5176b10a6d83961dd3c1ac88b59b2dc327aa4"),
				},
			))
			Expect(realizedResources[0].Outputs[1]).To(MatchFields(IgnoreExtras,
				Fields{
					"Name":    Equal("revision"),
					"Preview": Equal("bye"),
					"Digest":  Equal("sha256:b49f425a7e1f9cff3856329ada223f2f9d368f15a00cf48df16ca95986137fe8"),
				},
			))
			Expect(realizedResources[0].Outputs[0].LastTransitionTime).ToNot(Equal(previousTime))
			Expect(realizedResources[0].Outputs[1].LastTransitionTime).ToNot(Equal(previousTime))

			Expect(len(realizedResources[1].Outputs)).To(Equal(1))
			Expect(realizedResources[1].Outputs[0].LastTransitionTime).To(Equal(previousTime))

			Expect(len(realizedResources[2].Outputs)).To(Equal(1))
			Expect(realizedResources[2].Outputs[0].LastTransitionTime).To(Equal(previousTime))
		})

		Context("there is an error realizing resource 1", func() {
			BeforeEach(func() {
				resourceRealizer.DoReturnsOnCall(0, nil, nil, nil, errors.New("im in a bad state"))
				resourceRealizer.DoReturnsOnCall(1, nil, nil, nil, errors.New("im missing inputs"))

				obj = &unstructured.Unstructured{}
				obj.SetName("StampedObj")

				resourceRealizer.DoReturnsOnCall(2, templateModel3, obj, nil, nil)
			})

			It("the status uses the previous resource for resource 1 and resource 2", func() {
				realizedResources, err := rlzr.Realize(context.TODO(), resourceRealizer, delivery, previousResources)
				Expect(err).To(MatchError("im in a bad state"))
				Expect(realizedResources).To(HaveLen(3))

				Expect(realizedResources[0]).To(Equal(previousResources[0]))
				Expect(realizedResources[1]).To(Equal(previousResources[1]))
				Expect(realizedResources[2]).ToNot(Equal(previousResources[2]))
				Expect(realizedResources[2].StampedRef.Name).To(Equal("StampedObj"))
			})
		})
	})
})
