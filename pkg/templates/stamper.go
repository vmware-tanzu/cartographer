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

	"github.com/valyala/fasttemplate"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/pointer"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/eval"
)

type Inputs struct {
	Sources []SourceInput `json:"sources"`
	Images  []ImageInput  `json:"images"`
	Configs []ConfigInput `json:"configs"`
}

type StampContext struct {
	Params   Params             `json:"params"`
	Workload *v1alpha1.Workload `json:"workload"`
	Sources  []SourceInput      `json:"sources"`
	Images   []ImageInput       `json:"images"`
	Configs  []ConfigInput      `json:"configs"`

	Labels map[string]string
}

func StampContextBuilder(workload *v1alpha1.Workload, labels map[string]string, params Params, inputs *Inputs) StampContext {
	return StampContext{
		Params:   params,
		Workload: workload,
		Sources:  inputs.Sources,
		Images:   inputs.Images,
		Configs:  inputs.Configs,

		Labels: labels,
	}
}

type loopDetector []string

func (d loopDetector) checkItem(item string) (loopDetector, error) {
	var potentialLoop []string
	newStack := append(d, item)
	for _, currentItem := range newStack {
		if currentItem == item {
			potentialLoop = append(potentialLoop, currentItem)
		} else if len(potentialLoop) > 0 {
			potentialLoop = append(potentialLoop, currentItem)
		}
	}
	if len(potentialLoop) > 1 {
		return newStack, fmt.Errorf("infinite tag loop detected: %+v", formatExpressionLoop(potentialLoop))
	}

	return newStack, nil
}

func formatExpressionLoop(expressionLoop []string) string {
	result := ""
	for _, expression := range expressionLoop {
		if result != "" {
			result = result + " -> "
		}
		result = result + expression
	}
	return result
}

func (c *StampContext) recursivelyEvaluateTemplates(jsonValue interface{}, loopDetector loopDetector) (interface{}, error) {
	switch typedJSONValue := jsonValue.(type) {
	case string:
		stamperTagInterpolator := StandardTagInterpolator{
			Context:   *c,
			Evaluator: eval.EvaluatorBuilder(),
		}
		loopDetector, err := loopDetector.checkItem(typedJSONValue)
		if err != nil {
			return nil, err
		}

		stampedLeafNode, err := InterpolateLeafNode(fasttemplate.ExecuteFuncStringWithErr, []byte(typedJSONValue), stamperTagInterpolator)
		if err != nil {
			return nil, fmt.Errorf("interpolating: %w", err)
		}
		if jsonValue == stampedLeafNode {
			return stampedLeafNode, nil
		} else {
			return c.recursivelyEvaluateTemplates(stampedLeafNode, loopDetector)
		}
	case map[string]interface{}:
		stampedMap := make(map[string]interface{})
		for key, value := range typedJSONValue {
			stampedValue, err := c.recursivelyEvaluateTemplates(value, loopDetector)
			if err != nil {
				return nil, fmt.Errorf("interpolating map value %v: %w", value, err)
			}
			stampedMap[key] = stampedValue
		}
		return stampedMap, nil
	case []interface{}:
		var stampedSlice []interface{}
		for _, sliceElement := range typedJSONValue {
			stampedElement, err := c.recursivelyEvaluateTemplates(sliceElement, loopDetector)
			if err != nil {
				return nil, fmt.Errorf("interpolating map value %v: %w", sliceElement, err)
			}
			stampedSlice = append(stampedSlice, stampedElement)
		}
		return stampedSlice, nil
	default:
		return typedJSONValue, nil
	}
}

func (c *StampContext) Stamp(resourceTemplate []byte) (*unstructured.Unstructured, error) {
	var resourceTemplateJSON interface{}
	err := json.Unmarshal(resourceTemplate, &resourceTemplateJSON)
	if err != nil {
		return nil, fmt.Errorf("unmarshal to JSON: %w", err)
	}

	stampedObjectJSON, err := c.recursivelyEvaluateTemplates(resourceTemplateJSON, loopDetector{})
	if err != nil {
		return nil, fmt.Errorf("recursively stamp json values: %w", err)
	}

	unstructuredContent, ok := stampedObjectJSON.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("stamped resource is not a map[string]interface{}: %+v", stampedObjectJSON)
	}
	stampedObject := &unstructured.Unstructured{}
	stampedObject.SetUnstructuredContent(unstructuredContent)

	if stampedObject.GetNamespace() == "" {
		stampedObject.SetNamespace(c.Workload.GetNamespace())
	}

	apiVersion, kind := c.Workload.TypeMeta.GroupVersionKind().ToAPIVersionAndKind()
	stampedObject.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         apiVersion,
			Kind:               kind,
			Name:               c.Workload.Name,
			UID:                c.Workload.UID,
			BlockOwnerDeletion: pointer.BoolPtr(true),
			Controller:         pointer.BoolPtr(true),
		},
	})

	c.mergeLabels(stampedObject)

	return stampedObject, nil
}

func (c *StampContext) mergeLabels(obj *unstructured.Unstructured) {
	labels := obj.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}

	for key, value := range c.Labels {
		labels[key] = value
	}

	obj.SetLabels(labels)
}
