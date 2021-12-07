# Owner Custom Resources

## Workload

`Workload` allows the developer to pass information about the app to be delivered through the supply chain.

```yaml
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  name: spring-petclinic
  labels:
    # label to be matched against a `ClusterSupplyChain`s label selector.
    #
    app.tanzu.vmware.com/workload-type: web   # (1)

spec:
  # service account with permissions to create resources submitted by the supply chain
  # if not set, will use serviceAccountName from supply chain
  # if that is also not set, will use the default service account in the workload's namespace
  #
  serviceAccountName: workload-service-account

  source:
    # source code location in a git repository.
    #
    git:
      url: https://github.com/scothis/spring-petclinic.git
      ref:
        branch: "main"
        tag: "v0.0.1"
        commit: "b4df00d"

    # image containing the source code to be used throughout
    # the supply chain
    #
    image: harbor-repo.vmware.com/tanzu_desktop/golang-sample-source@sha256:e508a587

  build:
    # environment variables to propagate to a resource responsible
    # for performing a build in the supplychain.
    #
    env:
      - name: CGO_ENABLED
        value: "0"


  # serviceClaims to be bound through service-bindings
  #
  serviceClaims:
    - name: broker
      ref:
        apiVersion: services.tanzu.vmware.com/v1alpha1
        kind: RabbitMQ
        name: rabbit-broker


  # image with the app already built
  #
  image: foo/docker-built@sha256:b4df00d      # (2)

  # environment variables to be passed to the main container
  # running the application.
  #
  env:
    - name: SPRING_PROFILES_ACTIVE
      value: mysql


  # resource constraints for the main application.
  #
  resources:
    requests:
      memory: 1Gi
      cpu: 100m
    limits:
      memory: 1Gi
      cpu: 4000m

  # any other parameters that don't fit the ones already typed.
  params:
    - name: java-version
      # name of the parameter. should match a supply chain parameter name
      value: 11
```

Notes:

1. labels serve as a way of indirectly selecting `ClusterSupplyChain` - `Workload`s without labels that match
   a `ClusterSupplyChain`'s `spec.selector` won't be reconciled and will stay in an `Errored` state.
2. `spec.image` is useful for enabling workflows that are not based on building the container image from within the
   supplychain, but outside.

_ref: [pkg/apis/v1alpha1/workload.go](../../../../pkg/apis/v1alpha1/workload.go)_

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
    app.tanzu.vmware.com/deliverable-type: web---deliverable   # (1)

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
  params: [ ]

  # service account with requisite permissions to create objects specified in the delivery
  serviceAccountName: super-secure-service-account
```

Notes:

1. labels serve as a way of indirectly selecting `ClusterDelivery`

_ref: [pkg/apis/v1alpha1/deliverable.go](../../../../pkg/apis/v1alpha1/deliverable.go)_

