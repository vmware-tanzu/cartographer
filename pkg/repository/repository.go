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

package repository

import (
	"context"
	"fmt"

	api_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate sigs.k8s.io/controller-runtime/pkg/client.Client

//counterfeiter:generate . Repository
type Repository interface {
	EnsureObjectExistsOnCluster(obj *unstructured.Unstructured, allowUpdate bool) error
	GetClusterTemplate(reference v1alpha1.ClusterTemplateReference) (templates.Template, error)
	GetTemplate(reference v1alpha1.TemplateReference) (templates.Template, error)
	GetSupplyChainsForWorkload(workload *v1alpha1.Workload) ([]v1alpha1.ClusterSupplyChain, error)
	GetWorkload(name string, namespace string) (*v1alpha1.Workload, error)
	GetSupplyChain(name string) (*v1alpha1.ClusterSupplyChain, error)
	StatusUpdate(object client.Object) error
	GetScheme() *runtime.Scheme
	GetPipeline(name string, namespace string) (*v1alpha1.Pipeline, error)
}

type repository struct {
	rc RepoCache
	cl client.Client
}

func NewRepository(client client.Client, repoCache RepoCache) Repository {
	return &repository{
		rc: repoCache,
		cl: client,
	}
}

func (r *repository) EnsureObjectExistsOnCluster(obj *unstructured.Unstructured, allowUpdate bool) error {
	unstructuredList, err := r.listUnstructured(obj)
	if err != nil {
		return err
	}

	cacheHit := r.rc.UnchangedSinceCached(obj, unstructuredList)
	if cacheHit != nil {
		r.rc.Refresh(obj.DeepCopy())
		*obj = *cacheHit
		return nil
	}

	var outdatedObject *unstructured.Unstructured
	if allowUpdate {
		outdatedObject = getOutdatedUnstructuredByName(obj, unstructuredList)
	}

	if outdatedObject != nil {
		return r.patchUnstructured(outdatedObject, obj)
	} else {
		return r.createUnstructured(obj)
	}
}

func getOutdatedUnstructuredByName(target *unstructured.Unstructured, candidates []unstructured.Unstructured) *unstructured.Unstructured {
	for _, candidate := range candidates {
		if candidate.GetName() == target.GetName() && candidate.GetNamespace() == target.GetNamespace() {
			return &candidate
		}
	}
	return nil
}

func (r *repository) listUnstructured(obj *unstructured.Unstructured) ([]unstructured.Unstructured, error) {
	unstructuredList := &unstructured.UnstructuredList{}
	unstructuredList.SetGroupVersionKind(obj.GroupVersionKind())

	opts := []client.ListOption{
		client.InNamespace(obj.GetNamespace()),
		client.MatchingLabels(obj.GetLabels()),
	}
	err := r.cl.List(context.TODO(), unstructuredList, opts...)
	if err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}

	return unstructuredList.Items, nil
}

func (r *repository) GetClusterTemplate(ref v1alpha1.ClusterTemplateReference) (templates.Template, error) {
	apiTemplate, err := v1alpha1.GetAPITemplate(ref.Kind)
	if err != nil {
		return nil, fmt.Errorf("get api template: %w", err)
	}

	err = r.cl.Get(context.TODO(), client.ObjectKey{
		Name: ref.Name,
	}, apiTemplate)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	template, err := templates.NewModelFromAPI(apiTemplate)
	if err != nil {
		return nil, fmt.Errorf("new model from api: %w", err)
	}

	return template, nil
}

// FIXME: Probably should not return a templates.Template, not sure it adheres to interface
func (r *repository) GetTemplate(ref v1alpha1.TemplateReference) (templates.Template, error) {
	apiTemplate, err := v1alpha1.GetAPITemplate(ref.Kind)
	if err != nil {
		return nil, fmt.Errorf("get api template: %w", err)
	}

	err = r.cl.Get(context.TODO(), client.ObjectKey{
		Name:      ref.Name,
		Namespace: ref.Namespace,
	}, apiTemplate)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	template, err := templates.NewModelFromAPI(apiTemplate)
	if err != nil {
		return nil, fmt.Errorf("new model from api: %w", err)
	}

	return template, nil
}

func (r *repository) createUnstructured(obj *unstructured.Unstructured) error {
	submitted := obj.DeepCopy()
	if err := r.cl.Create(context.TODO(), obj); err != nil {
		return fmt.Errorf("create: %w", err)
	}

	r.rc.Set(submitted, obj.DeepCopy())
	return nil
}

func (r *repository) patchUnstructured(existingObj *unstructured.Unstructured, obj *unstructured.Unstructured) error {
	submitted := obj.DeepCopy()
	obj.SetResourceVersion(existingObj.GetResourceVersion())
	if err := r.cl.Patch(context.TODO(), obj, client.MergeFrom(existingObj)); err != nil {
		return fmt.Errorf("patch: %w", err)
	}

	r.rc.Set(submitted, obj.DeepCopy())
	return nil
}

func (r *repository) GetSupplyChainsForWorkload(workload *v1alpha1.Workload) ([]v1alpha1.ClusterSupplyChain, error) {
	list := &v1alpha1.ClusterSupplyChainList{}
	if err := r.cl.List(context.TODO(), list); err != nil {
		return nil, fmt.Errorf("list supply chains: %w", err)
	}

	var clusterSupplyChains []v1alpha1.ClusterSupplyChain
	for _, supplyChain := range list.Items {
		if supplyChainSelectorMatchesWorkloadLabels(supplyChain.Spec.Selector, workload.Labels) {
			clusterSupplyChains = append(clusterSupplyChains, supplyChain)
		}
	}

	return clusterSupplyChains, nil
}

func (r *repository) GetWorkload(name string, namespace string) (*v1alpha1.Workload, error) {
	workload := v1alpha1.Workload{}

	err := r.cl.Get(context.TODO(),
		client.ObjectKey{
			Name:      name,
			Namespace: namespace,
		},
		&workload,
	)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	return &workload, nil
}

func (r *repository) GetPipeline(name string, namespace string) (*v1alpha1.Pipeline, error) {
	pipeline := &v1alpha1.Pipeline{}

	err := r.cl.Get(context.TODO(),
		client.ObjectKey{
			Name:      name,
			Namespace: namespace,
		},
		pipeline,
	)

	if err != nil {
		return nil, fmt.Errorf("get-pipeline: %w", err)
	}

	return pipeline, nil
}

func supplyChainSelectorMatchesWorkloadLabels(selector map[string]string, labels map[string]string) bool {
	for key, value := range selector {
		if labels[key] != value {
			return false
		}
	}

	return true
}

func (r *repository) GetSupplyChain(name string) (*v1alpha1.ClusterSupplyChain, error) {
	supplyChain := v1alpha1.ClusterSupplyChain{}

	err := r.cl.Get(context.TODO(),
		client.ObjectKey{
			Name: name,
		},
		&supplyChain,
	)
	if err != nil && !api_errors.IsNotFound(err) {
		return nil, fmt.Errorf("get: %w", err)
	}

	if api_errors.IsNotFound(err) {
		return nil, nil
	}

	return &supplyChain, nil
}

func (r *repository) StatusUpdate(object client.Object) error {
	return r.cl.Status().Update(context.TODO(), object)
}

func (r *repository) GetScheme() *runtime.Scheme {
	return r.cl.Scheme()
}
