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

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/errors"
	"github.com/vmware-tanzu/cartographer/pkg/logger"
	realizerclient "github.com/vmware-tanzu/cartographer/pkg/realizer/client"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/healthcheck"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/runnable/gc"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/selector"
	"github.com/vmware-tanzu/cartographer/pkg/stamp"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

//go:generate go run -modfile ../../hack/tools/go.mod github.com/maxbrunsfeld/counterfeiter/v6 -generate

type ContextGenerator interface {
	Generate(templateParams TemplateParams, resource OwnerResource, outputs OutputsGetter) map[string]interface{}
}

type resourceRealizer struct {
	owner             client.Object
	systemRepo        repository.Repository
	ownerRepo         repository.Repository
	templatingContext ContextGenerator
	resourceLabeler   ResourceLabeler
}

type ResourceLabeler func(resource OwnerResource) templates.Labels

type ResourceRealizerBuilder func(authToken string, owner client.Object, templatingContext ContextGenerator, systemRepo repository.Repository, resourceLabeler ResourceLabeler) (ResourceRealizer, error)

//counterfeiter:generate sigs.k8s.io/controller-runtime/pkg/client.Client
func NewResourceRealizerBuilder(repositoryBuilder repository.RepositoryBuilder, clientBuilder realizerclient.ClientBuilder, cache repository.RepoCache) ResourceRealizerBuilder {
	return func(authToken string, owner client.Object, templatingContext ContextGenerator, systemRepo repository.Repository, resourceLabeler ResourceLabeler) (ResourceRealizer, error) {
		ownerClient, _, err := clientBuilder(authToken, false)
		if err != nil {
			return nil, fmt.Errorf("can't build client: %w", err)
		}

		ownerRepo := repositoryBuilder(ownerClient, cache)

		return &resourceRealizer{
			owner:             owner,
			systemRepo:        systemRepo,
			ownerRepo:         ownerRepo,
			templatingContext: templatingContext,
			resourceLabeler:   resourceLabeler,
		}, nil
	}
}

