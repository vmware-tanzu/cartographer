/*
Copyright 2021 VMware

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	cartov1alpha1 "github.com/vmware-tanzu/cartographer/pkg/apis/carto/v1alpha1"
	versioned "github.com/vmware-tanzu/cartographer/pkg/generated/clientset/versioned"
	internalinterfaces "github.com/vmware-tanzu/cartographer/pkg/generated/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/vmware-tanzu/cartographer/pkg/generated/listers/carto/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// ClusterDeliveryInformer provides access to a shared informer and lister for
// ClusterDeliveries.
type ClusterDeliveryInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.ClusterDeliveryLister
}

type clusterDeliveryInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewClusterDeliveryInformer constructs a new informer for ClusterDelivery type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewClusterDeliveryInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredClusterDeliveryInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredClusterDeliveryInformer constructs a new informer for ClusterDelivery type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredClusterDeliveryInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CartoV1alpha1().ClusterDeliveries(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CartoV1alpha1().ClusterDeliveries(namespace).Watch(context.TODO(), options)
			},
		},
		&cartov1alpha1.ClusterDelivery{},
		resyncPeriod,
		indexers,
	)
}

func (f *clusterDeliveryInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredClusterDeliveryInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *clusterDeliveryInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&cartov1alpha1.ClusterDelivery{}, f.defaultInformer)
}

func (f *clusterDeliveryInformer) Lister() v1alpha1.ClusterDeliveryLister {
	return v1alpha1.NewClusterDeliveryLister(f.Informer().GetIndexer())
}
