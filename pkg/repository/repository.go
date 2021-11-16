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
	"strings"

	api_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate sigs.k8s.io/controller-runtime/pkg/client.Client

//counterfeiter:generate . Repository
type Repository interface {
	EnsureObjectExistsOnCluster(obj *unstructured.Unstructured, allowUpdate bool) error
	GetClusterTemplate(reference v1alpha1.ClusterTemplateReference) (client.Object, error)
	GetDeliveryClusterTemplate(reference v1alpha1.DeliveryClusterTemplateReference) (client.Object, error)
	GetRunTemplate(reference v1alpha1.TemplateReference) (*v1alpha1.ClusterRunTemplate, error)
	GetSupplyChainsByLabels(workload *v1alpha1.Workload) ([]v1alpha1.ClusterSupplyChain, error)
	GetDeliveriesForDeliverable(deliverable *v1alpha1.Deliverable) ([]v1alpha1.ClusterDelivery, error)
	GetWorkload(name string, namespace string) (*v1alpha1.Workload, error)
	GetDeliverable(name string, namespace string) (*v1alpha1.Deliverable, error)
	GetSupplyChain(name string) (*v1alpha1.ClusterSupplyChain, error)
	StatusUpdate(object client.Object) error
	GetScheme() *runtime.Scheme
	GetRunnable(name string, namespace string) (*v1alpha1.Runnable, error)
	ListUnstructured(obj *unstructured.Unstructured) ([]*unstructured.Unstructured, error)
	GetDelivery(name string) (*v1alpha1.ClusterDelivery, error)
}

type repository struct {
	rc     RepoCache
	cl     client.Client
	logger Logger
}

func NewRepository(client client.Client, repoCache RepoCache, logger Logger) Repository {
	return &repository{
		rc:     repoCache,
		cl:     client,
		logger: logger,
	}
}

func (r *repository) GetDelivery(name string) (*v1alpha1.ClusterDelivery, error) {
	delivery := &v1alpha1.ClusterDelivery{}

	key := client.ObjectKey{
		Name: name,
	}

	err := r.cl.Get(context.TODO(), key, delivery)

	if err != nil && !api_errors.IsNotFound(err) {
		return nil, fmt.Errorf("get: %w", err)
	}

	if api_errors.IsNotFound(err) {
		return nil, nil
	}

	return delivery, nil
}

func (r *repository) EnsureObjectExistsOnCluster(obj *unstructured.Unstructured, allowUpdate bool) error {
	unstructuredList, err := r.ListUnstructured(obj)

	var names []string
	for _, considered := range unstructuredList {
		names = append(names, considered.GetName())
	}
	r.logger.Info("considering objects from apiserver", "consideredList", strings.Join(names, ", "))

	if err != nil {
		return err
	}

	cacheHit := r.rc.UnchangedSinceCached(obj, unstructuredList)
	if cacheHit != nil {
		*obj = *cacheHit
		return nil
	}

	var outdatedObject *unstructured.Unstructured
	if allowUpdate {
		outdatedObject = getOutdatedUnstructuredByName(obj, unstructuredList)
	}

	if outdatedObject != nil {
		r.logger.Info("patching object", "name", obj.GetName(), "namespace", obj.GetNamespace(), "kind", obj.GetKind())
		return r.patchUnstructured(outdatedObject, obj)
	} else {
		r.logger.Info("creating object", "name", obj.GetName(), "namespace", obj.GetNamespace(), "kind", obj.GetKind())
		return r.createUnstructured(obj)
	}
}

func getOutdatedUnstructuredByName(target *unstructured.Unstructured, candidates []*unstructured.Unstructured) *unstructured.Unstructured {
	for _, candidate := range candidates {
		if candidate.GetName() == target.GetName() && candidate.GetNamespace() == target.GetNamespace() {
			return candidate
		}
	}
	return nil
}

func (r *repository) ListUnstructured(obj *unstructured.Unstructured) ([]*unstructured.Unstructured, error) {
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

	pointersToUnstructureds := make([]*unstructured.Unstructured, len(unstructuredList.Items))

	for i, item := range unstructuredList.Items {
		pointersToUnstructureds[i] = item.DeepCopy()
	}
	return pointersToUnstructureds, nil
}

func (r *repository) GetClusterTemplate(ref v1alpha1.ClusterTemplateReference) (client.Object, error) {
	return r.getTemplate(ref.Name, ref.Kind)
}

