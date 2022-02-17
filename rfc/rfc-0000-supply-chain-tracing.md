# Meta
[meta]: #meta
- Name: Supply Chain Tracing
- Start Date: 2020-02-15
- Author(s): @scothis
- Status: Draft <!-- Acceptable values: Draft, Approved, On Hold, Superseded -->
- RFC Pull Request: (leave blank)
- Supersedes: N/A

Note: Unless stated otherwise, the concepts in this RFC attributed to `Workload`s and `ClusterSupplyChain`s apply equally to `Deliverable`s and `ClusterDelivery`.

# Summary
[summary]: #summary

Cartographer users create `Workload` and `Deliverable` resources that define work to be processed by `ClusterSupplyChain` and `ClusterDelivery` resources respectively. The work preformed by Cartographer on behalf of the user is largely opaque, and difficult to debug when something goes wrong. This RFC defines a way to trace the activity performed by Cartographer on behalf of the user's `Workload`/`Deliverable` by recording Cartographer's observed state of the world on the status of the resource that triggered the work.

While traceability is an important aspect of artifact provenance, the goal for this RFC is to increase visibility of Cartographers current behavior. This work may serve as a foundation to later attestations of provenance, but is itself not.

# Motivation
[motivation]: #motivation

For every `Workload` and `Deliverable` we should be able to answer:

- What resources were stamped?
- Which template was used to stamp out resources?
- What outputs did each stamped resource produce?

These questions are answerable today so long as you have permission on the cluster to view all the necessary resources, but doing so requires a deep understanding of Cartographer and the templates that define the resources to be stamped out. Instead we can distill the current state on the status of the resource in a form that can be used by client tools and users.

There are a number of use cases for this information, internal and external:

- tooling can visualize the resource graph of a given workload.
- Cartographer can be more aware of previous resources to aid garbage collection of orphaned resources.

# What it is
[what-it-is]: #what-it-is

After Cartographer matches a `ClusterSupplyChain` to a `Workload`, the resources defined by the supply chain are resolved to a template and template is applied to stamp out a kubernetes resource on the cluster. After the resource is created, outputs located at paths also defined by the template are read and propagated forward to the next resource in the chain. As Cartographer applies this behavior, it can keep a trace for each resource containing:

1. the name in the ClusterSupplyChain
2. a reference to the template
3. a reference to the stamped resource
4. named input resources as defined by the `ClusterSupplyChain`
5. the name and value of each output resolved from the stamped resource

The status for a Workload or Deliverable should be enhanced to add:

```yaml
... existing fields ...
status:
  ... existing fields ...
  resources:
  - name:                  # string
    templateRef:           # corev1.ObjectReference
      apiVersion:            # string
      kind:                  # string
      namespace:             # string (empty for cluster scoped resources)
      name:                  # string
    stampedRef:            # corev1.ObjectReference
      apiVersion:            # string
      kind:                  # string
      namespace:             # string (empty for cluster scoped resources)
      name:                  # string
    inputs:
    - name:                  # string
    outputs:
    - name:                  # string
      value:                 # json.RawExtension
      lastTransitionTime:    # metav1.Time
    observedGeneration:    # int64
```

