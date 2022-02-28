# Meta
[meta]: #meta
- Name: Add matchFields and matchLabels to top level of blueprints
- Start Date: 2022-02-10
- Author(s): @jwntrs
- Status: Draft
- RFC Pull Request:
- Supersedes: N/A

# Summary
[summary]: #summary

The [template switching RFC](https://github.com/vmware-tanzu/cartographer/pull/75) introduced template `options` that included a `matchFields` selector. We should include this same selector at the `blueprint` level, and we should move the current labels to a `matchLabels` selector.

# Motivation
[motivation]: #motivation

- Allow entire `blueprints` to match an `owner` based on certain fields.

- Ensure that the behaviour of all `selectors` is consistent.

# What it is
[what-it-is]: #what-it-is

```yaml
apiVersion: carto.run/v1alpha2                                  # <======== v2 spec
kind: ClusterSupplyChain
metadata:
  name: supply-chain
spec:
  selector:
    matchLabels:                                                # <=========== move existing labels under this heading
      app.tanzu.vmware.com/workload-type: web
    matchFields:                                                                             # <=========== add this
      - { key: "spec.image", operation: exists }                                             # <=========== 
    matchExpressions:                                                                        # <=========== add this
    - { key: app.tanzu.vmware.com/workload-type, operator: In, values [web] }                # <===========
```


```yaml
apiVersion: carto.run/v1alpha1                                  # <======== v1 spec
kind: ClusterSupplyChain
metadata:
  name: supply-chain
spec:
  selector:
    app.tanzu.vmware.com/workload-type: web
  selectorMatchFields:                                                                       # <=========== add this
    - { key: "spec.image", operation: exists }                                               # <=========== 
  selectorMatchExpressions:                                                                  # <=========== add this
    - { key: app.tanzu.vmware.com/workload-type, operator: In, values [web] }                # <===========
```


# How it Works
[how-it-works]: #how-it-works

The same way as `matchFields` from [template switching RFC](https://github.com/vmware-tanzu/cartographer/pull/75)

This will also impact how best match works at the supply chain level. Right now a best-match is indicated by the number of labels. We should broaden this definition to include the number of labels and number of fields matched.

# Migration
[migration]: #migration

This is most likely a breaking change which would require a bump to the `ClusterSupplyChain` `apiVersion` and introducing a conversion webhook.

```yaml
  selector:
    app.tanzu.vmware.com/workload-type: web
```

needs to be transformed into this:

```yaml
  selector:
    matchLabels:
      app.tanzu.vmware.com/workload-type: web
```


# Drawbacks
[drawbacks]: #drawbacks

This is a breaking change.

# Alternatives
[alternatives]: #alternatives


# Prior Art
[prior-art]: #prior-art

[template switching RFC](https://github.com/vmware-tanzu/cartographer/pull/75) matchFields

# Unresolved Questions
[unresolved-questions]: #unresolved-questions

# Spec. Changes (OPTIONAL)
[spec-changes]: #spec-changes

```yaml
apiVersion: carto.run/v1alpha1                                 # <======== v1 spec
kind: ClusterSupplyChain
metadata:
  name: supply-chain
spec:
  selector:
    app.tanzu.vmware.com/workload-type: web
  selectorMatchFields:                                                                       # <=========== add this
    - { key: "spec.image", operation: exists }                                               # <=========== 
  selectorMatchExpressions:                                                                  # <=========== add this
    - { key: app.tanzu.vmware.com/workload-type, operator: In, values [web] }                # <===========
```
