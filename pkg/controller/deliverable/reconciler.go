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

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/deliverable"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

const reconcileInterval = 5 * time.Second

type Reconciler struct {
	repo                    repository.Repository
	conditionManager        conditions.ConditionManager
	conditionManagerBuilder conditions.ConditionManagerBuilder
	realizer                realizer.Realizer
	logger                  logr.Logger
}

func NewReconciler(repo repository.Repository, conditionManagerBuilder conditions.ConditionManagerBuilder, realizer realizer.Realizer) *Reconciler {
	return &Reconciler{
		repo:                    repo,
		conditionManagerBuilder: conditionManagerBuilder,
		realizer:                realizer,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = logr.FromContext(ctx).
		WithValues("name", req.Name, "namespace", req.Namespace)
	r.logger.Info("started")
	defer r.logger.Info("finished")

	deliverable, err := r.repo.GetDeliverable(req.Name, req.Namespace)
	if err != nil || deliverable == nil {
		if kerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("get deliverable: %w", err)
	}

	r.conditionManager = r.conditionManagerBuilder(v1alpha1.DeliverableReady, deliverable.Status.Conditions)

	delivery, err := r.getDeliveriesForDeliverable(deliverable)
	if err != nil {
		return r.completeReconciliation(deliverable, err)
	}

	deliveryGVK, err := utils.GetObjectGVK(delivery, r.repo.GetScheme())
	if err != nil {
		return r.completeReconciliation(deliverable, fmt.Errorf("get object gvk: %w", err))
	}

	deliverable.Status.DeliveryRef.Kind = deliveryGVK.Kind
	deliverable.Status.DeliveryRef.Name = delivery.Name

	err = r.checkDeliveryReadiness(delivery)
	if err != nil {
		r.conditionManager.AddPositive(MissingReadyInDeliveryCondition(getDeliveryReadyCondition(delivery)))
		return r.completeReconciliation(deliverable, err)
	}
	r.conditionManager.AddPositive(DeliveryReadyCondition())

	err = r.realizer.Realize(ctx, realizer.NewResourceRealizer(deliverable, r.repo), delivery)
	if err != nil {
		switch typedErr := err.(type) {
		case realizer.GetDeliveryClusterTemplateError:
			r.conditionManager.AddPositive(TemplateObjectRetrievalFailureCondition(typedErr))
		case realizer.StampError:
			r.conditionManager.AddPositive(TemplateStampFailureCondition(typedErr))
		case realizer.ApplyStampedObjectError:
			r.conditionManager.AddPositive(TemplateRejectedByAPIServerCondition(typedErr))
		case realizer.RetrieveOutputError:
			r.conditionManager.AddPositive(MissingValueAtPathCondition(typedErr.ComponentName(), typedErr.JsonPathExpression()))
			err = nil
		default:
			r.conditionManager.AddPositive(UnknownResourceErrorCondition(typedErr))
		}

		return r.completeReconciliation(deliverable, err)
	}

	r.conditionManager.AddPositive(ResourcesSubmittedCondition())

	return r.completeReconciliation(deliverable, nil)
}

func (r *Reconciler) completeReconciliation(deliverable *v1alpha1.Deliverable, err error) (ctrl.Result, error) {
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

	if !r.conditionManager.IsSuccessful() { // TODO: Discuss rename to IsReady
		return ctrl.Result{}, fmt.Errorf("deliverable not ready")
	}

	return ctrl.Result{RequeueAfter: reconcileInterval}, nil
}

func (r *Reconciler) checkDeliveryReadiness(delivery *v1alpha1.ClusterDelivery) error {
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

func (r *Reconciler) getDeliveriesForDeliverable(deliverable *v1alpha1.Deliverable) (*v1alpha1.ClusterDelivery, error) {
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
