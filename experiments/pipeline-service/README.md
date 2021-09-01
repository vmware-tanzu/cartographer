# pipeline service

_**THIS IS JUST A DRAFT/WORK IN PROGRESS/EXPERIMENT**, do not take it seriously
yet (specially the name)_.

`pipeline-service` extends Kubernetes to provide the ability to express (in a
declarative form) the desire to have a pipeline run and ensuring that it does
so on any change to its specification. think of it as `kpack` but for pipeline
runners (tekton, argo workflows, jenkins, concourse jobs...), or, more broadly,
anything that has a "submit an invocation object to _run_" semantic.


## trying out

Under [`./examples`](./examples) you'll find a walkthrough of pipeline-service
going from simple to elaborate examples.

In the elaborate case (pipeline in a cartographer supplychain), it continuously
runs tests for a Go app on every commit, then builds a container image that
gets pushed to a container image registry, then deploys that using a plain
Kubernetes Deployment.



                                 SupplyChain


                            (src)                  (src)                 (img)
        SOURCE-PROVIDER  <--------- TESTS ----------------------  IMAGE --------- DEPLOYMENT
             .                        .                             .                 .
             .                        .                             .                 .
             .               carto.run/Pipeline                   .           appsv1/Deployment
      flux/GitRepository              |                        kpack/Image
                                      '                             |     
                            tekton.dev/PipelineRun                  '     
                                                               kpack/Build
                                                          
                                                          

Check out [./examples](./examples).


### installing

You can find a release manifest under the [releases directory](./releases). 

All it takes to install pipeline-service is submitting that manifest to your
Kubernetes cluster:

```bash
kapp deploy -a pipeline-service -f ./releases/release.yaml
```
```console
TODO
```

Somes examples in this repository depends on a few other controllers (flux'
source-controller, cert-manager, cartographer, kpack, and tekton), which must
be installed if you want to run them.

Make sure you check out the examples' READMEs before proceeding with them.
