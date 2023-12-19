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
	"sort"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/events"
	"github.com/vmware-tanzu/cartographer/pkg/logger"
)

//go:generate go run -modfile ../../hack/tools/go.mod github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate sigs.k8s.io/controller-runtime/pkg/client.Client

//counterfeiter:generate . Repository
type Repository interface {
	EnsureImmutableObjectExistsOnCluster(ctx context.Context, obj *unstructured.Unstructured, labels map[string]string) error
	EnsureMutableObjectExistsOnCluster(ctx context.Context, obj *unstructured.Unstructured) error
	GetTemplate(ctx context.Context, name, kind string) (client.Object, error)
	GetRunTemplate(ctx context.Context, ref v1alpha1.TemplateReference) (*v1alpha1.ClusterRunTemplate, error)
	GetSupplyChainsForWorkload(ctx context.Context, workload *v1alpha1.Workload) ([]*v1alpha1.ClusterSupplyChain, error)
	GetDeliveriesForDeliverable(ctx context.Context, deliverable *v1alpha1.Deliverable) ([]*v1alpha1.ClusterDelivery, error)
	GetWorkload(ctx context.Context, name string, namespace string) (*v1alpha1.Workload, error)
	GetDeliverable(ctx context.Context, name string, namespace string) (*v1alpha1.Deliverable, error)
	GetSupplyChain(ctx context.Context, name string) (*v1alpha1.ClusterSupplyChain, error)
	StatusUpdate(ctx context.Context, object client.Object) error
	GetRunnable(ctx context.Context, name string, namespace string) (*v1alpha1.Runnable, error)
	GetUnstructured(ctx context.Context, obj *unstructured.Unstructured) (*unstructured.Unstructured, error)
	ListUnstructured(ctx context.Context, gvk schema.GroupVersionKind, namespace string, labels map[string]string) ([]*unstructured.Unstructured, error)
	GetDelivery(ctx context.Context, name string) (*v1alpha1.ClusterDelivery, error)
	GetScheme() *runtime.Scheme
	GetRESTMapper() meta.RESTMapper
	GetServiceAccount(ctx context.Context, serviceAccountName, ns string) (*corev1.ServiceAccount, error)
	Delete(ctx context.Context, objToDelete *unstructured.Unstructured) error
}

type RepositoryBuilder func(client client.Client, repoCache RepoCache) Repository

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

func (r *repository) Delete(ctx context.Context, objToDelete *unstructured.Unstructured) error {
	log := logr.FromContextOrDiscard(ctx).WithValues("delete object", fmt.Sprintf("%s/%s", objToDelete.GetNamespace(), objToDelete.GetName()))
	ctx = logr.NewContext(ctx, log)
	log.V(logger.DEBUG).Info("Delete")

	err := r.deleteObject(ctx, objToDelete)
	if err != nil {
		log.Error(err, "failed to delete object")
		return fmt.Errorf("failed to delete object [%s/%s]: %w", objToDelete.GetNamespace(), objToDelete.GetName(), err)
	}

	log.V(logger.DEBUG).Info("object deleted successfully")
	rec := events.FromContextOrDie(ctx)
	rec.ResourceEventf(events.NormalType, events.StampedObjectRemovedReason, "Deleted object [%Q]", objToDelete)

	return nil
}

func (r *repository) GetServiceAccount(ctx context.Context, name, namespace string) (*corev1.ServiceAccount, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues("service account", fmt.Sprintf("%s/%s", namespace, name))
	ctx = logr.NewContext(ctx, log)
	log.V(logger.DEBUG).Info("GetServiceAccount")

	serviceAccount := &corev1.ServiceAccount{}
	err := r.getObject(ctx, name, namespace, serviceAccount)
	if err != nil {
		log.Error(err, "failed to get service account object from api server")
		return nil, fmt.Errorf("failed to get service account object from api server [%s/%s]: %w", namespace, name, err)
	}

	return serviceAccount, nil
}

func (r *repository) GetDelivery(ctx context.Context, name string) (*v1alpha1.ClusterDelivery, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.V(logger.DEBUG).Info("GetDelivery")

	delivery := &v1alpha1.ClusterDelivery{}

	key := client.ObjectKey{
		Name: name,
	}

	err := r.cl.Get(ctx, key, delivery)
	if kerrors.IsNotFound(err) {
		log.V(logger.DEBUG).Info("delivery is not found on api server")
		return nil, nil
	}
	if err != nil {
		log.Error(err, "failed to get delivery object from api server")
		return nil, fmt.Errorf("failed to get delivery object from api server [%s]: %w", name, err)
	}

	return delivery, nil
}

