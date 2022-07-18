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
)

// +kubebuilder:webhook:path=/validate-carto-run-v1alpha1-runnable,mutating=false,failurePolicy=fail,sideEffects=none,admissionReviewVersions=v1beta1;v1,groups=carto.run,resources=runnables,verbs=create;update,versions=v1alpha1,name=runnable-validator.cartographer.com

var _ webhook.Validator = &Runnable{}

func (r *Runnable) ValidateCreate() error {
	err := validateName(r.ObjectMeta)
	if err != nil {
		return fmt.Errorf("error validating runnable [%s]: %w", r.Name, err)
	}
	return nil
}

func (r *Runnable) ValidateUpdate(_ runtime.Object) error {
	err := validateName(r.ObjectMeta)
	if err != nil {
		return fmt.Errorf("error validating runnable [%s]: %w", r.Name, err)
	}
	return nil
}

func (r *Runnable) ValidateDelete() error {
	return nil
}

func (r *Runnable) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}
