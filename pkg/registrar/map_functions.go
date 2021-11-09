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

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

//counterfeiter:generate sigs.k8s.io/controller-runtime/pkg/client.Client

//counterfeiter:generate . Logger
type Logger interface {
	Error(err error, msg string, keysAndValues ...interface{})
}

type Mapper struct {
	Client client.Client
	// fixme We should accept the context, not the logger - then we get the right logger and so does the client
	Logger Logger
}

func (mapper *Mapper) TemplateToDeliverableRequests(template client.Object) []reconcile.Request {
	deliveries := mapper.templateToDeliveries(template)

	var requests []reconcile.Request
	for _, delivery := range deliveries {
		reqs := mapper.ClusterDeliveryToDeliverableRequests(delivery)
		requests = append(requests, reqs...)
	}

	return requests
}

func (mapper *Mapper) TemplateToWorkloadRequests(template client.Object) []reconcile.Request {
	supplyChains := mapper.templateToSupplyChains(template)

	var requests []reconcile.Request
	for _, supplyChain := range supplyChains {
		reqs := mapper.ClusterSupplyChainToWorkloadRequests(&supplyChain)
		requests = append(requests, reqs...)
	}

	return requests
}

func (mapper *Mapper) templateToSupplyChains(template client.Object) []v1alpha1.ClusterSupplyChain {
	templateName := template.GetName()

	err := mapper.addGVK(template)
	if err != nil {
		mapper.Logger.Error(err, fmt.Sprintf("could not get GVK for template: %s", templateName))
		return nil
	}

	list := &v1alpha1.ClusterSupplyChainList{}

	err = mapper.Client.List(
		context.TODO(),
		list,
	)

	if err != nil {
		mapper.Logger.Error(err, "list ClusterSupplyChains")
		return nil
	}

	templateKind := template.GetObjectKind().GroupVersionKind().Kind

	var supplyChains []v1alpha1.ClusterSupplyChain
	for _, sc := range list.Items {
		for _, res := range sc.Spec.Resources {
			if res.TemplateRef.Kind == templateKind && res.TemplateRef.Name == templateName {
				supplyChains = append(supplyChains, sc)
			}
		}
	}
	return supplyChains
}

func (mapper *Mapper) ClusterSupplyChainToWorkloadRequests(object client.Object) []reconcile.Request {
	var err error

	supplyChain, ok := object.(*v1alpha1.ClusterSupplyChain)
	if !ok {
		mapper.Logger.Error(nil, "cluster supply chain to workload requests: cast to ClusterSupplyChain failed")
		return nil
	}

	list := &v1alpha1.WorkloadList{}

	err = mapper.Client.List(context.TODO(), list,
		client.InNamespace(supplyChain.Namespace),
		client.MatchingLabels(supplyChain.Spec.Selector))
	if err != nil {
		mapper.Logger.Error(fmt.Errorf("client list: %w", err), "cluster supply chain to workload requests: client list")
		return nil
	}

	var requests []reconcile.Request
	for _, workload := range list.Items {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      workload.Name,
				Namespace: workload.Namespace,
			},
		})
	}

	return requests
}

func (mapper *Mapper) ClusterDeliveryToDeliverableRequests(object client.Object) []reconcile.Request {
	var err error

	delivery, ok := object.(*v1alpha1.ClusterDelivery)
	if !ok {
		mapper.Logger.Error(nil, "cluster delivery to deliverable requests: cast to ClusterDelivery failed")
		return nil
	}

	list := &v1alpha1.DeliverableList{}

	err = mapper.Client.List(context.TODO(), list,
		client.InNamespace(delivery.Namespace),
		client.MatchingLabels(delivery.Spec.Selector))
	if err != nil {
		mapper.Logger.Error(fmt.Errorf("client list: %w", err), "cluster delivery to deliverable requests: client list")
		return nil
	}

	var requests []reconcile.Request
	for _, deliverable := range list.Items {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      deliverable.Name,
				Namespace: deliverable.Namespace,
			},
		})
	}

	return requests
}

func (mapper *Mapper) RunTemplateToRunnableRequests(object client.Object) []reconcile.Request {
	var err error

	runTemplate, ok := object.(*v1alpha1.ClusterRunTemplate)
	if !ok {
		mapper.Logger.Error(nil, "run template to runnable requests: cast to run template failed")
		return nil
	}

	list := &v1alpha1.RunnableList{}

	err = mapper.Client.List(context.TODO(), list)
	if err != nil {
		mapper.Logger.Error(fmt.Errorf("client list: %w", err), "run template to runnable requests: client list")
		return nil
	}

	var requests []reconcile.Request
	for _, runnable := range list.Items {

		if runTemplateRefMatch(runnable.Spec.RunTemplateRef, runTemplate) {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      runnable.Name,
					Namespace: runnable.Namespace,
				},
			})
		}
	}

	return requests
}

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

func (mapper *Mapper) TemplateToSupplyChainRequests(template client.Object) []reconcile.Request {
	supplyChains := mapper.templateToSupplyChains(template)

	var requests []reconcile.Request
	for _, supplyChain := range supplyChains {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name: supplyChain.Name,
			},
		})
	}

	return requests
}

func (mapper *Mapper) TemplateToDeliveryRequests(template client.Object) []reconcile.Request {
	deliveries := mapper.templateToDeliveries(template)

	var requests []reconcile.Request
	for _, delivery := range deliveries {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name: delivery.Name,
			},
		})
	}

	return requests
}

func (mapper *Mapper) templateToDeliveries(template client.Object) []*v1alpha1.ClusterDelivery {
	templateName := template.GetName()

	err := mapper.addGVK(template)
	if err != nil {
		mapper.Logger.Error(err, fmt.Sprintf("could not get GVK for template: %s", templateName))
		return nil
	}

	list := &v1alpha1.ClusterDeliveryList{}

	err = mapper.Client.List(
		context.TODO(),
		list,
	)

	if err != nil {
		mapper.Logger.Error(err, "list ClusterDeliveries")
		return nil
	}

	templateKind := template.GetObjectKind().GroupVersionKind().Kind

	var deliveries []*v1alpha1.ClusterDelivery
	for dIndex := range list.Items {
		d := list.Items[dIndex]
		for _, res := range d.Spec.Resources {
			if res.TemplateRef.Kind == templateKind && res.TemplateRef.Name == templateName {
				deliveries = append(deliveries, &d)
			}
		}
	}
	return deliveries
}

func runTemplateRefMatch(ref v1alpha1.TemplateReference, runTemplate *v1alpha1.ClusterRunTemplate) bool {
	if ref.Name != runTemplate.Name {
		return false
	}

	return ref.Kind == "ClusterRunTemplate" || ref.Kind == ""
}
