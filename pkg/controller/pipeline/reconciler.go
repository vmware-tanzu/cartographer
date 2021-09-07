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

	// get pipeline
	pipeline, err := r.Repository.GetPipeline(request.Name, request.Namespace)

	if kerrors.IsNotFound(err) {
		logger.Info("pipeline no longer exists")
		// FIXME: duplicated finished messages
		logger.Info("finished")
		return ctrl.Result{}, nil
	}

	// realize
	conditionManager := conditions.NewConditionManager(v1alpha1.PipelineReady, pipeline.Status.Conditions)
	condition := r.realize(pipeline, logger)

	// conditions
	if condition != nil {
		conditionManager.AddPositive(*condition)
		pipeline.Status.Conditions, _ = conditionManager.Finalize()
		err = r.Repository.StatusUpdate(pipeline) // FIXME: deal with errors!
		if err != nil {
			panic("badbad")
		}
	}

	logger.Info("finished")
	return ctrl.Result{}, nil
}

func (r *Reconciler) realize(pipeline *v1alpha1.Pipeline, logger logr.Logger) *metav1.Condition {
	template, err := r.Repository.GetTemplate(pipeline.Spec.RunTemplate)

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

	err = r.Repository.Create(stampedObject)
	if err != nil {
		errorMessage := "could not create object"
		logger.Error(err, errorMessage)

		return StampedObjectRejectedByAPIServerCondition(fmt.Errorf("%s: %w", errorMessage, err))
	}

	return nil
}
