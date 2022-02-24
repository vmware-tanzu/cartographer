[Draft PR](https://github.com/vmware-tanzu/cartographer/pull/319)
# RFC 0015 Supply Chain selection through traits

## Summary

A mechanism to allow a cluster to have multiple Supply Chains declared
1. With similar purposes (or where one extends the behaviour of another)
2. That will match the developer's requirements without the developer needing to understand naming permutations.


```gherkin
Given Supply Chains that represent permutations of a behaviour tree
When A developer annotates their workload to select for a permutation
Then the correct supply chain is selected.
```

## Motivation

Examples of app characteristics that might make up a permutation:

* `type`:
    * `web`: an app that will deploy as a web service endpoint
    * `agent`: an app that monitors other services

* `input`:
    * `image`: an app that is provided as a OCI image
    * `source`: an app with source

* `test`:
    * `tekton`: the devloper has a tekton pipeline that can be used to test the code

* `scan`:
    * `security`: the developer wants an image scanner to run across their code.

The permutations, if using a single selector would be vast:

* `web-source`
* `web-source-test`
* `web-source-test-scan`
* `web-image`
* `web-image-test`
* `web-image-test-scan`
* etc...

Further, the developer would need to know that `web-image-test-scan` is valid, not `web-image-test-scan`  

## Possible Solutions


[Draft Text](https://github.com/vmware-tanzu/cartographer/blob/rfc-0015-supply-chain-selection-through-traits-impl/rfc/rfc-0015-supply-chain-selection-through-traits.md)
### 1. Characteristics

With operators defining a simple set of available `characteristics', developers can select for appropriate supply chains
in a more readable manner. Further, the operators can change the available set of supply chains without breaking existing
workloads.

Given these Supply chains:
```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSupplyChain
metadata:
  name: image-web
spec:
  selector:
    workload/type: web
    workload/input: image
```

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSupplyChain
metadata:
  name: image-web-tekton
spec:
  selector:
    workload/type: web
    workload/input: image
    workload/test: tekton
```

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSupplyChain
metadata:
  name: image-web-tekton-security
spec:
  selector:
    workload/type: web
    workload/input: image
    workload/test: tekton
    workload/scan: security
```

Workloads:

1. would select `image-web-tekton`
    ```yaml
    apiVersion: carto.run/v1alpha1
    kind: Workload
    metadata:
      labels:
        workload/type: web
        workload/input: image
        workload/test: tekton
    ```

1. would select `image-web`
    ```yaml
    apiVersion: carto.run/v1alpha1
    kind: Workload
    metadata:
      labels:
        workload/type: web
        workload/input: image
        workload/test: unsupported
    ```

1. would fail to select
    ```yaml
    apiVersion: carto.run/v1alpha1
    kind: Workload
    metadata:
      labels:
        workload/type: web
        workload/test: tekton
    ```

where the rule is:
A workload will only be selected by (in priority order) 
   1. a perfect match
   2. an over-match which does not match another supply chain more perfectly

### 2. Selecting on fields

In the "Characteristics" solution, it's assumed that a workload with `workload/input: source` would have the following 
snippet:

```yaml
spec:
  source:
    git:
      url: https://github.com/spring-projects/spring-petclinic.git
      ref:
        branch: main

```

and `workload/input: image` would have:
```yaml
spec:
  image: example.com/my/image
```

Selectors that allow us to match a Supply Chain against field presence (eg: `spec.source.git`) might be less redundant. 
This could be a combined solution, that is, Characteristics and Field-Selectors are not exclusive.

## Cross References and Prior Art

n/a