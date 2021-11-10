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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/deliverable"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/tracker"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

type Reconciler struct {
	Repo                    repository.Repository
	ConditionManagerBuilder conditions.ConditionManagerBuilder
	Realizer                realizer.Realizer
	DynamicTracker          tracker.DynamicTracker
	conditionManager        conditions.ConditionManager
	logger                  logr.Logger
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = logr.FromContext(ctx).
		WithValues("name", req.Name, "namespace", req.Namespace)
	r.logger.Info("started")
	defer r.logger.Info("finished")

	deliverable, err := r.Repo.GetDeliverable(req.Name, req.Namespace)
	if err != nil || deliverable == nil {
		if kerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("get deliverable: %w", err)
	}

	r.conditionManager = r.ConditionManagerBuilder(v1alpha1.DeliverableReady, deliverable.Status.Conditions)

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

	deliveryGVK, err := utils.GetObjectGVK(delivery, r.Repo.GetScheme())
	if err != nil {
		return r.completeReconciliation(deliverable, controller.NewUnhandledError(fmt.Errorf("get object gvk: %w", err)))
	}

	deliverable.Status.DeliveryRef.Kind = deliveryGVK.Kind
	deliverable.Status.DeliveryRef.Name = delivery.Name

	if !r.isDeliveryReady(delivery) {
		r.conditionManager.AddPositive(MissingReadyInDeliveryCondition(getDeliveryReadyCondition(delivery)))
		return r.completeReconciliation(deliverable, nil)
	}
	r.conditionManager.AddPositive(DeliveryReadyCondition())

	stampedObjects, err := r.Realizer.Realize(ctx, realizer.NewResourceRealizer(deliverable, r.Repo), delivery)
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
			err = r.DynamicTracker.Watch(r.logger, stampedObject, &handler.EnqueueRequestForOwner{OwnerType: &v1alpha1.Deliverable{}})
			if err != nil {
				r.logger.Error(err, "dynamic tracker watch")
			}
		}
	}

	return r.completeReconciliation(deliverable, nil)
}

func (r *Reconciler) completeReconciliation(deliverable *v1alpha1.Deliverable, err error) (ctrl.Result, error) {
	var changed bool
	deliverable.Status.Conditions, changed = r.conditionManager.Finalize()

	var updateErr error
	if changed || (deliverable.Status.ObservedGeneration != deliverable.Generation) {
		deliverable.Status.ObservedGeneration = deliverable.Generation
		updateErr = r.Repo.StatusUpdate(deliverable)
		if updateErr != nil {
			r.logger.Error(updateErr, "update error")
			if err == nil {
				return ctrl.Result{}, fmt.Errorf("update deliverable status: %w", updateErr)
			}
		}
	}

	// log handled errors?

	if err != nil && controller.IsUnhandledError(err) {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) isDeliveryReady(delivery *v1alpha1.ClusterDelivery) bool {
	readyCondition := getDeliveryReadyCondition(delivery)
	return readyCondition.Status == "True"
}

func getDeliveryReadyCondition(delivery *v1alpha1.ClusterDelivery) metav1.Condition {
	for _, condition := range delivery.Status.Conditions {
		if condition.Type == "Ready" {
			return condition
		}
	}
	return metav1.Condition{}
}

func (r *Reconciler) getDeliveriesForDeliverable(deliverable *v1alpha1.Deliverable) (*v1alpha1.ClusterDelivery, error) {
	if len(deliverable.Labels) == 0 {
		r.conditionManager.AddPositive(DeliverableMissingLabelsCondition())
		return nil, fmt.Errorf("deliverable is missing required labels")
	}

	deliveries, err := r.Repo.GetDeliveriesForDeliverable(deliverable)
	if err != nil {
		//return nil, controller.NewUnhandledError(fmt.Errorf("get delivery by label: %w", err))
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
