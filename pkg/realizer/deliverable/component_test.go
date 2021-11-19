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
	"encoding/json"
	"errors"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/deliverable"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

var _ = Describe("Resource", func() {

	var (
		ctx            context.Context
		resource       v1alpha1.ClusterDeliveryResource
		deliverable    v1alpha1.Deliverable
		outputs        realizer.Outputs
		deliveryName   string
		deliveryParams []v1alpha1.OverridableParam
		fakeRepo       repositoryfakes.FakeRepository
		r              realizer.ResourceRealizer
	)

	BeforeEach(func() {
		ctx = context.Background()
		resource = v1alpha1.ClusterDeliveryResource{
			Name: "resource-1",
			TemplateRef: v1alpha1.DeliveryClusterTemplateReference{
				Kind: "ClusterSourceTemplate",
				Name: "source-template-1",
			},
		}

		deliveryName = "delivery-name"

		deliveryParams = []v1alpha1.OverridableParam{}

		outputs = realizer.NewOutputs()

		fakeRepo = repositoryfakes.FakeRepository{}
		deliverable = v1alpha1.Deliverable{}
		r = realizer.NewResourceRealizer(&deliverable, &fakeRepo)
	})

	Describe("Do", func() {
		When("passed a deliverable with outputs", func() {
			BeforeEach(func() {
				resource.Sources = []v1alpha1.ResourceReference{
					{
						Name:     "source-provider",
						Resource: "previous-resource",
					},
				}

				outputs.AddOutput("previous-resource", &templates.Output{Source: &templates.Source{
					URL:      "some-url",
					Revision: "some-revision",
				}})

				configMap := &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "example-config-map",
						Namespace: "some-namespace",
					},
					Data: map[string]string{
						"player_current_lives": `$(source.url)$`,
						"some_other_info":      `$(sources.source-provider.revision)$`,
					},
				}

				dbytes, err := json.Marshal(configMap)
				Expect(err).ToNot(HaveOccurred())

				templateAPI := &v1alpha1.ClusterSourceTemplate{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterSourceTemplate",
						APIVersion: "carto.run/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "source-template-1",
						Namespace: "some-namespace",
					},
					Spec: v1alpha1.SourceTemplateSpec{
						TemplateSpec: v1alpha1.TemplateSpec{
							Template: &runtime.RawExtension{Raw: dbytes},
						},
						URLPath:      "data.player_current_lives",
						RevisionPath: "data.some_other_info",
					},
				}

				fakeRepo.GetDeliveryClusterTemplateReturns(templateAPI, nil)
				fakeRepo.EnsureObjectExistsOnClusterReturns(nil)
			})

			It("creates a stamped object and returns the outputs and stampedObjects", func() {
				returnedStampedObject, out, err := r.Do(ctx, &resource, deliveryName, deliveryParams, outputs)
				Expect(err).ToNot(HaveOccurred())

				actualCtx, stampedObject, allowUpdate := fakeRepo.EnsureObjectExistsOnClusterArgsForCall(0)
				Expect(actualCtx).To(Equal(ctx))

				Expect(returnedStampedObject).To(Equal(stampedObject))
				Expect(allowUpdate).To(BeTrue())

				metadata := stampedObject.Object["metadata"]
				metadataValues, ok := metadata.(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(metadataValues["name"]).To(Equal("example-config-map"))
				Expect(metadataValues["namespace"]).To(Equal("some-namespace"))
				Expect(metadataValues["ownerReferences"]).To(Equal([]interface{}{
					map[string]interface{}{
						"apiVersion":         "",
						"kind":               "",
						"name":               "",
						"uid":                "",
						"controller":         true,
						"blockOwnerDeletion": true,
					},
				}))
				Expect(metadataValues["labels"]).To(Equal(map[string]interface{}{
					"carto.run/cluster-delivery-name": "delivery-name",
					"carto.run/resource-name":         "resource-1",
					"carto.run/cluster-template-name": "source-template-1",
					"carto.run/deliverable-name":      "",
					"carto.run/deliverable-namespace": "",
					"carto.run/template-kind":         "ClusterSourceTemplate",
				}))
				Expect(stampedObject.Object["data"]).To(Equal(map[string]interface{}{"player_current_lives": "some-url", "some_other_info": "some-revision"}))

				Expect(out.Source.Revision).To(Equal("some-revision"))
				Expect(out.Source.URL).To(Equal("some-url"))
			})
		})

		When("unable to get the template ref from repo", func() {
			BeforeEach(func() {
				fakeRepo.GetDeliveryClusterTemplateReturns(nil, errors.New("bad template"))
			})

			It("returns GetDeliveryClusterTemplateError", func() {
				_, _, err := r.Do(ctx, &resource, deliveryName, deliveryParams, outputs)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("unable to get template 'source-template-1'"))
				Expect(err.Error()).To(ContainSubstring("bad template"))
				Expect(reflect.TypeOf(err).String()).To(Equal("deliverable.GetDeliveryClusterTemplateError"))
			})
		})

		When("unable to create a template model from apiTemplate", func() {
			BeforeEach(func() {
				templateAPI := &v1alpha1.Workload{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "carto.run/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "not-a-template",
						Namespace: "some-namespace",
					},
				}

				fakeRepo.GetClusterTemplateReturns(templateAPI, nil)
			})

			It("returns a helpful error", func() {
				_, _, err := r.Do(ctx, &resource, deliveryName, deliveryParams, outputs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("new model from api:"))
			})
		})

		When("unable to Stamp a new template", func() {
			BeforeEach(func() {
				templateAPI := &v1alpha1.ClusterSourceTemplate{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterSourceTemplate",
						APIVersion: "carto.run/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "source-template-1",
						Namespace: "some-namespace",
					},
					Spec: v1alpha1.SourceTemplateSpec{
						TemplateSpec: v1alpha1.TemplateSpec{
							Template: &runtime.RawExtension{},
						},
					},
				}

				fakeRepo.GetDeliveryClusterTemplateReturns(templateAPI, nil)
			})

			It("returns StampError", func() {
				_, _, err := r.Do(ctx, &resource, deliveryName, deliveryParams, outputs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unable to stamp object for resource 'resource-1'"))
				Expect(reflect.TypeOf(err).String()).To(Equal("deliverable.StampError"))
			})
		})

		When("unable to retrieve the output from the stamped object", func() {
			BeforeEach(func() {
				configMap := &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "example-config-map",
						Namespace: "some-namespace",
					},
					Data: map[string]string{
						"player_current_lives": "9",
						"some_other_info":      "10",
					},
				}

				dbytes, err := json.Marshal(configMap)
				Expect(err).ToNot(HaveOccurred())

				templateAPI := &v1alpha1.ClusterSourceTemplate{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterSourceTemplate",
						APIVersion: "carto.run/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "image-template-1",
						Namespace: "some-namespace",
					},
					Spec: v1alpha1.SourceTemplateSpec{
						TemplateSpec: v1alpha1.TemplateSpec{
							Template: &runtime.RawExtension{Raw: dbytes},
						},
						URLPath: "data.does-not-exist",
					},
				}

				fakeRepo.GetDeliveryClusterTemplateReturns(templateAPI, nil)
				fakeRepo.EnsureObjectExistsOnClusterReturns(nil)
			})

			It("returns RetrieveOutputError", func() {
				_, _, err := r.Do(ctx, &resource, deliveryName, deliveryParams, outputs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("find results: does-not-exist is not found"))
				Expect(reflect.TypeOf(err).String()).To(Equal("deliverable.RetrieveOutputError"))
			})
		})

		When("unable to EnsureObjectExistsOnCluster the stamped object", func() {
			BeforeEach(func() {
				resource.Sources = []v1alpha1.ResourceReference{
					{
						Name:     "source-provider",
						Resource: "previous-resource",
					},
				}

				outputs.AddOutput("previous-resource", &templates.Output{Source: &templates.Source{
					URL:      "some-url",
					Revision: "some-revision",
				}})

				configMap := &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "example-config-map",
						Namespace: "some-namespace",
					},
					Data: map[string]string{
						"player_current_lives": `$(sources.source-provider.url)$`,
						"some_other_info":      `$(sources.source-provider.revision)$`,
					},
				}

				dbytes, err := json.Marshal(configMap)
				Expect(err).ToNot(HaveOccurred())

				templateAPI := &v1alpha1.ClusterSourceTemplate{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterSourceTemplate",
						APIVersion: "carto.run/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "source-template-1",
						Namespace: "some-namespace",
					},
					Spec: v1alpha1.SourceTemplateSpec{
						TemplateSpec: v1alpha1.TemplateSpec{
							Template: &runtime.RawExtension{Raw: dbytes},
						},
						URLPath: "data.some_other_info",
					},
				}

				fakeRepo.GetDeliveryClusterTemplateReturns(templateAPI, nil)
				fakeRepo.EnsureObjectExistsOnClusterReturns(errors.New("bad object"))
			})

			It("returns ApplyStampedObjectError", func() {
				_, _, err := r.Do(ctx, &resource, deliveryName, deliveryParams, outputs)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("bad object"))
				Expect(reflect.TypeOf(err).String()).To(Equal("deliverable.ApplyStampedObjectError"))
			})
		})
	})
})
