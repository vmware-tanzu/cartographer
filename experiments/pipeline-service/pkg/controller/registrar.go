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

package controller

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/vmware-tanzu/cartographer/experiments/pipeline-service/pkg/apis/v1alpha1"
)

func AddToScheme(scheme *runtime.Scheme) error {
	if err := corev1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("corev1 add to scheme: %w", err)
	}

	if err := v1alpha1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("ppservice v1alpha1 add to scheme: %w", err)
	}

	return nil
}

func RegisterControllers(mgr manager.Manager) error {
	if err := RegisterPipelineController(mgr); err != nil {
		return fmt.Errorf("register pipeline controller: %w", err)
	}

	return nil
}

func RegisterPipelineController(mgr manager.Manager) error {
	ctrl, err := controller.New("pipeline", mgr, controller.Options{
		Reconciler: &PipelineReconciler{
			Client: mgr.GetClient(),
			Logger: mgr.GetLogger().WithName("pipeline"),
		},
	})
	if err != nil {
		return fmt.Errorf("controller new: %w", err)
	}

	if err := ctrl.Watch(
		&source.Kind{Type: &v1alpha1.Pipeline{}},
		&handler.EnqueueRequestForObject{},
	); err != nil {
		return fmt.Errorf("watch: %w", err)
	}

	return nil
}
