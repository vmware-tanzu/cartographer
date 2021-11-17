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

package root

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/registrar"
)

type Command struct {
	Port    int
	CertDir string
	Context context.Context
	Logger  logr.Logger
}

func (cmd *Command) Execute() error {
	log.SetLogger(cmd.Logger)
	l := log.Log.WithName("cartographer")

	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	scheme := runtime.NewScheme()
	if err := registrar.AddToScheme(scheme); err != nil {
		return fmt.Errorf("add to scheme: %w", err)
	}

	mgr, err := manager.New(cfg, manager.Options{
		Port:               cmd.Port,
		CertDir:            cmd.CertDir,
		Scheme:             scheme,
		MetricsBindAddress: "0",
	})

	if err != nil {
		return fmt.Errorf("manager new: %w", err)
	}

	if err := registrar.RegisterControllers(mgr); err != nil {
		return fmt.Errorf("register controllers: %w", err)
	}

	if err := registrar.IndexResources(mgr, cmd.Context); err != nil {
		return fmt.Errorf("index resources: %w", err)
	}

	if cmd.CertDir == "" {
		l.Info("Not registering the webhook server. Must pass a directory containing tls.crt and tls.key to --cert-dir")
	} else {
		if err := controllerruntime.NewWebhookManagedBy(mgr).
			For(&v1alpha1.ClusterSupplyChain{}).
			Complete(); err != nil {
			return fmt.Errorf("clustersupplychain webhook: %w", err)
		}
		if err := controllerruntime.NewWebhookManagedBy(mgr).
			For(&v1alpha1.ClusterConfigTemplate{}).
			Complete(); err != nil {
			return fmt.Errorf("clusterconfigtemplate webhook: %w", err)
		}
		if err := controllerruntime.NewWebhookManagedBy(mgr).
			For(&v1alpha1.ClusterImageTemplate{}).
			Complete(); err != nil {
			return fmt.Errorf("clusterimagetemplate webhook: %w", err)
		}
		if err := controllerruntime.NewWebhookManagedBy(mgr).
			For(&v1alpha1.ClusterSourceTemplate{}).
			Complete(); err != nil {
			return fmt.Errorf("clustersourcetemplate webhook: %w", err)
		}
		if err := controllerruntime.NewWebhookManagedBy(mgr).
			For(&v1alpha1.ClusterTemplate{}).
			Complete(); err != nil {
			return fmt.Errorf("clustertemplate webhook: %w", err)
		}
		if err := controllerruntime.NewWebhookManagedBy(mgr).
			For(&v1alpha1.ClusterDelivery{}).
			Complete(); err != nil {
			return fmt.Errorf("clusterdelivery webhook: %w", err)
		}
		if err := controllerruntime.NewWebhookManagedBy(mgr).
			For(&v1alpha1.ClusterDeploymentTemplate{}).
			Complete(); err != nil {
			return fmt.Errorf("clusterdeploymenttemplate webhook: %w", err)
		}
	}

	if err := mgr.Start(cmd.Context); err != nil {
		return fmt.Errorf("manager start: %w", err)
	}

	return nil
}
