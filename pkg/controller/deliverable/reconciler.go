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

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/deliverable"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

type Reconciler interface {
	Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error)
	AddTracking(dynamicTracker DynamicTracker)
}

type reconciler struct {
	repo                    repository.Repository
	conditionManager        conditions.ConditionManager
	conditionManagerBuilder conditions.ConditionManagerBuilder
	realizer                realizer.Realizer
	logger                  logr.Logger
	dynamicTracker          DynamicTracker
}

func NewReconciler(repo repository.Repository, conditionManagerBuilder conditions.ConditionManagerBuilder, realizer realizer.Realizer) Reconciler {
	return &reconciler{
		repo:                    repo,
		conditionManagerBuilder: conditionManagerBuilder,
		realizer:                realizer,
	}
}

//counterfeiter:generate . DynamicTracker
type DynamicTracker interface {
	Watch(log logr.Logger, obj runtime.Object, handler handler.EventHandler) error
}

func (r *reconciler) AddTracking(dynamicTracker DynamicTracker) {
	r.dynamicTracker = dynamicTracker
}

func (r *reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = logr.FromContext(ctx).
		WithValues("name", req.Name, "namespace", req.Namespace)
	r.logger.Info("started")
	defer r.logger.Info("finished")

	deliverable, err := r.repo.GetDeliverable(req.Name, req.Namespace)
	if err != nil || deliverable == nil {
		if kerrors.IsNotFound(err) {
			// 1. Does not exist, we watch deliverables, do not requeue
			return ctrl.Result{}, nil
		}

		// 2. Server error, we should requeue
		return ctrl.Result{}, fmt.Errorf("get deliverable: %w", err)
	}

	r.conditionManager = r.conditionManagerBuilder(v1alpha1.DeliverableReady, deliverable.Status.Conditions)

	delivery, err := r.getDeliveriesForDeliverable(deliverable)
	if err != nil {
		// 3.a len(deliverable.Labels) == 0, we watch deliverables, do not requeue
		// 3.b GetDeliveriesForDeliverable, server error, we should requeue
		// 3.c len(deliveries) == 0, we watch deliveries, do not requeue
		//       SHOULD THIS BE A STATUS?
		// 3.d len(deliveries) > 1, we watch deliveries, do not requeue
		//       SHOULD THIS BE A STATUS?
		return r.completeReconciliation(deliverable, err)
	}

	deliveryGVK, err := utils.GetObjectGVK(delivery, r.repo.GetScheme())
	if err != nil {
		// 4. This would be a real error, I'm not sure how it would ever be fixed tho?, we should requeue
		return r.completeReconciliation(deliverable, fmt.Errorf("get object gvk: %w", err))
	}

	deliverable.Status.DeliveryRef.Kind = deliveryGVK.Kind
	deliverable.Status.DeliveryRef.Name = delivery.Name

	err = r.checkDeliveryReadiness(delivery)
	if err != nil {
		// 5. If readyCondition.Status != "True", we watch deliveries, do not requeue, should not even be an error
		r.conditionManager.AddPositive(MissingReadyInDeliveryCondition(getDeliveryReadyCondition(delivery)))
		return r.completeReconciliation(deliverable, err)
	}
	r.conditionManager.AddPositive(DeliveryReadyCondition())

	stampedObjects, err := r.realizer.Realize(ctx, realizer.NewResourceRealizer(deliverable, r.repo), delivery)
	if err != nil {
		switch typedErr := err.(type) {
		case realizer.GetDeliveryClusterTemplateError:
			// 6.a invalid kind, we watch templates, do not requeue
			// 6.b get object, server error, requeue
			// 6.c impossible to get here??, do not requeue
			r.conditionManager.AddPositive(TemplateObjectRetrievalFailureCondition(typedErr))
		case realizer.StampError:
			// 7.a. unknwon resource template type, we watch templates, do not requeue
			// 7.b. templatize, we watch templates, deliverables, deliveries, do not requeue
			r.conditionManager.AddPositive(TemplateStampFailureCondition(typedErr))
		case realizer.ApplyStampedObjectError:
			// 8.a list - server error, requeue
			// 8.b patch - server error, requeue
			// 8.c create - server error, requeue
			// 8.d invalid unstructured..... do not requeue, FUTURE
			r.conditionManager.AddPositive(TemplateRejectedByAPIServerCondition(typedErr))
		case realizer.RetrieveOutputError:
			// 9.a evaluate json path - ???, do not requeue
			r.conditionManager.AddPositive(MissingValueAtPathCondition(typedErr.ResourceName(), typedErr.JsonPathExpression()))
		default:
			// 10. ?????????????????????, requeue
			r.conditionManager.AddPositive(UnknownResourceErrorCondition(typedErr))
		}
	} else {
		r.conditionManager.AddPositive(ResourcesSubmittedCondition())
	}

	if len(stampedObjects) > 0 {
		for _, stampedObject := range stampedObjects {
			err = r.dynamicTracker.Watch(r.logger, stampedObject, &handler.EnqueueRequestForOwner{OwnerType: &v1alpha1.Deliverable{}})
			if err != nil {
				r.logger.Error(err, "dynamic tracker watch")
			}
		}
	}

	return r.completeReconciliation(deliverable, nil)
}

func (r *reconciler) completeReconciliation(deliverable *v1alpha1.Deliverable, err error) (ctrl.Result, error) {
	var changed bool
	deliverable.Status.Conditions, changed = r.conditionManager.Finalize()

	var updateErr error
	if changed || (deliverable.Status.ObservedGeneration != deliverable.Generation) {
		deliverable.Status.ObservedGeneration = deliverable.Generation
		updateErr = r.repo.StatusUpdate(deliverable)
		if updateErr != nil {
			r.logger.Error(updateErr, "update error")
			if err == nil {
				return ctrl.Result{}, fmt.Errorf("update deliverable status: %w", updateErr)
			}
		}
	}

	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *reconciler) checkDeliveryReadiness(delivery *v1alpha1.ClusterDelivery) error {
	readyCondition := getDeliveryReadyCondition(delivery)
	if readyCondition.Status == "True" {
		return nil
	}
	return fmt.Errorf("delivery is not in ready condition")
}

func getDeliveryReadyCondition(delivery *v1alpha1.ClusterDelivery) metav1.Condition {
	for _, condition := range delivery.Status.Conditions {
		if condition.Type == "Ready" {
			return condition
		}
	}
	return metav1.Condition{}
}

func (r *reconciler) getDeliveriesForDeliverable(deliverable *v1alpha1.Deliverable) (*v1alpha1.ClusterDelivery, error) {
	if len(deliverable.Labels) == 0 {
		r.conditionManager.AddPositive(DeliverableMissingLabelsCondition())
		return nil, fmt.Errorf("deliverable is missing required labels")
	}

	deliveries, err := r.repo.GetDeliveriesForDeliverable(deliverable)
	if err != nil {
		r.conditionManager.AddPositive(DeliveryNotFoundCondition(deliverable.Labels))
		return nil, fmt.Errorf("get delivery by label: %w", err)
	}

	if len(deliveries) == 0 {
		r.conditionManager.AddPositive(DeliveryNotFoundCondition(deliverable.Labels))
		return nil, fmt.Errorf("no delivery found where full selector is satisfied by labels: %v", deliverable.Labels)
	}

	if len(deliveries) > 1 {
		r.conditionManager.AddPositive(TooManyDeliveryMatchesCondition())
		return nil, fmt.Errorf("too many deliveries match the deliverable selector label")
	}

	return &deliveries[0], nil
}
