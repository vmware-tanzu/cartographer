# Draft RFC 20 Read Resources Only When In Success State

## Summary

Cartographer is currently unable to connect an output of a resource to the inputs of said resource. Cartographer is
also unable to determine when an input to a resource has reached success, failed, or is still processing.
Templates should enable authors to specify the indications of successful reconciliation. Given that, Carto should
only read the status of an object when said object has completed reconciling and is in a good state. (Updates to the
object may happen continuously)

## Motivation

Connecting an output of a resource to an input is necessary for establishing tracing. That is to state, "The app
currently running on the cluster is a result of resource X producing Y which was fed into resource Z which produced..."
it is necessary to tie a resource output to the input that produced it. Waiting for success/failure before update
achieves this.

Tying inputs to outputs also allows Carto to harden supply chains to tampering. Currently, an enemy could update a
resource. Carto will not read said resource without first updating it with the proper spec. But the resource may still
produce the enemy output (followed by the correct output). Carto must be able to associate inputs to outputs if it is
not to propagate the enemy output.

## Proposed Solution

Cartographer can only read the status of an object when said object has completed reconciling and is in a good state.

Other proposed solutions are inadequate as illustrated in scenarios below.

_For the purposes of this discussion, we will assert only observer the status of an object when it represents the
current spec of the object. Strategies for assuring this (e.g. checking that .metadata.generation ==
.status.observedGeneration) are discussed in the implementation section._

Assumptions for the scenario:

- There exists a resource type A. An instance of this is an object A.
- Resource A reports `latestGoodOutput`. That is, the controller A reads the spec of object A, combines it with
knowledge that it has of the rest of the world, and computes an output. If the output is 'good' (by the internal logic
of controller A) then the output is reported in object A's .status.latestGoodOutput field. If the computed output is
bad, the .status.latestGoodOutput field is untouched. Therefore, if a previous good output had been calculated, it
would be in the field.
- Resource A reports Ready:true, Ready:false or Ready:unknown. If Ready:true, controller A has reconciled the current
spec against the current state of the world and the output of the most reconcile was good. If Ready:false, the
controller has similarly reconciled the most recent spec, but the resulting output was bad. If Ready:unknown,
controller is currently attempting to reconcile the most recent spec against the current state of the world.
- Changes in the state of the world will trigger the controller to reconcile using object A's most recent spec.
- (Resource A is a thin facsimile of Runnable or kpack)

_Carto can update continuously, but can only read when an object has completed reconciling (in a good or bad state)_:
Cartographer updates object A. Before the object completes reconciliation, Cartographer updates the object twice more.
Later, Cartographer goes to read the status. It sees object A is Ready:false. It also sees that there is a new
latestGoodOutput. It knows that this output is the result of the first or second update, but there is no way to
determine which. Sad Carto.

_Cartographer can only update the spec of an object when said object has completed reconciling, but Carto reads
continuously_: Cartographer observes that object A is Ready:true. Cartographer submits an
update of object A's spec. Unbeknownst to Cartographer, an instant before that the state of the world changes,
causing controller A to reconcile the previous spec of object A with the current state of the world. The reconcile is
successful, so object A's .status.lastGoodOutput is updated. Object A remains Ready:unknown, because of the new spec
that has been submitted. But because reading the state of the object is not constrained (in this scenario) Cartographer
reads the new lastGoodOutput and incorrectly attributes it as the result of the most recently submitted spec. This
is incorrect. Bad Carto.

_Cartographer can only read or update the spec of an object when said object has completed reconciling (in good or bad
state)_: Cartographer observes that object A is Ready:true. Cartographer submits an update of object A's spec. Later,
Cartographer observes that Ready:false _and_ that there is a new latestGoodObject. Either the old spec or the new spec
could have caused this latestGoodOutput:
- the previous spec: immediately before the update is submitted, the state of the world changes. The latestGoodOutput
is a result of the previous spec and the new state of the world. The new spec is then reconciled against the new
state of the world and the 2nd reconcile fails. Ready:false.
- the new spec: **after** the update is submitted, the state of the world changes. The latestGoodOutput is a result of the
new spec and the previous state of the world. The new spec is then reconciled against the new state of the world and
the 2nd reconcile fails. Ready:false.
Cartographer cannot attribute the latestGoodOutput. Sad Carto.

By constraining reading to occur only when object A is in Ready:true state, all of these scenarios can be addressed.
By definition, Ready:true indicates that the current output is the result of reconciliation of "the current spec
against the current state of the world and the output of the most reconcile was good." Other strategies are
insufficient.

### How to ensure the most recent successful?

Let us again consider resource A. Two updates are made to resource A in quick succession. The first succeeds and
updates the `latestGoodOutput`. But the status is still Ready:unknown, so Cartographer does not propagate this output.
The second fails and results in `Ready:false`. Cartographer will not propagate any output because `Ready:false`.
How can this be addressed?

Insufficient Strategy:
Cartographer could wait to submit a new spec until the previous spec has finished reconciling. This would still need
to be paired with only ready when Ready:true (as demonstrated above). But even more than that we can see that this will
either leave Carto vulnerable to an equivalent problem or introduce potential bottlenecks in the supply chain. The
difference is in the method of submitting waiting updates.
a. Assume a strategy of not submitting while work is occurring, when work is done, submit the most recent update
(pop from the stack). Here we find a problem analogous to the original. Consider update 1. While it works, update 2
(good) and 3 (bad) come in. Once update 1 finishes processing, update 3 will be submitted. Update 3, being bad, does
not result in a valid output. And the latest update that would produce an output (update 2) is never submitted. Bad Carto.
b. Again assume a strategy of not submitting while work is occurring. This time, when work is done, submit the earliest
update (first in the queue). This will result in every update being submitted. But it will be slow. If resource A takes
a long time to reconcile, users may not want to wait on outdated updates.

