# Meta
[meta]: #meta
- Name: Add matchParams selector
- Start Date: 2022-02-14
- Author(s): @jwntrs
- Status: Draft
- RFC Pull Request:
- Supersedes: N/A

# Summary
[summary]: #summary

The [template switching RFC](https://github.com/vmware-tanzu/cartographer/pull/75) introduced the ability to match templates based on fields in the workload. However matching on params is not well defined. It is currently not possible to match on a field in the supply chain, and matching on fields in the workload is complex. You would need to define something as follows:

```yaml
  selector:
    matchFields:
      - key: workload.spec.params[?(@.name=="promotion")].value    #< ===== not so nice
        operator: In
        values: ["gitops"]
```

Lets introduce a matchParams selector:

```yaml
  selector:
    matchParams:           #< ===== use something like this instead
      - key: promotion    
        operator: In
        values: ["gitops"]
```


# Motivation
[motivation]: #motivation

We should be able to match against default supply chain params in our field selectors, and provide a simpler syntax for matching against workload params.

# What it is
[what-it-is]: #what-it-is

```yaml
apiVersion: kontinue.io/v1alpha1
kind: ClusterSupplyChain
metadata:
  name: responsible-ops
spec:
  selector:
    matchLabels:
      app: web
    matchParams:      #< ===== introduce convenience matcher for workload params
      - key: language
        operator: In
        values: ["java"]    
        
  params:
  - name: promotion       #< ===== default params can be set here
    default: (gitops|regops)
  resources:
   ...
    - name: promote
      templateRef:
        kind: ClusterTemplate
        options:
        - name: git-promotion
          selector:
            matchParams:      #< ===== introduce `matchParams` selector (uses param hierarchy)
              - key: promotion           
                operator: In
                values: ["gitops"]
        - name: registry-promotion
          selector:
            matchParams:
              - key: promotion
                operator: In
                values: ["regops"]

---
apiVersion: kontinue.io/v1alpha1
kind: Workload
metadata:
  name: petclinic
  labels:
    app: web
spec:
  params:
  - name: promotion           #< ===== params can be overriden here
    value: (gitops|regops)
  source:
    git:
      url: https://github.com/spring-projects/spring-petclinic.git
      ref:
        branch: main
```

# How it Works
[how-it-works]: #how-it-works

Compute the available params from the workload and supply chain based on the [param hierarchy](https://cartographer.sh/docs/v0.2.0/architecture/#parameter-hierarchy) and pass it to the params matcher.

# Migration
[migration]: #migration

For the top level `matchParams` selector, this assumes we've already migrated the top level selector to use `matchLabels`.

# Drawbacks
[drawbacks]: #drawbacks

Ultimately we want our template `selector` and top-level supply chain `selector` to work the same. However, if we introduce a matchParams selector, then it would cause some divergence in these two cases. 

The top level selector only has access to params defined in the workload.

The template selectors would have access to params from the workload as well as default values from the supply chain.

Neither selector would have access to params from the templates.


# Alternatives
[alternatives]: #alternatives

???

# Prior Art
[prior-art]: #prior-art

???

# Unresolved Questions
[unresolved-questions]: #unresolved-questions

???

# Spec. Changes (OPTIONAL)
[spec-changes]: #spec-changes

```yaml
  selector:
    matchParams:         #< =================== introduce `matchParams` selector for templates and top level selector
      - key: promotion           
        operator: In
        values: ["gitops"]
```


