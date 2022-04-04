# Using Runnable in a Supply Chain

## Overview

In the previous tutorial we saw how Runnable brings updateable behavior to immutable kubernetes objects. In this
tutorial, we’ll see how we can use Runnable in our supply chains for common behavior like linting, scanning, and
testing.

## Environment setup

For this tutorial you will need a kubernetes cluster with Cartographer, kpack and Tekton installed. You can find
[Cartographer's installation instructions here](https://github.com/vmware-tanzu/cartographer#installation), kpack's
[here](https://github.com/pivotal/kpack/blob/main/docs/install.md) and Tekton's
[here](https://github.com/pivotal/kpack/blob/main/docs/install.md).

You will also need an image registry for which you have read and write permission.

You may choose to use the [./hack/setup.sh](https://github.com/vmware-tanzu/cartographer/blob/main/hack/setup.sh) script
to install a kind cluster with Cartographer, Tekton, kpack and a local registry. _This script is meant for our
end-to-end testing and while we rely on it working in that role, no user guarantees are made about the script._

Command to run from the Cartographer directory:

```shell
$ ./hack/setup.sh cluster cartographer-latest example-dependencies
```

If you later wish to tear down this generated cluster, run

```shell
$ ./hack/setup.sh teardown
```

## Scenario

Continuing the scenario from [the previous tutorial](runnable.md), we remember that our CTO is interested in putting
quality controls in place; only code that passes certain checks should be built and deployed. They want to start small,
and have decided all source code repositories that are built must pass markdown linting. In order to do this we’re going
to leverage
[the markdown linting pipeline in the TektonCD catalog](https://github.com/tektoncd/catalog/tree/main/task/markdown-lint/0.1).

Last tutorial we saw how to use Cartographer’s Runnable to give us easy updating behavior of Tekton (no need for Tekton
Triggers and Github Webhooks). In this following tutorial we’ll complete the scenario by using Runnable in a supply
chain.

## Steps

### App Operator Steps

Much of our work from the previous tutorial remains the same. We deployed a Tekton pipeline and two Tekton Tasks in the
cluster. These objects will be applied in this tutorial with no change.

We deployed a ClusterRunTemplate that templated a Tekton PipelineRun. This will stay largely the same, with only a
change to the outputs (We're going to define outputs that are slightly more useful than the ones we chose last
tutorial). We will still deploy this object directly.

The object that we’ll deploy differently is the Runnable. To use Runnable in a supply chain we’ll wrap it in a
Cartographer template. This template will be referenced in our supply chain. This work should feel very familiar to the
steps we took in the [Build Your First Supply Chain](first-supply-chain.md) tutorial!

#### Supply Chain

Let’s begin by thinking through what template we need and where it will go in our supply chain. Our goal is to ensure
that the only repos that are built and deployed are those that pass linting. So we’ll need our new step to be the first
step in a supply chain. This step will receive the location of a source code and if the source code passes linting it
will pass that location inforation to the next step in the supply chain. Do you remember what template is meant to
expose information about the location of source code? That’s right, the ClusterSourceTemplate.

Let’s define our supply chain now. We’ll start with the supply chain we created in the Extending a Supply Chain
tutorial. The resources then looked like this:

<!-- prettier-ignore-start -->
```yaml
  resources:
    - name: build-image
      templateRef:
        kind: ClusterImageTemplate
        name: image-builder
    - name: deploy
      templateRef:
        kind: ClusterTemplate
        name: app-deploy-from-sc-image
      images:
        - resource: build-image
          name: built-image
```
<!-- prettier-ignore-end -->

We’ll add a new first step, lint source code. As we determined before, this will refer to a ClusterSourceTemplate. Our
second step will remain a ClusterImageTemplate, but it will have to be a new template. This is because it will consume
the source code from the previous step rather than directly from the workload. The rest of the resources will remain the
same.

<!-- prettier-ignore-start -->
```yaml
  resources:
    - name: lint-source
      templateRef:
        kind: ClusterSourceTemplate
        Name: source-linter
    - name: build-image
      templateRef:
        kind: ClusterImageTemplate
        name: image-builder-from-previous-step
      sources:
      - resource: lint-source
        name: source
    - name: deploy
      templateRef:
        kind: ClusterTemplate
        name: app-deploy-from-sc-image
      images:
        - resource: build-image
          name: built-image
```
<!-- prettier-ignore-end -->

Our final step with the supply chain will be referring to a service-account. Let's think through what permissions we
need:

- the `source-linter` template will create a runnable
- the `image-builder-from-previous-step` template will create a kpack image (just as the supply chain from the
  [Extending a Supply Chain](extending-a-supply-chain.md) tutorial)
- the `app-deploy-from-sc-image` template will create a deployment (just as the supply chain from the
  [Extending a Supply Chain](extending-a-supply-chain.md) tutorial)

The only new object created here is a runnable, which is a Cartographer resource. The Cartographer controller already
has permission to manipulate Cartographer resources. So we do not need to do any alterations, we can simply reuse the
service account (and roles and role bindings) from the [Extending a Supply Chain](extending-a-supply-chain.md) tutorial.

_Note that while the supply chain refers to a service account, the Runnable itself also refers to a service account.
More on that in a moment._

Here is our complete supply chain.

{{< tutorial supply-chain.yaml >}}

#### Templates

There are two new templates that need to be written, `image-builder-from-previous-step` and `source-linter`. Creating
the `image-builder-from-previous-step` will be left as an exercise for the reader. Refer to the Extending a Supply Chain
tutorial for help. Let's turn to creating the `source-linter` template.

We know we’ll wrap our Runnable in a ClusterSourceTemplate. We’ll begin as always, taking our previously created
Runnable and simply pasting it into a ClusterSourceTemplate:

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSourceTemplate
metadata:
  name: source-linter
spec:
  template:
    apiVersion: carto.run/v1alpha1
    kind: Runnable
    metadata:
      name: linter
    spec:
      runTemplateRef:
        name: md-linting-pipelinerun
      inputs:
        repository: https://github.com/waciumawanjohi/demo-hello-world
        revision: main
      serviceAccountName: pipeline-run-management-sa
  urlPath: ???
  revisionPath: ???
```

We can quickly see that there are hardcoded values that we’ll want to replace with references to the workload (so that
our supply chain can work with many different workloads). These are the `inputs` values. Remember, these fields are
inputs to the ClusterRunTemplate. If a value could be hardcoded, it should not be a field in the Runnable’s inputs at
all, it should simply be hardcoded in the ClusterRunTemplate (e.g. every field in the runnable `spec.inputs` should
become a Cartographer variable). Let’s look at how we’ll change `inputs`:

<!-- prettier-ignore-start -->
```yaml
      inputs:
        repository: $(workload.spec.source.git.url)$
        revision: $(workload.spec.source.git.revision)$
```
<!-- prettier-ignore-end -->

Quite straightforward and familiar. Simply pull the requisite values from the workload. The next step will be
straightforward as well; we need to make sure that multiple apps don’t overwrite one Runnable object. We need to
template the Runnable’s name:

<!-- prettier-ignore-start -->
```yaml
    metadata:
      name: $(workload.metadata.name)$-linter
```
<!-- prettier-ignore-end -->

Finally, we need to fill the `urlPath` and `revisionPath` to tell the ClusterSourceTemplate what field of the Runnable
to expose for the `url` and `revision`. Remember, that’s the contract of a ClusterSourceTemplate, it exposes those two
values to the rest of the supply chain. Runnables have a `.status` field, which we'll use. The contents of that field
are determined by fields on the ClusterRunTemplate. In the [previous tutorial](runnable.md) the ClusterRunTemplate
declared that the output would be called `lastTransitionTime`. Let's declare our intention now to change the
ClusterRunTemplate. In a moment we'll alter it to set new outputs named `url` and `revision`. That will allow us to
finish the ClusterSourceTemplate wrapping our runnable.

```yaml
spec:
  template: ...
  urlPath: .status.outputs.url
  revisionPath: .status.outputs.revision
```

Wonderful. Our ClusterSourceTemplate is complete:

{{< tutorial source-linter-template.yaml >}}

#### ClusterRunTemplate

Our supply-chain will now stamp out a Runnable. But we still have to change the ClusterRunTemplate to ensure that the
status of the Runnable has the fields we want. Just a moment ago we decided that these fields should be `url` and
`revision`. Before we alter the previous ClusterRunTemplate from the previous tutorial, let's look at it:

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
      generateName: linter-pipeline-run-
    spec:
      pipelineRef:
        name: linter-pipeline
      params:
        - name: repository
          value: $(runnable.spec.inputs.repository)$
        - name: revision
          value: $(runnable.spec.inputs.revision)$
      workspaces:
        - name: shared-workspace
          volumeClaimTemplate:
            spec:
              accessModes:
                - ReadWriteOnce
              resources:
                requests:
                  storage: 256Mi
  outputs:
    lastTransitionTime: .status.conditions[0].lastTransitionTime
```

We can see that the `spec.outputs` field was used to prescribe a `lastTransitionTime` field. We’ll change that to a
`url` and `revision` field:

<!-- prettier-ignore-start -->
```yaml
  outputs:
    url: ???
    revision: ???
```
<!-- prettier-ignore-end -->

Great. Now let’s think; where on the object that’s created will we find these values. Actually, why think about it;
let’s take a look at the pipelinerun that was created in the last tutorial!

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
status:
  completionTime: ...
  conditions: ...
  pipelineSpec: ...
  startTime: ...
  taskRuns: ...
```

While it feels most natural to read outputs from the `.status` of objects, in this case the value we want is in the
`.spec`. We can see that the `.spec.params` of this object has the url and revision of the source code. That's the value
we wish to pass on if the pipeline-run is successful. We'll use jsonpath in our outputs to grab these values:

<!-- prettier-ignore-start -->
```yaml
  outputs:
    url: spec.params[?(@.name=="repository")].value
    revision: spec.params[?(@.name=="revision")].value
```
<!-- prettier-ignore-end -->

Great, our outputs are complete.

There’s one more thing that we can do to make things easy on ourselves, differentiate the name of pipeline-runs of one
workload from those of another. Technically, we do not have to do this. Cartographer is smart enough to stamp out these
objects and differentiate objects created from the same prefix from Runnable 1 and Runnable 2. But we’re human and it’ll
be nice for us to quickly be able to distinguish.

```yaml
spec:
  template:
    metadata:
      generateName: $(runnable.metadata.name)$-pipeline-run-
```

Let’s pull it all together! Our ClusterRunTemplate now reads:

{{< tutorial cluster-run-template.yaml >}}

#### Runnable's Service Account

The last object to mention is the serviceAccount object referred to in the ClusterSourceTemplate wrapped Runnable. We
could refer to the same service account referred to in the supply chain. Then we would simply bind an additional role to
the account, one that allows creation of a Tekton PipelineRun. Or we can refer to a different service account with these
permissions. We created such a service account in the [previous tutorial](runnable.md) and our runnable still refers to
that name. We will simply deploy that service account.

Now we're ready. Let’s submit all these objects and step into our role as App Devs.

### App Dev Steps

As devs, our work is easy! We submit a workload. We’re being asked for the same information as ever from the templates,
a url and a revision for the location of the source code. We can submit the same workload from the
[Extending a Supply Chain tutorial](extending-a-supply-chain.md):

{{< tutorial workload.yaml >}}

## Observe

Using [kubectl tree](https://github.com/ahmetb/kubectl-tree) we can see our workload is parent to a runnable which in
turn is parent to a pipeline-run.

```shell
$ kubectl tree workload hello-again
```

```console
NAMESPACE  NAME                                                                    READY  REASON        AGE
default    Workload/hello-again                                                    True   Ready
default    ├─Deployment/hello-again-deployment                                     -
default    │ └─ReplicaSet/hello-again-deployment-67b7dc6d5                         -
default    │   ├─Pod/hello-again-deployment-67b7dc6d5-djtph                        True
default    │   ├─Pod/hello-again-deployment-67b7dc6d5-f2nkv                        True
default    │   └─Pod/hello-again-deployment-67b7dc6d5-p4m9c                        True
default    ├─Image/hello-again                                                     True
default    │ ├─Build/hello-again-build-1                                           -
default    │ │ └─Pod/hello-again-build-1-build-pod                                 False  PodCompleted
default    │ ├─PersistentVolumeClaim/hello-again-cache                             -
default    │ └─SourceResolver/hello-again-source                                   True
default    └─Runnable/hello-again-linter                                           True   Ready
default      └─PipelineRun/hello-again-linter-pipeline-run-x123x                   -
default        ├─PersistentVolumeClaim/pvc-1a89a7e201                              -
default        ├─TaskRun/hello-again-linter-pipeline-run-x123x-fetch-repository    -
default        │ └─Pod/hello-again-linter-pipeline-run-x123x-fetch-repository-pod  False  PodCompleted
default        └─TaskRun/hello-again-linter-pipeline-run-x123x-md-lint-run         -
default          └─Pod/hello-again-linter-pipeline-run-x123x-md-lint-run-pod       False  PodCompleted
```

We also see that the workload is in a ready state, as are all of the pods of our deployment.

## Wrap Up

You’ve now built a supply chain leveraging runnables. Your app platform can now provide testing, scanning, linting and
more to all the applications brought by your dev teams. Let’s look at what you learned in this tutorial:

- How to wrap a Runnable in a Cartographer template
- How to align the outputs from a ClusterRunTemplate with the values exposed by the template wrapping the Runnable

Before we leave this tutorial, we should note that the supply chain that we’ve created deals very well with new apps
that are brought to the platform. That is, when a workload is submitted, the app will be linted and upon success will
proceed to be built. But this supply chain is not resilient to changes made to the source code of said app. What will
happen if the code was good, but is changed to a bad state? The linting step won’t rerun, as from the tekton
perspective, no value has changed; it still has the same url and revision.

In order to address this problem, production supply chains should leverage a resource like
[fluxCD’s source controller](https://fluxcd.io/docs/components/source/). This kubernetes resource will translate a
revision like `main` into the revision of the current commit on main (e.g. commit `abc123`). When main is updated,
source controller will ensure that it outputs the new commit that is the head of main. Leveraging this resource, we can
avoid the dilemma presented above.

Users can see supply chains that use fluxCD’s source controller in
[Cartographer’s examples](https://github.com/vmware-tanzu/cartographer/tree/main/examples).
