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
	"reflect"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/logger"
)

//counterfeiter:generate . Realizer
type Realizer interface {
	Realize(ctx context.Context, resourceRealizer ResourceRealizer, supplyChain *v1alpha1.ClusterSupplyChain, previousResources []v1alpha1.RealizedResource) ([]v1alpha1.RealizedResource, error)
}

type realizer struct{}

func NewRealizer() Realizer {
	return &realizer{}
}

func (r *realizer) Realize(ctx context.Context, resourceRealizer ResourceRealizer, supplyChain *v1alpha1.ClusterSupplyChain, previousResources []v1alpha1.RealizedResource) ([]v1alpha1.RealizedResource, error) {
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

		outputs := template.GetResourceOutput()

		currTime := metav1.NewTime(time.Now())
		for j, output := range outputs {
			outputs[j].LastTransitionTime = currTime
			for _, previousResource := range previousResources {
				if previousResource.Name == resource.Name {
					for _, previousOutput := range previousResource.Outputs {
						if previousOutput.Name == output.Name && reflect.DeepEqual(previousOutput.Value, output.Value) {
							outputs[j].LastTransitionTime = previousOutput.LastTransitionTime
						}
					}
				}
			}
		}

		realizedResources = append(realizedResources, v1alpha1.RealizedResource{
			Name: resource.Name,
			StampedRef: corev1.ObjectReference{
				Kind:       stampedObject.GetKind(),
				Namespace:  stampedObject.GetNamespace(),
				Name:       stampedObject.GetName(),
				APIVersion: stampedObject.GetAPIVersion(),
			},
			TemplateRef: corev1.ObjectReference{
				Kind:       template.GetKind(),
				Namespace:  "",
				Name:       template.GetName(),
				APIVersion: v1alpha1.SchemeGroupVersion.String(),
			},
			Inputs:             inputs,
			Outputs:            outputs,
			ObservedGeneration: stampedObject.GetGeneration(),
		})

		outs.AddOutput(resource.Name, out)
	}

	return realizedResources, firstError
}
