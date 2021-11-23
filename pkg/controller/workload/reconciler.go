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

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/controller"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/workload"
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
	r.logger = logr.FromContext(ctx)
	r.logger.Info("started")
	defer r.logger.Info("finished")

	workload, err := r.Repo.GetWorkload(ctx, req.Name, req.Namespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("get workload: %w", err)
	}

	if workload == nil {
		r.logger.Info("workload no longer exists")
		return ctrl.Result{}, nil
	}

	r.conditionManager = r.ConditionManagerBuilder(v1alpha1.WorkloadReady, workload.Status.Conditions)

	supplyChain, err := r.getSupplyChainsForWorkload(ctx, workload)
	if err != nil {
		return r.completeReconciliation(ctx, workload, err)
	}

	supplyChainGVK, err := utils.GetObjectGVK(supplyChain, r.Repo.GetScheme())
	if err != nil {
		return r.completeReconciliation(ctx, workload, controller.NewUnhandledError(fmt.Errorf("get object gvk: %w", err)))
	}

	workload.Status.SupplyChainRef.Kind = supplyChainGVK.Kind
	workload.Status.SupplyChainRef.Name = supplyChain.Name

	if !r.isSupplyChainReady(supplyChain) {
		r.conditionManager.AddPositive(MissingReadyInSupplyChainCondition(getSupplyChainReadyCondition(supplyChain)))
		return r.completeReconciliation(ctx, workload, fmt.Errorf("supply chain is not in ready state"))
	}
	r.conditionManager.AddPositive(SupplyChainReadyCondition())

	stampedObjects, err := r.Realizer.Realize(ctx, realizer.NewResourceRealizer(workload, r.Repo), supplyChain)
	if err != nil {
		switch typedErr := err.(type) {
		case realizer.GetClusterTemplateError:
			r.conditionManager.AddPositive(TemplateObjectRetrievalFailureCondition(typedErr))
			err = controller.NewUnhandledError(err)
		case realizer.StampError:
			r.conditionManager.AddPositive(TemplateStampFailureCondition(typedErr))
		case realizer.ApplyStampedObjectError:
			r.conditionManager.AddPositive(TemplateRejectedByAPIServerCondition(typedErr))
			err = controller.NewUnhandledError(err)
		case realizer.RetrieveOutputError:
			r.conditionManager.AddPositive(MissingValueAtPathCondition(typedErr.Resource.Name, typedErr.JsonPathExpression()))
		default:
			r.conditionManager.AddPositive(UnknownResourceErrorCondition(typedErr))
			err = controller.NewUnhandledError(err)
		}
	} else {
		r.conditionManager.AddPositive(ResourcesSubmittedCondition())
	}

	var trackingError error
	if len(stampedObjects) > 0 {
		for _, stampedObject := range stampedObjects {
			trackingError = r.DynamicTracker.Watch(r.logger, stampedObject, &handler.EnqueueRequestForOwner{OwnerType: &v1alpha1.Workload{}})
			if trackingError != nil {
				r.logger.Error(err, "dynamic tracker watch")
				err = controller.NewUnhandledError(trackingError)
			}
		}
	}

	return r.completeReconciliation(ctx, workload, err)
}

func (r *Reconciler) completeReconciliation(ctx context.Context, workload *v1alpha1.Workload, err error) (ctrl.Result, error) {
	var changed bool
	workload.Status.Conditions, changed = r.conditionManager.Finalize()

	var updateErr error
	if changed || (workload.Status.ObservedGeneration != workload.Generation) {
		workload.Status.ObservedGeneration = workload.Generation
		updateErr = r.Repo.StatusUpdate(ctx, workload)
		if updateErr != nil {
			return ctrl.Result{}, fmt.Errorf("update workload status: %w", updateErr)
		}
	}

	if err != nil {
		if controller.IsUnhandledError(err) {
			return ctrl.Result{}, err
		}
		r.logger.Info("handled error", "error", err)
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) isSupplyChainReady(supplyChain *v1alpha1.ClusterSupplyChain) bool {
	supplyChainReadyCondition := getSupplyChainReadyCondition(supplyChain)
	return supplyChainReadyCondition.Status == "True"
}

func getSupplyChainReadyCondition(supplyChain *v1alpha1.ClusterSupplyChain) metav1.Condition {
	for _, condition := range supplyChain.Status.Conditions {
		if condition.Type == "Ready" {
			return condition
		}
	}
	return metav1.Condition{}
}

func (r *Reconciler) getSupplyChainsForWorkload(ctx context.Context, workload *v1alpha1.Workload) (*v1alpha1.ClusterSupplyChain, error) {
	if len(workload.Labels) == 0 {
		r.conditionManager.AddPositive(WorkloadMissingLabelsCondition())
		return nil, fmt.Errorf("workload is missing required labels")
	}

	supplyChains, err := r.Repo.GetSupplyChainsForWorkload(ctx, workload)
	if err != nil {
		return nil, controller.NewUnhandledError(fmt.Errorf("get supply chain for workload: %w", err))
	}

	if len(supplyChains) == 0 {
		r.conditionManager.AddPositive(SupplyChainNotFoundCondition(workload.Labels))
		return nil, fmt.Errorf("no supply chain found where full selector is satisfied by labels: %v", workload.Labels)
	}

	if len(supplyChains) > 1 {
		r.conditionManager.AddPositive(TooManySupplyChainMatchesCondition())
		return nil, fmt.Errorf("too many supply chains match the workload selector")
	}

	return supplyChains[0], nil
}
