# Runnable Custom Resources

## Runnable

A `Runnable` object declares the intention of having immutable objects submitted to Kubernetes according to a template (
via ClusterRunTemplate) whenever any of the inputs passed to it changes. i.e., it allows us to provide a mutable spec
that drives the creation of immutable objects whenever that spec changes.

{{< crd  carto.run_runnables.yaml >}}

## ClusterRunTemplate

A `ClusterRunTemplate` defines how an immutable object should be stamped out based on data provided by a `Runnable`.

{{< crd  carto.run_clusterruntemplates.yaml >}}

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
