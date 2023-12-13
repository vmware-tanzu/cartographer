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

package mapper

//go:generate go run -modfile ../../hack/tools/go.mod github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/tracker/dependency"
)

//counterfeiter:generate sigs.k8s.io/controller-runtime/pkg/client.Client

//counterfeiter:generate . Logger
type Logger interface {
	Error(err error, msg string, keysAndValues ...interface{})
}

type Mapper struct {
	Client client.Client
	// fixme We should accept the context, not the logger - then we get the right logger and so does the client
	Logger  Logger
	Tracker dependency.DependencyTracker
}

// Workload

func (mapper *Mapper) ClusterSupplyChainToWorkloadRequests(_ context.Context, _ client.Object) []reconcile.Request {
	workloadList := &v1alpha1.WorkloadList{}
	err := mapper.Client.List(context.TODO(), workloadList)
	if err != nil {
		mapper.Logger.Error(err, "cluster supply chain to workload requests: client list workloads")
		return nil
	}

	var requests []reconcile.Request
	for _, workload := range workloadList.Items {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      workload.Name,
				Namespace: workload.Namespace,
			},
		})
	}

	return requests
}

func (mapper *Mapper) ServiceAccountToWorkloadRequests(_ context.Context, serviceAccountObject client.Object) []reconcile.Request {
	err := mapper.addGVK(serviceAccountObject)
	if err != nil {
		mapper.Logger.Error(err, fmt.Sprintf("could not get GVK for service account: %s", serviceAccountObject.GetName()))
		return nil
	}

	wks := mapper.Tracker.Lookup(dependency.Key{
		GroupKind: serviceAccountObject.GetObjectKind().GroupVersionKind().GroupKind(),
		NamespacedName: types.NamespacedName{
			Namespace: serviceAccountObject.GetNamespace(),
			Name:      serviceAccountObject.GetName(),
		},
	})

	var requests []reconcile.Request
	for _, wk := range wks {
		requests = append(requests, reconcile.Request{NamespacedName: wk})
	}

	return requests
}

func (mapper *Mapper) RoleBindingToWorkloadRequests(ctx context.Context, roleBindingObject client.Object) []reconcile.Request {
	roleBinding, ok := roleBindingObject.(*rbacv1.RoleBinding)
	if !ok {
		mapper.Logger.Error(nil, "role binding to workload requests: cast to RoleBinding failed")
		return nil
	}

	serviceAccounts := mapper.getServiceAccounts(roleBinding.Subjects, "role binding to workload requests")

	var requests []reconcile.Request
	for _, serviceAccount := range serviceAccounts {
		requests = append(requests, mapper.ServiceAccountToWorkloadRequests(ctx, serviceAccount)...)
	}

	return requests
}

func (mapper *Mapper) ClusterRoleBindingToWorkloadRequests(ctx context.Context, clusterRoleBindingObject client.Object) []reconcile.Request {
	clusterRoleBinding, ok := clusterRoleBindingObject.(*rbacv1.ClusterRoleBinding)
	if !ok {
		mapper.Logger.Error(nil, "cluster role binding to workload requests: cast to ClusterRoleBinding failed")
		return nil
	}

	serviceAccounts := mapper.getServiceAccounts(clusterRoleBinding.Subjects, "cluster role binding to workload requests")

	var requests []reconcile.Request
	for _, serviceAccount := range serviceAccounts {
		requests = append(requests, mapper.ServiceAccountToWorkloadRequests(ctx, serviceAccount)...)
	}

	return requests
}

func (mapper *Mapper) RoleToWorkloadRequests(ctx context.Context, roleObject client.Object) []reconcile.Request {
	role, ok := roleObject.(*rbacv1.Role)
	if !ok {
		mapper.Logger.Error(nil, "role to workload requests: cast to Role failed")
		return nil
	}

	list := &rbacv1.RoleBindingList{}

	err := mapper.Client.List(context.TODO(), list, client.InNamespace(role.Namespace))
	if err != nil {
		mapper.Logger.Error(fmt.Errorf("client list: %w", err), "role to workload requests: list role bindings")
		return nil
	}

	var requests []reconcile.Request
	for _, roleBinding := range list.Items {
		if roleBinding.RoleRef.APIGroup == "" && roleBinding.RoleRef.Kind == "Role" && roleBinding.RoleRef.Name == role.Name {
			requests = append(requests, mapper.RoleBindingToWorkloadRequests(ctx, &roleBinding)...)
		}
	}

	return requests
}

