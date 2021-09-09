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

package workload_test

import (
	"encoding/json"
	"errors"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/eval"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/workload"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

var _ = Describe("Component", func() {

	var (
		component       v1alpha1.SupplyChainComponent
		workload        v1alpha1.Workload
		outputs         realizer.Outputs
		supplyChainName string
		fakeRepo        repositoryfakes.FakeRepository
		r               realizer.ComponentRealizer
	)

	BeforeEach(func() {
		component = v1alpha1.SupplyChainComponent{
			Name: "component-1",
			TemplateRef: v1alpha1.ClusterTemplateReference{
				Kind: "ClusterImageTemplate",
				Name: "image-template-1",
			},
		}

		supplyChainName = "supply-chain-name"

		outputs = realizer.NewOutputs()

		fakeRepo = repositoryfakes.FakeRepository{}
		workload = v1alpha1.Workload{}
		r = realizer.NewComponentRealizer(&workload, &fakeRepo)
	})

	Describe("Do", func() {
		When("passed a workload with outputs", func() {
			BeforeEach(func() {
				component.Sources = []v1alpha1.ComponentReference{
					{
						Name:      "source-provider",
						Component: "previous-component",
					},
				}

				outputs.AddOutput("previous-component", &templates.Output{Source: &templates.Source{
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
						"player_current_lives": "$(sources[0].url)$",
						"some_other_info":      "$(sources[?(@.name==\"source-provider\")].revision)$",
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
						Template:  runtime.RawExtension{Raw: dbytes},
						ImagePath: "data.some_other_info",
					},
				}

				template := templates.NewClusterImageTemplateModel(templateAPI, eval.EvaluatorBuilder())
				fakeRepo.GetClusterTemplateReturns(template, nil)
				fakeRepo.AssureObjectExistsOnClusterReturns(nil)
			})

			It("creates a stamped object and returns the outputs", func() {
				out, err := r.Do(&component, supplyChainName, outputs)
				Expect(err).ToNot(HaveOccurred())

				stampedObject, allowUpdate := fakeRepo.AssureObjectExistsOnClusterArgsForCall(0)
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
					"carto.run/cluster-supply-chain-name": "supply-chain-name",
					"carto.run/component-name":            "component-1",
					"carto.run/cluster-template-name":     "image-template-1",
					"carto.run/workload-name":             "",
					"carto.run/workload-namespace":        "",
				}))
				Expect(stampedObject.Object["data"]).To(Equal(map[string]interface{}{"player_current_lives": "some-url", "some_other_info": "some-revision"}))

				Expect(out.Image).To(Equal("some-revision"))
			})
		})

		When("unable to get the template ref from repo", func() {
			BeforeEach(func() {
				fakeRepo.GetClusterTemplateReturns(nil, errors.New("bad template"))
			})

			It("returns GetClusterTemplateError", func() {
				_, err := r.Do(&component, supplyChainName, outputs)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("unable to get template 'image-template-1'"))
				Expect(err.Error()).To(ContainSubstring("bad template"))
				Expect(reflect.TypeOf(err).String()).To(Equal("workload.GetClusterTemplateError"))
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
						Template: runtime.RawExtension{},
					},
				}

				template := templates.NewClusterImageTemplateModel(templateAPI, eval.EvaluatorBuilder())
				fakeRepo.GetClusterTemplateReturns(template, nil)
			})

			It("returns StampError", func() {
				_, err := r.Do(&component, supplyChainName, outputs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unable to stamp object for component 'component-1'"))
				Expect(reflect.TypeOf(err).String()).To(Equal("workload.StampError"))
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
						Template:  runtime.RawExtension{Raw: dbytes},
						ImagePath: "data.does-not-exist",
					},
				}

				template := templates.NewClusterImageTemplateModel(templateAPI, eval.EvaluatorBuilder())
				fakeRepo.GetClusterTemplateReturns(template, nil)
				fakeRepo.AssureObjectExistsOnClusterReturns(nil)
			})

			It("returns RetrieveOutputError", func() {
				_, err := r.Do(&component, supplyChainName, outputs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("find results: does-not-exist is not found"))
				Expect(reflect.TypeOf(err).String()).To(Equal("workload.RetrieveOutputError"))
			})
		})

		When("unable to AssureObjectExistsOnCluster the stamped object", func() {
			BeforeEach(func() {
				component.Sources = []v1alpha1.ComponentReference{
					{
						Name:      "source-provider",
						Component: "previous-component",
					},
				}

				outputs.AddOutput("previous-component", &templates.Output{Source: &templates.Source{
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
						"player_current_lives": "$(sources[0].url)$",
						"some_other_info":      "$(sources[?(@.name==\"source-provider\")].revision)$",
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
						Template:  runtime.RawExtension{Raw: dbytes},
						ImagePath: "data.some_other_info",
					},
				}

				template := templates.NewClusterImageTemplateModel(templateAPI, eval.EvaluatorBuilder())
				fakeRepo.GetClusterTemplateReturns(template, nil)
				fakeRepo.AssureObjectExistsOnClusterReturns(errors.New("bad object"))
			})
			It("returns ApplyStampedObjectError", func() {
				_, err := r.Do(&component, supplyChainName, outputs)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("bad object"))
				Expect(reflect.TypeOf(err).String()).To(Equal("workload.ApplyStampedObjectError"))
			})
		})
	})
})
