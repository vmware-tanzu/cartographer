package pipeline

import (
	"context"
	"fmt"

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
	logger := logr.FromContext(ctx)
	// FIXME: seems to already have name and namespace, this causes a duplicate
	//logger := logr.FromContext(ctx).
	//	WithValues("name", request.Name, "namespace", request.Namespace)
	logger.Info("started")

	r.realize(request, logger)

	logger.Info("finished")
	return ctrl.Result{}, nil
}

func (r *Reconciler) realize(request ctrl.Request, logger logr.Logger) {
	pipeline, err := r.Repository.GetPipeline(request.Name, request.Namespace)

	if kerrors.IsNotFound(err) {
		logger.Info("pipeline no longer exists")
		return
	}

	conditionManager := conditions.NewConditionManager(v1alpha1.PipelineReady, pipeline.Status.Conditions)

	template, err := r.Repository.GetTemplate(pipeline.Spec.RunTemplate)

	if err != nil {
		errorMessage := fmt.Sprintf("could not get RunTemplate '%s'", pipeline.Spec.RunTemplate.Name)
		logger.Error(err, errorMessage)

		conditionManager.AddPositive(RunTemplateMissingCondition(fmt.Errorf("%s: %w", errorMessage, err)))
		pipeline.Status.Conditions, _ = conditionManager.Finalize()
		_ = r.Repository.StatusUpdate(pipeline) // FIXME: deal with errors!

		return
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

	// TODO: remove
	logger.Info(fmt.Sprintf("EMJ TEST WE STAMPED?? %+v", stampedObject))

	// FIXME untested err
	// FIXME must use create only.
	err = r.Repository.AssureObjectExistsOnCluster(stampedObject)
	if err != nil {
		logger.Error(err, "could not create object")
	}

}
