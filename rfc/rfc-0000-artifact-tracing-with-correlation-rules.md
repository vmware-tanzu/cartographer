# Meta

[meta]: #meta

- Name: Artifact Tracing with Correlation Rules
- Start Date: 2022-06-09
- Author(s): waciumawanjohi
- Status: Draft <!-- Acceptable values: Draft, Approved, On Hold, Superseded -->
- RFC Pull Request:
- Supersedes: [Input-Output correlation](https://github.com/vmware-tanzu/cartographer/pull/799)

# Summary

[summary]: #summary

Artifact Tracing requires Cartographer to determine which set of inputs (which update) of a stamped object led to a
given output (status) of the object. This RFC allows such a determination if the templated object reports some inputs
alongside outputs. Template authors write a set of correlation rules pointing to these reported inputs.

There are separate RFCs for correlation for other types of resources:

- [a resource which reports an observed-generation alongside outputs](https://github.com/vmware-tanzu/cartographer/pull/893)
- [a resource that provides no information about inputs with outputs](https://github.com/vmware-tanzu/cartographer/pull/891)

# Motivation

[motivation]: #motivation

Cartographer has received consistent feedback that users are interested in historical runs. A pre-requisite to providing
historical runs is to be able to provide a current snapshot of what has happened to an artifact (e.g. a source commit or
a given image) in the current run. We can call this artifact tracing. In order to establish artifact tracing,
Cartographer must be able to determine for an individual object which update led to a current output. This determination
can be easier or harder depending on what information is written to the status of an object. This RFC handles the
specific case where the status reports an output along with fields in the input that led to that output.

# What it is

[what-it-is]: #what-it-is

## Definitions

### artifact tracing

the ability to determine which update of the object definition led to the presence of a particular field in the object
status. e.g. after submitting 5 updates to the definition of a kpack Image object, an observer sees the `latestImage`
field in the status of the Image object. When the observer can definitively say which update was responsible for that
latestImage, the observer has achieved input-output correlation.

### output

the value at the `imagePath`, `configPath`, `urlPath` or `revisionPath` on a templated object in the cluster.

### stamp context

the fields available to a template from the workload and supply chain. This includes fields like `params`, values on the
workload like `workload.metadata.name` and values from earlier steps in the supply chain like `sources.first-source.url`

### input

some individual value from the templating context that was used to create a templated object

## Examples

A template is written with a set of correlation rules.

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterImageTemplate
metadata:
  name: image-creator
spec:
  imagePath: .status.latestGoodImage.location
  template:
    apiVersion: my-image-resource/v1
    kind: ImageMaker
    metadata:
      name: $(workload.metadata.name)$
    spec:
      scanUrl: "$(sources.source.url)$"
      sourceRevision: "$(sources.source.revision)$"
  artifactTracing:
    correlationRules:
      - stampContextPath: sources.source.url                      #evaluated against inputs in template context
        objectStatusPath: .status.latestGoodImage.sourceUrl       #evaluated against the in-cluster resource
      - stampContextPath: sources.source.revision                 #evaluated against inputs in template context
        objectStatusPath: .status.latestGoodImage.sourceRevision  #evaluated against the in-cluster resource
```

Assume at a given time the object on the cluster is:

```yaml
apiVersion: my-scanner/v1
kind: SourceScan
metadata:
  name: ...
status:
  latestGoodImage:
    location: ghcr.io/some-project/some-repo:xyz
    sourceUrl: https://github.com/some-project/some-repo
    sourceRevision: abc123
```

Then Cartographer will be able to associate the image output `ghcr.io/some-project/some-repo:xyz` with the source input
url `https://github.com/some-project/some-repo` and revision `abc123`.

# How it Works

[how-it-works]: #how-it-works

When a template author specifies `correlationRules`, Cartographer will associate the outputs with the proper set of
previous inputs:

When Cartographer stamps an object, it uses a stamping context to fill templated fields in a template. From the example
above, something like the following object was submitted:

```yaml
apiVersion: my-image-resource/v1
kind: ImageMaker
metadata:
  name: app
spec:
  scanUrl: https://github.com/some-project/some-repo
  sourceRevision: abc123
```

At some later point, Cartographer will read the object on the cluster. At that point there will be some output
(from the example above `ghcr.io/some-project/some-repo:xyz`). There will also be a set of input values at paths
determined in the correlationRules (e.g. the value `abc123` and `https://github.com/some-project/some-repo`).
Cartographer relies on this information to establish that the output is the result of the indicated inputs.

# Migration

[migration]: #migration

This is additive work and has no implications for migration. If users do not specify
a `.spec.artifactTracing.correlationRules` field, no changes from current behavior will be observed.

# Drawbacks

[drawbacks]: #drawbacks

- This would represent an additional manner of determining the source of outputs. Documentation will be harder for that.
- This hypothesizes how resource authors may choose to write their resources. If we are incorrect, that will be wasted
  effort on our part.

# Alternatives

[alternatives]: #alternatives

As mentioned in [summary], there are two other proposals for correlating
outputs. [Artifact Tracing with Health Rules](https://github.com/vmware-tanzu/cartographer/pull/891) is general
purpose (though it entails a performance penalty on Cartographer) and could be used if this proposal is not adopted.

# Prior Art

[prior-art]: #prior-art

This RFC draws heavily on the RFC [Input-Output correlation](https://github.com/vmware-tanzu/cartographer/pull/799). An
important difference is the matter of caching. In the earlier RFC, the currently stamped input is compared to the
currently read output. If the values match, the output is cached and passed forward. If the values do not match, the
cached value is passed forward. This has the effect of waiting for the object to finish reconciling the most recent
commit before passing forward (or caching) an output. This can lead to starvation of the supply chain in the case that
the object is updated at a rate faster than it can reconcile.

By contrast, in this RFC no caching is done and reads/writes are never held.

# Unresolved Questions

[unresolved-questions]: #unresolved-questions

- Once Cartographer determines that a given input caused a given output, how will that information be conveyed? RFC
  [Workload Report Artifact Provenance](https://github.com/vmware-tanzu/cartographer/pull/519) addresses this question.

# Spec. Changes (OPTIONAL)

[spec-changes]: #spec-changes

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

    # Correlation Rules specify template values that are written into the spec
    # of an object and where those values are expected to be found in the
    # status of the reconciled object
    correlationRules:
      - # The path in the stamp context used as an input for the template
        # example: .workload.metadata.name
        stampContextPath: <string>
        # The path in the reconciled object where the above value can be found.
        # This path should begin with `.status`
        objectStatusPath: <string>
```
