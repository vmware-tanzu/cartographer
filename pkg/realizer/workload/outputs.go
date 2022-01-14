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
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

type Outputs map[string]*templates.Output

func NewOutputs() Outputs {
	return make(Outputs)
}

func (o Outputs) AddOutput(name string, output *templates.Output) {
	o[name] = output
}

func (o Outputs) getResourceSource(resourceName string) *templates.Source {
	output := o[resourceName]
	if output == nil {
		return nil
	}

	return output.Source
}

func (o Outputs) getResourceImage(resourceName string) templates.Image {
	output := o[resourceName]
	if output == nil {
		return ""
	}
	return output.Image
}

func (o Outputs) getResourceConfig(resourceName string) templates.Config {
	output := o[resourceName]
	if output == nil {
		return ""
	}
	return output.Config
}

func (o Outputs) GenerateInputs(resource *v1alpha1.SupplyChainResource) *templates.Inputs {
	inputs := &templates.Inputs{
		Sources: map[string]templates.SourceInput{},
		Images:  map[string]templates.ImageInput{},
		Configs: map[string]templates.ConfigInput{},
	}

	for _, referenceSource := range resource.Sources {
		source := o.getResourceSource(referenceSource.Resource)
		if source != nil {
			inputs.Sources[referenceSource.Name] = templates.SourceInput{
				URL:      source.URL,
				Revision: source.Revision,
				Name:     referenceSource.Name,
			}
		}
	}

	for _, referenceImage := range resource.Images {
		image := o.getResourceImage(referenceImage.Resource)
		if image != "" {
			inputs.Images[referenceImage.Name] = templates.ImageInput{
				Image: image,
				Name:  referenceImage.Name,
			}
		}
	}

	for _, referenceConfig := range resource.Configs {
		config := o.getResourceConfig(referenceConfig.Resource)
		if config != "" {
			inputs.Configs[referenceConfig.Name] = templates.ConfigInput{
				Config: config,
				Name:   referenceConfig.Name,
			}
		}
	}

	return inputs
}
