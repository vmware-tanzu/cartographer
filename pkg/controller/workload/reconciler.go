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

//go:generate go run -modfile ../../../hack/tools/go.mod github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/controller"
	"github.com/vmware-tanzu/cartographer/pkg/logger"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/workload"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/dependency"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/stamped"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

type Reconciler struct {
	Repo                    repository.Repository
	ConditionManagerBuilder conditions.ConditionManagerBuilder
	ResourceRealizerBuilder realizer.ResourceRealizerBuilder
	Realizer                realizer.Realizer
	conditionManager        conditions.ConditionManager
	StampedTracker          stamped.StampedTracker
	DependencyTracker       dependency.DependencyTracker
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.Info("started")
	defer log.Info("finished")

	log = log.WithValues("workload", req.NamespacedName)
	ctx = logr.NewContext(ctx, log)

	workload, err := r.Repo.GetWorkload(ctx, req.Name, req.Namespace)
	if err != nil {
		log.Error(err, "failed to get workload")
		return ctrl.Result{}, fmt.Errorf("failed to get workload [%s]: %w", req.NamespacedName, err)
	}

	if workload == nil {
		log.Info("workload no longer exists")
		r.DependencyTracker.ClearTracked(types.NamespacedName{
			Namespace: req.Namespace,
			Name:      req.Name,
		})

		return ctrl.Result{}, nil
	}

	r.conditionManager = r.ConditionManagerBuilder(v1alpha1.WorkloadReady, workload.Status.Conditions)

	supplyChain, err := r.getSupplyChainsForWorkload(ctx, workload)
	if err != nil {
		return r.completeReconciliation(ctx, workload, err)
	}

	log = log.WithValues("supply chain", supplyChain.Name)
	ctx = logr.NewContext(ctx, log)

	supplyChainGVK, err := utils.GetObjectGVK(supplyChain, r.Repo.GetScheme())
	if err != nil {
		log.Error(err, "failed to get object gvk for supply chain")
		return r.completeReconciliation(ctx, workload, controller.NewUnhandledError(
			fmt.Errorf("failed to get object gvk for supply chain [%s]: %w", supplyChain.Name, err)),
		)
	}

	workload.Status.SupplyChainRef.Kind = supplyChainGVK.Kind
	workload.Status.SupplyChainRef.Name = supplyChain.Name

	if !r.isSupplyChainReady(supplyChain) {
		r.conditionManager.AddPositive(MissingReadyInSupplyChainCondition(getSupplyChainReadyCondition(supplyChain)))
		log.Info("supply chain is not in ready state")
		return r.completeReconciliation(ctx, workload, fmt.Errorf("supply chain [%s] is not in ready state", supplyChain.Name))
	}
	r.conditionManager.AddPositive(SupplyChainReadyCondition())

	serviceAccountName, serviceAccountNS := getServiceAccountNameAndNamespace(workload, supplyChain)

	secret, err := r.Repo.GetServiceAccountSecret(ctx, serviceAccountName, serviceAccountNS)
	if err != nil {
		r.conditionManager.AddPositive(ServiceAccountSecretNotFoundCondition(err))
		log.Info("failed to get service account secret", "service account", fmt.Sprintf("%s/%s", serviceAccountNS, serviceAccountName))
		return r.completeReconciliation(ctx, workload, fmt.Errorf("failed to get service account secret [%s]: %w", fmt.Sprintf("%s/%s", serviceAccountNS, serviceAccountName), err))
	}

	resourceRealizer, err := r.ResourceRealizerBuilder(secret, workload, r.Repo, supplyChain.Spec.Params)
	if err != nil {
		r.conditionManager.AddPositive(ResourceRealizerBuilderErrorCondition(err))
		log.Error(err, "failed to build resource realizer")
		return r.completeReconciliation(ctx, workload, controller.NewUnhandledError(
			fmt.Errorf("failed to build resource realizer: %w", err)))
	}

	selectedTemplates, stampedObjects, err := r.Realizer.Realize(ctx, resourceRealizer, supplyChain)
	if err != nil {
		log.V(logger.DEBUG).Info("failed to realize")
		switch typedErr := err.(type) {
		case realizer.GetSupplyChainTemplateError:
			r.conditionManager.AddPositive(TemplateObjectRetrievalFailureCondition(typedErr))
			err = controller.NewUnhandledError(err)
		case realizer.StampError:
			r.conditionManager.AddPositive(TemplateStampFailureCondition(typedErr))
		case realizer.ApplyStampedObjectError:
			r.conditionManager.AddPositive(TemplateRejectedByAPIServerCondition(typedErr))
			if !kerrors.IsForbidden(typedErr.Err) {
				err = controller.NewUnhandledError(err)
			}
		case realizer.RetrieveOutputError:
			r.conditionManager.AddPositive(MissingValueAtPathCondition(typedErr.StampedObject, typedErr.JsonPathExpression()))
		case realizer.ResolveTemplateOptionError:
			r.conditionManager.AddPositive(ResolveTemplateOptionsErrorCondition(typedErr))
		case realizer.TemplateOptionsMatchError:
			r.conditionManager.AddPositive(TemplateOptionsMatchErrorCondition(typedErr))
		default:
			r.conditionManager.AddPositive(UnknownResourceErrorCondition(typedErr))
			err = controller.NewUnhandledError(err)
		}
	} else {
		if log.V(logger.DEBUG).Enabled() {
			for _, stampedObject := range stampedObjects {
				log.V(logger.DEBUG).Info("realized object",
					"object", stampedObject)
			}
		}
		r.conditionManager.AddPositive(ResourcesSubmittedCondition())
	}

	r.trackDependencies(workload, selectedTemplates, serviceAccountName, serviceAccountNS)

	var trackingError error
	if len(stampedObjects) > 0 {
		for _, stampedObject := range stampedObjects {
			trackingError = r.StampedTracker.Watch(log, stampedObject, &handler.EnqueueRequestForOwner{OwnerType: &v1alpha1.Workload{}})
			if trackingError != nil {
				log.Error(err, "failed to add informer for object",
					"object", stampedObject)
				err = controller.NewUnhandledError(trackingError)
			} else {
				log.V(logger.DEBUG).Info("added informer for object",
					"object", stampedObject)
			}
		}
	}

	return r.completeReconciliation(ctx, workload, err)
}

