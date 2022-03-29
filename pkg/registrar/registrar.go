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

//go:generate go run -modfile ../../hack/tools/go.mod github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	"github.com/vmware-tanzu/cartographer/pkg/controller/runnable"
	"github.com/vmware-tanzu/cartographer/pkg/controller/supplychain"
	"github.com/vmware-tanzu/cartographer/pkg/controller/workload"
	"github.com/vmware-tanzu/cartographer/pkg/enqueuer"
	realizerclient "github.com/vmware-tanzu/cartographer/pkg/realizer/client"
	realizerdeliverable "github.com/vmware-tanzu/cartographer/pkg/realizer/deliverable"
	realizerrunnable "github.com/vmware-tanzu/cartographer/pkg/realizer/runnable"
	realizerworkload "github.com/vmware-tanzu/cartographer/pkg/realizer/workload"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/dependency"
)

type Timer struct{}

const defaultResyncTime = 10 * time.Hour

func (t Timer) Now() metav1.Time {
	return metav1.Now()
}

func AddToScheme(scheme *runtime.Scheme) error {
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("cartographer v1alpha1 add to scheme: %w", err)
	}

	if err := corev1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("core v1 add to scheme: %w", err)
	}

	if err := rbacv1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("rbac v1 add to scheme: %w", err)
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

	if err := registerRunnableController(mgr); err != nil {
		return fmt.Errorf("register runnable-service controller: %w", err)
	}

	return nil
}

func registerWorkloadController(mgr manager.Manager) error {
	repo := repository.NewRepository(
		mgr.GetClient(),
		repository.NewCache(mgr.GetLogger().WithName("workload-repo-cache")),
	)

	recorder := mgr.GetEventRecorderFor("workload-controller")

	reconciler := &workload.Reconciler{
		Repo:                    repo,
		ConditionManagerBuilder: conditions.NewConditionManager,
		ResourceRealizerBuilder: realizerworkload.NewResourceRealizerBuilder(repository.NewRepository, realizerclient.NewClientBuilder(mgr.GetConfig()), repository.NewCache(mgr.GetLogger().WithName("workload-stamping-repo-cache"))),
		Realizer:                realizerworkload.NewRealizer(recorder),
		DependencyTracker:       dependency.NewDependencyTracker(2*defaultResyncTime, mgr.GetLogger().WithName("tracker-workload")),
		EventRecorder:           recorder,
	}

	ctrl, err := pkgcontroller.New("workload", mgr, pkgcontroller.Options{
		Reconciler: reconciler,
	})
	if err != nil {
		return fmt.Errorf("controller new: %w", err)
	}

	reconciler.StampedTracker = &external.ObjectTracker{Controller: ctrl}

	if err := ctrl.Watch(
		&source.Kind{Type: &v1alpha1.Workload{}},
		&handler.EnqueueRequestForObject{},
	); err != nil {
		return fmt.Errorf("watch: %w", err)
	}

	mapper := Mapper{
		Client:  mgr.GetClient(),
		Logger:  mgr.GetLogger().WithName("workload"),
		Tracker: reconciler.DependencyTracker,
	}

	watches := map[client.Object]handler.MapFunc{
		&v1alpha1.ClusterSupplyChain{}: mapper.ClusterSupplyChainToWorkloadRequests,
		&corev1.ServiceAccount{}:       mapper.ServiceAccountToWorkloadRequests,
		&rbacv1.Role{}:                 mapper.RoleToWorkloadRequests,
		&rbacv1.RoleBinding{}:          mapper.RoleBindingToWorkloadRequests,
		&rbacv1.ClusterRole{}:          mapper.ClusterRoleToWorkloadRequests,
		&rbacv1.ClusterRoleBinding{}:   mapper.ClusterRoleBindingToWorkloadRequests,
	}

	for kindType, mapFunc := range watches {
		if err := ctrl.Watch(
			&source.Kind{Type: kindType},
			handler.EnqueueRequestsFromMapFunc(mapFunc),
		); err != nil {
			return fmt.Errorf("watch %T: %w", kindType, err)
		}
	}

	for _, template := range v1alpha1.ValidSupplyChainTemplates {
		if err := ctrl.Watch(
			&source.Kind{Type: template},
			enqueuer.EnqueueTracked(template, reconciler.DependencyTracker, mgr.GetScheme()),
		); err != nil {
			return fmt.Errorf("watch %T: %w", template, err)
		}
	}

	return nil
}

