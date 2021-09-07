# Draft RFC 0004 - Handle multiple supply chains selecting workload

## Summary

Elegantly handle the case where multiple supply chains select a workload.

## Motivation

What do we do when a workload is selected by multiple supply-chains? Two
scenarios can give rise to this:

1. Two supply chains can have the same selector.

2. A workload can have multiple labels. Each label could match with a different
   supply-chain.

Currently, we operate as if workloads will only reconcile with a single
supply-chain. Each workload has a status which reports progress on a single
stream of work. But if a workload is reconciled with two or more supplyChains
then there would need to be disambiguation between the state of the
workload-supply-chain-A pair and the state of the workload-supply-chain-B pair.

### Example Setup
For the scenarios below, assume that the user has submitted the following objects:
```yaml
apiVersion: kontinue.io/v1alpha1
kind: SupplyChain
metadata:
  name: full-ops
spec:
  selector:
    waciuma.com: web-app
  components:
    - name: source-provider
      templateRef:
        kind: SourceTemplate
        name: git-repository-battery

    - name: image-provider
      templateRef:
        kind: BuildTemplate
        name: kpack-battery

    - name: opinion-provider
      templateRef:
        kind: OpinionTemplate
        name: opinion-service-battery

    - name: cluster-sink
      templateRef:
        kind: ConfigTemplate
        name: cluster-sink-battery

---

apiVersion: kontinue.io/v1alpha1
kind: SupplyChain
metadata:
  name: just-build
spec:
  selector:
    waciuma.com: web-app
  components:
    - name: source-provider
      templateRef:
        kind: SourceTemplate
        name: git-repository-custom

    - name: image-provider
      templateRef:
        kind: BuildTemplate
        name: dockerfile-builder

---

apiVersion: kontinue.io/v1alpha1
kind: Workload
metadata:
  name: petclinic
  labels:
    waciuma.com: web-app
spec:
  git:
    url: https://github.com/spring-projects/spring-petclinic.git
    ref:
      branch: main
```

## Possible Solutions

### Report work for multiple supply chains in the workload
We could have the workload directly reconcile with each supply-chain that
selects them. The workload status would have a single `Ready` condition,
and then a list of supply-chains it was working with that each had conditions.

This structure has a `conditions` field with only one condition, `Ready`. It
then has a field for listing all of the supply-chains that have paired with the
workload. Each member of that list has a `conditions` field that matches the
conditions of the workload as they are currently on the main branch.

```yaml
apiVersion: kontinue.io/v1alpha1
kind: Workload
metadata:
  name: petclinic
  labels:
    waciuma.com: web-app
status:
  conditions:
    - type: Ready
      status: "False"
      reason: ErrorWithASupplyChain
  supplyChainsPaired:
    - supplyChainRef:
        name: full-ops
      conditions:
        - type: SupplyChainReady
          status: "True"
          reason: Ready
        - type: ComponentsSubmitted
          status: "True"
          reason: ComponentSubmissionComplete
        - type: Ready
          status: "True"
          reason: Ready
    - supplyChainRef:
        name: just-build
      conditions:
        - type: ComponentsSubmitted
          status: "False"
          reason: TemplateRejectedByAPIServer
        - type: Ready
          status: "False"
          reason: ComponentsSubmitted

```

### Report work in separate child objects
We can create workload-supply-chain pair child objects. Each would act as the
single workload does now. The parent workload status would simply be a single
`Ready` condition which is `True` only if all children are ready, `Unknown` if
any child is `Unknown` and `False` if any child is `False`. The
workload's status could contain pointers to each workload-supply-chain child.
Alternatively, users could use labels/ownerRefs to find these children.

What should this new object be called? `WorkloadSupplyChainChild`? Perhaps more
clearly, this new object should be a `SupplyChain` and the CRD we have been
calling a supply-chain should become a `SupplyChainBlueprint`.

#### The case for SupplyChainBlueprint
The name SupplyChainBlueprint more properly represents the
relationship of the objects, even under the current architecture. The
current supply chain CRD represents a template for stamping out multiple
instantiations of a supply chain, instantiations that only occur once a workload
is present. To put it another way, currently for every supply-chain object in
a cluster, there are multiple logical supply chains.

#### Example

```yaml
apiVersion: kontinue.io/v1alpha1
kind: Workload
metadata:
  name: petclinic
  labels:
    waciuma.com: web-app
status:
  conditions:
    - type: Ready
      status: "False"
      reason: SupplyChainNotReady
      message: "ready status of workloadSupplyChainChild petclinic/just-build is false"

---

apiVersion: kontinue.io/v1alpha1
kind: SupplyChainBlueprint
metadata:
  name: full-ops
spec:
  selector:
    waciuma.com: web-app
  components:
    ...

---

apiVersion: kontinue.io/v1alpha1
kind: SupplyChainBlueprint
metadata:
  name: just-build
spec:
  selector:
    waciuma.com: web-app
  components:
    ...

---

apiVersion: kontinue.io/v1alpha1
kind: SupplyChain
metadata:
  name: petclinic/full-ops
status:
  conditions:
    - type: SupplyChainBlueprintReady
      status: "True"
      reason: Ready
    - type: ComponentsSubmitted
      status: "True"
      reason: ComponentSubmissionComplete
    - type: Ready
      status: "True"
      reason: Ready

---

apiVersion: kontinue.io/v1alpha1
kind: SupplyChain
metadata:
  name: petclinic/just-build
status:
  conditions:
    - type: SupplyChainBlueprintReady
      status: "True"
      reason: Ready
    - type: ComponentsSubmitted
      status: "False"
      reason: TemplateRejectedByAPIServer
    - type: Ready
      status: "False"
      reason: ComponentsSubmitted

```

## Cross References and Prior Art

## Previous discussion:
[Previous home of this RFC](https://docs.google.com/document/d/1b1VijMasimW95Q3xE99aifwuXyh-qD6Jdn2W38TeHxQ/edit#)

## Notes
If the team decides to adopt the change from SupplyChain to SupplyChainBlueprint,
there is a [draft of steps/stories to carry out the work](https://gitlab.eng.vmware.com/kontinue/kontinue/-/issues/35).
