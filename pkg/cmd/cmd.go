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

package cmd

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/registrar"
)

type Command struct {
	Port    int
	CertDir string
	Logger  logr.Logger
}

func (cmd *Command) Execute(ctx context.Context) error {
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

	if err := registrar.IndexResources(ctx, mgr); err != nil {
		return fmt.Errorf("index resources: %w", err)
	}

	if cmd.CertDir != "" {
		if err := registerWebhooks(mgr); err != nil {
			return fmt.Errorf("failed to register webhooks: %w", err)
		}
	} else {
		l.Info("Not registering the webhook server. Must pass a directory containing tls.crt and tls.key to --cert-dir")
	}

	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("manager start: %w", err)
	}

	return nil
}

func registerWebhooks(mgr manager.Manager) error {
	if err := (&v1alpha1.ClusterSupplyChain{}).SetupWebhookWithManager(mgr); err != nil {
		return fmt.Errorf("failed to setup cluster supply chain webhook: %w", err)
	}

	if err := (&v1alpha1.ClusterDelivery{}).SetupWebhookWithManager(mgr); err != nil {
		return fmt.Errorf("failed to setup cluster deliver webhook: %w", err)
	}

	if err := (&v1alpha1.ClusterConfigTemplate{}).SetupWebhookWithManager(mgr); err != nil {
		return fmt.Errorf("failed to setup cluster config template webhook: %w", err)
	}

	if err := (&v1alpha1.ClusterDeploymentTemplate{}).SetupWebhookWithManager(mgr); err != nil {
		return fmt.Errorf("failed to setup cluster deployment template webhook: %w", err)
	}

	if err := (&v1alpha1.ClusterImageTemplate{}).SetupWebhookWithManager(mgr); err != nil {
		return fmt.Errorf("failed to setup cluster image template webhook: %w", err)
	}

	if err := (&v1alpha1.ClusterRunTemplate{}).SetupWebhookWithManager(mgr); err != nil {
		return fmt.Errorf("failed to setup cluster run template webhook: %w", err)
	}

	if err := (&v1alpha1.ClusterSourceTemplate{}).SetupWebhookWithManager(mgr); err != nil {
		return fmt.Errorf("failed to setup cluster source template webhook: %w", err)
	}

	if err := (&v1alpha1.ClusterTemplate{}).SetupWebhookWithManager(mgr); err != nil {
		return fmt.Errorf("failed to setup cluster template webhook: %w", err)
	}

	return nil
}
