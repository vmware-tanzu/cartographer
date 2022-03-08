# Workload and Supply Chain Custom Resources

## Workload

`Workload` allows the developer to pass information about the app to be delivered through the supply chain.

{{< crd  carto.run_workloads.yaml >}}

Notes:

1. labels serve as a way of indirectly selecting `ClusterSupplyChain` - `Workload`s without labels that match a
   `ClusterSupplyChain`'s `spec.selector` won't be reconciled and will stay in an `Errored` state.
2. `spec.image` is useful for enabling workflows that are not based on building the container image from within the
   supplychain, but outside.

_ref:
[pkg/apis/v1alpha1/workload.go](https://github.com/vmware-tanzu/cartographer/tree/main/pkg/apis/v1alpha1/workload.go)_

## ClusterSupplyChain

With a `ClusterSupplyChain`, app operators describe which "shape of applications" they deal with (via `spec.selector`),
and what series of resources are responsible for creating an artifact that delivers it (via `spec.resources`).

Those `Workload`s that match `spec.selector` then go through the resources specified in `spec.resources`.

A resource can emit values, which the supply chain can make available to other resources.

{{< crd  carto.run_clustersupplychains.yaml >}}

_ref:
[pkg/apis/v1alpha1/cluster_supply_chain.go](https://github.com/vmware-tanzu/cartographer/tree/main/pkg/apis/v1alpha1/cluster_supply_chain.go)_
