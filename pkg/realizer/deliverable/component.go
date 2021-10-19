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

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . ResourceRealizer
type ResourceRealizer interface {
	Do(ctx context.Context, resource *v1alpha1.ClusterDeliveryResource, deliveryName string, outputs Outputs) (*templates.Output, error)
}

type resourceRealizer struct {
	deliverable *v1alpha1.Deliverable
	repo        repository.Repository
}

func NewResourceRealizer(deliverable *v1alpha1.Deliverable, repo repository.Repository) ResourceRealizer {
	return &resourceRealizer{
		deliverable: deliverable,
		repo:        repo,
	}
}

func (r *resourceRealizer) Do(ctx context.Context, resource *v1alpha1.ClusterDeliveryResource, deliveryName string, outputs Outputs) (*templates.Output, error) {
	template, err := r.repo.GetDeliveryClusterTemplate(resource.TemplateRef)
	if err != nil {
		return nil, GetDeliveryClusterTemplateError{
			Err:         err,
			TemplateRef: resource.TemplateRef,
		}
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
		"params":      templates.ParamsBuilder(template.GetDefaultParams(), resource.Params),
		"sources":     inputs.Sources,
		"configs":     inputs.Configs,
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
		return nil, StampError{
			Err:      err,
			Resource: resource,
		}
	}

	err = r.repo.EnsureObjectExistsOnCluster(stampedObject, true)
	if err != nil {
		return nil, ApplyStampedObjectError{
			Err:           err,
			StampedObject: stampedObject,
		}
	}

	output, err := template.GetOutput(stampedObject)
	if err != nil {
		return nil, RetrieveOutputError{
			Err:      err,
			resource: resource,
		}
	}

	return output, nil
}
