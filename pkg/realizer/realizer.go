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

//go:generate go run -modfile ../../hack/tools/go.mod github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/strings/slices"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/events"
	"github.com/vmware-tanzu/cartographer/pkg/logger"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/healthcheck"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/statuses"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

func MakeSupplychainOwnerResources(supplyChain *v1alpha1.ClusterSupplyChain) []OwnerResource {
	var resources []OwnerResource
	for _, resource := range supplyChain.Spec.Resources {
		resources = append(resources, OwnerResource{
			Name: resource.Name,
			TemplateRef: v1alpha1.TemplateReference{
				Kind: resource.TemplateRef.Kind,
				Name: resource.TemplateRef.Name,
			},
			TemplateOptions: resource.TemplateRef.Options,
			Params:          resource.Params,
			Sources:         resource.Sources,
			Images:          resource.Images,
			Configs:         resource.Configs,
		})
	}
	return resources
}

func MakeDeliveryOwnerResources(delivery *v1alpha1.ClusterDelivery) []OwnerResource {
	var resources []OwnerResource
	for _, resource := range delivery.Spec.Resources {
		resources = append(resources, OwnerResource{
			Name: resource.Name,
			TemplateRef: v1alpha1.TemplateReference{
				Kind: resource.TemplateRef.Kind,
				Name: resource.TemplateRef.Name,
			},
			TemplateOptions: resource.TemplateRef.Options,
			Params:          resource.Params,
			Sources:         resource.Sources,
			Configs:         resource.Configs,
			Deployment:      resource.Deployment,
		})
	}
	return resources
}

//counterfeiter:generate . Realizer
type Realizer interface {
	Realize(ctx context.Context, resourceRealizer ResourceRealizer, blueprintName string, ownerResources []OwnerResource, resourceStatuses statuses.ResourceStatuses) error
}

type realizer struct {
	healthyConditionEvaluator HealthyConditionEvaluator
	mapper                    meta.RESTMapper
}

type HealthyConditionEvaluator func(rule *v1alpha1.HealthRule, realizedResource *v1alpha1.RealizedResource, stampedObject *unstructured.Unstructured) metav1.Condition

//counterfeiter:generate k8s.io/apimachinery/pkg/api/meta.RESTMapper
func NewRealizer(healthyConditionEvaluator HealthyConditionEvaluator, mapper meta.RESTMapper) Realizer {
	if healthyConditionEvaluator == nil {
		healthyConditionEvaluator = healthcheck.DetermineHealthCondition
	}
	return &realizer{
		healthyConditionEvaluator: healthyConditionEvaluator,
		mapper:                    mapper,
	}
}

