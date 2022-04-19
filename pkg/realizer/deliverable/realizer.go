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
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/logger"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

//counterfeiter:generate . Realizer
type Realizer interface {
	Realize(ctx context.Context, resourceRealizer ResourceRealizer, delivery *v1alpha1.ClusterDelivery, previousResources []v1alpha1.RealizedResource) ([]v1alpha1.RealizedResource, error)
}

type realizer struct{}

func NewRealizer() Realizer {
	return &realizer{}
}

func (r *realizer) Realize(ctx context.Context, resourceRealizer ResourceRealizer, delivery *v1alpha1.ClusterDelivery, previousResources []v1alpha1.RealizedResource) ([]v1alpha1.RealizedResource, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.V(logger.DEBUG).Info("Realize")

	outs := NewOutputs()
	var realizedResources []v1alpha1.RealizedResource
	var firstError error

	for i := range delivery.Spec.Resources {
		resource := delivery.Spec.Resources[i]
		log = log.WithValues("resource", resource.Name)
		ctx = logr.NewContext(ctx, log)
		template, stampedObject, out, err := resourceRealizer.Do(ctx, &resource, delivery.Name, outs)

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

		realizedResource := generateRealizedResource(resource, template, stampedObject, out, previousResources)

		conditionManagerBuilder := conditions.NewConditionManager
		conditionManager := conditionManagerBuilder(v1alpha1.ResourceReady, getPreviousResourceConditions(resource.Name, previousResources))

		if err != nil {
			conditions.AddConditionForDeliverableError(&conditionManager, v1alpha1.ResourceSubmitted, err)
		} else {
			conditionManager.AddPositive(conditions.ResourceSubmittedCondition())
		}

		conditions, _ := conditionManager.Finalize()
		realizedResource.Conditions = conditions

		realizedResources = append(realizedResources, realizedResource)

		outs.AddOutput(resource.Name, out)
	}

	return realizedResources, firstError
}

func generateRealizedResource(resource v1alpha1.DeliveryResource, template templates.Template, stampedObject *unstructured.Unstructured, output *templates.Output, previousResources []v1alpha1.RealizedResource) v1alpha1.RealizedResource {
	if stampedObject == nil || template == nil {
		for _, previousResource := range previousResources {
			if previousResource.Name == resource.Name {
				return previousResource
			}
		}
	}

	var inputs []v1alpha1.Input
	for _, source := range resource.Sources {
		inputs = append(inputs, v1alpha1.Input{Name: source.Resource})
	}
	for _, config := range resource.Configs {
		inputs = append(inputs, v1alpha1.Input{Name: config.Resource})
	}
	if resource.Deployment != nil {
		inputs = append(inputs, v1alpha1.Input{Name: resource.Deployment.Resource})
	}

	var templateRef *corev1.ObjectReference
	var outputs []v1alpha1.Output
	if template != nil {
		templateRef = &corev1.ObjectReference{
			Kind:       template.GetKind(),
			Name:       template.GetName(),
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		}

		outputs = getOutputs(resource.Name, template, previousResources, output)
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
	}
}

func getPreviousResourceConditions(resourceName string, previousResources []v1alpha1.RealizedResource) []metav1.Condition {
	for _, previousResource := range previousResources {
		if previousResource.Name == resourceName {
			return previousResource.Conditions
		}
	}
	return nil
}

func getOutputs(resourceName string, template templates.Template, previousResources []v1alpha1.RealizedResource, output *templates.Output) []v1alpha1.Output {
	outputs, err := template.GenerateResourceOutput(output)
	if err != nil {
		for _, previousResource := range previousResources {
			if previousResource.Name == resourceName {
				outputs = previousResource.Outputs
				break
			}
		}
	} else {
		currTime := metav1.NewTime(time.Now())
		for j, out := range outputs {
			outputs[j].LastTransitionTime = currTime
			for _, previousResource := range previousResources {
				if previousResource.Name == resourceName {
					for _, previousOutput := range previousResource.Outputs {
						if previousOutput.Name == out.Name {
							if previousOutput.Digest == out.Digest {
								outputs[j].LastTransitionTime = previousOutput.LastTransitionTime
							}
							break
						}
					}
					break
				}
			}
		}
	}

	return outputs
}
