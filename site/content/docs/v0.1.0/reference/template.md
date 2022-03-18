# Template Custom Resources

## ClusterSourceTemplate

`ClusterSourceTemplate` indicates how the supply chain could instantiate an object responsible for providing source
code.

The `ClusterSourceTemplate` requires definition of a `urlPath` and `revisionPath`. `ClusterSourceTemplate` will update
its status to emit `url` and `revision` values, which are reflections of the values at the path on the created objects.
The supply chain may make these values available to other resources.

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
      # default value if not specified in the resource that references
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
      gitImplementation: $(params.git-implementation.value)$
      ignore: ""
```

_ref:
[pkg/apis/v1alpha1/cluster_source_template.go](https://github.com/vmware-tanzu/cartographer/tree/main/pkg/apis/v1alpha1/cluster_source_template.go)_

## ClusterImageTemplate

`ClusterImageTemplate` instructs how the supply chain should instantiate an object responsible for supplying container
images, for instance, one that takes source code, builds a container image out of it.

The `ClusterImageTemplate` requires definition of an `imagePath`. `ClusterImageTemplate` will update its status to emit
an `image` value, which is a reflection of the value at the path on the created object. The supply chain may make this
value available to other resources.

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
    apiVersion: kpack.io/v1alpha2
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
          url: $(sources.provider.url)$
```

_ref:
[pkg/apis/v1alpha1/cluster_image_template.go](https://github.com/vmware-tanzu/cartographer/tree/main/pkg/apis/v1alpha1/cluster_image_template.go)_

## ClusterConfigTemplate

Instructs the supply chain how to instantiate a Kubernetes object that knows how to make Kubernetes configurations
available to further resources in the chain.

The `ClusterConfigTemplate` requires definition of a `configPath`. `ClusterConfigTemplate` will update its status to
emit a `config` value, which is a reflection of the value at the path on the created object. The supply chain may make
this value available to other resources.

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterConfigTemplate
metadata:
  name: deployer
spec:
  # default parameters. see ClusterSourceTemplate for more info. (optional)
  #
  params: []

  # jsonpath expression to instruct where in the object templated out config
  # information can be found. (required)
  #
  configPath: .data

  # how to template out the kubernetes object. (required)
  #
  template:
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: $(workload.metadata.name)$
    data:
      service.yml: |
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
                - image: $(images.<name-of-image-provider>.image)$
                  securityContext:
                    runAsUser: 1000
```

_ref:
[pkg/apis/v1alpha1/cluster_config_template.go](https://github.com/vmware-tanzu/cartographer/tree/main/pkg/apis/v1alpha1/cluster_config_template.go)_

## ClusterDeploymentTemplate

A `ClusterDeploymentTemplate` indicates how the delivery should configure the environment (namespace/cluster).

The `ClusterDeploymentTemplate` consumes configuration from the `deployment` values provided by the `ClusterDelivery`.
The `ClusterDeploymentTemplate` outputs these same values. The `ClusterDeploymentTemplate` is able to consume additional
configuration from the `sources` provided by the `ClusterDelivery`.

`ClusterDeploymentTemplate` must specify criteria to determine whether the templated object has successfully completed
its role in configuring the environment. Once the criteria are met, the `ClusterDeploymentTemplate` will output the
`deployment` values. The criteria may be specified in `spec.observedMatches` or in `spec.observedCompletion`.

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterDeploymentTemplate
metadata:
  name: app-deploy---deliverable
spec:
  # criteria for determining templated object has completed configuration of environment.
  # (mutually exclusive with observedCompletion; one or the other required)
  observedMatches:
    # set of input:output pairs
    # when the value of input == output for each set, the criteria has been satisfied
    #
    # (one or more required)
    - input: "spec.value.some-key"
      output: "status.some-key"
      # input is expected to be some field specified before the templated object is reconciled
    - input: "spec.value.another-key"
      # output is expected to be some field set after reconciliation (e.g. in the status)
      output: "status.another-key"

  # criteria for determining templated object has completed configuration of environment.
  # this criteria requires that the templated object reports `status.observedGeneration`
  #
  # if templated object's:
  # 1. `status.observedGeneration` == `metadata.generation`
  # 2.  the field at the specified succeeded key == the specified value
  # then the criteria has been satisfied
  #
  # if the templated object's:
  # 1. `status.observedGeneration` == `metadata.generation`
  # 2.  the field at the specified failed key == the specified value
  # then the criteria cannot be met
  #
  # (mutually exclusive with observedMatches; one or the other required)
  observedCompletion:
    # (required)
    succeeded:
      # field to inspect on the templated object
      # (required)
      key: 'status.conditions[?(@.type=="Succeeded")].status'
      # value to expect at the inspected field
      # (required)
      value: "True"
    # (optional)
    failed:
      # (required)
      key: 'status.conditions[?(@.type=="Failed")].status'
      # (required)
      value: "True"
  # template for configuring the environment/deploying an application
  template:
    apiVersion: kappctrl.k14s.io/v1alpha1
    kind: App
    metadata:
      name: $(deliverable.metadata.name)$
    spec:
      serviceAccountName: default
      fetch:
        - http:
            url: $(source.url)$
      template:
        - ytt: {}
      deploy:
        - kapp: {}
```

_ref:
[pkg/apis/v1alpha1/cluster_deployment_template.go](https://github.com/vmware-tanzu/cartographer/tree/main/pkg/apis/v1alpha1/cluster_deployment_template.go)_

## ClusterTemplate

A `ClusterTemplate` instructs the supply chain to instantiate a Kubernetes object that has no outputs to be supplied to
other objects in the chain, for instance, a resource that deploys a container image that has been built by other
ancestor resources.

The `ClusterTemplate` does not emit values to the supply chain.

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterTemplate
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
      name: $(workload.metadata.name)$
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
                        - image: $(images.<name-of-image-provider>.image)$
                          securityContext:
                            runAsUser: 1000
      template:
        - ytt: {}
      deploy:
        - kapp: {}
```

_ref:
[pkg/apis/v1alpha1/cluster_template.go](https://github.com/vmware-tanzu/cartographer/tree/main/pkg/apis/v1alpha1/cluster_template.go)_
