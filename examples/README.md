# Examples

These examples demonstrate how common patterns can be implemented in Cartographer. Each example demonstrates increasing
assurance steps along the path to production. The suggested order is:

- [Supply Chain: Source => Image => App](basic-sc/README.md): demonstrates how one can set up a `ClusterSupplyChain`
  that monitors source code updates, creates a new image for each update and deploys that image as a running application
  in the cluster.
- [Runnable: Updating a Tekton Task without triggers](runnable-tekton/README.md):
  demonstrates how Cartographer's `Runnable` wraps immutable objects to provide updatable behavior. The example uses a
  Tekton Task that tests source code and demonstrates how to
  create [TaskRuns](https://tekton.dev/docs/pipelines/taskruns/)
  with objects that don't require triggers or events.
- [Supply Chain: Source => Test => Image => App](testing-sc/README.md): expands upon the example above by adding
  testing, assuring that only commits with passing tests result in deployed apps.
- [Supply Chain: Source => Image => Git](gitwriter-sc/README.md): whereas the first example deploys an app, this example
  takes the app configuration and writes it to git. This configuration can then be read by a Delivery in another
  cluster. This allows all build processes to be isolated in a cluster completely separate from the production cluster
  exposed to customers and end users.
- [Delivery: Git => App](basic-delivery/README.md): this example picks up app configuration that has been written to
  git (it is written to work with the gitwriter example above). The configuration is then applied to the cluster. This
  minimizes processes occuring in the production cluster exposed to customers and end users.

## Resource Requirements

### Usage

When the examples are run, they generally consume 4 cpu cores and 4 GiB of memory.

### Requests

The examples rely on a number of controllers being installed. These controllers create deployments which request
resources from the kubernetes cluster at rest. All told, the controllers request 1.7 cpu cores and 1.6 GiB of memory
"at rest".

When Cartographer runs these examples in testing, we
use [the following overlay](../hack/overlays/remove-resource-requests-from-deployments.yaml) to remove these requests.
