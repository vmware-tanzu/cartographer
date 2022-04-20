# Meta

[meta]: #meta

- Name: Input-Output correlation
- Start Date: 2022-04-07
- Author(s): idoru
- Status: Draft
- RFC Pull Request: (leave blank)
- Supersedes:
  [Draft RFC Read Resources Only When In Success State](https://github.com/vmware-tanzu/cartographer/pull/556)
- Depends on: [Allow Resources to Report Status](https://github.com/vmware-tanzu/cartographer/pull/738)

# Summary

[summary]: #summary

Cartographer may stamp out arbitrary resources in servicing a `Workload` or `Deliverable`. These resources produce
output artifacts, and in, turn, may also consume, as inputs, the outputs from other resources.

As of this writing, there is no mechanism to assert which input artifacts are implicated for a given output.

This RFC introduces mechanism which allows template authors to provide additional context on the stamped resources which
Cartographer can use to make such assertions possible.

It borrows heavily on prior art
in [Draft RFC Read Resources Only When In Success State](https://github.com/vmware-tanzu/cartographer/pull/556)

Part of the mechanics requires [Allow Resources to Report Status](https://github.com/vmware-tanzu/cartographer/pull/738)
or a similar mechanism.

# Motivation

[motivation]: #motivation

## Why should we do this?

It is desirable to know which inputs yielded an output, as this knowledge is central to tracing the preceding artifact
inputs which ultimately result in said output.

While the necessary information to do so may exist on the stamped resources, and thus could be inferred from inspecting
the resource, Cartographer does not provide an agnostic means to leverage this. This means users must delve into these
resources and make these inferences themselves. Doing so requires more explicit knowledge of exploring resources, is
laborious, and error-prone.

While this RFC does not directly propose UX changes to enable this, it describes mechanisms that enable Cartographer to
internally recognize which inputs yield a given output, which can be used as the basis for UX changes.

## What use cases does it support?

The changes proposed here do not surface the correlation between inputs and outputs, but they introduce mechanics which
will enable building features which do so.

An interesting side effect of these mechanics is that Cartographer will not propagate artifacts output by a resource to
the next resource unless they can be correlated to input provided to the resource by Cartographer. This is desirable
from a security perspective, and the outcomes of this behavior are explored in detail in the
[Draft RFC Read Resources Only When In Success State](https://github.com/vmware-tanzu/cartographer/pull/556).

# What it is

[what-it-is]: #what-it-is

New fields are added to templates which allow template authors to specify correlation rules using `correlationRules`,
which informs Cartographer of how to compare the resource's output (published in the owner's status) with its inputs
(based on where the input values were substituted into the resource stamped bt template). Where the output does not
expose useful matching information, it is possible to match against the resource's `spec` instead, with some caveats.

Template authors are thought to be the ideal target, as they are likely to have the most intimate knowledge of the
resource that their templates stamp out.

Example matching inputs against outputs with the `correlationRules` field:

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSourceTemplate
metadata:
  name: example-source-scanning-thing
spec:
  template:
    apiVersion: my-scanner/v1
    kind: SourceScan
    metadata:
      name: ...
    spec:
      scanUrl: "$(sources.source.url)$"
      sourceRevision: "$(sources.source.url)$"
  correlationRules:
  - expectedValue: "$(sources.source.url)$"        #evaluated against inputs in template context
    actualPath: ".status.outputs.url"              #evaluated against the in-cluster resource
  - expectedValue: "$(sources.source.revision)$"   #evaluated against inputs in template context
    actualPath: ".status.outputs.scannedRevision"  #evaluated against the in-cluster resource
```

# How it Works

[how-it-works]: #how-it-works

When Cartographer processes a resource, the jsonpath expressions in each `correlationRules` `input` and `output` fields
are evaluated and compared. If they are all equal, then the output is considered to be correlated with the input(s).

When output cannot be correlated to input, Cartographer will not propagate the output forward. Doing so ensures we can
always correlate outputs with inputs on subsequent resources, and also mitigates behavior of bad actors in the cluster
as described
in [Draft RFC Read Resources Only When In Success State](https://github.com/vmware-tanzu/cartographer/pull/556).

In the absence of `correlationRules` Cartographer will continue to propagate outputs without attempting to correlate
inputs to preserve existing functionality.

For cases where the output does not contain necessary information to match against inputs, one may instead match against
the spec of the resource:

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterImageTemplate
metadata:
  name: example-image-building-thing
spec:
  template:
    apiVersion: my-builder/v1
    kind: ImageBuilderThatDoesntExposeInputDetailsInItsOutput
    metadata:
      name: ...
    spec:
      sourceUrl: "$(sources.source.url)$"
      sourceRevision: "$(sources.source.url)$"
observesGeneration: true                         #signifies resource complies with observedGeneration pattern
correlationTimeout: 600                          #maximum timeout in seconds
correlationRules:
  - expectedValue: "$(sources.source.url)$"      #evaluated against inputs in template context
    actualPath: ".spec.sourceUrl"                #evaluated against the in-cluster resource
  - expectedValue: "$(sources.source.revision)$" #evaluated against inputs in template context
    actualPath: ".spec.sourceRevision"           #evaluated against the in-cluster resource
```

Since the spec of the resource may change independently of the output, when the `correlationRules` makes any assertion
against the `spec`, Cartographer must also ensure that the current output is a result of processing the current spec by
making additional assertions:

* Resource is healthy, as per [Allow Resources to Report Status](https://github.com/vmware-tanzu/cartographer/pull/738)
* Resource `observedGeneration == generation`

Template authors must explicitly signal to Cartographer that the resource's behavior cooperates with the assertion that
the combination of these facts is proof that output extracted from the resource is generated by its current generation
spec. This is set with the `observesGeneration` property.

Additionally, relying on the spec to make assertions imposes a further restriction to avoid stalling artifact
propagation: Cartographer must not stamp out any further changes to the resource until it has finished processing the
current spec completely. That is, it must have reached either a Healthy or Unhealthy status (as per
[Allow Resources to Report Status](https://github.com/vmware-tanzu/cartographer/pull/738)).

To prevent halting when a resource never reaches either of these statuses, a `correlationTimeout` property can specify
the maximum time to wait, after which Cartographer assumes the resource has reached the Unhealthy status. If this is
accepted, exposing a mechanism to override timeouts per resource in owner blueprints will likely be necessary.

# Migration

[migration]: #migration

Templates will continue to work as they do today, but `correlationRules` must be added if input-output correlation is
desired for the resources they will stamp out.

Any template containing `correlationRules` referencing the resource's `spec` must declare `observesGeneration: true`
also also be configured to identify success, failure and unknown states - which map to Healthy, Unhealthy and Unknown
if the mechanism to be relied upon is
[Allow Resources to Report Status](https://github.com/vmware-tanzu/cartographer/pull/738).

# Drawbacks

[drawbacks]: #drawbacks

Making assertions against resources which require matching against the spec introduces a few downsides:

* Holding spec updates is another bottleneck which can further slow processing;
* Resources must implement `observedGeneration` pattern;
* It must be possible to determine if the resources are still processing (Unknown state, as per
  [Allow Resources to Report Status](https://github.com/vmware-tanzu/cartographer/pull/738)).
* Great responsibility is placed on template authors to write correct `correlationRules` and health determination rules,
  as well as setting `observesGeneration: true` only when appropriate. If not written correctly, Cartographer may not
  propagate artifacts correctly.

# Prior Art

[prior-art]: #prior-art

[Draft RFC Read Resources Only When In Success State](https://github.com/vmware-tanzu/cartographer/pull/556)

[Allow Resources to Report Status](https://github.com/vmware-tanzu/cartographer/pull/738)

# Unresolved Questions

[unresolved-questions]: #unresolved-questions

Ultimately, artifact tracing will require being able to correlate inputs with outputs for all resources stamped out for
a blueprint owner.

If a resource does not behave in such a way that the above mechanics can be applied meaningfully, then the resource
cannot be used if artifact tracing is desired.

# Spec. Changes

[spec-changes]: #spec-changes

All templates will now have an optional `correlationRules` field, which must be an array containing at least one element
identifying a pair of jsonpath expressions, representing details which must be matched against an input and an output.

`correlationRules` are only be allowed if a `healthRule` is specified, as per
[Allow Resources to Report Status](https://github.com/vmware-tanzu/cartographer/pull/738) or a similar mechanism.

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterTemplate
metadata:
  name: example-template
spec:
  template:
    apiVersion: some-api/v1
    kind: NotImportant
    metadata:
      name: ...
    spec:
      ...
correlationRules:
  - expectedValue: "$(sources.source.url)$"        #evaluated against inputs in template context
    actualPath: ".status.outputs.url"              #evaluated against the in-cluster resource
  - expectedValue: "$(sources.source.revision)$"   #evaluated against inputs in template context
    actualPath: ".status.outputs.scannedRevision"  #evaluated against the in-cluster resource
```

`expectedValue` is a template evaluated in template context. Expressions may only reference `sources`, `images`,
`configs`, `deployments` from the resource's templating context.

`actualPath` is a Jsonpath expression evaluated against the in-cluster resource. If it accesses the `spec` of the
resource, then the template must also specify `observesGeneration: true` or the tempate is invalid.