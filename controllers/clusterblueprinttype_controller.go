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
	blueprintsv1alpha1 "carto.run/blueprints/api/v1alpha1"
	"github.com/vmware-labs/reconciler-runtime/reconcilers"
)

// ClusterBlueprintTypeReconciler reconciles ClusterBlueprintType
//+kubebuilder:rbac:groups=blueprints.carto.run,resources=clusterblueprinttypes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=blueprints.carto.run,resources=clusterblueprinttypes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=blueprints.carto.run,resources=clusterblueprinttypes/finalizers,verbs=update
func ClusterBlueprintTypeReconciler(c reconcilers.Config) *reconcilers.ResourceReconciler {
	return &reconcilers.ResourceReconciler{
		Name:       "Function",
		Type:       &blueprintsv1alpha1.ClusterBlueprintType{},
		Reconciler: reconcilers.Sequence{},
		Config:     c,
	}
}
