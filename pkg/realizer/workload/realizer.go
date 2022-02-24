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

//go:generate go run -modfile ../../../hack/tools/go.mod github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/logger"
	"github.com/vmware-tanzu/cartographer/pkg/supplychains"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

//counterfeiter:generate . Realizer
type Realizer interface {
	Realize(ctx context.Context, resourceRealizer ResourceRealizer, supplyChain supplychains.SupplyChain) (map[string]templates.Template, []*unstructured.Unstructured, error)
}

type realizer struct{}

func NewRealizer() Realizer {
	return &realizer{}
}

func (r *realizer) Realize(ctx context.Context, resourceRealizer ResourceRealizer, supplyChain supplychains.SupplyChain) (map[string]templates.Template, []*unstructured.Unstructured, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.V(logger.DEBUG).Info("Realize")

	outs := NewOutputs()
	var stampedObjects []*unstructured.Unstructured
	var selectedTemplates map[string]templates.Template
	var firstError error

	supplyChainResources := supplyChain.GetResources()
	for i := range supplyChainResources {
		resource := supplyChainResources[i]
		if strings.Contains(resource.TemplateRef.Kind, "SupplyChain") {
			sc, err := resourceRealizer.FindMatchingSupplyChain(ctx, &resource, supplyChain.GetName())
			if err != nil {
				panic(err)
			}
			usedTemplates, stampedObjs, err := r.Realize(ctx, resourceRealizer, sc)
			for resourceName, template := range usedTemplates {
				if template != nil {
					selectedTemplates[resourceName] = template
				}
			}

			for _, stampedObj := range stampedObjs {
				if stampedObj != nil {
					log.V(logger.DEBUG).Info("realized resource as object",
						"object", stampedObj)
					stampedObjects = append(stampedObjects, stampedObj)
				}
			}

			if template, ok := usedTemplates[sc.GetOutputResource()]; ok {
				out, err := template.GetOutput()
				if err != nil {
					panic(err)
				}
				outs.AddOutput(resource.Name, out)
			}
		} else {
			template, stampedObject, out, err := resourceRealizer.Do(ctx, &resource, supplyChain.GetName(), outs)

			if template != nil {
				selectedTemplates[resource.Name] = template
			}

			if stampedObject != nil {
				log.V(logger.DEBUG).Info("realized resource as object",
					"object", stampedObject)
				stampedObjects = append(stampedObjects, stampedObject)
			}

			if err != nil {
				log.Error(err, "failed to realize resource")

				if firstError == nil {
					firstError = err
				}
			}
			outs.AddOutput(resource.Name, out)
		}
	}

	return selectedTemplates, stampedObjects, firstError
}
