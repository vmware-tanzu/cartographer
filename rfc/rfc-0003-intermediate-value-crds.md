# RFC 0003 Intermediate Value CRDs

## Summary

Create a CRD type that will be used to store the values emitted by a node in a supply chain.
These new objects will give operators increased visibility into the what is happening with
a supply chain.

## Motivation

Kontinue currently passes values from one component of a supply chain to another in memory.
This is not observable to end users and can make it difficult to reason about the state of
the supply chain.

E.g.: when component N points to a SourceTemplate, that component implicitly emits a url
and a revision, which can be consumed by later components in the SupplyChain DAG. Imagine that
an operator has made a mistake in writing their SourceTemplate and both the urlPath and the
revisionPath point to the same field in the GitResource. The supply chain will break in some
strange way (probably kpack will complain that it is misconfigured). If the operator examines
the GitResource, they will see that it is behaving as expected. By creating an intermediate
value CRD, the operator would be able to examine the field and directly observe the mistake.

## Detailed Explanation

Create an IntermediateValue CRD. This can have 4 fields: url, revision, build, opinion.

When a component stamps out an object (e.g. when component 1 uses a SourceTemplate to stamp
out a GitResource), the component then stamps out a IntermediateValueStore and writes the
expected values in. Similar to the stamped out objects, the IntermediateValueStore must be:
- Owned by the workload
- Labeled with the workload, supply-chain, and component

When a follow-on template then references an input, the field in question will exist not simply in the controller’s memory, but also in a k8s object.


```yaml
## SourceTemplate with mistake
apiVersion: kontinue.io/v1alpha1
kind: ClusterSourceTemplate
metadata:
  name: git-template
spec:
  urlPath: $(status.artifact.url)$
  revisionPath: $(status.artifact.url)$  # <--- mistake
  template:
    apiVersion: source.toolkit.fluxcd.io/v1beta1
    kind: GitRepository
    metadata:
      name: $(workload.name)$-source
    spec:
      interval: 5m
      url: $(workload.git.url)$
      ref: $(workload.git.ref)$
```

```yaml
## Resulting proper GitResource
apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: GitRepository
metadata:
  creationTimestamp: "2021-03-02T22:25:30Z"
  finalizers:
  - finalizers.fluxcd.io
  generation: 22
  labels:
    experimental.kontinue.io/digest: 3a2531a9382654b15865220e9b51f8d4
  name: example-app
  namespace: default
  ownerReferences:
  - apiVersion: experimental.kontinue.io/v1
    blockOwnerDeletion: true
    controller: true
    kind: Workload
    name: example-app
    uid: 82698cb8-2c1a-4c55-8742-25b13b39521a
  resourceVersion: "53158"
  uid: 0d5b2a20-adb9-41f7-b61c-ee434596416e
spec:
  gitImplementation: go-git
  interval: 1m0s
  ref:
    branch: master
  secretRef:
    name: gitlab-credentials
  timeout: 20s
  url: https://gitlab.eng.vmware.com/tanzu-delivery-pipeline/example-app
status:
  artifact:
    checksum: fb5edd0cd4ab245791af41aea8ab8654a835ee87
    lastUpdateTime: "2021-03-02T22:48:03Z"
    path: gitrepository/default/example-app/115fc159352e68d4b4b17fd358924faa9fce1d22.tar.gz
    revision: master/115fc159352e68d4b4b17fd358924faa9fce1d22
    url: http://source-controller.gitops-toolkit.svc.cluster.local./gitrepository/default/example-app/115fc159352e68d4b4b17fd358924faa9fce1d22.tar.gz
  conditions:
  - lastTransitionTime: "2021-03-02T22:48:03Z"
    message: 'Fetched revision: master/115fc159352e68d4b4b17fd358924faa9fce1d22'
    reason: GitOperationSucceed
    status: "True"
    type: Ready
  observedGeneration: 22
  url: http://source-controller.gitops-toolkit.svc.cluster.local./gitrepository/default/example-app/latest.tar.gz
```

```yaml
## Example Intermediate Value CRD where error can be observed
apiVersion: kontinue.io/v1alpha1
kind: IntermediateValue
metadata:
  name: git-template
  labels:
    kontinue.io/workload-name: petclinic
    kontinue.io/cluster-supply-chain-name: responsible-ops
    kontinue.io/component-name: source-provider
    kontinue.io/cluster-template-name: git-template
  ownerReferences:
    - apiVersion: kontinue.io/v1alpha1
      kind: Workload
      name: petclinic
spec:
  url: http://source-controller.gitops-toolkit.svc.cluster.local./gitrepository/default/example-app/115fc159352e68d4b4b17fd358924faa9fce1d22.tar.gz
  revision: http://source-controller.gitops-toolkit.svc.cluster.local./gitrepository/default/example-app/115fc159352e68d4b4b17fd358924faa9fce1d22.tar.gz
```

## Rationale and Alternatives

- Create an IntermediateSourceValue, IntermediateImageValue, IntermediateConfigValue. This is
  a small interation on IntermediateValue CRD.
- Store the information on the workload status. This would mean creating fewer objects in etcd
  (but the same amount of writes would be made to etcd; an expert on that component should
  weigh in on whether that is a meaningful benefit)
- Store the information on the template status. Templates are leveraged by multiple
  supply-chains, so this may lead to leaking information about different teams' repos and
  build artifacts.

## Implementation

The smallest change to the code base would simply be an additive change. After the realizer
makes a call to DO, the output could be written into a patchable `IntermediateValue` object.

## Prior Art
