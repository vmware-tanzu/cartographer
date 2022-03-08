# Workload and Supply Chain Custom Resources

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
    app.tanzu.vmware.com/workload-type: web # (1)

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
  image: foo/docker-built@sha256:b4df00d # (2)

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

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSupplyChain
metadata:
  name: supplychain
spec:
  # specifies the label key-value pair to select workloads. (required, one one)
  #
  selector:
    app.tanzu.vmware.com/workload-type: web

  # specifies the service account to be used to create resources if one
  # is not specified in the workload
  #
  # (optional)
  serviceAccountRef:
    name: service-account
    namespace: default

  # parameters to override the defaults from the templates.
  # if a resource in the supply-chain specifies a parameter
  # of the same name that resource parameter clobber what is
  # specified here at the top level (this includes specification
  # as `value` vs `default`)
  #
  # in a template, these can be consumed as:
  #
  #   $(params.<name>)
  #
  # (optional)
  params:
    # name of the parameter. (required, unique in this list, and should match
    # a pre-defined parameter name in a template)
    #
    - name: java-version
      # value to be passed down to the template's parameters,  supporting
      # interpolation.
      #
      value: 6
      # when specified as `value`, a parameter of the same name on the workload will
      # be disregarded.
      #
    - name: jvm
      value: openjdk
      # when specified as `default`, a parameter of the same name on the workload will
      # overwrite this default value.
      #

  # set of resources that will take care of bringing the application to a
  # deliverable state. (required, at least 1)
  #
  resources:
    # name of the resource to be referenced by further resources in the chain.
    # (required, unique)
    #
    - name: source-provider
      # object reference to a template object that instructs how to
      # instantiate and keep the resource up to date. (required)
      #
      templateRef:
        kind: ClusterSourceTemplate
        name: git-repository-battery

    - name: built-image-provider
      templateRef:
        kind: ClusterImageTemplate
        name: kpack-battery

      # a set of resources that provide source information, that is, url and
      # revision.
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
        - resource: source-provider
          # name to be referenced in the template via a query over the list of
          # sources (for instance, `$(sources.provider.url)`.
          #
          # (required, unique in this list)
          #
          name: provider

      # (optional) set of resources that provide image information.
      #
      # in a template, these can be consumed as:
      #
      #   $(images.<name>.image)
      #
      # if there is only one image, it can be consumed as:
      #
      #   $(image)
      #
      images: []

      # (optional) set of resources that provide kubernetes configuration,
      # for instance, podTemplateSpecs.
      # in a template, these can be consumed as:
      #
      #   $(configs.<name>.config)
      #
      # if there is only one config, it can be consumed as:
      #
      #   $(config)
      #
      configs: []

      # parameters to override the defaults from the templates.
      # resource parameters override any parameter of the same name set
      # for the overall supply-chain in spec.params
      # in a template, these can be consumed as:
      #
      #   $(params.<name>)
      #
      # (optional)
      params:
        # name of the parameter. (required, unique in this list, and should match
        # template's pre-defined set of parameters)
        #
        - name: java-version
          # value to be passed down to the template's parameters,  supporting
          # interpolation.
          #
          default: 9
          # when specified as `default`, a parameter of the same name on the workload will
          # overwrite this default value.
          #
        - name: jvm
          value: openjdk
          # when specified as `value`, a parameter of the same name on the workload will
          # be disregarded
          #
```

_ref:
[pkg/apis/v1alpha1/cluster_supply_chain.go](https://github.com/vmware-tanzu/cartographer/tree/main/pkg/apis/v1alpha1/cluster_supply_chain.go)_
