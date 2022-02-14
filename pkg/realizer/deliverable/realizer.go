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

//go:generate go run -modfile ../../../hack/tools/go.mod github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"
	"github.com/vmware-tanzu/cartographer/pkg/templates"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/logger"
)

//counterfeiter:generate . Realizer
type Realizer interface {
	Realize(ctx context.Context, resourceRealizer ResourceRealizer, delivery *v1alpha1.ClusterDelivery) ([]templates.Template, []*unstructured.Unstructured, error)
}

type realizer struct{}

func NewRealizer() Realizer {
	return &realizer{}
}

func (r *realizer) Realize(ctx context.Context, resourceRealizer ResourceRealizer, delivery *v1alpha1.ClusterDelivery) ([]templates.Template, []*unstructured.Unstructured, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.V(logger.DEBUG).Info("Realize")

	outs := NewOutputs()
	var stampedObjects []*unstructured.Unstructured
	var selectedTemplates []templates.Template
	var firstError error

	for i := range delivery.Spec.Resources {
		resource := delivery.Spec.Resources[i]
		log = log.WithValues("resource", resource.Name)
		ctx = logr.NewContext(ctx, log)
		template, stampedObject, out, err := resourceRealizer.Do(ctx, &resource, delivery.Name, outs)
		if template != nil {
			selectedTemplates = append(selectedTemplates, template)
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

	return selectedTemplates, stampedObjects, firstError
}
