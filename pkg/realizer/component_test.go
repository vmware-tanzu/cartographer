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
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/realizer"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/realizerfakes"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

var _ = Describe("Resource", func() {

	var (
		ctx                      context.Context
		resource                 realizer.OwnerResource
		workload                 v1alpha1.Workload
		outputs                  realizer.Outputs
		blueprintName            string
		fakeSystemRepo           repositoryfakes.FakeRepository
		fakeOwnerRepo            repositoryfakes.FakeRepository
		clientForBuiltRepository client.Client
		cacheForBuiltRepository  repository.RepoCache
		theAuthToken             string
		authTokenForBuiltClient  string
		r                        realizer.ResourceRealizer
		out                      *Buffer
		repoCache                repository.RepoCache
		supplyChainParams        []v1alpha1.BlueprintParam
		fakeMapper               *realizerfakes.FakeRESTMapper
	)

	BeforeEach(func() {
		var err error

		ctx = context.Background()
		resource = realizer.OwnerResource{
			Name: "resource-1",
			TemplateRef: v1alpha1.TemplateReference{
				Kind: "ClusterImageTemplate",
				Name: "image-template-1",
			},
		}

		blueprintName = "supply-chain-name"
		supplyChainParams = []v1alpha1.BlueprintParam{}

		outputs = realizer.NewOutputs()

		fakeSystemRepo = repositoryfakes.FakeRepository{}
		fakeOwnerRepo = repositoryfakes.FakeRepository{}
		workload = v1alpha1.Workload{}

		repositoryBuilder := func(client client.Client, repoCache repository.RepoCache) repository.Repository {
			clientForBuiltRepository = client
			cacheForBuiltRepository = repoCache
			return &fakeOwnerRepo
		}

		builtClient := &repositoryfakes.FakeClient{}
		clientBuilder := func(authToken string, _ bool) (client.Client, discovery.DiscoveryInterface, error) {
			authTokenForBuiltClient = authToken
			return builtClient, nil, nil
		}
		out = NewBuffer()
		logger := zap.New(zap.WriteTo(out))

		repoCache = repository.NewCache(logger)
		resourceRealizerBuilder := realizer.NewResourceRealizerBuilder(repositoryBuilder, clientBuilder, repoCache)

		theAuthToken = "tis-but-a-flesh-wound"

		placeholderLabeler := func(resource realizer.OwnerResource, reader templates.Reader) templates.Labels {
			return templates.Labels{"expected-labels-from-labeler-placeholder": "labeler"}
		}
		r, err = resourceRealizerBuilder(theAuthToken, &workload, realizer.NewContextGenerator(&workload, []v1alpha1.OwnerParam{}, supplyChainParams), &fakeSystemRepo, placeholderLabeler)

		Expect(err).NotTo(HaveOccurred())

		fakeMapper = &realizerfakes.FakeRESTMapper{}
	})

	It("creates a resource realizer with the existing client, as well as one with the the supplied secret mixed in", func() {
		Expect(authTokenForBuiltClient).To(Equal(theAuthToken))
		Expect(clientForBuiltRepository).To(Equal(clientForBuiltRepository))
	})

	It("creates a resource realizer with the existing cache", func() {
		Expect(cacheForBuiltRepository).To(Equal(repoCache))
	})

	Describe("Do", func() {
		When("passed outputs that are populated", func() {
			var templateAPI *v1alpha1.ClusterSourceTemplate
			var expectedObject unstructured.Unstructured
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
						Name: "example-config-map",
					},
					Data: map[string]string{
						"player_current_lives": `$(source.url)$`,
						"some_other_info":      `$(sources.source-provider.revision)$`,
					},
				}

				dbytes, err := json.Marshal(configMap)
				Expect(err).ToNot(HaveOccurred())

				templateAPI = &v1alpha1.ClusterSourceTemplate{
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

				expectedObject = unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "example-config-map",
							"ownerReferences": []interface{}{
								map[string]interface{}{
									"apiVersion":         "",
									"kind":               "",
									"name":               "",
									"uid":                "",
									"controller":         true,
									"blockOwnerDeletion": true,
								},
							},
							"creationTimestamp": nil,
							"labels": map[string]interface{}{
								"expected-labels-from-labeler-placeholder": "labeler",
							},
						},
						"data": map[string]interface{}{
							"player_current_lives": "some-url",
							"some_other_info":      "some-revision",
						},
					},
				}
			})

			When("template is mutable", func() {
				BeforeEach(func() {
					fakeSystemRepo.GetTemplateReturns(templateAPI, nil)
					fakeOwnerRepo.EnsureMutableObjectExistsOnClusterReturns(nil)
				})

				It("creates a stamped object and returns the outputs and stampedObjects", func() {
					template, returnedStampedObject, out, isPassThrough, templateRefName, err := r.Do(ctx, resource, blueprintName, outputs, fakeMapper)
					Expect(err).ToNot(HaveOccurred())
					Expect(template).ToNot(BeNil())
					Expect(isPassThrough).To(BeFalse())
					Expect(templateRefName).To(Equal("image-template-1"))
					Expect(returnedStampedObject.Object).To(Equal(expectedObject.Object))

					Expect(fakeOwnerRepo.EnsureMutableObjectExistsOnClusterCallCount()).To(Equal(1))

					_, stampedObject := fakeOwnerRepo.EnsureMutableObjectExistsOnClusterArgsForCall(0)

					Expect(returnedStampedObject).To(Equal(stampedObject))

					metadata := stampedObject.Object["metadata"]
					metadataValues, ok := metadata.(map[string]interface{})
					Expect(ok).To(BeTrue())
					Expect(metadataValues["name"]).To(Equal("example-config-map"))
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
					Expect(metadataValues["labels"]).To(Equal(map[string]interface{}{"expected-labels-from-labeler-placeholder": "labeler"}))

					Expect(out.Source.Revision).To(Equal("some-revision"))
					Expect(out.Source.URL).To(Equal("some-url"))
				})
			})

			When("template is immutable", func() {
				BeforeEach(func() {
					templateAPI.Spec.TemplateSpec.Lifecycle = "immutable"
					templateAPI.Spec.TemplateSpec.RetentionPolicy = &v1alpha1.RetentionPolicy{
						MaxFailedRuns:     10,
						MaxSuccessfulRuns: 10,
					}

					fakeSystemRepo.GetTemplateReturns(templateAPI, nil)
				})

				When("call to ensure immutable object succeeds", func() {
					BeforeEach(func() {
						fakeOwnerRepo.EnsureImmutableObjectExistsOnClusterReturns(nil)
					})

					When("call to list objects succeeds", func() {
						BeforeEach(func() {
							stampedObjectWithTime := expectedObject.DeepCopy()

							stampedObjectWithTime.SetCreationTimestamp(metav1.NewTime(time.Unix(1, 0)))

							fakeOwnerRepo.ListUnstructuredReturns([]*unstructured.Unstructured{stampedObjectWithTime}, nil)
						})

						When("no returned object meets the healthRule", func() {
							BeforeEach(func() {
								templateAPI.Spec.TemplateSpec.HealthRule = &v1alpha1.HealthRule{
									SingleConditionType: "Ready",
								}
							})
							It("creates a stamped object, but returns an error and no output", func() {
								template, returnedStampedObject, out, isPassThrough, templateRefName, err := r.Do(ctx, resource, blueprintName, outputs, fakeMapper)
								Expect(template).ToNot(BeNil())
								Expect(isPassThrough).To(BeFalse())
								Expect(templateRefName).To(Equal("image-template-1"))
								Expect(returnedStampedObject.Object).To(Equal(expectedObject.Object))
								Expect(out).To(BeNil())

								Expect(fakeOwnerRepo.EnsureImmutableObjectExistsOnClusterCallCount()).To(Equal(1))

								_, stampedObject, _ := fakeOwnerRepo.EnsureImmutableObjectExistsOnClusterArgsForCall(0)

								Expect(returnedStampedObject).To(Equal(stampedObject))

								metadata := stampedObject.Object["metadata"]
								metadataValues, ok := metadata.(map[string]interface{})
								Expect(ok).To(BeTrue())
								Expect(metadataValues["name"]).To(Equal("example-config-map"))
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
								Expect(metadataValues["labels"]).To(Equal(map[string]interface{}{"expected-labels-from-labeler-placeholder": "labeler"}))

								Expect(err).To(HaveOccurred())
								Expect(err.Error()).To(ContainSubstring("unable to retrieve outputs for resource [resource-1] in supply chain [supply-chain-name]: failed to find any healthy object in the set of immutable stamped objects"))
								Expect(reflect.TypeOf(err).String()).To(Equal("errors.NoHealthyImmutableObjectsError"))
							})
						})

						When("at least one returned object meets the healthRule", func() {
							BeforeEach(func() {
								templateAPI.Spec.TemplateSpec.HealthRule = &v1alpha1.HealthRule{
									AlwaysHealthy: &runtime.RawExtension{Raw: []byte{}},
								}
							})
							It("creates a stamped object and returns the outputs and stampedObjects", func() {
								template, returnedStampedObject, out, isPassThrough, templateRefName, err := r.Do(ctx, resource, blueprintName, outputs, fakeMapper)
								Expect(err).ToNot(HaveOccurred())
								Expect(template).ToNot(BeNil())
								Expect(isPassThrough).To(BeFalse())
								Expect(templateRefName).To(Equal("image-template-1"))
								Expect(returnedStampedObject.Object).To(Equal(expectedObject.Object))

								Expect(fakeOwnerRepo.EnsureImmutableObjectExistsOnClusterCallCount()).To(Equal(1))

								_, stampedObject, _ := fakeOwnerRepo.EnsureImmutableObjectExistsOnClusterArgsForCall(0)

								Expect(returnedStampedObject).To(Equal(stampedObject))

								metadata := stampedObject.Object["metadata"]
								metadataValues, ok := metadata.(map[string]interface{})
								Expect(ok).To(BeTrue())
								Expect(metadataValues["name"]).To(Equal("example-config-map"))
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
								Expect(metadataValues["labels"]).To(Equal(map[string]interface{}{"expected-labels-from-labeler-placeholder": "labeler"}))

								Expect(out.Source.Revision).To(Equal("some-revision"))
								Expect(out.Source.URL).To(Equal("some-url"))
							})
						})
					})

					When("the call to list objects fails", func() {
						BeforeEach(func() {
							fakeOwnerRepo.ListUnstructuredReturns([]*unstructured.Unstructured{}, fmt.Errorf("some error"))
						})

						It("returns ListCreatedObjectsError", func() {
							template, _, _, isPassThrough, templateRefName, err := r.Do(ctx, resource, blueprintName, outputs, fakeMapper)
							Expect(template).ToNot(BeNil())
							Expect(isPassThrough).To(BeFalse())
							Expect(templateRefName).To(Equal("image-template-1"))

							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("some error"))
							Expect(reflect.TypeOf(err).String()).To(Equal("errors.ListCreatedObjectsError"))
						})
					})
				})
				When("call to ensure immutable object fails", func() {
					BeforeEach(func() {
						fakeOwnerRepo.EnsureImmutableObjectExistsOnClusterReturns(fmt.Errorf("bad object"))
					})

					It("returns ApplyStampedObjectError", func() {
						template, _, _, isPassThrough, templateRefName, err := r.Do(ctx, resource, blueprintName, outputs, fakeMapper)
						Expect(template).ToNot(BeNil())
						Expect(isPassThrough).To(BeFalse())
						Expect(templateRefName).To(Equal("image-template-1"))

						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("bad object"))
						Expect(reflect.TypeOf(err).String()).To(Equal("errors.ApplyStampedObjectError"))
					})
				})
			})
		})

		When("unable to get the template ref from repo", func() {
			BeforeEach(func() {
				fakeSystemRepo.GetTemplateReturns(nil, errors.New("bad template"))
			})

			It("returns GetTemplateError", func() {
				template, _, _, isPassThrough, _, err := r.Do(ctx, resource, blueprintName, outputs, fakeMapper)
				Expect(err).To(HaveOccurred())
				Expect(template).To(BeNil())
				Expect(isPassThrough).To(BeFalse())

				Expect(err.Error()).To(ContainSubstring("unable to get template [image-template-1]"))
				Expect(err.Error()).To(ContainSubstring("bad template"))
				Expect(reflect.TypeOf(err).String()).To(Equal("errors.GetTemplateError"))
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

				fakeSystemRepo.GetTemplateReturns(templateAPI, nil)
			})

			It("returns a helpful error", func() {
				template, _, _, isPassThrough, templateRefName, err := r.Do(ctx, resource, blueprintName, outputs, fakeMapper)

				Expect(template).To(BeNil())
				Expect(isPassThrough).To(BeFalse())
				Expect(templateRefName).To(Equal("image-template-1"))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to get cluster template [{Kind:ClusterImageTemplate Name:image-template-1}]: resource does not match a known template"))
			})
		})

		When("unable to Stamp a new template", func() {
			BeforeEach(func() {
				templateAPI := &v1alpha1.ClusterImageTemplate{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterImageTemplate",
						APIVersion: "carto.run/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "image-template-1",
						Namespace: "some-namespace",
					},
					Spec: v1alpha1.ImageTemplateSpec{
						TemplateSpec: v1alpha1.TemplateSpec{
							Template: &runtime.RawExtension{},
						},
					},
				}

				fakeSystemRepo.GetTemplateReturns(templateAPI, nil)
			})

			It("returns StampError", func() {
				template, _, _, isPassThrough, templateRefName, err := r.Do(ctx, resource, blueprintName, outputs, fakeMapper)
				Expect(template).ToNot(BeNil())
				Expect(isPassThrough).To(BeFalse())
				Expect(templateRefName).To(Equal("image-template-1"))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unable to stamp object for resource [resource-1]"))
				Expect(reflect.TypeOf(err).String()).To(Equal("errors.StampError"))
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
						Name: "example-config-map",
					},
					Data: map[string]string{
						"player_current_lives": "9",
						"some_other_info":      "10",
					},
				}

				dbytes, err := json.Marshal(configMap)
				Expect(err).ToNot(HaveOccurred())

				templateAPI := &v1alpha1.ClusterImageTemplate{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterImageTemplate",
						APIVersion: "carto.run/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "image-template-1",
						Namespace: "some-namespace",
					},
					Spec: v1alpha1.ImageTemplateSpec{
						TemplateSpec: v1alpha1.TemplateSpec{
							Template: &runtime.RawExtension{Raw: dbytes},
						},
						ImagePath: "data.does-not-exist",
					},
				}

				fakeSystemRepo.GetTemplateReturns(templateAPI, nil)
				fakeOwnerRepo.EnsureMutableObjectExistsOnClusterReturns(nil)

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

			It("returns RetrieveOutputError", func() {
				template, _, _, isPassThrough, templateRefName, err := r.Do(ctx, resource, blueprintName, outputs, fakeMapper)
				Expect(template).ToNot(BeNil())
				Expect(isPassThrough).To(BeFalse())
				Expect(templateRefName).To(Equal("image-template-1"))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("jsonpath returned empty list: data.does-not-exist"))
				Expect(err.Error()).To(ContainSubstring("of type [FOO.EXAMPLE.COM]"))
				Expect(reflect.TypeOf(err).String()).To(Equal("errors.RetrieveOutputError"))
			})
		})

		When("unable to EnsureMutableObjectExistsOnCluster the stamped object", func() {
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
						Name: "example-config-map",
					},
					Data: map[string]string{
						"player_current_lives": `$(sources.source-provider.url)$`,
						"some_other_info":      `$(sources.source-provider.revision)$`,
					},
				}

				dbytes, err := json.Marshal(configMap)
				Expect(err).ToNot(HaveOccurred())

				templateAPI := &v1alpha1.ClusterImageTemplate{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterImageTemplate",
						APIVersion: "carto.run/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "image-template-1",
						Namespace: "some-namespace",
					},
					Spec: v1alpha1.ImageTemplateSpec{
						TemplateSpec: v1alpha1.TemplateSpec{
							Template: &runtime.RawExtension{Raw: dbytes},
						},
						ImagePath: "data.some_other_info",
					},
				}

				fakeSystemRepo.GetTemplateReturns(templateAPI, nil)
				fakeOwnerRepo.EnsureMutableObjectExistsOnClusterReturns(errors.New("bad object"))
			})
			It("returns ApplyStampedObjectError", func() {
				template, _, _, isPassThrough, templateRefName, err := r.Do(ctx, resource, blueprintName, outputs, fakeMapper)
				Expect(template).ToNot(BeNil())
				Expect(isPassThrough).To(BeFalse())
				Expect(templateRefName).To(Equal("image-template-1"))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("bad object"))
				Expect(reflect.TypeOf(err).String()).To(Equal("errors.ApplyStampedObjectError"))
			})
		})

		When("resource template has namespace specified", func() {
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

				templateAPI := &v1alpha1.ClusterImageTemplate{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterImageTemplate",
						APIVersion: "carto.run/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "image-template-1",
						Namespace: "some-namespace",
					},
					Spec: v1alpha1.ImageTemplateSpec{
						TemplateSpec: v1alpha1.TemplateSpec{
							Template: &runtime.RawExtension{Raw: dbytes},
						},
						ImagePath: "data.does-not-exist",
					},
				}

				fakeSystemRepo.GetTemplateReturns(templateAPI, nil)
			})

			It("returns StampError", func() {
				template, _, _, isPassThrough, templateRefName, err := r.Do(ctx, resource, blueprintName, outputs, fakeMapper)
				Expect(template).ToNot(BeNil())
				Expect(isPassThrough).To(BeFalse())
				Expect(templateRefName).To(Equal("image-template-1"))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cannot set namespace in resource template"))
				Expect(reflect.TypeOf(err).String()).To(Equal("errors.StampError"))
			})
		})

		When("template ref has options", func() {
			BeforeEach(func() {
				url := "https://example.com"
				branch := "main"
				workload = v1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name: "my-workload",
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       "Workload",
						APIVersion: "v1alpha",
					},
					Spec: v1alpha1.WorkloadSpec{
						Source: &v1alpha1.Source{
							Git: &v1alpha1.GitSource{
								URL: &url,
								Ref: &v1alpha1.GitRef{
									Branch: &branch,
								},
							},
						},
						Env: []corev1.EnvVar{
							{
								Name:  "some-name",
								Value: "some-value",
							},
						},
					},
				}

				resource = realizer.OwnerResource{
					Name: "resource-1",
					TemplateRef: v1alpha1.TemplateReference{
						Kind: "ClusterImageTemplate",
					},
				}
			})

			When("the option that matches is pass through", func() {
				BeforeEach(func() {
					resource.TemplateOptions = []v1alpha1.TemplateOption{
						{
							Name: "template-not-chosen",
							Selector: v1alpha1.Selector{
								MatchFields: []v1alpha1.FieldSelectorRequirement{
									{
										Key:      "spec.source.image",
										Operator: "Exists",
									},
								},
							},
						},
						{
							PassThrough: "my-input",
							Selector: v1alpha1.Selector{
								MatchFields: []v1alpha1.FieldSelectorRequirement{
									{
										Key:      "spec.source.git.url",
										Operator: "Exists",
									},
								},
							},
						},
					}

					resource.Images = []v1alpha1.ResourceReference{
						{
							Name:     "my-input",
							Resource: "my-input",
						},
					}

					outputs.AddOutput("my-input", &templates.Output{Image: "my-image"})
				})

				It("returns the input as an output", func() {
					template, stamped, output, isPassThrough, _, err := r.Do(ctx, resource, blueprintName, outputs, fakeMapper)
					Expect(template).To(BeNil())
					Expect(stamped).To(BeNil())
					Expect(isPassThrough).To(BeTrue())
					Expect(err).NotTo(HaveOccurred())

					Expect(output.Image).To(Equal("my-image"))
				})

				It("does not call to the repo", func() {
					_, _, _, _, _, _ = r.Do(ctx, resource, blueprintName, outputs, fakeMapper)
					Expect(fakeSystemRepo.GetTemplateCallCount()).To(Equal(0))
					Expect(fakeOwnerRepo.EnsureMutableObjectExistsOnClusterCallCount()).To(Equal(0))
				})

				When("output cannot be retrieved", func() {
					BeforeEach(func() {
						resource.Images = []v1alpha1.ResourceReference{}
					})

					It("returns an error", func() {
						template, stamped, output, isPassThrough, _, err := r.Do(ctx, resource, blueprintName, outputs, fakeMapper)
						Expect(template).To(BeNil())
						Expect(stamped).To(BeNil())
						Expect(output).To(BeNil())
						Expect(isPassThrough).To(BeTrue())

						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("unable to retrieve outputs from pass through [my-input] for resource [resource-1] in supply chain [supply-chain-name]: input [my-input] not found in images"))
					})
				})

				When("pass through reader cannot be created", func() {
					BeforeEach(func() {
						resource.TemplateRef.Kind = "something-bad"
					})

					It("returns an error", func() {
						template, stamped, output, isPassThrough, _, err := r.Do(ctx, resource, blueprintName, outputs, fakeMapper)
						Expect(template).To(BeNil())
						Expect(stamped).To(BeNil())
						Expect(output).To(BeNil())
						Expect(isPassThrough).To(BeTrue())

						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("failed to create new stamp pass through reader: kind does not match a known template"))
					})
				})
			})

			When("the option that matches is a template", func() {
				BeforeEach(func() {
					resource.TemplateOptions = []v1alpha1.TemplateOption{
						{
							Name: "template-not-chosen",
							Selector: v1alpha1.Selector{
								MatchFields: []v1alpha1.FieldSelectorRequirement{
									{
										Key:      "spec.source.image",
										Operator: "Exists",
									},
								},
							},
						},
						{
							Name: "template-chosen",
							Selector: v1alpha1.Selector{
								MatchFields: []v1alpha1.FieldSelectorRequirement{
									{
										Key:      "spec.source.git.url",
										Operator: "Exists",
									},
								},
							},
						},
					}

					configMap := &corev1.ConfigMap{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ConfigMap",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "example-config-map",
						},
						Data: map[string]string{
							"some_other_info": "hello",
						},
					}

					dbytes, err := json.Marshal(configMap)
					Expect(err).ToNot(HaveOccurred())

					templateAPI := &v1alpha1.ClusterImageTemplate{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ClusterImageTemplate",
							APIVersion: "carto.run/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "template-chosen",
						},
						Spec: v1alpha1.ImageTemplateSpec{
							TemplateSpec: v1alpha1.TemplateSpec{
								Template: &runtime.RawExtension{Raw: dbytes},
							},
							ImagePath: "data.some_other_info",
						},
					}

					fakeSystemRepo.GetTemplateReturns(templateAPI, nil)
					fakeOwnerRepo.EnsureMutableObjectExistsOnClusterReturns(nil)
				})

				When("one option matches", func() {
					It("finds the correct template", func() {
						template, _, _, isPassThrough, templateRefName, err := r.Do(ctx, resource, blueprintName, outputs, fakeMapper)
						Expect(template).ToNot(BeNil())
						Expect(isPassThrough).To(BeFalse())
						Expect(err).NotTo(HaveOccurred())
						Expect(templateRefName).To(Equal("template-chosen"))

						_, name, kind := fakeSystemRepo.GetTemplateArgsForCall(0)
						Expect(name).To(Equal("template-chosen"))
						Expect(kind).To(Equal("ClusterImageTemplate"))
					})
				})

				When("more than one option matches", func() {
					It("returns a TemplateOptionsMatchError", func() {
						resource.TemplateOptions[0].Selector.MatchFields[0].Key = "spec.source.git.ref.branch"

						template, _, _, isPassThrough, _, err := r.Do(ctx, resource, blueprintName, outputs, fakeMapper)
						Expect(template).To(BeNil())
						Expect(isPassThrough).To(BeFalse())

						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("expected exactly 1 option to match, found [2] matching options [template-not-chosen, template-chosen] for resource [resource-1] in supply chain [supply-chain-name]"))
					})

				})

				When("zero options match", func() {
					It("returns a TemplateOptionsMatchError", func() {
						resource.TemplateOptions[0].Selector.MatchFields[0].Key = "spec.source.image"
						resource.TemplateOptions[1].Selector.MatchFields[0].Key = "spec.source.subPath"

						template, _, _, isPassThrough, _, err := r.Do(ctx, resource, blueprintName, outputs, fakeMapper)

						Expect(template).To(BeNil())
						Expect(isPassThrough).To(BeFalse())

						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("expected exactly 1 option to match, found [0] matching options for resource [resource-1] in supply chain [supply-chain-name]"))
					})
				})

				When("one option has key that does not exist in the spec", func() {
					It("does not error", func() {
						resource.TemplateOptions[0].Selector.MatchFields[0].Key = `spec.env[?(@.name=="some-name")].bad`

						template, _, _, isPassThrough, templateRefName, err := r.Do(ctx, resource, blueprintName, outputs, fakeMapper)
						Expect(template).ToNot(BeNil())
						Expect(isPassThrough).To(BeFalse())
						Expect(templateRefName).To(Equal("template-chosen"))

						Expect(err).NotTo(HaveOccurred())
						_, name, kind := fakeSystemRepo.GetTemplateArgsForCall(0)
						Expect(name).To(Equal("template-chosen"))
						Expect(kind).To(Equal("ClusterImageTemplate"))
					})
				})

				When("key is malformed", func() {
					It("returns a ResolveTemplateOptionError", func() {
						resource.TemplateOptions[0].Selector.MatchFields[0].Key = `spec.env[`

						template, _, _, isPassThrough, _, err := r.Do(ctx, resource, blueprintName, outputs, fakeMapper)
						Expect(template).To(BeNil())
						Expect(isPassThrough).To(BeFalse())

						Expect(err).To(HaveOccurred())
						Expect(reflect.TypeOf(err).String()).To(Equal("errors.ResolveTemplateOptionError"))
						Expect(err.Error()).To(ContainSubstring(`error matching against template option [template-not-chosen] for resource [resource-1] in supply chain [supply-chain-name]`))
						Expect(err.Error()).To(ContainSubstring(`failed to evaluate selector matchFields: unable to match field requirement with key [spec.env[] operator [Exists] values [[]]: evaluate: failed to parse jsonpath '{.spec.env[}': unterminated array`))
					})
				})

				When("one option matches with multiple fields", func() {
					It("finds the correct template", func() {
						resource.TemplateOptions[0].Selector.MatchFields = append(resource.TemplateOptions[0].Selector.MatchFields, v1alpha1.FieldSelectorRequirement{
							Key:      "spec.source.git.ref.branch",
							Operator: "Exists",
						})

						template, _, _, isPassThrough, templateRefName, err := r.Do(ctx, resource, blueprintName, outputs, fakeMapper)
						Expect(template).ToNot(BeNil())
						Expect(isPassThrough).To(BeFalse())
						Expect(templateRefName).To(Equal("template-chosen"))

						Expect(err).NotTo(HaveOccurred())
						_, name, kind := fakeSystemRepo.GetTemplateArgsForCall(0)
						Expect(name).To(Equal("template-chosen"))
						Expect(kind).To(Equal("ClusterImageTemplate"))
					})
				})
			})
		})
	})
})
