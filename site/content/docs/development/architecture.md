# Architecture and Concepts

## Overview

Cartographer is an open-sourced Supply Chain Choreographer for Kubernetes. Cartographer provides a set of Kubernetes
controllers and CRDs that allow a platform operator to create an application platform by specifying Supply Chains and 
application Delivery workflows.

## Concepts

### Templates

Templates create or update resources (i.e. kubectl apply).

Templates consist of:
* Inputs
    * Outputs of other resources
    * Params from Owners and Blueprints
    * Any field in the Owner - simplifies extension
* Resource, the actual yaml to apply to kubernetes
* Output, a single typed object from the stamped resource
    * Config
    * Image
    * Source
    * Deployment


### Blueprints

Blueprints is a list of templates and how they depend upon each other. It forms the dependency graph of
your supply chain or delivery.

### Owners

Owners represent the workload or deliverable, which in many cases refer to a single application's source or image 
location.

### Theory of operation

Given an owner that matches a blueprint, cartographer reconciles the resources referenced by the blueprint.
The resources are only created when the inputs are satisfied, and a resource is only updated when it's inputs change.
This results in a system where a new intrinsic result from one resource can cause other resources to change.

Although Cartographer is not a 'runner of things', a resource can be something as simple as a Job or a CI pipeline.
However, one advantage of Cartographer's design, is that a resource can also be untriggered. Imagine a Build resource 
that discovers new base OCI images. If it rebuilds your image, then cartographer will see this new image and update 
other linked resources.

![Generic Blueprint](../img/generic.jpg)
<!-- https://miro.com/app/board/uXjVOeb8u5o=/ -->

### Reconciles blueprint
When Cartographer reconciles an owner, each resource in the matching blueprint is reconciled:

1. Generate Inputs: Using the **blueprint resource `inputs` as a reference, select outputs from previously applied **Kubernetes Resources**
2. Generate Params: Using the [Parameter Hierarchy](architecture.md#parameter-hierarchy), generate parameter values   
3. Generate and apply resource spec: Apply the result of interpolating `spec.Template` (or `ytt`), inputs, params and owner spec. 
4. Retrieve Output: Store the output from the applied resource. The output to use is specified in the **Template Output Path**.  

<!-- new diagram https://miro.com/app/board/uXjVOeb8u5o=/?moveToWidget=3458764514330138805&cot=14 -->

![Realize](../img/realize.jpg)
<!-- https://miro.com/app/board/uXjVOeb8u5o=/ -->


### Types of templates

Templates are typed by the output they produce.

| Output      | Template |
| ----------- | ----------- |
| Config | ClusterConfigTemplate |
| Image | ClusterImageTemplate |
| Source | ClusterSourceTemplate |
| Deployment | ClusterDeploymentTemplate |
| | ClusterTemplate |

### Types of blueprints

<!-- insert image of simplified supply chain into delivery -->

| Blueprint    | Owner | Valid Templates |
| ----------- | ----------- | ----------- |
| ClusterSupplyChain | Workload | ClusterSourceTemplate, ClusterImageTemplate, ClusterTemplate, ClusterConfigTemplate |
| ClusterDelivery | Deliverable | ClusterSourceTemplate, ClusterDeploymentTemplate, ClusterTemplate |

#### ClusterSupplyChain
is a blueprint which continuously integrates and builds your app.

#### ClusterDelivery
is a blueprint which continuously deploys and validates images to a cluster. This blueprint has the ability to lock 
(and unlock) templates which pauses the continuous deploy.

#### Selectors
An owner's labels will determine which blueprint will select for it. A blueprint's `spec.selector` will match on the 
owner's labels.

### Types of Inputs

### Parameter Hierarchy

### Complete Supply Chain and Delivery Example

<!-- insert very specific diagram with logos -->