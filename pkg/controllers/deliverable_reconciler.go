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

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/controllers/external"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/enqueuer"
	cerrors "github.com/vmware-tanzu/cartographer/pkg/errors"
	"github.com/vmware-tanzu/cartographer/pkg/logger"
	"github.com/vmware-tanzu/cartographer/pkg/mapper"
	"github.com/vmware-tanzu/cartographer/pkg/realizer"
	realizerclient "github.com/vmware-tanzu/cartographer/pkg/realizer/client"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/statuses"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/dependency"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/stamped"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

type DeliverableReconciler struct {
	Repo                    repository.Repository
	ConditionManagerBuilder conditions.ConditionManagerBuilder
	ResourceRealizerBuilder realizer.ResourceRealizerBuilder
	Realizer                realizer.Realizer
	StampedTracker          stamped.StampedTracker
	DependencyTracker       dependency.DependencyTracker
	conditionManager        conditions.ConditionManager
}

func (r *DeliverableReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.Info("started")
	defer log.Info("finished")

	log = log.WithValues("deliverable", req.NamespacedName)
	ctx = logr.NewContext(ctx, log)

	deliverable, err := r.Repo.GetDeliverable(ctx, req.Name, req.Namespace)
	if err != nil {
		log.Error(err, "failed to get deliverable")
		return ctrl.Result{}, fmt.Errorf("failed to get deliverable [%s]: %w", req.NamespacedName, err)
	}

	if deliverable == nil {
		log.Info("deliverable no longer exists")
		r.DependencyTracker.ClearTracked(types.NamespacedName{
			Namespace: req.Namespace,
			Name:      req.Name,
		})

		return ctrl.Result{}, nil
	}

	r.conditionManager = r.ConditionManagerBuilder(v1alpha1.OwnerReady, deliverable.Status.Conditions)

	delivery, err := r.getDeliveriesForDeliverable(ctx, deliverable)
	if err != nil {
		return r.completeReconciliation(ctx, deliverable, nil, err)
	}

	log = log.WithValues("delivery", delivery.Name)
	ctx = logr.NewContext(ctx, log)

	deliveryGVK, err := utils.GetObjectGVK(delivery, r.Repo.GetScheme())
	if err != nil {
		log.Error(err, "failed to get object gvk for delivery")
		return r.completeReconciliation(ctx, deliverable, nil, cerrors.NewUnhandledError(
			fmt.Errorf("failed to get object gvk for delivery [%s]: %w", delivery.Name, err)))
	}

	deliverable.Status.DeliveryRef.Kind = deliveryGVK.Kind
	deliverable.Status.DeliveryRef.Name = delivery.Name

	if !r.isDeliveryReady(delivery) {
		r.conditionManager.AddPositive(conditions.MissingReadyInDeliveryCondition(getDeliveryReadyCondition(delivery)))
		log.Info("delivery is not in ready state")
		return r.completeReconciliation(ctx, deliverable, nil, fmt.Errorf("delivery [%s] is not in ready state", delivery.Name))
	}
	r.conditionManager.AddPositive(conditions.DeliveryReadyCondition())

	serviceAccountName, serviceAccountNS := getServiceAccountNameAndNamespaceForDeliverable(deliverable, delivery)

	secret, err := r.Repo.GetServiceAccountSecret(ctx, serviceAccountName, serviceAccountNS)
	if err != nil {
		r.conditionManager.AddPositive(conditions.ServiceAccountSecretNotFoundCondition(err))
		return r.completeReconciliation(ctx, deliverable, nil, fmt.Errorf("failed to get secret for service account [%s]: %w", fmt.Sprintf("%s/%s", serviceAccountNS, serviceAccountName), err))
	}

	resourceRealizer, err := r.ResourceRealizerBuilder(secret, deliverable, deliverable.Spec.Params, r.Repo, delivery.Spec.Params, buildDeliverableResourceLabeler(deliverable, delivery))

	if err != nil {
		r.conditionManager.AddPositive(conditions.ResourceRealizerBuilderErrorCondition(err))
		return r.completeReconciliation(ctx, deliverable, nil, cerrors.NewUnhandledError(fmt.Errorf("failed to build resource realizer: %w", err)))
	}

	resourceStatuses := statuses.NewResourceStatuses(deliverable.Status.Resources, conditions.AddConditionForResourceSubmittedDeliverable)
	err = r.Realizer.Realize(ctx, resourceRealizer, delivery.Name, realizer.MakeDeliveryOwnerResources(delivery), resourceStatuses)

	if err != nil {
		conditions.AddConditionForResourceSubmittedDeliverable(&r.conditionManager, true, err)
	} else {
		r.conditionManager.AddPositive(conditions.ResourcesSubmittedCondition(true))
	}

	if err != nil {
		log.V(logger.DEBUG).Info("failed to realize")
		if cerrors.IsUnhandledErrorType(err) {
			err = cerrors.NewUnhandledError(err)
		}
	} else {
		if log.V(logger.DEBUG).Enabled() {
			for _, resource := range resourceStatuses.GetCurrent() {
				log.V(logger.DEBUG).Info("realized object",
					"object", resource.StampedRef)
			}
		}
	}

	r.trackDependencies(deliverable, resourceStatuses.GetCurrent(), serviceAccountName, serviceAccountNS)

	cleanupErr := r.cleanupOrphanedObjects(ctx, deliverable.Status.Resources, resourceStatuses.GetCurrent())
	if cleanupErr != nil {
		log.Error(cleanupErr, "failed to cleanup orphaned objects")
	}

	var trackingError error
	for _, resource := range resourceStatuses.GetCurrent() {
		if resource.StampedRef == nil {
			continue
		}
		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(resource.StampedRef.GroupVersionKind())

		trackingError = r.StampedTracker.Watch(log, obj, &handler.EnqueueRequestForOwner{OwnerType: &v1alpha1.Deliverable{}})
		if trackingError != nil {
			log.Error(err, "failed to add informer for object",
				"object", resource.StampedRef)
			err = cerrors.NewUnhandledError(trackingError)
		} else {
			log.V(logger.DEBUG).Info("added informer for object",
				"object", resource.StampedRef)
		}
	}

	return r.completeReconciliation(ctx, deliverable, resourceStatuses, err)
}

