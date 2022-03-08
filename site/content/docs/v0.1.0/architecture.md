# Architecture and Concepts

## Overview

Cartographer is an open-source Supply Chain Choreographer for Kubernetes. Cartographer provides a set of Kubernetes
controllers and CRDs that allow a platform operator to create an application platform by specifying repeatable, reusable
**code-to-production** blueprints.

Two kinds of blueprint work together to provide **code-to-production**, [Supply Chains](#clustersupplychain) and
[Delivery](#clusterdelivery).

## Concepts

### Blueprints

| Blueprint                                                   | Owner                                            | Valid Templates                                                                                                                                                                                                                                              |
| ----------------------------------------------------------- | ------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [ClusterSupplyChain](reference/workload#clustersupplychain) | [Workload](reference/workload#workload)          | [ClusterSourceTemplate](reference/template#clustersourcetemplate), [ClusterImageTemplate](reference/template#clusterimagetemplate), [ClusterConfigTemplate](reference/template#clusterconfigtemplate), [ClusterTemplate](reference/template#clustertemplate) |
| [ClusterDelivery](reference/deliverable#clusterdelivery)    | [Deliverable](reference/deliverable#deliverable) | [ClusterSourceTemplate](reference/template#clustersourcetemplate), [ClusterDeploymentTemplate](reference/template#clusterdeploymenttemplate), [ClusterTemplate](reference/template#clustertemplate)                                                          |

Blueprints are a list of templates (called resources) that defines how the templates depend upon each other. It forms
the dependency graph of your supply chain or delivery.

The dependencies are formed by specifying which resource(s) are used as inputs.

Blueprints consist of:

- A **selector** to match owners
- **Parameters** to pass to all resources
- **Resources**:
  - A **templateRef** pointing to the template for the resource
  - **Parameters** to pass to the template
  - **Inputs**, which specify dependencies for the template

{{< figure src="../img/blueprint.svg" alt="Blueprint" width="400px" >}}

<!-- https://miro.com/app/board/uXjVOeb8u5o=/ -->

### Templates

Templates create or update resources (i.e. kubectl apply).

Templates consist of:

- Parameters to pass to `spec.template` or `spec.ytt`
- The Kubernetes resource yaml as `spec.template` or `spec.ytt` see [Templating](templating#templating)
- **Output paths** which tell Cartographer where to find the output of the Kubernetes resource
  - The path field depends upon the specific template kind.
  - These paths are interpolated and subsequent templates can use them via the input accessors. see
    [Inputs](templating#inputs)

Templates are typed by the output their underlying resource produces.

| Output     | Template                                                                  | Output Path                         | Input Accessor                                              |
| ---------- | ------------------------------------------------------------------------- | ----------------------------------- | ----------------------------------------------------------- |
| Config     | [ClusterConfigTemplate](reference/template#clusterconfigtemplate)         | `spec.configPath`                   | `configs.<input-name>`                                      |
| Image      | [ClusterImageTemplate](reference/template#clusterimagetemplate)           | `spec.imagePath`                    | `images.<input-name>`                                       |
| Source     | [ClusterSourceTemplate](reference/template#clustersourcetemplate)         | `spec.urlPath`, `spec.revisionPath` | `sources.<input-name>.url`, `sources.<input-name>.revision` |
| Deployment | [ClusterDeploymentTemplate](reference/template#clusterdeploymenttemplate) | `spec.urlPath`, `spec.revisionPath` | `sources.<input-name>.url`, `sources.<input-name>.revision` |
|            | [ClusterTemplate](reference/template#clustertemplate)                     |

{{< figure src="../img/template.svg" alt="Template" width="400px" >}}

### Owners

| Owner       | Blueprint          |
| ----------- | ------------------ |
| Workload    | ClusterSupplyChain |
| Deliverable | ClusterDelivery    |

Owners represent the **workload** or **deliverable**, which in many cases refer to a single application's source or
image location.

Owners are the developer provided configuration which cause a blueprint to be reconciled into resources. Owners
reference the primary **source** or **image** for the **blueprint**

They consist of:

- **Labels**: blueprints will select based on the labels of an owner, see [selectors](#selectors)
- **Params**: parameters supplied to the blueprint, see [Parameter Hierarchy](#parameter-hierarchy)
- **Source**: The source reference for the input to the Supply Chain or Delivery Blueprints, see
  [Workload](reference/workload#workload) and [Deliverable](reference/deliverable#deliverable)

{{< figure src="../img/owner.svg" alt="Owner" width="400px" >}}

## Theory of Operation

Given an owner that matches a blueprint, Cartographer reconciles the resources referenced by the blueprint. The
resources are only created when the inputs are satisfied, and a resource is only updated when its inputs change. This
results in a system where a new result from one resource can cause other resources to change.

![Generic Blueprint](../img/generic.jpg)

<!-- https://miro.com/app/board/uXjVOeb8u5o=/ -->

Although Cartographer is not a 'runner of things', a resource can be something as simple as a Job.

However, one advantage of Cartographer's design is that resources that self-mutate can cause downstream change.

For example, a Build resource that discovers new base OCI images. If it rebuilds your image, then Cartographer will see
this new image and update downstream resources.

When Cartographer reconciles an owner, each resource in the matching blueprint is applied:

1. **Generate Inputs**: Using the **blueprint resource's** `inputs` as a reference, select outputs from previously
   applied **Kubernetes resources**
2. **Generate Params**: Using the [Parameter Hierarchy](architecture.md#parameter-hierarchy), generate parameter values
3. **Generate and apply resource spec**: Apply the result of interpolating `spec.template` (or `spec.ytt`), **inputs**,
   **params** and the **owner spec**.
4. **Retrieve Output**: Store the output from the applied resource. The output to use is specified in the **template
   output path**.

![Realize](../img/realize.jpg)

### Complete Supply Chain and Delivery with GitOps

![Gitops](../img/gitops.jpg)

## Blueprint Details

### ClusterSupplyChain

A ClusterSupplyChain blueprint continuously integrates and builds your app.

![ClusterSupplyChain](../img/supplychain.png)

### ClusterDelivery

A ClusterDelivery blueprint continuously deploys and validates Kubernetes configuration to a cluster.

[comment]: <> (Not implemented yet) [comment]: <> (A ClusterDelivery has the ability to lock &#40;and unlock&#41;
templates which pauses the continuous deploy. ) [comment]: <> (TODO - more on locking)

![ClusterDelivery](../img/delivery.jpg)

### Selectors

An owner's labels will determine which blueprint will select for it. The controller will do a "best match" on a
blueprint's `spec.selector` with an owner's labels.

A "best match" follows the rules:

1. If all labels are fully contained in the selector, reconcile the owner with that blueprint.
2. If not all labels match, we choose the blueprint with the most matched labels.
3. If two blueprints match all labels, the blueprint with the more concise match (less non-matching labels) is selected.

Note: Despite the rules, the controller can still return more than one match. If more than one match is returned, no
blueprint will reconcile for the owner.

## Parameter Hierarchy

<!--- @TODO Image of params -->

Templates can specify default values for **parameters** in `spec.params`.

These parameters may be overridden by the **blueprint**, which allows operators to specify:

- a default value which can be overridden by the **owner's** `spec.params`
- a value which cannot be overridden by the **owner**

Blueprint parameters can be specified globally in `spec.params` or per resource `spec.resource[].params` If the **per
resource param** is specified, the global blueprint param is ignored.
