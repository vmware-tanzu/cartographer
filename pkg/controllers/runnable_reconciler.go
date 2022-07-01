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
	"reflect"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/controllers/external"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/enqueuer"
	cerrors "github.com/vmware-tanzu/cartographer/pkg/errors"
	"github.com/vmware-tanzu/cartographer/pkg/logger"
	"github.com/vmware-tanzu/cartographer/pkg/mapper"
	realizerclient "github.com/vmware-tanzu/cartographer/pkg/realizer/client"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/runnable"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/dependency"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/stamped"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

type RunnableReconciler struct {
	Repo                    repository.Repository
	Realizer                realizer.Realizer
	ConditionManagerBuilder conditions.ConditionManagerBuilder
	conditionManager        conditions.ConditionManager
	RepositoryBuilder       repository.RepositoryBuilder
	ClientBuilder           realizerclient.ClientBuilder
	RunnableCache           repository.RepoCache
	StampedTracker          stamped.StampedTracker
	DependencyTracker       dependency.DependencyTracker
}

func (r *RunnableReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.Info("started")
	defer log.Info("finished")

	log = log.WithValues("runnable", req.NamespacedName)
	ctx = logr.NewContext(ctx, log)

	runnable, err := r.Repo.GetRunnable(ctx, req.Name, req.Namespace)
	if err != nil {
		log.Error(err, "failed to get runnable")
		return ctrl.Result{}, fmt.Errorf("failed to get runnable [%s]: %w", req.NamespacedName, err)
	}

	if runnable == nil {
		log.Info("runnable no longer exists")
		r.DependencyTracker.ClearTracked(types.NamespacedName{
			Namespace: req.Namespace,
			Name:      req.Name,
		})

		return ctrl.Result{}, nil
	}

	r.conditionManager = r.ConditionManagerBuilder(v1alpha1.RunnableReady, runnable.Status.Conditions)

	serviceAccountName := "default"
	if runnable.Spec.ServiceAccountName != "" {
		serviceAccountName = runnable.Spec.ServiceAccountName
	}

	r.trackDependencies(runnable, serviceAccountName)

	secret, err := r.Repo.GetServiceAccountSecret(ctx, serviceAccountName, req.Namespace)
	if err != nil {
		r.conditionManager.AddPositive(conditions.RunnableServiceAccountSecretNotFoundCondition(err))
		return r.completeReconciliation(ctx, runnable, nil, fmt.Errorf("failed to get secret for service account [%s]: %w", fmt.Sprintf("%s/%s", req.Namespace, serviceAccountName), err))
	}

	runnableClient, discoveryClient, err := r.ClientBuilder(secret, true)
	if err != nil {
		r.conditionManager.AddPositive(conditions.ClientBuilderErrorCondition(err))
		return r.completeReconciliation(ctx, runnable, nil, cerrors.NewUnhandledError(fmt.Errorf("failed to build resource realizer: %w", err)))
	}

	stampedObject, outputs, err := r.Realizer.Realize(ctx, runnable, r.Repo, r.RepositoryBuilder(runnableClient, r.RunnableCache), discoveryClient)
	if err != nil {
		log.V(logger.DEBUG).Info("failed to realize")
		switch typedErr := err.(type) {
		case cerrors.RunnableGetRunTemplateError:
			r.conditionManager.AddPositive(conditions.RunTemplateMissingCondition(typedErr))
			err = cerrors.NewUnhandledError(err)
		case cerrors.RunnableResolveSelectorError:
			r.conditionManager.AddPositive(conditions.RunnableTemplateStampFailureCondition(typedErr))
		case cerrors.RunnableStampError:
			r.conditionManager.AddPositive(conditions.RunnableTemplateStampFailureCondition(typedErr))
		case cerrors.RunnableApplyStampedObjectError:
			r.conditionManager.AddPositive(conditions.StampedObjectRejectedByAPIServerCondition(typedErr))
			if !kerrors.IsForbidden(typedErr.Err) {
				err = cerrors.NewUnhandledError(err)
			}
		case cerrors.RunnableListCreatedObjectsError:
			r.conditionManager.AddPositive(conditions.FailedToListCreatedObjectsCondition(typedErr))
			err = cerrors.NewUnhandledError(err)
		case cerrors.RunnableRetrieveOutputError:
			r.conditionManager.AddPositive(conditions.OutputPathNotSatisfiedCondition(typedErr.StampedObject, typedErr.Error()))
		default:
			r.conditionManager.AddPositive(conditions.UnknownErrorCondition(typedErr))
			err = cerrors.NewUnhandledError(err)
		}
	} else {
		log.V(logger.DEBUG).Info("realized object", "object", stampedObject)
		r.conditionManager.AddPositive(conditions.RunTemplateReadyCondition())
	}

	var stampedObjectStatusPresent = false
	var trackingError error

	if stampedObject != nil {
		stampedCondition := utils.ExtractConditions(stampedObject).ConditionWithType("Succeeded")
		if stampedCondition != nil {
			r.conditionManager.AddPositive(conditions.StampedObjectConditionKnown(stampedCondition))
			stampedObjectStatusPresent = true
		}
		trackingError = r.StampedTracker.Watch(log, stampedObject, &handler.EnqueueRequestForOwner{OwnerType: &v1alpha1.Runnable{}})
		if trackingError != nil {
			log.Error(err, "failed to add informer for object", "object", stampedObject)
			err = cerrors.NewUnhandledError(trackingError)
		} else {
			log.V(logger.DEBUG).Info("added informer for object", "object", stampedObject)
		}
	}
	if !stampedObjectStatusPresent {
		r.conditionManager.AddPositive(conditions.StampedObjectConditionUnknown())
	}

	return r.completeReconciliation(ctx, runnable, outputs, err)
}

