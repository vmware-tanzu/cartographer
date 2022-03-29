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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/enqueuer"
	cerrors "github.com/vmware-tanzu/cartographer/pkg/errors"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/dependency"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

type Reconciler struct {
	Repo              repository.Repository
	DependencyTracker dependency.DependencyTracker
	conditionManager  conditions.ConditionManager
}

func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.Info("started")
	defer log.Info("finished")

	log = log.WithValues("delivery", req.NamespacedName)
	ctx = logr.NewContext(ctx, log)

	delivery, err := r.Repo.GetDelivery(ctx, req.Name)
	if err != nil {
		log.Error(err, "failed to get delivery")
		return ctrl.Result{}, fmt.Errorf("failed to get delivery [%s]: %w", req.NamespacedName, err)
	}

	if delivery == nil {
		log.Info("delivery no longer exists")
		return ctrl.Result{}, nil
	}

	r.conditionManager = conditions.NewConditionManager(v1alpha1.DeliveryReady, delivery.Status.Conditions)

	err = r.reconcileDelivery(ctx, delivery)

	return r.completeReconciliation(ctx, delivery, err)
}

func (r *Reconciler) reconcileDelivery(ctx context.Context, delivery *v1alpha1.ClusterDelivery) error {
	log := logr.FromContextOrDiscard(ctx)
	var resourcesNotFound []string

	for _, resource := range delivery.Spec.Resources {
		if resource.TemplateRef.Name != "" {
			found, err := r.validateResource(ctx, delivery, resource.TemplateRef.Name, resource.TemplateRef.Kind)
			if err != nil {
				log.Error(err, "failed to get delivery cluster template", "template",
					fmt.Sprintf("%s/%s", resource.TemplateRef.Kind, resource.TemplateRef.Name))
				return cerrors.NewUnhandledError(fmt.Errorf("failed to get delivery cluster template: %w", err))
			}

			if !found {
				resourcesNotFound = append(resourcesNotFound, resource.Name)
			}
		} else {
			for _, option := range resource.TemplateRef.Options {
				found, err := r.validateResource(ctx, delivery, option.Name, resource.TemplateRef.Kind)
				if err != nil {
					log.Error(err, "failed to get delivery cluster template", "template",
						fmt.Sprintf("%s/%s", resource.TemplateRef.Kind, resource.TemplateRef.Name))
					return cerrors.NewUnhandledError(fmt.Errorf("failed to get delivery cluster template: %w", err))
				}

				if !found {
					resourcesNotFound = append(resourcesNotFound, resource.Name)
				}
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

func (r *Reconciler) validateResource(ctx context.Context, delivery *v1alpha1.ClusterDelivery, templateName, templateKind string) (bool, error) {
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
		Namespace: delivery.Namespace,
		Name:      delivery.Name,
	})
	return template != nil, nil
}

func (r *Reconciler) completeReconciliation(ctx context.Context, delivery *v1alpha1.ClusterDelivery, err error) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx)

	var changed bool
	delivery.Status.Conditions, changed = r.conditionManager.Finalize()

	var updateErr error
	if changed || (delivery.Status.ObservedGeneration != delivery.Generation) {
		delivery.Status.ObservedGeneration = delivery.Generation
		updateErr = r.Repo.StatusUpdate(ctx, delivery)
		if updateErr != nil {
			log.Error(err, "failed to update status for delivery")
			return ctrl.Result{}, fmt.Errorf("failed to update status for delivery: %w", updateErr)
		}
	}

	if err != nil {
		if cerrors.IsUnhandledError(err) {
			log.Error(err, "unhandled error reconciling delivery")
			return ctrl.Result{}, err
		}
		log.Info("handled error reconciling delivery", "handled error", err)
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Repo = repository.NewRepository(
		mgr.GetClient(),
		repository.NewCache(mgr.GetLogger().WithName("delivery-repo-cache")),
	)

	r.DependencyTracker = dependency.NewDependencyTracker(
		2*utils.DefaultResyncTime,
		mgr.GetLogger().WithName("tracker-delivery"),
	)

	builder := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ClusterDelivery{})

	for _, template := range v1alpha1.ValidDeliveryTemplates {
		builder = builder.Watches(
			&source.Kind{Type: template},
			enqueuer.EnqueueTracked(template, r.DependencyTracker, mgr.GetScheme()),
		)
	}

	return builder.Complete(r)
}
