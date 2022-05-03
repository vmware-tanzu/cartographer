# Meta
[meta]: #meta
- Name: Allow developers to provide service configuration for builds
- Start Date: 2022-05-02
- Author(s): [squeedee](https://github.com/squeedee)
- Status: Draft
- RFC Pull Request: (leave blank)
- Supersedes: "N/A"

# Summary
[summary]: #summary

Just as developers can provide environment variables for build time, they need the ability to provide service bindings
as well (dependent on how the supply chain and templates are designed)

As a rule-of-thumb, build configuration should be provided on a per-supply-chain basis, yet some users have clear
requirements for workload-level configuration.

# Motivation
[motivation]: #motivation

Some resources that might be used in a supply chain can only access secrets (especially artifact repository credentials)
using a service binding. Quite often these need to be provided on a per-workload basis.

# What it is
[what-it-is]: #what-it-is

Two alternatives for providing services that bind at build time.

## Alternative: match current env design

Workloads currently have environment variables for build and runtime, so we could add a serviceClaims key to `spec.build`:

```yaml
spec:
  env:                  # Runtime ENV vars (existing)
    - name: VAR_NAME
      value: var_value
  serviceClaims:        # Runtime service claims (existing)
    - name: db
      ref: <database service ref>
  build:
    env:                # Build time ENV vars (existing)
      - name: VAR_NAME
        value: var_value
    serviceClaims:      # build time service claims !! NEW, proposed by this RFC !!
        - name: artifactory
          ref:
            apiVersion: v1
            kind: Secret
            name: artifactory_maven_repo
```

## Alternative: use labels to avoid 




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
