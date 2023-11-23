# Draft RFC 0005 Snippets: Support composing supply chains

## Summary

Allow users to define a new resource type: a snippet. A snippet itself defines
templates to be stamped out (similar to a supply chain).

## Motivation

At our favorite company, there is a team that is expert at building images.
They know exactly how each team should do image creation, and that process
takes a couple of steps of a supply chain. The image building team wants to
write a master image building supply chain. Every ops team in the company will
write a supply chain that starts with their own definitions (get code, test, etc),
then calls out to the image building team’s supply chain to build the image,
then comes back to the ops team’s supply chain for final steps.

### Implications/Assumptions
The image-building team does not expect their piece of the supply-chain to
reconcile with workloads on its own. They are writing a snippet that will only
make sense in the context of another supply-chain because:
- If the image-building team's supply-chain reconciles directly with supply-chains,
  then that supply-chain must have the same selector as any other supply-chain that
  wants to use it. That doesn't sound like it fits the use case.
- Presumably the image-building team's supply-chain will be deployed earlier than
  other supply-chains relying on it. If this early supply-chain tries to reconcile
  with a workload directly, but the components expect to consume some composed
  component output, the image-building team's supply-chain will error out immediately.

If the supply-chain isn't reconciling with a workload, it's not really a supply-chain.
It's some supply-chain snippet.

## Possible Solutions

Kontinue should provide a SupplyChainSnippet CRD.

A supply chain will refer to a snippet by reference (as if the snippet were a
*Template).

The snippet will define components, just as a supply chain does. A snippet will
be written with the assumption that it is called by a SupplyChain. The snippet
will be able to consume values from components in the SupplyChain as if they were
components in the snippet. SupplyChainSnippet should have no selector.

The snippet will emit some value/type, that would then be available to components
in the SupplyChain that called the snippet. There are a few way the emitted value
could be determined:
- By convention, the last component listed in a Snippet could be the component
  emitting values to components in the SupplyChain.
- The Snippet could explicitly define which component could emit values.

_Could a Snippet emit multiple values?_ It should not. If a snippet emitted values
from two ImageTemplates, for example, there would be no way for the calling
SupplyChain to differentiate these two values (as the SupplyChain treats the Snippet
as just another template).

_Should snippets be typed?_ This would result in having a SupplyChainSourceSnippet,
SupplyChainImageSnippet, SupplyChainConfigSnippet, etc. This multiplication of
resources adds cognitive and code overhead in exchange for dubious value.

### Example
```yaml
apiVersion: kontinue.io/v1alpha1
kind: ClusterSupplyChainSnippet
metadata:
  name: responsible-ops-middle
spec:
  resources:
    - name: built-image-provider
      templateRef:
        kind: ClusterImageTemplate
        name: kpack-battery
      sources:
        - component: source-provider
          name: solo-source-provider
    - name: opinion-service-workload-template-provider
      templateRef:
        kind: ClusterConfigTemplate
        name: opinion-service-battery
      images:
        - component: built-image-provider
          name: solo-image-provider
  exportedComponent:
    name: opinion-service-workload-template-provider

---

apiVersion: kontinue.io/v1alpha1
kind: ClusterSupplyChain
metadata:
  name: responsible-ops-composed
spec:
  selector:
    integration-test: "my-workload"

  resources:
    - name: source-provider
      templateRef:
        kind: ClusterSourceTemplate
        name: git-repository-battery

    - name: composed-supply-chain-snippet
      templateRef:
        kind: ClusterSupplyChainSnippet
        name: responsible-ops-middle
      sources:
        - component: source-provider
          name: solo-source-provider

    - name: cluster-sink
      templateRef:
        kind: ClusterTemplate
        name: cluster-sink-battery
      config:
        - component: composed-supply-chain-snippet
          name: singular-workload-template-provider

```

## Use Case

RFC 9 introduced the ability for a resource in a supply chain to swap 1-1 between
different templates, based on fields present on the workload. There is still
need to allow a supply chain to swap 1-N templates, that is to allow an optional
path through a supply chain to be of arbitrary length, unbound by the length of
other paths. Snippets can allow this. A resource could choose between a regular
template and a snippet. That snippet can be of arbitrary length.

### Example

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSupplyChain
metadata:
  name: supply-chain
spec:
  selector:
    matchLabels:
      workload-type: web

  resources:
    - name: provide-image
      templateRef:
        kind: ClusterImageTemplate
        options:
          - name: building-image
            selector:
              matchFields:
              - key: "spec.source.url"
                operation: exists
          - name: building-image
            selector:
              matchFields:
              - key: "spec.source.image"
                operation: exists
          - name: image-provider
            selector:
              matchFields:
              - { key: "spec.image", operation: exists }

    - name: configure
      templateRef:
        kind: ClusterConfigTemplate
        name: configure
      images:
        - resource: provide-image
          name: image

    - name: gitops
      templateRef:
        kind: ClusterTemplate
        options:
          - name: git-pusher
            selector:
              matchFields:
              - key: "metadata.labels.target"
                operation: In
                value: ["gitops"]
          - name: registry-pusher
            selector:
              matchFields:
              - key: "metadata.labels.target"
                operation: In
                value: ["repository"]
      configs:
        - resource: configure
          name: config

---
apiVersion: kontinue.io/v1alpha1
kind: ClusterSupplyChainSnippet
metadata:
  name: building-image
spec:
  resources:
    - name: provide-source
      templateRef:
        kind: ClusterSourceTemplate
        options:
          - name: providing-and-testing-source
            selector:
              matchFields:
              - key: "metadata.labels.has-tests"
                operation: In
                value: ["true"]
          - name: source-from-git-repo
            selector:
              matchFields:
              - { key: "spec.source.url", operation: exists }
              - { key: "metadata.labels.has-tests", operation: NotIn, value: ["true"] }
          - name: source-from-image-registry
            selector:
              matchFields:
              - { key: "spec.source.image", operation: exists }
              - { key: "metadata.labels.has-tests", operation: NotIn, value: ["true"] }
    - name: image-builder
      templateRef:
        kind: ClusterImageTemplate
        name: build-image
      sources:
        - resource: provide-source
          name: source

---
apiVersion: kontinue.io/v1alpha1
kind: ClusterSupplyChainSnippet
metadata:
  name: providing-and-testing-source
spec:
  resources:
    - name: provide-source
      templateRef:
        kind: ClusterSourceTemplate
        options:
          - name: source-from-git-repo
            selector:
              matchFields:
              - { key: "spec.source.url", operation: exists }
          - name: source-from-image-registry
            selector:
              matchFields:
              - { key: "spec.source.image", operation: exists }
    - name: test-source
      templateRef:
        kind: ClusterSourceTemplate
        name: source-tester
```

## Cross References and Prior Art

- If RFC 0004 is adopted, this should become a _Blueprint_ Snippet (or more
  verbosely, a SupplyChainBlueprintSnippet)
