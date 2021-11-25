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

package runnable

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/controller"
	realizerclient "github.com/vmware-tanzu/cartographer/pkg/realizer/client"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/runnable"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/tracker"
)

type Reconciler struct {
	Repo                    repository.Repository
	Realizer                realizer.Realizer
	DynamicTracker          tracker.DynamicTracker
	ConditionManagerBuilder conditions.ConditionManagerBuilder
	conditionManager        conditions.ConditionManager
	RepositoryBuilder       repository.RepositoryBuilder
	ClientBuilder           realizerclient.ClientBuilder
	RunnableCache           repository.RepoCache
	logger                  logr.Logger
}

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	r.logger = logr.FromContext(ctx)
	r.logger.Info("started")
	defer r.logger.Info("finished")

	runnable, err := r.Repo.GetRunnable(ctx, request.Name, request.Namespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("get runnable: %w", err)
	}

	if runnable == nil {
		r.logger.Info("runnable no longer exists")
		return ctrl.Result{}, nil
	}

	r.conditionManager = r.ConditionManagerBuilder(v1alpha1.RunnableReady, runnable.Status.Conditions)

	serviceAccountName := "default"
	if runnable.Spec.ServiceAccountName != "" {
		serviceAccountName = runnable.Spec.ServiceAccountName
	}

	secret, err := r.Repo.GetServiceAccountSecret(ctx, serviceAccountName, request.Namespace)
	if err != nil {
		r.conditionManager.AddPositive(ServiceAccountSecretNotFoundCondition(err))
		return r.completeReconciliation(ctx, runnable, nil, fmt.Errorf("get secret for service account '%s': %w", serviceAccountName, err))
	}

	runnableClient, err := r.ClientBuilder(secret)
	if err != nil {
		r.conditionManager.AddPositive(ClientBuilderErrorCondition(err))
		return r.completeReconciliation(ctx, runnable, nil, controller.NewUnhandledError(fmt.Errorf("build resource realizer: %w", err)))
	}

	stampedObject, outputs, err := r.Realizer.Realize(ctx, runnable, r.Repo, r.RepositoryBuilder(runnableClient, r.RunnableCache))
	if err != nil {
		switch typedErr := err.(type) {
		case realizer.GetRunTemplateError:
			r.conditionManager.AddPositive(RunTemplateMissingCondition(typedErr))
			err = controller.NewUnhandledError(err)
		case realizer.ResolveSelectorError:
			r.conditionManager.AddPositive(TemplateStampFailureCondition(typedErr))
		case realizer.StampError:
			r.conditionManager.AddPositive(TemplateStampFailureCondition(typedErr))
		case realizer.ApplyStampedObjectError:
			r.conditionManager.AddPositive(StampedObjectRejectedByAPIServerCondition(typedErr))
			if !kerrors.IsForbidden(typedErr.Err) {
				err = controller.NewUnhandledError(err)
			}
		case realizer.ListCreatedObjectsError:
			r.conditionManager.AddPositive(FailedToListCreatedObjectsCondition(typedErr))
			err = controller.NewUnhandledError(err)
		case realizer.RetrieveOutputError:
			r.conditionManager.AddPositive(OutputPathNotSatisfiedCondition(typedErr))
		default:
			r.conditionManager.AddPositive(UnknownErrorCondition(typedErr))
			err = controller.NewUnhandledError(err)
		}
	} else {
		r.conditionManager.AddPositive(RunTemplateReadyCondition())
	}

	var trackingError error
	if stampedObject != nil {
		trackingError = r.DynamicTracker.Watch(r.logger, stampedObject, &handler.EnqueueRequestForOwner{OwnerType: &v1alpha1.Runnable{}})
		if trackingError != nil {
			r.logger.Error(err, "dynamic tracker watch")
			err = controller.NewUnhandledError(trackingError)
		}
	}

	return r.completeReconciliation(ctx, runnable, outputs, err)
}

func (r *Reconciler) completeReconciliation(ctx context.Context, runnable *v1alpha1.Runnable, outputs map[string]apiextensionsv1.JSON, err error) (ctrl.Result, error) {
	var changed bool
	runnable.Status.Conditions, changed = r.conditionManager.Finalize()

	if changed || (runnable.Status.ObservedGeneration != runnable.Generation) {
		runnable.Status.Outputs = outputs
		statusUpdateError := r.Repo.StatusUpdate(ctx, runnable)
		if statusUpdateError != nil {
			return ctrl.Result{}, fmt.Errorf("update runnable status: %w", statusUpdateError)
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
