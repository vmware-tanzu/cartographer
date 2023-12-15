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

// +kubebuilder:webhook:path=/validate-carto-run-v1alpha1-clusterdelivery,mutating=false,failurePolicy=fail,sideEffects=none,admissionReviewVersions=v1beta1;v1,groups=carto.run,resources=clusterdeliveries,verbs=create;update,versions=v1alpha1,name=delivery-validator.cartographer.com

var _ webhook.Validator = &ClusterDelivery{}

func (c *ClusterDelivery) ValidateCreate() (admission.Warnings, error) {
	err := c.validateNewState()
	if err != nil {
		return nil, fmt.Errorf("error validating clusterdelivery [%s]: %w", c.Name, err)
	}
	return nil, nil
}

func (c *ClusterDelivery) ValidateUpdate(_ runtime.Object) (admission.Warnings, error) {
	err := c.validateNewState()
	if err != nil {
		return nil, fmt.Errorf("error validating clusterdelivery [%s]: %w", c.Name, err)
	}
	return nil, nil
}

func (c *ClusterDelivery) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

func (c *ClusterDelivery) validateNewState() error {
	if len(c.Spec.Selector) == 0 && len(c.Spec.SelectorMatchExpressions) == 0 && len(c.Spec.SelectorMatchFields) == 0 {
		return fmt.Errorf("at least one selector, selectorMatchExpression, selectorMatchField must be specified")
	}

	if err := c.validateParams(); err != nil {
		return err
	}

	if err := validateLegacySelector(c.Spec.LegacySelector, ValidDeliverablePaths, ValidDeliverablePrefixes); err != nil {
		return err
	}

	if err := c.validateResourceNamesUnique(); err != nil {
		return err
	}

	for _, resource := range c.Spec.Resources {
		optionNames := make(map[string]bool)
		for _, option := range resource.TemplateRef.Options {
			if _, ok := optionNames[option.Name]; ok {
				return fmt.Errorf(
					"duplicate template name [%s] found in options for resource [%s]",
					option.Name,
					resource.Name,
				)
			}
			optionNames[option.Name] = true
		}
	}

	for _, resource := range c.Spec.Resources {
		if err := validateDeliveryTemplateRef(resource.TemplateRef); err != nil {
			return fmt.Errorf("error validating resource [%s]: %w", resource.Name, err)
		}
	}

	for _, resource := range c.Spec.Resources {
		for _, option := range resource.TemplateRef.Options {
			if option.PassThrough != "" {
				var found bool
				if resource.TemplateRef.Kind == "ClusterSourceTemplate" {
					found = isPassThroughInputFound(resource.Sources, option.PassThrough)
				} else if resource.TemplateRef.Kind == "ClusterConfigTemplate" {
					found = isPassThroughInputFound(resource.Configs, option.PassThrough)
				} else {
					return fmt.Errorf("error validating resource [%s]: TemplateRef.Kind [%s] is not a known type", resource.Name, resource.TemplateRef.Kind)
				}

				if !found {
					return fmt.Errorf("error validating resource [%s]: pass through [%s] does not refer to a known input", resource.Name, option.PassThrough)
				}
			}
		}
	}

	if err := c.validateDeploymentPassedToProperReceivers(); err != nil {
		return err
	}

	return c.validateDeploymentTemplateDidNotReceiveConfig()
}

func (c *ClusterDelivery) validateDeploymentPassedToProperReceivers() error {
	for _, resource := range c.Spec.Resources {
		if resource.TemplateRef.Kind == "ClusterDeploymentTemplate" && resource.Deployment == nil {
			return fmt.Errorf("spec.resources[%s] is a ClusterDeploymentTemplate and must receive a deployment", resource.Name)
		}

		if resource.Deployment != nil && resource.TemplateRef.Kind != "ClusterDeploymentTemplate" {
			return fmt.Errorf("spec.resources[%s] receives a deployment but is not a ClusterDeploymentTemplate", resource.Name)
		}
	}
	return nil
}

func (c *ClusterDelivery) validateResourceNamesUnique() error {
	names := map[string]bool{}

	for idx, resource := range c.Spec.Resources {
		if names[resource.Name] {
			return fmt.Errorf("spec.resources[%d].name \"%s\" cannot appear twice", idx, resource.Name)
		}
		names[resource.Name] = true
	}
	return nil
}

func (c *ClusterDelivery) validateDeploymentTemplateDidNotReceiveConfig() error {
	for _, resource := range c.Spec.Resources {
		if resource.TemplateRef.Kind == "ClusterDeploymentTemplate" && resource.Configs != nil {
			return fmt.Errorf("spec.resources[%s] is a ClusterDeploymentTemplate and must not receive config", resource.Name)
		}
	}
	return nil
}

func (c *ClusterDelivery) validateParams() error {
	for _, param := range c.Spec.Params {
		err := param.validate()
		if err != nil {
			return err
		}
	}

	for _, resource := range c.Spec.Resources {
		for _, param := range resource.Params {
			err := param.validate()
			if err != nil {
				return fmt.Errorf("resource [%s] is invalid: %w", resource.Name, err)
			}
		}
	}

	return nil
}

func validateDeliveryTemplateRef(ref DeliveryTemplateReference) error {
	if ref.Name != "" && len(ref.Options) > 0 {
		return fmt.Errorf("exactly one of templateRef.Name or templateRef.Options must be specified, found both")
	}

	if ref.Name == "" && len(ref.Options) < 2 {
		if len(ref.Options) == 1 {
			return fmt.Errorf("templateRef.Options must have more than one option")
		}
		return fmt.Errorf("exactly one of templateRef.Name or templateRef.Options must be specified, found neither")
	}

	if err := validateResourceOptions(ref.Options, ValidDeliverablePaths, ValidDeliverablePrefixes); err != nil {
		return err
	}
	return nil
}

func (c *ClusterDelivery) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(c).
		Complete()
}
