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

import "fmt"

func (c *ClusterDelivery) validateNewState() error {
	if len(c.Spec.Selector) == 0 && len(c.Spec.SelectorMatchExpressions) == 0 && len(c.Spec.SelectorMatchFields) == 0 {
		return fmt.Errorf("at least one selector, selectorMatchExpression, selectorMatchField must be specified")
	}

	if err := c.validateParams(); err != nil {
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