func (r *Reconciler) completeReconciliation(ctx context.Context, workload *v1alpha1.Workload, err error) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx)
	var changed bool
	workload.Status.Conditions, changed = r.conditionManager.Finalize()

	var updateErr error
	if changed || (workload.Status.ObservedGeneration != workload.Generation) {
		workload.Status.ObservedGeneration = workload.Generation
		updateErr = r.Repo.StatusUpdate(ctx, workload)
		if updateErr != nil {
			log.Error(err, "failed to update status for workload")
			return ctrl.Result{}, fmt.Errorf("failed to update status for workload: %w", updateErr)
		}
	}

	if err != nil {
		if controller.IsUnhandledError(err) {
			log.Error(err, "unhandled error reconciling workload")
			return ctrl.Result{}, err
		}
		log.Info("handled error reconciling workload", "handled error", err)
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
	log := logr.FromContextOrDiscard(ctx)
	if len(workload.Labels) == 0 {
		r.conditionManager.AddPositive(WorkloadMissingLabelsCondition())
		log.Info("workload is missing required labels")
		return nil, fmt.Errorf("workload [%s/%s] is missing required labels",
			workload.Namespace, workload.Name)
	}

	supplyChains, err := r.Repo.GetSupplyChainsForWorkload(ctx, workload)
	if err != nil {
		log.Error(err, "failed to get supply chains for workload")
		return nil, controller.NewUnhandledError(fmt.Errorf("failed to get supply chains for workload [%s/%s]: %w",
			workload.Namespace, workload.Name, err))
	}

	if len(supplyChains) == 0 {
		r.conditionManager.AddPositive(SupplyChainNotFoundCondition(workload.Labels))
		log.Info("no supply chain found where full selector is satisfied by label",
			"labels", workload.Labels)
		return nil, fmt.Errorf("no supply chain [%s/%s] found where full selector is satisfied by labels: %v",
			workload.Namespace, workload.Name, workload.Labels)
	}

	if len(supplyChains) > 1 {
		r.conditionManager.AddPositive(TooManySupplyChainMatchesCondition())
		log.Info("more than one supply chain selected for workload",
			"supply chains", getSupplyChainNames(supplyChains))
		return nil, fmt.Errorf("more than one supply chain selected for workload [%s/%s]: %+v",
			workload.Namespace, workload.Name, getSupplyChainNames(supplyChains))
	}

	log.V(logger.DEBUG).Info("supply chain matched for workload", "supply chain", supplyChains[0].Name)
	return supplyChains[0], nil
}

func (r *Reconciler) trackDependencies(workload *v1alpha1.Workload, selectedTemplates []templates.Template, serviceAccountName, serviceAccountNS string) {
	r.DependencyTracker.ClearTracked(types.NamespacedName{
		Namespace: workload.Namespace,
		Name:      workload.Name,
	})

	r.DependencyTracker.Track(dependency.Key{
		GroupKind: schema.GroupKind{
			Group: corev1.SchemeGroupVersion.Group,
			Kind:  rbacv1.ServiceAccountKind,
		},
		NamespacedName: types.NamespacedName{
			Namespace: serviceAccountNS,
			Name:      serviceAccountName,
		},
	}, types.NamespacedName{
		Namespace: workload.Namespace,
		Name:      workload.Name,
	})

	for _, selectedTemplate := range selectedTemplates {
		r.DependencyTracker.Track(
			dependency.Key{
				GroupKind: schema.GroupKind{
					Group: v1alpha1.SchemeGroupVersion.Group,
					Kind:  selectedTemplate.GetKind(),
				},
				NamespacedName: types.NamespacedName{
					Name: selectedTemplate.GetName(),
				},
			},
			types.NamespacedName{
				Namespace: workload.Namespace,
				Name:      workload.Name,
			},
		)
	}
}

func getSupplyChainNames(objs []*v1alpha1.ClusterSupplyChain) []string {
	var names []string
	for _, obj := range objs {
		names = append(names, obj.GetName())
	}

	return names
}

func getServiceAccountNameAndNamespace(workload *v1alpha1.Workload, supplyChain *v1alpha1.ClusterSupplyChain) (string, string) {
	serviceAccountName := "default"
	serviceAccountNS := workload.Namespace

	if workload.Spec.ServiceAccountName != "" {
		serviceAccountName = workload.Spec.ServiceAccountName
	} else if supplyChain.Spec.ServiceAccountRef.Name != "" {
		serviceAccountName = supplyChain.Spec.ServiceAccountRef.Name
		if supplyChain.Spec.ServiceAccountRef.Namespace != "" {
			serviceAccountNS = supplyChain.Spec.ServiceAccountRef.Namespace
		}
	}

	return serviceAccountName, serviceAccountNS
}
