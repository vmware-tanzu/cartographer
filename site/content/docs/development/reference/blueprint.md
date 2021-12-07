# Blueprint Custom Resources

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
      images: [ ]

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
      configs: [ ]

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

_ref: [pkg/apis/v1alpha1/cluster_supply_chain.go](../../../../pkg/apis/v1alpha1/cluster_supply_chain.go)_

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
  params: [ ]

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
      params: [ ]
```

_ref: [pkg/apis/v1alpha1/cluster_delivery.go](../../../../pkg/apis/v1alpha1/cluster_delivery.go)_

