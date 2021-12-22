# Architecture and Concepts

## Overview

Runnable is an open source component of Cartographer. The `Runnable` CRD enables updating what are normally
immutable test resources. Tekton does not allow updating an object, and so we'll update Runnable to
test new commits.

## Concepts
### Event vs level trigger

### ClusterRunTemplate
With the addition of Runnable, there is a new template, `ClusterRunTemplate`. A `ClusterRunTemplate` will _always_
create resources (i.e. `kubectl create`).

ClusterRunTemplate consists of:
* The Kubernetes resource yaml as `spec.template`
* **Output paths** which tell Cartographer where to find the output of the Kubernetes resource
  * The outputs will be added to `runnable.status.outputs`

  
{{< figure src="../../img/runnable/clusterruntemplate.svg" alt="Template" width="400px" >}}

### Runnable
They consist of:
* **RunTemplateRef**: a reference to a `ClusterRunTemplate` which contains the yaml of the immutable resource to be
created
* **Selector**: the resource that it matches with is availble in the template data as `selected`
* **Inputs**: triggers

{{< figure src="../../img/runnable/runnable-outline.svg" alt="Runnable" width="400px" >}}

### Template Data
See [Template Data](templating#template-data)

The ClusterRunTemplate is provided a data structure that contains:
- runnable
- selected


## Theory of Operation

When Cartographer reconciles a runnable, each resource in the matching blueprint is applied:

1. **Resolve Selector**: Using the **blueprint resource's** `inputs` as a reference, select outputs from previously applied **Kubernetes resources**
2. **Generate and apply resource spec**: Apply the result of interpolating `spec.template`, **selected** and the **runnable spec**. 
3. **Retrieve Output**: 
   1. Get the output from the most recently created resource, where `status.conditions[?(@.type=="Succeeded")].status == True`.
   2. The output to use is specified in the **template output path**. 
   3. Store the output in `runnable.status.outputs`.

![Realize](../../img/runnable/runnable-realize-new.jpg)
<!-- https://miro.com/app/board/uXjVOeb8u5o=/ -->


## Runnable Details
### IDK
The `Runnable` CRD enables updating what are normally immutable test resources.

![Runnable](../../img/runnable/runnable-new.jpg)

### Using Runnable with a Supply Chain

![Runnable-SupplyChain](../../img/runnable/runnable-supplychain-new.jpg)

To see an example of the rest of the supply chain, see [ClusterSupplyChain](../architecture#clustersupplychain).
