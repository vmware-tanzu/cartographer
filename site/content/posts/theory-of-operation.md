---
title: "Cartographer: Theory of Operation"
slug: theory-of-operation
date: 2022-02-28
author: Rasheed Abdul-Aziz
authorLocation: https://github.com/squeedee
image: /img/posts/theory-of-operation/cover-image.png
excerpt: "Theory Of Operation"
tags: ['faq', 'theory']
---

# Assumptions

To explain Cartographer's operation, we're going to make some assumptions:

1. You are familiar with creating and modifying Kubernetes resources
2. You know the difference between a job and a pod.

We're going to focus on Supply Chains and Runnable in this post. Delivery/Deliverable are omitted for clarity.

# Document conventions

| Tables                                                     |
|------------------------------------------------------------|
| contain core definitions of Cartographer's behavior |

# Long Running Resources

When you create a Pod, that is typically a container that runs an executable indefinitely, or at least tries to keep it running.
The purpose of a long running resource, like a pod, is to accept the `spec` supplied and try to represent that specification
as a service. The service is not meant to terminate, it's meant to continually accept requests.

Similarly, many custom operators in the Kubernetes ecosystem are designed as long-running resources. You create a resource
and the resource runs as a service. The noticeable difference between a network service like a Pod, is that these operators
communicate through their `spec` and `status` fields, not incoming requests. 

For example, 
the [Flux CD GitRepository Resource](https://fluxcd.io/docs/components/source/gitrepositories/) accepts the repository
location in `spec.url` and tells you where the internally hosted URL is at `status.url`.

| Important Definitions                                                                     |
|-------------------------------------------------------------------------------------------|
| A changing value (usually in `status`) is an **output**.                                  |
| The **output** is a **level trigger**. We'll explain shortly.                             |
| **Inputs** to a resource come in the form of **fields in the spec**                       |  
| **Inputs** can change, causing an eventual **change in the output**                       |
| **Inputs** can refer to **external resources** that also cause a **change in the output** |

With regard to **external resources**:
* The input might remain unchanged, but the resource it refers to can change. For example, a reference to a git repository
  branch might not change, however there might be a new commit. In the case of the Flux CD GitRepository resource, this will
  eventually trigger a **change in the output**

# Level Triggers
Why are we so focused on **change in the outputs**? Well, remember this is a long-running resource, so to observe
that there is new outputs, we look at these **changing outputs** instead of expecting triggering events.

Level triggering is core to the concept of reconciling state in Kubernetes. [Learn more](https://hackernoon.com/level-triggering-and-reconciliation-in-kubernetes-1f17fe30333d)

# Cartographer Templates

To introduce a long-running Kubernetes resource into a supply chain, its definition must first be wrapped in a Cartographer Template.

| Definitions       |                                                                                                                                                                     |
|-------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Inputs            | Templates use **outputs from other resources** as part of the specification of your resource, without needing to know anything about the other resources.           |
| Output            | Templates expose a **single output** to other resources, such that other resources do not need to know how your resource works.                                     |
| Outputs are Typed | The template `kind` specifies the **kind of output** your resource produces. See the [table describing templates in the docs](/docs/v0.2.0/architecture/#templates) |

The template standardizes the resource so that you can use it in place of another resource, anywhere this **kind** of resource
is used in a Supply Chain.

Other than referencing inputs in your template, you can also refer to any field(s) in the Workload, and specify parameter's
(with sane defaults if they're optional).

The template can accept more than one input, although it's not that common.

Let's look at a typical template:

{{< figure src="/img/posts/theory-of-operation/typical-template.svg" alt="Typical Template" width="800px" >}}



# Terminating (Run-and-Done) Resources

## Runnable and ClusterRunTemplate
