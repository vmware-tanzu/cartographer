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
	"net/http"
	"net/http/pprof"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/controllers"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

type Command struct {
	Port                    int
	CertDir                 string
	MetricsPort             int
	PprofPort               int
	Logger                  logr.Logger
	MaxConcurrentDeliveries int
	MaxConcurrentWorkloads  int
	MaxConcurrentRunnables  int
}

func (cmd *Command) Execute(ctx context.Context) error {
	log.SetLogger(cmd.Logger)
	klog.SetLogger(cmd.Logger)
	l := log.Log.WithName("cartographer")

	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	scheme := runtime.NewScheme()
	if err := utils.AddToScheme(scheme); err != nil {
		return fmt.Errorf("add to scheme: %w", err)
	}

	mgrOpts := manager.Options{
		WebhookServer: webhook.NewServer(webhook.Options{
			Port:    cmd.Port,
			CertDir: cmd.CertDir,
		}),
		Scheme: scheme,
		Metrics: server.Options{
			BindAddress: "0",
		},
	}

	if cmd.MetricsPort != 0 {
		mgrOpts.Metrics.BindAddress = fmt.Sprintf(":%d", cmd.MetricsPort)
	}

	if cmd.PprofPort != 0 {
		startPprof(cmd.PprofPort)
	}

	mgr, err := manager.New(cfg, mgrOpts)
	if err != nil {
		return fmt.Errorf("failed to create new manager: %w", err)
	}

	if err := cmd.registerControllers(mgr); err != nil {
		return fmt.Errorf("failed to register controllers: %w", err)
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

func (cmd *Command) registerControllers(mgr manager.Manager) error {
	if err := (&controllers.WorkloadReconciler{}).SetupWithManager(mgr, cmd.MaxConcurrentWorkloads); err != nil {
		return fmt.Errorf("failed to register workload controller: %w", err)
	}

	if err := (&controllers.SupplyChainReconciler{}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("failed to register supply chain controller: %w", err)
	}

	if err := (&controllers.DeliverableReconciler{}).SetupWithManager(mgr, cmd.MaxConcurrentDeliveries); err != nil {
		return fmt.Errorf("failed to register deliverable controller: %w", err)
	}

	if err := (&controllers.DeliveryReconiler{}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("failed to register delivery controller: %w", err)
	}

	if err := (&controllers.RunnableReconciler{}).SetupWithManager(mgr, cmd.MaxConcurrentRunnables); err != nil {
		return fmt.Errorf("failed to register runnable controller: %w", err)
	}

	return nil
}

func registerWebhooks(mgr manager.Manager) error {
	if err := (&v1alpha1.ClusterSupplyChain{}).SetupWebhookWithManager(mgr); err != nil {
		return fmt.Errorf("failed to setup cluster supply chain webhook: %w", err)
	}

	if err := (&v1alpha1.ClusterDelivery{}).SetupWebhookWithManager(mgr); err != nil {
		return fmt.Errorf("failed to setup cluster delivery webhook: %w", err)
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

func startPprof(port int) {
	mux := &http.ServeMux{}
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	go func() {
		_ = http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
	}()
}