func (mapper *Mapper) ClusterRoleToWorkloadRequests(ctx context.Context, clusterRoleObject client.Object) []reconcile.Request {
	clusterRole, ok := clusterRoleObject.(*rbacv1.ClusterRole)
	if !ok {
		mapper.Logger.Error(nil, "cluster role to workload requests: cast to ClusterRole failed")
		return nil
	}

	clusterRoleBindingList := &rbacv1.ClusterRoleBindingList{}

	err := mapper.Client.List(context.TODO(), clusterRoleBindingList)
	if err != nil {
		mapper.Logger.Error(fmt.Errorf("client list: %w", err), "cluster role to workload requests: list cluster role bindings")
		return nil
	}

	var requests []reconcile.Request

	for _, clusterRoleBinding := range clusterRoleBindingList.Items {
		if clusterRoleBinding.RoleRef.APIGroup == "" && clusterRoleBinding.RoleRef.Kind == "ClusterRole" && clusterRoleBinding.RoleRef.Name == clusterRole.Name {
			requests = append(requests, mapper.ClusterRoleBindingToWorkloadRequests(ctx, &clusterRoleBinding)...)
		}
	}

	roleBindingList := &rbacv1.RoleBindingList{}

	err = mapper.Client.List(context.TODO(), roleBindingList)
	if err != nil {
		mapper.Logger.Error(fmt.Errorf("client list: %w", err), "cluster role role to workload requests: list role bindings")
		return nil
	}

	for _, roleBinding := range roleBindingList.Items {
		if roleBinding.RoleRef.APIGroup == "" && roleBinding.RoleRef.Kind == "ClusterRole" && roleBinding.RoleRef.Name == clusterRole.Name {
			requests = append(requests, mapper.RoleBindingToWorkloadRequests(ctx, &roleBinding)...)
		}
	}

	return requests
}

// Deliverable

func (mapper *Mapper) ClusterDeliveryToDeliverableRequests(_ context.Context, _ client.Object) []reconcile.Request {
	deliverableList := &v1alpha1.DeliverableList{}
	err := mapper.Client.List(context.TODO(), deliverableList)
	if err != nil {
		mapper.Logger.Error(err, "cluster delivery to deliverable requests: client list deliverables")
		return nil
	}

	var requests []reconcile.Request
	for _, deliverable := range deliverableList.Items {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      deliverable.Name,
				Namespace: deliverable.Namespace,
			},
		})
	}

	return requests
}

func (mapper *Mapper) ServiceAccountToDeliverableRequests(_ context.Context, serviceAccountObject client.Object) []reconcile.Request {
	err := mapper.addGVK(serviceAccountObject)
	if err != nil {
		mapper.Logger.Error(err, fmt.Sprintf("could not get GVK for service account: %s", serviceAccountObject.GetName()))
		return nil
	}

	deliverables := mapper.Tracker.Lookup(dependency.Key{
		GroupKind: serviceAccountObject.GetObjectKind().GroupVersionKind().GroupKind(),
		NamespacedName: types.NamespacedName{
			Namespace: serviceAccountObject.GetNamespace(),
			Name:      serviceAccountObject.GetName(),
		},
	})

	var requests []reconcile.Request
	for _, deliverable := range deliverables {
		requests = append(requests, reconcile.Request{NamespacedName: deliverable})
	}

	return requests
}

func (mapper *Mapper) RoleBindingToDeliverableRequests(ctx context.Context, roleBindingObject client.Object) []reconcile.Request {
	roleBinding, ok := roleBindingObject.(*rbacv1.RoleBinding)
	if !ok {
		mapper.Logger.Error(nil, "role binding to deliverable requests: cast to RoleBinding failed")
		return nil
	}

	serviceAccounts := mapper.getServiceAccounts(roleBinding.Subjects, "role binding to deliverable requests")

	var requests []reconcile.Request
	for _, serviceAccount := range serviceAccounts {
		requests = append(requests, mapper.ServiceAccountToDeliverableRequests(ctx, serviceAccount)...)
	}

	return requests
}

func (mapper *Mapper) ClusterRoleBindingToDeliverableRequests(ctx context.Context, clusterRoleBindingObject client.Object) []reconcile.Request {
	clusterRoleBinding, ok := clusterRoleBindingObject.(*rbacv1.ClusterRoleBinding)
	if !ok {
		mapper.Logger.Error(nil, "cluster role binding to deliverable requests: cast to ClusterRoleBinding failed")
		return nil
	}

	serviceAccounts := mapper.getServiceAccounts(clusterRoleBinding.Subjects, "cluster role binding to deliverable requests")

	var requests []reconcile.Request
	for _, serviceAccount := range serviceAccounts {
		requests = append(requests, mapper.ServiceAccountToDeliverableRequests(ctx, serviceAccount)...)
	}

	return requests
}

