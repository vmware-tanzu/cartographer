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

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/logger"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate sigs.k8s.io/controller-runtime/pkg/client.Client

//counterfeiter:generate . Repository
type Repository interface {
	EnsureObjectExistsOnCluster(ctx context.Context, obj *unstructured.Unstructured, allowUpdate bool) error
	GetClusterTemplate(ctx context.Context, ref v1alpha1.ClusterTemplateReference) (client.Object, error)
	GetDeliveryClusterTemplate(ctx context.Context, ref v1alpha1.DeliveryClusterTemplateReference) (client.Object, error)
	GetRunTemplate(ctx context.Context, ref v1alpha1.TemplateReference) (*v1alpha1.ClusterRunTemplate, error)
	GetSupplyChainsForWorkload(ctx context.Context, workload *v1alpha1.Workload) ([]*v1alpha1.ClusterSupplyChain, error)
	GetDeliveriesForDeliverable(ctx context.Context, deliverable *v1alpha1.Deliverable) ([]*v1alpha1.ClusterDelivery, error)
	GetWorkload(ctx context.Context, name string, namespace string) (*v1alpha1.Workload, error)
	GetDeliverable(ctx context.Context, name string, namespace string) (*v1alpha1.Deliverable, error)
	GetSupplyChain(ctx context.Context, name string) (*v1alpha1.ClusterSupplyChain, error)
	StatusUpdate(ctx context.Context, object client.Object) error
	GetRunnable(ctx context.Context, name string, namespace string) (*v1alpha1.Runnable, error)
	ListUnstructured(ctx context.Context, obj *unstructured.Unstructured) ([]*unstructured.Unstructured, error)
	GetDelivery(ctx context.Context, name string) (*v1alpha1.ClusterDelivery, error)
	GetScheme() *runtime.Scheme
	GetServiceAccountSecret(ctx context.Context, serviceAccountName, ns string) (*corev1.Secret, error)
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

func (r *repository) GetServiceAccountSecret(ctx context.Context, name, namespace string) (*corev1.Secret, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues("service account", fmt.Sprintf("%s/%s", namespace, name))
	ctx = logr.NewContext(ctx, log)
	log.V(logger.DEBUG).Info("GetServiceAccountSecret")

	serviceAccount := corev1.ServiceAccount{}
	err := r.getObject(ctx, name, namespace, &serviceAccount)
	if err != nil {
		log.Error(err, "failed to get service account object from api server")
		return nil, fmt.Errorf("failed to get service account object from api server [%s/%s]: %w", namespace, name, err)
	}

	if len(serviceAccount.Secrets) == 0 {
		log.V(logger.DEBUG).Info("service account does not have any secrets")
		return nil, fmt.Errorf("service account [%s/%s] does not have any secrets", namespace, name)
	}

	for _, secretRef := range serviceAccount.Secrets {
		secret := corev1.Secret{}
		err := r.getObject(ctx, secretRef.Name, namespace, &secret)
		if err != nil {
			log.Error(err, "failed to get secret object from api server",
				"secret", fmt.Sprintf("%s/%s", namespace, secretRef.Name))
			return nil, fmt.Errorf("failed to get secret object from api server: %w", err)
		}

		if secret.Type == corev1.SecretTypeServiceAccountToken {
			return &secret, nil
		}
	}

	log.V(logger.DEBUG).Info("service account does not have any token secrets")
	return nil, fmt.Errorf("service account [%s/%s] does not have any token secrets", namespace, name)
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

func (r *repository) EnsureObjectExistsOnCluster(ctx context.Context, obj *unstructured.Unstructured, allowUpdate bool) error {
	log := logr.FromContextOrDiscard(ctx)
	log.V(logger.DEBUG).Info("EnsureObjectExistsOnCluster")

	unstructuredList, err := r.ListUnstructured(ctx, obj)

	for _, considered := range unstructuredList {
		log.V(logger.DEBUG).Info("considering objects from api server",
			"considered", considered)
	}

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
		log.Info("patching object", "object", obj)
		return r.patchUnstructured(ctx, outdatedObject, obj)
	} else {
		log.Info("creating object", "object", obj)
		return r.createUnstructured(ctx, obj)
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

func (r *repository) ListUnstructured(ctx context.Context, obj *unstructured.Unstructured) ([]*unstructured.Unstructured, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.V(logger.DEBUG).Info("ListUnstructured")

	unstructuredList := &unstructured.UnstructuredList{}
	unstructuredList.SetGroupVersionKind(obj.GroupVersionKind())

	opts := []client.ListOption{
		client.InNamespace(obj.GetNamespace()),
		client.MatchingLabels(obj.GetLabels()),
	}
	log.V(logger.DEBUG).Info("list unstructured with namespace and labels",
		"namespace", obj.GetNamespace(), "labels", obj.GetLabels())
	err := r.cl.List(ctx, unstructuredList, opts...)
	if err != nil {
		log.Error(err, "unable to list from api server")
		return nil, fmt.Errorf("unable to list from api server: %w", err)
	}

	pointersToUnstructureds := make([]*unstructured.Unstructured, len(unstructuredList.Items))

	//FIXME: why are we taking a deep copy?
	for i, item := range unstructuredList.Items {
		log.V(logger.DEBUG).Info("unstructured that matched",
			"namespace", obj.GetNamespace(), "labels", obj.GetLabels(), "unstructured", item)
		pointersToUnstructureds[i] = item.DeepCopy()
	}
	return pointersToUnstructureds, nil
}

func (r *repository) GetClusterTemplate(ctx context.Context, ref v1alpha1.ClusterTemplateReference) (client.Object, error) {
	return r.getTemplate(ctx, ref.Name, ref.Kind)
}

func (r *repository) GetDeliveryClusterTemplate(ctx context.Context, ref v1alpha1.DeliveryClusterTemplateReference) (client.Object, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.V(logger.DEBUG).Info("GetDeliveryClusterTemplate")

	return r.getTemplate(ctx, ref.Name, ref.Kind)
}

func (r *repository) getTemplate(ctx context.Context, name string, kind string) (client.Object, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.V(logger.DEBUG).Info("getTemplate")

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

func (r *repository) createUnstructured(ctx context.Context, obj *unstructured.Unstructured) error {
	submitted := obj.DeepCopy()
	if err := r.cl.Create(ctx, obj); err != nil {
		return fmt.Errorf("create: %w", err)
	}

	r.rc.Set(submitted, obj.DeepCopy())
	return nil
}

func (r *repository) patchUnstructured(ctx context.Context, existingObj *unstructured.Unstructured, obj *unstructured.Unstructured) error {
	submitted := obj.DeepCopy()

	obj.SetResourceVersion(existingObj.GetResourceVersion())
	if err := r.cl.Patch(ctx, obj, client.MergeFrom(existingObj)); err != nil {
		return fmt.Errorf("patch: %w", err)
	}

	r.rc.Set(submitted, obj.DeepCopy())
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

	selectorGetters := []SelectorGetter{}
	for _, item := range list.Items {
		item := item
		selectorGetters = append(selectorGetters, &item)
	}

	supplyChains := []*v1alpha1.ClusterSupplyChain{}
	for _, matchingObject := range BestLabelMatches(workload, selectorGetters) {
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

	selectorGetters := []SelectorGetter{}
	for _, item := range list.Items {
		item := item
		selectorGetters = append(selectorGetters, &item)
	}

	deliveries := []*v1alpha1.ClusterDelivery{}
	for _, matchingObject := range BestLabelMatches(deliverable, selectorGetters) {
		log.V(logger.DEBUG).Info("delivery matched deliverable",
			"delivery", matchingObject)
		deliveries = append(deliveries, matchingObject.(*v1alpha1.ClusterDelivery))
	}

	log.V(logger.DEBUG).Info("deliveries matched deliverable",
		"deliveries", deliveries)
	return deliveries, nil
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
		//TODO: use namespacedName everywhere in repo, not everything has a namespace
		var namespacedName string
		if namespace == "" {
			namespacedName = name
		} else {
			namespacedName = fmt.Sprintf("%s/%s", namespace, name)
		}
		log.Error(err, "failed to get object from api server", "object", namespacedName)
		return fmt.Errorf("failed to get object [%s] from api server: %w", namespacedName, err)
	}

	return nil
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
