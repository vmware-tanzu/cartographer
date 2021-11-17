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
	"github.com/vmware-tanzu/cartographer/pkg/controller"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
)

type Timer interface {
	Now() metav1.Time
}

type Reconciler struct {
	Repo                    repository.Repository
	ConditionManagerBuilder conditions.ConditionManagerBuilder
	conditionManager        conditions.ConditionManager
	logger                  logr.Logger
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = logr.FromContext(ctx).
		WithValues("name", req.Name, "namespace", req.Namespace)
	r.logger.Info("started")
	defer r.logger.Info("finished")

	supplyChain, err := r.Repo.GetSupplyChain(req.Name)
	if err != nil || supplyChain == nil {
		if kerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("get supply chain: %w", err)
	}

	r.conditionManager = r.ConditionManagerBuilder(v1alpha1.SupplyChainReady, supplyChain.Status.Conditions)

	err = r.reconcileSupplyChain(supplyChain)

	return r.completeReconciliation(supplyChain, err)
}

func (r *Reconciler) completeReconciliation(supplyChain *v1alpha1.ClusterSupplyChain, err error) (ctrl.Result, error) {
	var changed bool
	supplyChain.Status.Conditions, changed = r.conditionManager.Finalize()

	var updateErr error
	if changed || (supplyChain.Status.ObservedGeneration != supplyChain.Generation) {
		supplyChain.Status.ObservedGeneration = supplyChain.Generation
		updateErr = r.Repo.StatusUpdate(supplyChain)
		if updateErr != nil {
			return ctrl.Result{}, fmt.Errorf("update supply chain status: %w", updateErr)
		}
	}

	if err != nil && controller.IsUnhandledError(err) {
		return ctrl.Result{}, err
	}

	r.logger.Info("handled error", "error", err)
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
				return controller.NewUnhandledError(fmt.Errorf("get cluster template: %w", err))
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
