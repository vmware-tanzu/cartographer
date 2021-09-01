# Matching user-provided objects

In the [previous example], we went through the process of creating a
`carto.run/RunTemplate` that describes how invocations of `tekton/Pipeline`s
(i.e., `tekton/PipelineRun` objects) should look like.

Here we make use of an additional feature of `carto.run/Pipeline`: supplying
information about external objects to the `carto.run/RunTemplate` templates.


## The example

Imagine that we have two groups:

- operators, not familiar with the details of the applications being developed,
  but very well acquainted with Tekton and Kubernetes in general

- contracted app developers, experts in Go and the applications their
  developing, but not familiar with neither Kubernetes nor Tekton.


In this case, it'd be great if:

1. operators could provide the definitions of how Tekton pipelines should be run
  according to their companies best practices

2. contracted developers could "just plug in" scripts to run their tests

Still being Kubernetes-native, **developers** would then supply their test
configuration via ConfigMaps, labelled with their "team id":

```yaml
kind: ConfigMap
apiVersion: v1
metadata:
  name: team-1-testing-scripts
  labels:
    pipelines.carto.run/id: "one"
data:
  01-unit.sh: |-
    go test -v ./...
```

**operators**, very familiar with Kubernetes, would then take care of defining
the common tekton/Pipeline, the definition of how to invoke the tekton/Pipeline
(i.e., the `carto.run/RunTemplate`), and the definition of the
`carto.run/Pipeline`s which point at that single `carto.run/RunTemplate`.

The detail here is that those operators would need to include in the
`carto.run/RunTemplate` information about the `ConfigMap`s supplied by the
developers.

To enable that, `pipeline-service` provides an extra field under
carto.run/Pipeline: `spec.selector`.

```yaml
apiVersion: carto.run/v1alpha1
kind: Pipeline
metadata:
  name: pipeline-1
spec:
  runTemplateName: tekton

  inputs:
    url: https://github.com/kontinue/hello-world
    revision: 19769456b6b229b3e78f2b90eced15a353eb4e7c

  selector:
    resource:
      apiVersion: v1
      kind: ConfigMap
    matchingLabels:
      pipelines.carto.run/id: "one"
```

making available to the `carto.run/RunTemplate` as extra interpolation field:
`$(selected.<>)$`.

```yaml
apiVersion: carto.run/v1alpha1
kind: RunTemplate
metadata:
  name: tekton
spec:
  template:
    apiVersion: tekton.dev/v1beta1
    kind: PipelineRun
    metadata:
      generateName: $(pipeline.metadata.name)$-
    spec:
      # ...
      workspaces:
        - name: commands
          configmap:
            name: $(selected.metadata.name)$
```

In summary, we have the following:


```
.── devs
│   │                          (no k8s knowledge required)
│   ├── commands-dev-1.yaml           // configmap
│   └── commands-dev-2.yaml           // configmap
│
└── operators
    ├── pipeline-service-pipeline-1.yaml      // carto.run/Pipeline matching on dev-1
    ├── pipeline-service-pipeline-2.yaml      // carto.run/Pipeline matching on dev-2
    ├── pipeline-service-run-template.yaml    // single carto.run/RunTemplate
    └── tekton-pipeline.yaml                  // single tekton/Pipeline and set of common tasks
```


[previous example]: ../02-simple-pipeline-service
