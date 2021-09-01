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

package main

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/vmware-tanzu/cartographer/experiments/pipeline-service/pkg/controller"
)

func init() {
	log.SetLogger(zap.New(zap.UseDevMode(true)))
}

func run(ctx context.Context) error {
	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	scheme := runtime.NewScheme()
	if err := controller.AddToScheme(scheme); err != nil {
		return fmt.Errorf("add to scheme: %w", err)
	}

	mgr, err := manager.New(cfg, manager.Options{
		Scheme:             scheme,
		MetricsBindAddress: "0",
	})
	if err != nil {
		return fmt.Errorf("manager new: %w", err)
	}

	if err := controller.RegisterControllers(mgr); err != nil {
		return fmt.Errorf("register controllers: %w", err)
	}

	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("manager start: %w", err)
	}

	return nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := run(ctx); err != nil {
		panic(err)
	}
}
