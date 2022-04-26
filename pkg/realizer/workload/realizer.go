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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/logger"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

//counterfeiter:generate . Realizer
type Realizer interface {
	Realize(ctx context.Context, resourceRealizer ResourceRealizer, supplyChain *v1alpha1.ClusterSupplyChain, previousResources []v1alpha1.ResourceStatus) ([]v1alpha1.ResourceStatus, bool, error)
}

type realizer struct{}

func NewRealizer() Realizer {
	return &realizer{}
}

func (r *realizer) Realize(ctx context.Context, resourceRealizer ResourceRealizer, supplyChain *v1alpha1.ClusterSupplyChain, previousResources []v1alpha1.ResourceStatus) ([]v1alpha1.ResourceStatus, bool, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.V(logger.DEBUG).Info("Realize")

	outs := NewOutputs()
	var realizedResources []v1alpha1.ResourceStatus
	var firstError error
	resourcesStatusChanged := false

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

		resourceStatus := generateResourceStatus(resource, template, stampedObject, out, previousResources)
		previousResourceStatus := getPreviousResourceStatus(resource.Name, previousResources)

		// set previousResourceStatus.Conditions to nil
		resourceStatus.Conditions = nil
		previousResourceStatusConditions := previousResourceStatus.Conditions
		previousResourceStatus.Conditions = nil
		if !reflect.DeepEqual(resourceStatus, &previousResourceStatus) {
			resourcesStatusChanged = true
		}
		// DeepEqual resourceStatus and previousResourceStatus, if diff mark resourcesStatusChanged = true

		// set back
		previousResourceStatus.Conditions = previousResourceStatusConditions
		conditionManager := conditions.NewConditionManager(v1alpha1.ResourceReady, previousResourceStatus.Conditions)

		if err != nil {
			conditions.AddConditionForResourceSubmitted(&conditionManager, false, err)
		} else {
			conditionManager.AddPositive(conditions.ResourceSubmittedCondition())
		}

		// if resourceStatus and previousResourceStatus are different, did the conditions (minus the time change on the resource, if yes, mark resourcesStatusChanged = true
		// DONE
		conditions, changed := conditionManager.Finalize()
		if !resourcesStatusChanged && changed {
			resourcesStatusChanged = true
		}
		resourceStatus.Conditions = conditions

		realizedResources = append(realizedResources, resourceStatus)

		outs.AddOutput(resource.Name, out)
	}

	return realizedResources, resourcesStatusChanged, firstError
}

func generateResourceStatus(resource v1alpha1.SupplyChainResource, template templates.Template, stampedObject *unstructured.Unstructured, output *templates.Output, previousResources []v1alpha1.ResourceStatus) v1alpha1.ResourceStatus {
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
	for _, image := range resource.Images {
		inputs = append(inputs, v1alpha1.Input{Name: image.Resource})
	}
	for _, config := range resource.Configs {
		inputs = append(inputs, v1alpha1.Input{Name: config.Resource})
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

	return v1alpha1.ResourceStatus{
		Name:        resource.Name,
		StampedRef:  stampedRef,
		TemplateRef: templateRef,
		Inputs:      inputs,
		Outputs:     outputs,
	}
}

func getPreviousResourceStatus(resourceName string, previousResources []v1alpha1.ResourceStatus) *v1alpha1.ResourceStatus {
	for _, previousResource := range previousResources {
		if previousResource.Name == resourceName {
			return &previousResource
		}
	}
	return nil
}

func getOutputs(resourceName string, template templates.Template, previousResources []v1alpha1.ResourceStatus, output *templates.Output) []v1alpha1.Output {
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