func (mapper *Mapper) RoleToDeliverableRequests(ctx context.Context, roleObject client.Object) []reconcile.Request {
	role, ok := roleObject.(*rbacv1.Role)
	if !ok {
		mapper.Logger.Error(nil, "role to deliverable requests: cast to Role failed")
		return nil
	}

	list := &rbacv1.RoleBindingList{}

	err := mapper.Client.List(context.TODO(), list, client.InNamespace(role.Namespace))
	if err != nil {
		mapper.Logger.Error(fmt.Errorf("client list: %w", err), "role to deliverable requests: list role bindings")
		return nil
	}

	var requests []reconcile.Request
	for _, roleBinding := range list.Items {
		if roleBinding.RoleRef.APIGroup == "" && roleBinding.RoleRef.Kind == "Role" && roleBinding.RoleRef.Name == role.Name {
			requests = append(requests, mapper.RoleBindingToDeliverableRequests(ctx, &roleBinding)...)
		}
	}

	return requests
}

func (mapper *Mapper) ClusterRoleToDeliverableRequests(ctx context.Context, clusterRoleObject client.Object) []reconcile.Request {
	clusterRole, ok := clusterRoleObject.(*rbacv1.ClusterRole)
	if !ok {
		mapper.Logger.Error(nil, "cluster role to deliverable requests: cast to ClusterRole failed")
		return nil
	}

	clusterRoleBindingList := &rbacv1.ClusterRoleBindingList{}

	err := mapper.Client.List(context.TODO(), clusterRoleBindingList)
	if err != nil {
		mapper.Logger.Error(fmt.Errorf("client list: %w", err), "cluster role to deliverable requests: list cluster role bindings")
		return nil
	}

	var requests []reconcile.Request

	for _, clusterRoleBinding := range clusterRoleBindingList.Items {
		if clusterRoleBinding.RoleRef.APIGroup == "" && clusterRoleBinding.RoleRef.Kind == "ClusterRole" && clusterRoleBinding.RoleRef.Name == clusterRole.Name {
			requests = append(requests, mapper.ClusterRoleBindingToDeliverableRequests(ctx, &clusterRoleBinding)...)
		}
	}

	roleBindingList := &rbacv1.RoleBindingList{}

	err = mapper.Client.List(context.TODO(), roleBindingList)
	if err != nil {
		mapper.Logger.Error(fmt.Errorf("client list: %w", err), "cluster role role to deliverable requests: list role bindings")
		return nil
	}

	for _, roleBinding := range roleBindingList.Items {
		if roleBinding.RoleRef.APIGroup == "" && roleBinding.RoleRef.Kind == "ClusterRole" && roleBinding.RoleRef.Name == clusterRole.Name {
			requests = append(requests, mapper.RoleBindingToDeliverableRequests(ctx, &roleBinding)...)
		}
	}

	return requests
}

// Runnable

func (mapper *Mapper) ServiceAccountToRunnableRequests(ctx context.Context, serviceAccountObject client.Object) []reconcile.Request {
	err := mapper.addGVK(serviceAccountObject)
	if err != nil {
		mapper.Logger.Error(err, fmt.Sprintf("could not get GVK for service account: %s", serviceAccountObject.GetName()))
		return nil
	}

	runnables := mapper.Tracker.Lookup(dependency.Key{
		GroupKind: serviceAccountObject.GetObjectKind().GroupVersionKind().GroupKind(),
		NamespacedName: types.NamespacedName{
			Namespace: serviceAccountObject.GetNamespace(),
			Name:      serviceAccountObject.GetName(),
		},
	})

	var requests []reconcile.Request
	for _, runnable := range runnables {
		requests = append(requests, reconcile.Request{NamespacedName: runnable})
	}

	return requests
}

func (mapper *Mapper) RoleBindingToRunnableRequests(ctx context.Context, roleBindingObject client.Object) []reconcile.Request {
	roleBinding, ok := roleBindingObject.(*rbacv1.RoleBinding)
	if !ok {
		mapper.Logger.Error(nil, "role binding to runnable requests: cast to RoleBinding failed")
		return nil
	}

	serviceAccounts := mapper.getServiceAccounts(roleBinding.Subjects, "role binding to runnable requests")

	var requests []reconcile.Request
	for _, serviceAccount := range serviceAccounts {
		requests = append(requests, mapper.ServiceAccountToRunnableRequests(ctx, serviceAccount)...)
	}

	return requests
}

func (mapper *Mapper) ClusterRoleBindingToRunnableRequests(ctx context.Context, clusterRoleBindingObject client.Object) []reconcile.Request {
	clusterRoleBinding, ok := clusterRoleBindingObject.(*rbacv1.ClusterRoleBinding)
	if !ok {
		mapper.Logger.Error(nil, "cluster role binding to runnable requests: cast to ClusterRoleBinding failed")
		return nil
	}

	serviceAccounts := mapper.getServiceAccounts(clusterRoleBinding.Subjects, "cluster role binding to runnable requests")

	var requests []reconcile.Request
	for _, serviceAccount := range serviceAccounts {
		requests = append(requests, mapper.ServiceAccountToRunnableRequests(ctx, serviceAccount)...)
	}

	return requests
}