func (r *realizer) Realize(ctx context.Context, resourceRealizer ResourceRealizer, blueprintName string, ownerResources []OwnerResource, resourceStatuses statuses.ResourceStatuses) error {
	log := logr.FromContextOrDiscard(ctx)
	log.V(logger.DEBUG).Info("Realize")

	outs := NewOutputs()
	var firstError error

	for _, resource := range ownerResources {
		log = log.WithValues("resource", resource.Name)
		ctx = logr.NewContext(ctx, log)
		template, stampedObject, out, err := resourceRealizer.Do(ctx, resource, blueprintName, outs, r.mapper)

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

		outs.AddOutput(resource.Name, out)

		previousResourceStatus := resourceStatuses.GetPreviousResourceStatus(resource.Name)

		var realizedResource *v1alpha1.RealizedResource

		var additionalConditions []metav1.Condition
		if (stampedObject == nil || template == nil) && previousResourceStatus != nil {
			realizedResource = &previousResourceStatus.RealizedResource
			if previousResourceStatusHealthyCondition := utils.ConditionList(previousResourceStatus.Conditions).ConditionWithType(v1alpha1.ResourceHealthy); previousResourceStatusHealthyCondition != nil {
				additionalConditions = []metav1.Condition{*previousResourceStatusHealthyCondition}
			}
		} else {
			var previousRealizedResource *v1alpha1.RealizedResource
			if previousResourceStatus != nil {
				previousRealizedResource = &previousResourceStatus.RealizedResource
			}
			realizedResource = r.generateRealizedResource(ctx, resource, template, stampedObject, out, previousRealizedResource)
			var previousOutputs []v1alpha1.Output
			if previousRealizedResource != nil {
				previousOutputs = previousRealizedResource.Outputs
			}
			if !reflect.DeepEqual(previousOutputs, realizedResource.Outputs) {
				rec := events.FromContextOrDie(ctx)
				rec.ResourceEventf(events.NormalType, events.ResourceOutputChangedReason, "[%s] found a new output in [%Q]", stampedObject, realizedResource.Name)
			}
			if template != nil {
				additionalConditions = []metav1.Condition{r.healthyConditionEvaluator(template.GetHealthRule(), realizedResource, stampedObject)}
			}
		}
		resourceStatuses.Add(realizedResource, err, additionalConditions...)
		if slices.Contains(resourceStatuses.ChangedConditionTypes(realizedResource.Name), v1alpha1.ResourceHealthy) {
			newStatus := metav1.ConditionUnknown
			newHealthyCondition := resourceStatuses.GetCurrent().ConditionsForResourceNamed(realizedResource.Name).ConditionWithType(v1alpha1.ResourceHealthy)
			if newHealthyCondition != nil {
				newStatus = newHealthyCondition.Status
			}
			events.FromContextOrDie(ctx).ResourceEventf(events.NormalType, events.ResourceHealthyStatusChangedReason, "[%s] found healthy status in [%Q] changed to [%s]", stampedObject, realizedResource.Name, newStatus)
		}
	}

	return firstError
}

func (r *realizer) generateRealizedResource(ctx context.Context, resource OwnerResource, template templates.Template, stampedObject *unstructured.Unstructured, output *templates.Output, previousRealizedResource *v1alpha1.RealizedResource) *v1alpha1.RealizedResource {
	log := logr.FromContextOrDiscard(ctx)

	if previousRealizedResource == nil {
		previousRealizedResource = &v1alpha1.RealizedResource{}
	}

	var inputs []v1alpha1.Input
	for _, source := range resource.Sources {
		inputs = append(inputs, v1alpha1.Input{Name: source.Resource})
	}

	for _, image := range resource.Images {
		inputs = append(inputs, v1alpha1.Input{Name: image.Resource})
	}

	if resource.Deployment != nil {
		inputs = append(inputs, v1alpha1.Input{Name: resource.Deployment.Resource})
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

		outputs = getOutputs(template, previousRealizedResource, output)
	}

	var stampedRef *v1alpha1.StampedRef
	if stampedObject != nil {
		qualifiedResource, err := utils.GetQualifiedResource(r.mapper, stampedObject)
		if err != nil {
			log.Error(err, "failed to retrieve qualified resource name", "object", stampedObject)
			qualifiedResource = "could not fetch - see logs for 'failed to retrieve qualified resource name'"
		}

		stampedRef = &v1alpha1.StampedRef{
			ObjectReference: &corev1.ObjectReference{
				Kind:       stampedObject.GetKind(),
				Namespace:  stampedObject.GetNamespace(),
				Name:       stampedObject.GetName(),
				APIVersion: stampedObject.GetAPIVersion(),
			},
			Resource: qualifiedResource,
		}
	}

	return &v1alpha1.RealizedResource{
		Name:        resource.Name,
		StampedRef:  stampedRef,
		TemplateRef: templateRef,
		Inputs:      inputs,
		Outputs:     outputs,
	}
}

func getOutputs(template templates.Template, previousRealizedResource *v1alpha1.RealizedResource, output *templates.Output) []v1alpha1.Output {
	outputs, err := template.GenerateResourceOutput(output)
	if err != nil {
		outputs = previousRealizedResource.Outputs
	} else {
		currTime := metav1.NewTime(time.Now())
		for j, out := range outputs {
			outputs[j].LastTransitionTime = currTime
			for _, previousOutput := range previousRealizedResource.Outputs {
				if previousOutput.Name == out.Name {
					if previousOutput.Digest == out.Digest {
						outputs[j].LastTransitionTime = previousOutput.LastTransitionTime
					}
					break
				}
			}
		}
	}

	return outputs
}
