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

package v1alpha1

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/util/jsonpath"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func validateResourceOptions(options []TemplateOption, validPaths map[string]bool, validPrefixes []string) error {
	passThroughCount := 0
	for _, option := range options {
		if option.PassThrough != "" {
			passThroughCount += 1
		}
	}
	if passThroughCount > 1 {
		return fmt.Errorf("cannot have more than one pass through option, found %d", passThroughCount)
	}

	for _, option := range options {
		if option.Name != "" && option.PassThrough != "" {
			return fmt.Errorf("exactly one of option.Name or option.PassThrough must be specified, found both")
		}

		if option.Name == "" && option.PassThrough == "" {
			return fmt.Errorf("exactly one of option.Name or option.PassThrough must be specified, found neither")
		}

		optionName := option.Name
		if optionName == "" {
			optionName = "passThrough"
		}

		if err := validateSelector(option.Selector, validPaths, validPrefixes); err != nil {
			return fmt.Errorf("error validating option [%s] selector: %w", optionName, err)
		}
	}

	for _, option1 := range options {
		name1 := option1.Name
		if name1 == "" {
			name1 = "passThrough"
		}
		for _, option2 := range options {
			name2 := option2.Name
			if name2 == "" {
				name2 = "passThrough"
			}
			if option1.Name != option2.Name && reflect.DeepEqual(option1.Selector, option2.Selector) {
				return fmt.Errorf("duplicate selector found in options [%s, %s]", name1, name2)
			}
		}
	}

	return nil
}

func validateSelector(selector Selector, validPaths map[string]bool, validPrefixes []string) error {
	if len(selector.MatchLabels) == 0 && len(selector.MatchExpressions) == 0 && len(selector.MatchFields) == 0 {
		return fmt.Errorf("at least one of matchLabels, matchExpressions or MatchFields must be specified")
	}

	var err error

	labelSelector := &metav1.LabelSelector{
		MatchLabels: selector.MatchLabels,
	}

	_, err = metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return fmt.Errorf("matchLabels are not valid: %w", err)
	}

	labelSelector = &metav1.LabelSelector{
		MatchExpressions: selector.MatchExpressions,
	}

	_, err = metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return fmt.Errorf("matchExpressions are not valid: %w", err)
	}

	if len(selector.MatchFields) != 0 {
		err = validateFieldSelectorRequirements(selector.MatchFields, validPaths, validPrefixes)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateFieldSelectorRequirements(reqs []FieldSelectorRequirement, validPaths map[string]bool, validPrefixes []string) error {
	for _, req := range reqs {
		switch req.Operator {
		case FieldSelectorOpExists, FieldSelectorOpDoesNotExist:
			if len(req.Values) != 0 {
				return fmt.Errorf("cannot specify values with operator [%s]", req.Operator)
			}
		case FieldSelectorOpIn, FieldSelectorOpNotIn:
			if len(req.Values) == 0 {
				return fmt.Errorf("must specify values with operator [%s]", req.Operator)
			}
		default:
			return fmt.Errorf("operator [%s] is invalid", req.Operator)
		}

		if !validPath(req.Key, validPaths, validPrefixes) {
			return fmt.Errorf("requirement key [%s] is not a valid path", req.Key)
		}

		if err := validJsonpath(req.Key); err != nil {
			return fmt.Errorf("invalid jsonpath for key [%s]: %w", req.Key, err)
		}
	}
	return nil
}

func validJsonpath(path string) error {
	parser := jsonpath.New("")

	return parser.Parse(path)
}

func validPath(path string, validPaths map[string]bool, validPrefixes []string) bool {
	if validPaths[path] {
		return true
	}

	for _, prefix := range validPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

func validateLegacySelector(selectors LegacySelector, validPaths map[string]bool, validPrefixes []string) error {
	var err error

	labelSelector := &metav1.LabelSelector{
		MatchLabels: selectors.Selector,
	}

	_, err = metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return fmt.Errorf("selector is not valid: %w", err)
	}

	labelSelector = &metav1.LabelSelector{
		MatchExpressions: selectors.SelectorMatchExpressions,
	}

	_, err = metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return fmt.Errorf("selectorMatchExpressions are not valid: %w", err)
	}

	if len(selectors.SelectorMatchFields) != 0 {
		err = validateFieldSelectorRequirements(selectors.SelectorMatchFields, validPaths, validPrefixes)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *TemplateSpec) validate() (admission.Warnings, error) {
	if t.Template == nil && t.Ytt == "" {
		return nil, fmt.Errorf("invalid template: must specify one of template or ytt, found neither")
	}
	if t.Template != nil && t.Ytt != "" {
		return nil, fmt.Errorf("invalid template: must specify one of template or ytt, found both")
	}
	if t.Template != nil {
		obj := unstructured.Unstructured{}
		if err := json.Unmarshal(t.Template.Raw, &obj); err != nil {
			return nil, fmt.Errorf("invalid template: failed to parse object: %w", err)
		}
		if obj.GetNamespace() != metav1.NamespaceNone {
			return nil, fmt.Errorf("invalid template: template should not set metadata.namespace on the child object")
		}
	}
	if t.HealthRule != nil {
		return nil, t.HealthRule.validate()
	}

	if t.RetentionPolicy != nil && t.Lifecycle == "mutable" {
		return nil, fmt.Errorf("invalid template: if lifecycle is mutable, no retention policy may be set")
	}

	return nil, nil
}

func (r *HealthRule) validate() error {
	nRules := 0
	if r.AlwaysHealthy != nil {
		nRules++
	}
	if r.SingleConditionType != "" {
		nRules++
	}
	if r.MultiMatch != nil {
		nRules++
	}
	if nRules == 0 {
		return fmt.Errorf("invalid health rule: must specify one of alwaysHealthy, singleConditionType or multiMatch, found neither")
	}
	if nRules > 1 {
		return fmt.Errorf("invalid health rule: must specify one of alwaysHealthy, singleConditionType or multiMatch, found multiple")
	}
	if r.MultiMatch != nil {
		return r.MultiMatch.validate()
	}
	return nil
}

func (m *MultiMatchHealthRule) validate() error {
	if len(m.Unhealthy.MatchConditions) == 0 && len(m.Unhealthy.MatchFields) == 0 {
		return fmt.Errorf("invalid multi match health rule: unhealthy rule has no matchFields or matchConditions")
	}
	if len(m.Healthy.MatchConditions) == 0 && len(m.Healthy.MatchFields) == 0 {
		return fmt.Errorf("invalid multi match health rule: healthy rule has no matchFields or matchConditions")
	}
	return nil
}