func (r *DeliverableReconciler) completeReconciliation(ctx context.Context, deliverable *v1alpha1.Deliverable, resourceStatuses statuses.ResourceStatuses, err error) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx)
	var changed bool
	deliverable.Status.Conditions, changed = r.conditionManager.Finalize()

	var updateErr error
	if changed || (deliverable.Status.ObservedGeneration != deliverable.Generation) || (resourceStatuses != nil && resourceStatuses.IsChanged()) {
		if resourceStatuses != nil {
			deliverable.Status.Resources = resourceStatuses.GetCurrent()
		}

		deliverable.Status.ObservedGeneration = deliverable.Generation
		updateErr = r.Repo.StatusUpdate(ctx, deliverable)
		if updateErr != nil {
			log.Error(err, "failed to update status for deliverable")
			return ctrl.Result{}, fmt.Errorf("failed to update status for deliverable: %w", updateErr)
		}
	}

	if err != nil {
		if cerrors.IsUnhandledError(err) {
			log.Error(err, "unhandled error reconciling deliverable")
			return ctrl.Result{}, err
		}
		log.Info("handled error reconciling deliverable", "handled error", err)
	}

	return ctrl.Result{}, nil
}

func (r *DeliverableReconciler) isDeliveryReady(delivery *v1alpha1.ClusterDelivery) bool {
	readyCondition := getDeliveryReadyCondition(delivery)
	return readyCondition.Status == "True"
}

func buildDeliverableResourceLabeler(owner, blueprint client.Object) realizer.ResourceLabeler {
	return func(resource realizer.OwnerResource) templates.Labels {
		return templates.Labels{
			"carto.run/deliverable-name":      owner.GetName(),
			"carto.run/deliverable-namespace": owner.GetNamespace(),
			"carto.run/delivery-name":         blueprint.GetName(),
			"carto.run/resource-name":         resource.Name,
			"carto.run/template-kind":         resource.TemplateRef.Kind,
			"carto.run/cluster-template-name": resource.TemplateRef.Name,
		}
	}
}

