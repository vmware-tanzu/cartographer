# Using carto.run/Pipeline to run Tekton Pipelines

`pipeline-service` extends Kubernetes to provide the ability to express (in a
declarative form) the desire to have a pipeline run and ensure that it does so
on any change to its specification. 

Like we mentioned in the [previous example], in order to get a
[tekton/Pipeline] to run, we must submit an "invocation object":
[tekton/PipelineRun].



```

      """""""""                         """"""""""
      blueprint                         invocation
      """""""""                         """"""""""


      tekton/Pipeline                   tekton/PipelineRun
        name: foo         <------.        name: invocation-1
                                 |
                                 '------  pipelineRef: {name: foo}
        params:                           params:
          - name: url                       - {name: url,      value: git.com/foo}
          - name: revision                  - {name: revision, value: f00b4r}

        tasks:
          - steps:
            - script: |-
                git clone $(params.url)$
                git checkout $(params.revision)$
```

To make the creation of such invocations declarative, `pipeline-service` has
the `carto.run/RunTemplate` CRD:


```yaml
kind: RunTemplate
apiVersion: carto.run/v1alpha1
metadata:
  name: template
spec:
  #
  #  definition of how the controller would be able to tell if the
  #  invocation finished successfully or not.
  #
  completion:
    #
    # let the controller know that to consider a PipelineRun succeeded, it must
    # find under the conditions array an object with the field `type` with 
    # value 'Succeeded' to evaluate to `True`.
    #
    #
    # for instance:
    #
    #     status:
    #       conditions:
    #         - type: Succeeded     ==>   this run succeeded!
    #           status: True
    # 
    # on the other hand, for detecting a failure:
    # 
    #     status:
    #       conditions:
    #         - type: Succeeded     ==>   this run failed!
    #           status: False
    #
    #
    succeeded: {key: 'status.conditions.#(type=="Succeeded").status', value: "True"}
    failed:    {key: 'status.conditions.#(type=="Succeeded").status', value: "False"}

  #
  #  what should be stamped out to invoke a pipeline.
  #
  template:
    apiVersion: tekton.dev/v1beta1
    kind: PipelineRun
    metadata:
      generateName: $(pipeline.metadata.name)$-
    spec:
      pipelineRef: {name: simple}
      params:
        - name: source-url
          value: $(pipeline.spec.inputs.url)$
        - name: source-revision
          value: $(pipeline.spec.inputs.revision)$
```

With the definition of how invocations should be created, the next step is
providing the object that controls the inputs that are supplied to this
template:


```yaml
kind: Pipeline
apiVersion: carto.run/v1alpha1
metadata:
  name: pipeline
spec:
  #
  #  name of the carto.run/RunTemplate object to pick up from this
  #  namespace to base invocations of
  #
  runTemplateName: simple

  #
  #   inputs that should be supplied to the pipeline invocations
  #
  inputs:
    url: https://github.com/kontinue/hello-world
    revision: 19769456b6b229b3e78f2b90eced15a353eb4e7c
```

This way, whenever an update to `pipeline.spec.inputs` occurs, the controller
takes care of instantiating a new invocation based on that
`carto.run/RunTemplate` and keeping track of whether it succeeded/failed.


## Running the example

1. submit to kubernetes the objects in this directory

```bash
kapp deploy -a example-simple -f .
```
```console
Target cluster 'https://127.0.0.1:39233' (nodes: kind-control-plane)

Changes

Namespace  Name    Kind         Conds.  Age  Op      Op st.  Wait to    Rs  Ri
default    simple  Pipeline     -       -    create  -       reconcile  -   -
^          simple  Pipeline     -       -    create  -       reconcile  -   -
^          simple  RunTemplate  -       -    create  -       reconcile  -   -

Op:      3 create, 0 delete, 0 update, 0 noop
Wait to: 3 reconcile, 0 delete, 0 noop

Continue? [yN]: y
...
8:17:27AM: ok: reconcile pipeline/simple (tekton.dev/v1beta1) namespace: default
8:17:27AM: ---- applying complete [3/3 done] ----
8:17:27AM: ---- waiting complete [3/3 done] ----

Succeeded
```


2. observe that we got a tekton/PipelineRun (execution of a Tekton pipeline)

```bash
kubectl tree pipelines.carto.run simple
```
```console
GVK                               NAME
. carto.run/Pipeline              simple
'---- tekton/PipelineRun          simple-abcd
       '---- tekton/TaskRun       simple-abcd-test-defg
             '---- Pod            simple-abcd-test-defg-pod-deas
```

3. point the carto.run/Pipeline at a different revision

```diff
  apiVersion: carto.run/v1alpha1
  kind: Pipeline
  metadata:
    name: pipeline
  spec:
    inputs:
      url: https://github.com/kontinue/hello-world
-     revision: 19769456b6b229b3e78f2b90eced15a353eb4e7c
+     revision: 3d42c19a618bb8fc13f72178b8b5e214a2f989c4
```

6. observe that a new tekton/PipelineRun was created

```diff
  . carto.run/Pipeline              simple
  +---- tekton/PipelineRun          simple-abcd
  |      '---- tekton/TaskRun       simple-abcd-test-defg
  |
+ '---- tekton/PipelineRun          simple-ncbd-test
+        '---- tekton/TaskRun       simple-ncbd-test-lksd
```


[tekton/PipelineRun]: https://github.com/tektoncd/pipeline/blob/main/docs/pipelineruns.md
[tekton/Pipeline]: https://github.com/tektoncd/pipeline/blob/main/docs/pipelines.md

[previous example]: ../01-tekton-standalone

