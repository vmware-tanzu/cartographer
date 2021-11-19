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

package workload

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . ResourceRealizer
type ResourceRealizer interface {
	Do(ctx context.Context, resource *v1alpha1.SupplyChainResource, supplyChainName string, supplyChainParams []v1alpha1.OverridableParam, outputs Outputs) (*unstructured.Unstructured, *templates.Output, error)
}

type resourceRealizer struct {
	workload *v1alpha1.Workload
	repo     repository.Repository
}

func NewResourceRealizer(workload *v1alpha1.Workload, repo repository.Repository) ResourceRealizer {
	return &resourceRealizer{
		workload: workload,
		repo:     repo,
	}
}

func (r *resourceRealizer) Do(ctx context.Context, resource *v1alpha1.SupplyChainResource, supplyChainName string, supplyChainParams []v1alpha1.OverridableParam, outputs Outputs) (*unstructured.Unstructured, *templates.Output, error) {
	apiTemplate, err := r.repo.GetClusterTemplate(ctx, resource.TemplateRef)
	if err != nil {
		return nil, nil, GetClusterTemplateError{
			Err:         err,
			TemplateRef: resource.TemplateRef,
		}
	}

	template, err := templates.NewModelFromAPI(apiTemplate)
	if err != nil {
		return nil, nil, fmt.Errorf("new model from api: %w", err)
	}

	labels := map[string]string{
		"carto.run/workload-name":             r.workload.Name,
		"carto.run/workload-namespace":        r.workload.Namespace,
		"carto.run/cluster-supply-chain-name": supplyChainName,
		"carto.run/resource-name":             resource.Name,
		"carto.run/template-kind":             template.GetKind(),
		"carto.run/cluster-template-name":     template.GetName(),
	}

	inputs := outputs.GenerateInputs(resource)
	workloadTemplatingContext := map[string]interface{}{
		"workload": r.workload,
		"params":   templates.ParamsBuilder(template.GetDefaultParams(), supplyChainParams, resource.Params, r.workload.Spec.Params),
		"sources":  inputs.Sources,
		"images":   inputs.Images,
		"configs":  inputs.Configs,
	}

	// Todo: this belongs in Stamp.
	if inputs.OnlyConfig() != nil {
		workloadTemplatingContext["config"] = inputs.OnlyConfig()
	}
	if inputs.OnlyImage() != nil {
		workloadTemplatingContext["image"] = inputs.OnlyImage()
	}
	if inputs.OnlySource() != nil {
		workloadTemplatingContext["source"] = inputs.OnlySource()
	}

	stampContext := templates.StamperBuilder(r.workload, workloadTemplatingContext, labels)
	stampedObject, err := stampContext.Stamp(ctx, template.GetResourceTemplate())
	if err != nil {
		return nil, nil, StampError{
			Err:      err,
			Resource: resource,
		}
	}

	err = r.repo.EnsureObjectExistsOnCluster(ctx, stampedObject, true)
	if err != nil {
		return nil, nil, ApplyStampedObjectError{
			Err:           err,
			StampedObject: stampedObject,
		}
	}

	template.SetStampedObject(stampedObject)

	output, err := template.GetOutput()
	if err != nil {
		return stampedObject, nil, RetrieveOutputError{
			Err:      err,
			Resource: resource,
		}
	}

	return stampedObject, output, nil
}
