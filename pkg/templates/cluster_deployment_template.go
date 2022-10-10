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

package templates

import (
	"crypto/sha256"
	"fmt"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/strings"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/eval"
)

type clusterDeploymentTemplate struct {
	template  *v1alpha1.ClusterDeploymentTemplate
	evaluator evaluator
	inputs    Inputs
}

func NewClusterDeploymentTemplateModel(template *v1alpha1.ClusterDeploymentTemplate, eval evaluator) *clusterDeploymentTemplate {
	return &clusterDeploymentTemplate{template: template, evaluator: eval}
}

func (t *clusterDeploymentTemplate) SetInputs(inputs Inputs) {
	t.inputs = inputs
}

func (t *clusterDeploymentTemplate) GetOutput(stampedObject *unstructured.Unstructured) (*Output, error) {
	if err := t.outputReady(stampedObject); err != nil {
		return nil, err
	}

	output := &Output{Source: &Source{}}

	deployment := t.inputs.GetDeployment()
	if deployment == nil {
		return nil, fmt.Errorf("deployment not found in upstream template")
	}

	output.Source.URL = deployment.URL
	output.Source.Revision = deployment.Revision

	return output, nil
}

func (t *clusterDeploymentTemplate) GenerateResourceOutput(output *Output) ([]v1alpha1.Output, error) {
	if output == nil || output.Source == nil {
		return nil, nil
	}

	urlBytes, err := yaml.Marshal(output.Source.URL)
	if err != nil {
		return nil, err
	}

	urlSHA := sha256.Sum256(urlBytes)

	revBytes, err := yaml.Marshal(output.Source.Revision)
	if err != nil {
		return nil, err
	}

	revSHA := sha256.Sum256(revBytes)

	return []v1alpha1.Output{
		{
			Name:    "url",
			Preview: strings.ShortenString(string(urlBytes), PREVIEW_CHARACTER_LIMIT),
			Digest:  fmt.Sprintf("sha256:%x", urlSHA),
		},
		{
			Name:    "revision",
			Preview: strings.ShortenString(string(revBytes), PREVIEW_CHARACTER_LIMIT),
			Digest:  fmt.Sprintf("sha256:%x", revSHA),
		},
	}, nil
}

func (t *clusterDeploymentTemplate) GetResourceTemplate() v1alpha1.TemplateSpec {
	return t.template.Spec.TemplateSpec
}

func (t *clusterDeploymentTemplate) GetDefaultParams() v1alpha1.TemplateParams {
	return t.template.Spec.Params
}

func (t *clusterDeploymentTemplate) outputReady(stampedObject *unstructured.Unstructured) error {
	if t.template.Spec.ObservedCompletion != nil {
		return t.observedCompletionReady(stampedObject)
	} else {
		return t.observedMatchesReady(stampedObject)
	}
}

func (t *clusterDeploymentTemplate) observedMatchesReady(stampedObject *unstructured.Unstructured) error {
	for _, match := range t.template.Spec.ObservedMatches {
		input, err := t.evaluator.EvaluateJsonPath(match.Input, stampedObject.UnstructuredContent())
		if err != nil {
			return DeploymentConditionError{
				Err: fmt.Errorf("could not find value on input [%s]: %w", match.Input, err),
			}
		}

		output, err := t.evaluator.EvaluateJsonPath(match.Output, stampedObject.UnstructuredContent())
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

func (t *clusterDeploymentTemplate) observedCompletionReady(stampedObject *unstructured.Unstructured) error {
	generation, err := t.evaluator.EvaluateJsonPath("metadata.generation", stampedObject.UnstructuredContent())
	if err != nil {
		return JsonPathError{
			Err:        fmt.Errorf("failed to evaluate metadata.generation: %w", err),
			expression: "metadata.generation",
		}
	}

	observedGeneration, err := t.evaluator.EvaluateJsonPath("status.observedGeneration",
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

	observedCompletion := t.template.Spec.ObservedCompletion
	if t.template.Spec.ObservedCompletion.FailedCondition != nil {
		failedObserved, err := t.evaluator.EvaluateJsonPath(observedCompletion.FailedCondition.Key, stampedObject.UnstructuredContent())
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

	succeededObserved, err := t.evaluator.EvaluateJsonPath(observedCompletion.SucceededCondition.Key, stampedObject.UnstructuredContent())
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

func (t *clusterDeploymentTemplate) GetHealthRule() *v1alpha1.HealthRule {
	return t.template.Spec.HealthRule
}
