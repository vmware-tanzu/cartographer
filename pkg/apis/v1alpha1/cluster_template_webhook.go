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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-carto-run-v1alpha1-clustertemplate,mutating=false,failurePolicy=fail,sideEffects=none,admissionReviewVersions=v1beta1;v1,groups=carto.run,resources=clustertemplates,verbs=create;update,versions=v1alpha1,name=template-validator.cartographer.com


var _ webhook.Validator = &ClusterTemplate{}

func (c *ClusterTemplate) ValidateCreate() (admission.Warnings, error) {
	return c.Spec.validate()
	
}

func (c *ClusterTemplate) ValidateUpdate(_ runtime.Object) (admission.Warnings, error) {
	return c.Spec.validate()
}

func (c *ClusterTemplate) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

func (c *ClusterTemplate) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(c).
		Complete()
}
