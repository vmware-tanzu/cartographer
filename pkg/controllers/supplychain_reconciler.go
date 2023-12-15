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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/enqueuer"
	cerrors "github.com/vmware-tanzu/cartographer/pkg/errors"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/dependency"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

type Timer interface {
	Now() metav1.Time
}

type SupplyChainReconciler struct {
	Repo                    repository.Repository
	ConditionManagerBuilder conditions.ConditionManagerBuilder
	DependencyTracker       dependency.DependencyTracker
}

func (r *SupplyChainReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.Info("started")
	defer log.Info("finished")

	log = log.WithValues("supply chain", req.NamespacedName)
	ctx = logr.NewContext(ctx, log)

	supplyChain, err := r.Repo.GetSupplyChain(ctx, req.Name)
	if err != nil {
		log.Error(err, "failed to get supply chain")
		return ctrl.Result{}, fmt.Errorf("failed to get supply chain [%s]: %w", req.NamespacedName, err)
	}

	if supplyChain == nil {
		log.Info("supply chain no longer exists")
		return ctrl.Result{}, nil
	}

	conditionManager := r.ConditionManagerBuilder(v1alpha1.BlueprintReady, supplyChain.Status.Conditions)

	err = r.reconcileSupplyChain(ctx, supplyChain, conditionManager)

	return r.completeReconciliation(ctx, supplyChain, conditionManager, err)
}

func (r *SupplyChainReconciler) completeReconciliation(ctx context.Context, supplyChain *v1alpha1.ClusterSupplyChain, conditionManager conditions.ConditionManager, err error) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx)

	var changed bool
	supplyChain.Status.Conditions, changed = conditionManager.Finalize()

	var updateErr error
	if changed || (supplyChain.Status.ObservedGeneration != supplyChain.Generation) {
		supplyChain.Status.ObservedGeneration = supplyChain.Generation
		updateErr = r.Repo.StatusUpdate(ctx, supplyChain)
		if updateErr != nil {
			log.Error(err, "failed to update status for supply chain")
			return ctrl.Result{}, fmt.Errorf("failed to update status for supply chain: %w", updateErr)
		}
	}

	if err != nil {
		if cerrors.IsUnhandledError(err) {
			log.Error(err, "unhandled error reconciling supply chain")
			return ctrl.Result{}, err
		}
		log.Info("handled error reconciling supply chain", "handled error", err)
	}

	return ctrl.Result{}, nil
}

func (r *SupplyChainReconciler) reconcileSupplyChain(ctx context.Context, chain *v1alpha1.ClusterSupplyChain, conditionManager conditions.ConditionManager) error {
	log := logr.FromContextOrDiscard(ctx)
	var resourcesNotFound []string

	for _, resource := range chain.Spec.Resources {
		if resource.TemplateRef.Name != "" {
			found, err := r.validateResource(ctx, chain, resource.TemplateRef.Name, resource.TemplateRef.Kind)
			if err != nil {
				log.Error(err, "failed to get cluster template", "template",
					fmt.Sprintf("%s/%s", resource.TemplateRef.Kind, resource.TemplateRef.Name))
				return cerrors.NewUnhandledError(fmt.Errorf("failed to get cluster template: %w", err))
			}

			if !found {
				resourcesNotFound = append(resourcesNotFound, resource.Name)
			}
		} else {
			for _, option := range resource.TemplateRef.Options {
				if option.Name != "" {
					found, err := r.validateResource(ctx, chain, option.Name, resource.TemplateRef.Kind)
					if err != nil {
						log.Error(err, "failed to get cluster template", "template",
							fmt.Sprintf("%s/%s", resource.TemplateRef.Kind, resource.TemplateRef.Name))
						return cerrors.NewUnhandledError(fmt.Errorf("failed to get cluster template: %w", err))
					}

					if !found {
						resourcesNotFound = append(resourcesNotFound, resource.Name)
					}
				}
			}
		}
	}

	if len(resourcesNotFound) > 0 {
		conditionManager.AddPositive(conditions.TemplatesNotFoundCondition(resourcesNotFound))
	} else {
		conditionManager.AddPositive(conditions.TemplatesFoundCondition())
	}

	return nil
}

func (r *SupplyChainReconciler) validateResource(ctx context.Context, supplyChain *v1alpha1.ClusterSupplyChain, templateName, templateKind string) (bool, error) {
	template, err := r.Repo.GetTemplate(ctx, templateName, templateKind)
	if err != nil {
		return false, err
	}

	r.DependencyTracker.Track(dependency.Key{
		GroupKind: schema.GroupKind{
			Group: v1alpha1.SchemeGroupVersion.Group,
			Kind:  templateKind,
		},
		NamespacedName: types.NamespacedName{
			Name: templateName,
		},
	}, types.NamespacedName{
		Namespace: supplyChain.Namespace,
		Name:      supplyChain.Name,
	})

	return template != nil, nil
}

func (r *SupplyChainReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Repo = repository.NewRepository(
		mgr.GetClient(),
		repository.NewCache(mgr.GetLogger().WithName("supply-chain-repo-cache")),
	)

	r.ConditionManagerBuilder = conditions.NewConditionManager
	r.DependencyTracker = dependency.NewDependencyTracker(
		2*utils.DefaultResyncTime,
		mgr.GetLogger().WithName("tracker-supply-chain"),
	)

	builder := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ClusterSupplyChain{})

	for _, template := range v1alpha1.ValidSupplyChainTemplates {
		builder = builder.Watches(
			template,
			enqueuer.EnqueueTracked(template, r.DependencyTracker, mgr.GetScheme()),
		)
	}

	return builder.Complete(r)
}
