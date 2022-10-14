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

package stamp

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/eval"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

type DeploymentInput interface {
	GetDeployment() *templates.SourceInput
}

type PassThroughInput interface {
	GetSources() map[string]templates.SourceInput
	GetImages() map[string]templates.ImageInput
	GetConfigs() map[string]templates.ConfigInput
	GetDeployment() *templates.SourceInput
}

type Reader interface {
	// fixme: output as a one-of is so weird
	GetOutput(stampedObject *unstructured.Unstructured) (*templates.Output, error)
}

func NewPassThroughReader(kind string, name string, inputReader PassThroughInput) (Reader, error) {
	switch {

	case kind == "ClusterSourceTemplate":
		return NewSourcePassThroughReader(name, inputReader), nil
	case kind == "ClusterImageTemplate":
		return NewImagePassThroughReader(name, inputReader), nil
	case kind == "ClusterConfigTemplate":
		return NewConfigPassThroughReader(name, inputReader), nil
	case kind == "ClusterTemplate":
		return NewNoOutputReader(), nil
	case kind == "ClusterDeploymentTemplate":
		return NewDeploymentPassThroughReader(inputReader, nil), nil
	}

	return nil, fmt.Errorf("kind does not match a known template")
}

func NewReader(template client.Object, inputReader DeploymentInput) (Reader, error) {
	switch v := template.(type) {

	case *v1alpha1.ClusterSourceTemplate:
		return NewSourceOutputReader(v), nil
	case *v1alpha1.ClusterImageTemplate:
		return NewImageOutputReader(v), nil
	case *v1alpha1.ClusterConfigTemplate:
		return NewConfigOutputReader(v), nil
	case *v1alpha1.ClusterDeploymentTemplate:
		return NewDeploymentPassThroughReader(inputReader, v), nil
	case *v1alpha1.ClusterTemplate:
		return NewNoOutputReader(), nil
	}
	return nil, fmt.Errorf("template does not match a known template")
}

type SourceOutputReader struct {
	template *v1alpha1.ClusterSourceTemplate
}

func (r *SourceOutputReader) GetOutput(stampedObject *unstructured.Unstructured) (*templates.Output, error) {
	// TODO: We don't need a Builder
	evaluator := eval.EvaluatorBuilder()
	url, err := evaluator.EvaluateJsonPath(r.template.Spec.URLPath, stampedObject.UnstructuredContent())
	if err != nil {
		return nil, JsonPathError{
			Err: fmt.Errorf("failed to evaluate the url path [%s]: %w",
				r.template.Spec.URLPath, err),
			expression: r.template.Spec.URLPath,
		}
	}

	revision, err := evaluator.EvaluateJsonPath(r.template.Spec.RevisionPath, stampedObject.UnstructuredContent())
	if err != nil {
		return nil, JsonPathError{
			Err: fmt.Errorf("failed to evaluate the revision path [%s]: %w",
				r.template.Spec.RevisionPath, err),
			expression: r.template.Spec.RevisionPath,
		}
	}
	return &templates.Output{
		Source: &templates.Source{
			URL:      url,
			Revision: revision,
		},
	}, nil
}

func NewSourceOutputReader(template *v1alpha1.ClusterSourceTemplate) Reader {
	return &SourceOutputReader{template: template}
}

type ConfigOutputReader struct {
	template *v1alpha1.ClusterConfigTemplate
}

func (r *ConfigOutputReader) GetOutput(stampedObject *unstructured.Unstructured) (*templates.Output, error) {
	evaluator := eval.EvaluatorBuilder()
	config, err := evaluator.EvaluateJsonPath(r.template.Spec.ConfigPath, stampedObject.UnstructuredContent())
	if err != nil {
		return nil, JsonPathError{
			Err: fmt.Errorf("failed to evaluate spec.configPath [%s]: %w",
				r.template.Spec.ConfigPath, err),
			expression: r.template.Spec.ConfigPath,
		}
	}

	return &templates.Output{
		Config: config,
	}, nil
}

