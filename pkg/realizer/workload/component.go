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

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/eval"
	"github.com/vmware-tanzu/cartographer/pkg/logger"
	realizerclient "github.com/vmware-tanzu/cartographer/pkg/realizer/client"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/selector"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

//go:generate go run -modfile ../../../hack/tools/go.mod github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . ResourceRealizer
type ResourceRealizer interface {
	Do(ctx context.Context, resource *v1alpha1.SupplyChainResource, supplyChainName string, outputs Outputs) (templates.Template, *unstructured.Unstructured, *templates.Output, error)
}

type resourceRealizer struct {
	workload          *v1alpha1.Workload
	systemRepo        repository.Repository
	workloadRepo      repository.Repository
	supplyChainParams []v1alpha1.BlueprintParam
}

type ResourceRealizerBuilder func(secret *corev1.Secret, workload *v1alpha1.Workload, systemRepo repository.Repository, supplyChainParams []v1alpha1.BlueprintParam) (ResourceRealizer, error)

//counterfeiter:generate sigs.k8s.io/controller-runtime/pkg/client.Client
func NewResourceRealizerBuilder(repositoryBuilder repository.RepositoryBuilder, clientBuilder realizerclient.ClientBuilder, cache repository.RepoCache) ResourceRealizerBuilder {
	return func(secret *corev1.Secret, workload *v1alpha1.Workload, systemRepo repository.Repository, supplyChainParams []v1alpha1.BlueprintParam) (ResourceRealizer, error) {
		workloadClient, err := clientBuilder(secret)
		if err != nil {
			return nil, fmt.Errorf("can't build client: %w", err)
		}

		workloadRepo := repositoryBuilder(workloadClient, cache)

		return &resourceRealizer{
			workload:          workload,
			systemRepo:        systemRepo,
			workloadRepo:      workloadRepo,
			supplyChainParams: supplyChainParams,
		}, nil
	}
}

func (r *resourceRealizer) Do(ctx context.Context, resource *v1alpha1.SupplyChainResource, supplyChainName string, outputs Outputs) (templates.Template, *unstructured.Unstructured, *templates.Output, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues("template", resource.TemplateRef)
	ctx = logr.NewContext(ctx, log)

	var templateName string
	var err error
	if len(resource.TemplateRef.Options) > 0 {
		templateName, err = r.findMatchingTemplateName(resource, supplyChainName)
		if err != nil {
			return nil, nil, nil, err
		}
	} else {
		templateName = resource.TemplateRef.Name
	}

	log.V(logger.DEBUG).Info("realizing template", "template", fmt.Sprintf("[%s/%s]", resource.TemplateRef.Kind, templateName))

	apiTemplate, err := r.systemRepo.GetSupplyChainTemplate(ctx, templateName, resource.TemplateRef.Kind)
	if err != nil {
		log.Error(err, "failed to get cluster template")
		return nil, nil, nil, GetSupplyChainTemplateError{
			Err:             err,
			SupplyChainName: supplyChainName,
			Resource:        resource,
		}
	}

	template, err := templates.NewModelFromAPI(apiTemplate)
	if err != nil {
		log.Error(err, "failed to get cluster template")
		return nil, nil, nil, fmt.Errorf("failed to get cluster template [%+v]: %w", resource.TemplateRef, err)
	}

	labels := map[string]string{
		"carto.run/workload-name":             r.workload.Name,
		"carto.run/workload-namespace":        r.workload.Namespace,
		"carto.run/cluster-supply-chain-name": supplyChainName,
		"carto.run/resource-name":             resource.Name,
		"carto.run/template-kind":             template.GetKind(),
		"carto.run/cluster-template-name":     template.GetName(),
	}

	inputs := outputs.GenerateInputs(resource)
	workloadTemplatingContext := map[string]interface{}{
		"workload": r.workload,
		"params":   templates.ParamsBuilder(template.GetDefaultParams(), r.supplyChainParams, resource.Params, r.workload.Spec.Params),
		"sources":  inputs.Sources,
		"images":   inputs.Images,
		"configs":  inputs.Configs,
	}

	// Todo: this belongs in Stamp.
	if inputs.OnlyConfig() != nil {
		workloadTemplatingContext["config"] = inputs.OnlyConfig()
	}
	if inputs.OnlyImage() != nil {
		workloadTemplatingContext["image"] = inputs.OnlyImage()
	}
	if inputs.OnlySource() != nil {
		workloadTemplatingContext["source"] = inputs.OnlySource()
	}

	stampContext := templates.StamperBuilder(r.workload, workloadTemplatingContext, labels)
	stampedObject, err := stampContext.Stamp(ctx, template.GetResourceTemplate())
	if err != nil {
		log.Error(err, "failed to stamp resource")
		return template, nil, nil, StampError{
			Err:             err,
			Resource:        resource,
			SupplyChainName: supplyChainName,
		}
	}

	err = r.workloadRepo.EnsureMutableObjectExistsOnCluster(ctx, stampedObject)
	if err != nil {
		log.Error(err, "failed to ensure object exists on cluster", "object", stampedObject)
		return template, nil, nil, ApplyStampedObjectError{
			Err:             err,
			StampedObject:   stampedObject,
			SupplyChainName: supplyChainName,
			Resource:        resource,
		}
	}

	template.SetStampedObject(stampedObject)

	output, err := template.GetOutput()
	if err != nil {
		log.Error(err, "failed to retrieve output from object", "object", stampedObject)
		return template, stampedObject, nil, RetrieveOutputError{
			Err:             err,
			Resource:        resource,
			SupplyChainName: supplyChainName,
			StampedObject:   stampedObject,
		}
	}

	return template, stampedObject, output, nil
}

func (r *resourceRealizer) findMatchingTemplateName(resource *v1alpha1.SupplyChainResource, supplyChainName string) (string, error) {
	var templateName string
	var matchingOptions []string

	for _, option := range resource.TemplateRef.Options {
		matchedAllFields := true
		for _, field := range option.Selector.MatchFields {
			wkContext := map[string]interface{}{
				"workload": r.workload,
			}
			matched, err := selector.Matches(field, wkContext)
			if err != nil {
				if _, ok := err.(eval.JsonPathDoesNotExistError); !ok {
					return "", ResolveTemplateOptionError{
						Err:             err,
						SupplyChainName: supplyChainName,
						Resource:        resource,
						OptionName:      option.Name,
						Key:             field.Key,
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
			SupplyChainName: supplyChainName,
			Resource:        resource,
			OptionNames:     matchingOptions,
		}
	} else {
		templateName = matchingOptions[0]
	}

	return templateName, nil
}
