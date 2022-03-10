# Multi-Cluster

## Overview

It is often desirable to separate build and production environments; this can be achieved with multiple clusters.
Cartographer helps pave a path to production across clusters with
[ClusterSupplyChains](reference/workload.md#clustersupplychain) and
[ClusterDeliveries](reference/deliverable.md#clusterdelivery), leveraging [GitOps](https://www.gitops.tech/).

A ClusterSupplyChain can be installed into an isolated build cluster to manage the path from source code to Kubernetes
configuration in a Git repository. Meanwhile, a ClusterDelivery can be installed in the production cluster, which will
pick up that configuration from the Git repository to be deployed and tested. This allows for the same artifact to be
promoted through multiple environments to first test/validate and finally run in production.

## Examples

### Example PR based flow

0. Developer commits new source code
1. ClusterSupplyChain in the build cluster generates Kubernetes configuration and commits it to a git repository (i.e.
   staging/feature-a)
2. "Staging environment maintainer" opens a PR and merges staging/feature-a into staging
3. ClusterDelivery in the staging cluster picks up the merged PR and deploys the Kubernetes configuration to be tested
   and commits it to a git repository (i.e. production/feature-a)
4. "Production environment maintainer" opens a PR and merges production/feature-a into production
5. ClusterDelivery in the production cluster picks up the merged PR and deploys the Kubernetes configuration.

![Multi-Cluster](../img/multi-cluster.jpg)

### Basic example without PRs

- [Supply Chain: Source ➡️ Image ➡️ Git](https://github.com/vmware-tanzu/cartographer/tree/main/examples/gitwriter-sc/README.md)
  - This simple ClusterSupplyChain takes source code, builds an image, generates app configuration, and writes it to a
    Git repository.
- [Delivery: Git ➡️ App](https://github.com/vmware-tanzu/cartographer/tree/main/examples/basic-delivery/README.md)
  - This simple ClusterDelivery picks up app configuration from a Git repository and deploys to the cluster.
