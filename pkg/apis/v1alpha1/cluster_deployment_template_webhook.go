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

package v1alpha1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-carto-run-v1alpha1-clusterdeploymenttemplate,mutating=false,failurePolicy=fail,sideEffects=none,admissionReviewVersions=v1beta1;v1,groups=carto.run,resources=clusterdeploymenttemplates,verbs=create;update,versions=v1alpha1,name=deployment-template-validator.cartographer.com

var _ webhook.Validator = &ClusterDeploymentTemplate{}

func (c *ClusterDeploymentTemplate) ValidateCreate() (admission.Warnings, error) {
	return nil, c.validate()
}

func (c *ClusterDeploymentTemplate) ValidateUpdate(_ runtime.Object) (admission.Warnings, error) {
	return nil, c.validate()
}

func (c *ClusterDeploymentTemplate) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

func (c *ClusterDeploymentTemplate) validate() error {
	_, err := c.Spec.TemplateSpec.validate()
	if err != nil {
		return err
	}

	if c.bothConditionsSet() || c.neitherConditionSet() {
		return fmt.Errorf("invalid spec: must set exactly one of spec.ObservedMatches and spec.ObservedCompletion")
	}

	return nil
}

func (c *ClusterDeploymentTemplate) bothConditionsSet() bool {
	return c.Spec.ObservedMatches != nil && c.Spec.ObservedCompletion != nil
}

func (c *ClusterDeploymentTemplate) neitherConditionSet() bool {
	return c.Spec.ObservedMatches == nil && c.Spec.ObservedCompletion == nil
}

func (c *ClusterDeploymentTemplate) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(c).
		Complete()
}
