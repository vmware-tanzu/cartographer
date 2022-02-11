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

import (
	"context"
	"fmt"
	"github.com/vmware-tanzu/cartographer/pkg/eval"
	"github.com/vmware-tanzu/cartographer/pkg/logger"
	"github.com/vmware-tanzu/cartographer/pkg/selector"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	realizerclient "github.com/vmware-tanzu/cartographer/pkg/realizer/client"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

//go:generate go run -modfile ../../../hack/tools/go.mod github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . ResourceRealizer
type ResourceRealizer interface {
	Do(ctx context.Context, resource *v1alpha1.DeliveryResource, deliveryName string, outputs Outputs) (*unstructured.Unstructured, *templates.Output, error)
}

type resourceRealizer struct {
	deliverable     *v1alpha1.Deliverable
	systemRepo      repository.Repository
	deliverableRepo repository.Repository
	deliveryParams  []v1alpha1.BlueprintParam
}

type ResourceRealizerBuilder func(secret *corev1.Secret, deliverable *v1alpha1.Deliverable, repo repository.Repository, deliveryParams []v1alpha1.BlueprintParam) (ResourceRealizer, error)

func NewResourceRealizerBuilder(repositoryBuilder repository.RepositoryBuilder, clientBuilder realizerclient.ClientBuilder, cache repository.RepoCache) ResourceRealizerBuilder {
	return func(secret *corev1.Secret, deliverable *v1alpha1.Deliverable, systemRepo repository.Repository, deliveryParams []v1alpha1.BlueprintParam) (ResourceRealizer, error) {
		client, err := clientBuilder(secret)
		if err != nil {
			return nil, fmt.Errorf("can't build client: %w", err)
		}
		deliverableRepo := repositoryBuilder(client, cache)
		return &resourceRealizer{
			deliverable:     deliverable,
			systemRepo:      systemRepo,
			deliverableRepo: deliverableRepo,
			deliveryParams:  deliveryParams,
		}, nil
	}
}

func (r *resourceRealizer) Do(ctx context.Context, resource *v1alpha1.DeliveryResource, deliveryName string, outputs Outputs) (*unstructured.Unstructured, *templates.Output, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues("template", resource.TemplateRef)
	ctx = logr.NewContext(ctx, log)

	var templateName string
	var err error
	if len(resource.TemplateRef.Options) > 0 {
		templateName, err = r.findMatchingTemplateName(resource, deliveryName)
		if err != nil {
			return nil, nil, err
		}
	} else {
		templateName = resource.TemplateRef.Name
	}

	log.V(logger.DEBUG).Info("realizing template", "template", fmt.Sprintf("[%s/%s]", resource.TemplateRef.Kind, templateName))

	apiTemplate, err := r.systemRepo.GetDeliveryTemplate(ctx, templateName, resource.TemplateRef.Kind)
	if err != nil {
		log.Error(err, "failed to get delivery cluster template")
		return nil, nil, GetDeliveryTemplateError{
			Err:          err,
			DeliveryName: deliveryName,
			Resource:     resource,
		}
	}

	template, err := templates.NewModelFromAPI(apiTemplate)
	if err != nil {
		log.Error(err, "failed to get delivery cluster template")
		return nil, nil, fmt.Errorf("failed to get delivery cluster template [%+v]: %w", resource.TemplateRef, err)
	}

	labels := map[string]string{
		"carto.run/deliverable-name":      r.deliverable.Name,
		"carto.run/deliverable-namespace": r.deliverable.Namespace,
		"carto.run/cluster-delivery-name": deliveryName,
		"carto.run/resource-name":         resource.Name,
		"carto.run/template-kind":         template.GetKind(),
		"carto.run/cluster-template-name": template.GetName(),
	}

	inputs := outputs.GenerateInputs(resource)
	templatingContext := map[string]interface{}{
		"deliverable": r.deliverable,
		"params":      templates.ParamsBuilder(template.GetDefaultParams(), r.deliveryParams, resource.Params, r.deliverable.Spec.Params),
		"sources":     inputs.Sources,
		"configs":     inputs.Configs,
		"deployment":  inputs.Deployment,
	}

	// Todo: this belongs in Stamp.
	if inputs.OnlyConfig() != nil {
		templatingContext["config"] = inputs.OnlyConfig()
	}
	if inputs.OnlySource() != nil {
		templatingContext["source"] = inputs.OnlySource()
	}

	stampContext := templates.StamperBuilder(r.deliverable, templatingContext, labels)
	stampedObject, err := stampContext.Stamp(ctx, template.GetResourceTemplate())
	if err != nil {
		log.Error(err, "failed to stamp resource")
		return nil, nil, StampError{
			Err:          err,
			Resource:     resource,
			DeliveryName: deliveryName,
		}
	}

	err = r.deliverableRepo.EnsureMutableObjectExistsOnCluster(ctx, stampedObject)
	if err != nil {
		log.Error(err, "failed to ensure object exists on cluster", "object", stampedObject)
		return nil, nil, ApplyStampedObjectError{
			Err:           err,
			StampedObject: stampedObject,
			DeliveryName:  deliveryName,
			Resource:      resource,
		}
	}

	template.SetInputs(inputs)
	template.SetStampedObject(stampedObject)

	output, err := template.GetOutput()
	if err != nil {
		log.Error(err, "failed to retrieve output from object", "object", stampedObject)
		return stampedObject, nil, RetrieveOutputError{
			Err:           err,
			Resource:      resource,
			DeliveryName:  deliveryName,
			StampedObject: stampedObject,
		}
	}

	return stampedObject, output, nil
}

func (r *resourceRealizer) findMatchingTemplateName(resource *v1alpha1.DeliveryResource, deliveryName string) (string, error) {
	var templateName string
	var matchingOptions []string

	for _, option := range resource.TemplateRef.Options {
		matchedAllFields := true
		for _, field := range option.Selector.MatchFields {
			dlContext := map[string]interface{}{
				"deliverable": r.deliverable,
			}
			matched, err := selector.Matches(field, dlContext)
			if err != nil {
				if _, ok := err.(eval.JsonPathDoesNotExistError); !ok {
					return "", ResolveTemplateOptionError{
						Err:          err,
						DeliveryName: deliveryName,
						Resource:     resource,
						OptionName:   option.Name,
						Key:          field.Key,
					}
				}
			}
			if !matched {
				matchedAllFields = false
				break
			}
		}
		if matchedAllFields {
			matchingOptions = append(matchingOptions, option.Name)
		}
	}

	if len(matchingOptions) != 1 {
		return "", TemplateOptionsMatchError{
			DeliveryName: deliveryName,
			Resource:     resource,
			OptionNames:  matchingOptions,
		}
	} else {
		templateName = matchingOptions[0]
	}

	return templateName, nil
}
