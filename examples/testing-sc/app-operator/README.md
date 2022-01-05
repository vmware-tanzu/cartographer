# App Operator

In this directory you'll find Kubernetes objects necessary for the app
operators to submit to Kubernetes:

- a cluster run template that accepts values from a runnable
- a cluster source template that creates a runnable
- a supply chain that includes the cluster source template

Note: The supply chain also references resources that are created in the examples/shared
directory.
