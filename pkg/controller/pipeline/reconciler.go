package pipeline

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	realizerpipeline "github.com/vmware-tanzu/cartographer/pkg/realizer/pipeline"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
)

type Reconciler interface {
	Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error)
}

func NewReconciler(repository repository.Repository, realizer realizerpipeline.Realizer) Reconciler {
	return &reconciler{
		repository: repository,
		realizer:   realizer,
	}
}

type reconciler struct {
	repository repository.Repository
	realizer   realizerpipeline.Realizer
}

func (r *reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	logger := logr.FromContext(ctx).
		WithValues("name", request.Name, "namespace", request.Namespace)
	logger.Info("started")

	pipeline, getPipelineError := r.repository.GetPipeline(request.Name, request.Namespace)

	if kerrors.IsNotFound(getPipelineError) {
		logger.Info("pipeline no longer exists")
		logger.Info("finished")
		return ctrl.Result{}, nil
	} else if getPipelineError != nil {
		logger.Info("finished")
		return ctrl.Result{}, getPipelineError
	} else if getPipelineError == nil {
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
	logger.Info("finished")
	return ctrl.Result{}, nil
}