func (r *repository) EnsureMutableObjectExistsOnCluster(ctx context.Context, obj *unstructured.Unstructured) error {
	log := logr.FromContextOrDiscard(ctx)
	log.V(logger.DEBUG).Info("EnsureMutableObjectExistsOnCluster")

	existingObj, err := r.GetUnstructured(ctx, obj)
	log.V(logger.DEBUG).Info("considering object from api server",
		"considered", obj)

	if err != nil {
		return err
	}

	if existingObj != nil {
		cacheHit := r.rc.UnchangedSinceCached(obj, existingObj)
		if cacheHit != nil {
			*obj = *cacheHit
			return nil
		}
		log.Info("patching object", "object", obj)
		return r.patchUnstructured(ctx, existingObj, obj)
	} else {
		log.Info("creating object", "object", obj)
		return r.createUnstructured(ctx, obj, "")
	}
}

func (r *repository) EnsureImmutableObjectExistsOnCluster(ctx context.Context, obj *unstructured.Unstructured, labels map[string]string) error {
	log := logr.FromContextOrDiscard(ctx)
	log.V(logger.DEBUG).Info("EnsureImmutableObjectExistsOnCluster")

	unstructuredList, err := r.ListUnstructured(ctx, obj.GroupVersionKind(), obj.GetNamespace(), labels)

	for _, considered := range unstructuredList {
		log.V(logger.DEBUG).Info("considering objects from api server",
			"considered", considered)
	}

	if err != nil {
		return err
	}

	ownerDiscriminant := buildOwnerDiscriminant(labels)

	cacheHit := r.rc.UnchangedSinceCachedFromList(obj, unstructuredList, ownerDiscriminant)
	if cacheHit != nil {
		*obj = *cacheHit
		return nil
	}

	log.Info("creating object", "object", obj)
	return r.createUnstructured(ctx, obj, ownerDiscriminant)
}

func (r *repository) GetUnstructured(ctx context.Context, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.V(logger.DEBUG).Info("GetUnstructured")

	objKey := client.ObjectKey{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}

	log.V(logger.DEBUG).Info("get unstructured with name and namespace",
		"name", obj.GetName(), "namespace", obj.GetNamespace())

	returnObj := &unstructured.Unstructured{}
	returnObj.SetGroupVersionKind(obj.GroupVersionKind())
	err := r.cl.Get(ctx, objKey, returnObj)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil, nil
		}
		namespacedName := types.NamespacedName{
			Namespace: obj.GetNamespace(),
			Name:      obj.GetName(),
		}
		log.Error(err, "failed to get unstructured from api server", "object", namespacedName)
		return nil, fmt.Errorf("failed to get unstructured [%s] from api server: %w", namespacedName, err)
	}

	return returnObj, nil
}

func (r *repository) ListUnstructured(ctx context.Context, gvk schema.GroupVersionKind, namespace string, labels map[string]string) ([]*unstructured.Unstructured, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.V(logger.DEBUG).Info("ListUnstructured")

	unstructuredList := &unstructured.UnstructuredList{}
	unstructuredList.SetGroupVersionKind(gvk)

	opts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(labels),
	}
	log.V(logger.DEBUG).Info("list unstructured with namespace and labels",
		"namespace", namespace, "labels", labels)
	err := r.cl.List(ctx, unstructuredList, opts...)
	if err != nil {
		log.Error(err, "unable to list from api server")
		return nil, fmt.Errorf("unable to list from api server: %w", err)
	}

	pointersToUnstructureds := make([]*unstructured.Unstructured, len(unstructuredList.Items))

	//FIXME: why are we taking a deep copy?
	for i, item := range unstructuredList.Items {
		log.V(logger.DEBUG).Info("unstructured that matched",
			"namespace", namespace, "labels", labels, "unstructured", item)
		pointersToUnstructureds[i] = item.DeepCopy()
	}
	return pointersToUnstructureds, nil
}

func (r *repository) GetTemplate(ctx context.Context, name string, kind string) (client.Object, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.V(logger.DEBUG).Info("GetTemplate")

	apiTemplate, err := v1alpha1.GetAPITemplate(kind)
	if err != nil {
		log.Error(err, "unable to get api template")
		return nil, fmt.Errorf("unable to get api template [%s/%s]: %w", kind, name, err)
	}

	err = r.getObject(ctx, name, "", apiTemplate)
	//TODO: Remove IsNotFound check, this should just be an error, breaks kuttl test
	if kerrors.IsNotFound(err) {
		log.V(logger.DEBUG).Info("template is not found on api server")
		return nil, nil
	}
	if err != nil {
		log.Error(err, "failed to get template object from api server")
		return nil, fmt.Errorf("failed to get template object from api server [%s/%s]: %w", kind, name, err)
	}

	return apiTemplate, nil
}

