# Architecture and Concepts

## Overview

Cartographer is an open-sourced Supply Chain Choreographer for Kubernetes. Cartographer provides a set of Kubernetes
controllers and CRDs that allow a platform operator to create an application platform by specifying Supply Chains and 
application Delivery workflows.

## Concepts

### Blueprints

Blueprints are a list of templates (called resources) and how they depend upon each other. It forms the dependency graph of
your supply chain or delivery.

The dependencies are formed by specifying which resource(s) are used as inputs.

Blueprints consist of:
* A **selector** to match owners
* **Parameters** to pass to all resources 
* **Resources**: 
  * A **templateRef** pointing to the template for the resource
  * **Parameters** to pass to the template
  * **Inputs**, which specify dependencies for the template

![Blueprint](../img/blueprint.jpg)
<!-- https://miro.com/app/board/uXjVOeb8u5o=/ -->

| Blueprint    | Owner | Valid Templates |
| ----------- | ----------- | ----------- |
| ClusterSupplyChain | Workload | ClusterSourceTemplate, ClusterImageTemplate, ClusterTemplate, ClusterConfigTemplate |
| ClusterDelivery | Deliverable | ClusterSourceTemplate, ClusterDeploymentTemplate, ClusterTemplate |

### Templates

Templates create or update resources (i.e. kubectl apply).

Templates consist of:
* Parameters to pass to `spec.template` or `spec.ytt`
* The Kubernetes resource yaml as `spec.template` or `spec.ytt` [see Templating](#tbd)
* **Output paths** which tell Cartographer where to find the output of the Kubernetes resource
  * The path field depends upon the specific template kind.

![Template](../img/template.jpg)

Templates are typed by the output they produce.

| Output      | Template |
| ----------- | ----------- |
| Config | ClusterConfigTemplate |
| Image | ClusterImageTemplate |
| Source | ClusterSourceTemplate |
| Deployment | ClusterDeploymentTemplate |
| | ClusterTemplate |

### Owners

Owners represent the workload or deliverable, which in many cases refer to a single application's source or image 
location.

Owners are the developer provided configuration which cause a blueprint to be reconciled into resources.

They consist of:
* **Labels**: blueprints will select based on the labels of an owner, see [selectors](#selectors) 
* **Params**: parameters supplied to the blueprint, see [Parameter Hierarchy](#parameter-hierarchy)
* **Source**: The source reference for the input to the Supply Chain or Delivery Blueprints,
see [Workload](reference.md/#workload) and [Deliverable](reference.md/#deliverable)

![Owner](../img/owner.jpg)

| Owner      | Blueprint |
| ----------- | ----------- |
| Workload | ClusterSupplyChain |
| Deliverable | ClusterDelivery |

## Theory of operation

Given an owner that matches a blueprint, cartographer reconciles the resources referenced by the blueprint.
The resources are only created when the inputs are satisfied, and a resource is only updated when it's inputs change.
This results in a system where a new intrinsic result from one resource can cause other resources to change.

Although Cartographer is not a 'runner of things', a resource can be something as simple as a Job or a CI pipeline.
However, one advantage of Cartographer's design, is that a resource can also be untriggered. Imagine a Build resource 
that discovers new base OCI images. If it rebuilds your image, then cartographer will see this new image and update 
other linked resources.

![Generic Blueprint](../img/generic.jpg)
<!-- https://miro.com/app/board/uXjVOeb8u5o=/ -->

When Cartographer reconciles an owner, each resource in the matching blueprint is reconciled:

1. **Generate Inputs**: Using the **blueprint resource's** `inputs` as a reference, select outputs from previously applied **Kubernetes resources**
2. **Generate Params**: Using the [Parameter Hierarchy](architecture.md#parameter-hierarchy), generate parameter values   
3. **Generate and apply resource spec**: Apply the result of interpolating `spec.template` (or `spec.ytt`), **inputs**, **params** and the **owner spec**. 
4. **Retrieve Output**: Store the output from the applied resource. The output to use is specified in the **template output path**.  

![Realize](../img/realize.jpg)

### Complete Supply Chain and Delivery with GitOps

![Gitops](../img/gitops.jpg)

## Blueprint Details

### ClusterSupplyChain
A ClusterSupplyChain blueprint continuously integrates and builds your app.

![ClusterSupplyChain](../img/supplychain.jpg)

### ClusterDelivery
A ClusterDelivery blueprint continuously deploys and validates images to a cluster. A ClusterDelivery has the ability to lock 
(and unlock) templates which pauses the continuous deploy. 

<!--- @TODO MORE ON LOCKING -->

![ClusterDelivery](../img/delivery.jpg)

### Selectors
An owner's labels will determine which blueprint will select for it. The controller will do a "best match" on a blueprint's 
`spec.selector` with an owner's labels.

A "best match" follows the rules:
1. If all labels are fully contained in the selector, reconcile the owner with that blueprint
2. If more than one blueprint has all the labels that the owner has, pick the most identical to the owner
3. If multiple blueprints match the owner labels, reconcile with the blueprint with the most label matches

Note:  Despite the rules, the controller can still return more than one match. If more than one match is returned, 
no blueprint will reconcile for the owner.

## Parameter Hierarchy

<!--- @TODO Image of params -->


Templates specify the **parameters** they accept in `spec.params`. These can have a default value.

These parameters can be fulfilled by the **blueprint**, which allows operators to specify:
* a default value which can be overridden by the **owner's** `spec.params`
* a value which cannot be overridden by the **owner**

Blueprint parameters can be specified globally in `spec.params` or per resource `spec.resource[].params`
If the **per resource param** is specified, the global blueprint param is ignored.
