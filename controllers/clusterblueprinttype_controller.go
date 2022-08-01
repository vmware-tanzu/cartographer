/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	blueprintsv1alpha1 "carto.run/blueprints/api/v1alpha1"
	"github.com/vmware-labs/reconciler-runtime/reconcilers"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
)

// ClusterBlueprintTypeReconciler reconciles ClusterBlueprintType
//+kubebuilder:rbac:groups=blueprints.carto.run,resources=clusterblueprinttypes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=blueprints.carto.run,resources=clusterblueprinttypes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=blueprints.carto.run,resources=clusterblueprinttypes/finalizers,verbs=update
func ClusterBlueprintTypeReconciler(c reconcilers.Config) *reconcilers.ResourceReconciler {
	return &reconcilers.ResourceReconciler{
		Name: "Function",
		Type: &blueprintsv1alpha1.ClusterBlueprintType{},
		Reconciler: reconcilers.Sequence{
			ValidateSchema(),
		},
		Config: c,
	}
}

// ValidateSchema ensures the spec.Schema is valid OpenAPI v3 schema
//+kubebuilder:rbac:groups=blueprints.carto.run,resources=clusterblueprinttypes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=blueprints.carto.run,resources=clusterblueprinttypes/status,verbs=get;update;patch
func ValidateSchema() reconcilers.SubReconciler {
	return &reconcilers.SyncReconciler{
		Name: "ValidateSchema",
		Sync: func(ctx context.Context, parent *blueprintsv1alpha1.ClusterBlueprintType) error {
			//log := logr.FromContextOrDiscard(ctx)
			//c := reconcilers.RetrieveConfigOrDie(ctx)

			schema := parent.Spec.Schema

			apiextensions.JSONSchemaProps{}
			//sources := &conventionsv1alpha1.ClusterPodConventionList{}
			//if err := c.List(ctx, sources); err != nil {
			//	return err
			//}
			//var conventions binding.Conventions
			//conditionManager := parent.GetConditionSet().Manage(&parent.Status)
			//for i := range sources.Items {
			//	source := sources.Items[i].DeepCopy()
			//	source.Default()
			//	convention := binding.Convention{
			//		Name:      source.Name,
			//		Selectors: source.Spec.Selectors,
			//		Priority:  source.Spec.Priority,
			//	}
			//	if source.Spec.Webhook != nil {
			//		clientConfig := source.Spec.Webhook.ClientConfig.DeepCopy()
			//		if source.Spec.Webhook.Certificate != nil {
			//			caBundle, err := getCABundle(ctx, c, source.Spec.Webhook.Certificate, parent)
			//			if err != nil {
			//				conditionManager.MarkFalse(conventionsv1alpha1.PodIntentConditionConventionsApplied, "CABundleResolutionFailed", "failed to authenticate: %v", err.Error())
			//				log.Error(err, "failed to get CABundle", "ClusterPodConvention", source.Name)
			//				return nil
			//			}
			//			// inject the CA data
			//			clientConfig.CABundle = caBundle
			//		}
			//		convention.ClientConfig = *clientConfig
			//	}
			//	conventions = append(conventions, convention)
			//}
			//StashConventions(ctx, conventions)
			return nil
		},
	}
}
