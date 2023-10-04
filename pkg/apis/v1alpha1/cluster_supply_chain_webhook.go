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

// +kubebuilder:webhook:path=/validate-carto-run-v1alpha1-clustersupplychain,mutating=false,failurePolicy=fail,sideEffects=none,admissionReviewVersions=v1beta1;v1,groups=carto.run,resources=clustersupplychains,verbs=create;update,versions=v1alpha1,name=supply-chain-validator.cartographer.com

var _ webhook.Validator = &ClusterSupplyChain{}

func (c *ClusterSupplyChain) ValidateCreate() (admission.Warnings, error) {
	err := c.validateNewState()
	if err != nil {
		return nil, fmt.Errorf("error validating clustersupplychain [%s]: %w", c.Name, err)
	}
	return nil, nil
}

func (c *ClusterSupplyChain) ValidateUpdate(_ runtime.Object) (admission.Warnings, error) {
	err := c.validateNewState()
	if err != nil {
		return nil, fmt.Errorf("error validating clustersupplychain [%s]: %w", c.Name, err)
	}
	return nil, nil
}

func (c *ClusterSupplyChain) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

func (c *ClusterSupplyChain) validateNewState() error {
	names := make(map[string]bool)

	if len(c.Spec.Selector) == 0 && len(c.Spec.SelectorMatchExpressions) == 0 && len(c.Spec.SelectorMatchFields) == 0 {
		return fmt.Errorf("at least one selector, selectorMatchExpression, selectorMatchField must be specified")
	}

	if err := c.validateParams(); err != nil {
		return err
	}

	if err := validateLegacySelector(c.Spec.LegacySelector, ValidWorkloadPaths, ValidWorkloadPrefixes); err != nil {
		return err
	}

	for _, resource := range c.Spec.Resources {
		if _, ok := names[resource.Name]; ok {
			return fmt.Errorf("duplicate resource name [%s] found", resource.Name)
		}
		names[resource.Name] = true
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
		if err := validateSupplyChainTemplateRef(resource.TemplateRef); err != nil {
			return fmt.Errorf("error validating resource [%s]: %w", resource.Name, err)
		}
	}

	for _, resource := range c.Spec.Resources {
		for _, option := range resource.TemplateRef.Options {
			if option.PassThrough != "" {
				var found bool
				if resource.TemplateRef.Kind == "ClusterSourceTemplate" {
					found = isPassThroughInputFound(resource.Sources, option.PassThrough)
				} else if resource.TemplateRef.Kind == "ClusterImageTemplate" {
					found = isPassThroughInputFound(resource.Images, option.PassThrough)
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

	for _, resource := range c.Spec.Resources {
		if err := c.validateResourceRefs(resource.Sources, "ClusterSourceTemplate"); err != nil {
			return fmt.Errorf(
				"invalid sources for resource [%s]: %w",
				resource.Name,
				err,
			)
		}

		if err := c.validateResourceRefs(resource.Images, "ClusterImageTemplate"); err != nil {
			return fmt.Errorf(
				"invalid images for resource [%s]: %w",
				resource.Name,
				err,
			)
		}

		if err := c.validateResourceRefs(resource.Configs, "ClusterConfigTemplate"); err != nil {
			return fmt.Errorf(
				"invalid configs for resource [%s]: %w",
				resource.Name,
				err,
			)
		}
	}

	return nil
}

func isPassThroughInputFound(refs []ResourceReference, passThrough string) bool {
	for _, ref := range refs {
		if ref.Name == passThrough {
			return true
		}
	}
	return false
}

func (c *ClusterSupplyChain) validateParams() error {
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

func (c *ClusterSupplyChain) validateResourceRefs(references []ResourceReference, targetKind string) error {
	for _, ref := range references {
		referencedResource := c.getResourceByName(ref.Resource)
		if referencedResource == nil {
			return fmt.Errorf(
				"[%s] is provided by unknown resource [%s]",
				ref.Name,
				ref.Resource,
			)
		}
		if referencedResource.TemplateRef.Kind != targetKind {
			return fmt.Errorf(
				"resource [%s] providing [%s] must reference a %s",
				referencedResource.Name,
				ref.Name,
				targetKind,
			)
		}
	}
	return nil
}

func validateSupplyChainTemplateRef(ref SupplyChainTemplateReference) error {
	if ref.Name != "" && len(ref.Options) > 0 {
		return fmt.Errorf("exactly one of templateRef.Name or templateRef.Options must be specified, found both")
	}

	if ref.Name == "" && len(ref.Options) < 2 {
		if len(ref.Options) == 1 {
			return fmt.Errorf("templateRef.Options must have more than one option")
		}
		return fmt.Errorf("exactly one of templateRef.Name or templateRef.Options must be specified, found neither")
	}

	if err := validateResourceOptions(ref.Options, ValidWorkloadPaths, ValidWorkloadPrefixes); err != nil {
		return err
	}
	return nil
}

func (c *ClusterSupplyChain) getResourceByName(name string) *SupplyChainResource {
	for _, resource := range c.Spec.Resources {
		if resource.Name == name {
			return &resource
		}
	}

	return nil
}

func (c *ClusterSupplyChain) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(c).
		Complete()
}
