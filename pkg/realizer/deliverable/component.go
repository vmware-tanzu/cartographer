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

////go:generate go run -modfile ../../../hack/tools/go.mod github.com/maxbrunsfeld/counterfeiter/v6 -generate

// //counterfeiter:generate . ResourceRealizer
//type ResourceRealizer interface {
//	Do(ctx context.Context, resource *v1alpha1.DeliveryResource, deliveryName string, outputs Outputs) (templates.Template, *unstructured.Unstructured, *templates.Output, error)
//}
//
//type resourceRealizer struct {
//	deliverable     *v1alpha1.Deliverable
//	systemRepo      repository.Repository
//	deliverableRepo repository.Repository
//	deliveryParams  []v1alpha1.BlueprintParam
//}
//
//type ResourceRealizerBuilder func(secret *corev1.Secret, deliverable *v1alpha1.Deliverable, repo repository.Repository, deliveryParams []v1alpha1.BlueprintParam) (ResourceRealizer, error)
//
//func NewResourceRealizerBuilder(repositoryBuilder repository.RepositoryBuilder, clientBuilder realizerclient.ClientBuilder, cache repository.RepoCache) ResourceRealizerBuilder {
//	return func(secret *corev1.Secret, deliverable *v1alpha1.Deliverable, systemRepo repository.Repository, deliveryParams []v1alpha1.BlueprintParam) (ResourceRealizer, error) {
//		client, _, err := clientBuilder(secret, false)
//		if err != nil {
//			return nil, fmt.Errorf("can't build client: %w", err)
//		}
//		deliverableRepo := repositoryBuilder(client, cache)
//		return &resourceRealizer{
//			deliverable:     deliverable,
//			systemRepo:      systemRepo,
//			deliverableRepo: deliverableRepo,
//			deliveryParams:  deliveryParams,
//		}, nil
//	}
//}
//
//func (r *resourceRealizer) Do(ctx context.Context, resource *v1alpha1.DeliveryResource, deliveryName string, outputs Outputs) (templates.Template, *unstructured.Unstructured, *templates.Output, error) {
//	log := logr.FromContextOrDiscard(ctx).WithValues("template", resource.TemplateRef)
//	ctx = logr.NewContext(ctx, log)
//
//	var templateName string
//	var err error
//	if len(resource.TemplateRef.Options) > 0 {
//		templateName, err = r.findMatchingTemplateName(resource, deliveryName)
//		if err != nil {
//			return nil, nil, nil, err
//		}
//	} else {
//		templateName = resource.TemplateRef.Name
//	}
//
//	log.V(logger.DEBUG).Info("realizing template", "template", fmt.Sprintf("[%s/%s]", resource.TemplateRef.Kind, templateName))
//
//	apiTemplate, err := r.systemRepo.GetTemplate(ctx, templateName, resource.TemplateRef.Kind)
//	if err != nil {
//		log.Error(err, "failed to get delivery cluster template")
//		return nil, nil, nil, errors.GetTemplateError{
//			Err:           err,
//			ResourceName:  resource.Name,
//			TemplateName:  templateName,
//			BlueprintName: deliveryName,
//			BlueprintType: errors.Delivery,
//		}
//	}
//
//	template, err := templates.NewModelFromAPI(apiTemplate)
//	if err != nil {
//		log.Error(err, "failed to get delivery cluster template")
//		return nil, nil, nil, fmt.Errorf("failed to get delivery cluster template [%+v]: %w", resource.TemplateRef, err)
//	}
//
//	labels := map[string]string{
//		"carto.run/deliverable-name":      r.deliverable.Name,
//		"carto.run/deliverable-namespace": r.deliverable.Namespace,
//		"carto.run/delivery-name":         deliveryName,
//		"carto.run/resource-name":         resource.Name,
//		"carto.run/template-kind":         template.GetKind(),
//		"carto.run/cluster-template-name": template.GetName(),
//	}
//
//	inputs := outputs.GenerateInputs(resource)
//	templatingContext := map[string]interface{}{
//		"deliverable": r.deliverable,
//		"params":      templates.ParamsBuilder(template.GetDefaultParams(), r.deliveryParams, resource.Params, r.deliverable.Spec.Params),
//		"sources":     inputs.Sources,
//		"configs":     inputs.Configs,
//		"deployment":  inputs.Deployment,
//	}
//
//	// Todo: this belongs in Stamp.
//	if inputs.OnlyConfig() != nil {
//		templatingContext["config"] = inputs.OnlyConfig()
//	}
//	if inputs.OnlySource() != nil {
//		templatingContext["source"] = inputs.OnlySource()
//	}
//
//	stampContext := templates.StamperBuilder(r.deliverable, templatingContext, labels)
//	stampedObject, err := stampContext.Stamp(ctx, template.GetResourceTemplate())
//	if err != nil {
//		log.Error(err, "failed to stamp resource")
//		return template, nil, nil, errors.StampError{
//			Err:           err,
//			ResourceName:  resource.Name,
//			BlueprintName: deliveryName,
//			BlueprintType: errors.Delivery,
//		}
//	}
//
//	err = r.deliverableRepo.EnsureMutableObjectExistsOnCluster(ctx, stampedObject)
//	if err != nil {
//		log.Error(err, "failed to ensure object exists on cluster", "object", stampedObject)
//		return template, nil, nil, errors.ApplyStampedObjectError{
//			Err:           err,
//			StampedObject: stampedObject,
//			ResourceName:  resource.Name,
//			BlueprintName: deliveryName,
//			BlueprintType: errors.Delivery,
//		}
//	}
//
//	template.SetInputs(inputs)
//	template.SetStampedObject(stampedObject)
//
//	output, err := template.GetOutput()
//	if err != nil {
//		log.Error(err, "failed to retrieve output from object", "object", stampedObject)
//		return template, stampedObject, nil, errors.RetrieveOutputError{
//			Err:           err,
//			ResourceName:  resource.Name,
//			StampedObject: stampedObject,
//			BlueprintName: deliveryName,
//			BlueprintType: errors.Delivery,
//		}
//	}
//
//	return template, stampedObject, output, nil
//}
//
//func (r *resourceRealizer) findMatchingTemplateName(resource *v1alpha1.DeliveryResource, deliveryName string) (string, error) {
//	bestMatchingTemplateOptionsIndices, err := selector.BestSelectorMatchIndices(r.deliverable, v1alpha1.TemplateOptionSelectors(resource.TemplateRef.Options))
//
//	if err != nil {
//		return "", errors.ResolveTemplateOptionError{
//			Err:           err,
//			ResourceName:  resource.Name,
//			OptionName:    resource.TemplateRef.Options[err.SelectorIndex()].Name,
//			BlueprintName: deliveryName,
//			BlueprintType: errors.Delivery,
//		}
//	}
//
//	if len(bestMatchingTemplateOptionsIndices) != 1 {
//		var optionNames []string
//		for _, optionIndex := range bestMatchingTemplateOptionsIndices {
//			optionNames = append(optionNames, resource.TemplateRef.Options[optionIndex].Name)
//		}
//
//		return "", errors.TemplateOptionsMatchError{
//			ResourceName:  resource.Name,
//			OptionNames:   optionNames,
//			BlueprintName: deliveryName,
//			BlueprintType: errors.Delivery,
//		}
//	}
//
//	return resource.TemplateRef.Options[bestMatchingTemplateOptionsIndices[0]].Name, nil
//}