func (r *resourceRealizer) Do(ctx context.Context, resource OwnerResource, blueprintName string, outputs Outputs, mapper meta.RESTMapper) (templates.Reader, *unstructured.Unstructured, *templates.Output, bool, string, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues("template", resource.TemplateRef)
	ctx = logr.NewContext(ctx, log)

	var templateName string
	var templateOption v1alpha1.TemplateOption
	var stampReader stamp.Outputter
	var stampedObject *unstructured.Unstructured
	var template templates.Reader
	var output *templates.Output
	var apiTemplate client.Object
	var err error
	var allRunnableStampedObjects []*unstructured.Unstructured
	var qualifiedResource string

	passThrough := false

	// TODO: consider: should we build this only once, and pass it to the contextGenerator also?
	inputGenerator := NewInputGenerator(resource, outputs)

	if len(resource.TemplateOptions) > 0 {
		var err error
		templateOption, err = r.findMatchingTemplateOption(resource, blueprintName)
		if err != nil {
			return nil, nil, nil, passThrough, templateName, err
		}
		if templateOption.PassThrough != "" {
			passThrough = true
		} else {
			templateName = templateOption.Name
		}
	} else {
		templateName = resource.TemplateRef.Name
	}

	if passThrough {
		log.V(logger.DEBUG).Info("pass through template", "passThrough", fmt.Sprintf("[%s]", templateOption.PassThrough))

		stampReader, err = stamp.NewPassThroughReader(resource.TemplateRef.Kind, templateOption.PassThrough, inputGenerator)
		if err != nil {
			log.Error(err, "failed to create new stamp pass through reader")
			return nil, nil, nil, passThrough, templateName, fmt.Errorf("failed to create new stamp pass through reader: %w", err)
		}

		output, err = stampReader.Output(stampedObject)
		if err != nil {
			log.Error(err, "failed to retrieve output from pass through", "passThrough", templateOption.PassThrough)
		}
	} else {
		log.V(logger.DEBUG).Info("realizing template", "template", fmt.Sprintf("[%s/%s]", resource.TemplateRef.Kind, templateName))

		apiTemplate, err = r.systemRepo.GetTemplate(ctx, templateName, resource.TemplateRef.Kind)
		if err != nil {
			log.Error(err, "failed to get cluster template")
			return nil, nil, nil, passThrough, templateName, errors.GetTemplateError{
				Err:           err,
				ResourceName:  resource.Name,
				TemplateName:  templateName,
				BlueprintName: blueprintName,
				BlueprintType: errors.SupplyChain,
			}
		}

		template, err = templates.NewReaderFromAPI(apiTemplate)
		if err != nil {
			log.Error(err, "failed to get cluster template")
			return nil, nil, nil, passThrough, templateName, fmt.Errorf("failed to get cluster template [%+v]: %w", resource.TemplateRef, err)
		}

		labels := r.resourceLabeler(resource)
		labelsWithLifecycle := make(map[string]string)
		for k, v := range labels {
			labelsWithLifecycle[k] = v
		}
		labelsWithLifecycle["carto.run/template-lifecycle"] = string(*template.GetLifecycle())

		stamper := templates.StamperBuilder(r.owner, r.templatingContext.Generate(template, resource, outputs), labelsWithLifecycle)
		stampedObject, err = stamper.Stamp(ctx, template.GetResourceTemplate())
		if err != nil {
			log.Error(err, "failed to stamp resource")
			return template, nil, nil, passThrough, templateName, errors.StampError{
				Err:           err,
				TemplateName:  templateName,
				TemplateKind:  resource.TemplateRef.Kind,
				ResourceName:  resource.Name,
				BlueprintName: blueprintName,
				BlueprintType: errors.SupplyChain,
			}
		}

		stampReader, err = stamp.NewReader(apiTemplate, inputGenerator)
		if err != nil {
			log.Error(err, "failed to create new stamp reader")
			return nil, nil, nil, passThrough, templateName, fmt.Errorf("failed to create new stamp reader: %w", err)
		}

		if template.GetLifecycle().IsImmutable() {
			err = r.ownerRepo.EnsureImmutableObjectExistsOnCluster(ctx, stampedObject, labelsWithLifecycle)
			if err != nil {
				log.Error(err, "failed to ensure object exists on cluster", "object", stampedObject)
				return template, nil, nil, passThrough, templateName, errors.ApplyStampedObjectError{
					Err:           err,
					StampedObject: stampedObject,
					ResourceName:  resource.Name,
					BlueprintName: blueprintName,
					BlueprintType: errors.SupplyChain,
				}
			}

			allRunnableStampedObjects, err = r.ownerRepo.ListUnstructured(ctx, stampedObject.GroupVersionKind(), stampedObject.GetNamespace(), labels)
			if err != nil {
				log.Error(err, "failed to list objects")
				return template, nil, nil, passThrough, templateName, errors.ListCreatedObjectsError{
					Err:       err,
					Namespace: stampedObject.GetNamespace(),
					Labels:    labels,
				}
			}

			healthRule := template.GetHealthRule()
			if healthRule == nil && *template.GetLifecycle() == templates.Tekton {
				healthRule = &v1alpha1.HealthRule{SingleConditionType: "Succeeded"}
			}

			var examinedObjects []*stamp.ExaminedObject

			for _, someStampedObject := range allRunnableStampedObjects {
				health := healthcheck.DetermineStampedObjectHealth(healthRule, someStampedObject)

				examinedObjects = append(examinedObjects, &stamp.ExaminedObject{
					StampedObject: someStampedObject,
					Health:        health,
				})
			}

			gc.CleanupRunnableStampedObjects(ctx, examinedObjects, template.GetRetentionPolicy(), r.ownerRepo)

			latestSuccessfulObject := stamp.GetLatestSuccessfulObjFromExaminedObject(examinedObjects)
			if latestSuccessfulObject == nil {
				for _, obj := range allRunnableStampedObjects {
					log.V(logger.DEBUG).Info("failed to retrieve output from any object", "considered", obj)
				}
			}

			output, err = stampReader.Output(latestSuccessfulObject)
		} else {
			err = r.ownerRepo.EnsureMutableObjectExistsOnCluster(ctx, stampedObject)
			if err != nil {
				log.Error(err, "failed to ensure object exists on cluster", "object", stampedObject)
				return template, nil, nil, passThrough, templateName, errors.ApplyStampedObjectError{
					Err:           err,
					StampedObject: stampedObject,
					ResourceName:  resource.Name,
					BlueprintName: blueprintName,
					BlueprintType: errors.SupplyChain,
				}
			}

			output, err = stampReader.Output(stampedObject)

			if err != nil {
				log.Error(err, "failed to retrieve output from object", "object", stampedObject)
			}
		}

		if err != nil {
			var rErr error
			qualifiedResource, rErr = utils.GetQualifiedResource(mapper, stampedObject)
			if rErr != nil {
				log.Error(err, "failed to retrieve qualified resource name", "object", stampedObject)
				qualifiedResource = "could not fetch - see the log line for 'failed to retrieve qualified resource name'"
			}
		}
	}

	if err != nil {
		return template, stampedObject, nil, passThrough, templateName, errors.RetrieveOutputError{
			Err:               err,
			ResourceName:      resource.Name,
			StampedObject:     stampedObject,
			BlueprintName:     blueprintName,
			BlueprintType:     errors.SupplyChain,
			QualifiedResource: qualifiedResource,
			PassThroughInput:  templateOption.PassThrough,
		}
	}

	return template, stampedObject, output, passThrough, templateName, nil
}

func (r *resourceRealizer) findMatchingTemplateOption(resource OwnerResource, supplyChainName string) (v1alpha1.TemplateOption, error) {
	bestMatchingTemplateOptionsIndices, err := selector.BestSelectorMatchIndices(r.owner, v1alpha1.TemplateOptionSelectors(resource.TemplateOptions))

	if err != nil {
		return v1alpha1.TemplateOption{}, errors.ResolveTemplateOptionError{
			Err:           err,
			ResourceName:  resource.Name,
			OptionName:    resource.TemplateOptions[err.SelectorIndex()].Name,
			BlueprintName: supplyChainName,
			BlueprintType: errors.SupplyChain,
		}
	}

	if len(bestMatchingTemplateOptionsIndices) != 1 {
		var optionNames []string
		for _, optionIndex := range bestMatchingTemplateOptionsIndices {
			optionNames = append(optionNames, resource.TemplateOptions[optionIndex].Name)
		}

		return v1alpha1.TemplateOption{}, errors.TemplateOptionsMatchError{
			ResourceName:  resource.Name,
			OptionNames:   optionNames,
			BlueprintName: supplyChainName,
			BlueprintType: errors.SupplyChain,
		}
	}

	return resource.TemplateOptions[bestMatchingTemplateOptionsIndices[0]], nil
}
