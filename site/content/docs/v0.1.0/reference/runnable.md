# Runnable Custom Resources

## Runnable

A `Runnable` object declares the intention of having immutable objects submitted to Kubernetes according to a template (
via ClusterRunTemplate) whenever any of the inputs passed to it changes. i.e., it allows us to provide a mutable spec
that drives the creation of immutable objects whenever that spec changes.

```yaml
apiVersion: carto.run/v1alpha1
kind: Runnable
metadata:
  name: test-runner
spec:
  # service account with permissions to create resources submitted by the runnable
  # if not set, will use the default service account in the runnable's namespace
  #
  serviceAccountName: runnable-service-account

  # data to be made available to the template of ClusterRunTemplate
  # that we point at.
  #
  # this field takes as value an object that maps strings to values of any
  # kind, which can then be reference in a template using jsonpath such as
  # `$(runnable.spec.inputs.<key...>)$`.
  #
  # (required)
  #
  inputs:
    serviceAccount: bla
    params:
      - name: foo
        value: bar

  # reference to a ClusterRunTemplate that defines how objects should be
  # created referencing the data passed to the Runnable.
  #
  # (required)
  #
  runTemplateRef:
    name: job-runner

  # an optional selection rule for finding an object that should be used
  # together with the one being stamped out by the runnable.
  #
  # an object found using the rules described here are made available during
  # interpolation time via `$(selected.<...object>)$`.
  #
  # (optional)
  #
  selector:
    resource:
      kind: Pipeline
      apiVersion: tekton.dev/v1beta1
    matchingLabels:
      pipelines.foo.bar: testing
```

## ClusterRunTemplate

A `ClusterRunTemplate` defines how an immutable object should be stamped out based on data provided by a `Runnable`.

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterRunTemplate
metadata:
  name: image-builder
spec:
  # data to be gathered from the objects that it interpolates once they #
  # succeeded (based on the object presenting a condition with type 'Succeeded'
  # and status `True`).
  #
  # (optional)
  #
  outputs:
    # e.g., make available under the Runnable `outputs` section in `status` a
    # field called "latestImage" that exposes the result named 'IMAGE-DIGEST'
    # of a tekton task that builds a container image.
    #
    latestImage: .status.results[?(@.name=="IMAGE-DIGEST")].value

  # definition of the object to interpolate and submit to kubernetes.
  #
  # data available for interpolation:
  #   - `runnable`: the Runnable object that referenced this template.
  #
  #                 e.g.:  params:
  #                        - name: revison
  #                          value: $(runnable.spec.inputs.git-revision)$
  #
  #
  #   - `selected`: a related object that got selected using the Runnable
  #                 selector query.
  #
  #                 e.g.:  taskRef:
  #                          name: $(selected.metadata.name)$
  #                          namespace: $(selected.metadata.namespace)$
  #
  # (required)
  #
  template:
    apiVersion: tekton.dev/v1beta1
    kind: TaskRun
    metadata:
      generateName: $(runnable.metadata.name)$-
    spec:
      serviceAccountName: $(runnable.spec.inputs.serviceAccount)$
      taskRef: $(runnable.spec.inputs.taskRef)$
      params: $(runnable.spec.inputs.params)$
```

ClusterRunTemplate differs from supply chain templates in many aspects:

- ClusterRunTemplate cannot be referenced directly by a ClusterSupplyChain object (it can only be reference by a
  Runnable)

- `outputs` provide a free-form way of exposing any form of results from what has been run (i.e., submitted by the
  Runnable) to the status of the Runnable object (as opposed to typed "source", "image", and "config" from supply
  chains)

- Templating context (values provided to the interpolation) is specific to the Runnable: the runnable object itself and
  the object resulting from the selection query.

- Templated object metadata.name should not be set. differently from ClusterSupplyChain, a Runnable has the semantics of
  creating new objects on change, rather than patching. This means that on every input set change, a new name must be
  derived. To be sure that a name can always be generated, `metadata.generateName` should be set rather than
  `metadata.name`.

Similarly to other templates, ClusterRunTemplate has a `template` field where data is taken (in this case, from Runnable
and selected objects via `runnable.spec.selector`) and via `$()$` allows one to interpolate such data to form a final
object.
