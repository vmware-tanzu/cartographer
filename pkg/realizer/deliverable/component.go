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
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	realizerclient "github.com/vmware-tanzu/cartographer/pkg/realizer/client"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . ResourceRealizer
type ResourceRealizer interface {
	Do(ctx context.Context, resource *v1alpha1.ClusterDeliveryResource, deliveryName string, outputs Outputs) (*unstructured.Unstructured, *templates.Output, error)
}

type resourceRealizer struct {
	deliverable     *v1alpha1.Deliverable
	systemRepo      repository.Repository
	deliverableRepo repository.Repository
	deliveryParams  []v1alpha1.DelegatableParam
}

type ResourceRealizerBuilder func(secret *corev1.Secret, deliverable *v1alpha1.Deliverable, repo repository.Repository, deliveryParams []v1alpha1.DelegatableParam) (ResourceRealizer, error)

func NewResourceRealizerBuilder(repositoryBuilder repository.RepositoryBuilder, clientBuilder realizerclient.ClientBuilder, cache repository.RepoCache) ResourceRealizerBuilder {
	return func(secret *corev1.Secret, deliverable *v1alpha1.Deliverable, systemRepo repository.Repository, deliveryParams []v1alpha1.DelegatableParam) (ResourceRealizer, error) {
		client, err := clientBuilder(secret)
		if err != nil {
			return nil, fmt.Errorf("can't build client: %w", err)
		}
		deliverableRepo := repositoryBuilder(client, cache)
		return &resourceRealizer{
			deliverable:     deliverable,
			systemRepo:      systemRepo,
			deliverableRepo: deliverableRepo,
			deliveryParams:  deliveryParams,
		}, nil
	}
}

func (r *resourceRealizer) Do(ctx context.Context, resource *v1alpha1.ClusterDeliveryResource, deliveryName string, outputs Outputs) (*unstructured.Unstructured, *templates.Output, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues("template", resource.TemplateRef)
	ctx = logr.NewContext(ctx, log)

	apiTemplate, err := r.systemRepo.GetDeliveryClusterTemplate(ctx, resource.TemplateRef)
	if err != nil {
		log.Error(err, "failed to get delivery cluster template")
		return nil, nil, GetDeliveryClusterTemplateError{
			Err:         err,
			TemplateRef: resource.TemplateRef,
		}
	}

	template, err := templates.NewModelFromAPI(apiTemplate)
	if err != nil {
		log.Error(err, "failed to get delivery cluster template")
		return nil, nil, fmt.Errorf("failed to get delivery cluster template [%+v]: %w", resource.TemplateRef, err)
	}

	labels := map[string]string{
		"carto.run/deliverable-name":      r.deliverable.Name,
		"carto.run/deliverable-namespace": r.deliverable.Namespace,
		"carto.run/cluster-delivery-name": deliveryName,
		"carto.run/resource-name":         resource.Name,
		"carto.run/template-kind":         template.GetKind(),
		"carto.run/cluster-template-name": template.GetName(),
	}

	inputs := outputs.GenerateInputs(resource)
	templatingContext := map[string]interface{}{
		"deliverable": r.deliverable,
		"params":      templates.ParamsBuilder(template.GetDefaultParams(), r.deliveryParams, resource.Params, r.deliverable.Spec.Params),
		"sources":     inputs.Sources,
		"configs":     inputs.Configs,
		"deployment":  inputs.Deployment,
	}

	// Todo: this belongs in Stamp.
	if inputs.OnlyConfig() != nil {
		templatingContext["config"] = inputs.OnlyConfig()
	}
	if inputs.OnlySource() != nil {
		templatingContext["source"] = inputs.OnlySource()
	}

	stampContext := templates.StamperBuilder(r.deliverable, templatingContext, labels)
	stampedObject, err := stampContext.Stamp(ctx, template.GetResourceTemplate())
	if err != nil {
		log.Error(err, "failed to stamp resource")
		return nil, nil, StampError{
			Err:      err,
			Resource: resource,
		}
	}

	err = r.deliverableRepo.EnsureObjectExistsOnCluster(ctx, stampedObject, true)
	if err != nil {
		log.Error(err, "failed to ensure object exists on cluster", "object", stampedObject)
		return nil, nil, ApplyStampedObjectError{
			Err:           err,
			StampedObject: stampedObject,
		}
	}

	template.SetInputs(inputs)
	template.SetStampedObject(stampedObject)

	output, err := template.GetOutput()
	if err != nil {
		log.Error(err, "failed to retrieve output from object", "object", stampedObject)
		return stampedObject, nil, RetrieveOutputError{
			Err:      err,
			Resource: resource,
		}
	}

	return stampedObject, output, nil
}
