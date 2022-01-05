# App Operator

In this directory you'll find Kubernetes objects necessary for the app
operators to submit to Kubernetes:

- a cluster config template that instantiates a `ConfigMap` containing Kubernetes
  configuration for a knative `Service`. This example uses
  ytt instead of a simple template for more complex templating.
- a secret with credentials for pushing to a git repository
- a cluster run template that creates a taskrun for writing objects to git.
- a cluster template `git-writer` that writes the ConfigMap data to git.
- a supply chain that references these templates

Note: The supply chain also references resources that are created in the examples/shared
directory.
