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

package errors

import (
	"fmt"
	"strings"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

const NoJsonpathContext = "<no jsonpath context>"

const SupplyChain = "supply chain"
const Delivery = "delivery"

type GetTemplateError struct {
	Err           error
	ResourceName  string
	TemplateName  string
	BlueprintName string
	BlueprintType string
}

func (e GetTemplateError) Error() string {
	return fmt.Errorf("unable to get template [%s] for resource [%s] in %s [%s]: %w",
		e.TemplateName,
		e.ResourceName,
		e.BlueprintType,
		e.BlueprintName,
		e.Err,
	).Error()
}

type ResolveTemplateOptionError struct {
	Err           error
	ResourceName  string
	OptionName    string
	BlueprintName string
	BlueprintType string
}

func (e ResolveTemplateOptionError) Error() string {
	return fmt.Errorf("error matching against template option [%s] for resource [%s] in %s [%s]: %w",
		e.OptionName,
		e.ResourceName,
		e.BlueprintType,
		e.BlueprintName,
		e.Err,
	).Error()
}

type TemplateOptionsMatchError struct {
	ResourceName  string
	OptionNames   []string
	BlueprintName string
	BlueprintType string
}

func (e TemplateOptionsMatchError) Error() string {
	var optionNamesList string
	if len(e.OptionNames) != 0 {
		optionNamesList = "[" + strings.Join(e.OptionNames, ", ") + "] "
	}
	return fmt.Errorf("expected exactly 1 option to match, found [%d] matching options %sfor resource [%s] in %s [%s]",
		len(e.OptionNames),
		optionNamesList,
		e.ResourceName,
		e.BlueprintType,
		e.BlueprintName,
	).Error()
}

type ApplyStampedObjectError struct {
	Err           error
	StampedObject *unstructured.Unstructured
	ResourceName  string
	BlueprintName string
	BlueprintType string
}

func (e ApplyStampedObjectError) Error() string {
	return fmt.Errorf("unable to apply object [%s/%s] for resource [%s] in %s [%s]: %w",
		e.StampedObject.GetNamespace(),
		e.StampedObject.GetName(),
		e.ResourceName,
		e.BlueprintType,
		e.BlueprintName,
		e.Err,
	).Error()
}

type StampError struct {
	Err           error
	ResourceName  string
	TemplateName  string
	TemplateKind  string
	BlueprintName string
	BlueprintType string
}

func (e StampError) Error() string {
	return fmt.Errorf("unable to stamp object for resource [%s] for template [%s/%s] in %s [%s]: %w",
		e.ResourceName,
		e.TemplateKind,
		e.TemplateName,
		e.BlueprintType,
		e.BlueprintName,
		e.Err,
	).Error()
}

type RetrieveOutputError struct {
	Err           error
	ResourceName  string
	StampedObject *unstructured.Unstructured
	BlueprintName string
	BlueprintType string
}

// TODO: Part 2
// utils.GetFullyQualifiedType is wrong
func (e RetrieveOutputError) Error() string {
	if e.JsonPathExpression() == NoJsonpathContext {
		return fmt.Errorf("unable to retrieve outputs from stamped object [%s/%s] of type [%s] for resource [%s] in %s [%s]: %w",
			e.StampedObject.GetNamespace(),
			e.StampedObject.GetName(),
			utils.GetFullyQualifiedType(e.StampedObject),
			e.ResourceName,
			e.BlueprintType,
			e.BlueprintName,
			e.Err,
		).Error()
	}
	return fmt.Errorf("unable to retrieve outputs [%s] from stamped object [%s/%s] of type [%s] for resource [%s] in %s [%s]: %w",
		e.JsonPathExpression(),
		e.StampedObject.GetNamespace(),
		e.StampedObject.GetName(),
		utils.GetFullyQualifiedType(e.StampedObject),
		e.ResourceName,
		e.BlueprintType,
		e.BlueprintName,
		e.Err,
	).Error()
}

func (e RetrieveOutputError) JsonPathExpression() string {
	jsonPathErrorContext, ok := e.Err.(JsonPathErrorContext)
	if ok {
		return jsonPathErrorContext.JsonPathExpression()
	}
	return NoJsonpathContext
}

type JsonPathErrorContext interface {
	JsonPathExpression() string
}

func (e RetrieveOutputError) GetResourceName() string {
	return e.ResourceName
}

func WrapUnhandledError(err error) error {
	if IsUnhandledErrorType(err) {
		return NewUnhandledError(err)
	}
	return err
}

func IsUnhandledErrorType(err error) bool {
	switch typedErr := err.(type) {
	case GetTemplateError:
		return true
	case ApplyStampedObjectError:
		if !kerrors.IsForbidden(typedErr.Err) {
			return true
		} else {
			return false
		}
	case StampError, RetrieveOutputError, ResolveTemplateOptionError, TemplateOptionsMatchError:
		return false
	default:
		return true
	}
}
