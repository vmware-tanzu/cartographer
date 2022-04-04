# Meta

[meta]: #meta

- Name: Include submitted condition in owner.status.resources.conditions
- Start Date: 2022-4-1
- Author(s): [Rasheed Abdul-Aziz](https://github.com/squeedee) [Todd Ritchie](https://github.com/pivotal-todd-ritchie)
- Status: Draft
- RFC Pull Request: (leave blank)
- Supersedes: N/A

# Summary

[summary]: #summary

Cartographer shows a `reason:ResourcesSubmitted` in conditions. If there is an issue (temporary or error) with the
stamping of an issue, this condition lets the user know about the _first resource_ where an issue id encountered. We
should present issues per-resource, in the Owner's `status.resources[].conditions` as well.

# Motivation

[motivation]: #motivation

Top level conditions make it difficult to present a list of issues, and we now have a better UX architecture in Owners
to display per-resource concerns. Users want to see, on a resource-by-resource basis, what the state of the supplychain
or delivery is.

We expect automation and graphical interfaces to present the status of resources just by querying the workload, and for
text users (kubectl extreme!) to see issues at a glance.

# What it is

[what-it-is]: #what-it-is

We add or augment `owner.status.resources[].conditions` to contain a `Submitted` condition such as:

```yaml
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  name: my-workload
    namespace: default
spec: { ... }
status:
  supplyChainRef:
    kind: ClusterSupplyChain
    name: supply-chain
  conditions:
    - type: Ready
      status: Unknown
      reason: MissingValueAtPath
      message: waiting to read value [.status.latestImage] from resource [image.kpack.io/testing-sc]
        in namespace [default]
      lastTransitionTime: "2022-03-22T20:15:21Z"
    - type: SupplyChainReady
      status: "True"
      reason: Ready
      message: ""
      lastTransitionTime: "2022-03-22T20:15:05Z"
    - type: ResourcesSubmitted
      status: Unknown
      reason: MissingValueAtPath
      message: waiting to read value [.status.latestImage] from resource [image.kpack.io/testing-sc]
        in namespace [default].
      lastTransitionTime: "2022-03-22T20:15:21Z"
  resources:
    - name: my-resource
      conditions:
        - type: Ready
          status: Unknown
          reason: MissingValueAtPath
        - type: Submitted     # <<=== this is the new section 
          status: Unknown
          reason: MissingValueAtPath
          message: waiting to read value [.status.latestImage] from resource [image.kpack.io/testing-sc] in namespace [default].
    - name: another-resource
      conditions:
        - type: Ready
          status: Unknown
          reason: MissingValueAtPath
        - type: Submitted     # One for each resource in spec.resources 
          status: Unknown
          reason: MissingValueAtPath
          message: ...
```

# How it Works

[how-it-works]: #how-it-works

During the creation of all stamped objects, the realizer provides a conditions collection for each resource. This
includes any existing conditions (
see [Resource Status on Workload RFC](https://github.com/vmware-tanzu/cartographer/blob/rfc-resources-report-status/rfc/rfc-0000-allow-resources-to-report-status.md))
and a calculated `type:Ready` condition.

Each `type:Submitted` condition follows the same implementation as the top level `type:ResourcesSubmitted` condition.

The type:Ready condition should follow the standard rules of our
existing [ConditionManager](../pkg/conditions/condition_manager.go).

# Migration

[migration]: #migration

There should be no breaking changes to the API

# Drawbacks

[drawbacks]: #drawbacks

None that we are aware of.

# Alternatives

[alternatives]: #alternatives

This proposal provides a denormalization of information that is otherwise hidden by the top level
condition `ResourcesSubmitted`. In theory, we could provide all failed submissions in the message
for `ResourcesSubmitted` but it would be messy and close to impossible to combine different 'reasons'.

This proposal uses standard [Condition types](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1@v0.23.5#Condition)
and maintains consistency with reporting on the status of resource creation and status that has already been accepted.

# Prior Art

[prior-art]: #prior-art

Extends the work of
this [Resource Status on Workload](https://github.com/vmware-tanzu/cartographer/blob/rfc-resources-report-status/rfc/rfc-0000-allow-resources-to-report-status.md)
RFC

# Unresolved Questions

[unresolved-questions]: #unresolved-questions

N/A

# Spec. Changes (OPTIONAL)

[spec-changes]: #spec-changes

See [What it is](#what-it-is) above for spec changes.