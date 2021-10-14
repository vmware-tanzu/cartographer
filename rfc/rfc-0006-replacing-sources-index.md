
# Draft RFC 0006 Replace sources[0]

[Draft PR](https://github.com/vmware-tanzu/cartographer/pull/73)


## Summary

Users have provided feedback that they do not like using the notation:
`sources[0].url` The following is a discussion of 4 changes that could be made
together or separately.

## Motivation

Users report disliking the notation `sources[0].url`

## Questions
- How should supply chains declare inputs?
    - List? Map?
- How should templates consume inputs...
    - When there is just one input
    - When the input type has just one field
    - When there are more than one of each type


## Current Solution

SupplyChain defines a list for each type of input:
```yaml
---
kind: ClusterSupplyChain
spec:
  components:
    ...
    - name: opinion-provider
      templateRef:
        kind: ClusterOpinionTemplate
        name: example-opinion
      images:
        - component: build-provider
          name: solo-build-provider
```

Templates declare a jsonPath to consume the field(s) of that input. There are two main ways to write such a jsonpath:
```yaml
apiVersion: kontinue.io/v1alpha1
kind: ClusterOpinionTemplate
metadata:
  name: example-opinion
spec:
  template:
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: example-build-configmap
    data:
      some_value: $(images[0].image)$                                   # <--- index notation
      same_value: $(images[?(@.name=="solo-source-provider")].image)$   # <--- expression notation
  ...
```

## Possible Solutions

### SupplyChain defined as map
SupplyChains could define the inputs as a map:
```yaml
---
kind: ClusterSupplyChain
spec:
  components:
    ...
    - name: opinion-provider
      templateRef:
        kind: ClusterOpinionTemplate
        name: example-opinion
      images:
        solo-build-provider: build-provider
```

### *Templates could simplify consuming a single instance of an input type
For all of the input types (images, sources, opinions), *Templates could request the singular noun. In that case, they would match if the SupplyChain only provided 1 of that input type.
```yaml
apiVersion: kontinue.io/v1alpha1
kind: ClusterOpinionTemplate
metadata:
  name: example-opinion
spec:
  template:
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: example-build-configmap
    data:
      some_value: $(image.image)$
  ...
```

This SupplyChain would be in a bad state, because it provides the template two images, while the template requests only one:
```yaml
---
kind: ClusterSupplyChain
spec:
  components:
    ...
    - name: opinion-provider
      templateRef:
        kind: ClusterOpinionTemplate
        name: example-opinion
      images:
        - component: build-provider
          name: first-build-provider
        - component: sidecar-provider
          name: second-build-provider
```

### *Template consuming an input with only one field
If the input type exposes only one field, then the template need not specify which field to read from the input. The applicable types are image and opinion. The excluded type is source.
```yaml
apiVersion: kontinue.io/v1alpha1
kind: ClusterOpinionTemplate
metadata:
  name: example-opinion
spec:
  template:
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: example-build-configmap
    data:
      some_value: $(image)$ # <== If combined with the '*Template consuming one input' proposal above
      
  ...
```

### *Templates consume as from maps
When multiple of an input type are needed, *Templates could consume them with map (rather than list) designation

Given this SupplyChain
```yaml
---
kind: ClusterSupplyChain
spec:
  components:
    ...
    - name: opinion-provider
      templateRef:
        kind: ClusterOpinionTemplate
        name: example-opinion
      images:
        - component: build-provider
          name: first-build-provider
        - component: sidecar-provider
          name: second-build-provider
```

A *Template could consume it with:
```yaml
apiVersion: kontinue.io/v1alpha1
kind: ClusterOpinionTemplate
metadata:
  name: example-opinion
spec:
  template:
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: example-build-configmap
    data:
      some_value: $(image["first-build-provider"].image)$
      another_value: $(image["second-build-provider"].image)$
  ...
```

## Proposed Values
### Objects match paths
I propose that it is easier for Kontinue authors to write objects if the notation on the objects matches the in memory structure they will consume. So if the *Template consumes an input using map notation, we should have SupplyChains declare inputs using a map as well.

## Cross References and Prior Art