func (r *DeliverableReconciler) trackDependencies(deliverable *v1alpha1.Deliverable, realizedResources []v1alpha1.ResourceStatus, serviceAccountName, serviceAccountNS string) {
	r.DependencyTracker.ClearTracked(types.NamespacedName{
		Namespace: deliverable.Namespace,
		Name:      deliverable.Name,
	})

	r.DependencyTracker.Track(dependency.Key{
		GroupKind: schema.GroupKind{
			Group: corev1.SchemeGroupVersion.Group,
			Kind:  rbacv1.ServiceAccountKind,
		},
		NamespacedName: types.NamespacedName{
			Namespace: serviceAccountNS,
			Name:      serviceAccountName,
		},
	}, types.NamespacedName{
		Namespace: deliverable.Namespace,
		Name:      deliverable.Name,
	})

	for _, resource := range realizedResources {
		if resource.TemplateRef == nil {
			continue
		}
		r.DependencyTracker.Track(dependency.Key{
			GroupKind: schema.GroupKind{
				Group: v1alpha1.SchemeGroupVersion.Group,
				Kind:  resource.TemplateRef.Kind,
			},
			NamespacedName: types.NamespacedName{
				Namespace: "",
				Name:      resource.TemplateRef.Name,
			},
		}, types.NamespacedName{
			Namespace: deliverable.Namespace,
			Name:      deliverable.Name,
		})
	}
}

func (r *DeliverableReconciler) cleanupOrphanedObjects(ctx context.Context, previousResources, realizedResources []v1alpha1.ResourceStatus) error {
	log := logr.FromContextOrDiscard(ctx)

	var orphanedObjs []*corev1.ObjectReference
	for _, prevResource := range previousResources {
		if prevResource.StampedRef == nil {
			continue
		}
		orphaned := true
		for _, realizedResource := range realizedResources {
			if realizedResource.StampedRef == nil {
				continue
			}
			if realizedResource.StampedRef.GroupVersionKind() == prevResource.StampedRef.GroupVersionKind() &&
				realizedResource.StampedRef.Namespace == prevResource.StampedRef.Namespace &&
				realizedResource.StampedRef.Name == prevResource.StampedRef.Name {
				orphaned = false
				break
			}
		}
		if orphaned {
			orphanedObjs = append(orphanedObjs, prevResource.StampedRef)
		}
	}

	for _, orphanedObj := range orphanedObjs {
		obj := &unstructured.Unstructured{}
		obj.SetNamespace(orphanedObj.Namespace)
		obj.SetName(orphanedObj.Name)
		obj.SetGroupVersionKind(orphanedObj.GroupVersionKind())

		log.V(logger.DEBUG).Info("deleting orphaned object", "object", orphanedObj)
		err := r.Repo.Delete(ctx, obj)
		if err != nil {
			return err
		}
	}

	return nil
}

func getDeliveryReadyCondition(delivery *v1alpha1.ClusterDelivery) metav1.Condition {
	for _, condition := range delivery.Status.Conditions {
		if condition.Type == "Ready" {
			return condition
		}
	}
	return metav1.Condition{}
}

func (r *DeliverableReconciler) getDeliveriesForDeliverable(ctx context.Context, deliverable *v1alpha1.Deliverable) (*v1alpha1.ClusterDelivery, error) {
	log := logr.FromContextOrDiscard(ctx)
	if len(deliverable.Labels) == 0 {
		r.conditionManager.AddPositive(conditions.DeliverableMissingLabelsCondition())
		log.Info("deliverable is missing required labels")
		return nil, fmt.Errorf("deliverable [%s/%s] is missing required labels",
			deliverable.Namespace, deliverable.Name)
	}

	deliveries, err := r.Repo.GetDeliveriesForDeliverable(ctx, deliverable)
	if err != nil {
		log.Error(err, "failed to get deliveries for deliverable")
		return nil, cerrors.NewUnhandledError(fmt.Errorf("failed to get deliveries for deliverable [%s/%s]: %w",
			deliverable.Namespace, deliverable.Name, err))
	}

	if len(deliveries) == 0 {
		r.conditionManager.AddPositive(conditions.DeliveryNotFoundCondition(deliverable.Labels))
		log.Info("no delivery found where full selector is satisfied by label",
			"labels", deliverable.Labels)
		return nil, fmt.Errorf("no delivery [%s/%s] found where full selector is satisfied by labels: %v",
			deliverable.Namespace, deliverable.Name, deliverable.Labels)
	}

	if len(deliveries) > 1 {
		r.conditionManager.AddPositive(conditions.TooManyDeliveryMatchesCondition())
		log.Info("more than one delivery selected for deliverable",
			"deliveries", getDeliveryNames(deliveries))
		return nil, fmt.Errorf("more than one delivery selected for deliverable [%s/%s]: %+v",
			deliverable.Namespace, deliverable.Name, getDeliveryNames(deliveries))
	}

	delivery := deliveries[0]
	log.V(logger.DEBUG).Info("delivery matched for deliverable", "delivery", delivery.Name)
	return delivery, nil
}

