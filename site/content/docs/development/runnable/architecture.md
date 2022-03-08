# Architecture and Concepts

## Overview

Runnable is a component of Cartographer. The `Runnable` CRD provides an intermediate layer to update immutable
resources. For example, Tekton does not allow updating TaskRuns and PipelineRuns, so we'll update Runnable to test new
commits.

## Concepts

[comment]: <> (### Edge vs level-driven triggers) [comment]: <> (TODO)

### ClusterRunTemplate

With the addition of Runnable, there is a new template, `ClusterRunTemplate`. A `ClusterRunTemplate` will _always_
create resources (i.e. `kubectl create`).

ClusterRunTemplate consists of:

- The Kubernetes resource yaml as `spec.template`
- **Output paths** which tell Cartographer where to find the output of the Kubernetes resource
  - The outputs will be added to `runnable.status.outputs`

{{< figure src="../../img/runnable/clusterruntemplate.svg" alt="Template" width="400px" >}}

### Runnable

Runnables consist of:

- **RunTemplateRef**: a reference to a `ClusterRunTemplate` which contains the yaml of the immutable resource to be
  created
- **Selector**: used to dynamically discover a resource that is needed in the `ClusterRunTemplate`. The matching
  resource is available in the template data as `selected`.
- **Inputs**: arbitrary key, value pairs that are passed along to the `ClusterRunTemplate`.

{{< figure src="../../img/runnable/runnable-outline.svg" alt="Runnable" width="400px" >}}

### Template Data

See [Template Data](../../templating#template-data) for templating in Cartographer.

The ClusterRunTemplate is provided a data structure that contains:

- runnable
- selected

#### Runnable

The entire Runnable resource is available for retrieving values. To use a Runnable value, use the format:

- **Simple template**: `$(runnable.<field-name>.(...))$`
- **ytt**: not currently supported, see [issue](https://github.com/vmware-tanzu/cartographer/issues/214)

**Runnable Examples**

| Simple template              | ytt |
| ---------------------------- | --- |
| `$(runnable.metadata.name)$` | N/A |
| `$(runnable.spec.inputs)$`   | N/A |

#### Selected

The entire selected resource is available for retrieving values. To use selected value, use the format:

- **Simple template**: `$(selected.<field-name>.(...))$`
- **ytt**: not currently supported, see [issue](https://github.com/vmware-tanzu/cartographer/issues/214)

**Selected Examples**

| Simple template              | ytt |
| ---------------------------- | --- |
| `$(selected.metadata.name)$` | N/A |

## Theory of Operation

When Cartographer reconciles a Runnable, the resource in the specified `ClusterRunTemplate` is applied:

1. **Resolve Selector**: attempt to find a resource that matches the selector
2. **Generate and apply resource spec**: Apply the result of interpolating `spec.template` in the `ClusterRunTemplate`,
   **selected**, and the **runnable spec**.
3. **Retrieve Output**: The output to use is specified in the **template output path**
   1. Get the output from the most recently created resource, where
      `status.conditions[?(@.type=="Succeeded")].status == True`.
   2. Store the output in `runnable.status.outputs`.

![Realize](../../img/runnable/runnable-realize-new.jpg)

<!-- https://miro.com/app/board/uXjVOeb8u5o=/ -->

## Runnable In Action

![Runnable](../../img/runnable/runnable-new.jpg)

### Using Runnable with a Supply Chain

![Runnable-SupplyChain](../../img/runnable/runnable-supplychain-new.jpg)

To see an example of the rest of the supply chain, see [ClusterSupplyChain](../architecture#clustersupplychain).
