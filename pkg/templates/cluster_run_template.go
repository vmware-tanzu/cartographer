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

const StatusPath = `status.conditions[?(@.type=="Succeeded")].status`

// Fixme: do we want a ptr receiver?
func (t runTemplate) GetLatestSuccessfulOutput(stampedObjects []*unstructured.Unstructured) (Outputs, *unstructured.Unstructured, error) {
	var (
		latestTime           time.Time
		latestMatchingObject *unstructured.Unstructured
	)

	latestOutputs := Outputs{}

	for _, stampedObject := range stampedObjects {
		matched, currentOutputs := t.matchOutputs(stampedObject)
		if !matched {
			continue
		}

		currentTime := stampedObject.GetCreationTimestamp().Time
		if currentTime.After(latestTime) {
			latestMatchingObject = stampedObject
			latestOutputs = currentOutputs
		}
	}

	return latestOutputs, latestMatchingObject, nil
}

func (t runTemplate) matchOutputs(stampedObject *unstructured.Unstructured) (bool, Outputs) {
	status, err := t.evaluator.EvaluateJsonPath(StatusPath, stampedObject.UnstructuredContent())
	if err != nil {
		return false, Outputs{}
	}

	if status == "True" {
		outputError, outputs := t.getOutputsOfSingleObject(t.evaluator, *stampedObject)
		if outputError != nil {
			return true, Outputs{}
		}

		return true, outputs
	}
	return false, Outputs{}
}

func (t runTemplate) getOutputsOfSingleObject(evaluator eval.Evaluator, stampedObject unstructured.Unstructured) (error, Outputs) {
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

func (t runTemplate) GetName() string {
	return t.template.Name
}

func (t runTemplate) GetResourceTemplate() v1alpha1.TemplateSpec {
	return v1alpha1.TemplateSpec{
		Template: &t.template.Spec.Template,
	}
}
