# Draft RFC 20 Read Resources Only When In Success State

# Summary

Cartographer is currently unable to connect an output of a resource to the inputs of said resource. Cartographer is also
unable to determine when an input to a resource has reached success, failed, or is still processing. Templates should
enable authors to specify the indications of successful reconciliation. Given that, Carto should only read the status of
an object when said object has completed reconciling and is in a good state. (Updates to the object may happen
continuously)

# Motivation

Connecting an output of a resource to an input is necessary for establishing tracing. That is to state, "The app
currently running on the cluster is a result of resource X producing Y which was fed into resource Z which produced..."
it is necessary to tie a resource output to the input that produced it. Waiting for success/failure before update
achieves this.

Tying inputs to outputs also allows Carto to harden supply chains to tampering. Currently, an enemy could update a
resource. Carto will not read said resource without first updating it with the proper spec. But the resource may still
produce the enemy output (followed by the correct output). Carto must be able to associate inputs to outputs if it is
not to propagate the enemy output.

# Proposed Solution

Cartographer can only read the status of an object when said object has completed reconciling and is in a good state.

Other proposed solutions are inadequate as illustrated in scenarios below.

_For the purposes of this discussion, we will assert that Carto only observes the status of an object when it represents
the current spec of the object. Strategies for assuring this (e.g. checking that .metadata.generation ==
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
continuously_: Cartographer observes that object A is Ready:true. Cartographer submits an update of object A's spec.
Unbeknownst to Cartographer, an instant before that the state of the world changes, causing controller A to reconcile
the previous spec of object A with the current state of the world. The reconcile is successful, so object A's
.status.lastGoodOutput is updated. Object A remains Ready:unknown, because of the new spec that has been submitted. But
because reading the state of the object is not constrained (in this scenario) Cartographer reads the new lastGoodOutput
and incorrectly attributes it as the result of the most recently submitted spec. This is incorrect. Bad Carto.

_Cartographer can only read or update the spec of an object when said object has completed reconciling (in good or bad
state)_: Cartographer observes that object A is Ready:true. Cartographer submits an update of object A's spec. Later,
Cartographer observes that Ready:false _and_ that there is a new latestGoodObject. Either the old spec or the new spec
could have caused this latestGoodOutput:

- the previous spec: immediately before the update is submitted, the state of the world changes. The latestGoodOutput is
  a result of the previous spec and the new state of the world. The new spec is then reconciled against the new state of
  the world and the 2nd reconcile fails. Ready:false.
- the new spec: **after** the update is submitted, the state of the world changes. The latestGoodOutput is a result of
  the new spec and the previous state of the world. The new spec is then reconciled against the new state of the world
  and the 2nd reconcile fails. Ready:false. Cartographer cannot attribute the latestGoodOutput. Sad Carto.

By constraining reading to occur only when object A is in Ready:true state, all of these scenarios can be addressed. By
definition, Ready:true indicates that the current output is the result of reconciliation of "the current spec against
the current state of the world and the output of the most reconcile was good." Other strategies are insufficient.

# Implementation details

## Use ObservedCompletion and ObservedMatch (from DeploymentTemplate)

Templates that expose outputs (SourceTemplate, ImageTemplate...) can include success conditions. These can include
expected values at a given path on the stamped object. These can alternately be an expectation that one field (e.g. on
the status) matches another field (e.g. on the spec).

- An observed completion: includes a mandatory SucceededCondition and an optional FailedCondition. Both conditions are
  defined by a path and value. When the object's observedGeneration == generation, and the value at the specified path
  matches the stated value, then this condition is met.
- An observed matches: a list of matches. Each match is a definition of two paths. When the values at the two paths are
  the same, then this condition is met. This can be used for resources that do not report observedGeneration, but whose
  status does include relevant fields in the spec.

## Examples:

### observedCompletion:

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

### observedMatches:

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterBuildTemplate
  #metadata:
  #  name: example-build---consume-output-of-components
  #spec:
  #  template:
  #    apiVersion: Job
  #    kind: Image
  #    metadata:
  #      name: ...
  #    spec:
  #      ...
  #  imagePath: $(status.latestImage)$
  observedMatches:
    - input:
      output:
```

## Handling non-reconciling objects

Some objects are not reconciled and are immediately valid (for example, configmaps). How should these be handled?

1. Users could write trivial ObservedMatches. For example asserting that the `.data` field is equal to the `.data`
   field.
2. The absence of either an ObservedCompletion or an ObservedMatch could be taken as indication that the object being
   created is immediately succesful. (This seems dangerous and ill-advised)
3. An additional exclusive field (one that could be specified instead of ObservedMatches and ObservedCompletion) could
   be defined. This field could be `AlwaysSuccessful: True`

## Limitations

There are a few limitations to the current setup of observedCompletion and ObservedMatches:

1. ObservedCompletion is limited to matching a single path and value. If more than one path must be interrogated, this
   spec is not sufficient.
2. ObservedMatches cannot define a failure state.

# Possible Extensions / Open Questions

## Limit Cartographer updating when reconciliation is not complete

We must limit Cartographer from reading unless the object is in a successful state. We can additionally limit
Cartographer's updating of a resource when it is reconciling/in an unknown state. There are two use cases for this
additional limit, both revolving around inputs to an object that come in faster than the object can produce outputs.

### Avoid hidden successes

Consider object A. Two updates are made to object A in quick succession. The first succeeds and updates
the `latestGoodOutput`. But the status is still Ready:unknown, so Cartographer does not propagate this output. The
second fails and results in `Ready:false`. Cartographer will not propagate any output because `Ready:false`. This hides
the success of the first update.

### Avoid starvation during fast commits

Consider object B. A stream of updates are made to object B, each update made before the last update completes. In this
scenario, object B will remain in Ready:Unknown state. Therefore, Cartographer will not be able to read the output of
the resource. As a result, the resources in the supply chain after resource B will never receive updates. Sad
hypothetical, [will the devs owning the workload simply have to take a chocolate break](https://xkcd.com/303/)?

### Strategy

By waiting until reconciliation is complete, Cartographer can prevent the situations above (hidden success and
starvation). If every update is applied sequentially and is allowed to run to completion, there will be a steady stream
of work for downstream resources and no success will be missed.

But such a sequential strategy would be very inefficient. Users will care far more about the most recent successful
change than about every change. Instead of submitting updates from a queue, Cartographer should pop updates from a
stack. Updates should be popped from the stack until an update produces a good output. At that point the stack can be
cleared. This assures that the most recent good update results in an output, but does not waste time needlessly
calculating the output of updates that have already been superseded.

### Example

Object A recieves update 1. While processing, it receives update 2 (good), 3 (good), 4 (bad). When update 1 completes,
update 4 is applied. Because update 4 is bad, when it completes object A is in a failure state. So update 3 is popped
from the stack. This is a good update and eventually object A is in a successful state. As a result the stack of older
updates (update 2) can be cleared without ever being processed.

Note that if Cartographer is going to wait on updating while a resource reconciles, there is a danger that the resource
will get into a bad state; some infinite loop where it never exits ready:unknown. Because of that, Cartographer should
have some timeout after which it would update a resource even if it is not yet ready:true/ready:false. Note that this
does not affect the ability to match inputs with outputs, as Cartographer would still only read when Ready:true.

## Prevent Supply Chain Deadlock When Object is not Ready:True

Currently, Cartographer basks in the _eternal sunshine of the spotless mind_; each reconcile loop for resource N it can
only pass on the values it _just_ read for resource N-1, N-2, N-3... So if any of those resources are in a bad state, a
supply chain is locked. E.g. until resource N-2 outputs a value, resource N-3 that relies on that value can never be
created. Resources like our example resource A (kpack, runnable) are good actors in this system, as they constantly
report the most recent good output. After a single good input in the life of object A, they always pass a value and
never be a concern for stopping the supply chain.

If we only read status when Ready:true, even a resource reporting `latestGoodOutput` will lock downstream resources when
it is updating. If no other changes are undertaken, Cartographer would be unable to pass a value from resource A to a
downstream resource when resource A had the status Ready:false or even Ready:unknown. Thus Cartographer could not stamp
out the desired spec of the downstream resource. Without knowledge of the desired spec of the downstream resource,
Cartographer would not read the status of the downstream resource. The supply chain would grind to a halt.

### Cache outputs

When Cartographer sees a good value, it must keep a _memento_. **Each time a new output is readable on an object
(Ready:true), Cartographer must cache that value. Whenever Cartographer cannot read the value from an object
(Ready:false or Ready:unknown) Cartographer should propagate the most recent cached values to the downstream objects.**

The implementation for such a cache is thankfully proposed in
[RFC 18](https://github.com/vmware-tanzu/cartographer/pull/519). RFC 18 currently assumes that multiple artifacts from a
single object could be cached at once (in the case where a downstream object is still a child of the earlier state; that
the new state has not propagated through entire supply chain). Cartographer will need to determine which cached value to
pass to downstream objects. There is no currently proposed field that can be leveraged for this determination
(resourceVersions are not guaranteed to increase monotonically). One additional `artifact` field will be necessary to
those proposed in RFC 18, a timestamp. (Alternatively Carto could flag the most recent artifact from each resource with
a `latest` flag.)

## Allow boolean operations

- An OR condition: An OR contains a list of conditions. When any is met, the OR condition is met.
- An AND condition: A list of conditions. When all conditions are met, this condition is met. Meant primarily for
  nesting in OR conditions (as AND is the default relation of a list of conditions).
- A NOT condition: holds a condition. The NOT condition is true only when its condition is false.

## Read other objects on the cluster

It may be useful to compare the stamped object to another object on the cluster. Or to simply read a value from another
object on the cluster.

# Alternatives

The goal of this RFC is to associate an object's outputs with the inputs that led to them. The strategy above assumes
that reconciled objects report a very limited set of information:

- This status was written after the reconciler was aware of this spec.
- The reconciler has finished working.
- The most recent work was successful.

While it's not clear that there is any strategy forward relying on less information from the choreographed resources,
Cartographer could certainly ask for _more_ information. In particular, Cartographer can simply expect choreographed
resources to report on every output, "What were the relevant inputs"?

As an example, if a kpack image reported the git commit that led to the `latestImage` field in its status, Cartographer
could connect the source input that led to an image output.

## Possible implementation

An additional field would be defined on every template producing output, let us call it, `inputs-on-status`. This field
would a list of tuples. Each tuple would define 1. a path in the templating context, 2. a path on the status of the
object being templated. When the values at the object path equals the values at the templating context path, then
Cartographer will know that the outputs are the result of these inputs.

## Example

Let's assume that kpack did expose the git commit that led to the `latestImage`. Let's assume the structure were
something like

```yaml
apiVersion: kpack.io/v1alpha2
kind: Image
spec:
  ...
status:
  latestImage:
    image: index.docker.io/projectcartographer/hello-world@sha256:27452d42b
    source:
      url: https://github.com/kontinue/hello-world
      revision: 3d42c19a618bb8fc13f72178b8b5e214a2f989c4
```

Then we could template the object as before and add the new `inputs-on-status` field.

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterImageTemplate
metadata:
  name: image
spec:
  params:
    - name: image_prefix
      default: some-default-prefix-

  imagePath: .status.latestImage

  template:
    apiVersion: kpack.io/v1alpha2
    kind: Image
    metadata:
      name: $(workload.metadata.name)$
    spec:
      tag: $(params.image_prefix)$$(workload.metadata.name)$
      serviceAccountName: cartographer-example-registry-creds-sa
      builder:
        kind: ClusterBuilder
        name: go-builder
      source:
        blob:
          url: $(sources.source.url)$
          revision: $(sources.source.revision)$
      build:
        env: $(workload.spec.build.env)$

  inputs-on-status: # <--- new field
    - templating-context-path: sources.source.url
      object-path: status.latestImage.source.url
    - templating-context-path: sources.source.revision
      object-path: status.latestImage.source.revision
```

Whenever a new latest image is created, Cartographer can simply read off the fields which it cares about.

## Advantages

- Cartographer need not change its read or write cadence. It can read all outputs from an object at all times.
- Easy to extend to tracking new outputs as the result of workload changes. Since the workload is part of the templating
  context, it can be part of the tracked input set.

## Disadvantages

- Relying on resources that report inputs limits the sorts of resources Cartographer can choreograph. E.g. none of the
  resources in the example templates expose this information.
- ytt templates that switch the kind of object stamped could not use this. This is not a terrible disadvantage, as there
  is a desire to get away from using ytt to switch templates (see RFCs on [switching templates], [snippets] and
  [multipath])

# Cross References and Prior Art

- The Deployment Template currently requires either an `ObservedMatches` or `ObservedCompletion` field.
- kpack includes a [waitRules](https://carvel.dev/kapp/docs/v0.45.0/config/#waitrules) field. (hat tip
  [Scott Andrews](https://github.com/scothis))

[switching templates]: https://github.com/vmware-tanzu/cartographer/pull/75

[snippets]: https://github.com/vmware-tanzu/cartographer/pull/72

[multipath]: https://github.com/vmware-tanzu/cartographer/pull/570
