# Examples

Here you'll find a guided tour of how one can make use of `pipeline-service`, going from using Tekton in a standalone manner, to a full integration within a [cartographer] supply chain.


1. [Standalone Tekton](./01-tekton-standalone): running Tekton pipelines by
   hand to build up the reasoning behind something like pipeline-service.
2. [Triggering Tekton pipelines using pipeline
   service](./02-simple-pipeline-service): using the declarative nature of
   pipeline-service to trigger those previously manually submitted pipeline
   runs.
3. [Reusing RunTemplate and Tekton
   pipelines](./03-pipeline-service-with-selector): making use of the ability
   to interpolate invocations using extra objects we build a Tekton pipeline
   that runs scripts that developers provide via ConfigMaps.
4. [Integrating Pipeline in a simple SupplyChain](./04-simple-supply-chain):
   similar to second example but integrating GitRepository with Pipeline so
   that we run tests on every commit to a repository
5. [Complete SupplyChain with developer-provided Tekton
   Pipelines](./05-complete-supply-chain): a definition of a supply chain from
   fetching a commit, to running tests, building a container image, and
   ultimately deploying it
6. [SupplyChain with developer-provided
   ConfigMaps](./06-supply-chain-configmap): similar to the previous example,
   but reduced to just continuously run tests from configmaps (no image
   building)


[cartographer]: https://github.com/vmware-tanzu/cartographer
