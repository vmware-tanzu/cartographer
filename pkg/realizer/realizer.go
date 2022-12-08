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
	"crypto/sha256"
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/strings"
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

//counterfeiter:generate . ResourceRealizer
type ResourceRealizer interface {
	Do(ctx context.Context, resource OwnerResource, blueprintName string, outputs Outputs, mapper meta.RESTMapper) (templates.Reader, *unstructured.Unstructured, *templates.Output, bool, string, error)
}

type realizer struct {
	healthyConditionEvaluator HealthyConditionEvaluator
	mapper                    meta.RESTMapper
}

type HealthyConditionEvaluator func(rule *v1alpha1.HealthRule, realizedResource *v1alpha1.RealizedResource, stampedObject *unstructured.Unstructured) metav1.Condition

//counterfeiter:generate k8s.io/apimachinery/pkg/api/meta.RESTMapper
func NewRealizer(healthyConditionEvaluator HealthyConditionEvaluator, mapper meta.RESTMapper) *realizer {
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
		template, stampedObject, out, isPassThrough, templateName, err := resourceRealizer.Do(ctx, resource, blueprintName, outs, r.mapper)

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
			realizedResource = r.generateRealizedResource(ctx, resource, template, stampedObject, out, previousRealizedResource, isPassThrough, templateName)

			var previousOutputs []v1alpha1.Output
			if previousRealizedResource != nil {
				previousOutputs = previousRealizedResource.Outputs
			}

			if !reflect.DeepEqual(previousOutputs, realizedResource.Outputs) {
				rec := events.FromContextOrDie(ctx)
				if isPassThrough {
					rec.Eventf(events.NormalType, events.ResourceOutputChangedReason, "[%s] passed through a new output", realizedResource.Name)
				} else {
					rec.ResourceEventf(events.NormalType, events.ResourceOutputChangedReason, "[%s] found a new output in [%Q]", stampedObject, realizedResource.Name)
				}
			}

			if template != nil {
				additionalConditions = []metav1.Condition{r.healthyConditionEvaluator(template.GetHealthRule(), realizedResource, stampedObject)}
			}
		}
		resourceStatuses.Add(realizedResource, err, isPassThrough, additionalConditions...)
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

func (r *realizer) generateRealizedResource(ctx context.Context, resource OwnerResource, template templates.Reader,
	stampedObject *unstructured.Unstructured, output *templates.Output, previousRealizedResource *v1alpha1.RealizedResource,
	isPassThrough bool, templateName string) *v1alpha1.RealizedResource {
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
			Kind:       resource.TemplateRef.Kind,
			Name:       templateName,
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		}
		outputs = getOutputs(previousRealizedResource, output)
	}

	if isPassThrough {
		outputs = getOutputs(previousRealizedResource, output)
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
			Resource:  qualifiedResource,
			Lifecycle: stampedObject.GetLabels()["carto.run/template-lifecycle"],
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

func getOutputs(previousRealizedResource *v1alpha1.RealizedResource, output *templates.Output) []v1alpha1.Output {
	outputs, err := generateResourceOutput(output)
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

// TODO: This should be polymorphic

func generateResourceOutput(output *templates.Output) ([]v1alpha1.Output, error) {
	if output == nil {
		return nil, nil
	}

	var result []v1alpha1.Output

	if output.Source != nil {
		urlOut, err := buildOneOutput("url", output.Source.URL)
		if err != nil {
			return nil, err
		}
		result = append(result, urlOut)

		revisionOut, err := buildOneOutput("revision", output.Source.Revision)
		if err != nil {
			return nil, err
		}
		result = append(result, revisionOut)
	} else if output.Image != nil {
		out, err := buildOneOutput("image", output.Image)
		if err != nil {
			return nil, err
		}
		result = append(result, out)
	} else if output.Config != nil {
		out, err := buildOneOutput("config", output.Config)
		if err != nil {
			return nil, err
		}
		result = append(result, out)

	} else {
		return nil, nil
	}
	return result, nil
}

const PreviewCharacterLimit = 1024

func buildOneOutput(name string, value any) (v1alpha1.Output, error) {
	bytes, err := yaml.Marshal(value)
	if err != nil {
		return v1alpha1.Output{}, err
	}

	sha := sha256.Sum256(bytes)

	return v1alpha1.Output{
		Name:    name,
		Preview: strings.ShortenString(string(bytes), PreviewCharacterLimit),
		Digest:  fmt.Sprintf("sha256:%x", sha),
	}, nil

}
