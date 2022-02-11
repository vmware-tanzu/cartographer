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

package deliverable

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

const NoJsonpathContext = "<no jsonpath context>"

type GetDeliveryTemplateError struct {
	Err          error
	DeliveryName string
	Resource     *v1alpha1.DeliveryResource
}

func (e GetDeliveryTemplateError) Error() string {
	return fmt.Errorf("unable to get template [%s] for resource [%s] in delivery [%s]: %w",
		e.Resource.TemplateRef.Name,
		e.Resource.Name,
		e.DeliveryName,
		e.Err,
	).Error()
}

type ResolveTemplateOptionError struct {
	Err          error
	DeliveryName string
	Resource     *v1alpha1.DeliveryResource
	OptionName   string
	Key          string
}

func (e ResolveTemplateOptionError) Error() string {
	return fmt.Errorf("key [%s] is invalid in template option [%s] for resource [%s] in delivery [%s]: %w",
		e.Key,
		e.OptionName,
		e.Resource.Name,
		e.DeliveryName,
		e.Err,
	).Error()
}

type TemplateOptionsMatchError struct {
	DeliveryName string
	Resource     *v1alpha1.DeliveryResource
	OptionNames  []string
}

func (e TemplateOptionsMatchError) Error() string {
	var optionNamesList string
	if len(e.OptionNames) != 0 {
		optionNamesList = "[" + strings.Join(e.OptionNames, ", ") + "] "
	}
	return fmt.Errorf("expected exactly 1 option to match, found [%d] matching options %sfor resource [%s] in delivery [%s]",
		len(e.OptionNames),
		optionNamesList,
		e.Resource.Name,
		e.DeliveryName,
	).Error()
}

type ApplyStampedObjectError struct {
	Err           error
	DeliveryName  string
	StampedObject *unstructured.Unstructured
	Resource      *v1alpha1.DeliveryResource
}

func (e ApplyStampedObjectError) Error() string {
	return fmt.Errorf("unable to apply object [%s/%s] for resource [%s] in delivery [%s]: %w",
		e.StampedObject.GetNamespace(),
		e.StampedObject.GetName(),
		e.Resource.Name,
		e.DeliveryName,
		e.Err,
	).Error()
}

type StampError struct {
	Err          error
	DeliveryName string
	Resource     *v1alpha1.DeliveryResource
}

func (e StampError) Error() string {
	return fmt.Errorf("unable to stamp object for resource [%s] in delivery [%s]: %w",
		e.Resource.Name,
		e.DeliveryName,
		e.Err,
	).Error()
}

type RetrieveOutputError struct {
	Err           error
	DeliveryName  string
	Resource      *v1alpha1.DeliveryResource
	StampedObject *unstructured.Unstructured
}

type JsonPathErrorContext interface {
	JsonPathExpression() string
}

func (e RetrieveOutputError) Error() string {
	if e.JsonPathExpression() == NoJsonpathContext {
		return fmt.Errorf("unable to retrieve outputs from stamped object [%s/%s] of type [%s] for resource [%s] in delivery [%s]: %w",
			e.StampedObject.GetNamespace(),
			e.StampedObject.GetName(),
			utils.GetFullyQualifiedType(e.StampedObject),
			e.Resource.Name,
			e.DeliveryName,
			e.Err,
		).Error()
	}
	return fmt.Errorf("unable to retrieve outputs [%s] from stamped object [%s/%s] of type [%s] for resource [%s] in delivery [%s]: %w",
		e.JsonPathExpression(),
		e.StampedObject.GetNamespace(),
		e.StampedObject.GetName(),
		utils.GetFullyQualifiedType(e.StampedObject),
		e.Resource.Name,
		e.DeliveryName,
		e.Err,
	).Error()
}

func (e RetrieveOutputError) ResourceName() string {
	return e.Resource.Name
}

func (e RetrieveOutputError) JsonPathExpression() string {
	jsonPathErrorContext, ok := e.Err.(JsonPathErrorContext)
	if ok {
		return jsonPathErrorContext.JsonPathExpression()
	}
	return NoJsonpathContext
}
