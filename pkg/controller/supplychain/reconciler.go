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

package supplychain

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
)

type Timer interface {
	Now() metav1.Time
}

type Reconciler struct {
	Repo                    repository.Repository
	ConditionManagerBuilder conditions.ConditionManagerBuilder
	conditionManager        conditions.ConditionManager
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logr.FromContext(ctx).
		WithValues("name", req.Name, "namespace", req.Namespace)
	logger.Info("started")

	reconcileCtx := logr.NewContext(ctx, logger)

	sc, err := r.Repo.GetSupplyChain(req.Name)
	if err != nil || sc == nil {
		if kerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("get supplyChain: %w", err)
	}

	// fixme: discuss DeepCopy() as a prophylactic
	supplyChain := sc.DeepCopy()

	r.conditionManager = r.ConditionManagerBuilder(v1alpha1.SupplyChainReady, supplyChain.Status.Conditions)

	err = r.reconcileSupplyChain(supplyChain)

	return r.completeReconciliation(reconcileCtx, supplyChain, err)
}

func (r *Reconciler) completeReconciliation(ctx context.Context, supplyChain *v1alpha1.ClusterSupplyChain, err error) (ctrl.Result, error) {
	logger := logr.FromContext(ctx)

	var changed bool
	supplyChain.Status.Conditions, changed = r.conditionManager.Finalize()

	var updateErr error
	if changed || (supplyChain.Status.ObservedGeneration != supplyChain.Generation) {
		supplyChain.Status.ObservedGeneration = supplyChain.Generation
		updateErr = r.Repo.StatusUpdate(supplyChain)
		if updateErr != nil {
			logger.Error(updateErr, "update error")
			if err == nil {
				logger.Info("finished")
				return ctrl.Result{}, fmt.Errorf("update supply-chain status: %w", updateErr)
			}
		}
	}

	logger.Info("finished")
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) reconcileSupplyChain(chain *v1alpha1.ClusterSupplyChain) error {
	var (
		err               error
		resourcesNotFound []string
	)

	for _, resource := range chain.Spec.Resources {
		_, err = r.Repo.GetClusterTemplate(resource.TemplateRef)
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
