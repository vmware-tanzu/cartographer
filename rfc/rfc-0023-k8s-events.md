# Meta

[meta]: #meta

- Name: Introduce Kubernetes Events
- Start Date: 2022-03-28
- Author(s): [Rasheed Abdul-Aziz](https://github.com/squeedee)
- Status: Approved
- RFC Pull Request: https://github.com/vmware-tanzu/cartographer/pull/756
- Supersedes: N/A

# Summary

[summary]: #summary

Introduce the use
of [K8s Events](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#events)
to provide user's awareness of state-changing event's that are meaningful for debugging and situational awareness.

# Motivation

[motivation]: #motivation

Developers and Operators need to debug issues where workloads stall, do not make it to deployment, or where other level
changes do not propagate as they should.

Cartographer concerns itself with creating **Stamped Objects** from **Templates** and passing **Artifact**
references between them. By design, the passing of references (and desired state definitions in the **owner** spec)
is _by value_. A user can only see if a state change is transferred from source to sink, not when it happened, or how
often, or even what the specific cause was for a resource to be updated.

This approach lends itself well to the declarative, eventual consistency model of Kubernetes, where "I want this, I
don't care how you get me there" reigns. However, there are times when a user needs to know more detail about the
machinations of the Cartographer controller, especially in debugging failures. For these situations, Kubernetes
provides [events](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#events)
to expose the user to these temporal triggers/actions from inside controller processes.

**Stamping** (creation of the resource definition from template and inputs) and **Submitting** (applying the resource
definition to the cluster) are two very common events, with associated causes, that currently need further exposure to
users.

# What it is

[what-it-is]: #what-it-is

## Developer View

Cartographer presents useful information to developers in the **owner** (workload/deliverable). This information is
derived from stamped objects and contributing resources such as **templates** and **blueprints**:

This RFC addresses the absence of awareness that "something is happening", especially after a developer has committed
new code. With events we can provide information that shows activity and causality not present in the current snapshot
of state.

We add events to the **Owners** (Workload/Deliverable) to enable debug and reasoning about Cartographer's behavior.

## Operator View

The situation is worse for users managing the **blueprints** (Devops/Ops). Some metrics can be recorded in **
blueprints**
and this might be a valid topic for another RFC, however events also present a useful way for authors to avoid trawling
through logs to get a quick overview of temporal behaviour of **owners** (workload/delivery) as they rely on the
author's
**blueprint**.

# How it Works

[how-it-works]: #how-it-works

To avoid over-communicating, and making the view of events overwhelming and dilute for users we will start with two
key guiding principles:

1. > Let it be something that isn't satisfactorily presented in other user views (eg: `k get workload`).
2. > Let the event list be as meaningful as is necessary, and no more.

These are examples of types of events that could be emitted. This RFC makes no recommendation as to which events are
required. The cartographer controller could emit the following events:

| Reason | Message Format | Description | involvedObject |
| --- | --- | --- | --- |
| StampedObjectExternalSpecChange | an external actor changed the spec of "\<resource>.\<group>/\<name>" | Our cache of the spec, and the spec we just generated match, but the API server has a different one. A lot of these is evidence of thrashing with external resources | Owner |
| StampedObjectExternalSpecChange | an external actor changed the spec of resource: "\<resource name>" for object "\<resource>.\<group>" | Our cache of the spec, and the spec we just generated match, but the API server has a different one. A lot of these is evidence of thrashing with external resources | Blueprint |
| StampedObjectGetError | "\<resource name>" could not retrieve "\<resource>.\<group>/\<name>" due to error: "\<error message>" | Loading a resource is failing due to a client.get issue (missing is not an error) | Owner |
| StampedObjectGetError | "\<resource name>" could not retrieve "\<resource>.\<group>" due to error: "\<error message>" | Loading a resources is failing due to a client.get issue (missing is not an error) - this could be spammy | Blueprint |
| StampedObjectGarbageCollected | "\<resource>.\<group>/\<name>" is no longer referenced | This owner has selected a different template, either by supply chain selection or templating, and this object is no longer needed | Owner |
| ImmutableStampedObjectGarbageCollected | "\<n> * \<resource>.\<group>" historical objects deleted due to garbage collection policy | This runnable's GC policy has caused "n" objects to be removed | Runnable |
| StampedObjectInvalid | "\<resource name>" could not be applied as "\<resource>.\<group>/\<name>" due to API server error "\<error message>" | This object was (probably) malformed. | Owner |
| StampedObjectInvalid | "\<resource name>" could not be applied as "\<resource>.\<group>" due to API server error "\<error message>" | This object was (probably) malformed. This lets operators know their templates might have issues | Blueprint |
| StampedObjectKindNotFound | "\<resource name>" could not be applied because "\<resource>.\<group>" does not exist on this cluster | Did someone forget to install the CRDs? Otherwise it's a malformed template | Owner |
| StampedObjectKindNotFound | "\<resource name>" could not be applied because "\<resource>.\<group>" does not exist on this cluster | Did someone forget to install the CRDs? Otherwise it's a malformed template | Blueprint |
| StampedObjectApplied | "\<resource name>" was applied as "\<resource>.\<group>/\<name>" | a stamped object needed to be created/updated | Owner |
| StampedObjectKindChanged | "\<resource name>" was "\<old group>.\<old kind>", now "\<new group>.\<new kind>" | YTT selection for a template GVK changed, or templated values in GVK changed | Owner |
| ResourceOutputChanged | "\<resource name>" found a new output in "\<resource>.\<group>/\<name>" | a resource produced a new output | Owner |
| ResourceHealthyStatusChanged | "\<resource name>" found a new status in "\<resource>.\<group>/\<name>" | a resource produced a new healthy status | Owner |
| SupplyChainChanged | supply chain changed from "\<old supply chain name>" to "\<new supply chain name>" | Workload selected for a new or different supply chain. Note: "none" is a possible name. | Owner |
| WorkloadSelected | supply chain matched a workload | A workload was selected for (really, this is when the supplyChainRef changes). This makes the count a useful metric | Blueprint |

**Note:** `<resource name>` represents the blueprint's `.spec.resource[].name` and *not* the stamped object's `.metadata.name`

**Note:** Blueprint event `messages` are usually designed to aggregate on the resource kind, not the individual
resource (we omit `<resource name>`)

**Note:** On naming, we're tring to start with the location in the object diagram (first diagram) that represents the
change. `stamped object` is not well known to users, but we need to rectify that, or use a clearer term to describe "the
resource on etcd", which would normally be "Resource". Unfortunately we've overloaded resource to mean the collective
concept of a template, a configuration, its submission to the API and its resulting changes from external events.

There are a lot of events here, and more could exist, but we should review these during the RFC review process. I hope
we can prune the list and make it `as meaningful as is necessary, and no more`

Example output:

```text
  Type    Reason                      Age                    From                 Message
  ----    ------                      ----                   ----                 -------
Normal    ResourceExternalSpecChange  1m59s (x2 over 1m59s)  workload-controller  an external actor changed the spec of kpack.io.image/java-sample-app
Normal    ResourceExternalSpecChange  35s (x4 over 22s)      workload-controller  an external actor changed the spec of kappctrl.k14s.io/java-sample-app
```

# Migration

[migration]: #migration

This RFC only adds to the API, there are no breaking changes that require migration.

# Drawbacks

[drawbacks]: #drawbacks

Very little, this is the right thing to be doing.

* Making API calls to emit events will slow processing down, and increase API consumption.
* In code we either end up passing more contextual information to the code-sites that events originate from, or we
  create a better abstraction than controller-runtimes event recorder to keep value passing to a minimum. For example
  an `InvolvedEvent` function that curries the `Workload`, `Delivery`, `SupplyChain` or `Deliverable` as the involved
  object.

# Alternatives

[alternatives]: #alternatives

- What other designs have been considered?
    - Long tail historic information in the `workload.status` etc. This would be obnoxious for users and require some GC
      over time, whereas events are built for this purpose.
- Why is this proposal the best?
    - This is what events are for. It's idiomatic and well supported
- What is the impact of not doing this?
    - Leaving users without temporal event information, making them understand resources intimately, when the side
      effects are all the user really cares about.

# Prior Art

[prior-art]: #prior-art

See k8s native Pods etc.

# Unresolved Questions

[unresolved-questions]: #unresolved-questions

- How many of these events do we really need
- How do we avoid an exhaustive and hard to parse list of events.