func NewConfigOutputReader(template *v1alpha1.ClusterConfigTemplate) Reader {
	return &ConfigOutputReader{
		template: template,
	}
}

type ImageOutputReader struct {
	template *v1alpha1.ClusterImageTemplate
}

func (r *ImageOutputReader) GetOutput(stampedObject *unstructured.Unstructured) (*templates.Output, error) {
	evaluator := eval.EvaluatorBuilder()
	image, err := evaluator.EvaluateJsonPath(r.template.Spec.ImagePath, stampedObject.UnstructuredContent())
	if err != nil {
		return nil, JsonPathError{
			Err: fmt.Errorf("failed to evaluate the url path [%s]: %w",
				r.template.Spec.ImagePath, err),
			expression: r.template.Spec.ImagePath,
		}
	}

	return &templates.Output{
		Image: image,
	}, nil
}

func NewImageOutputReader(template *v1alpha1.ClusterImageTemplate) Reader {
	return &ImageOutputReader{
		template: template,
	}
}

type DeploymentPassThroughReader struct {
	inputs   DeploymentInput
	template *v1alpha1.ClusterDeploymentTemplate
}

func (r *DeploymentPassThroughReader) GetOutput(stampedObject *unstructured.Unstructured) (*templates.Output, error) {
	if err := r.outputReady(stampedObject); err != nil {
		return nil, err
	}

	output := &templates.Output{Source: &templates.Source{}}

	deployment := r.inputs.GetDeployment()
	if deployment == nil {
		return nil, fmt.Errorf("deployment not found in upstream template")
	}

	output.Source.URL = deployment.URL
	output.Source.Revision = deployment.Revision

	return output, nil
}

func (r *DeploymentPassThroughReader) outputReady(stampedObject *unstructured.Unstructured) error {
	if r.template == nil {
		return nil
	}
	if r.template.Spec.ObservedCompletion != nil {
		return r.observedCompletionReady(stampedObject)
	} else {
		return r.observedMatchesReady(stampedObject)
	}
}

func (r *DeploymentPassThroughReader) observedMatchesReady(stampedObject *unstructured.Unstructured) error {
	evaluator := eval.EvaluatorBuilder()

	for _, match := range r.template.Spec.ObservedMatches {
		input, err := evaluator.EvaluateJsonPath(match.Input, stampedObject.UnstructuredContent())
		if err != nil {
			return DeploymentConditionError{
				Err: fmt.Errorf("could not find value on input [%s]: %w", match.Input, err),
			}
		}

		output, err := evaluator.EvaluateJsonPath(match.Output, stampedObject.UnstructuredContent())
		if err != nil {
			return DeploymentConditionError{
				Err: fmt.Errorf("could not find value on output [%s]: %w", match.Output, err),
			}
		}

		if input != output {
			return DeploymentConditionError{
				Err: fmt.Errorf("input [%s] and output [%s] do not match: %s != %s", match.Input, match.Output, input, output),
			}
		}
	}

	return nil
}

