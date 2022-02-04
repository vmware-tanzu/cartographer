# Meta
[meta]: #meta
- Name: More Option Selectors
- Start Date: 2022-02-04
- Author(s): Waciuma Wanjohi, Ciro Costa
- Status: Draft <!-- Acceptable values: Draft, Approved, On Hold, Superseded -->
- RFC Pull Request: (leave blank)
- Supersedes: N/A

# Summary
[summary]: #summary

RFC 09 introduces the use of the selector MatchFields in supply-chain resource options. These
borrow their shape from
[the selectors defined on Pods](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#resources-that-support-set-based-requirements).
MatchFields can support matching on all (non-status) fields in the workload object. Carto
options selector should also support the label specific `matchExpressions` and `matchLabels`.

# Motivation
[motivation]: #motivation

Users may find the MatchLabels syntax simpler than the MatchFields syntax.

# What it is
[what-it-is]: #what-it-is

When definining an option for a blueprint resource, users may use either
a MatchLabel, MatchExpression, MatchField. Both MatchLabel and MatchExpression
examine only the labels on a workload. These fields are mutually exclusive for
any given option. The selectors used on one option do not determine the selectors
that may be used on another option.

```yaml
  resources:
    - name: source-provider
      templateRef:
        kind: ClusterSourceTemplate
        options: 
        - name: git-template
          selector:
            matchFields:
              - key: workload.spec.source.git       # would match if the workload has this field filled in
                operator: Exists
        - name: imgpkg-bundle-template
          selector:
            matchLabels:   # or the label set below
              app.tanzu.vmware.com/source-type: imgpkg-bundle
```

# How it Works
[how-it-works]: #how-it-works

TBD

# Migration
[migration]: #migration

This is an additive change and does not create any migration implications for users.

# Drawbacks
[drawbacks]: #drawbacks

Labels are a field that can be accessed by MatchFields. E.g. these two options are
equivalent:

```yaml
  resources:
    - name: source-provider
      templateRef:
        kind: ClusterSourceTemplate
        options: 
        - name: git-template
          selector:
            matchFields:
              - key: metadata.labels.source-type
                operator: In
                values: ["imgpkg-bundle"]
        - name: git-template
          selector:
            matchLabels:
              source-type: imgpkg-bundle
```

We may ask users to rely on MatchFields until we have evidence of dissatisfaction, directing
engineering efforts to other topics.

# Alternatives
[alternatives]: #alternatives

See [drawbacks].

# Prior Art
[prior-art]: #prior-art

https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#resources-that-support-set-based-requirements

# Unresolved Questions
[unresolved-questions]: #unresolved-questions


# Spec. Changes (OPTIONAL)
[spec-changes]: #spec-changes

Adds the option for `MatchLabels` and `MatchExpressions` under resource.options in the supply-chain.