func (r *RunnableReconciler) completeReconciliation(ctx context.Context, runnable *v1alpha1.Runnable, outputs map[string]apiextensionsv1.JSON, err error) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx)
	var changed bool
	runnable.Status.Conditions, changed = r.conditionManager.Finalize()

	if changed || (runnable.Status.ObservedGeneration != runnable.Generation) || !reflect.DeepEqual(runnable.Status.Outputs, outputs) {
		runnable.Status.Outputs = outputs
		runnable.Status.ObservedGeneration = runnable.Generation
		statusUpdateError := r.Repo.StatusUpdate(ctx, runnable)
		if statusUpdateError != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update status for runnable: %w", statusUpdateError)
		}
	}

	if err != nil {
		if cerrors.IsUnhandledError(err) {
			log.Error(err, "unhandled error reconciling runnable")
			return ctrl.Result{}, err
		}
		log.Info("handled error reconciling runnable", "handled error", err)
	}

	return ctrl.Result{}, nil
}

func (r *RunnableReconciler) trackDependencies(runnable *v1alpha1.Runnable, serviceAccountName string) {
	r.DependencyTracker.ClearTracked(types.NamespacedName{
		Namespace: runnable.Namespace,
		Name:      runnable.Name,
	})

	r.DependencyTracker.Track(dependency.Key{
		GroupKind: schema.GroupKind{
			Group: corev1.SchemeGroupVersion.Group,
			Kind:  rbacv1.ServiceAccountKind,
		},
		NamespacedName: types.NamespacedName{
			Namespace: runnable.Namespace,
			Name:      serviceAccountName,
		},
	}, types.NamespacedName{
		Namespace: runnable.Namespace,
		Name:      runnable.Name,
	})

	r.DependencyTracker.Track(dependency.Key{
		GroupKind: schema.GroupKind{
			Group: v1alpha1.SchemeGroupVersion.Group,
			Kind:  "ClusterRunTemplate",
		},
		NamespacedName: types.NamespacedName{
			Name: runnable.Spec.RunTemplateRef.Name,
		},
	}, types.NamespacedName{
		Namespace: runnable.Namespace,
		Name:      runnable.Name,
	})
}

func (r *RunnableReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Repo = repository.NewRepository(
		mgr.GetClient(),
		repository.NewCache(mgr.GetLogger().WithName("runnable-repo-cache")),
	)

	r.Realizer = realizer.NewRealizer()
	r.RunnableCache = repository.NewCache(mgr.GetLogger().WithName("runnable-stamping-repo-cache"))
	r.RepositoryBuilder = repository.NewRepository
	r.ClientBuilder = realizerclient.NewClientBuilder(mgr.GetConfig())
	r.ConditionManagerBuilder = conditions.NewConditionManager
	r.DependencyTracker = dependency.NewDependencyTracker(
		2*utils.DefaultResyncTime,
		mgr.GetLogger().WithName("tracker-runnable"),
	)

	builder := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Runnable{}).
		Watches(&source.Kind{Type: &v1alpha1.ClusterRunTemplate{}},
			enqueuer.EnqueueTracked(&v1alpha1.ClusterRunTemplate{}, r.DependencyTracker, mgr.GetScheme()),
		)

	m := mapper.Mapper{
		Client:  mgr.GetClient(),
		Logger:  mgr.GetLogger().WithName("runnable"),
		Tracker: r.DependencyTracker,
	}

	watches := map[client.Object]handler.MapFunc{
		&corev1.ServiceAccount{}:     m.ServiceAccountToRunnableRequests,
		&rbacv1.Role{}:               m.RoleToRunnableRequests,
		&rbacv1.RoleBinding{}:        m.RoleBindingToRunnableRequests,
		&rbacv1.ClusterRole{}:        m.ClusterRoleToRunnableRequests,
		&rbacv1.ClusterRoleBinding{}: m.ClusterRoleBindingToRunnableRequests,
	}

	for kindType, mapFunc := range watches {
		builder = builder.Watches(
			&source.Kind{Type: kindType},
			handler.EnqueueRequestsFromMapFunc(mapFunc),
		)
	}

	controller, err := builder.Build(r)
	if err != nil {
		return fmt.Errorf("failed to build controller for runnable: %w", err)
	}
	r.StampedTracker = &external.ObjectTracker{Controller: controller}

	return nil
}