Based on the [basic supply chain example](https://github.com/vmware-tanzu/cartographer/tree/53402edb8b8914b4cd36ace82da85f83f3daefc1/examples/basic-sc), the Workload status would have the form:

```yaml
... existing fields ...
status:
  ... existing fields ...
  resources:
  - name: source-provider
    stampedRef:
      apiVersion: source.toolkit.fluxcd.io/v1beta1
      kind: GitRepository
      namespace: default
      name: my-workload
    templateRef:
      apiVersion: carto.run/v1
      kind: ClusterSourceTemplate
      name: source
    outputs:
    - name: url
      value: http://source-controller.flux-system.svc.cluster.local./gitrepository/default/my-workload/3d42c19a618bb8fc13f72178b8b5e214a2f989c4.tar.gz
      lastTransitionTime: "2022-02-16T03:29:52Z"
    - name: revision
      value: main/3d42c19a618bb8fc13f72178b8b5e214a2f989c4
      lastTransitionTime: "2022-02-16T03:29:52Z"
    observedGeneration: 1
  - name: image-builder
    stampedRef:
      apiVersion: kpack.io/v1alpha2
      kind: Image
      namespace: default
      name: my-workload
    templateRef:
      apiVersion: carto.run/v1
      kind: ClusterImageTemplate
      name: image
    inputs:
    - name: source-provider
    outputs:
    - name: image
      value: registry.example/supply-chain/my-workload@sha256:68f8e8fc6e8ede7a411db9182cd695eac7b3e7e19e4ff9dcb9ba21205c135697
      lastTransitionTime: "2022-02-16T03:23:37Z"
      observedGeneration: 14
  - name: deployer
    stampedRef:
      apiVersion: kappctrl.k14s.io/v1alpha1
      kind: App
      namespace: default
      name: my-workload
    templateRef:
      apiVersion: carto.run/v1
      kind: ClusterTemplate
      name: app-deploy
    inputs:
    - name: image-builder
    observedGeneration: 19
```

# How it Works
[how-it-works]: #how-it-works

- `.status.resources[*]` should match the order and length of `.spec.resource` for a ClusterSupplyChain.
- `.status.resources[*].name` is the name of the resources as defined within the ClusterSupplyChain resource's array.
- `.status.resources[*].templateRef` object reference to the template resource referenced by the ClusterSupplyChain.
- `.status.resources[*].stampedRef` object reference to the Kubernetes resource created by Cartographer for this resource.
- `.status.resources[*].inputs[*]` inputs model the relationships between resources within the SupplyChain graph. The value of an input can be read from the referenced resource's outputs.
- `.status.resources[*].inputs[*].name` the name of the resource backing this input. 
- `.status.resources[*].outputs[*].name` the name of the output. Output names are fixed and defined by the template type.
- `.status.resources[*].outputs[*].value` raw value of the output. This may be large for config templates.
- `.status.resources[*].outputs[*].lastTransitionTime` time of the most recent observed change of value. this is not the last time the value observed. This value is helpful to know if a supply chain is progressing. As this field is the result of observation of an edge triggered change, it should not be relied upon when accuracy matters. This is the same behavior as the lastTransitionTime field within a Condition.
- `.status.resources[*].observedGeneration` the `.metadata.generation` of the stamped resource the outputs are read from. This value is not read from the `.status.observeredGeneration` field that may exist on the stamped resource. A future RFC may choose to take the stamped resource's observed generation under advisement, as that is a change to how Cartographer processes resources, it is inappropriate for this RFC.

In the future, additional information about particular resources may be added. Like an indication of health with an error message when unhealthy. Support is not included in this RFC as it will require additional support within the template to know how to interpret the health of a resource and should be defined in a separate RFC.

# Migration
[migration]: #migration

All of the data collected by this RFC exists within Cartographer today but is not exposed outwardly by the `Workload`. There are no changes to the runtime behavior required. All changes to existing CRDs are net new and will not break existing clients.

# Drawbacks
[drawbacks]: #drawbacks

Some users may consider the runtime behavior of a `ClusterSupplyChain` to be an implementation detail that a user creating a `Workload` should not have visibility into. That a particular resource is created by Cartographer for a `Workload` may be viewed as an information disclosure. This change shines light on the internal behavior of a supply chain. The exact definition for each template is not exposed, unless the user also has access to view that resource. The outputs produced for a resource are the most sensitive values captured, but should never contain any confidential data.

If there is significant user concern, a toggle could be developed to partially or fully disable tracing. Such a toggle is not in scope for this RFC.

# Alternatives
[alternatives]: #alternatives

The alternative to adding trace information to the Workload status is to continue not adding that information and require clients to reconstruct the values on their own. Other RFCs listed below under prior art have broader scopes that treat tracing as a side effect of deeper semantic changes to how Cartographer processes Workloads.

# Prior Art
[prior-art]: #prior-art

Several RFC have attempted to make stronger pushes into this space, including:

- [First draft of RFC 0014](https://github.com/vmware-tanzu/cartographer/pull/274)
- [Introduce RFC 18 Workload Report Artifact Provenance](https://github.com/vmware-tanzu/cartographer/pull/519)

# Unresolved Questions
[unresolved-questions]: #unresolved-questions

- Are the proposed field names well aligned with existing Cartographer-isms?
- Are resource level observedGeneration and output level lastTransitionTime values helpful and meaningful?
- Is all of the information collected and exposed appropriate to expose to all users who can view the workload resource?
- Is there value in capturing the uuid for referenced resources?

# Spec. Changes (OPTIONAL)
[spec-changes]: #spec-changes

For both the `Workload` and `Deliverable` resources their status is adding a `resources` array, see the [What it is](#what-it-is) section for details.
