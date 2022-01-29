# Meta
[meta]: #meta
- Name: Multiple Paths through a blueprint
- Start Date: 1-21-22
- Author(s): @squeedee
- Status: Draft 
- RFC Pull Request: 
- Supersedes: Partially: [RFC-0005](https://github.com/vmware-tanzu/cartographer/pull/72)

# Summary
[summary]: #summary

Extend [RFC 0009](https://github.com/vmware-tanzu/cartographer/blob/rfc-0009-supply-chain-switches-template-on-flag/rfc/rfc-0009-supply-chain-switches-template-on-flag.md) to support input matching in options. This provides a mechanism for an option to be chosen as part of 
a path through a complex supply chain.

# Motivation
[motivation]: #motivation

The use cases supported by this RFC are any scenario where changing content in the workload/delivery means that more or less
resources in the supply chain should be stamped out.


A concrete example is when a workload indicates it `has-tests`, and we want to stamp a tekton pipeline runner for those tests, or
not stamp the resoruce if there are no tests.

```text
workload
  -- [label.has_tests == true] --> image-tester 
        --> configure    # we want to arrive here regardless of has-tests
```


Another example is when a workload provides `spec.image`, we want to skip `source-provider` and `source-tester`, using 
`image-provider` instead. 

```text
workload
  -- [spec.image.defined?] --> image-provider 
    --> configure
  -- [spec.image.notDefined?] --> source-provider 
    --> source-tester
       --> configure
```



# What it is
[what-it-is]: #what-it-is

This RFC introduces `input selection`: the ability to add `sources:`, `images:` and `configs:` to the `spec.resources[].options[]` field.
It describes `selection sets` which is the sum total of `selectors` and `input selection` leading to an option or resource.

This spec produces [this graph](https://mermaid.live/view/#pako:eNqNVU1PwzAM_StRuK4S7DChTnCCG-IAxwVNoXHbaGtSJSljQvx33KZrtn6snGYnz_az9-L-0EQLoDFN9_qQ5Nw48vbCFCHJnlv7BCmxujIJvCKKWGf0DuKb1SpdeDs6SOHyeFl-r-sgW31mhpc5MWC3PjByYB2YzYX3UYMJEdJA4qRWbU1CdOm2t9saFLUBB-lyjNo5rTaMTtwwpm7ul8vVmuTcNjXsgzMVMNpUAiUG7GTBM4hKo7-kQHqX7gw_D06NLiJvGsgkDuSIDCfvkGNLpxntXKqzua-v9_BZyX1oofVmOmhQviJyPvMCSw-9-w90jFuiVSqzysCms2Y4hQhGO3vIZwY2xiWTTpd2439mWCAoKiub40AZDU5QmOMmA_fgk_W4nf68kKF3MkhjoNQWaaA-pvm3gu_E2vNnOmrRjcbqhuqSyGzseFyjY8i-PE8TOMcO3sb05Vjha8kmXsfV9UGi6HFsnCHwTNx9sN9bJ-jdNegwbyfTDnrxVkPWKeAwZVBmBw2PQU3ocQilC1qAKbgU-An4qQNxx-ZQ4OaM0RTc7Bhl6hdxVSm4g2eBujc0rrfrgvLK6fejSmic8r2FE-hJchRv0Z7-_gFDOk2E):
```yaml
---
apiVersion: carto.run/v1alpha1
kind: ClusterSupplyChain
metadata:
  name: supply-chain
spec:
  resources:
    - name: source-provider
      templateRef:
        kind: ClusterSourceTemplate
        options:
          - name: source-from-git-repo
            selector:
              matchFields:
                { key: "spec.source.url", operation: exists }
          - name: source-from-image-registry
            selector:
              matchFields:
                { key: "spec.source.image", operation: exists }

    - name: source-tester
      templateRef:
        kind: ClusterSourceTemplate
        name: test-source-with-tekton
      selector:
        matchLabels:
          has-tests: "true"
      sources:
        - resource: source-provider
          name: source

    - name: image-provider
      templateRef:
        kind: ClusterImageTemplate
        name: image-from-image-registry
      selector:
        matchFields:
          { key: "spec.image", operation: exists }

    - name: image-builder
      templateRef:
        kind: ClusterImageTemplate
        options:
          - name: build-image
            images:
              - resource: source-tester
                name: source
          - name: build-image
            images:
              - resource: source-provider
                name: source

    - name: configure
      templateRef:
        kind: ClusterConfigTemplate
        options:
          - name: configure
            sources:
              - resource: image-builder
                name: image
          - name: configure
            sources:
              - resource: image-provider
                name: image

    - name: gitops
      templateRef:
        kind: ClusterSourceTemplate
        options:
          - name: git-pusher
            selector:
              matchLabels:
                target: gitops
            configs:
              - resource: configure
                name: config
          - name: registry-pusher
            selector:
              matchLabels:
                target: repostiory
            configs:
              - resource: configure
                name: config
```

If label `has-tests` is "true", then `image-builder` chooses the first option, from source-tester, because source-tester _will_ eventually be stamped out. 
Otherwise `image-builder` selects the second option, from source-provider.

This lets a resource be skipped if none of its options match.

`configure` shows a similar pathing, based this time on matchFields not matchLabels.

# How it Works
[how-it-works]: #how-it-works

An option is selected for one of these reasons:

> It's the only option and all other criteria are met.

OR

> It matches on labels and fields (if no selector is specified, that is a match)

AND

> It's inputs are fulfilled using these same rules.

<TBD>: Empirical tests for priority lists.

Options are a priority list to disambiguate inputs that match in more than one case. It should still be possible
to warn supply chain authors of unreachable components. Each of source-tester's options are reachable because each input
has a different 'selection set' 
``` yaml
- name: source-tester
  templateRef:
    kind: ClusterImageTemplate
    options:
      - name: build-image
        images:
          - resource: source-tester
            name: source
      - name: build-image
        images:
          - resource: source-provider
            name: source
```

# Migration
[migration]: #migration

This does not break the existing API.

# Drawbacks
[drawbacks]: #drawbacks

<TBD>

Why should we *not* do this?

# Alternatives
[alternatives]: #alternatives

<TBD>
- What other designs have been considered?
- Why is this proposal the best?
- What is the impact of not doing this?

# Prior Art
[prior-art]: #prior-art
<TBD>
Discuss prior art, both the good and bad.

# Unresolved Questions
[unresolved-questions]: #unresolved-questions

- Empiricial tests for complex paths. I have a repository for working these through here: https://github.com/squeedee/vizit 
- Can we shake the tree fast enough that this works well as a pre-realize step?

# Spec. Changes
[spec-changes]: #spec-changes

Adds:
```
spec.resources[].templateRef.options[].images
spec.resources[].templateRef.options[].configs
spec.resources[].templateRef.options[].sources
```

All three follow the top level spec for
```
spec.resources[].images
spec.resources[].configs
spec.resources[].sources
```