func getDeliveryNames(objs []*v1alpha1.ClusterDelivery) []string {
	var names []string
	for _, obj := range objs {
		names = append(names, obj.GetName())
	}

	return names
}

func getServiceAccountNameAndNamespaceForDeliverable(deliverable *v1alpha1.Deliverable, delivery *v1alpha1.ClusterDelivery) (string, string) {
	serviceAccountName := "default"
	serviceAccountNS := deliverable.Namespace

	if deliverable.Spec.ServiceAccountName != "" {
		serviceAccountName = deliverable.Spec.ServiceAccountName
	} else if delivery.Spec.ServiceAccountRef.Name != "" {
		serviceAccountName = delivery.Spec.ServiceAccountRef.Name
		if delivery.Spec.ServiceAccountRef.Namespace != "" {
			serviceAccountNS = delivery.Spec.ServiceAccountRef.Namespace
		}
	}

	return serviceAccountName, serviceAccountNS
}

func (r *DeliverableReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Repo = repository.NewRepository(
		mgr.GetClient(),
		repository.NewCache(mgr.GetLogger().WithName("deliverable-repo-cache")),
	)

	r.ConditionManagerBuilder = conditions.NewConditionManager
	r.ResourceRealizerBuilder = realizer.NewResourceRealizerBuilder(
		repository.NewRepository,
		realizerclient.NewBuilderWithRestMapper(mgr.GetConfig(), mgr.GetClient().RESTMapper()),
		repository.NewCache(mgr.GetLogger().WithName("deliverable-stamping-repo-cache")),
	)
	r.Realizer = realizer.NewRealizer()
	r.DependencyTracker = dependency.NewDependencyTracker(
		2*utils.DefaultResyncTime,
		mgr.GetLogger().WithName("tracker-deliverable"),
	)

	builder := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Deliverable{})

	m := mapper.Mapper{
		Client:  mgr.GetClient(),
		Logger:  mgr.GetLogger().WithName("deliverable"),
		Tracker: r.DependencyTracker,
	}

	watches := map[client.Object]handler.MapFunc{
		&v1alpha1.ClusterDelivery{}:  m.ClusterDeliveryToDeliverableRequests,
		&corev1.ServiceAccount{}:     m.ServiceAccountToDeliverableRequests,
		&rbacv1.Role{}:               m.RoleToDeliverableRequests,
		&rbacv1.RoleBinding{}:        m.RoleBindingToDeliverableRequests,
		&rbacv1.ClusterRole{}:        m.ClusterRoleToDeliverableRequests,
		&rbacv1.ClusterRoleBinding{}: m.ClusterRoleBindingToDeliverableRequests,
	}

	for kindType, mapFunc := range watches {
		builder = builder.Watches(
			&source.Kind{Type: kindType},
			handler.EnqueueRequestsFromMapFunc(mapFunc),
		)
	}

	for _, template := range v1alpha1.ValidDeliveryTemplates {
		builder = builder.Watches(
			&source.Kind{Type: template},
			enqueuer.EnqueueTracked(template, r.DependencyTracker, mgr.GetScheme()),
		)
	}

	controller, err := builder.Build(r)
	if err != nil {
		return fmt.Errorf("failed to build controller for deliverable: %w", err)
	}
	r.StampedTracker = &external.ObjectTracker{Controller: controller}

	return nil
}
