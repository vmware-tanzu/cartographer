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

package runnable

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

type GetRunTemplateError struct {
	Err      error
	Runnable *v1alpha1.Runnable
}

func (e GetRunTemplateError) Error() string {
	return fmt.Errorf("unable to get runnable '%s/%s': '%w'",
		e.Runnable.Namespace, e.Runnable.Name, e.Err).Error()
}

type ResolveSelectorError struct {
	Err      error
	Selector *v1alpha1.ResourceSelector
}

func (e ResolveSelectorError) Error() string {
	return fmt.Errorf("unable to resolve selector '(apiVersion:%s kind:%s labels:%v)': '%w'",
		e.Selector.Resource.APIVersion,
		e.Selector.Resource.Kind,
		e.Selector.MatchingLabels,
		e.Err).Error()
}

type StampError struct {
	Err      error
	Runnable *v1alpha1.Runnable
}

func (e StampError) Error() string {
	return fmt.Errorf("unable to stamp object '%s/%s': '%w'",
		e.Runnable.Namespace, e.Runnable.Name, e.Err).Error()
}

type ApplyStampedObjectError struct {
	Err           error
	StampedObject *unstructured.Unstructured
}

func (e ApplyStampedObjectError) Error() string {
	name := e.StampedObject.GetName()
	if name == "" {
		name = e.StampedObject.GetGenerateName()
	}
	return fmt.Errorf("unable to apply stamped object '%s/%s': '%w'",
		e.StampedObject.GetNamespace(), name, e.Err).Error()
}

type ListCreatedObjectsError struct {
	Err       error
	Namespace string
	Labels    map[string]string
}

func (e ListCreatedObjectsError) Error() string {
	return fmt.Errorf("unable to list objects in namespace '%s' with labels '%v': '%w'",
		e.Namespace, e.Labels, e.Err).Error()
}

type RetrieveOutputError struct {
	Err      error
	Runnable *v1alpha1.Runnable
}

func (e RetrieveOutputError) Error() string {
	return fmt.Errorf("unable to retrieve outputs from stamped object for runnable '%s/%s': %w",
		e.Runnable.Namespace, e.Runnable.Name, e.Err).Error()
}
