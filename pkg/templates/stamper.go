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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/cartographer/pkg/eval"
)

type Inputs struct {
	Sources []SourceInput `json:"sources"`
	Images  []ImageInput  `json:"images"`
	Configs []ConfigInput `json:"configs"`
}

type Labels map[string]string

// JsonPathContext is any structure that you intend for jsonpath to treat as it's context.
// typically any struct with template-specific json structure tags
type JsonPathContext interface{}

type Stamper struct {
	TemplatingContext JsonPathContext
	Owner             client.Object
	Labels            Labels
}

func StamperBuilder(owner client.Object, templatingContext JsonPathContext, labels Labels) Stamper {
	return Stamper{
		TemplatingContext: templatingContext,
		Owner:             owner,
		Labels:            labels,
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

func (s *Stamper) recursivelyEvaluateTemplates(jsonValue interface{}, loopDetector loopDetector) (interface{}, error) {
	switch typedJSONValue := jsonValue.(type) {
	case string:
		stamperTagInterpolator := StandardTagInterpolator{
			Context:   s.TemplatingContext,
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
			return s.recursivelyEvaluateTemplates(stampedLeafNode, loopDetector)
		}
	case map[string]interface{}:
		stampedMap := make(map[string]interface{})
		for key, value := range typedJSONValue {
			stampedValue, err := s.recursivelyEvaluateTemplates(value, loopDetector)
			if err != nil {
				return nil, fmt.Errorf("interpolating map value %v: %w", value, err)
			}
			stampedMap[key] = stampedValue
		}
		return stampedMap, nil
	case []interface{}:
		var stampedSlice []interface{}
		for _, sliceElement := range typedJSONValue {
			stampedElement, err := s.recursivelyEvaluateTemplates(sliceElement, loopDetector)
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

func (s *Stamper) Stamp(resourceTemplate []byte) (*unstructured.Unstructured, error) {
	var resourceTemplateJSON interface{}
	err := json.Unmarshal(resourceTemplate, &resourceTemplateJSON)
	if err != nil {
		return nil, fmt.Errorf("unmarshal to JSON: %w", err)
	}

	stampedObjectJSON, err := s.recursivelyEvaluateTemplates(resourceTemplateJSON, loopDetector{})
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
		stampedObject.SetNamespace(s.Owner.GetNamespace())
	}

	apiVersion, kind := s.Owner.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()
	stampedObject.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         apiVersion,
			Kind:               kind,
			UID:                s.Owner.GetUID(),
			Name:               s.Owner.GetName(),
			BlockOwnerDeletion: pointer.BoolPtr(true),
			Controller:         pointer.BoolPtr(true),
		},
	})

	s.mergeLabels(stampedObject)

	return stampedObject, nil
}

func (s *Stamper) mergeLabels(obj *unstructured.Unstructured) {
	labels := obj.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}

	for key, value := range s.Labels {
		labels[key] = value
	}

	obj.SetLabels(labels)
}
