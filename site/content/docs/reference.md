# Spec Reference

## GVK

### Version

All of the custom resources that Cartographer is working on are being written under `v1alpha1` to indicate that our first version of it is at the "alpha stability level", and that it's our first iteration on it.

See [versions in CustomResourceDefinitions].

[versions in CustomResourceDefinitions]: https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definition-versioning/


### Group

All of our custom resources under the `carto.run` group[^1].

For instance:

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSupplyChain
```

## Resources

Cartographer is composed of several custom resources, some of them being cluster-wide:

- `ClusterSupplyChain`
- `ClusterSourceTemplate`
- `ClusterImageTemplate`
- `ClusterConfigTemplate`
- `ClusterTemplate`

and two that are namespace-scoped:

- `Workload`


### Workload

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
  #
  params:
    - name: my-company.com/defaults/java-version
      value: 11
    - name: debug
      value: true
```

notes:

1. labels serve as a way of indirectly selecting `ClusterSupplyChain` - `Workload`s without labels that match a `ClusterSupplyChain`'s `spec.selector` won't be reconciled and will stay in an `Errored` state.

2. `spec.image` is useful for enabling workflows that are not based on building the container image from within the supplychain, but outside. 

_ref: [pkg/apis/v1alpha1/workload.go](https://github.com/vmware-tanzu/cartographer/blob/v0.0.3/pkg/apis/v1alpha1/workload.go)_


### ClusterSupplyChain

With a `ClusterSupplyChain`, app operators describe which "shape of applications" they deal with (via `spec.selector`), and what series of components are responsible for creating an artifact that delivers it (via `spec.components`).

Those `Workload`s that match `spec.selector` then go through the components specified in `spec.components`.


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


  # set of components that will take care of bringing the application to a
  # deliverable state. (required, at least 1)
  #
  components:
    # name of the component to be referenced by further components in the chain.
    # (required, unique)
    #
    - name: source-provider
      # object reference to a template object that instructs how to
      # instantiate and keep the component up to date. (required)
      #
      templateRef:
        kind: ClusterSourceTemplate
        name: git-repository-battery

    - name: built-image-provider
      templateRef:
        kind: ClusterImageTemplate
        name: kpack-battery

      # a set of components that provide source information, that is, url and
      # revision.
      # 
      # in a template, these can be consumed as: 
      #
      #    $(sources[<idx>]url)$
      #    $(sources[<idx>]revision)$
      #
      # (optional)
      sources:
        # name of the component to provide the source information. (required)
        #
        - component: source-provider
          # name to be referenced in the template via a query over the list of
          # sources (for instance, `$(sources.$(name=="provider").url)`.
          #
          # (required, unique in this list)
          #
          name: provider

      # (optional) set of components that provide image information.
      #
      # in a template, these can be consumed as:
      #
      #   $(images[<idx>].image)
      #
      images: []

      # (optional) set of components that provide kubernetes configuration,
      # for instance, podTemplateSpecs.
      # in a template, these can be consumed as:
      #
      #   $(configs[<idx>].config)
      #
      configs: []

      # parameters to override the defaults from the templates.
      # (optional)
      #
      params:
        # name of the parameter. (required, unique in this list, and must match
        # template's pre-defined set of parameters)
        #
        - name: java-version
          # value to be passed down to the template's parameters, supporting
          # interpolation.
          #
          value: $(workload.spec.params[?(@.name=="nebhale-io/java-version")])$
        - name: jvm
          value: openjdk
```


_ref: [pkg/apis/v1alpha1/cluster_supply_chain.go](https://github.com/vmware-tanzu/cartographer/blob/v0.0.3/pkg/apis/v1alpha1/cluster_supply_chain.go)_


### ClusterSourceTemplate

`ClusterSourceTemplate` indicates how the supply chain could instantiate a provider of source code information (url and revision).


```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSourceTemplate
metadata:
  name: git-repository-battery
spec:
  # default set of parameters. (optional)
  #
  params:
      # name of the parameter (required, unique in this list)
      #
    - name: git-implementation
      # default value if not specified in the component that references 
      # this templateClusterSupplyChain (required)
      #
      default: libgit2

  # jsonpath expression to instruct where in the object templated out source
  # code url information can be found. (required)
  #
  urlPath: .status.artifact.url

  # jsonpath expression to instruct where in the object templated out 
  # source code revision information can be found. (required)
  #
  revisionPath: .status.artifact.revision

  # template for instantiating the source provider.
  #
  # data available for interpolation (`$(<json_path>)$`:
  #
  #     - workload  (access to the whole workload object)
  #     - params
  #     - sources   (if specified in the supply chain)
  #     - images    (if specified in the supply chain)
  #     - configs   (if specified in the supply chain)
  #
  # (required)
  #
  template:
    apiVersion: source.toolkit.fluxcd.io/v1beta1
    kind: GitRepository
    metadata:
      name: $(workload.metadata.name)$-source
    spec:
      interval: 3m
      url: $(workload.spec.source.git.url)$
      ref: $(workload.spec.source.git.ref)$
      gitImplementation: $(params[?(@.name=="git-implementation")].value)$
      ignore: ""
```

_ref: [pkg/apis/v1alpha1/cluster_source_template.go](https://github.com/vmware-tanzu/cartographer/blob/v0.0.3/pkg/apis/v1alpha1/cluster_source_template.go)_


### ClusterImageTemplate

`ClusterImageTemplate` instructs how the supply chain should instantiate an object responsible for supplying container images, for instance, one that takes source code, builds a container image out of it and presents under its `.status` the reference to that produced image.


```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterImageTemplate
metadata:
  name: kpack-battery
spec:
  # default set of parameters. see ClusterSourceTemplate for more
  # information. (optional)
  #
  params: []

  # jsonpath expression to instruct where in the object templated out container 
  # image information can be found. (required)
  #
  imagePath: .status.latestImage

  # template for instantiating the image provider.
  # same data available for interpolation as any other `*Template`. (required)
  #
  template:
    apiVersion: kpack.io/v1alpha1
    kind: Image
    metadata:
      name: $(workload.metadata.name)$-image
    spec:
      tag: projectcartographer/demo/$(workload.metadata.name)$
      serviceAccount: service-account
      builder:
        kind: ClusterBuilder
        name: java-builder
      source:
        blob:
          url: $(sources[0].url)$
```

_ref: [pkg/apis/v1alpha1/cluster_image_template.go](https://github.com/vmware-tanzu/cartographer/blob/v0.0.3/pkg/apis/v1alpha1/cluster_image_template.go)_

### ClusterConfigTemplate

Instructs the supply chain how to instantiate a Kubernetes object that knows how to make Kubernetes configurations available to further components in the chain.

For instance, a resource that given an image, exposes a complete `podTemplateSpec` to be embedded in a knative service.

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterConfigTemplate
metadata:
  name: convention-service-battery
spec:
  # default parameters. see ClusterSourceTemplate for more info. (optional)
  #
  params: []

  # jsonpath expression to instruct where in the object templated out an 
  # a kubernetes configuration can be found (required)
  #
  configPath: .status.template

  # how to template out the kubernetes object. (required)
  #
  template:
    apiVersion: opinions.local/v1alpha1
    kind: WorkloadDecorator
    metadata:
      name: $(workload.metadata.name)$-workload-template
    spec:
      template:
        metadata:
          labels:
            app.kubernetes.io/part-of: $(workload.metadata.name)$
          annotations:
            autoscaling.knative.dev/minScale: "1"
            autoscaling.knative.dev/maxScale: "1"
        spec:
          containers:
            - name: workload
              image: $(images[?(@.name=="solo-image-provider")].image)$
              env: $(workload.spec.env)$
              resources: $(workload.spec.resources)$
              securityContext:
                runAsUser: 1000
          imagePullSecrets:
            - name: registry-credentials
```

_ref: [pkg/apis/v1alpha1/cluster_config_template.go](https://github.com/vmware-tanzu/cartographer/blob/v0.0.3/pkg/apis/v1alpha1/cluster_config_template.go)_


### ClusterTemplate

A ClusterTemplate instructs the supply chain to instantiate a Kubernetes object that has no outputs to be supplied to other objects in the chain, for instance, a resource that deploys a container image that has been built by other ancestor components.


```yaml
apiVersion: carto.run/v1alpha1
kind: ConfigTemplate
metadata:
  name: deployer
spec:
  # default parameters. see ClusterSourceTemplate for more info. (optional)
  #
  params: []

  # how to template out the kubernetes object. (required)
  #
  template:
    apiVersion: kappctrl.k14s.io/v1alpha1
    kind: App
    metadata:
      name: $(workload.metadata.name)
    spec:
      serviceAccountName: service-account
      fetch:
        - inline:
            paths:
              manifest.yml: |
                ---
                apiVersion: kapp.k14s.io/v1alpha1
                kind: Config
                rebaseRules:
                  - path: [metadata, annotations, serving.knative.dev/creator]
                    type: copy
                    sources: [new, existing]
                    resourceMatchers: &matchers
                      - apiVersionKindMatcher: {apiVersion: serving.knative.dev/v1, kind: Service}
                  - path: [metadata, annotations, serving.knative.dev/lastModifier]
                    type: copy
                    sources: [new, existing]
                    resourceMatchers: *matchers
                ---
                apiVersion: serving.knative.dev/v1
                kind: Service
                metadata:
                  name: links
                  labels:
                    app.kubernetes.io/part-of: $(workload.metadata.labels['app\.kubernetes\.io/part-of'])$
                spec:
                  template:
                    spec:
                      containers:
                        - image: $(images[0].image)$
                          securityContext:
                            runAsUser: 1000
      template:
        - ytt: {}
      deploy:
        - kapp: {}
```

_ref: [pkg/apis/v1alpha1/cluster_template.go](https://github.com/vmware-tanzu/cartographer/blob/v0.0.3/pkg/apis/v1alpha1/cluster_template.go)_
