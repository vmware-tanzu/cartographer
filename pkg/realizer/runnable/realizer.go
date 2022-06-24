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

//go:generate go run -modfile ../../../hack/tools/go.mod github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/errors"
	"github.com/vmware-tanzu/cartographer/pkg/logger"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/runnable/gc"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

//counterfeiter:generate . Realizer
type Realizer interface {
	Realize(ctx context.Context, runnable *v1alpha1.Runnable, systemRepo repository.Repository, runnableRepo repository.Repository, discoveryClient discovery.DiscoveryInterface) (*unstructured.Unstructured, templates.Outputs, error)
}

func NewRealizer() Realizer {
	return &runnableRealizer{}
}

type runnableRealizer struct{}

type TemplatingContext struct {
	Runnable *v1alpha1.Runnable     `json:"runnable"`
	Selected map[string]interface{} `json:"selected"`
}

//counterfeiter:generate k8s.io/client-go/discovery.DiscoveryInterface
func (r *runnableRealizer) Realize(ctx context.Context, runnable *v1alpha1.Runnable, systemRepo repository.Repository, runnableRepo repository.Repository, discoveryClient discovery.DiscoveryInterface) (*unstructured.Unstructured, templates.Outputs, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues("template", runnable.Spec.RunTemplateRef)
	ctx = logr.NewContext(ctx, log)

	runnable.Spec.RunTemplateRef.Kind = "ClusterRunTemplate"
	apiRunTemplate, err := systemRepo.GetRunTemplate(ctx, runnable.Spec.RunTemplateRef)

	if err != nil {
		log.Error(err, "failed to get runnable cluster template")
		return nil, nil, errors.RunnableGetRunTemplateError{
			Err:         err,
			TemplateRef: &runnable.Spec.RunTemplateRef,
		}
	}

	template := templates.NewRunTemplateModel(apiRunTemplate)

	labels := map[string]string{
		"carto.run/runnable-name":     runnable.Name,
		"carto.run/run-template-name": template.GetName(),
	}

	selected, err := r.resolveSelector(ctx, runnable.Spec.Selector, runnableRepo, discoveryClient, runnable.GetNamespace())
	if err != nil {
		log.Error(err, "failed to resolve selector", "selector", runnable.Spec.Selector)
		return nil, nil, errors.RunnableResolveSelectorError{
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
		log.Error(err, "failed to stamp resource")
		return nil, nil, errors.RunnableStampError{
			Err:         err,
			TemplateRef: &runnable.Spec.RunTemplateRef,
		}
	}

	err = runnableRepo.EnsureImmutableObjectExistsOnCluster(ctx, stampedObject, map[string]string{"carto.run/runnable-name": runnable.Name})
	if err != nil {
		log.Error(err, "failed to ensure object exists on cluster", "object", stampedObject)
		return nil, nil, errors.RunnableApplyStampedObjectError{
			Err:           err,
			StampedObject: stampedObject,
			TemplateRef:   &runnable.Spec.RunTemplateRef,
		}
	}

	allRunnableStampedObjects, err := runnableRepo.ListUnstructured(ctx, stampedObject.GroupVersionKind(), stampedObject.GetNamespace(), labels)
	if err != nil {
		log.Error(err, "failed to list objects")
		return stampedObject, nil, errors.RunnableListCreatedObjectsError{
			Err:       err,
			Namespace: stampedObject.GetNamespace(),
			Labels:    labels,
		}
	}

	err = gc.CleanupRunnableStampedObjects(ctx, allRunnableStampedObjects, runnable.Spec.RetentionPolicy, runnableRepo)
	if err != nil {
		log.Error(err, "failed to cleanup runnable stamped objects")
	}

	outputs, outputSource, err := template.GetLatestSuccessfulOutput(allRunnableStampedObjects)
	if err != nil {
		for _, obj := range allRunnableStampedObjects {
			log.V(logger.DEBUG).Info("failed to retrieve output from any object", "considered", obj)
		}
		log.Error(err, "failed to retrieve output from object")
		return stampedObject, nil, errors.RunnableRetrieveOutputError{
			Err:           err,
			StampedObject: stampedObject,
			TemplateRef:   &runnable.Spec.RunTemplateRef,
		}
	}

	if outputSource != nil {
		log.V(logger.DEBUG).Info("retrieved output from stamped object", "stamped object", outputSource)
	}

	if len(outputs) == 0 {
		log.V(logger.DEBUG).Info("no outputs retrieved, getting outputs from runnable.Status.Outputs")
		outputs = runnable.Status.Outputs
	}

	return stampedObject, outputs, nil
}

func (r *runnableRealizer) resolveSelector(ctx context.Context, selector *v1alpha1.ResourceSelector, repository repository.Repository, discoveryClient discovery.DiscoveryInterface, namespace string) (map[string]interface{}, error) {
	log := logr.FromContextOrDiscard(ctx)

	if selector == nil {
		return nil, nil
	}

	apiResourceList, err := discoveryClient.ServerResourcesForGroupVersion(selector.Resource.APIVersion)
	if err != nil {
		log.Error(err, "failed to list server api resources")
		return nil, fmt.Errorf("failed to list server api resources: %w", err)
	}

	var namespaced bool
	for _, apiResource := range apiResourceList.APIResources {
		if apiResource.Kind == selector.Resource.Kind {
			namespaced = apiResource.Namespaced
			break
		}
	}

	var results []*unstructured.Unstructured
	if namespaced {
		results, err = repository.ListUnstructured(ctx, schema.FromAPIVersionAndKind(selector.Resource.APIVersion, selector.Resource.Kind), namespace, selector.MatchingLabels)
		if err != nil {
			log.Error(err, "failed to list objects in namespace matching selector", "selector", selector.MatchingLabels)
			return nil, fmt.Errorf("failed to list objects in namespace matching selector [%+v]: %w", selector.MatchingLabels, err)
		}
	} else {
		results, err = repository.ListUnstructured(ctx, schema.FromAPIVersionAndKind(selector.Resource.APIVersion, selector.Resource.Kind), "", selector.MatchingLabels)
		if err != nil {
			log.Error(err, "failed to list objects at cluster scope matching selector", "selector", selector.MatchingLabels)
			return nil, fmt.Errorf("failed to list objects at cluster scope matching selector [%+v]: %w", selector.MatchingLabels, err)
		}
	}

	if len(results) == 0 {
		log.V(logger.DEBUG).Info("selector did not match any objects")
		return nil, fmt.Errorf("selector did not match any objects")
	} else if len(results) > 1 {
		log.V(logger.DEBUG).Info("selector matched multiple objects")
		return nil, fmt.Errorf("selector matched multiple objects")
	}
	return results[0].Object, nil
}
