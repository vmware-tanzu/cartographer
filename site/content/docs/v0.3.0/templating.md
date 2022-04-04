# Templating

There are two options for templating in Cartographer, simple templates and ytt:

## Simple Templates

Define simple templates in `spec.template` in [Templates](./reference/template)

Simple templates provide string interpolation in a `$(...)$` tag with
[jsonpath](https://pkg.go.dev/k8s.io/client-go/util/jsonpath) syntax.

## ytt

Define [ytt](https://carvel.dev/ytt/) templates in `spec.ytt` in [Templates](./reference/template)

[ytt](https://carvel.dev/ytt/) is complete YAML aware templating language.

Use `ytt` when your templates contain complex logic, such as **conditionals** or **looping over collections**.

## Template Data

Both options for templating are provided a data structure that contains:

- owner resource (workload, deliverable)
- inputs, that are specified in the blueprint for the template (sources, images, configs, deployments)
- parameters

Note: all `ytt` examples assume you have loaded the data module with `#@ load("@ytt:data", "data")`.

### Owner

The entire owner resource is available for retrieving values. To use an owner value, use the format:

- **Simple template**: `$(<workload|deliverable>.<field-name>.(...))$`
- **ytt**: `#@ data.values.<workload|deliverable>.<field-name>.(...)`

#### Owner Examples

| Simple template                   | ytt                                          |
| --------------------------------- | -------------------------------------------- |
| `$(workload.metadata.name)$`      | `#@ data.values.workload.metadata.name`      |
| `$(deliverable.spec.source.url)$` | `#@ data.values.deliverable.spec.source.url` |

### Inputs

The template specifies the inputs required in the blueprint.

You may specify a combination of one or more of these input types:

| Input type  | Accessor                                                    |
| ----------- | ----------------------------------------------------------- |
| sources     | `sources.<input-name>.url`, `sources.<input-name>.revision` |
| images      | `images.<input-name>`                                       |
| configs     | `configs.<input-name>`                                      |
| deployments | `sources.<input-name>.url`, `sources.<input-name>.revision` |

Where the `<input-name>` corresponds to the `spec.resources[].<sources|images|configs|deployments>[].name` expected in
the blueprint.

Specifying inputs in a template:

- **Simple template**: `$(<sources|images|configs|deployments>.<input-name>(.<value>))$`
- **ytt**: `#@ data.values.<sources|images|configs|deployments>.<input-name>(.<value>)`

#### Inputs Examples

Given a supply chain where a resource has multiple sources and a config:

```yaml
---
spec:
  resources:
    - name: my-template
      sources:
        - resource: source-tester
          name: tested
        - resource: source-original
          name: original
      configs:
        - resource: configurator
          name: app-configuration
```

They could be used in the template as follows:

| Simple template                 | ytt                                        |
| ------------------------------- | ------------------------------------------ |
| `$(sources.original.url)$`      | `#@ data.values.sources.original.url`      |
| `$(sources.tested.revision)$`   | `#@ data.values.sources.tested.revision`   |
| `$(configs.app-configuration)$` | `#@ data.values.configs.app-configuration` |

#### Input Aliases

If only one input of a given input-type is required, refer to it in the singular and omit the input-name.

- **Simple template**: `$<sources|images|configs|deployments>(.<value>)$`
- **ytt**: `#@ data.values.<sources|images|configs|deployments>(.<value>)`

#### Input Alias Examples

Given a supply chain where a resource has a single source and a single config:

```yaml
---
spec:
  resources:
    - name: my-template
      sources:
        - resource: source-original
          name: original
      configs:
        - resource: configurator
          name: app-configuration
```

They could be used in the template as follows:

| Simple template       | ytt                              |
| --------------------- | -------------------------------- |
| `$(source.url)$`      | `#@ data.values.source.url`      |
| `$(source.revision)$` | `#@ data.values.source.revision` |
| `$(config)$`          | `#@ data.values.config`          |

### Parameters

See [Parameter Hierarchy](architecture#parameter-hierarchy) for more information on the precedence of parameters for
owner, blueprint and templates.

To use a parameter in the template, use the format:

- **Simple template**: `$(params.<param-name>)$`
- **ytt**: `data.values.params.<param-name>`

#### Parameter Example

| Simple template           | ytt                                  |
| ------------------------- | ------------------------------------ |
| `$(params.image_prefix)$` | `#@ data.values.params.image_prefix` |

## Complete Examples

### Simple Template

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterImageTemplate
metadata:
  name: kpack-template
spec:
  params:
    - name: image_prefix
      default: projectcartographer/demo-

  imagePath: .status.latestImage

  template:
    apiVersion: kpack.io/v1alpha2
    kind: Image
    metadata:
      name: $(workload.metadata.name)$
    spec:
      tag: $(params.image_prefix)$$(workload.metadata.name)$
      serviceAccountName: service-account
      builder:
        kind: ClusterBuilder
        name: go-builder
      source:
        blob:
          url: $(sources.source.url)$
      build:
        env: $(workload.spec.build.env)$
```

### ytt

```yaml
---
apiVersion: carto.run/v1alpha1
kind: ClusterImageTemplate
metadata:
  name: kpack-template
spec:
  imagePath: .status.latestImage
  params:
    - name: serviceAccount
      default: default
    - name: clusterBuilder
      default: default
    - name: registry
      default: {}
  ytt: |
    #@ load("@ytt:data", "data")

    #@ def image():
    #@   return "/".join([
    #@    data.values.params.registry.server,
    #@    data.values.params.registry.repository,
    #@    "-".join([
    #@      data.values.workload.metadata.name,
    #@      data.values.workload.metadata.namespace,
    #@    ])
    #@   ])
    #@ end

    apiVersion: kpack.io/v1alpha2
    kind: Image
    metadata:
      name: #@ data.values.workload.metadata.name
      labels:
        app.kubernetes.io/component: build
        #@ if/end hasattr(data.values.workload.metadata, "labels") and hasattr(data.values.workload.metadata.labels, "app.kubernetes.io/part-of"):
        app.kubernetes.io/part-of: #@ data.values.workload.metadata.labels["app.kubernetes.io/part-of"]
    spec:
      tag: #@ image()
      serviceAccountName: #@ data.values.params.serviceAccount
      builder:
        kind: ClusterBuilder
        name: #@ data.values.params.clusterBuilder
      source:
        blob:
          url: #@ data.values.source.url
        #@ if/end hasattr(data.values.workload.spec.source, "subPath"):
        subPath: #@ data.values.workload.spec.source.subPath
      build:
        env:
        - name: BP_OCI_SOURCE
          value: #@ data.values.source.revision
        #@ if hasattr(data.values.workload.spec.build, "env"):
        #@ for var in data.values.workload.spec.build.env:
        - name: #@ var.name
          value: #@ var.value
        #@ end
        #@ end
```