func (r *DeploymentPassThroughReader) observedCompletionReady(stampedObject *unstructured.Unstructured) error {
	evaluator := eval.EvaluatorBuilder()

	generation, err := evaluator.EvaluateJsonPath("metadata.generation", stampedObject.UnstructuredContent())
	if err != nil {
		return JsonPathError{
			Err:        fmt.Errorf("failed to evaluate metadata.generation: %w", err),
			expression: "metadata.generation",
		}
	}

	observedGeneration, err := evaluator.EvaluateJsonPath("status.observedGeneration",
		stampedObject.UnstructuredContent())
	if err != nil {
		return ObservedGenerationError{
			Err: fmt.Errorf("failed to evaluate status.observedGeneration: %w", err),
		}
	}

	if observedGeneration != generation {
		return DeploymentConditionError{
			Err: fmt.Errorf("status.observedGeneration does not equal metadata.generation: %s != %s",
				observedGeneration, generation),
		}
	}

	observedCompletion := r.template.Spec.ObservedCompletion
	if r.template.Spec.ObservedCompletion.FailedCondition != nil {
		failedObserved, err := evaluator.EvaluateJsonPath(observedCompletion.FailedCondition.Key, stampedObject.UnstructuredContent())
		if err != nil {
			if _, ok := err.(eval.JsonPathDoesNotExistError); !ok {
				return JsonPathError{
					Err:        fmt.Errorf("failed to evaluate %s: %w", observedCompletion.FailedCondition.Key, err),
					expression: observedCompletion.FailedCondition.Key,
				}
			}
		}

		if failedObserved == observedCompletion.FailedCondition.Value {
			return DeploymentFailedConditionMetError{
				Err: fmt.Errorf("deployment failure condition [%s] was: %s",
					observedCompletion.FailedCondition.Key, failedObserved),
			}
		}
	}

	succeededObserved, err := evaluator.EvaluateJsonPath(observedCompletion.SucceededCondition.Key, stampedObject.UnstructuredContent())
	if err != nil {
		return DeploymentConditionError{
			Err: fmt.Errorf("failed to evaluate succeededCondition.Key [%s]: %w",
				observedCompletion.SucceededCondition.Key, err),
		}
	}

	if succeededObserved != observedCompletion.SucceededCondition.Value {
		return DeploymentConditionError{
			Err: fmt.Errorf("deployment success condition [%s] was: %s, expected: %s",
				observedCompletion.SucceededCondition.Key, succeededObserved, observedCompletion.SucceededCondition.Value),
		}
	}

	return nil
}

func NewDeploymentPassThroughReader(inputReader DeploymentInput, template *v1alpha1.ClusterDeploymentTemplate) Reader {
	return &DeploymentPassThroughReader{
		inputs:   inputReader,
		template: template,
	}
}

type NoOutputReader struct{}

func (r *NoOutputReader) GetOutput(_ *unstructured.Unstructured) (*templates.Output, error) {
	return &templates.Output{}, nil
}

func NewNoOutputReader() Reader {
	return &NoOutputReader{}
}

type SourcePassThroughReader struct {
	inputs PassThroughInput
	name   string
}

func NewSourcePassThroughReader(name string, inputReader PassThroughInput) Reader {
	return &SourcePassThroughReader{
		name:   name,
		inputs: inputReader,
	}
}

func (r *SourcePassThroughReader) GetOutput(_ *unstructured.Unstructured) (*templates.Output, error) {
	sources := r.inputs.GetSources()
	if _, ok := sources[r.name]; !ok {
		return nil, fmt.Errorf("input [%s] not found in sources", r.name)
	}

	return &templates.Output{
		Source: &templates.Source{
			URL:      sources[r.name].URL,
			Revision: sources[r.name].Revision,
		},
	}, nil
}

type ImagePassThroughReader struct {
	inputs PassThroughInput
	name   string
}

func NewImagePassThroughReader(name string, inputReader PassThroughInput) Reader {
	return &ImagePassThroughReader{
		name:   name,
		inputs: inputReader,
	}
}

func (r *ImagePassThroughReader) GetOutput(_ *unstructured.Unstructured) (*templates.Output, error) {
	images := r.inputs.GetImages()
	if _, ok := images[r.name]; !ok {
		return nil, fmt.Errorf("input [%s] not found in images", r.name)
	}

	return &templates.Output{
		Image: images[r.name].Image,
	}, nil
}

type ConfigPassThroughReader struct {
	inputs PassThroughInput
	name   string
}

func NewConfigPassThroughReader(name string, inputReader PassThroughInput) Reader {
	return &ConfigPassThroughReader{
		name:   name,
		inputs: inputReader,
	}
}

func (r *ConfigPassThroughReader) GetOutput(_ *unstructured.Unstructured) (*templates.Output, error) {
	config := r.inputs.GetConfigs()
	if _, ok := config[r.name]; !ok {
		return nil, fmt.Errorf("input [%s] not found in configs", r.name)
	}

	return &templates.Output{
		Config: config[r.name].Config,
	}, nil
}
