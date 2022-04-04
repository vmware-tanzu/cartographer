# Runnable: Templating Objects That Cannot Update

## Overview

In this tutorial we’ll explore a new Cartographer resource: runnable. Runnables will enable us to choreograph resources
that do not support standard update behavior. We’ll see how we can wrap Runnables around tools useful for testing, like
Tekton, and how that will provide us easy updating behavior of our testing objects.

## Environment setup

For this tutorial you will need a kubernetes cluster with Cartographer and Tekton installed. You can find
[Cartographer's installation instructions here](https://github.com/vmware-tanzu/cartographer#installation) and
[Tekton's installation instructions are here](https://github.com/pivotal/kpack/blob/main/docs/install.md).

Alternatively, you may choose to use the
[./hack/setup.sh](https://github.com/vmware-tanzu/cartographer/blob/main/hack/setup.sh) script to install a kind cluster
with Cartographer and Tekton. _This script is meant for our end-to-end testing and while we rely on it working in that
role, no user guarantees are made about the script._

Command to run from the Cartographer directory:

```shell
$ ./hack/setup.sh cluster cartographer-latest example-dependencies
```

If you later wish to tear down this generated cluster, run

```shell
$ ./hack/setup.sh teardown
```

## Scenario

Our CTO is interested in putting quality controls in place; only code that passes certain checks should be built and
deployed. They want to start small: all source code repositories that are built must pass markdown linting. In order to
do this we’re going to leverage
[the markdown linting pipeline in the TektonCD catalog](https://github.com/tektoncd/catalog/tree/main/task/markdown-lint/0.1).

In this tutorial we’ll see how to use Cartographer’s Runnable to give us easy updating behavior of Tekton (no need for
Tekton Triggers and Github Webhooks). In [the next tutorial](runnable-in-a-supply-chain.md) we’ll complete the scenario
by using Runnable in a supply chain.

## Steps

### Tekton Basics

Before using Cartographer, let’s think about how we would use Tekton on its own to lint a repo. First we would define a
pipeline:

{{< tutorial pipeline.yaml >}}

We would apply this pipeline to the cluster, along with the tasks. Those tasks are in the TektonCD Catalog:

```shell
$ kubectl apply -f https://raw.githubusercontent.com/tektoncd/catalog/main/task/git-clone/0.3/git-clone.yaml
$ kubectl apply -f https://raw.githubusercontent.com/tektoncd/catalog/main/task/markdown-lint/0.1/markdown-lint.yaml
```

Finally, we need to create a pipeline-run object. This object provides the param and workspace values defined at the top
level `.spec` field of the pipeline.

```yaml
---
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: linter-pipeline-run
spec:
  pipelineRef:
    name: linter-pipeline
  params:
    - name: repository
      value: https://github.com/waciumawanjohi/demo-hello-world
    - name: revision
      value: main
  workspaces:
    - name: shared-workspace
      volumeClaimTemplate:
        spec:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 256Mi
```

Importantly, this pipeline-run object will kick off a single run of the pipeline. The run will either succeed or fail.
The outcome will be written into the pipeline-run’s status. No later changes to the pipeline-run object will change
those outcomes; the run happens once.

To see this in action, let’s apply the above pipeline-run. If we watch the object, we’ll soon see that it succeeds.

```shell
$ watch 'kubectl get -o yaml pipelinerun linter-pipeline-run | yq .status.conditions'
```

Eventually yields the result:

```yaml
- lastTransitionTime: ...
  message: "Tasks Completed: 2 (Failed: 0, Cancelled 0), Skipped: 0"
  reason: "Succeeded"
  status: "True"
  type: "Succeeded"
```

### Templating Pipeline Runs

This seems like a very easy step to encode in a supply chain. All we would need to do is ensure that the Tekton tasks
and pipeline are created beforehand. Then a supply chain could stamp out a templated pipeline-run object. This template
will pull the repository and revision value from the workload. But there’s a problem... what happens if an app dev
changes one of these values on the workload? Our supply chain would not properly reflect the change, because the
pipeline-run object cannot be updated!

Fortunately there’s an easy fix. We’re going to use a pair of new Cartographer resources: Runnable and
ClusterRunTemplate.

It will be the responsibility of the ClusterRunTemplate to template our desired object (in this case, our Tekton
pipeline-run). The Runnable will be responsible for providing values to the ClusterRunTemplate, the values that we
expect to vary (in our example, the url and revision of the source code). When the set of values from the Runnable
changes, a new object will be created (rather than the old object updated).

The Runnable will also expose results. Of course, multiple results will exist, a result for each of the objects created.
Runnable will only expose the results from the most recently submitted successful object.

In this manner, we get a wrapper (Runnable) that is updateable and updates results in its status. This is similar to the
default behavior of kuberenetes objects. By wrapping Tekton pipeline-runs (or any immutable resource) in a Runnable, we
will be able to use the resource in a supply chain as if it were mutable.

Let’s see the Runnable and the ClusterRunTemplate at work. Once we’re solid on those, we’ll use Runnable in a Supply
Chain in the next tutorial.

#### ClusterRunTemplate

Let’s start with the ClusterRunTemplate. As can be expected from the name, there’s a `.spec.template` field in it, where
we will write something very similar to our pipeline-run above. In fact, let’s write that exact pipeline-run in and then
look at the values that will need to change:

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterRunTemplate
metadata:
  name: md-linting-pipelinerun
spec:
  template:
    apiVersion: tekton.dev/v1beta1
    kind: PipelineRun
    metadata:
      name: linter-pipeline-run # <=== Can’t all have the same name
    spec:
      pipelineRef:
        name: linter-pipeline
      params: # <=== These param values will change
        - name: repository
          value: https://github.com/waciumawanjohi/demo-hello-world
        - name: revision
          value: main
      workspaces:
        - name: shared-workspace
          volumeClaimTemplate:
            spec:
              accessModes:
                - ReadWriteOnce
              resources:
                requests:
                  storage: 256Mi
```

Most fields are fine. The name field is not. Why not? If we want to change the values for the pipeline-run, we’re not
going to update the templated object. We’re going to create an entirely new object. And of course that new object can’t
have the same hardcoded name. To handle this, every object templated in a ClusterRunTemplate specifies a `generateName`
rather than a `name`. We can use `linter-pipeline-run-` and kubernetes will handle putting a unique suffix on the name
of each pipeline-run.

<!-- prettier-ignore-start -->
```yaml
    metadata:
      generateName: linter-pipeline-run-
```
<!-- prettier-ignore-end -->

The other change we want to make is to the values on the params. It doesn’t do much good to hardcode
`https://github.com/waciumawanjohi/demo-hello-world` into the repository param; this is the value we want to update. As
we said, we intend to update the Runnable and have the value in that update stamped out in a new pipeline-run object. So
we replace the hardcoded value with a Cartographer parameter. We’ll find the value that we want on the runnable. That
will look like `$(runnable.spec.inputs.repository)$`. This specifies that the value we want will be found in the
runnable spec, in a field named inputs, one of which will have the name `repository`.

There’s only one more thing that we need to specify on the ClusterRunTemplate: what the outputs will be! We said that
the Runnable status will reflect results of a successful run. The ClusterRunTemplate specifies what results. For now
let’s simply report the lastTransition time that we saw on the conditions. We use jsonpath to indicate the location of
this value on the objects that are being stamped:

<!-- prettier-ignore-start -->
```yaml
  outputs:
    lastTransitionTime: .status.conditions[0].lastTransitionTime
```
<!-- prettier-ignore-end -->

Let’s look at our complete ClusterRunTemplate:

{{< tutorial cluster-run-template.yaml >}}

#### Runnable

Now let’s make the Runnable. First we’ll specify which ClusterRunTemplate our Runnable works with. We do this in the
Runnable's `.spec.runTemplateRef.` field and we refer to the name of the ClusterRunTemplate we just created.

<!-- prettier-ignore-start -->
```yaml
  runTemplateRef:
    name: md-linting-pipelinerun
```
<!-- prettier-ignore-end -->

Next we’ll fill in the inputs. These are the paths that we templated into the ClusterRunTemplate param values, the
`runnable.spec.inputs.repository` and `runnable.spec.inputs.revision`. We’ll fill these with the values we previously
hardcoded in the pipelinerun. With runnable we’ll be able to update them later as we like.

<!-- prettier-ignore-start -->
```yaml
  inputs:
    repository: https://github.com/waciumawanjohi/demo-hello-world
    revision: main
```
<!-- prettier-ignore-end -->

Finally, we need a serviceAccountName. Just like the supply-chain, with Runnable Cartographer could be stamping out
_anything_. Using RBAC we expect a service account to provide permissions to Cartographer and limit it to creating only
the types of objects we expect. We'll create a service account named `pipeline-run-management-sa`. We’ll put that name
in our Runnable object. The full object looks like this:

{{< tutorial runnable.yaml >}}

Let’s quickly create the service account we referenced:

{{< tutorial service-account.yaml >}}

Great! Let’s deploy these objects.

## Observe

Let’s observe the pipeline-run objects in the cluster:

```shell
$ kubectl get pipelineruns
```

We can see that a new pipelinerun has been created with the `linter-pipeline-run-` prefix:

```console
NAME                        SUCCEEDED   REASON      STARTTIME   COMPLETIONTIME
linter-pipeline-run-123az   True        Succeeded   2m48s       2m35s
```

Examining the created object it’s a non-trivial 300 lines:

```shell
$ kubectl get -o yaml pipelineruns linter-pipeline-run-123az
```

In the metadata we can see familiar labels indicating Carto objects used to create this templated object. We can also
see that the object is owned by the runnable.

```yaml
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: linter-pipeline-run-123az
  generateName: linter-pipeline-run-
  labels:
    carto.run/run-template-name: md-linting-pipelinerun
    carto.run/runnable-name: linter
    tekton.dev/pipeline: linter-pipeline
  ownerReferences:
    - apiVersion: carto.run/v1alpha1
      blockOwnerDeletion: true
      controller: true
      kind: Runnable
      name: linter
      uid: ...
  ...
```

The spec contains the spec that we templated out. Looks great.

```yaml
spec:
  params:
    - name: repository
      value: https://github.com/waciumawanjohi/demo-hello-world
    - name: revision
      value: main
  pipelineRef:
    name: linter-pipeline
  serviceAccountName: default
  timeout: 1h0m0s
  workspaces:
    - name: shared-workspace
      volumeClaimTemplate:
        metadata:
          creationTimestamp: null
        spec:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 256Mi
        status: {}
```

The status contains fields expected of Tekton:

```yaml
status:
  completionTime: ...
  conditions:
    - lastTransitionTime: "2022-03-07T19:24:35Z"
      message: "Tasks Completed: 2 (Failed: 0, Cancelled 0), Skipped: 0"
      reason: Succeeded
      status: "True"
      type: Succeeded
  pipelineSpec: ...
  startTime: ...
  taskRuns: ...
```

To learn more about Tekton’s behavior, readers will want to refer to [Tekton documentation](https://tekton.dev/docs/).

Now we examine the Cartographer Runnable object. We expect it to expose values from our successful object.

```shell
$ kubectl get -o yaml runnable linter
```

```yaml
apiVersion: carto.run/v1alpha1
kind: Runnable
Metadata: ...
Spec: ...
status:
  conditions: ...
  observedGeneration: ...
  outputs:
    lastTransitionTime: "2022-03-07T19:24:35Z"
```

Wonderful! The value from the field we specified in the ClusterRunTemplate is now in the outputs of the Runnable.

Finally, let's update our runnable with a new repository:

```yaml
apiVersion: carto.run/v1alpha1
kind: Runnable
metadata:
  name: linter
spec:
  runTemplateRef:
    name: md-linting-pipelinerun
  inputs:
    repository: https://github.com/kelseyhightower/nocode # <=== new repo
    revision: master # <=== new revision
  serviceAccountName: pipeline-run-management-sa
```

When we apply this to the cluster, we can observe:

- The spec of our runnable is updated
- The runnable causes the creation of a new pipelinerun
- The new pipeline run fails because the new repo does not pass linting
- Since the result of the new pipeline run is failure, our runnable output remains the output of our previous
  (successful) pipeline run

## Wrap Up

In this tutorial you learned:

- Some useful kubernetes objects cannot be updated
- Runnable allows you to add updateable behavior to those objects
- How to write Runnables with input values for ClusterRunTemplates
- How to write ClusterRunTemplates that specify outputs for the Runnable
- How to read output values from the Runnable

We’ve got an updateable object, Runnable, that can manage tekton pipelines and tasks. Our next step is going to be using
Runnable in a supply chain.
