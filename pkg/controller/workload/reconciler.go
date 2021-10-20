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

package workload

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
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/workload"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

const reconcileInterval = 5 * time.Second

type Reconciler struct {
	repo                    repository.Repository
	conditionManager        conditions.ConditionManager
	conditionManagerBuilder conditions.ConditionManagerBuilder
	realizer                realizer.Realizer
}

func NewReconciler(repo repository.Repository, conditionManagerBuilder conditions.ConditionManagerBuilder, realizer realizer.Realizer) *Reconciler {
	return &Reconciler{
		repo:                    repo,
		conditionManagerBuilder: conditionManagerBuilder,
		realizer:                realizer,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logr.FromContext(ctx).
		WithValues("name", req.Name, "namespace", req.Namespace)
	ctx = logr.NewContext(ctx, logger)
	logger.Info("started")

	reconcileCtx := logr.NewContext(ctx, logger)

	workload, err := r.repo.GetWorkload(req.Name, req.Namespace)
	if err != nil || workload == nil {
		if kerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("get workload: %w", err)
	}

	r.conditionManager = r.conditionManagerBuilder(v1alpha1.WorkloadReady, workload.Status.Conditions)

	supplyChain, err := r.getSupplyChainsForWorkload(workload)
	if err != nil {
		return r.completeReconciliation(reconcileCtx, workload, err)
	}

	supplyChainGVK, err := utils.GetObjectGVK(supplyChain, r.repo.GetScheme())
	if err != nil {
		return r.completeReconciliation(reconcileCtx, workload, fmt.Errorf("get object gvk: %w", err))
	}

	workload.Status.SupplyChainRef.Kind = supplyChainGVK.Kind
	workload.Status.SupplyChainRef.Name = supplyChain.Name

	err = r.checkSupplyChainReadiness(supplyChain)
	if err != nil {
		r.conditionManager.AddPositive(MissingReadyInSupplyChainCondition(getSupplyChainReadyCondition(supplyChain)))
		return r.completeReconciliation(reconcileCtx, workload, err)
	}
	r.conditionManager.AddPositive(SupplyChainReadyCondition())

	err = r.realizer.Realize(ctx, realizer.NewComponentRealizer(workload, r.repo), supplyChain)
	if err != nil {
		switch typedErr := err.(type) {
		case realizer.GetClusterTemplateError:
			r.conditionManager.AddPositive(TemplateObjectRetrievalFailureCondition(typedErr))
		case realizer.StampError:
			r.conditionManager.AddPositive(TemplateStampFailureCondition(typedErr))
		case realizer.ApplyStampedObjectError:
			r.conditionManager.AddPositive(TemplateRejectedByAPIServerCondition(typedErr))
		case realizer.RetrieveOutputError:
			r.conditionManager.AddPositive(MissingValueAtPathCondition(typedErr.ComponentName(), typedErr.JsonPathExpression()))
			err = nil
		default:
			r.conditionManager.AddPositive(UnknownComponentErrorCondition(typedErr))
		}

		return r.completeReconciliation(reconcileCtx, workload, err)
	}

	r.conditionManager.AddPositive(ComponentsSubmittedCondition())

	return r.completeReconciliation(reconcileCtx, workload, nil)
}

func (r *Reconciler) completeReconciliation(ctx context.Context, workload *v1alpha1.Workload, err error) (ctrl.Result, error) {
	logger := logr.FromContext(ctx)

	workload.Status.Conditions = r.conditionManager.Finalize()

	workload.Status.ObservedGeneration = workload.Generation
	updateErr := r.repo.StatusUpdate(workload)
	if updateErr != nil {
		logger.Error(updateErr, "update error")
		if err == nil {
			logger.Info("finished")
			return ctrl.Result{}, fmt.Errorf("update workload status: %w", updateErr)
		}
	}

	logger.Info("finished")

	if err != nil {
		return ctrl.Result{}, err
	}

	if !r.conditionManager.IsSuccessful() { // TODO: Discuss rename to IsReady
		return ctrl.Result{}, fmt.Errorf("workload not ready")
	}

	return ctrl.Result{RequeueAfter: reconcileInterval}, nil
}

func (r *Reconciler) checkSupplyChainReadiness(supplyChain *v1alpha1.ClusterSupplyChain) error {
	supplyChainReadyCondition := getSupplyChainReadyCondition(supplyChain)
	if supplyChainReadyCondition.Status == "True" {
		return nil
	}
	return fmt.Errorf("supply-chain is not in ready condition")
}

func getSupplyChainReadyCondition(supplyChain *v1alpha1.ClusterSupplyChain) metav1.Condition {
	for _, condition := range supplyChain.Status.Conditions {
		if condition.Type == "Ready" {
			return condition
		}
	}
	return metav1.Condition{}
}

func (r *Reconciler) getSupplyChainsForWorkload(workload *v1alpha1.Workload) (*v1alpha1.ClusterSupplyChain, error) {
	if len(workload.Labels) == 0 {
		r.conditionManager.AddPositive(WorkloadMissingLabelsCondition())
		return nil, fmt.Errorf("workload is missing required labels")
	}

	supplyChains, err := r.repo.GetSupplyChainsForWorkload(workload)
	if err != nil || len(supplyChains) == 0 {
		r.conditionManager.AddPositive(SupplyChainNotFoundCondition(workload.Labels))

		if err != nil {
			return nil, fmt.Errorf("get supply chain by label: %w", err)
		} else {
			return nil, fmt.Errorf("no supply chain found where full selector is satisfied by labels: %v", workload.Labels)
		}
	} else if len(supplyChains) > 1 {
		r.conditionManager.AddPositive(TooManySupplyChainMatchesCondition())
		return nil, fmt.Errorf("too many supply chains match the workload selector")
	}

	return supplyChains[0].DeepCopy(), nil
}
