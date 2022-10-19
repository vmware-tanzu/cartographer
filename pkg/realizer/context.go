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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

// Todo: Pass an interface for owner and ownerParams that supports getParams and getObject
func NewContextGenerator(owner client.Object, ownerParams []v1alpha1.OwnerParam, blueprintParams []v1alpha1.BlueprintParam) *contextGenerator {
	return &contextGenerator{
		blueprintParams: blueprintParams,
		ownerParams:     ownerParams,
		owner:           owner,
	}
}

type contextGenerator struct {
	blueprintParams []v1alpha1.BlueprintParam
	ownerParams     []v1alpha1.OwnerParam
	owner           client.Object
}

// Generate builds a context based on the template, owner and resource
func (c contextGenerator) Generate(templateParams TemplateParams, resource OwnerResource, outputs OutputsGetter) map[string]interface{} {
	inputGenerator := NewInputGenerator(resource, outputs)
	merger := NewParamMerger(resource.Params, c.blueprintParams, c.ownerParams)

	configs := inputGenerator.GetConfigs()
	sources := inputGenerator.GetSources()
	images := inputGenerator.GetImages()
	result := map[string]interface{}{
		"workload":    c.owner,
		"deliverable": c.owner,
		"params":      merger.Merge(templateParams),
		"sources":     sources,
		"images":      images,
		"configs":     configs,
		"deployment":  inputGenerator.GetDeployment(),
	}

	if len(sources) == 1 {
		for _, source := range sources {
			result["source"] = &source
		}
	}

	if len(images) == 1 {
		for _, image := range images {
			result["image"] = image.Image
		}
	}

	if len(configs) == 1 {
		for _, config := range configs {
			result["config"] = config.Config
		}
	}

	return result
}
