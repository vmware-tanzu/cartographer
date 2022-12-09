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

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/realizer"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/statuses"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

//go:generate go run -modfile ../../hack/tools/go.mod github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Realizer
type Realizer interface {
	Realize(ctx context.Context, resourceRealizer realizer.ResourceRealizer, blueprintName string, ownerResources []realizer.OwnerResource, resourceStatuses statuses.ResourceStatuses) error
}

func mutableEquivalenceTest(realizedResource v1alpha1.ResourceStatus, prevResource v1alpha1.ResourceStatus, _ context.Context, _ repository.Repository) (bool, error) {
	return realizedResource.StampedRef.GroupVersionKind() == prevResource.StampedRef.GroupVersionKind() &&
		realizedResource.StampedRef.Namespace == prevResource.StampedRef.Namespace &&
		realizedResource.StampedRef.Name == prevResource.StampedRef.Name, nil
}

func immutableEquivalenceTest(realizedResource v1alpha1.ResourceStatus, prevResource v1alpha1.ResourceStatus, ctx context.Context, repo repository.Repository) (bool, error) {
	obj := &unstructured.Unstructured{}

	obj.SetNamespace(prevResource.StampedRef.ObjectReference.Namespace)
	obj.SetName(prevResource.StampedRef.ObjectReference.Name)
	obj.SetGroupVersionKind(prevResource.StampedRef.ObjectReference.GroupVersionKind())

	prevObj, err := repo.GetUnstructured(ctx, obj)
	if err != nil {
		return false, fmt.Errorf("get unstructured: %w", err)
	}

	prevLabels := prevObj.GetLabels()

	if lifecycleLabel, ok := prevLabels["carto.run/template-lifecycle"]; !ok || lifecycleLabel == "mutable" {
		return mutableEquivalenceTest(realizedResource, prevResource, ctx, repo)
	}

	return realizedResource.TemplateRef.Name == prevResource.TemplateRef.Name &&
		realizedResource.TemplateRef.Kind == prevResource.TemplateRef.Kind &&
		realizedResource.Name == prevResource.Name, nil
}

func getEquivalenceTest(ctx context.Context, repo repository.Repository, prevResource v1alpha1.ResourceStatus) (
	func(v1alpha1.ResourceStatus, v1alpha1.ResourceStatus, context.Context, repository.Repository) (bool, error),
	error,
) {
	log := logr.FromContextOrDiscard(ctx)

	apiTemplate, err := repo.GetTemplate(ctx, prevResource.TemplateRef.Name, prevResource.TemplateRef.Kind)
	if err != nil {
		log.Error(err, "unable to get api template")
		return nil, fmt.Errorf("unable to get api template [%s/%s]: %w", prevResource.TemplateRef.Kind, prevResource.TemplateRef.Name, err)
	}
	reader, err := templates.NewReaderFromAPI(apiTemplate)
	if err != nil {
		log.Error(err, "failed to get reader for apiTemplate")
		return nil, fmt.Errorf("failed to get reader for apiTemplate [%s/%s]: %w", prevResource.TemplateRef.Kind, prevResource.TemplateRef.Name, err)
	}

	if reader.GetLifecycle().IsImmutable() {
		return immutableEquivalenceTest, nil
	} else {
		return mutableEquivalenceTest, nil
	}
}
