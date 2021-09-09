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

package pipeline

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	realizer "github.com/vmware-tanzu/cartographer/pkg/realizer/pipeline"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
)

type Reconciler interface {
	Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error)
}

func NewReconciler(repository repository.Repository, realizer realizer.Realizer) Reconciler {
	return &reconciler{
		repository: repository,
		realizer:   realizer,
	}
}

type reconciler struct {
	repository repository.Repository
	realizer   realizer.Realizer
}

func (r *reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	logger := logr.FromContext(ctx).
		WithValues("name", request.Name, "namespace", request.Namespace)
	logger.Info("started")
	defer logger.Info("finished")

	pipeline, err := r.repository.GetPipeline(request.Name, request.Namespace)

	if kerrors.IsNotFound(err) {
		logger.Info("pipeline no longer exists")
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	} else if err == nil {
		conditionManager := conditions.NewConditionManager(v1alpha1.PipelineReady, pipeline.Status.Conditions)

		condition := r.realizer.Realize(pipeline, logger, r.repository)

		if condition != nil {
			conditionManager.AddPositive(*condition)
			//TODO: deal with changed
			pipeline.Status.Conditions, _ = conditionManager.Finalize()
			statusUpdateError := r.repository.StatusUpdate(pipeline)
			if statusUpdateError != nil {
				logger.Info("finished")
				return ctrl.Result{}, fmt.Errorf("update workload status: %w", statusUpdateError)
			}
		}
	}
	return ctrl.Result{}, nil
}
