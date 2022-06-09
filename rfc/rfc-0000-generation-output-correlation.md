# Meta

[meta]: #meta

- Name: Artifact Tracing with Generation-Output Correlation
- Start Date: 2022-06-06
- Author(s): waciumawanjohi
- Status: Draft <!-- Acceptable values: Draft, Approved, On Hold, Superseded -->
- RFC Pull Request:
- Supersedes: N/A

# Summary

[summary]: #summary

Cartographer creates objects and updates their definitions. Other controllers reconcile those objects, in effect
creating outputs from the definitions. At present, it is not possible for Cartographer to determine which definition led
to which output. This RFC hypothesizes that resources may decide to report the generation of the object that led to an
output. Cartographer should be able to handle that pattern.

# Motivation

[motivation]: #motivation

There exist two proposals for correlating the output of a resource to the input that led to it:

- [Artifact Tracing with Correlation Rules](https://github.com/vmware-tanzu/cartographer/pull/892) specifies how
  Cartographer can reason about resources which report a set of inputs in their outputs. (This
  supersedes [Input-Output Correlation](https://github.com/vmware-tanzu/cartographer/pull/799))
- [Artifact Tracing with Health Rules](https://github.com/vmware-tanzu/cartographer/pull/891) specifies how Cartographer
  can reason about resources which provide no such information. (This
  supersedes [Read Resources Only When In Success State](https://github.com/vmware-tanzu/cartographer/pull/556))

But there is a third, reasonable set of resources to consider: those which report the generation of the object
definition that led to the outputs.

Reasonable question, "Isn't this the same as Input-Output Correlation?" No, for three reasons.

1. Input-Output Correlation will require Cartographer end users to specify correlation rules, associating particular
   fields in the object status with fields stamped by the user. Such correlation rules are not necessary when
   correlating a reported generation.
2. The implementation of input-output correlation will differ from the implementation of generation-output correlation.
   The generation of an object is not an input in the definition Cartographer stamps, but rather a value that
   Cartographer would learn upon submitting a definition to the apiServer.
3. This proposal assumes a different model of resource authors. Input-Output Correlation assumes that resource authors
   will choose some subset of the spec of an object as being "actually" important in the creation of outputs, with other
   fields in the spec being non-essential. It is not clear that this model is rational. While a particular template in
   cartographer may only vary a subset of fields, we should assume that all fields are meaningful from the resource
   author's standpoint. Rather than reporting all of these meaningful fields in their output, these authors would more
   reasonably report generation, which corresponds to a particular tuple of all the inputs.

# What it is

[what-it-is]: #what-it-is

## Definitions

### artifact tracing

[artifact tracing]: #artifact-tracing
the ability to determine which update of the object definition led to the presence of a particular field in the object
status. e.g. after submitting 5 updates to the definition of a kpack Image object, an observer sees the `latestImage`
field in the status of the Image object. When the observer can definitively say which update was responsible for that
latestImage, the observer has achieved artifact tracing.

### output generation

[output generation]: #output-generation
this refers to an `observedGeneration` value that is tied to a particular field in an object's status. To be clear, this
is almost certainly not the status' top level `observedGeneration` field, as the conventions around its use do not meet
the expectations here. An example would be the `observedGeneration` field in `metav1.Conditions`, where each condition
may have its own `observedGeneration` pointing to the generation that first caused its current value. If such a
condition were an output, we would call the `observedGeneration` field an "
output generation".

### stamp

[stamp]: #stamp
an object definition. Cartographer stamps an object based on a template and inputs from a workload and previous steps in
a supply chain. This stamp is then submitted to the cluster.

### output path

[output path]: #output-path
the path specified in the template where an object's output can be found. This is a generic term for
`imagePath`, `configPath` and the pair `urlPath` and `revisionPath`.

### monotonically increasing

[monotonically increasing]: #monotonically-increasing
A series which always increases or remaining constant, and never decreases.

- Example: 1, 2, 3, 3, 6, 7
- Counter Example: 1, 2, 5, 4

## Example

Template authors will be able to specify the field in an object that corresponds to the output generation and thereby
enable artifact tracing. In other words, given a resource with output generations (that reports outputs alongside the
spec generation that led to the output), users will be able to write a template that specifies the location of this
generation field and to enable artifact tracing* in Cartographer.

Let us hypothesize a new kpack image which reports its latest image in the following manner:

```yaml
status:
  ...
  latestImage:
    imageLocation: some-image-uri
    observedGeneration: 6
```

Template authors will be able to write a template thusly:

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterImageTemplate
metadata:
  name: image
spec:
  imagePath: .status.latestImage.imageLocation
  artifactTracing:
    observedGenerationPath: .status.latestImage.observedGeneration

  template:
    apiVersion: kpack.io/v1alpha2
    kind: Image
    ...
```

# How it Works

[how-it-works]: #how-it-works

## Nest observedGenerationPath in artifactTracing field

There have been two proposals for artifact tracing. Previous discussions have led to a growing consensus that both will
be valuable. One innovation in this proposal is the template field `artifactTracing`, under which any of the proposed
new fields should nest. E.g. the `correlationRules` field from
the [Input-Output Correlation RFC](https://github.com/vmware-tanzu/cartographer/pull/799) and the
`observedGenerationPath` from this proposal will both be fields nested under the `artifactTracing` field. Any additional
approach adopted for artifact tracing should be similarly nested.

## Implementation

When a template author specifies an `observedGenerationPath`, Cartographer will determine what input led to the
specified outputs:

When Cartographer stamps an object, it fills fields in a template with values from the workload and supply chain. Upon
submitting this object to the cluster, the apiServer will return the new object definition to Cartographer. This
definition will include `.metadata.generation`. Cartographer will be responsible for caching this generation value along
with the inputs that were used in the stamp.

When reading an object, Carto will read both the value at the output path and the value at
the `.artifactTracing. observedGenerationPath`. Cartographer will look up the reported generation in the cache to
determine which inputs were involved in the creation of that generation definition. Cartographer will then report to
users that these inputs led to the given output.

### Not described: reporting information to users

When a template author specifies an `observedGenerationPath`, Cartographer will report what input led to a given output
for each object in a supply chain. The implementation of that reporting (where and how that reporting is done) will be
handled in a separate RFC. For an idea of what that might look like, the reader is directed
to [Workload Report Artifact Provenance](https://github.com/vmware-tanzu/cartographer/pull/519).

### Caching

There are broadly three approaches to caching such a value:

1. Write the value in the workload status.
2. Write the value in some other object on the cluster.
3. Write the value in some database to which Cartographer has credentials.

These lead to a variety of concerns including the size limit of the workload object (1MB), managing objects on the
cluster (naming, garbage collecting, etc), deploying and managing a service on the cluster (given an on cluster
database) or documenting required external dependencies (given a database outside of the server).

Balancing these concerns, Cartographer should write the generation and input values for each stamp submitted to the
cluster in a different object in the cluster. This is intermediary information and not necessary to expose to users.
Implementers of this RFC are welcome to determine if 1 such data object per Carto controller, 1 per supply chain, 1 per
step/object or 1 per stamp is most reasonable. (Presumably one of the middle approaches will win out).

#### Garbage collection of the cache

Cartographer may assume that object status reflects [monotonically increasing] generations. Therefore, when output
generation N has been observed, the cache of all input-generation tuples for gen 1 to N-1 may be discarded. (i.e. no
earlier generation cached value will be seen or used).

# Migration

[migration]: #migration

This is additive work and has no implications for migration.

# Drawbacks

[drawbacks]: #drawbacks

- This would represent an additional manner of determining the source of outputs. Documentation will be harder for that.
- This hypothesizes how resource authors may choose to write their resources. If we are incorrect, that will be wasted
  effort on our part.

# Alternatives

[alternatives]: #alternatives

As mentioned in [motivation](#motivation), there are two other proposals for correlating
outputs. [Read Resources Only When In Success State](https://github.com/vmware-tanzu/cartographer/pull/556) is general
purpose (though it entails a performance penalty on Cartographer) and could be used if this proposal is not adopted.

# Prior Art

[prior-art]: #prior-art

## observedGeneration

Reporting conditions as part of outputs is a growing pattern. There is a convention to report `observedGeneration`
at the top level of an object's status (to be clear, this is not a value that can be relied upon in this proposal). In
addition, `metav1.Conditions` contain an optional field `ObservedGeneration`. They are described as such:
> observedGeneration represents the .metadata.generation that the condition was set based upon.
> For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the
> condition is out of date
> with respect to the current state of the instance.

It is reasonable to suggest to resource authors that as they implement observedGeneration in of their conditions, so
should they implement an observedGeneration of their outputs.

## artifact tracing

There have been two proposals for artifact tracing. Previous discussions have led to a growing consensus that both will
be valuable. One innovation in this proposal is the template field `artifactTracing`, under which any of the proposed
new fields should nest.

# Unresolved Questions

[unresolved-questions]: #unresolved-questions

- How will Cartographer report the connection of inputs and outputs to end
  users? [Read more](#not-described-reporting-information-to-users)
- How will Cartographer cache? [Read more](#caching)
- How will Cartographer handle resources that do not report an output generation?

# Spec. Changes (OPTIONAL)

[spec-changes]: #spec-changes

All templates will have an additional top level field in their spec `artifactTracing` with the nested field
`observedGenerationPath`.

e.g.

```yaml
---
apiVersion: carto.run/v1alpha1
kind: ClusterSourceTemplate
metadata: { }
spec:
  # Artifact Tracing is the behavior of Cartographer reporting which inputs
  # for an object were responsible for a particular output.
  # In order to enable this behavior, a subfield of Artifact Tracing must be
  # specified.
  artifactTracing:

    # Observed Generation Path is the path on the object where an output's
    # observed generation is observed.
    # Note: it is very unlikely that .status.observedGeneration is the correct
    # value for this field. Likely candidates will be observedGeneration fields
    # on subfields of the status.
    observedGenerationPath: <string>
```
