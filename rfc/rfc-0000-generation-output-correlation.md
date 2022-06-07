# Meta

[meta]: #meta

- Name: Generation-Output Correlation
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

- [Input-Output Correlation](https://github.com/vmware-tanzu/cartographer/pull/799) specifies how Cartographer can
  reason about resources which report a set of inputs in their outputs.
- [Read Resources Only When In Success State](https://github.com/vmware-tanzu/cartographer/pull/556) specifies how
  Cartographer can reason about resources which provide no such information.

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



This provides a high level overview of the feature.

- Define any new terminology.
- Define the target persona: application developer, supply chain/delivery author, template author, and/or project
  contributor.
- Explaining the feature largely in terms of examples.

# How it Works

[how-it-works]: #how-it-works

This is the technical portion of the RFC, where you explain the design in sufficient detail.

The section should return to the examples given in the previous section, and explain more fully how the detailed
proposal makes those examples work.

# Migration

[migration]: #migration

This is additive work and has no implications for migration.

# Drawbacks

[drawbacks]: #drawbacks

- This would represent a third manner of determining the source of outputs. Documentation will be harder for that.
- This hypothesizes how resource authors may choose to write their resources. If we are incorrect, that will be wasted
  effort on our part.

# Alternatives

[alternatives]: #alternatives

As mentioned in [motivation](#motivation), there are two other proposals for correlating
outputs. [Read Resources Only When In Success State](https://github.com/vmware-tanzu/cartographer/pull/556) is general
purpose (though it entails a performance penalty on Cartographer) and could be used if this proposal is not adopted.

# Prior Art

[prior-art]: #prior-art

Reporting conditions as part of outputs is a growing pattern. There is a convention to report `observedGeneration`
at the top level of an object's status (to be clear, this is not a value that can be relied upon in this proposal). In
addition, `metav1.Conditions` contain an optional field `ObservedGeneration`. They are described as such:
> observedGeneration represents the .metadata.generation that the condition was set based upon.
> For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the
> condition is out of date
> with respect to the current state of the instance.

It is reasonable to suggest to resource authors that as they implement observedGeneration in of their conditions, so
should they implement an observedGeneration of their outputs.

# Unresolved Questions

[unresolved-questions]: #unresolved-questions

- What parts of the design do you expect to be resolved before this gets merged?
- What parts of the design do you expect to be resolved through implementation of the feature?
- What related issues do you consider out of scope for this RFC that could be addressed in the future independently of
  the solution that comes out of this RFC?

# Spec. Changes (OPTIONAL)

[spec-changes]: #spec-changes
Does this RFC entail any proposed changes to CRD specs? If so, please document changes here. This section is not
intended to be binding, but as discussion of an RFC unfolds, if spec changes are necessary, they should be documented
here.
