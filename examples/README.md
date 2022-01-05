# Examples

These examples demonstrate how common patterns can be implemented in Cartographer.
Each example demonstrates increasing assurance steps along the
path to production. The suggested order is:

- [Supply Chain: Source => Image => App](basic-sc/README.md): demonstrates how
  one can set up a `ClusterSupplyChain` that monitors source code updates, creates
  a new image for each update and deploys that image as a running application
  in the cluster.
- [Runnable: Updating a Tekton Task without triggers](runnable-tekton/README.md):
  demonstrates how Cartographer's `Runnable` wraps immutable objects to provide
  updatable behavior. The example uses a Tekton Task that tests source code and
  demonstrates how to create [TaskRuns](https://tekton.dev/docs/pipelines/taskruns/)
  with objects that don't require triggers or events.
- [Supply Chain: Source => Test => Image => App](testing-sc/README.md): expands upon
  the example above by adding testing, assuring that only commits with passing tests
  result in deployed apps.
- [Supply Chain: Source => Image => Git](gitwriter-sc/README.md): whereas the first
  example deploys an app, this example takes the app configuration and writes it to git. This
  configuration can then be read by a Delivery in another cluster. This allows all
  build processes to be isolated in a cluster completely separate from the
  production cluster exposed to customers and end users.