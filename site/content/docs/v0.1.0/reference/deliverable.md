# Deliverable and Delivery Custom Resources

## Deliverable

`Deliverable` allows the operator to pass information about the configuration to be applied to the environment to the
delivery.

```yaml
apiVersion: carto.run/v1alpha1
kind: Deliverable
metadata:
  name: spring-petclinic
  labels:
    # label to be matched against a `ClusterDelivery`s label selector.
    #
    app.tanzu.vmware.com/deliverable-type: web---deliverable # (1)

spec:
  source:
    # source code location in a git repository.
    #
    git:
      url: https://github.com/waciumawanjohi/spring-petclinic.git
      ref:
        branch: "main"
        tag: "v0.0.1"
        commit: "b4df00d"

    # source code location in an image repository
    #
    image: harbor-repo.vmware.com/tanzu_desktop/golang-sample-source@sha256:e508a587

    # subpath in the source code directory that contains the expected code
    # useful when multiple apps are in a single repository
    subPath: app-1

  # any other parameters that don't fit the ones already typed.
  params: []

  # service account with requisite permissions to create objects specified in the delivery
  serviceAccountName: super-secure-service-account
```

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

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterDelivery
metadata:
  name: supplychain
spec:
  # specifies the label key-value pair to select deliverables. (required)
  #
  selector:
    app.tanzu.vmware.com/deliverable-type: web---deliverable

  # see specification of params in supply-chain
  params: []

  # set of resources that will take care of bringing the application to a
  # deliverable state. (required, at least 1)
  #
  resources:
    # name of the resource to be referenced by further resources in the chain.
    # (required, unique)
    #
    - name: config-provider
      # reference to a template object. (required)
      #
      templateRef:
        kind: ClusterSourceTemplate
        name: config-source

    - name: additional-config-provider
      templateRef:
        kind: ClusterSourceTemplate
        name: another-config-source

    - name: deployer
      templateRef:
        kind: ClusterDeploymentTemplate
        name: app-deploy
      # a single resource that provides the location (url and revision) of
      # configuration required for deployment,
      # in a template, these can be consumed as:
      #
      #    $(deployment.url)$
      #    $(deployment.revision)$
      #
      # (required)
      deployment:
        # name of the resource to provide the source information. (required)
        #
        resource: config-provider

      # a set of resources that provide additional information for deployment located
      # at the specified url and revision.
      #
      # in a template, these can be consumed as:
      #
      #    $(sources.<name>.url)$
      #    $(sources.<name>.revision)$
      #
      # if there is only one source, it can be consumed as:
      #
      #    $(source.url)$
      #    $(sources.revision)$
      #
      # (optional)
      sources:
        # name of the resource to provide the source information. (required)
        #
        - resource: additional-config-provider
          # name to be referenced in the template via a query over the list of
          # sources (for instance, `$(sources.provider.url)`.
          #
          # (required, unique in this list)
          #
          name: addtnl

      # see specification for params in supply chain resources
      params: []
```

_ref:
[pkg/apis/v1alpha1/cluster_delivery.go](https://github.com/vmware-tanzu/cartographer/tree/main/pkg/apis/v1alpha1/cluster_delivery.go)_
