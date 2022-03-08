# Installing Cartographer

## Prerequisites

- Kubernetes cluster v1.19+
- cert-manager, see [cert-manager Installation](https://cert-manager.io/docs/installation/)

## Install

1. Apply `cartographer.yaml`

   ```bash
   kubectl apply -f https://github.com/vmware-tanzu/cartographer/releases/download/v0.2.0/cartographer.yaml
   ```

   Resources in file `cartographer.yaml`:

   ```console
   Namespace            Name                                  Kind
   (cluster)            cartographer-cluster-admin            ClusterRoleBinding
   ^                    cartographer-controller-admin         ClusterRole
   ^                    clusterconfigtemplates.carto.run      CustomResourceDefinition
   ^                    clusterdeliveries.carto.run           CustomResourceDefinition
   ^                    clusterdeploymenttemplates.carto.run  CustomResourceDefinition
   ^                    clusterimagetemplates.carto.run       CustomResourceDefinition
   ^                    clusterruntemplates.carto.run         CustomResourceDefinition
   ^                    clustersourcetemplates.carto.run      CustomResourceDefinition
   ^                    clustersupplychains.carto.run         CustomResourceDefinition
   ^                    clustersupplychainvalidator           ValidatingWebhookConfiguration
   ^                    clustertemplates.carto.run            CustomResourceDefinition
   ^                    deliverables.carto.run                CustomResourceDefinition
   ^                    deliveryvalidator                     ValidatingWebhookConfiguration
   ^                    runnables.carto.run                   CustomResourceDefinition
   ^                    workloads.carto.run                   CustomResourceDefinition

   cartographer-system  cartographer-controller               Deployment
   ^                    cartographer-controller               ServiceAccount
   ^                    cartographer-webhook                  Certificate
   ^                    cartographer-webhook                  Secret
   ^                    cartographer-webhook                  Service
   ^                    private-registry-credentials          Secret
   ^                    selfsigned-issuer                     Issuer
   ```

## Uninstall

1. Delete `cartographer.yaml`
   ```bash
   kubectl delete -f https://github.com/vmware-tanzu/cartographer/releases/download/v0.2.0/cartographer.yaml
   ```