Happy Middle Ground Strategy:
As mentioned, strategy A pops the most recent update from a stack. This suggests an optimal strategy. Pop updates from
the stack until an update produces a good output. At that point the stack can be cleared. This assures that the most
recent good update results in an output, but does not waste time needlessly calculating the output of updates that have
already been superseded.

## Implementation details

### Use ObservedCompletion and ObservedMatch (from DeploymentTemplate)

Templates that expose outputs (SourceTemplate, ImageTemplate...) can include success conditions. These can include
expected values at a given path on the stamped object. These can alternately be an expectation that one field (e.g.
on the status) matches another field (e.g. on the spec).

- An observed completion: includes a mandatory SucceededCondition and an optional FailedCondition. Both conditions are 
  defined by a path and value. When the object's observedGeneration == generation, and the value at the
  specified path matches the stated value, then this condition is met. 
- An observed matches: a list of matches. Each match is a definition of two paths. When the values at the two paths are
  the same, then this condition is met. This can be used for resources that do not report observedCondition, but whose
  status does include relevant fields in the spec.

### Example

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterBuildTemplate
metadata:
  name: example-build---consume-output-of-components
spec:
  template:
    apiVersion: kpack.io/v1alpha1
    kind: Image
    metadata:
      name: ...
    spec:
      ...
  imagePath: $(status.latestImage)$
  observedCompletion:
    succeeded:
      - key: status.conditions[?(@.type=="Ready")].status
        value: True
```

### Handling non-reconciling objects

Some objects are not reconciled and are immediately valid (for example, configmaps). How should these be handled?

1. Users could write trivial ObservedMatches. For example asserting that the `.data` field is equal to the `.data`
   field.
2. The absence of either an ObservedCompletion or an ObservedMatch could be taken as indication that the object
   being created is immediately succesful. (This seems dangerous and ill-advised)
3. An additional exclusive field (one that could be specified instead of ObservedMatches and ObservedCompletion) could
   be defined. This field could be `AlwaysSuccessful: True`

### Limitations

There are a few limitations to the current setup of observedCompletion and ObservedMatches:
1. ObservedCompletion is limited to matching a single path and value. If more than one path must be interrogated,
   this spec is not sufficient.
2. ObservedMatches cannot define a failure state.

## Tradeoffs

### The performance cost of tracing

Consider workload X, which is updated quickly and continuously with new commits. Assume that resource A takes a long
time to reconcile, longer than the time between 2 commits. In this scenario, resource A will remain in Ready:Unknown
state. Therefore, Cartographer will not be able to read the output of the resource. As a result, the resources in the
supply chain after resource A will never receive updates. Sad hypothetical, [the devs owning the workload will have to
take a chocolate break at some point](https://xkcd.com/303/).

### Deadlock

Currently, Cartographer basks in the _eternal sunshine of the spotless mind_; each reconcile loop for resource N it can
only pass on the values it _just_ read for resource N-1, N-2, N-3... So if any of those resources are in a bad state,
a supply chain is locked. Resources like our example resource A are good actors in this system, as they constantly
report the most recent good output. After a single good input in the life of object A, it would always pass a value
and never be a concern for stopping the supply chain.

With this proposal, even good teammate resource A can lock downstream resources simply by updating. If no other changes
are undertaken, Cartographer would be unable to pass a value from resource A to a downstream resource when resource A
had the status Ready:false or even Ready:unknown. Thus Cartographer could not stamp out the desired spec of the
downstream resource. Without knowledge of the desired spec of the downstream resource, Cartographer would not read
the status of the downstream resource. The supply chain would grind to a halt.

When Cartographer sees a good value, it must keep a _memento_. **Each time a new output is readable on an object
(Ready:true), Cartographer must cache that value. Whenever Cartographer cannot read the value from an object
(Ready:false or Ready:unknown) Cartographer should propagate the most recent cached values to the downstream objects.**

The implementation for such a cache is thankfully proposed in
[RFC 18](https://github.com/vmware-tanzu/cartographer/pull/519). One additional `artifact` field will be necessary to
those proposed in RFC 18, a timestamp. RFC 18 currently assumes that multiple artifacts from a single object could be
cached at once (in the case where a downstream object is still a child of the earlier state; that the new state has
not propagated through entire supply chain). Cartographer will need to determine which cached value to pass to
downstream objects. There is no currently proposed field that can be leveraged for this determination (resourceVersions
are not guaranteed to increase monotonically).

## Possible Extensions

### Allow boolean operations

- An OR condition: An OR contains a list of conditions. When any is met, the OR condition is met.
- An AND condition: A list of conditions. When all conditions are met, this condition is met. Meant primarily for
  nesting in OR conditions (as AND is the default relation of a list of conditions).
- A NOT condition: holds a condition. The NOT condition is true only when its condition is false.

### Read other objects on the cluster

It may be useful to compare the stamped object to another object on the cluster. Or to simply read a value from
another object on the cluster.

## Cross References and Prior Art

- The Deployment Template currently requires either an `ObservedMatches` or `ObservedCompletion` field.
- kpack includes a [waitRules](https://carvel.dev/kapp/docs/v0.45.0/config/#waitrules) field. (hat tip
[Scott Andrews](https://github.com/scothis))