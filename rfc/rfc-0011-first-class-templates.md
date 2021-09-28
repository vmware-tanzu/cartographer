This RFC aims to tackle at once:

1. The use of intermediate state crd's for templates
1. Treating templates as first class, such that they
    1. Work standalone
    1. Are easily treated as entries in a catalog of Templates
    1. Can be easily authored by tool creators, and shared with app-operators
    1. are fully isolated from references to a `workload`
1. Make the RunTemplate and Cluster*Templates closer cousins, as they already appear to be on paper.

# Current Vs Planned
Current:
![Current](./rfc-0011/now.jpg "Current design")

Planned:
![Current](./rfc-0011/planned.jpg "Planned changes in design")

# Workload (unchanged)
```yaml
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  name: petclinic
  labels:
    supply-chain: "a-good-one"
spec:
  source:
    git:
      url: https://github.com/spring-projects/spring-petclinic.git
```

# Supply Chain
```yaml
---
apiVersion: carto.run/v1alpha1
kind: ClusterSupplyChain
metadata:
  name: responsible-ops---consume-output-of-components
spec:
  selector:
    supply-chain: "a-good-one"
  components:
    - name: source-provider
      template: # or ytt
        apiVersion: carto.run/v1alpha1
        kind: Source
        metadata:
          name: $(workload.name)$
          labels:
            source-provider: "mr-git-fetch"
        spec:
          source: $(workload.spec.source)$
    - name: test-provider
      sources:
        - component: source-provider
          name: source
      template: # or ytt
        apiVersion: carto.run/v1alpha1
        kind: Pipeline
        metadata:
          name: $(workload.name)$
        spec:
          runTemplateRef:
            name: my-run-template-inputs
          inputs:
            source-url: $(source.url)$
            source-revision: $(source.revision)$
```

# Source-Provider template
```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSourceTemplate
metadata:
  name: source-template
spec:
  selector:
    source-provider: "mr-git-fetch"
  urlPath: .status.artifact.url
  revisionPath: .status.artifact.revision
  ytt: |
    #@ load("@ytt:data", "data")

    #@ if hasattr(data.values.spec.source, "git"):
    apiVersion: source.toolkit.fluxcd.io/v1beta1
    kind: GitRepository
    metadata:
      name: #@ data.values.metadata.name
      labels:
        app.kubernetes.io/component: source
        #@ if/end hasattr(data.values.metadata, "labels") and hasattr(data.values.metadata.labels, "app.kubernetes.io/part-of"):
        app.kubernetes.io/part-of: #@ data.values.metadata.labels["app.kubernetes.io/part-of"]
    spec:
      interval: 1m
      url: #@ data.values.spec.source.git.url
      ref: #@ data.values.spec.source.git.ref
      gitImplementation: libgit2
      ignore: ""
    #@ end

    #@ if hasattr(data.values.spec.source, "image"):
    apiVersion: source.apps.tanzu.vmware.com/v1alpha1
    kind: ImageRepository
    metadata:
      name: #@ data.values.metadata.name
      labels:
        app.kubernetes.io/component: source
        #@ if/end hasattr(data.values.metadata, "labels") and hasattr(data.values.metadata.labels, "app.kubernetes.io/part-of"):
        app.kubernetes.io/part-of: #@ data.values.metadata.labels["app.kubernetes.io/part-of"]
    spec:
      interval: 1m
      image: #@ data.values.spec.source.image
    #@ end

```