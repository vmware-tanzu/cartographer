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

package registrar

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/cache"
	"sigs.k8s.io/cluster-api/controllers/external"
	"sigs.k8s.io/controller-runtime/pkg/client"
	pkgcontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	"github.com/vmware-tanzu/cartographer/pkg/controller/deliverable"
	"github.com/vmware-tanzu/cartographer/pkg/controller/delivery"
	"github.com/vmware-tanzu/cartographer/pkg/controller/pipeline"
	"github.com/vmware-tanzu/cartographer/pkg/controller/supplychain"
	"github.com/vmware-tanzu/cartographer/pkg/controller/workload"
	realizerdeliverable "github.com/vmware-tanzu/cartographer/pkg/realizer/deliverable"
	realizerpipeline "github.com/vmware-tanzu/cartographer/pkg/realizer/pipeline"
	realizerworkload "github.com/vmware-tanzu/cartographer/pkg/realizer/workload"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
)

type Timer struct{}

func (t Timer) Now() metav1.Time {
	return metav1.Now()
}

func AddToScheme(scheme *runtime.Scheme) error {
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("cartographer v1alpha1 add to scheme: %w", err)
	}

	return nil
}

func RegisterControllers(mgr manager.Manager) error {
	if err := registerWorkloadController(mgr); err != nil {
		return fmt.Errorf("register workload controller: %w", err)
	}

	if err := registerSupplyChainController(mgr); err != nil {
		return fmt.Errorf("register supply-chain controller: %w", err)
	}

	if err := registerDeliveryController(mgr); err != nil {
		return fmt.Errorf("register delivery controller: %w", err)
	}

	if err := registerDeliverableController(mgr); err != nil {
		return fmt.Errorf("register deliverable controller: %w", err)
	}

	if err := registerPipelineServiceController(mgr); err != nil {
		return fmt.Errorf("register pipeline-service controller: %w", err)
	}

	return nil
}

func registerWorkloadController(mgr manager.Manager) error {
	repo := repository.NewRepository(mgr.GetClient(), repository.NewCache(cache.NewExpiring()))

	ctrl, err := pkgcontroller.New("workload", mgr, pkgcontroller.Options{
		Reconciler: workload.NewReconciler(repo, conditions.NewConditionManager, realizerworkload.NewRealizer()),
	})
	if err != nil {
		return fmt.Errorf("controller new: %w", err)
	}

	if err := ctrl.Watch(
		&source.Kind{Type: &v1alpha1.Workload{}},
		&handler.EnqueueRequestForObject{},
	); err != nil {
		return fmt.Errorf("watch: %w", err)
	}

	mapper := Mapper{
		Client: mgr.GetClient(),
		Logger: mgr.GetLogger().WithName("workload"),
	}

	if err := ctrl.Watch(
		&source.Kind{Type: &v1alpha1.ClusterSupplyChain{}},
		handler.EnqueueRequestsFromMapFunc(mapper.ClusterSupplyChainToWorkloadRequests),
	); err != nil {
		return fmt.Errorf("watch: %w", err)
	}

	return nil
}

func registerSupplyChainController(mgr manager.Manager) error {
	repo := repository.NewRepository(mgr.GetClient(), repository.NewCache(cache.NewExpiring()))

	ctrl, err := pkgcontroller.New("supply-chain", mgr, pkgcontroller.Options{
		Reconciler: supplychain.NewReconciler(repo, conditions.NewConditionManager),
	})
	if err != nil {
		return fmt.Errorf("controller new: %w", err)
	}

	if err := ctrl.Watch(
		&source.Kind{Type: &v1alpha1.ClusterSupplyChain{}},
		&handler.EnqueueRequestForObject{},
	); err != nil {
		return fmt.Errorf("watch: %w", err)
	}

	return nil
}

func registerDeliveryController(mgr manager.Manager) error {
	repo := repository.NewRepository(mgr.GetClient(), repository.NewCache(cache.NewExpiring()))

	ctrl, err := pkgcontroller.New("delivery", mgr, pkgcontroller.Options{
		Reconciler: delivery.NewReconciler(repo),
	})
	if err != nil {
		return fmt.Errorf("controller new: %w", err)
	}

	if err := ctrl.Watch(
		&source.Kind{Type: &v1alpha1.ClusterDelivery{}},
		&handler.EnqueueRequestForObject{},
	); err != nil {
		return fmt.Errorf("watch: %w", err)
	}

	return nil
}

func registerDeliverableController(mgr manager.Manager) error {
	repo := repository.NewRepository(mgr.GetClient(), repository.NewCache(cache.NewExpiring()))

	ctrl, err := pkgcontroller.New("deliverable", mgr, pkgcontroller.Options{
		Reconciler: deliverable.NewReconciler(repo, conditions.NewConditionManager, realizerdeliverable.NewRealizer()),
	})
	if err != nil {
		return fmt.Errorf("controller new: %w", err)
	}

	if err := ctrl.Watch(
		&source.Kind{Type: &v1alpha1.Deliverable{}},
		&handler.EnqueueRequestForObject{},
	); err != nil {
		return fmt.Errorf("watch: %w", err)
	}

	mapper := Mapper{
		Client: mgr.GetClient(),
		Logger: mgr.GetLogger().WithName("deliverable"),
	}

	if err := ctrl.Watch(
		&source.Kind{Type: &v1alpha1.ClusterDelivery{}},
		handler.EnqueueRequestsFromMapFunc(mapper.ClusterDeliveryToDeliverableRequests),
	); err != nil {
		return fmt.Errorf("watch: %w", err)
	}

	return nil
}

func registerPipelineServiceController(mgr manager.Manager) error {
	repo := repository.NewRepository(mgr.GetClient(), repository.NewCache(cache.NewExpiring()))

	reconciler := pipeline.NewReconciler(repo, realizerpipeline.NewRealizer())
	ctrl, err := pkgcontroller.New("pipeline-service", mgr, pkgcontroller.Options{
		Reconciler: reconciler,
	})
	if err != nil {
		return fmt.Errorf("controller new pipeline-service: %w", err)
	}

	reconciler.AddTracking(&external.ObjectTracker{
		Controller: ctrl,
	})

	if err := ctrl.Watch(
		&source.Kind{Type: &v1alpha1.Pipeline{}},
		&handler.EnqueueRequestForObject{},
	); err != nil {
		return fmt.Errorf("watch [pipeline-service]: %w", err)
	}

	mapper := Mapper{
		Client: mgr.GetClient(),
		Logger: mgr.GetLogger().WithName("pipeline"),
	}

	if err := ctrl.Watch(
		&source.Kind{Type: &v1alpha1.ClusterRunTemplate{}},
		handler.EnqueueRequestsFromMapFunc(mapper.RunTemplateToPipelineRequests),
	); err != nil {
		return fmt.Errorf("watch: %w", err)
	}

	return nil
}

func IndexResources(mgr manager.Manager, ctx context.Context) error {
	fieldIndexer := mgr.GetFieldIndexer()

	if err := indexSupplyChains(ctx, fieldIndexer); err != nil {
		return fmt.Errorf("index supply chain resource: %w", err)
	}

	return nil
}

func indexSupplyChains(ctx context.Context, fieldIndexer client.FieldIndexer) error {
	err := fieldIndexer.IndexField(ctx, &v1alpha1.ClusterSupplyChain{}, "spec.selector", v1alpha1.GetSelectorsFromObject)
	if err != nil {
		return fmt.Errorf("index field supply-chain.selector: %w", err)
	}

	return nil
}
