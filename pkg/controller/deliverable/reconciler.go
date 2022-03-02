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

package deliverable

//go:generate go run -modfile ../../../hack/tools/go.mod github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/controller"
	"github.com/vmware-tanzu/cartographer/pkg/logger"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/deliverable"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/dependency"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/stamped"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

type Reconciler struct {
	Repo                    repository.Repository
	ConditionManagerBuilder conditions.ConditionManagerBuilder
	ResourceRealizerBuilder realizer.ResourceRealizerBuilder
	Realizer                realizer.Realizer
	StampedTracker          stamped.StampedTracker
	DependencyTracker       dependency.DependencyTracker
	conditionManager        conditions.ConditionManager
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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

	r.conditionManager = r.ConditionManagerBuilder(v1alpha1.DeliverableReady, deliverable.Status.Conditions)

	delivery, err := r.getDeliveriesForDeliverable(ctx, deliverable)
	if err != nil {
		return r.completeReconciliation(ctx, deliverable, deliverable.Status.Resources, err)
	}

	log = log.WithValues("delivery", delivery.Name)
	ctx = logr.NewContext(ctx, log)

	deliveryGVK, err := utils.GetObjectGVK(delivery, r.Repo.GetScheme())
	if err != nil {
		log.Error(err, "failed to get object gvk for delivery")
		return r.completeReconciliation(ctx, deliverable, deliverable.Status.Resources, controller.NewUnhandledError(
			fmt.Errorf("failed to get object gvk for delivery [%s]: %w", delivery.Name, err)))
	}

	deliverable.Status.DeliveryRef.Kind = deliveryGVK.Kind
	deliverable.Status.DeliveryRef.Name = delivery.Name

	if !r.isDeliveryReady(delivery) {
		r.conditionManager.AddPositive(MissingReadyInDeliveryCondition(getDeliveryReadyCondition(delivery)))
		log.Info("delivery is not in ready state")
		return r.completeReconciliation(ctx, deliverable, deliverable.Status.Resources, fmt.Errorf("delivery [%s] is not in ready state", delivery.Name))
	}
	r.conditionManager.AddPositive(DeliveryReadyCondition())

	serviceAccountName, serviceAccountNS := getServiceAccountNameAndNamespace(deliverable, delivery)

	secret, err := r.Repo.GetServiceAccountSecret(ctx, serviceAccountName, serviceAccountNS)
	if err != nil {
		r.conditionManager.AddPositive(ServiceAccountSecretNotFoundCondition(err))
		return r.completeReconciliation(ctx, deliverable, deliverable.Status.Resources, fmt.Errorf("failed to get secret for service account [%s]: %w", fmt.Sprintf("%s/%s", serviceAccountNS, serviceAccountName), err))
	}

	resourceRealizer, err := r.ResourceRealizerBuilder(secret, deliverable, r.Repo, delivery.Spec.Params)
	if err != nil {
		r.conditionManager.AddPositive(ResourceRealizerBuilderErrorCondition(err))
		return r.completeReconciliation(ctx, deliverable, deliverable.Status.Resources, controller.NewUnhandledError(fmt.Errorf("failed to build resource realizer: %w", err)))
	}

	realizedResources, err := r.Realizer.Realize(ctx, resourceRealizer, delivery, deliverable.Status.Resources)
	if err != nil {
		log.V(logger.DEBUG).Info("failed to realize")
		switch typedErr := err.(type) {
		case realizer.GetTemplateError:
			r.conditionManager.AddPositive(TemplateObjectRetrievalFailureCondition(typedErr))
			err = controller.NewUnhandledError(err)
		case realizer.StampError:
			r.conditionManager.AddPositive(TemplateStampFailureCondition(typedErr))
		case realizer.ApplyStampedObjectError:
			r.conditionManager.AddPositive(TemplateRejectedByAPIServerCondition(typedErr))
			if !kerrors.IsForbidden(typedErr.Err) {
				err = controller.NewUnhandledError(err)
			}
		case realizer.RetrieveOutputError:
			switch typedErr.Err.(type) {
			case templates.ObservedGenerationError:
				r.conditionManager.AddPositive(TemplateStampFailureByObservedGenerationCondition(typedErr))
			case templates.DeploymentFailedConditionMetError:
				r.conditionManager.AddPositive(DeploymentFailedConditionMetCondition(typedErr))
			case templates.DeploymentConditionError:
				r.conditionManager.AddPositive(DeploymentConditionNotMetCondition(typedErr))
			case templates.JsonPathError:
				r.conditionManager.AddPositive(MissingValueAtPathCondition(typedErr.StampedObject, typedErr.JsonPathExpression()))
			default:
				r.conditionManager.AddPositive(UnknownResourceErrorCondition(typedErr))
			}
		case realizer.ResolveTemplateOptionError:
			r.conditionManager.AddPositive(ResolveTemplateOptionsErrorCondition(typedErr))
		case realizer.TemplateOptionsMatchError:
			r.conditionManager.AddPositive(TemplateOptionsMatchErrorCondition(typedErr))
		default:
			r.conditionManager.AddPositive(UnknownResourceErrorCondition(typedErr))
			err = controller.NewUnhandledError(err)
		}
	} else {
		if log.V(logger.DEBUG).Enabled() {
			for _, resource := range realizedResources {
				log.V(logger.DEBUG).Info("realized object",
					"object", resource.StampedRef)
			}
		}
		r.conditionManager.AddPositive(ResourcesSubmittedCondition())
	}

	r.trackDependencies(deliverable, realizedResources, serviceAccountName, serviceAccountNS)

	cleanupErr := r.cleanupOrphanedObjects(ctx, deliverable.Status.Resources, realizedResources)
	if cleanupErr != nil {
		log.Error(cleanupErr, "failed to cleanup orphaned objects")
	}

	var trackingError error
	for _, resource := range realizedResources {
		if resource.StampedRef == nil {
			continue
		}
		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(resource.StampedRef.GroupVersionKind())

		trackingError = r.StampedTracker.Watch(log, obj, &handler.EnqueueRequestForOwner{OwnerType: &v1alpha1.Deliverable{}})
		if trackingError != nil {
			log.Error(err, "failed to add informer for object",
				"object", resource.StampedRef)
			err = controller.NewUnhandledError(trackingError)
		} else {
			log.V(logger.DEBUG).Info("added informer for object",
				"object", resource.StampedRef)
		}
	}

	return r.completeReconciliation(ctx, deliverable, realizedResources, err)
}