func (r *repository) GetRunTemplate(ctx context.Context, ref v1alpha1.TemplateReference) (*v1alpha1.ClusterRunTemplate, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.V(logger.DEBUG).Info("GetRunTemplate")

	runTemplate := &v1alpha1.ClusterRunTemplate{}

	err := r.cl.Get(ctx, client.ObjectKey{
		Name: ref.Name,
	}, runTemplate)
	if err != nil {
		log.Error(err, "failed to get run template object from api server")
		return nil, fmt.Errorf("failed to get run template object from api server [%s/%s]: %w", ref.Kind, ref.Name, err)
	}

	return runTemplate, nil
}

func (r *repository) createUnstructured(ctx context.Context, obj *unstructured.Unstructured, ownerDiscriminant string) error {
	submitted := obj.DeepCopy()
	if err := r.cl.Create(ctx, obj); err != nil {
		return fmt.Errorf("create: %w", err)
	}

	r.rc.Set(submitted, obj.DeepCopy(), ownerDiscriminant)

	rec := events.FromContextOrDie(ctx)
	rec.ResourceEventf(events.NormalType, events.StampedObjectAppliedReason, "Created object [%Q]", obj)
	return nil
}

func (r *repository) patchUnstructured(ctx context.Context, existingObj *unstructured.Unstructured, obj *unstructured.Unstructured) error {
	submitted := obj.DeepCopy()

	obj.SetResourceVersion(existingObj.GetResourceVersion())
	if err := r.cl.Patch(ctx, obj, client.MergeFrom(existingObj)); err != nil {
		return fmt.Errorf("patch: %w", err)
	}

	r.rc.Set(submitted, obj.DeepCopy(), "")

	rec := events.FromContextOrDie(ctx)
	rec.ResourceEventf(events.NormalType, events.StampedObjectAppliedReason, "Patched object [%Q]", obj)
	return nil
}

func (r *repository) GetSupplyChainsForWorkload(ctx context.Context, workload *v1alpha1.Workload) ([]*v1alpha1.ClusterSupplyChain, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.V(logger.DEBUG).Info("GetSupplyChainsForWorkload")

	list := &v1alpha1.ClusterSupplyChainList{}
	if err := r.cl.List(ctx, list); err != nil {
		log.Error(err, "unable to list supply chains from api server")
		return nil, fmt.Errorf("unable to list supply chains from api server: %w", err)
	}

	var supplyChains []*v1alpha1.ClusterSupplyChain

	for _, item := range list.Items {
		itemValue := item
		supplyChains = append(supplyChains, &itemValue)
	}

	return GetSelectedSupplyChain(supplyChains, workload, log)
}

func GetSelectedSupplyChain(allSupplyChains []*v1alpha1.ClusterSupplyChain, workload *v1alpha1.Workload, log logr.Logger) ([]*v1alpha1.ClusterSupplyChain, error) {
	var selectorGetters []SelectingObject
	for _, item := range allSupplyChains {
		itemValue := item
		selectorGetters = append(selectorGetters, itemValue)
	}

	var supplyChains []*v1alpha1.ClusterSupplyChain
	matches, err := BestSelectorMatch(workload, selectorGetters)
	if err != nil {
		return nil, fmt.Errorf("evaluating supply chain selectors against workload [%s/%s] failed: %w", workload.Namespace, workload.Name, err)
	}
	for _, matchingObject := range matches {
		log.V(logger.DEBUG).Info("supply chain matched workload",
			"supply chain", matchingObject)
		supplyChains = append(supplyChains, matchingObject.(*v1alpha1.ClusterSupplyChain))
	}

	return supplyChains, nil
}

func (r *repository) GetDeliveriesForDeliverable(ctx context.Context, deliverable *v1alpha1.Deliverable) ([]*v1alpha1.ClusterDelivery, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.V(logger.DEBUG).Info("GetDeliveriesForDeliverable")

	list := &v1alpha1.ClusterDeliveryList{}
	if err := r.cl.List(ctx, list); err != nil {
		log.Error(err, "unable to list deliveries from api server")
		return nil, fmt.Errorf("unable to list deliveries from api server: %w", err)
	}

	var selectorGetters []SelectingObject
	for _, item := range list.Items {
		itemValue := item
		selectorGetters = append(selectorGetters, &itemValue)
	}

	var deliveries []*v1alpha1.ClusterDelivery
	matches, err := BestSelectorMatch(deliverable, selectorGetters)
	if err != nil {
		return nil, fmt.Errorf("evaluating supply chain selectors against deliverable [%s/%s] failed: %w", deliverable.Namespace, deliverable.Name, err)
	}
	for _, matchingObject := range matches {
		log.V(logger.DEBUG).Info("delivery matched deliverable",
			"delivery", matchingObject)
		deliveries = append(deliveries, matchingObject.(*v1alpha1.ClusterDelivery))
	}

	log.V(logger.DEBUG).Info("deliveries matched deliverable",
		"deliveries", deliveries)
	return deliveries, nil
}
func getNamespacedName(name string, namespace string) string {
	var namespacedName string
	if namespace == "" {
		namespacedName = name
	} else {
		namespacedName = fmt.Sprintf("%s/%s", namespace, name)
	}
	return namespacedName
}

