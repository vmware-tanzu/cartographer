# Automatically mutating carto.run/Pipeline to achieve testing for every commit

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


3. Cartographer

See installation at [cartographer]'s README.


[cartographer]: https://github.com/vmware-tanzu/cartographer

### Running

0. Fork the repository and update the `carto.run/Workload` object
   ([devs/workload.yaml](./devs/workload.yaml)) to point at the fork

```diff
  apiVersion: carto.run/v1alpha1
  kind: Workload
  metadata:
    name: dev-1-master
    labels:
      app.tanzu.vmware.com/workload-type: web
  spec:
    source:
      git:
-       url: https://github.com/kontinue/hello-world
+       url: https://github.com/YOUR_USER/hello-world
        ref: {branch: main}
```

1. Submit the files to Kubernetes

```bash
kapp deploy -a example -f .
```
```console
```

2. Observe that we get a `tekton/PipelineRun` for the latest commit in the
   repository


```bash
kubectl get pipelinerun
```
```console
```


3. See a new `tekton/PipelineRun` for a new commit that we push

First, start by cloning the forked repository and then pushing a change:

```bash
pushd `mktemp -d`
git clone https://github.com/YOUR_USER/hello-world .
touch no-op.txt
git add --all && git commit -m "no-op" && git push origin HEAD
popd
```

That done, observe that in a few moments a new `tekton/PipelineRun` gets
automatically created with no need of patching anything else in Kubernetes.

```bash
kubectl get pipelinerun
```
```console
```
