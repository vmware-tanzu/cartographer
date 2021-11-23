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

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

//counterfeiter:generate . Realizer
type Realizer interface {
	Realize(ctx context.Context, runnable *v1alpha1.Runnable, systemRepo repository.Repository, runnableRepo repository.Repository) (*unstructured.Unstructured, templates.Outputs, error)
}

func NewRealizer() Realizer {
	return &runnableRealizer{}
}

type runnableRealizer struct{}

type TemplatingContext struct {
	Runnable *v1alpha1.Runnable     `json:"runnable"`
	Selected map[string]interface{} `json:"selected"`
}

func (p *runnableRealizer) Realize(ctx context.Context, runnable *v1alpha1.Runnable, systemRepo repository.Repository, runnableRepo repository.Repository) (*unstructured.Unstructured, templates.Outputs, error) {
	runnable.Spec.RunTemplateRef.Kind = "ClusterRunTemplate"
	apiRunTemplate, err := systemRepo.GetRunTemplate(ctx, runnable.Spec.RunTemplateRef)

	if err != nil {
		return nil, nil, GetRunTemplateError{
			Err:      err,
			Runnable: runnable,
		}
	}

	template := templates.NewRunTemplateModel(apiRunTemplate)

	labels := map[string]string{
		"carto.run/runnable-name":     runnable.Name,
		"carto.run/run-template-name": template.GetName(),
	}

	selected, err := resolveSelector(ctx, runnable.Spec.Selector, runnableRepo, runnable.GetNamespace())
	if err != nil {
		return nil, nil, ResolveSelectorError{
			Err:      err,
			Selector: runnable.Spec.Selector,
		}
	}

	stampContext := templates.StamperBuilder(
		runnable,
		TemplatingContext{
			Runnable: runnable,
			Selected: selected,
		},
		labels,
	)

	stampedObject, err := stampContext.Stamp(ctx, template.GetResourceTemplate())
	if err != nil {
		return nil, nil, StampError{
			Err:      err,
			Runnable: runnable,
		}
	}

	err = runnableRepo.EnsureObjectExistsOnCluster(ctx, stampedObject.DeepCopy(), false)
	if err != nil {
		return nil, nil, ApplyStampedObjectError{
			Err:           err,
			StampedObject: stampedObject,
		}
	}

	objectForListCall := stampedObject.DeepCopy()
	objectForListCall.SetLabels(labels)

	allRunnableStampedObjects, err := runnableRepo.ListUnstructured(ctx, objectForListCall)
	if err != nil {
		return stampedObject, nil, ListCreatedObjectsError{
			Err:       err,
			Namespace: objectForListCall.GetNamespace(),
			Labels:    objectForListCall.GetLabels(),
		}
	}

	outputs, err := template.GetOutput(allRunnableStampedObjects)
	if err != nil {
		return stampedObject, nil, RetrieveOutputError{
			Err:      err,
			Runnable: runnable,
		}
	}

	if len(outputs) == 0 {
		outputs = runnable.Status.Outputs
	}

	return stampedObject, outputs, nil
}

func resolveSelector(ctx context.Context, selector *v1alpha1.ResourceSelector, repository repository.Repository, namespace string) (map[string]interface{}, error) {
	if selector == nil {
		return nil, nil
	}
	queryObj := &unstructured.Unstructured{}
	queryObj.SetGroupVersionKind(schema.FromAPIVersionAndKind(selector.Resource.APIVersion, selector.Resource.Kind))
	queryObj.SetLabels(selector.MatchingLabels)
	queryObj.SetNamespace(namespace)

	results, err := repository.ListUnstructured(ctx, queryObj)
	if err != nil {
		return nil, fmt.Errorf("could not list objects matching selector: %w", err)
	}

	if len(results) == 0 {
		return resolveClusterScopedSelector(ctx, selector, repository)
	} else if len(results) > 1 {
		return nil, fmt.Errorf("selector matched multiple objects")
	}
	return results[0].Object, nil
}

func resolveClusterScopedSelector(ctx context.Context, selector *v1alpha1.ResourceSelector, repository repository.Repository) (map[string]interface{}, error) {
	queryObj := &unstructured.Unstructured{}
	queryObj.SetGroupVersionKind(schema.FromAPIVersionAndKind(selector.Resource.APIVersion, selector.Resource.Kind))
	queryObj.SetLabels(selector.MatchingLabels)

	results, err := repository.ListUnstructured(ctx, queryObj)
	if err != nil {
		return nil, fmt.Errorf("could not list objects matching selector: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("selector did not match any objects")
	} else if len(results) > 1 {
		return nil, fmt.Errorf("selector matched multiple objects")
	}
	return results[0].Object, nil
}
