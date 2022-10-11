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
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

func NewInputGenerator(resource Resource, outputs OutputsGetter) *InputGenerator {
	return &InputGenerator{

		resource: resource,
		outputs:  outputs,
	}
}

type Resource interface {
	GetName() string
	GetSources() []v1alpha1.ResourceReference
	GetImages() []v1alpha1.ResourceReference
	GetConfigs() []v1alpha1.ResourceReference
	GetDeployment() *v1alpha1.DeploymentReference
}

type OutputsGetter interface {
	GetSource(resourceName string) *templates.Source
	GetImage(resourceName string) templates.Image
	GetConfig(resourceName string) templates.Config
}

type InputGenerator struct {
	resource Resource
	outputs  OutputsGetter
}

func (i *InputGenerator) GetSources() map[string]templates.SourceInput {
	inputs := map[string]templates.SourceInput{}

	for _, reference := range i.resource.GetSources() {
		source := i.outputs.GetSource(reference.Resource)
		if source != nil {
			inputs[reference.Name] = templates.SourceInput{
				URL:      source.URL,
				Revision: source.Revision,
				Name:     reference.Name,
			}
		}
	}

	return inputs
}

func (i *InputGenerator) GetImages() map[string]templates.ImageInput {
	inputs := map[string]templates.ImageInput{}

	for _, reference := range i.resource.GetImages() {
		image := i.outputs.GetImage(reference.Resource)
		if image != nil {
			inputs[reference.Name] = templates.ImageInput{
				Image: image,
				Name:  reference.Name,
			}
		}
	}

	return inputs
}

func (i *InputGenerator) GetConfigs() map[string]templates.ConfigInput {
	inputs := map[string]templates.ConfigInput{}

	for _, reference := range i.resource.GetConfigs() {
		config := i.outputs.GetConfig(reference.Resource)
		if config != nil {
			inputs[reference.Name] = templates.ConfigInput{
				Config: config,
				Name:   reference.Name,
			}
		}
	}

	return inputs
}

func (i *InputGenerator) GetDeployment() *templates.SourceInput {
	if i.resource.GetDeployment() != nil {
		deployment := i.outputs.GetSource(i.resource.GetDeployment().Resource)
		if deployment != nil {
			return &templates.SourceInput{
				URL:      deployment.URL,
				Revision: deployment.Revision,
			}
		}
	}

	return nil
}
