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

package delivery

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
)

type Reconciler struct {
	repo             repository.Repository
	conditionManager conditions.ConditionManager
	logger           logr.Logger
}

func NewReconciler(repo repository.Repository) *Reconciler {
	return &Reconciler{
		repo: repo,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (ctrl.Result, error) {
	r.logger = logr.FromContext(ctx).
		WithValues("name", req.Name, "namespace", req.Namespace)
	r.logger.Info("started")
	defer r.logger.Info("finished")

	delivery, err := r.repo.GetDelivery(req.Name)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("get delivery: %w", err)
	}

	if delivery == nil {
		r.logger.Info("delivery no longer exists")
		return ctrl.Result{}, nil
	}

	r.conditionManager = conditions.NewConditionManager(v1alpha1.DeliveryReady, delivery.Status.Conditions)

	err = r.reconcileDelivery(delivery)

	return r.completeReconciliation(delivery, err)
}

func (r *Reconciler) reconcileDelivery(delivery *v1alpha1.ClusterDelivery) error {
	var (
		err               error
		resourcesNotFound []string
	)

	for _, resource := range delivery.Spec.Resources {
		_, err = r.repo.GetDeliveryClusterTemplate(resource.TemplateRef)
		if err != nil {
			if !kerrors.IsNotFound(err) {
				return err
			}

			resourcesNotFound = append(resourcesNotFound, resource.Name)
		}
	}

	if len(resourcesNotFound) > 0 {
		r.conditionManager.AddPositive(TemplatesNotFoundCondition(resourcesNotFound))
	} else {
		r.conditionManager.AddPositive(TemplatesFoundCondition())
	}

	return nil
}

func (r *Reconciler) completeReconciliation(delivery *v1alpha1.ClusterDelivery, reconcileError error) (ctrl.Result, error) {
	delivery.Status.Conditions, _ = r.conditionManager.Finalize()

	delivery.Status.ObservedGeneration = delivery.Generation
	err := r.repo.StatusUpdate(delivery)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("status update: %w", err)
	}

	if reconcileError != nil {
		return ctrl.Result{}, reconcileError
	}

	return ctrl.Result{}, nil
}
