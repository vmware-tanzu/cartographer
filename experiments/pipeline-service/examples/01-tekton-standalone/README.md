# Standalone Tekton Pipeline

Here you'll find an example that demonstrates how you get to run Tekton
pipelines without the help of [cartographer] and pipeline-service.

## Context

Pipeline runners like [tekton], [argo-workflows], or even [kubernetes jobs]
don't provide a declarative way of expressing the intention of having a given
pipeline run whenever a set of inputs change.  

For instance, with Tekton one can define a [tekton/Pipeline], which defines a
series of [tekton/Task]s that accomplish a specific build or delivery goal, but
to actually run it, a [tekton/PipelineRun] object (pointing at the
tekton/Pipeline) must be submitted to Kubernetes so that Tekton actually runs
it.

i.e.:

1. define tekton/Tasks
2. define tekton/Pipeline that points at those tasks
3. define a tekton/PipelineRun with the parameters specified each time you want
   to actually run it

Because tekton itself doesn't provide a way of automatically running those
pipelines, [tekton-triggers] is usually used as a way of creating an endpoint
that source control management systems can make requests to so that on a new
commit, a new tekton/PipelineRun gets automatically created based on a
template.

i.e.:
1. define tekton/Tasks
2. define tekton/Pipeline
3. define triggers/EventListener, for creating the endpoint to listen for webhook requests
4. define triggers/TriggerBinding, for mapping webhook content to parameters
4. define triggers/TriggerTemplate, for mapping parameters to tekton/PipelineRun

_ps.: a very similar approach is taken by argo-workflows: [argo/Workflow] objects
represent the "running" of an [argo/WorkflowTemplate], but as you need somehow to
submit a Workflow anytime you want to run it, [argo-events] was created to let
you template out Workflow objects whenever there's a webhook request is made._

## Example

### Prerequisites

1. Kubernetes cluster, v1.19+

2. Tekton Pipelines

```bash
kapp deploy --yes -a tekton -f https://storage.googleapis.com/tekton-releases/pipeline/previous/v0.27.2/release.yaml
```
```console
Target cluster 'https://127.0.0.1:39233' (nodes: kind-control-plane)

Changes

Namespace         Name                                        Kind                            Conds.  Age  Op      Op st.  Wait to    Rs  Ri
(cluster)         clustertasks.tekton.dev                     CustomResourceDefinition        -       -    create  -       reconcile  -   -
^                 conditions.tekton.dev                       CustomResourceDefinition        -       -    create  -       reconcile  -   -
^                 config.webhook.pipeline.tekton.dev          ValidatingWebhookConfiguration  -       -    create  -       reconcile  -   -
^                 pipelineresources.tekton.dev                CustomResourceDefinition        -       -    create  -       reconcile  -   -
...

7:57:19AM: ok: reconcile deployment/tekton-pipelines-controller (apps/v1) namespace: tekton-pipelines
7:57:19AM: ok: reconcile deployment/tekton-pipelines-webhook (apps/v1) namespace: tekton-pipelines
7:57:19AM: ---- applying complete [47/47 done] ----
7:57:19AM: ---- waiting complete [47/47 done] ----

Succeeded
```

ps.: refer to [tekton] for more details.


### Running

In this example, we have two files:

- [00-tekton-pipeline.yaml](./00-tekton-pipeline.yaml) contains the definition
  of the tekton/Pipeline that declares the set of tasks we should run for a
  given set of parameters (as mentioned before, it's just a "blueprint", it
  doesn't run by itself)

- [01-tekton-pipeline-run.yaml](./01-tekton-pipeline-runs.yaml) contains two
  tekton/PipelineRun objects that represents the instantiation of runs - one
  for a particular commit, and another for another. For each commit that we'd
  like to run the pipeline against, we'd need to create one of these.

- [02-tekton-pipeline-run.yaml](./02-tekton-pipeline-runs.yaml) same as the
  other PipelineRun, but with a different set of parameters used in the
  invocation.

First, submit the `tekton/Pipeline` object from `00-tekton-pipeline.yaml`, the
definition of a blueprint of tasks that should be performed given a set of
parameters.

```bash
kubectl apply -f ./00-tekton-pipeline.yaml
```
```console
pipeline.tekton.dev/standalone created
```

With the definition there, now submit the first pipeline run invocation:

```bash
kubectl apply -f ./01-tekton-pipeline-run.yaml
```
```console
pipelinerun.tekton.dev/standalone-run-1 created
```

And then, using [tkn], the Tekton CLI, observe its result:

```bash
tkn pipelinerun describe standalone-run-1
```
```console
Name:              standalone-run-1
Namespace:         default
Pipeline Ref:      standalone
Service Account:   default
Timeout:           1h0m0s
Labels:
 tekton.dev/pipeline=standalone

🌡️  Status

STARTED         DURATION     STATUS
4 minutes ago   17 seconds   Succeeded


⚓ Params

 NAME                VALUE
 ∙ source-url        https://github.com/kontinue/hello-world
 ∙ source-revision   19769456b6b229b3e78f2b90eced15a353eb4e7c


🗂  Taskruns

 NAME                            TASK NAME   STARTED         DURATION     STATUS
 ∙ standalone-run-1-test-f62b4   test        4 minutes ago   17 seconds   Succeeded
```

But then now, if we'd like to actually run against the latest commit of our
repository, we can't just expect that the tekton/Pipeline will magically
discover that - we have to create a new pipeline run invocation object with the
parameter set with the revision that we want to run tests against:

```bash
kubectl apply -f ./02-tekton-pipeline-run.yaml
```
```console
pipelinerun.tekton.dev/standalone-run-2 created
```

---

As we'll see in [next example], we can instead of manually submitting those
tekton/PipelineRun objects, have that templated in a way that pipeline-service
can take care of ensuring that they get submitted when we have new revisions.


[argo-events]: https://github.com/argoproj/argo-events
[argo-workflows]: https://github.com/argoproj/argo-workflows
[argo/WorkflowTemplate]: https://github.com/argoproj/argo-workflows/blob/master/docs/workflow-templates.md
[argo/Workflow]: https://github.com/argoproj/argo-workflows/blob/master/docs/workflow-concepts.md
[cartographer]: https://github.com/vmware-tanzu/cartographer
[kubernetes jobs]: https://kubernetes.io/docs/concepts/workloads/controllers/job/
[tekton-triggers]: https://github.com/tektoncd/triggers
[tekton/PipelineRun]: https://github.com/tektoncd/pipeline/blob/main/docs/pipelineruns.md
[tekton/Pipeline]: https://github.com/tektoncd/pipeline/blob/main/docs/pipelines.md
[tekton/Task]: https://github.com/tektoncd/pipeline/blob/main/docs/tasks.md
[tekton]: https://github.com/tektoncd/pipeline
[tkn]: https://github.com/tektoncd/cli

[next example]: ../02-simple-pipeline-service