func registerSupplyChainController(mgr manager.Manager) error {
	repo := repository.NewRepository(
		mgr.GetClient(),
		repository.NewCache(mgr.GetLogger().WithName("supply-chain-repo-cache")),
	)

	reconciler := &supplychain.Reconciler{
		Repo:                    repo,
		ConditionManagerBuilder: conditions.NewConditionManager,
		DependencyTracker:       dependency.NewDependencyTracker(2*defaultResyncTime, mgr.GetLogger().WithName("tracker-supply-chain")),
	}
	ctrl, err := pkgcontroller.New("supply-chain", mgr, pkgcontroller.Options{
		Reconciler: reconciler,
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

	for _, template := range v1alpha1.ValidSupplyChainTemplates {
		if err := ctrl.Watch(
			&source.Kind{Type: template},
			enqueuer.EnqueueTracked(template, reconciler.DependencyTracker, mgr.GetScheme()),
		); err != nil {
			return fmt.Errorf("watch template: %w", err)
		}
	}

	return nil
}

func registerDeliveryController(mgr manager.Manager) error {
	repo := repository.NewRepository(
		mgr.GetClient(),
		repository.NewCache(mgr.GetLogger().WithName("delivery-repo-cache")),
	)

	reconciler := &delivery.Reconciler{
		Repo:              repo,
		DependencyTracker: dependency.NewDependencyTracker(2*defaultResyncTime, mgr.GetLogger().WithName("tracker-delivery")),
	}
	ctrl, err := pkgcontroller.New("delivery", mgr, pkgcontroller.Options{
		Reconciler: reconciler,
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

	for _, template := range v1alpha1.ValidDeliveryTemplates {
		if err := ctrl.Watch(
			&source.Kind{Type: template},
			enqueuer.EnqueueTracked(template, reconciler.DependencyTracker, mgr.GetScheme()),
		); err != nil {
			return fmt.Errorf("watch template: %w", err)
		}
	}

	return nil
}

func registerDeliverableController(mgr manager.Manager) error {
	repo := repository.NewRepository(
		mgr.GetClient(),
		repository.NewCache(mgr.GetLogger().WithName("deliverable-repo-cache")),
	)

	reconciler := &deliverable.Reconciler{
		Repo:                    repo,
		ConditionManagerBuilder: conditions.NewConditionManager,
		ResourceRealizerBuilder: realizerdeliverable.NewResourceRealizerBuilder(
			repository.NewRepository,
			realizerclient.NewClientBuilder(mgr.GetConfig()),
			repository.NewCache(mgr.GetLogger().WithName("deliverable-stamping-repo-cache")),
		),
		Realizer:          realizerdeliverable.NewRealizer(),
		DependencyTracker: dependency.NewDependencyTracker(2*defaultResyncTime, mgr.GetLogger().WithName("tracker-deliverable")),
	}

	ctrl, err := pkgcontroller.New("deliverable", mgr, pkgcontroller.Options{
		Reconciler: reconciler,
	})
	if err != nil {
		return fmt.Errorf("controller new: %w", err)
	}

	reconciler.StampedTracker = &external.ObjectTracker{Controller: ctrl}

	if err := ctrl.Watch(
		&source.Kind{Type: &v1alpha1.Deliverable{}},
		&handler.EnqueueRequestForObject{},
	); err != nil {
		return fmt.Errorf("watch: %w", err)
	}

	mapper := Mapper{
		Client:  mgr.GetClient(),
		Logger:  mgr.GetLogger().WithName("deliverable"),
		Tracker: reconciler.DependencyTracker,
	}

	watches := map[client.Object]handler.MapFunc{
		&v1alpha1.ClusterDelivery{}:  mapper.ClusterDeliveryToDeliverableRequests,
		&corev1.ServiceAccount{}:     mapper.ServiceAccountToDeliverableRequests,
		&rbacv1.Role{}:               mapper.RoleToDeliverableRequests,
		&rbacv1.RoleBinding{}:        mapper.RoleBindingToDeliverableRequests,
		&rbacv1.ClusterRole{}:        mapper.ClusterRoleToDeliverableRequests,
		&rbacv1.ClusterRoleBinding{}: mapper.ClusterRoleBindingToDeliverableRequests,
	}

	for kindType, mapFunc := range watches {
		if err := ctrl.Watch(
			&source.Kind{Type: kindType},
			handler.EnqueueRequestsFromMapFunc(mapFunc),
		); err != nil {
			return fmt.Errorf("watch %T: %w", kindType, err)
		}
	}

	for _, template := range v1alpha1.ValidDeliveryTemplates {
		if err := ctrl.Watch(
			&source.Kind{Type: template},
			enqueuer.EnqueueTracked(template, reconciler.DependencyTracker, mgr.GetScheme()),
		); err != nil {
			return fmt.Errorf("watch %T: %w", template, err)
		}
	}

	return nil
}

func registerRunnableController(mgr manager.Manager) error {
	repo := repository.NewRepository(
		mgr.GetClient(),
		repository.NewCache(mgr.GetLogger().WithName("runnable-repo-cache")),
	)

	reconciler := &runnable.Reconciler{
		Repo:                    repo,
		Realizer:                realizerrunnable.NewRealizer(),
		RunnableCache:           repository.NewCache(mgr.GetLogger().WithName("runnable-stamping-repo-cache")),
		RepositoryBuilder:       repository.NewRepository,
		ClientBuilder:           realizerclient.NewClientBuilder(mgr.GetConfig()),
		ConditionManagerBuilder: conditions.NewConditionManager,
		DependencyTracker:       dependency.NewDependencyTracker(2*defaultResyncTime, mgr.GetLogger().WithName("tracker-runnable")),
	}
	ctrl, err := pkgcontroller.New("runnable-service", mgr, pkgcontroller.Options{
		Reconciler: reconciler,
	})
	if err != nil {
		return fmt.Errorf("controller new runnable-service: %w", err)
	}

	reconciler.StampedTracker = &external.ObjectTracker{Controller: ctrl}

	if err := ctrl.Watch(
		&source.Kind{Type: &v1alpha1.Runnable{}},
		&handler.EnqueueRequestForObject{},
	); err != nil {
		return fmt.Errorf("watch [runnable-service]: %w", err)
	}

	mapper := Mapper{
		Client:  mgr.GetClient(),
		Logger:  mgr.GetLogger().WithName("runnable"),
		Tracker: reconciler.DependencyTracker,
	}

	watches := map[client.Object]handler.MapFunc{
		&corev1.ServiceAccount{}:     mapper.ServiceAccountToRunnableRequests,
		&rbacv1.Role{}:               mapper.RoleToRunnableRequests,
		&rbacv1.RoleBinding{}:        mapper.RoleBindingToRunnableRequests,
		&rbacv1.ClusterRole{}:        mapper.ClusterRoleToRunnableRequests,
		&rbacv1.ClusterRoleBinding{}: mapper.ClusterRoleBindingToRunnableRequests,
	}

	for kindType, mapFunc := range watches {
		if err := ctrl.Watch(
			&source.Kind{Type: kindType},
			handler.EnqueueRequestsFromMapFunc(mapFunc),
		); err != nil {
			return fmt.Errorf("watch %T: %w", kindType, err)
		}
	}

	if err := ctrl.Watch(
		&source.Kind{Type: &v1alpha1.ClusterRunTemplate{}},
		enqueuer.EnqueueTracked(&v1alpha1.ClusterRunTemplate{}, reconciler.DependencyTracker, mgr.GetScheme()),
	); err != nil {
		return fmt.Errorf("watch %T: %w", &v1alpha1.ClusterRunTemplate{}, err)
	}

	return nil
}

func IndexResources(ctx context.Context, mgr manager.Manager) error {
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
