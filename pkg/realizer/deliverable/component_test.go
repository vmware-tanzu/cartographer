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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/cartographer/pkg/repository"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/deliverable"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

var _ = Describe("Resource", func() {

	var (
		ctx                      context.Context
		resource                 v1alpha1.DeliveryResource
		deliverable              v1alpha1.Deliverable
		outputs                  realizer.Outputs
		deliveryName             string
		fakeSystemRepo           repositoryfakes.FakeRepository
		fakeDeliverableRepo      repositoryfakes.FakeRepository
		clientForBuiltRepository client.Client
		cacheForBuiltRepository  repository.RepoCache
		repoCache                repository.RepoCache
		builtClient              client.Client
		theSecret                *corev1.Secret
		secretForBuiltClient     *corev1.Secret
		r                        realizer.ResourceRealizer
		deliveryParams           []v1alpha1.BlueprintParam
	)

	BeforeEach(func() {
		ctx = context.Background()
		resource = v1alpha1.DeliveryResource{
			Name: "resource-1",
			TemplateRef: v1alpha1.DeliveryTemplateReference{
				Kind: "ClusterSourceTemplate",
				Name: "source-template-1",
			},
		}

		deliveryName = "delivery-name"

		deliveryParams = []v1alpha1.BlueprintParam{}

		outputs = realizer.NewOutputs()

		fakeSystemRepo = repositoryfakes.FakeRepository{}
		fakeDeliverableRepo = repositoryfakes.FakeRepository{}

		repositoryBuilder := func(client client.Client, repoCache repository.RepoCache) repository.Repository {
			clientForBuiltRepository = client
			cacheForBuiltRepository = repoCache
			return &fakeDeliverableRepo
		}

		builtClient = &repositoryfakes.FakeClient{}
		clientBuilder := func(secret *corev1.Secret) (client.Client, error) {
			secretForBuiltClient = secret
			return builtClient, nil
		}

		repoCache = &repositoryfakes.FakeRepoCache{} //TODO: can we verify right cache used?
		resourceRealizerBuilder := realizer.NewResourceRealizerBuilder(repositoryBuilder, clientBuilder, repoCache)

		deliverable = v1alpha1.Deliverable{}

		theSecret = &corev1.Secret{StringData: map[string]string{"blah": "blah"}}

		var err error
		r, err = resourceRealizerBuilder(theSecret, &deliverable, &fakeSystemRepo, deliveryParams)
		Expect(err).NotTo(HaveOccurred())
	})

	It("creates a resource realizer with the existing client, as well as one with the the supplied secret mixed in", func() {
		Expect(secretForBuiltClient).To(Equal(theSecret))
		Expect(clientForBuiltRepository).To(Equal(builtClient))
	})

	It("creates a resource realizer with the existing cache", func() {
		Expect(cacheForBuiltRepository).To(Equal(repoCache))
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

				fakeSystemRepo.GetDeliveryTemplateReturns(templateAPI, nil)
				fakeDeliverableRepo.EnsureMutableObjectExistsOnClusterReturns(nil)
			})

			It("creates a stamped object and returns the outputs and stampedObjects", func() {
				returnedStampedObject, out, err := r.Do(ctx, &resource, deliveryName, outputs)
				Expect(err).ToNot(HaveOccurred())

				_, stampedObject := fakeDeliverableRepo.EnsureMutableObjectExistsOnClusterArgsForCall(0)

				Expect(returnedStampedObject).To(Equal(stampedObject))

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
				Expect(stampedObject.Object["data"]).To(Equal(map[string]interface{}{"player_current_lives": "some-url", "some_other_info": "some-revision"}))
				Expect(metadataValues["labels"]).To(Equal(map[string]interface{}{
					"carto.run/cluster-delivery-name": "delivery-name",
					"carto.run/resource-name":         "resource-1",
					"carto.run/cluster-template-name": "source-template-1",
					"carto.run/deliverable-name":      "",
					"carto.run/deliverable-namespace": "",
					"carto.run/template-kind":         "ClusterSourceTemplate",
				}))

				Expect(out.Source.Revision).To(Equal("some-revision"))
				Expect(out.Source.URL).To(Equal("some-url"))
			})
		})

		When("unable to get the template ref from systemRepo", func() {
			BeforeEach(func() {
				fakeSystemRepo.GetDeliveryTemplateReturns(nil, errors.New("bad template"))
			})

			It("returns GetDeliveryTemplateError", func() {
				_, _, err := r.Do(ctx, &resource, deliveryName, outputs)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("unable to get template [source-template-1]"))
				Expect(err.Error()).To(ContainSubstring("bad template"))
				Expect(reflect.TypeOf(err).String()).To(Equal("deliverable.GetDeliveryTemplateError"))
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

				fakeSystemRepo.GetSupplyChainTemplateReturns(templateAPI, nil)
			})

			It("returns a helpful error", func() {
				_, _, err := r.Do(ctx, &resource, deliveryName, outputs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to get delivery cluster template [{Kind:ClusterSourceTemplate Name:source-template-1}]: resource does not match a known template"))
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

				fakeSystemRepo.GetDeliveryTemplateReturns(templateAPI, nil)
			})

			It("returns StampError", func() {
				_, _, err := r.Do(ctx, &resource, deliveryName, outputs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unable to stamp object for resource [resource-1]"))
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

				fakeSystemRepo.GetDeliveryTemplateReturns(templateAPI, nil)
				fakeDeliverableRepo.EnsureMutableObjectExistsOnClusterReturns(nil)
			})

			It("returns RetrieveOutputError", func() {
				_, _, err := r.Do(ctx, &resource, deliveryName, outputs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("find results: does-not-exist is not found"))
				Expect(reflect.TypeOf(err).String()).To(Equal("deliverable.RetrieveOutputError"))
			})
		})

		When("unable to EnsureImmutableObjectExistsOnCluster the stamped object", func() {
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

				fakeSystemRepo.GetDeliveryTemplateReturns(templateAPI, nil)
				fakeDeliverableRepo.EnsureMutableObjectExistsOnClusterReturns(errors.New("bad object"))
			})

			It("returns ApplyStampedObjectError", func() {
				_, _, err := r.Do(ctx, &resource, deliveryName, outputs)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("bad object"))
				Expect(reflect.TypeOf(err).String()).To(Equal("deliverable.ApplyStampedObjectError"))
			})
		})
	})
})
