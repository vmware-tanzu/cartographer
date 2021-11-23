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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/controller"
	"github.com/vmware-tanzu/cartographer/pkg/logger"
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
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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
		return ctrl.Result{}, nil
	}

	r.conditionManager = r.ConditionManagerBuilder(v1alpha1.RunnableReady, runnable.Status.Conditions)

	stampedObject, outputs, err := r.Realizer.Realize(ctx, runnable, r.Repo)
	if err != nil {
		log.V(logger.DEBUG).Info("failed to realize")
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
			err = controller.NewUnhandledError(err)
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
		log.V(logger.DEBUG).Info("realized object", "object", stampedObject)
		r.conditionManager.AddPositive(RunTemplateReadyCondition())
	}

	var trackingError error
	if stampedObject != nil {
		trackingError = r.DynamicTracker.Watch(log, stampedObject, &handler.EnqueueRequestForOwner{OwnerType: &v1alpha1.Runnable{}})
		if trackingError != nil {
			log.Error(err, "failed to add informer for object", "object", stampedObject)
			err = controller.NewUnhandledError(trackingError)
		} else {
			log.V(logger.DEBUG).Info("added informer for object", "object", stampedObject)
		}
	}

	var changed bool
	runnable.Status.Conditions, changed = r.conditionManager.Finalize()

	if changed || (runnable.Status.ObservedGeneration != runnable.Generation) {
		runnable.Status.Outputs = outputs
		statusUpdateError := r.Repo.StatusUpdate(ctx, runnable)
		if statusUpdateError != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update status for runnable: %w", statusUpdateError)
		}
	}

	if err != nil {
		if controller.IsUnhandledError(err) {
			log.Error(err, "unhandled error reconciling runnable")
			return ctrl.Result{}, err
		}
		log.Info("handled error reconciling runnable", "handled error", err)
	}

	return ctrl.Result{}, nil
}
