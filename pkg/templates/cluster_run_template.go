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

func NewRunTemplateModel(template *v1alpha1.ClusterRunTemplate) ClusterRunTemplate {
	return &runTemplate{
		template:  template,
		evaluator: eval.EvaluatorBuilder(),
	}
}

type Outputs map[string]apiextensionsv1.JSON

type ClusterRunTemplate interface {
	GetName() string
	GetResourceTemplate() v1alpha1.TemplateSpec
	GetLatestSuccessfulOutput(stampedObjects []*unstructured.Unstructured) (Outputs, *unstructured.Unstructured, error)
}

type runTemplate struct {
	template  *v1alpha1.ClusterRunTemplate
	evaluator eval.Evaluator
}

const SuccessStatusPath = `status.conditions[?(@.type=="Succeeded")].status`

// GetLatestSuccessfulOutput returns the most recent condition:Succeeded=True stamped object.
// If no output paths are specified, then you only receive the object and empty outputs.
// If the output path is specified but doesn't match anything in the latest "suceeded" object, then an error is returned
// along with the matched object.
// if the output paths are all satisfied, then the outputs from the latest object, and the object itself, are returned.
func (t *runTemplate) GetLatestSuccessfulOutput(stampedObjects []*unstructured.Unstructured) (Outputs, *unstructured.Unstructured, error) {
	latestMatchingObject := t.getLatestSuccessfulObject(stampedObjects)

	if latestMatchingObject == nil {
		return Outputs{}, nil, nil
	}

	outputError, outputs := t.getOutputsOfSingleObject(t.evaluator, *latestMatchingObject)

	return outputs, latestMatchingObject, outputError
}

func (t *runTemplate) getLatestSuccessfulObject(stampedObjects []*unstructured.Unstructured) *unstructured.Unstructured {
	var (
		latestTime           time.Time // zero value is used for comparison
		latestMatchingObject *unstructured.Unstructured
	)

	for _, stampedObject := range stampedObjects {
		status, err := t.evaluator.EvaluateJsonPath(SuccessStatusPath, stampedObject.UnstructuredContent())
		if !(err == nil && status == "True") {
			continue
		}

		currentTime := stampedObject.GetCreationTimestamp().Time
		if currentTime.After(latestTime) {
			latestMatchingObject = stampedObject
			latestTime = currentTime
		}

	}
	return latestMatchingObject
}

func (t *runTemplate) getOutputsOfSingleObject(evaluator eval.Evaluator, stampedObject unstructured.Unstructured) (error, Outputs) {
	var objectErr error
	provisionalOutputs := Outputs{}
	for key, path := range t.template.Spec.Outputs {
		output, err := evaluator.EvaluateJsonPath(path, stampedObject.UnstructuredContent())
		//TODO: get this path out to the user in case of error
		if err != nil {
			objectErr = fmt.Errorf("failed to evaluate path [%s]: %w", path, err)
			continue
		}

		result, err := json.Marshal(output)
		if err != nil {
			objectErr = fmt.Errorf("failed to marshal output for key [%s]: %w", key, err)
			continue
		}

		ext := apiextensionsv1.JSON{Raw: result}
		provisionalOutputs[key] = ext
	}
	return objectErr, provisionalOutputs
}

func (t *runTemplate) GetName() string {
	return t.template.Name
}

func (t *runTemplate) GetResourceTemplate() v1alpha1.TemplateSpec {
	return v1alpha1.TemplateSpec{
		Template: &t.template.Spec.Template,
		Ytt:      t.template.Spec.Ytt,
	}
}
