package pipeline

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Reconciler interface {
	Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error)
}

func NewReconciler(repository repository.Repository, realizer Realizer) Reconciler {
	return &reconciler{
		repository: repository,
		realizer:  realizer,
	}
}

type reconciler struct {
	repository repository.Repository
	realizer Realizer
}

type PipelineTemplatingContext struct {
	Pipeline *v1alpha1.Pipeline `json:"pipeline"`
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
	} else if getPipelineError == nil {
		conditionManager := conditions.NewConditionManager(v1alpha1.PipelineReady, pipeline.Status.Conditions)

		// realize
		condition := r.realizer.Realize(pipeline, logger,r.repository)

		// conditions
		if condition != nil {
			conditionManager.AddPositive(*condition)
			pipeline.Status.Conditions, _ = conditionManager.Finalize()
			// Fixme deal with setstatus!
			getPipelineError = r.repository.StatusUpdate(pipeline) // FIXME: deal with errors!
			if getPipelineError != nil {
				panic("badbad")
			}
		}
	}

	logger.Info("finished")
	return ctrl.Result{}, getPipelineError
}
