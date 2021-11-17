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
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/runnable"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/tracker"
)

type Reconciler struct {
	Repo           repository.Repository
	Realizer       realizer.Realizer
	DynamicTracker tracker.DynamicTracker
}

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	logger := logr.FromContext(ctx).
		WithValues("name", request.Name, "namespace", request.Namespace)
	logger.Info("started")
	defer logger.Info("finished")

	runnable, err := r.Repo.GetRunnable(request.Name, request.Namespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("get runnable: %w", err)
	}

	if runnable == nil {
		logger.Info("runnable no longer exists")
		return ctrl.Result{}, nil
	}

	condition, outputs, stampedObject := r.Realizer.Realize(ctx, runnable, logger, r.Repo)
	if stampedObject != nil {
		err = r.DynamicTracker.Watch(logger, stampedObject, &handler.EnqueueRequestForOwner{OwnerType: &v1alpha1.Runnable{}})
		if err != nil {
			logger.Error(err, "dynamic tracker watch")
		}
	}

	conditionManager := conditions.NewConditionManager(v1alpha1.RunnableReady, runnable.Status.Conditions)
	conditionManager.AddPositive(*condition)
	//TODO: deal with changed (story #84)
	runnable.Status.Conditions, _ = conditionManager.Finalize()
	runnable.Status.Outputs = outputs

	statusUpdateError := r.Repo.StatusUpdate(runnable)
	if statusUpdateError != nil {
		return ctrl.Result{}, fmt.Errorf("update runnable status: %w", statusUpdateError)
	}

	return ctrl.Result{}, nil
}
