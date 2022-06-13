# Meta
[meta]: #meta
- Name: Constrained Resources
- Start Date: 2022-06-13
- Author(s): squeedee
- Status: Draft 
- RFC Pull Request: (leave blank)
- Supersedes: (put "N/A" unless this replaces an existing RFC, then link to that RFC)

# Summary
[summary]: #summary

Provide a mechanism for multiple artifact inputs to a template that are constrained, meaning that intermediate artifacts
were triggered by the same upstream artifact.

# Motivation
[motivation]: #motivation

A few examples now exist of Supply Chain authors wanting to pass the source reference or the specific source revision to
the configuration writer. For example attaching a source reference to the gitops push. 

A possible supply chain design that will let DevOps ensure an accepted build is constantly patch, is one where the
workload is patched with a specific source revision, and the workload is pushed to the gitops repo, along with the config.
In this way, the workload can be applied to a "maintenance" namspace, re-patching a specific software revision.

No matter the case, user's assume that taking two inputs leads to a consistent, re-entrant result at their template
step.

The current architecture works like this:



# What it is
[what-it-is]: #what-it-is

This provides a high level overview of the feature.

- Define any new terminology.
- Define the target persona: application developer, supply chain/delivery author, template author, and/or project contributor.
- Explaining the feature largely in terms of examples.

# How it Works
[how-it-works]: #how-it-works

This is the technical portion of the RFC, where you explain the design in sufficient detail.

The section should return to the examples given in the previous section, and explain more fully how the detailed proposal makes those examples work.

# Migration
[migration]: #migration

This section should document breaks to public API and breaks in compatibility due to this RFC's proposed changes. In addition, it should document the proposed steps that one would need to take to work through these changes. Care should be give to include all applicable personas, such as application developers, supply chain/delivery authors, and template authors.

# Drawbacks
[drawbacks]: #drawbacks

Why should we *not* do this?

# Alternatives
[alternatives]: #alternatives

- What other designs have been considered?
- Why is this proposal the best?
- What is the impact of not doing this?

# Prior Art
[prior-art]: #prior-art

Discuss prior art, both the good and bad.

# Unresolved Questions
[unresolved-questions]: #unresolved-questions

- What parts of the design do you expect to be resolved before this gets merged?
- What parts of the design do you expect to be resolved through implementation of the feature?
- What related issues do you consider out of scope for this RFC that could be addressed in the future independently of the solution that comes out of this RFC?

# Spec. Changes (OPTIONAL)
[spec-changes]: #spec-changes
Does this RFC entail any proposed changes to CRD specs? If so, please document changes here.
This section is not intended to be binding, but as discussion of an RFC unfolds, if spec changes are necessary, they should be documented here.
