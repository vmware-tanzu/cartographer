---
title: "Announcing Cartographer"
slug: building-paths-to-prod
date: 2021-10-05
author: Cartographer Team
authorLocation: https://github.com/vmware-tanzu/cartographer/blob/main/MAINTAINERS.md
image: /img/posts/announcing-carto/cover-image.png
excerpt: "Announcing Cartographer a Choreographer for Kubernetes"
tags: ["Path to Production"]
---

A path to production is a way of codifying an application’s value stream map as it makes its way from a developer’s
workstation to a production environment.

![A Simple Path to Production](/img/posts/announcing-carto/path-to-prod.png)

Although Kubernetes has a rich ecosystem of APIs and tools to do the work, managing the communication paths between
these resources can still be a big pain.

This is why we’ve created and open-sourced Cartographer, a tightly scoped tool that can
[_choreograph_](https://tanzu.vmware.com/developer/guides/supply-chain-choreography/) events within the Kubernetes
ecosystem. Instead of building an entirely new platform to hide this complexity and offer convenience, we are building
on top of Kubernetes primitives to enable automation and encourage composability. You’ll find Cartographer on
[GitHub](https://github.com/vmware-tanzu/cartographer).

## What is Cartographer?

Cartographer allows users to configure Kubernetes resources into re-usable supply chains that can be used to define all
of the stages that an Application Workload must go through to get to an environment. We call these stages the path to
production.

Cartographer separates controls between a user responsible for defining a Supply Chain (known as an App Operator) and a
user responsible for creating the application (the Developer). These roles are not necessarily mutually exclusive, but
provide the ability to create a separation concerns which means developers can focus on writing code, and platforms
teams can focus on smoothing and securing their path to production.

### Want to learn more?

- [Getting Started with Cartographer](https://github.com/vmware-tanzu/cartographer#getting-started)
- Your First Supply Chain:
  [Source to Knative Service](https://github.com/vmware-tanzu/cartographer/blob/main/examples/source-to-knative-service/README.md)

## How Does It Work?

Cartographer allows users to define all of the steps that an application must go through to create an image and
Kubernetes configuration. Users achieve this with the
[**_Supply Chain_**](https://cartographer.sh/docs/latest/reference#clustersupplychain) abstraction.

The supply chain consists of resources that are specified via Templates. Each template acts as a wrapper for existing
Kubernetes resources and allows them to be used with Cartographer. Cartographer supply chain templates include:

- [**_Source Template_**](https://cartographer.sh/docs/latest/reference#clustersourcetemplate)
- [**_Image Template_**](https://cartographer.sh/docs/latest/reference#clusterimagetemplate)
- [**_Config Template_**](https://cartographer.sh/docs/latest/reference#clusterconfigtemplate)
- [**_Generic Template_**](https://cartographer.sh/docs/latest/reference#clustertemplate)

Unlike many other existing Kubernetes native workflow tools, Cartographer does not “run” any of the objects themselves.
Instead, it monitors the execution of each resource and templates the following resource in the supply chain after a
given resource has completed execution and updated its status.

The supply chain may also be extended to include integrations to existing CI/CD pipelines by using the Pipeline Service
(which is part of Cartographer Core). The Pipeline Service acts as a wrapper for existing CI and CD tooling (with
support for Tekton, and with plans to support more providers in the future) and provides a declarative way for pipelines
to be run inside of Cartographer.

While the supply chain is operator facing, Cartographer also provides an abstraction for developers called
[**_Workloads_**](https://cartographer.sh/docs/latest/reference#workload). Workloads allow developers to create
application specifications such as the location of their repository, environment variables and service claims.

By design, supply chains can be reused by many workloads. This allows an operator to specify the steps in the path to
production a single time, and for developers to specify their applications independently but for each to use the same
path to production. The intent is that developers are able to focus on providing value for their users and can reach
production quickly and easily, while providing peace of mind for app operators, who are ensured that each application
has passed through the steps of the path to production that they’ve defined.

![Cartographer High Level Diagram](/img/posts/announcing-carto/ownership-flow.png)

## It’s only just begun!

The team here is excited to share Cartographer with the community! We’re looking forward to having you try it out, use
it, provide feedback, and to get involved if you’re interested. Whether it’s by automating the turning of source code
into a running app, or integrating security scans into every application that’s deployed onto your platform, our hope is
to help make Kubernetes app development and operations an absolute breeze. Get Involved

We want Cartographer to be a project that embraces input from the community as it begins to find its place within the
Kubernetes ecosystem. From code contributions and documentation to sharing your usage in the field, there are many ways
to get involved. Try out the [latest release on GitHub](https://github.com/vmware-tanzu/cartographer/releases)!

- [Propose or request new features](https://github.com/vmware-tanzu/cartographer/blob/main/rfc/README.md)
- [Easy first issues tag](https://github.com/vmware-tanzu/cartographer/labels/good%20first%20issue)

## Join the Cartographer Community

- Get updates on Twitter [@OssCartographer](https://twitter.com/OssCartographer)
