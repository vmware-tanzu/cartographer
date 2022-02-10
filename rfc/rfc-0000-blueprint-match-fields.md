# Meta
[meta]: #meta
- Name: Blueprint selector matchFields
- Start Date: 2022-02-10
- Author(s): @jwntrs
- Status: Draft
- RFC Pull Request:
- Supersedes: N/A

# Summary
[summary]: #summary

The [template switching RFC](https://github.com/vmware-tanzu/cartographer/pull/75) introduced template `options` that included a `matchFields` selector. We should include this same selector at the `blueprint` level

# Motivation
[motivation]: #motivation

- Allow entire `blueprints` to match an `owner` based on certain fields.

- Ensure that the behaviour of all `selectors` is consistent.

# What it is
[what-it-is]: #what-it-is

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSupplyChain
metadata:
  name: supply-chain
spec:
  selector:
    matchFields:                                                # <=========== add this
      { key: "workload.labels.pipeline", operation: exists }    # <=========== 
```

# How it Works
[how-it-works]: #how-it-works

The same way as `matchFields` from [template switching RFC](https://github.com/vmware-tanzu/cartographer/pull/75)

# Migration
[migration]: #migration

This is most likely a breaking change which would require a bump to the `ClusterSupplyChain` `apiVersion` and introducing a conversion webhook.

```yaml
  selector:
    app.tanzu.vmware.com/workload-type: web
```

needs to become:

```yaml
  selector:
    matchFields:
      - key: workload.metadata.labels."app.tanzu.vmware.com/workload-type"
        operator: In
        values: ["web"]
```


# Drawbacks
[drawbacks]: #drawbacks

This is a breaking change.

# Alternatives
[alternatives]: #alternatives

If we introduced `matchLabels` at the same time as `matchFields`, when migrating we would simply need to hoist everything under the current `selector` over to the new `matchLabels` field.

```yaml
  selector:
    app.tanzu.vmware.com/workload-type: web
```

becomes

```yaml
  selector:
    matchLabels:
      app.tanzu.vmware.com/workload-type: web
```


# Prior Art
[prior-art]: #prior-art

[template switching RFC](https://github.com/vmware-tanzu/cartographer/pull/75) matchFields

# Unresolved Questions
[unresolved-questions]: #unresolved-questions

# Spec. Changes (OPTIONAL)
[spec-changes]: #spec-changes

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSupplyChain
metadata:
  name: supply-chain
spec:
  selector:
    matchFields:                                                # <=========== add this
      { key: "workload.labels.pipeline", operation: exists }    # <=========== 
```
