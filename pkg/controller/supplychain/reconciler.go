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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/controller"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/dependency"
)

type Timer interface {
	Now() metav1.Time
}

type Reconciler struct {
	Repo                    repository.Repository
	ConditionManagerBuilder conditions.ConditionManagerBuilder
	conditionManager        conditions.ConditionManager
	DependencyTracker       dependency.DependencyTracker
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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

	r.conditionManager = r.ConditionManagerBuilder(v1alpha1.SupplyChainReady, supplyChain.Status.Conditions)

	err = r.reconcileSupplyChain(ctx, supplyChain)

	return r.completeReconciliation(ctx, supplyChain, err)
}

func (r *Reconciler) completeReconciliation(ctx context.Context, supplyChain *v1alpha1.ClusterSupplyChain, err error) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx)

	var changed bool
	supplyChain.Status.Conditions, changed = r.conditionManager.Finalize()

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
		if controller.IsUnhandledError(err) {
			log.Error(err, "unhandled error reconciling supply chain")
			return ctrl.Result{}, err
		}
		log.Info("handled error reconciling supply chain", "handled error", err)
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) reconcileSupplyChain(ctx context.Context, chain *v1alpha1.ClusterSupplyChain) error {
	log := logr.FromContextOrDiscard(ctx)
	var resourcesNotFound []string

	for _, resource := range chain.Spec.Resources {
		if resource.TemplateRef.Name != "" {
			template, err := r.Repo.GetSupplyChainTemplate(ctx, resource.TemplateRef.Name, resource.TemplateRef.Kind)
			if err != nil {
				log.Error(err, "failed to get cluster template", "template", resource.TemplateRef)
				return controller.NewUnhandledError(fmt.Errorf("failed to get cluster template: %w", err))
			}

			if template == nil {
				log.Info("cluster template does not exist", "template", resource.TemplateRef)
				resourcesNotFound = append(resourcesNotFound, resource.Name)
			}

			r.DependencyTracker.Track(dependency.Key{
				GroupKind: schema.GroupKind{
					Group: v1alpha1.SchemeGroupVersion.Group,
					Kind:  resource.TemplateRef.Kind,
				},
				NamespacedName: types.NamespacedName{
					Name: resource.TemplateRef.Name,
				},
			}, types.NamespacedName{
				Namespace: chain.Namespace,
				Name:      chain.Name,
			})
		} else {
			for _, option := range resource.TemplateRef.Options {
				template, err := r.Repo.GetSupplyChainTemplate(ctx, option.Name, resource.TemplateRef.Kind)
				if err != nil {
					log.Error(err, "failed to get cluster template", "template",
						fmt.Sprintf("%s/%s", resource.TemplateRef.Kind, option.Name))
					return controller.NewUnhandledError(fmt.Errorf("failed to get cluster template: %w", err))
				}

				if template == nil {
					log.Info("cluster template does not exist", "template",
						fmt.Sprintf("%s/%s", resource.TemplateRef.Kind, option.Name))
					resourcesNotFound = append(resourcesNotFound, resource.Name)
				}

				r.DependencyTracker.Track(dependency.Key{
					GroupKind: schema.GroupKind{
						Group: v1alpha1.SchemeGroupVersion.Group,
						Kind:  resource.TemplateRef.Kind,
					},
					NamespacedName: types.NamespacedName{
						Name: option.Name,
					},
				}, types.NamespacedName{
					Namespace: chain.Namespace,
					Name:      chain.Name,
				})
			}
		}
	}

	if len(resourcesNotFound) > 0 {
		r.conditionManager.AddPositive(TemplatesNotFoundCondition(resourcesNotFound))
	} else {
		r.conditionManager.AddPositive(TemplatesFoundCondition())
	}

	return nil
}
