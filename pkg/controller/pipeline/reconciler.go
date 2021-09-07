package pipeline

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-logr/logr"
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Reconciler struct {
	Repository repository.Repository
}

type PipelineTemplatingContext struct {
	Pipeline *v1alpha1.Pipeline `json:"pipeline"`
}

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	logger := logr.FromContext(ctx).
		WithValues("name", request.Name, "namespace", request.Namespace)
	logger.Info("started")

	reconciler := &pipelineReconciler{
		Repository: r.Repository,
	}

	pipeline, getPipelineError := r.Repository.GetPipeline(request.Name, request.Namespace)

	if kerrors.IsNotFound(getPipelineError) {
		logger.Info("pipeline no longer exists")
		logger.Info("finished")
		return ctrl.Result{}, nil
	} else if getPipelineError == nil {
		conditionManager := conditions.NewConditionManager(v1alpha1.PipelineReady, pipeline.Status.Conditions)

		// realize
		condition := reconciler.Realize(pipeline, logger)

		// conditions
		if condition != nil {
			conditionManager.AddPositive(*condition)
			pipeline.Status.Conditions, _ = conditionManager.Finalize()
			getPipelineError = r.Repository.StatusUpdate(pipeline) // FIXME: deal with errors!
			if getPipelineError != nil {
				panic("badbad")
			}
		}
	}

	logger.Info("finished")
	return ctrl.Result{}, getPipelineError
}

// ------------------------------------------------------------------
// ------------------------------------------------------------------
// ------------------------------------------------------------------
// ------------------------------------------------------------------


type Realizer interface {
	Realize(pipeline *v1alpha1.Pipeline, logger logr.Logger) *metav1.Condition
}

type pipelineReconciler struct {
	Repository repository.Repository
}

func (p *pipelineReconciler) Realize(pipeline *v1alpha1.Pipeline, logger logr.Logger) *metav1.Condition {
	template, err := p.Repository.GetTemplate(pipeline.Spec.RunTemplate)

	if err != nil {
		errorMessage := fmt.Sprintf("could not get RunTemplate '%s'", pipeline.Spec.RunTemplate.Name)
		logger.Error(err, errorMessage)

		return RunTemplateMissingCondition(fmt.Errorf("%s: %w", errorMessage, err))
	}

	labels := map[string]string{}

	stampContext := templates.StamperBuilder(
		pipeline,
		PipelineTemplatingContext{
			Pipeline: pipeline,
		},
		labels,
	)

	stampedObject, err := stampContext.Stamp(template.GetResourceTemplate().Raw)
	// FIXME untested
	if err != nil {
		logger.Error(err, "could not stamp template")
	}

	err = p.Repository.Create(stampedObject)
	if err != nil {
		errorMessage := "could not create object"
		logger.Error(err, errorMessage)

		return StampedObjectRejectedByAPIServerCondition(fmt.Errorf("%s: %w", errorMessage, err))
	}

	return nil
}

