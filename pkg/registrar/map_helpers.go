package registrar

import (
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

//type deliveryMapToRequestFn func(delivery v1alpha1.ClusterDelivery) reconcile.Request
//
//func deliveriesMapToRequest(items []v1alpha1.ClusterDelivery, fn deliveryMapToRequestFn) []reconcile.Request {
//	var mappedList []reconcile.Request
//	for _, item := range items {
//		mappedList = append(mappedList, fn(item))
//	}
//	return mappedList
//}

type deliveryFilterFn func(delivery v1alpha1.ClusterDelivery) bool

func deliveriesFilter(items []v1alpha1.ClusterDelivery, fn deliveryFilterFn) []v1alpha1.ClusterDelivery {
	var filteredList []v1alpha1.ClusterDelivery
	for _, item := range items {
		if fn(item) {
			filteredList = append(filteredList, item)
		}
	}
	return filteredList
}

type deliveryResourceFilterFn func(resource v1alpha1.ClusterDeliveryResource) bool

func deliveryResourceAny(items []v1alpha1.ClusterDeliveryResource, fn deliveryResourceFilterFn) bool {
	for _, item := range items {
		if fn(item) {
			return true
		}
	}
	return false
}
