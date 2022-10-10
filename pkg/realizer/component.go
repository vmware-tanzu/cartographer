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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/errors"
	"github.com/vmware-tanzu/cartographer/pkg/logger"
	realizerclient "github.com/vmware-tanzu/cartographer/pkg/realizer/client"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/selector"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

//go:generate go run -modfile ../../hack/tools/go.mod github.com/maxbrunsfeld/counterfeiter/v6 -generate

type OwnerResource struct {
	TemplateRef     v1alpha1.TemplateReference
	TemplateOptions []v1alpha1.TemplateOption
	Params          []v1alpha1.BlueprintParam
	Name            string
	Sources         []v1alpha1.ResourceReference
	Images          []v1alpha1.ResourceReference
	Configs         []v1alpha1.ResourceReference
	Deployment      *v1alpha1.DeploymentReference
}

func (o OwnerResource) GetImages() []v1alpha1.ResourceReference {
	return o.Images
}

func (o OwnerResource) GetConfigs() []v1alpha1.ResourceReference {
	return o.Configs
}

func (o OwnerResource) GetDeployment() *v1alpha1.DeploymentReference {
	return o.Deployment
}

func (o OwnerResource) GetName() string {
	return o.Name
}

func (o OwnerResource) GetSources() []v1alpha1.ResourceReference {
	return o.Sources
}

//counterfeiter:generate . ResourceRealizer
type ResourceRealizer interface {
	Do(ctx context.Context, resource OwnerResource, blueprintName string, outputs Outputs) (templates.Reader, *unstructured.Unstructured, *templates.Output, error)
}

type resourceRealizer struct {
	owner           client.Object
	ownerParams     []v1alpha1.OwnerParam
	systemRepo      repository.Repository
	ownerRepo       repository.Repository
	blueprintParams []v1alpha1.BlueprintParam
	resourceLabeler ResourceLabeler
}

type ResourceLabeler func(resource OwnerResource) templates.Labels

type ResourceRealizerBuilder func(authToken string, owner client.Object, ownerParams []v1alpha1.OwnerParam, systemRepo repository.Repository, blueprintParams []v1alpha1.BlueprintParam, resourceLabeler ResourceLabeler) (ResourceRealizer, error)

//counterfeiter:generate sigs.k8s.io/controller-runtime/pkg/client.Client
func NewResourceRealizerBuilder(repositoryBuilder repository.RepositoryBuilder, clientBuilder realizerclient.ClientBuilder, cache repository.RepoCache) ResourceRealizerBuilder {
	return func(authToken string, owner client.Object, ownerParams []v1alpha1.OwnerParam, systemRepo repository.Repository, supplyChainParams []v1alpha1.BlueprintParam, resourceLabeler ResourceLabeler) (ResourceRealizer, error) {
		ownerClient, _, err := clientBuilder(authToken, false)
		if err != nil {
			return nil, fmt.Errorf("can't build client: %w", err)
		}

		ownerRepo := repositoryBuilder(ownerClient, cache)

		return &resourceRealizer{
			owner:           owner,
			ownerParams:     ownerParams,
			systemRepo:      systemRepo,
			ownerRepo:       ownerRepo,
			blueprintParams: supplyChainParams,
			resourceLabeler: resourceLabeler,
		}, nil
	}
}

