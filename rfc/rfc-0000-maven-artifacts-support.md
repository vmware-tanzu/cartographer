# Meta
[meta]: #meta
- Name: Maven Artifacts Support
- Start Date: 2022-03-30
- Author(s): @emmjohnson
- Status: Draft
- RFC Pull Request: (leave blank)
- Supersedes: N/A

# Summary
[summary]: #summary

Modify `owner.spec.source` to allow for specifying a JAR file.

# Motivation
[motivation]: #motivation

- Why should we do this?
  - Today we support workloads being built from git and image source. Many developers in the field also need support
    Maven and Artifactory repositories hosting their JAR and WAR files.
- What use cases does it support?
  - Maven and Artifactory repositories
- What is the expected outcome?
  - To enable more types of workloads.

# What it is
[what-it-is]: #what-it-is
We need to modify the Owner specs to create an unambiguous way for specifying a JAR file. There are two proposed
options, the goal is to get community buy-in on a solution before ratifying and the RFC will then be modified to suggest
the agreed upon spec.

This option allows for the developer to know what is necessary for a maven workload by examining the spec.
```yaml
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  name: my-workload
spec:
  source:
      maven: 
        url: https://maven.pkg.jetbrains.space/mycompany/p/projectkey/my-maven-repo

---
apiVersion: carto.run/v1alpha1
kind: ClusterSupplyChain
metadata:
  name: my-supply-chain
spec:
  selectorMatchFields:
    - key: spec.maven.url
      operator: Exists
  resources:
    - name: source-provider
      templateRef:
        name: source-template
        kind: ClusterSourceTemplate
```

This option allows us to support many artifacts, but the platform operator would have to communicate to the developer what parameter to use. 
```yaml
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  name: my-workload
spec:
  params:
    - name: source-type
      value: maven
  source:
      url: https://maven.pkg.jetbrains.space/mycompany/p/projectkey/my-maven-repo

---
apiVersion: carto.run/v1alpha1
kind: ClusterSupplyChain
metadata:
  name: my-supply-chain
spec:
  selectorMatchFields:
    - key: spec.params[?(@.name=="source-type")].value
      operator: In
      values:
        - "maven"
  resources:
    - name: source-provider
      templateRef:
        name: source-template
        kind: ClusterSourceTemplate
```

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
  - We did consider using options to select a template for a generic spec.source.url, however
    we don't believe matchFields will be able to distinguish between different repository types.
    That information tends to be hidden inside the particular protocol of the repository and not the URL.

- Why is this proposal the best?
- What is the impact of not doing this?

# Prior Art
[prior-art]: #prior-art

Discuss prior art, both the good and bad.

# Unresolved Questions
[unresolved-questions]: #unresolved-questions
- Do we need any additional information from the Owner besides the url of the Artifact?

# Spec. Changes (OPTIONAL)
[spec-changes]: #spec-changes
Does this RFC entail any proposed changes to CRD specs? If so, please document changes here.
This section is not intended to be binding, but as discussion of an RFC unfolds, if spec changes are necessary, they should be documented here.
