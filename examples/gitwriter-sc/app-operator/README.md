# App Operator

In this directory you'll find Kubernetes objects necessary for the app
operators to submit to Kubernetes:

- a cluster config template that instantiates a `ConfigMap` containing Kubernetes
  configuration for a knative `Service`. This example uses
  ytt instead of a simple template for more complex templating.
- a tekton cluster task that writes objects to git.
- a cluster run template that creates a taskrun of the tekton cluster task.
- a cluster template `git-writer` that creates a runnable to write the ConfigMap data to git.
- a supply chain that references these templates

Note: The supply chain also references resources that are created in the examples/shared
directory.