func (r *repository) getObject(ctx context.Context, name string, namespace string, obj client.Object) error {
	log := logr.FromContextOrDiscard(ctx)
	log.V(logger.DEBUG).Info("getObject")

	err := r.cl.Get(ctx,
		client.ObjectKey{
			Name:      name,
			Namespace: namespace,
		},
		obj,
	)
	if err != nil {
		namespacedName := getNamespacedName(name, namespace)
		return fmt.Errorf("failed to get object [%s] from api server: %w", namespacedName, err)
	}

	return nil
}

func (r *repository) deleteObject(ctx context.Context, unstructuredObj *unstructured.Unstructured) error {
	err := r.cl.Delete(ctx, unstructuredObj)

	return err
}

func (r *repository) GetWorkload(ctx context.Context, name string, namespace string) (*v1alpha1.Workload, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.V(logger.DEBUG).Info("GetWorkload")

	workload := v1alpha1.Workload{}
	err := r.getObject(ctx, name, namespace, &workload)
	if kerrors.IsNotFound(err) {
		log.V(logger.DEBUG).Info("workload is not found on api server")
		return nil, nil
	}
	if err != nil {
		log.Error(err, "failed to get workload object from api server")
		return nil, fmt.Errorf("failed to get workload object from api server [%s/%s]: %w", namespace, name, err)
	}

	return &workload, nil
}

func (r *repository) GetDeliverable(ctx context.Context, name string, namespace string) (*v1alpha1.Deliverable, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.V(logger.DEBUG).Info("GetDeliverable")

	deliverable := v1alpha1.Deliverable{}
	err := r.getObject(ctx, name, namespace, &deliverable)
	if kerrors.IsNotFound(err) {
		log.V(logger.DEBUG).Info("deliverable is not found on api server")
		return nil, nil
	}

	if err != nil {
		log.Error(err, "failed to get deliverable object from api server")
		return nil, fmt.Errorf("failed to get deliverable object from api server [%s/%s]: %w", namespace, name, err)
	}
	return &deliverable, nil
}

func (r *repository) GetRunnable(ctx context.Context, name string, namespace string) (*v1alpha1.Runnable, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.V(logger.DEBUG).Info("GetDeliverable")

	runnable := &v1alpha1.Runnable{}

	err := r.getObject(ctx, name, namespace, runnable)
	if kerrors.IsNotFound(err) {
		log.V(logger.DEBUG).Info("runnable is not found on api server")
		return nil, nil
	}
	if err != nil {
		log.Error(err, "failed to get runnable object from api server")
		return nil, fmt.Errorf("failed to get runnable object from api server [%s/%s]: %w", namespace, name, err)
	}

	return runnable, nil
}

func (r *repository) GetSupplyChain(ctx context.Context, name string) (*v1alpha1.ClusterSupplyChain, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.V(logger.DEBUG).Info("GetSupplyChain")

	supplyChain := v1alpha1.ClusterSupplyChain{}

	err := r.getObject(ctx, name, "", &supplyChain)
	if kerrors.IsNotFound(err) {
		log.V(logger.DEBUG).Info("supply chain is not found on api server")
		return nil, nil
	}
	if err != nil {
		log.Error(err, "failed to get supply chain object from api server")
		return nil, fmt.Errorf("failed to get supply chain object from api server [%s]: %w", name, err)
	}

	return &supplyChain, nil
}

func (r *repository) StatusUpdate(ctx context.Context, object client.Object) error {
	return r.cl.Status().Update(ctx, object)
}

func (r *repository) GetScheme() *runtime.Scheme {
	return r.cl.Scheme()
}

func (r *repository) GetRESTMapper() meta.RESTMapper {
	return r.cl.RESTMapper()
}

func buildOwnerDiscriminant(labels map[string]string) string {
	var discriminantComponents []string

	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		discriminantComponents = append(discriminantComponents, fmt.Sprintf("{%s:%s}", k, labels[k]))
	}

	return strings.Join(discriminantComponents, "")
}