func (r *resourceRealizer) Do(ctx context.Context, resource OwnerResource, blueprintName string, outputs Outputs) (templates.Reader, *unstructured.Unstructured, *templates.Output, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues("template", resource.TemplateRef)
	ctx = logr.NewContext(ctx, log)

	var templateName string
	var err error
	if len(resource.TemplateOptions) > 0 {
		templateName, err = r.findMatchingTemplateName(resource, blueprintName)
		if err != nil {
			return nil, nil, nil, err
		}
	} else {
		templateName = resource.TemplateRef.Name
	}

	log.V(logger.DEBUG).Info("realizing template", "template", fmt.Sprintf("[%s/%s]", resource.TemplateRef.Kind, templateName))

	apiTemplate, err := r.systemRepo.GetTemplate(ctx, templateName, resource.TemplateRef.Kind)
	if err != nil {
		log.Error(err, "failed to get cluster template")
		return nil, nil, nil, errors.GetTemplateError{
			Err:           err,
			ResourceName:  resource.Name,
			TemplateName:  templateName,
			BlueprintName: blueprintName,
			BlueprintType: errors.SupplyChain,
		}
	}

	template, err := templates.NewReaderFromAPI(apiTemplate)
	if err != nil {
		log.Error(err, "failed to get cluster template")
		return nil, nil, nil, fmt.Errorf("failed to get cluster template [%+v]: %w", resource.TemplateRef, err)
	}

	labels := r.resourceLabeler(resource)

	// TODO: lift me out of component.go
	ownerTemplatingContext := NewContextGenerator(r.owner, r.ownerParams, r.blueprintParams)

	stamper := templates.StamperBuilder(r.owner, ownerTemplatingContext.Generate(template, resource, outputs), labels)
	stampedObject, err := stamper.Stamp(ctx, template.GetResourceTemplate())
	if err != nil {
		log.Error(err, "failed to stamp resource")
		return template, nil, nil, errors.StampError{
			Err:           err,
			TemplateName:  templateName,
			TemplateKind:  resource.TemplateRef.Kind,
			ResourceName:  resource.Name,
			BlueprintName: blueprintName,
			BlueprintType: errors.SupplyChain,
		}
	}

	err = r.ownerRepo.EnsureMutableObjectExistsOnCluster(ctx, stampedObject)
	if err != nil {
		log.Error(err, "failed to ensure object exists on cluster", "object", stampedObject)
		return template, nil, nil, errors.ApplyStampedObjectError{
			Err:           err,
			StampedObject: stampedObject,
			ResourceName:  resource.Name,
			BlueprintName: blueprintName,
			BlueprintType: errors.SupplyChain,
		}
	}

	// This is for deployment pass through
	// TODO we need to lose this input generator
	inputGenerator := NewInputGenerator(resource, outputs)
	//template.SetInputs(inputGenerator)
	// FIXME why?!

	switch kind {
	case "ClusterDeploymentTemplate":
		outputReader = NewClusterDeploymentOutputReader(stampedObject, inputGenerator)
	case "ClusterSourceTemplate":
		outputReader = NewClusterSourceOutputReader(outputPathGetter, stampedObject)
	case "ClusterImageTemplate":
		outputReader = NewClusterImageOutputReader(outputPathGetter, stampedObject)
	case "ClusterConfigTemplate":
		outputReader = NewClusterConfigOutputReader(outputPathGetter, stampedObject)
	case "ClusterTemplate":
		outputReader = NewNoOutputReader()
	case "Passthrough":
		outputReader = NewPassthroughOutputReader(inputType, inputGenerator)
	}

	output, err := outputReader.Read()

	//output, err := template.GetOutput(stampedObject, inputGenerator)
	if err != nil {
		log.Error(err, "failed to retrieve output from object", "object", stampedObject)
		return template, stampedObject, nil, errors.RetrieveOutputError{
			Err:           err,
			ResourceName:  resource.Name,
			StampedObject: stampedObject,
			BlueprintName: blueprintName,
			BlueprintType: errors.SupplyChain,
		}
	}

	return template, stampedObject, output, nil
}

func (r *resourceRealizer) findMatchingTemplateName(resource OwnerResource, supplyChainName string) (string, error) {
	bestMatchingTemplateOptionsIndices, err := selector.BestSelectorMatchIndices(r.owner, v1alpha1.TemplateOptionSelectors(resource.TemplateOptions))

	if err != nil {
		return "", errors.ResolveTemplateOptionError{
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

		return "", errors.TemplateOptionsMatchError{
			ResourceName:  resource.Name,
			OptionNames:   optionNames,
			BlueprintName: supplyChainName,
			BlueprintType: errors.SupplyChain,
		}
	}

	return resource.TemplateOptions[bestMatchingTemplateOptionsIndices[0]].Name, nil
}