func (r *Reconciler) completeReconciliation(ctx context.Context, deliverable *v1alpha1.Deliverable, realizedResources []v1alpha1.RealizedResource, err error) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx)
	var changed bool
	deliverable.Status.Conditions, changed = r.conditionManager.Finalize()

	var updateErr error
	if changed || (deliverable.Status.ObservedGeneration != deliverable.Generation) || !reflect.DeepEqual(deliverable.Status.Resources, realizedResources) {
		deliverable.Status.Resources = realizedResources
		deliverable.Status.ObservedGeneration = deliverable.Generation
		updateErr = r.Repo.StatusUpdate(ctx, deliverable)
		if updateErr != nil {
			log.Error(err, "failed to update status for deliverable")
			return ctrl.Result{}, fmt.Errorf("failed to update status for deliverable: %w", updateErr)
		}
	}

	if err != nil {
		if controller.IsUnhandledError(err) {
			log.Error(err, "unhandled error reconciling deliverable")
			return ctrl.Result{}, err
		}
		log.Info("handled error reconciling deliverable", "handled error", err)
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) isDeliveryReady(delivery *v1alpha1.ClusterDelivery) bool {
	readyCondition := getDeliveryReadyCondition(delivery)
	return readyCondition.Status == "True"
}

func (r *Reconciler) trackDependencies(deliverable *v1alpha1.Deliverable, realizedResources []v1alpha1.RealizedResource, serviceAccountName, serviceAccountNS string) {
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

func (r *Reconciler) cleanupOrphanedObjects(ctx context.Context, previousResources, realizedResources []v1alpha1.RealizedResource) error {
	log := logr.FromContextOrDiscard(ctx)

	var orphanedObjs []*corev1.ObjectReference
	for _, prevResource := range previousResources {
		orphaned := true
		for _, realizedResource := range realizedResources {
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

func (r *Reconciler) getDeliveriesForDeliverable(ctx context.Context, deliverable *v1alpha1.Deliverable) (*v1alpha1.ClusterDelivery, error) {
	log := logr.FromContextOrDiscard(ctx)
	if len(deliverable.Labels) == 0 {
		r.conditionManager.AddPositive(DeliverableMissingLabelsCondition())
		log.Info("deliverable is missing required labels")
		return nil, fmt.Errorf("deliverable [%s/%s] is missing required labels",
			deliverable.Namespace, deliverable.Name)
	}

	deliveries, err := r.Repo.GetDeliveriesForDeliverable(ctx, deliverable)
	if err != nil {
		log.Error(err, "failed to get deliveries for deliverable")
		return nil, controller.NewUnhandledError(fmt.Errorf("failed to get deliveries for deliverable [%s/%s]: %w",
			deliverable.Namespace, deliverable.Name, err))
	}

	if len(deliveries) == 0 {
		r.conditionManager.AddPositive(DeliveryNotFoundCondition(deliverable.Labels))
		log.Info("no delivery found where full selector is satisfied by label",
			"labels", deliverable.Labels)
		return nil, fmt.Errorf("no delivery [%s/%s] found where full selector is satisfied by labels: %v",
			deliverable.Namespace, deliverable.Name, deliverable.Labels)
	}

	if len(deliveries) > 1 {
		r.conditionManager.AddPositive(TooManyDeliveryMatchesCondition())
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

func getServiceAccountNameAndNamespace(deliverable *v1alpha1.Deliverable, delivery *v1alpha1.ClusterDelivery) (string, string) {
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
