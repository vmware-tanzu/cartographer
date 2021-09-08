package pipeline

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"fmt"

	"github.com/go-logr/logr"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

//counterfeiter:generate . Realizer
type Realizer interface {
	Realize(pipeline *v1alpha1.Pipeline, logger logr.Logger, repository repository.Repository) *v1.Condition
}

func NewRealizer() Realizer {
	return &pipelineRealizer{}
}

type pipelineRealizer struct{}

func (p *pipelineRealizer) Realize(pipeline *v1alpha1.Pipeline, logger logr.Logger, repository repository.Repository) *v1.Condition {
	template, err := repository.GetTemplate(pipeline.Spec.RunTemplateRef)

	if err != nil {
		errorMessage := fmt.Sprintf("could not get RunTemplate '%s'", pipeline.Spec.RunTemplateRef.Name)
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
	if err != nil {
		errorMessage := "could not stamp template"
		logger.Error(err, errorMessage)
		return TemplateStampFailureCondition(fmt.Errorf("%s: %w", errorMessage, err))
	}

	err = repository.Create(stampedObject)
	if err != nil {
		errorMessage := "could not create object"
		logger.Error(err, errorMessage)
		return StampedObjectRejectedByAPIServerCondition(fmt.Errorf("%s: %w", errorMessage, err))
	}

	return RunTemplateReadyCondition()
}
