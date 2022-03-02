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

**Note:** we do not recommend letting developers opt out of security scans. We're using it
here purely for demonstration purposes.

The permutations, if using a single selector would be vast:

* `web-source`
* `web-source-test`
* `web-source-test-scan`
* `web-image`
* `web-image-test`
* `web-image-test-scan`
* etc...

Further, the developer would need to know that `web-image-test-scan` is valid, not `web-image-test-scan`  

## Solution
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

1. A workload is selected by the supply chain that matches ALL label selectors.
2. Where a workload matches more than one supply chain, the supply chain with the larger list
   of matched labels wins
3. Where there is multiple supply chains with the largest number of matches, an error is reported

## Alternatives/Extensions
### Selecting on fields

This option is now proposed in [this RFC](https://github.com/vmware-tanzu/cartographer/pull/591) (to complement
selecting on labels as proposed above).

## Cross References and Prior Art

n/a