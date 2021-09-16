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

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

//counterfeiter:generate . Realizer
type Realizer interface {
	Realize(ctx context.Context, pipeline *v1alpha1.Pipeline, logger logr.Logger, repository repository.Repository) (*v1.Condition, templates.Outputs, *unstructured.Unstructured)
}

func NewRealizer() Realizer {
	return &pipelineRealizer{}
}

type pipelineRealizer struct{}

type TemplatingContext struct {
	Pipeline *v1alpha1.Pipeline `json:"pipeline"`
}

func (p *pipelineRealizer) Realize(ctx context.Context, pipeline *v1alpha1.Pipeline, logger logr.Logger, repository repository.Repository) (*v1.Condition, templates.Outputs, *unstructured.Unstructured) {
	pipeline.Spec.RunTemplateRef.Kind = "RunTemplate"
	if pipeline.Spec.RunTemplateRef.Namespace == "" {
		pipeline.Spec.RunTemplateRef.Namespace = pipeline.Namespace
	}
	template, err := repository.GetRunTemplate(pipeline.Spec.RunTemplateRef)

	if err != nil {
		errorMessage := fmt.Sprintf("could not get RunTemplate '%s'", pipeline.Spec.RunTemplateRef.Name)
		logger.Error(err, errorMessage)

		return RunTemplateMissingCondition(fmt.Errorf("%s: %w", errorMessage, err)), nil, nil
	}

	labels := map[string]string{
		"carto.run/pipeline-name":          pipeline.Name,
		"carto.run/pipeline-namespace":     pipeline.Namespace,
		"carto.run/run-template-name":      template.GetName(),
		"carto.run/run-template-namespace": pipeline.Spec.RunTemplateRef.Namespace,
	}

	stampContext := templates.StamperBuilder(
		pipeline,
		TemplatingContext{
			Pipeline: pipeline,
		},
		labels,
	)

	stampedObject, err := stampContext.Stamp(ctx, template.GetResourceTemplate())
	if err != nil {
		errorMessage := "could not stamp template"
		logger.Error(err, errorMessage)
		return TemplateStampFailureCondition(fmt.Errorf("%s: %w", errorMessage, err)), nil, nil
	}

	err = repository.EnsureObjectExistsOnCluster(stampedObject.DeepCopy(), false)
	if err != nil {
		errorMessage := "could not create object"
		logger.Error(err, errorMessage)
		return StampedObjectRejectedByAPIServerCondition(fmt.Errorf("%s: %w", errorMessage, err)), nil, nil
	}

	stampedObject.SetLabels(labels)
	allPipelineStampedObjects, err := repository.ListUnstructured(stampedObject)
	if err != nil {
		errorMessage := fmt.Sprintf("could not list pipeline objects: %s", err.Error())
		logger.Info(errorMessage)
		return OutputPathNotSatisfiedCondition(err), nil, stampedObject
	}
	// TODO: handle the case where the pipeline has changed what type of object it stamps
	// TODO: handle the case where the pipeline has changed what runTemplate pairs with

	outputs, err := template.GetOutput(pipeline.Status.Outputs, allPipelineStampedObjects)

	if err != nil {
		errorMessage := fmt.Sprintf("could not get output: %s", err.Error())
		logger.Info(errorMessage)
		return OutputPathNotSatisfiedCondition(err), nil, stampedObject
	}

	return RunTemplateReadyCondition(), outputs, stampedObject
}
