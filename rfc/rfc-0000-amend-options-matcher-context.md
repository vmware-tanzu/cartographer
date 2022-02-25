# Meta
[meta]: #meta
- Name: Amend options matcher context to remove workload prefix
- Start Date: 2022-02-22
- Author(s): @jwntrs
- Status: Draft
- RFC Pull Request:
- Supersedes: N/A

# Summary
[summary]: #summary

Amend the options matcher context as defined in the [template switching RFC](https://github.com/vmware-tanzu/cartographer/pull/75) to remove the 'workload' prefix from the object matcher. This was first proposed in [this discussion](https://github.com/vmware-tanzu/cartographer/pull/602#discussion_r808047234).

# Motivation
[motivation]: #motivation

- Originally the workload prefix was included to leave room for other prefixes (such as params). However based on the direction from the [matchParams RFC](https://github.com/vmware-tanzu/cartographer/pull/618) this no longer makes sense.
- Be as consistent as possible with the direction of the [top level selector RFC](https://github.com/vmware-tanzu/cartographer/pull/602)


# What it is
[what-it-is]: #what-it-is

Change this:

```yaml
  resources:
    - name: source-provider
      templateRef:
        kind: ClusterSourceTemplate
        options: 
        - name: git-template
          selector:
            matchFields:
              - key: workload.spec.source.git       #< ========== remove this 'workload' prefix
                operator: Exists
```

To this:

```yaml
  resources:
    - name: source-provider
      templateRef:
        kind: ClusterSourceTemplate
        options: 
        - name: git-template
          selector:
            matchFields:
              - key: spec.source.git       #< ========== just reference the spec
                operator: Exists
```

# How it Works
[how-it-works]: #how-it-works


# Migration
[migration]: #migration

The previous change has not yet been released so no migration is needed.

# Drawbacks
[drawbacks]: #drawbacks


# Alternatives
[alternatives]: #alternatives

# Prior Art
[prior-art]: #prior-art


# Unresolved Questions
[unresolved-questions]: #unresolved-questions


# Spec. Changes (OPTIONAL)
[spec-changes]: #spec-changes

```yaml
  resources:
    - name: source-provider
      templateRef:
        kind: ClusterSourceTemplate
        options: 
        - name: git-template
          selector:
            matchFields:
              - key: workload.spec.source.git       #< ========== remove this 'workload' prefix
                operator: Exists
```
