# Template Custom Resources

## ClusterSourceTemplate

`ClusterSourceTemplate` indicates how the supply chain could instantiate an object responsible for providing source
code.

The `ClusterSourceTemplate` requires definition of a `urlPath` and `revisionPath`. `ClusterSourceTemplate` will update
its status to emit `url` and `revision` values, which are reflections of the values at the path on the created objects.
The supply chain may make these values available to other resources.

{{< crd  carto.run_clustersourcetemplates.yaml >}}

_ref:
[pkg/apis/v1alpha1/cluster_source_template.go](https://github.com/vmware-tanzu/cartographer/tree/main/pkg/apis/v1alpha1/cluster_source_template.go)_

## ClusterImageTemplate

`ClusterImageTemplate` instructs how the supply chain should instantiate an object responsible for supplying container
images, for instance, one that takes source code, builds a container image out of it.

The `ClusterImageTemplate` requires definition of an `imagePath`. `ClusterImageTemplate` will update its status to emit
an `image` value, which is a reflection of the value at the path on the created object. The supply chain may make this
value available to other resources.

{{< crd  carto.run_clusterimagetemplates.yaml >}}

_ref:
[pkg/apis/v1alpha1/cluster_image_template.go](https://github.com/vmware-tanzu/cartographer/tree/main/pkg/apis/v1alpha1/cluster_image_template.go)_

## ClusterConfigTemplate

Instructs the supply chain how to instantiate a Kubernetes object that knows how to make Kubernetes configurations
available to further resources in the chain.

The `ClusterConfigTemplate` requires definition of a `configPath`. `ClusterConfigTemplate` will update its status to
emit a `config` value, which is a reflection of the value at the path on the created object. The supply chain may make
this value available to other resources.

{{< crd  carto.run_clusterconfigtemplates.yaml >}}

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

{{< crd  carto.run_clusterdeploymenttemplates.yaml >}}

_ref:
[pkg/apis/v1alpha1/cluster_deployment_template.go](https://github.com/vmware-tanzu/cartographer/tree/main/pkg/apis/v1alpha1/cluster_deployment_template.go)_

## ClusterTemplate

A `ClusterTemplate` instructs the supply chain to instantiate a Kubernetes object that has no outputs to be supplied to
other objects in the chain, for instance, a resource that deploys a container image that has been built by other
ancestor resources.

The `ClusterTemplate` does not emit values to the supply chain.

{{< crd  carto.run_clustertemplates.yaml >}}

_ref:
[pkg/apis/v1alpha1/cluster_template.go](https://github.com/vmware-tanzu/cartographer/tree/main/pkg/apis/v1alpha1/cluster_template.go)_
