# Multi-cluster

## Overview
It is often desirable to separate development and production environments; this can be achieved with multiple clusters.
Cartographer helps pave a path to production across clusters with ClusterSupplyChains and ClusterDeliveries, leveraging
GitOps. 

ClusterSupplyChains can be installed in the development cluster to manage the path from source code to Kubernetes
configuration in a Git repository. ClusterDeliveries can be installed in the production cluster, which then pick up that 
configuration from the Git repository to be deployed and tested.

![Multi-cluster](../img/multi-cluster.jpg)

## Examples
- [Supply Chain: Source => Image => Git](https://github.com/vmware-tanzu/cartographer/tree/main/examples/gitwriter-sc/README.md): 
  This simple ClusterSupplyChain takes source code, builds an image, generates app configuration, and writes it to a Git repository. 
  
- [Delivery: Git => App](https://github.com/vmware-tanzu/cartographer/tree/main/examples/basic-delivery/README.md): 
  This simple ClusterDelivery picks up app configuration from a Git repository and deploys to the cluster. 




