# Meta
[meta]: #meta
- Name: Common Blueprint Architecture
- Start Date: 2022-03-29
- Author(s): @emmjohnson @martyspiewak
- Status: Draft
- RFC Pull Request: (leave blank)
- Supersedes: N/A

# Summary
[summary]: #summary

This RFC proposes collapsing all current "Blueprints" and templates (not including `ClusterRunTemplate`) into a single CRD, `ClusterBlueprint`.
As part of this, the RFC proposes that outputs be specified by template authors and remove the typed outputs that exist today (url, revision, image, config).
Additionally, the `ClusterBlueprint` will take an optional selector so that operators can decide whether the blueprint should be resolved directly against an Owner or only as part of a larger Blueprint.


# Motivation
[motivation]: #motivation

- Why should we do this?

  - Today we have seven CRDs that are needed to create a supply chain and delivery. To simplify the data model for users, we can
collapse all seven CRDs into one CRD. Additionally, whereas today the overhead to start reconciling templates would be adding five identical
reconcilers, this would enable us to have one reconciler for all blueprints and more easily add meaningful data to their statuses.

- What use cases does it support?

  - This would allow for snippets and users to specify any output they want.

- What is the expected outcome?

  - Simplified onboarding and user experience.

# What it is
[what-it-is]: #what-it-is

We have two options that we are presenting, but the goal is to agree on one before fleshing out the complete details of the RFC.

Option 1
```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterBlueprint
metadata:
  name:
spec:
  # optional
  selector:

  resources:
    - name: first-resource
      inputs:
        - resource: zero-resource
          output: zoo
          name: zoo-resource-name
      inline:
        # mutually exclusive - template, ytt
        template: {}
        ytt: {}

        outputs:
          - name: foo
            valuePath: .status.url

    - name: second-resource
      # mutually exclusive - blueprintName, options
      options:
        - name: another-blueprint
        - name: diff-blueprint
      inputs:
        - resource: first-resource
          output: foo
          name: a-new-resource-name
        - resource: zero-resource
          output: zoo
          name: a-new-resource-name

    - name: third-resource
      # mutually exclusive - blueprintName, options
      blueprintName: another-another-blueprint
      inputs:
        - resource: first-resource
          output: foo
          name: a-new-resource-name
        - resource: second-resource
          output: bar
          name: another-new-name

  outputs:
    - name: baz
      valuePath: resources[first-resource].outputs[foo]
    - name: fizz
      valuePath: resources[second-resource].outputs[bar]
```

Option 2
```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterBlueprint
metadata:
  name:
spec:
  # optional
  selector:

  # mutually exclusive - resources, template, ytt (i.e. a blueprint is either a template or a supply chain)
  resources:
    - name: first-resource
      # mutually exclusive - blueprintName, options
      blueprintName: another-blueprint
    - name: second-resource
      # mutually exclusive - blueprintName, options
      options:
        - name: another-another-blueprint
      inputs:
        - resource: first-resource
          output: foo
          name: a-new-resource-name

  template: {}

  ytt: {}

  outputs:
    - name: url
      valuePath: .status.url
```

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
  - Typed outputs
    ```yaml
    outputs:
      urlPath: <string>
      revisionPath: <string>
      configPath: <string>
      imagePath: <string>
    ```
  - Typed blueprints (`ClusterSourceBlueprint`, `ClusterImageBlueprint`, etc.)

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
