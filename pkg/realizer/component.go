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

package realizer

import (
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . ComponentRealizer
type ComponentRealizer interface {
	Do(component *v1alpha1.SupplyChainComponent, supplyChainName string, outputs Outputs) (*templates.Output, error)
}

type componentRealizer struct {
	workload *v1alpha1.Workload
	repo     repository.Repository
}

func NewComponentRealizer(workload *v1alpha1.Workload, repo repository.Repository) ComponentRealizer {
	return &componentRealizer{
		workload: workload,
		repo:     repo,
	}
}

func (r *componentRealizer) Do(component *v1alpha1.SupplyChainComponent, supplyChainName string, outputs Outputs) (*templates.Output, error) {
	template, err := r.repo.GetTemplate(component.TemplateRef)
	if err != nil {
		return nil, GetTemplateError{
			Err:         err,
			TemplateRef: component.TemplateRef,
		}
	}

	labels := map[string]string{
		"carto.run/workload-name":             r.workload.Name,
		"carto.run/workload-namespace":        r.workload.Namespace,
		"carto.run/cluster-supply-chain-name": supplyChainName,
		"carto.run/component-name":            component.Name,
		"carto.run/cluster-template-name":     template.GetName(),
	}

	stampContext := templates.StampContextBuilder(
		r.workload,
		labels,
		templates.ParamsBuilder(template.GetDefaultParams(), component.Params),
		outputs.GenerateInputs(component),
	)

	stampedObject, err := stampContext.Stamp(template.GetResourceTemplate().Raw)
	if err != nil {
		return nil, StampError{
			Err:       err,
			Component: component,
		}
	}

	err = r.repo.AssureObjectExistsOnCluster(stampedObject)
	if err != nil {
		return nil, ApplyStampedObjectError{
			Err:           err,
			StampedObject: stampedObject,
		}
	}

	output, err := template.GetOutput(stampedObject)
	if err != nil {
		return nil, RetrieveOutputError{
			Err:       err,
			component: component,
		}
	}

	return output, nil
}
