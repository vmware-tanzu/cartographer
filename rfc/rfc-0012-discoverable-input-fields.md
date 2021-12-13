# Draft RFC 0012 Discoverable Input Fields

## Summary

Templates and supply chains should report in their status what input fields they expect from a workload.

## Motivation

1. We want supply chain authors to be empowered to quickly swap templates.
In order to do so, they must verify that the workloads currently matched provide
all of the inputs expected by the new supply chain.
2. We want workload authors to be able to quickly know what inputs their supply chain expects.

## Possible Solutions

When a template is submitted to the cluster it is reconciled by a controller.
That controller adds a status field to the template. In the status there is a field: `expectedInputs`.
This is an array of
1. All of the named params used in the template.
2. All of the workload fields used in the template.

When a supply chain is submitted to the cluster, the supply chain controller updates its status
with all of the fields required by the referenced templates.

### Example

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSourceTemplate
metadata:
  name: test
spec:
  urlPath: .status.outputs.url
  revisionPath: .status.outputs.revision

  template:
    apiVersion: carto.run/v1alpha1
    kind: Runnable
    metadata:
      name: $(workload.metadata.name)$
    spec:
      serviceAccountName: $(workload.spec.serviceAccountName)$

      runTemplateRef:
        name: tekton-pipelinerun

      selector:
        resource:
          apiVersion: tekton.dev/v1beta1
          kind: Task
        matchingLabels:
          apps.tanzu.vmware.com/task: test
          some.other.label: $(params.other-label)$

      inputs:
        source: $(source)$
        params:
          - name: blob-url
            value: $(source.url)$
          - name: blob-revision
            value: $(source.revision)$
status:
  inputs:
    workload:
      metadata:
      - name
      spec:
      - serviceAccountName
    params:
      - other-label
    sources:
    - ANYNAME:
      - url
      - revision
```

## Cross References and Prior Art

{{Reference other similar implementations, and resources you are using to draw inspiration from}}
