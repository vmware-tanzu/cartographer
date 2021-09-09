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

type TemplatingContext struct {
	Pipeline *v1alpha1.Pipeline `json:"pipeline"`
}

func (p *pipelineRealizer) Realize(pipeline *v1alpha1.Pipeline, logger logr.Logger, repository repository.Repository) *v1.Condition {
	pipeline.Spec.RunTemplateRef.Kind = "RunTemplate"
	if pipeline.Spec.RunTemplateRef.Namespace == "" {
		pipeline.Spec.RunTemplateRef.Namespace = pipeline.Namespace
	}
	template, err := repository.GetTemplate(pipeline.Spec.RunTemplateRef)

	if err != nil {
		errorMessage := fmt.Sprintf("could not get RunTemplate '%s'", pipeline.Spec.RunTemplateRef.Name)
		logger.Error(err, errorMessage)

		return RunTemplateMissingCondition(fmt.Errorf("%s: %w", errorMessage, err))
	}

	labels := map[string]string{}

	stampContext := templates.StamperBuilder(
		pipeline,
		TemplatingContext{
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