func (r *repository) GetDeliveryClusterTemplate(ref v1alpha1.DeliveryClusterTemplateReference) (client.Object, error) {
	return r.getTemplate(ref.Name, ref.Kind)
}

func (r *repository) getTemplate(name string, kind string) (client.Object, error) {
	apiTemplate, err := v1alpha1.GetAPITemplate(kind)
	if err != nil {
		return nil, fmt.Errorf("get api template: %w", err)
	}

	err = r.getObject(name, "", apiTemplate)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}
	return apiTemplate, nil
}

func (r *repository) GetRunTemplate(ref v1alpha1.TemplateReference) (*v1alpha1.ClusterRunTemplate, error) {
	runTemplate := &v1alpha1.ClusterRunTemplate{}

	err := r.cl.Get(context.TODO(), client.ObjectKey{
		Name: ref.Name,
	}, runTemplate)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	return runTemplate, nil
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
	// FIXME: I'm untested. What am I for? Patch doesn't block on RV's (is this a historical artifact of .Update?)
	obj.SetResourceVersion(existingObj.GetResourceVersion())
	if err := r.cl.Patch(context.TODO(), obj, client.MergeFrom(existingObj)); err != nil {
		return fmt.Errorf("patch: %w", err)
	}

	r.rc.Set(submitted, obj.DeepCopy())
	return nil
}

func (r *repository) GetSupplyChainsByLabels(workload *v1alpha1.Workload) ([]v1alpha1.ClusterSupplyChain, error) {
	list := &v1alpha1.ClusterSupplyChainList{}
	if err := r.cl.List(context.TODO(), list); err != nil {
		return nil, fmt.Errorf("list supply chains: %w", err)
	}

	var clusterSupplyChains []v1alpha1.ClusterSupplyChain
	for _, supplyChain := range list.Items {
		if selectorMatchesLabels(supplyChain.Spec.Selector, workload.Labels) {
			clusterSupplyChains = append(clusterSupplyChains, supplyChain)
		}
	}

	return clusterSupplyChains, nil
}

func (r *repository) GetDeliveriesForDeliverable(deliverable *v1alpha1.Deliverable) ([]v1alpha1.ClusterDelivery, error) {
	list := &v1alpha1.ClusterDeliveryList{}
	if err := r.cl.List(context.TODO(), list); err != nil {
		return nil, fmt.Errorf("list deliveries: %w", err)
	}

	var clusterDeliveries []v1alpha1.ClusterDelivery
	for _, delivery := range list.Items {
		if selectorMatchesLabels(delivery.Spec.Selector, deliverable.Labels) {
			clusterDeliveries = append(clusterDeliveries, delivery)
		}
	}

	return clusterDeliveries, nil
}

func (r *repository) getObject(name string, namespace string, obj client.Object) error {
	err := r.cl.Get(context.TODO(),
		client.ObjectKey{
			Name:      name,
			Namespace: namespace,
		},
		obj,
	)
	if err != nil {
		return fmt.Errorf("get: %w", err)
	}

	return nil
}

func (r *repository) GetWorkload(name string, namespace string) (*v1alpha1.Workload, error) {
	workload := v1alpha1.Workload{}
	err := r.getObject(name, namespace, &workload)
	if err != nil {
		return nil, err
	}
	return &workload, nil
}

func (r *repository) GetDeliverable(name string, namespace string) (*v1alpha1.Deliverable, error) {
	deliverable := v1alpha1.Deliverable{}
	err := r.getObject(name, namespace, &deliverable)
	if err != nil {
		return nil, err
	}
	return &deliverable, nil
}

func (r *repository) GetRunnable(name string, namespace string) (*v1alpha1.Runnable, error) {
	runnable := &v1alpha1.Runnable{}

	err := r.getObject(name, namespace, runnable)

	if err != nil {
		return nil, fmt.Errorf("get-runnable: %w", err)
	}

	return runnable, nil
}

func selectorMatchesLabels(selector map[string]string, labels map[string]string) bool {
	for key, value := range selector {
		if labels[key] != value {
			return false
		}
	}

	return true
}

func (r *repository) GetSupplyChain(name string) (*v1alpha1.ClusterSupplyChain, error) {
	supplyChain := v1alpha1.ClusterSupplyChain{}

	err := r.getObject(name, "", &supplyChain)
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
