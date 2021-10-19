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
	"encoding/json"
	"fmt"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/eval"
)

type Outputs map[string]apiextensionsv1.JSON

type ClusterRunTemplate interface {
	GetName() string
	GetResourceTemplate() v1alpha1.TemplateSpec
	GetOutput(stampedObjects []*unstructured.Unstructured) (Outputs, error)
}

type runTemplate struct {
	template *v1alpha1.ClusterRunTemplate
}

func (t runTemplate) GetOutput(stampedObjects []*unstructured.Unstructured) (Outputs, error) {
	var (
		updateError                        error
		everyObjectErrored                 bool
		mostRecentlySubmittedSuccesfulTime *time.Time
	)

	outputs := Outputs{}

	evaluator := eval.EvaluatorBuilder()

	everyObjectErrored = true

	for _, stampedObject := range stampedObjects {
		objectErr, provisionalOutputs := t.getOutputsOfSingleObject(evaluator, *stampedObject)

		statusPath := `status.conditions[?(@.type=="Succeeded")].status`
		status, err := evaluator.EvaluateJsonPath(statusPath, stampedObject.UnstructuredContent())
		if err != nil {
			updateError = objectErr
			continue
		}

		if status == "True" && objectErr == nil {
			objectCreationTimestamp, err := getCreationTimestamp(stampedObject, evaluator)
			if err != nil {
				continue
			}

			if mostRecentlySubmittedSuccesfulTime == nil {
				mostRecentlySubmittedSuccesfulTime = objectCreationTimestamp
			} else if objectCreationTimestamp.After(*mostRecentlySubmittedSuccesfulTime) {
				mostRecentlySubmittedSuccesfulTime = objectCreationTimestamp
			} else {
				continue
			}

			outputs = provisionalOutputs
		}

		if objectErr != nil {
			updateError = objectErr
		} else {
			everyObjectErrored = false
		}
	}

	if everyObjectErrored {
		return nil, updateError
	}

	return outputs, nil
}

func getCreationTimestamp(stampedObject *unstructured.Unstructured, evaluator evaluator) (*time.Time, error) {
	creationTimestamp, err := evaluator.EvaluateJsonPath("metadata.creationTimestamp", stampedObject.UnstructuredContent())
	if err != nil {
		return nil, err
	}
	creationTimeString, ok := creationTimestamp.(string)
	if !ok {
		return nil, err
	}
	creationTime, err := time.Parse(time.RFC3339, creationTimeString)
	if err != nil {
		return nil, fmt.Errorf("parse creation timestamp: %w", err)
	}
	return &creationTime, nil
}

func (t runTemplate) getOutputsOfSingleObject(evaluator eval.Evaluator, stampedObject unstructured.Unstructured) (error, Outputs) {
	var objectErr error
	provisionalOutputs := Outputs{}
	for key, path := range t.template.Spec.Outputs {
		output, err := evaluator.EvaluateJsonPath(path, stampedObject.UnstructuredContent())
		if err != nil {
			objectErr = fmt.Errorf("get output: %w", err)
			continue
		}

		result, err := json.Marshal(output)
		if err != nil {
			objectErr = fmt.Errorf("get output could not marshal jsonpath output: %w", err)
			continue
		}

		ext := apiextensionsv1.JSON{Raw: result}
		provisionalOutputs[key] = ext
	}
	return objectErr, provisionalOutputs
}

func NewRunTemplateModel(template *v1alpha1.ClusterRunTemplate) ClusterRunTemplate {
	return &runTemplate{template: template}
}

func (t runTemplate) GetName() string {
	return t.template.Name
}

func (t runTemplate) GetResourceTemplate() v1alpha1.TemplateSpec {
	return v1alpha1.TemplateSpec{
		Template: &t.template.Spec.Template,
	}
}
