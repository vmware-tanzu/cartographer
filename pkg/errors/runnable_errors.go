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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

type RunnableGetRunTemplateError struct {
	Err         error
	TemplateRef *v1alpha1.TemplateReference
}

func (e RunnableGetRunTemplateError) Error() string {
	return fmt.Errorf("unable to get run template [%s]: %w",
		e.TemplateRef.Name,
		e.Err,
	).Error()
}

type RunnableResolveSelectorError struct {
	Err      error
	Selector *v1alpha1.ResourceSelector
}

func (e RunnableResolveSelectorError) Error() string {
	return fmt.Errorf("unable to resolve selector [%v], apiVersion [%s], kind [%s]: %w",
		e.Selector.MatchingLabels,
		e.Selector.Resource.APIVersion,
		e.Selector.Resource.Kind,
		e.Err,
	).Error()
}

type RunnableStampError struct {
	Err         error
	TemplateRef *v1alpha1.TemplateReference
}

func (e RunnableStampError) Error() string {
	return fmt.Errorf("unable to stamp object for run template [%s]: %w",
		e.TemplateRef.Name,
		e.Err,
	).Error()
}

type RunnableApplyStampedObjectError struct {
	Err           error
	StampedObject *unstructured.Unstructured
	TemplateRef   *v1alpha1.TemplateReference
}

func (e RunnableApplyStampedObjectError) Error() string {
	name := e.StampedObject.GetName()
	if name == "" {
		name = e.StampedObject.GetGenerateName()
	}
	return fmt.Errorf("unable to apply object [%s/%s] for run template [%s]: %w",
		e.StampedObject.GetNamespace(),
		name,
		e.TemplateRef.Name,
		e.Err,
	).Error()
}

type ListCreatedObjectsError struct {
	Err       error
	Namespace string
	Labels    map[string]string
}

func (e ListCreatedObjectsError) Error() string {
	return fmt.Errorf("unable to list objects in namespace [%s] with labels [%v]: %w",
		e.Namespace,
		e.Labels,
		e.Err,
	).Error()
}

type RunnableRetrieveOutputError struct {
	Err               error
	StampedObject     *unstructured.Unstructured
	TemplateRef       *v1alpha1.TemplateReference
	QualifiedResource string
}

func (e RunnableRetrieveOutputError) Error() string {
	name := e.StampedObject.GetName()
	if name == "" {
		name = e.StampedObject.GetGenerateName()
	}

	return fmt.Errorf("unable to retrieve outputs from stamped object [%s/%s] of type [%s] for run template [%s]: %w",
		e.StampedObject.GetNamespace(),
		name,
		e.QualifiedResource,
		e.TemplateRef.Name,
		e.Err,
	).Error()
}
