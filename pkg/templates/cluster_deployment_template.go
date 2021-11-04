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
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

type clusterDeploymentTemplate struct {
	template  *v1alpha1.ClusterDeploymentTemplate
	evaluator evaluator
}

func (t clusterDeploymentTemplate) GetKind() string {
	return t.template.Kind
}

func NewClusterDeploymentTemplateModel(template *v1alpha1.ClusterDeploymentTemplate, eval evaluator) *clusterDeploymentTemplate {
	return &clusterDeploymentTemplate{template: template, evaluator: eval}
}

func (t clusterDeploymentTemplate) GetName() string {
	return t.template.Name
}

func (t clusterDeploymentTemplate) GetOutput(stampedObject *unstructured.Unstructured, templatingContext map[string]interface{}) (*Output, error) {
	if err := t.outputReady(stampedObject); err != nil {
		return nil, err
	}

	output := &Output{Source: &Source{}}

	originalSource, ok := templatingContext["source"].(*SourceInput)
	if !ok {
		return nil, fmt.Errorf("original source not found in context: %v", templatingContext)
	}

	output.Source.URL = originalSource.URL
	output.Source.Revision = originalSource.Revision

	return output, nil
}

func (t clusterDeploymentTemplate) GetResourceTemplate() v1alpha1.TemplateSpec {
	return t.template.Spec.TemplateSpec
}

func (t clusterDeploymentTemplate) GetDefaultParams() v1alpha1.DefaultParams {
	return t.template.Spec.Params
}

func (t clusterDeploymentTemplate) outputReady(stampedObject *unstructured.Unstructured) error {
	if t.template.Spec.ObservedCompletion != nil {
		return t.observedCompletionReady(stampedObject)
	} else {
		return t.observedMatchesReady(stampedObject)
	}
}

func (t clusterDeploymentTemplate) observedMatchesReady(stampedObject *unstructured.Unstructured) error {
	for _, match := range t.template.Spec.ObservedMatches {
		input, err := t.evaluator.EvaluateJsonPath(match.Input, stampedObject.UnstructuredContent())
		if err != nil {
			return DeploymentConditionError{
				Err: fmt.Errorf("could not find value at key '%s': %w", match.Input, err),
			}
		}

		output, err := t.evaluator.EvaluateJsonPath(match.Output, stampedObject.UnstructuredContent())
		if err != nil {
			return DeploymentConditionError{
				Err: fmt.Errorf("could not find value at key '%s': %w", match.Output, err),
			}
		}

		if input != output {
			return DeploymentConditionError{
				Err: fmt.Errorf("expected '%s' to match '%s'", input, output),
			}
		}
	}

	return nil
}

func (t clusterDeploymentTemplate) observedCompletionReady(stampedObject *unstructured.Unstructured) error {
	generation, err := t.evaluator.EvaluateJsonPath("metadata.generation", stampedObject.UnstructuredContent())
	if err != nil {
		return JsonPathError{
			Err:        fmt.Errorf("generation json path: %w", err),
			expression: "metadata.generation",
		}
	}

	observedGeneration, err := t.evaluator.EvaluateJsonPath("status.observedGeneration", stampedObject.UnstructuredContent())
	if err != nil {
		return ObservedGenerationError{
			Err: fmt.Errorf("observed generation json path: %w", err),
		}
	}

	if observedGeneration != generation {
		return DeploymentConditionError{
			Err: fmt.Errorf("observedGeneration does not equal generation"),
		}
	}

	if t.template.Spec.ObservedCompletion.FailedCondition != nil {
		failedObserved, _ := t.evaluator.EvaluateJsonPath(t.template.Spec.ObservedCompletion.FailedCondition.Key, stampedObject.UnstructuredContent())

		if failedObserved == t.template.Spec.ObservedCompletion.FailedCondition.Value {
			return DeploymentFailedConditionMetError{
				Err: fmt.Errorf("'%s' was '%s'", t.template.Spec.ObservedCompletion.FailedCondition.Key, failedObserved),
			}
		}
	}

	succeededObserved, err := t.evaluator.EvaluateJsonPath(t.template.Spec.ObservedCompletion.SucceededCondition.Key, stampedObject.UnstructuredContent())
	if err != nil {
		return DeploymentConditionError{
			Err: fmt.Errorf("could not find value at key '%s': %w", t.template.Spec.ObservedCompletion.SucceededCondition.Key, err),
		}
	}

	if succeededObserved != t.template.Spec.ObservedCompletion.SucceededCondition.Value {
		return DeploymentConditionError{
			Err: fmt.Errorf("expected '%s' to be '%s' but found '%s'", t.template.Spec.ObservedCompletion.SucceededCondition.Key, t.template.Spec.ObservedCompletion.SucceededCondition.Value, succeededObserved),
		}
	}

	return nil
}
