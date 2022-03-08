# Deliverable and Delivery Custom Resources

## Deliverable

`Deliverable` allows the operator to pass information about the configuration to be applied to the environment to the
delivery.

{{< crd  carto.run_deliverables.yaml >}}

Notes:

1. labels serve as a way of indirectly selecting `ClusterDelivery`

_ref:
[pkg/apis/v1alpha1/deliverable.go](https://github.com/vmware-tanzu/cartographer/tree/main/pkg/apis/v1alpha1/deliverable.go)_

## ClusterDelivery

A `ClusterDelivery` is a cluster-scoped resources that enables application operators to define a continuous delivery
workflow. Delivery is analogous to SupplyChain, in that it specifies a list of resources that are created when requested
by the developer. Early resources in the delivery are expected to configure the k8s environment (for example by
deploying an application). Later resources validate the environment is healthy.

The SupplyChain resources `ClusterSourceTemplates` and `ClusterTemplates` are valid for delivery. Delivery additionally
has the resource `ClusterDeploymentTemplates`. Delivery can cast the values from a `ClusterSourceTemplate` so that they
may be consumed by a `ClusterDeploymentTemplate`.

`ClusterDeliveries` specify the type of configuration they accept through the `spec.selector` field. `Deliverable`s with
matching `spec.selector` then create a logical delivery. This makes the values in the `Deliverable` available to all of
the resources in the `ClusterDelivery`s `spec.resources`.

{{< crd  carto.run_clusterdeliveries.yaml >}}

_ref:
[pkg/apis/v1alpha1/cluster_delivery.go](https://github.com/vmware-tanzu/cartographer/tree/main/pkg/apis/v1alpha1/cluster_delivery.go)_
