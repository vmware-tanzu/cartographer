# Meta

[meta]: #meta

- Name: Artifact Tracing with Health Rules
- Start Date: 2022-06-08
- Author(s): waciumawanjohi
- Status: Draft <!-- Acceptable values: Draft, Approved, On Hold, Superseded -->
- RFC Pull Request: (leave blank)
- Supersedes: [Read Resources Only When In Success State](https://github.com/vmware-tanzu/cartographer/pull/556)

# Summary

[summary]: #summary

Tracing allows Cartographer to determine which set of inputs (which update) of a stamped object led to a given output (
status) of the object. Some resources may explicitly provide this information, but many do not. This RFC specifies how
tracing can be accomplished when no such fields can be leveraged in the resource.

In a template's `.spec.artifactTracing` users may specify `healthRule`. This rule will the same structure as the top
level `.spec.healthRule`. The `.spec.artifactTracing.healthRule` will have an optional nested `observedGenerationPath`.
(for most use cases, the expectation is that the health rule can be copied from the top level to the artifactTracing
field with the addition of an observedGenerationPath). Only when an object is healthy, can Cartographer read the object.
Only if an object is healthy/unhealthy can Cartographer update it; that is, if an object's health is
`Unknown` the object cannot be updated.

(The specification of `healthRule` is established in the
RFC [Allow Resources to Report Status](https://github.com/vmware-tanzu/cartographer/pull/738).)

# Motivation

[motivation]: #motivation

Connecting an output of a resource to an input is necessary for establishing tracing. That is to state, "The app
currently running on the cluster is a result of resource X producing Y which was fed into resource Z which produced..."
it is necessary to tie a resource output to the input that produced it. Waiting for success/failure before updating
achieves this.

There are separate RFCs for tracing when resources report:

- a set of inputs which produced an
  output: [Artifact Tracing with Correlation Rules](https://github.com/vmware-tanzu/cartographer/pull/894)
- the spec generation that produced an
  output: [Artifact tracing with Generation Output Correlation](https://github.com/vmware-tanzu/cartographer/pull/893)

If the object provides no indication of which of a batch of updates led to the current output, the only viable strategy
is to wait until the entire batch has been processed, at which point Cartographer will know that the most recent update
led to the most recent output. This is the 'read when ready' strategy.

There is a corner case where 'Read when ready' can lead to starvation. When updates are applied to an object faster than
the object reconciles, the object will never reach a ready state. In order to avoid this, we apply a related policy 'Do
not write when unknown'. This guarantees an object will always be able to complete reconciling.

Is 'Do not write when unknown' sufficient without 'Read when ready'? Unfortunately not. An input can have multiple valid
outputs based on factors external to the supply chain. Consider a GitRepository object, the status will change based on
changes to the spec and also based on commits to a git repo unknown by Cartographer. We can call these external
triggers. Cartographer may update an object concurrent with an external trigger firing. The first output may be the
result of the external trigger. If Cartographer is not using 'Read when Ready', e.g. if it reads this value while the
object is still in an Unknown state processing the spec update, Cartographer would not know if this output belonged to
the first or second spec.

# What it is

[what-it-is]: #what-it-is

The RFC proposes adding a `healthRule` field nested in template's `.spec.artifactTracing` field. Only when this rule is
satisfied can Cartographer read the object. Only when the rule indicates either Healthy or Unhealthy can Cartographer
update the object.

---
Single condition example (likely the most common use case). Reads/Updates disallowed when specified condition is in
`Unknown` state. Reads/Updates disallowed if the value at `observedGenerationPath` does not match the value
of `metadata.generation`:

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSourceTemplate
metadata:
  name: source
spec:
  artifactTracing:
    healthRule:
      singleConditionType: Ready
      observedGenerationPath: .status.observedGeneration

  template:
    apiVersion: source.toolkit.fluxcd.io/v1beta1
    kind: GitRepository
```

Given an object in the following state, the object can be updated and read.

```yaml
apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: GitRepository
metadata:
  generation: 5
status:
  observedGeneration: 5
  conditions:
    - type: Ready
      status: "True"
```

Given an object in the following state, the object can be updated.

```yaml
apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: GitRepository
metadata:
  generation: 8
status:
  observedGeneration: 8
  conditions:
    - type: Ready
      status: "False"
```

Given an object in the following state (specified condition is neither True nor False), the object cannot be updated or
read.

```yaml
apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: GitRepository
metadata:
  generation: 3
status:
  observedGeneration: 3
  conditions:
    - type: Ready
      status: "Unknown"
```

Given an object in the following state (observedGeneration != generation), the object cannot be updated or read.

```yaml
apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: GitRepository
metadata:
  generation: 4
status:
  observedGeneration: 3
  conditions:
    - type: Ready
      status: "True"
```

---
Multi match example with both conditions and fields. Object may be updated only when

- either the healthy or unhealthy fields have been satisfied
- generation == observedGeneration

Object may be read only when

- the healthy fields have been satisfied
- generation == observedGeneration

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterTemplate
metadata:
  name: deploy
spec:
  artifactTracing:
    healthRule:
      observedGenerationPath: .status.observedGeneration
      multiMatch:
        healthy:
          matchConditions:
            - type: ReconcileSucceeded
              status: 'True'
        unhealthy:
          matchFields:
            - key: 'status.conditions[?(@.type=="ReconcileFailed")].status'
              operator: 'In'
              values: [ 'True' ]
              messagePath: '.status.usefulErrorMessage'

  template:
    apiVersion: kappctrl.k14s.io/v1alpha1
    kind: App
```

When healthy is satisfied, the object can be updated and read:

```yaml
apiVersion: kappctrl.k14s.io/v1alpha1
kind: App
metadata:
  generation: 4
status:
  observedGeneration: 4
  conditions:
    - type: ReconcileSucceeded
      status: "True"
```

When unhealthy is satisfied, the object can be updated but not read:

```yaml
apiVersion: kappctrl.k14s.io/v1alpha1
kind: App
metadata:
  generation: 6
status:
  observedGeneration: 6
  conditions:
    - type: ReconcileFailed
      status: "True"
```

Otherwise Cartographer must wait to apply updates.

---

Users may omit `observedGenerationPath`, but this is only recommended for resources which do not implement an
observedGeneration field.

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSourceTemplate
metadata:
  name: source
spec:
  artifactTracing:
    healthRule:
      singleConditionType: Ready

  template:
    apiVersion: non.compliant.com/v1
    kind: ResourceWithoutObGen
```

Given an object in the following state, the object can be updated and read.

```yaml
apiVersion: non.compliant.com/v1
kind: ResourceWithoutObGen
metadata:
  generation: 5
status:
  conditions:
    - type: Ready
      status: "True"
```

---
Users may specify `alwaysHealthy: {}` in order to specify an object that never has a state where it is in the middle of
reconciling. When an update of the object is submitted to the apiServer, the apiServer acknowledgement contains the
entire expected state of the object.

```yaml
---
apiVersion: carto.run/v1alpha1
kind: ClusterTemplate
metadata:
  name: configmap-creator
spec:
  artifactTracing:
    healthRule:
      alwaysHealthy: { }

  template:
    apiVersion: v1
    kind: ConfigMap
```

A configmap stamped out by this template can always be read and updated.

# How it Works

[how-it-works]: #how-it-works

There are five large questions:

- [How does Cartographer determine if an object is healthy, unhealthy, or unknown?](#determine-health)
- [What does Cartographer do when creating/updating an object?](#behavior-at-creationupdate)
- [How does Cartographer behave if an object is healthy?](#behavior-when-healthy)
- [How does Cartographer behave if an object is unhealthy?](#behavior-when-unhealthy)
- [How does Cartographer behave if an object is in an unknown state?](#behavior-when-in-unknown-state)

## Determine health

A stamped resource can be in one of three states:

1. 'Healthy' (status `True`)
2. 'Unhealthy' (status `False`)
3. 'Unknown' (status `Unknown`)

Readers should refer to the How it Works Section of
the [Allow Resources to Report Status RFC](https://github.com/vmware-tanzu/cartographer/pull/738) for details.

The only addition to that section is the implementation of `observedGenerationPath` in HealthRule. If the value found at
this path on the object is NOT equal to the value found at `.metadata.generation`, then the object is in an
`Unknown` state.

## Behavior at creation/update

On every reconciliation cycle, for each object in a supply chain, Cartographer currently

1. Determines the appropriate definition for the object based on the template and values from the workload and supply
   chain.
2. Assures that this definition is currently applied to the cluster.
3. Reads the object that is on the cluster and passes forward any specified fields.

This process ensures that object definition changes by an outside actor are reverted.

But in this proposal, there will be reconciliation cycles where Cartographer will not write the "appropriate definition"
because the object is still reconciling the previous definition of the object. In order to continue to protect against
rogue definitions, Cartographer will have to keep a cache of the most recent definition applied to the cluster and refer
to it before reading any values. The flow will then become:

Determine if the cached definition of the object is currently applied to the cluster.

If the cached definition of the object is NOT currently applied to the cluster, that indicates that an outside actor has
changed the object definition. Cartographer should reapply the proper definition to the cluster.

If the cached definition of the object IS currently applied to the cluster Cartographer should determine the state of
the object:

- [If it is healthy](#behavior-when-healthy)
- [If it is unhealthy](#behavior-when-unhealthy)
- [If it is unknown](#behavior-when-in-unknown-state)

## Behavior when healthy

1. Read the object that is on the cluster.
2. Pass forward any specified fields.
3. Continue with the steps in [behavior when unhealthy](#behavior-when-unhealthy)

## Behavior when unhealthy

1. Submit the most recent stamp to the cluster. The stamp is calculated from the template and the values passed from the
   workload and the supply chain.
2. Cache the submitted and returned definitions.

## Behavior when in unknown state

If the object is in unknown state Cartographer will make no update to the object.

# Migration

[migration]: #migration

The `.artifactTracing.healthRule` field is new. If this field is not specified, Cartographer will continue as otherwise
spec'd.

# Drawbacks

[drawbacks]: #drawbacks

This design complicates reading and writing, which now follow different rules.

# Alternatives

[alternatives]: #alternatives

As stated above, Cartographer could restrict only reads, rather than reads and writes. This risks starvation.
Cartographer could restrict only writes, but this presents corner cases where incorrect behavior would be exhibited.

# Prior Art

[prior-art]: #prior-art

One might say this is a similar approach as kpack, which creates one build at a time. When a kpack image is created,
kpack will start a build. As updates are made to the image, no new image is created until the build in process succeeds
or fails. At that time the most recent definition leads to the creation of a new build object.

# Unresolved Questions

[unresolved-questions]: #unresolved-questions

- Cartographer currently makes no distinction between updates because the workload changed, the template changed, the
  supply chain definition (e.g. a param) changed, or a value from earlier in the supply chain changed. Should
  Cartographer continue to remain agnostic? Or should redefinitions of Carto objects cause stamped objects to be deleted
  and applied anew?
- An object that enters into an eternal loop could never be updated by Carto under this strategy. Should there be a
  timeout field?

# Spec. Changes (OPTIONAL)

[spec-changes]: #spec-changes

- All templates will now have an optional `artifactTracing.healthRule` which will take have an optional
  `observedGenerationPath` field and a required one of three fields {alwaysHealthy|singleConditionType|multiMatch}:

```yaml
apiVersion: carto.run/v1alpha1
kind: Cluster[Config|Deployment|Image|Source]Template
spec:
  artifactTracing:
    healthRule:
      observedGenerationPath: <string>
      alwaysHealthy: { }
      singleConditionType: <string>
      multiMatch:
        healthy:
          matchConditions:
            - type: <string>
              status: <string>
          matchFields:
            - key: <string>
              operator: <[In|NotIn|Exists|DoesNotExist]>
              values: [ <string> ]
              messagePath: <string>
        unhealthy:
          matchConditions:
            - type: <string>
              status: <string>
          matchFields:
            - key: <string>
              operator: <[In|NotIn|Exists|DoesNotExist]>
              values: [ <string> ]
              messagePath: <string>
```
