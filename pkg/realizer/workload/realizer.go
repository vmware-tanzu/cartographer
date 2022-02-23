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

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/logger"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

//counterfeiter:generate . Realizer
type Realizer interface {
	Realize(ctx context.Context, resourceRealizer ResourceRealizer, supplyChain *v1alpha1.ClusterSupplyChain) ([]v1alpha1.RealizedResource, error)
}

type realizer struct{}

func NewRealizer() Realizer {
	return &realizer{}
}

func (r *realizer) Realize(ctx context.Context, resourceRealizer ResourceRealizer, supplyChain *v1alpha1.ClusterSupplyChain) ([]v1alpha1.RealizedResource, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.V(logger.DEBUG).Info("Realize")

	outs := NewOutputs()
	var realizedResources []v1alpha1.RealizedResource
	var firstError error

	for i := range supplyChain.Spec.Resources {
		resource := supplyChain.Spec.Resources[i]
		template, stampedObject, out, err := resourceRealizer.Do(ctx, &resource, supplyChain.Name, outs)

		if stampedObject != nil {
			log.V(logger.DEBUG).Info("realized resource as object",
				"object", stampedObject)
		}

		if err != nil {
			log.Error(err, "failed to realize resource")

			if firstError == nil {
				firstError = err
			}
		}

		realizedResource, err := generateRealizedResource(resource, template, stampedObject, out)
		if err != nil {
			log.Error(err, "failed to generate realized resource data for stampedObject", "object", stampedObject)
		}
		realizedResources = append(realizedResources, realizedResource)

		outs.AddOutput(resource.Name, out)
	}

	return realizedResources, firstError
}

func generateRealizedResource(resource v1alpha1.SupplyChainResource, template templates.Template, stampedObject *unstructured.Unstructured, output *templates.Output) (v1alpha1.RealizedResource, error) {
	var inputs []v1alpha1.Input
	for _, source := range resource.Sources {
		inputs = append(inputs, v1alpha1.Input{Name: source.Resource})
	}
	for _, image := range resource.Images {
		inputs = append(inputs, v1alpha1.Input{Name: image.Resource})
	}
	for _, config := range resource.Configs {
		inputs = append(inputs, v1alpha1.Input{Name: config.Resource})
	}

	var templateRef *corev1.ObjectReference
	var outputs []v1alpha1.Output
	if template != nil {
		if output != nil {
			outputs = template.GenerateResourceOutput()
		}
		templateRef = &corev1.ObjectReference{
			Kind:       template.GetKind(),
			Name:       template.GetName(),
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		}
	}

	var stampedRef *corev1.ObjectReference
	if stampedObject != nil {
		stampedRef = &corev1.ObjectReference{
			Kind:       stampedObject.GetKind(),
			Namespace:  stampedObject.GetNamespace(),
			Name:       stampedObject.GetName(),
			APIVersion: stampedObject.GetAPIVersion(),
		}
	}

	return v1alpha1.RealizedResource{
		Name:        resource.Name,
		StampedRef:  stampedRef,
		TemplateRef: templateRef,
		Inputs:      inputs,
		Outputs:     outputs,
	}, nil
}