func (mapper *Mapper) RoleToRunnableRequests(ctx context.Context, roleObject client.Object) []reconcile.Request {
	role, ok := roleObject.(*rbacv1.Role)
	if !ok {
		mapper.Logger.Error(nil, "role to runnable requests: cast to Role failed")
		return nil
	}

	list := &rbacv1.RoleBindingList{}

	err := mapper.Client.List(context.TODO(), list, client.InNamespace(role.Namespace))
	if err != nil {
		mapper.Logger.Error(fmt.Errorf("client list: %w", err), "role to runnable requests: list role bindings")
		return nil
	}

	var requests []reconcile.Request
	for _, roleBinding := range list.Items {
		if roleBinding.RoleRef.APIGroup == "" && roleBinding.RoleRef.Kind == "Role" && roleBinding.RoleRef.Name == role.Name {
			requests = append(requests, mapper.RoleBindingToRunnableRequests(ctx, &roleBinding)...)
		}
	}

	return requests
}

func (mapper *Mapper) ClusterRoleToRunnableRequests(ctx context.Context, clusterRoleObject client.Object) []reconcile.Request {
	clusterRole, ok := clusterRoleObject.(*rbacv1.ClusterRole)
	if !ok {
		mapper.Logger.Error(nil, "cluster role to runnable requests: cast to ClusterRole failed")
		return nil
	}

	clusterRoleBindingList := &rbacv1.ClusterRoleBindingList{}

	err := mapper.Client.List(context.TODO(), clusterRoleBindingList)
	if err != nil {
		mapper.Logger.Error(fmt.Errorf("client list: %w", err), "cluster role to runnable requests: list cluster role bindings")
		return nil
	}

	var requests []reconcile.Request

	for _, clusterRoleBinding := range clusterRoleBindingList.Items {
		if clusterRoleBinding.RoleRef.APIGroup == "" && clusterRoleBinding.RoleRef.Kind == "ClusterRole" && clusterRoleBinding.RoleRef.Name == clusterRole.Name {
			requests = append(requests, mapper.ClusterRoleBindingToRunnableRequests(ctx, &clusterRoleBinding)...)
		}
	}

	roleBindingList := &rbacv1.RoleBindingList{}

	err = mapper.Client.List(context.TODO(), roleBindingList)
	if err != nil {
		mapper.Logger.Error(fmt.Errorf("client list: %w", err), "cluster role role to runnable requests: list role bindings")
		return nil
	}

	for _, roleBinding := range roleBindingList.Items {
		if roleBinding.RoleRef.APIGroup == "" && roleBinding.RoleRef.Kind == "ClusterRole" && roleBinding.RoleRef.Name == clusterRole.Name {
			requests = append(requests, mapper.RoleBindingToRunnableRequests(ctx, &roleBinding)...)
		}
	}

	return requests
}

// Shared

// addGVK fulfills the 'GVK of an object returned from the APIServer
// https://github.com/kubernetes-sigs/controller-runtime/issues/1517#issuecomment-844703142
func (mapper *Mapper) addGVK(obj client.Object) error {
	gvks, unversioned, err := mapper.Client.Scheme().ObjectKinds(obj)
	if err != nil {
		return fmt.Errorf("missing apiVersion or kind: %s err: %w", obj.GetName(), err)
	}

	if unversioned {
		return fmt.Errorf("unversioned object: %s", obj.GetName())
	}

	if len(gvks) != 1 {
		return fmt.Errorf("unexpected GVK count: %s", obj.GetName())
	}

	obj.GetObjectKind().SetGroupVersionKind(gvks[0])
	return nil
}

func (mapper *Mapper) getServiceAccounts(subjects []rbacv1.Subject, logPrefix string) []*corev1.ServiceAccount {
	var serviceAccounts []*corev1.ServiceAccount
	for _, subject := range subjects {
		if subject.APIGroup == "" && subject.Kind == "ServiceAccount" {
			namespace := "default"
			if subject.Namespace != "" {
				namespace = subject.Namespace
			}
			serviceAccountKey := client.ObjectKey{
				Namespace: namespace,
				Name:      subject.Name,
			}

			serviceAccountObject := &corev1.ServiceAccount{}
			err := mapper.Client.Get(context.TODO(), serviceAccountKey, serviceAccountObject)
			if err != nil {
				mapper.Logger.Error(fmt.Errorf("client get: %w", err), fmt.Sprintf("%s: get service account", logPrefix))
			}
			serviceAccounts = append(serviceAccounts, serviceAccountObject)
		}
	}
	return serviceAccounts
}
