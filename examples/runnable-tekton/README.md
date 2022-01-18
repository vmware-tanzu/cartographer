# Updatable Test Objects

**before you proceed**: the example in this directory illustrates the use of
the latest components and functionality of Cartographer (including some that
may not have been included in the latest release yet). Make sure to check out
the version of this document in a tag that matches the latest version (for
instance, https://github.com/vmware-tanzu/cartographer/tree/v0.0.7/examples).

---

The [basic-sc] example illustrates how an App Operator group could set up a software
supply chain such that source code gets continuously built using the best
practices from [buildpacks] via [kpack/Image] and deployed to the cluster using
[knative-serving]. The [testing-sc] example will add testing to the supply chain.
This example focuses on how the Cartographer CRD `Runnable` enables updating what are normally
immutable test resources.

## Prerequisites

1. Kubernetes v1.19+

```bash
kind create cluster --image kindest/node:v1.21.1
```

2. Install Cartographer. Refer to [README.md](../../README.md).

3. Install [Tekton](https://tekton.dev/docs/getting-started/#installation), which provides a
  mechanism to create pipelines and tasks for application testing, scanning, etc.

## Running the example in this directory

In order to demonstrate updatable testing, the example has a Tekton Task that will run `go test` on a
particular commit in a repo. Tekton does not allow updating an object, and so we'll update Runnable to
test new commits. Our Runnable is written to output the sha of passing tests, which we'll observe.

We start by submitting the setup objects: [./00-setup](./00-setup):

```bash
kubectl apply -f ./00-setup
```

Next we'll submit the runnable:

```bash
kubectl apply -f ./01-tests-pass/runnable.yml
```

Cartographer will use the Runnable to create a Tekton TaskRun. We can use the plugin
[kubectl tree](https://github.com/ahmetb/kubectl-tree) to see.

```bash
kubectl tree runnable test
```

```console
NAMESPACE  NAME                    READY  REASON        AGE
default    Runnable/test           True   Ready         2m39s
default    └─TaskRun/test-6w8lk    -                    2m37s
default      └─Pod/test-6w8lk-pod  False  PodCompleted  2m37s
```

The Runnable output reflects the most recent passing test.

```bash
kubectl get -o yaml runnable test
```

```yaml
apiVersion: carto.run/v1alpha1
kind: Runnable
metadata:
  name: test
spec: ...
status:
  conditions:
  - reason: Ready
    status: "True"
    type: RunTemplateReady
    ...
  - reason: Ready
    status: "True"
    type: Ready
    ...
  observedGeneration: 1
  outputs:
    revision: 19769456b6b229b3e78f2b90eced15a353eb4e7c
    url: https://github.com/kontinue/hello-world
```

Now let's update the Runnable with a different SHA, one where the tests fail:

```bash
kubectl patch runnable test --type merge --patch "$(cat 02-tests-fail/runnable-patch.yml)"
```

We can see that Runnable has a new child Tekton TaskRun:

```bash
kubectl tree runnable test
```

```console
NAMESPACE  NAME                    READY  REASON              AGE
default    Runnable/test           True   Ready               2m47s
default    ├─TaskRun/test-8rx94    -                          2m45s
default    │ └─Pod/test-8rx94-pod  False  PodCompleted        2m45s
default    └─TaskRun/test-zctzd    -                          37s
default      └─Pod/test-zctzd-pod  False  ContainersNotReady  36s
```

If we look at the logs of this new testing pod, the test has failed:

```bash
kubectl logs Pod/test-zctzd-pod
```

```console
...
+ go test -v ./...
=== RUN   TestDummy
--- FAIL: TestDummy (0.00s)
FAIL
FAIL    github.com/kontinue/hello-world 0.009s
FAIL
```

Runnable continues to output the sha of the passing test

```bash
kubectl get -o yaml runnable test
```

```yaml
apiVersion: carto.run/v1alpha1
kind: Runnable
metadata:
  name: test
spec: ...
status:
  outputs:
    revision: 19769456b6b229b3e78f2b90eced15a353eb4e7c  # <=== old sha
    url: https://github.com/kontinue/hello-world
```

Now let's update Runnable with a commit where tests again pass:

```bash
kubectl patch runnable test --type merge --patch "$(cat 03-tests-pass/runnable-patch.yml)"
```

We can see a new Tekton TaskRun created and completed

```bash
kubectl tree runnable test
```

```console
NAMESPACE  NAME                    READY  REASON              AGE
default    Runnable/test           True   Ready               6m40s
default    ├─TaskRun/test-8rx94    -                          6m38s
default    │ └─Pod/test-8rx94-pod  False  PodCompleted        6m38s
default    ├─TaskRun/test-sqhcj    -                          12s
default    │ └─Pod/test-sqhcj-pod  False  PodCompleted        12s
default    └─TaskRun/test-zctzd    -                          4m30s
default      └─Pod/test-zctzd-pod  False  ContainersNotReady  4m29s
```

And when we examine the Runnable, we can see that the output has changed
because of the Succeeded Tekton TaskRun:

```bash
kubectl get -o yaml runnable test
```

```yaml
apiVersion: carto.run/v1alpha1
kind: Runnable
metadata:
  name: test
spec: ...
status:
  outputs:
    revision: 3d42c19a618bb8fc13f72178b8b5e214a2f989c4  # <=== new sha
    url: https://github.com/kontinue/hello-world
  ...
```

While the Tekton TaskRuns are not updatable, runnable provides that behavior.

## Tearing down the example and the dependencies

```bash
kubectl delete runnable test
kubectl delete -f ./00-setup
```
Uninstall Tekton: replace `apply` with `delete` in the [installation instructions](https://tekton.dev/docs/getting-started/#installation)

## Explanation

We want to test our source code in Kubernetes. We're interested in doing so with objects that
are easy to update. But projects like [Tekton] create immutable runs for each test. While Tekton
users often leverage event triggers, we want to use simple objects that we submit to the server.
Here we'll walk step-by-step through the process of usign an easy to update object for tests and
scans: `Runnable`.

To begin, our developer uses [Tekton] to run tests on their repository.

```yaml
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: test
  labels:
    apps.tanzu.vmware.com/task: test
spec:
  params:
    - name: blob-url
    - name: blob-revision
  steps:
    - name: test
      image: golang
      command:
        - bash
        - -cxe
        - |-
          set -o pipefail

          git checkout $(params.blob-revision)

          cd `mktemp -d`
          git clone $(params.blob-url) && cd "`basename $(params.blob-url) .git`"
          go test -v ./...
```

A Tekton Task is a template to be run. This template is instantiated with
Tekton TaskRuns. A TaskRun object provides in its spec the necessary values to the
task template. When Tekton completes reconciliation of the TaskRun, a set of conditions
are included in the `.status.conditions` field of the TaskRun.

```yaml
apiVersion: tekton.dev/v1beta1
kind: TaskRun
metadata:
  name: test
spec:
  taskRef:
    name: test
  params:
  - name: blob-url
    value: https://github.com/kontinue/hello-world
  - name: blob-revision
    value: 3d42c19a618bb8fc13f72178b8b5e214a2f989c4
```

When these two objects are submitted, the TaskRun will execute and produce the status:

```yaml
apiVersion: tekton.dev/v1beta1
kind: TaskRun
metadata:
  name: test
spec:
  params:
    - name: blob-url
      value: https://github.com/kontinue/hello-world
    - name: blob-revision
      value: 3d42c19a618bb8fc13f72178b8b5e214a2f989c4
  ...
status:
  conditions:
  - message: All Steps have completed executing
    reason: Succeeded
    status: "True"
    type: Succeeded
    ...
  podName: test-pod
  startTime: "2021-12-14T20:50:30Z"
  ...
```

So we have a Kubernetes object that we submit to the cluster, it is reconciled and we can read state from it.
This sounds just like what we were doing with kpack in [basic-sc]. Why not just wrap the Tekton
TaskRun in a `ClusterSourceTemplate` and insert that into the supply chain?

There are two problems:

1. We want to read values from the TaskRun conditionally. Only if Succeeded is True do we want to pass the
  url and revision values forward in the supply chain.
2. TaskRuns are immutable. We cannot update the TaskRun when the next commit SHA is created.

This is the use case for Cartographer's `Runnable`. `Runnable`s are objects that bring updatable behavior
to immutable Kubernetes objects. When the runnable updates, it submits a new immutable object. The status
of the Runnable reflects state from the most recently successfully reconciled immutable object.

### Runnable: From immutable to mutable

Similar to how Tekton Tasks are tasks paired with TaskRuns, Cartographer Runnables are `Runnable`s paired
with `ClusterRunTemplate`s. The `ClusterRunTemplate` has 2 responsibilities:

1. Define the template of the immutable object that will be created.
2. Define the fields that will be read from a successfully reconciled object and what key those values
  will be written to in the Runnable status.

The `Runnable` is then responsible for supplying the `ClusterRunTemplate` with values to fill templated fields.

#### Creating new immutable objects

For simplicity sake, we're going to pretend that there is a Kubernetes resource `SuccessJob`. This resource is just like
`Job` except for one thing:

1. When the job completes, the condition "Succeeded" is "True" (rather than `Job`'s usual "Completed" condition)

Just as with the `Job` resource, the `.template.spec` field is immutable. We'll look at how Runnable can provide
updatable experience.

First a `SuccessJob` is written into a `ClusterRunTemplate`.

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterRunTemplate
metadata:
  name: mutate-job
spec:
  outputs:
    a-great-output: status.startTime
  template:
    apiVersion: batch/v1
    kind: SuccessJob
    metadata:
      generateName: $(runnable.metadata.name)$-
    spec:
      template:
        metadata:
          generateName: $(runnable.metadata.name)$-
        spec:
          containers:
            - name: say-something
              image: busybox
              command:
                - $(runnable.spec.inputs.command)$
                - $(runnable.spec.inputs.arg)$
          restartPolicy: OnFailure
```

And a Runnable is created with input fields that match.

```yaml
kind: Runnable
metadata:
  name: some-runnable
spec:
  serviceAccountName: service-account-with-role-to-create-jobs

  runTemplateRef:
    name: mutate-job

  inputs:
    command: "echo"
    arg: "be the change you wish to see in the world"
```

When these are submitted, a SuccessJob is created:

```yaml
apiVersion: batch/v1
kind: SuccessJob
metadata:
  generateName: say-
  name: say-wv5nr
spec:
  template:
    metadata:
      generateName: say-
    spec:
      containers:
        - command:
            - echo
            - be the change you wish to see in the world
          image: busybox
          imagePullPolicy: Always
          name: say-something
      ...
status:
  completionTime: "2021-12-14T17:57:16Z"
  conditions:
    - lastProbeTime: "2021-12-14T17:57:16Z"
      lastTransitionTime: "2021-12-14T17:57:16Z"
      status: "True"
      type: Succeeded
  startTime: "2021-12-14T17:57:14Z"
  succeeded: 1
```

If the definition of the `Runnable` is updated with a new input value, a new SuccessJob
is created.

```yaml
kind: Runnable
metadata:
  name: some-runnable
spec:
  serviceAccountName: service-account-with-role-to-create-jobs

  runTemplateRef:
    name: mutate-job

  inputs:
    command: "exit"
    arg: "1"

---
apiVersion: batch/v1
kind: SuccessJob
metadata:
  generateName: say-
  name: say-xyz987 # <=== new object, new name
spec:
  template:
    metadata:
      generateName: say-
    spec:
      containers:
        - command:
            - exit # <=== updated field from the Runnable
            - 1    # <=== updated field from the Runnable
          image: busybox
          imagePullPolicy: Always
          name: say-something
    ...
status:
  active: 1
  failed: 1
  startTime: "2021-12-14T18:42:03Z"
```

#### Reading from immutable fields and exposing on the Runnable

Cartographer expects that `Runnable`s are created in order to do some work and
report some status. Having seen how Runnable allows these immutable objects to be created,
let us look at how their values are returned.

Runnable reflects fields from the most recent successfully reconciled immutable object. _**Runnable
assumes that a successfully reconciled object has a `.status.conditions` object
with the `name` "Succeeded" and the `status` "True".**_ This is why our example is the imaginary `SuccessJob`.

Let's consider the Runnable that was updated above. Two SuccessJobs were created:

```yaml
---
apiVersion: batch/v1
kind: SuccessJob
metadata:
  name: say-wv5nr # <=== First SuccessJob Submitted
spec:
  template:
    ...
status:
  completionTime: "2021-12-14T17:57:16Z"
  conditions:
    - lastProbeTime: "2021-12-14T17:57:16Z"
      lastTransitionTime: "2021-12-14T17:57:16Z"
      status: "True"
      type: Succeeded                # <=== If this field were "Completed", this object would be the same as `Job`
  startTime: "2021-12-14T17:57:14Z"
  succeeded: 1

---
apiVersion: batch/v1
kind: SuccessJob
metadata:
  name: say-xyz987 # <=== First SuccessJob Submitted
spec:
  template:
    ...
status:
  active: 1
  failed: 1
  startTime: "2021-12-14T18:42:03Z"
```

SuccessJob `say-wv5nr` is the most recently submitted successful object. As such, Runnable
will expose values from it. To determine what will be exposed, we reference the `ClusterRunTemplate` above.

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterRunTemplate
metadata:
  name: mutate-job
spec:
  outputs:
    a-great-output: status.startTime # <=== key-value
  template:
    ...
```

The `ClusterRunTemplate`'s outputs fields are a set of key-value pairs. The value is the name of the field on the
immutable object whose value will be exposed. Here we see that the value at the path `.status.startTime` will
be exposed by the `Runnable`. On SuccessJob `say-wv5nr` we see the value at that path is `"2021-12-14T17:57:14Z"`.
The key is the name of the output field that will be created in the `Runnable`'s outputs. We can see here
the new status of the `Runnable`:

```yaml
kind: Runnable
metadata:
  name: some-runnable
spec:
  ...
status:
  outputs:
    a-great-output: "2021-12-14T17:57:14Z"
```

#### Automatic deletion of older created objects

Over time, the objects created by runnables can accumulate and consume resources in the cluster. For this reason
Cartographer will only retain a limited number of runs. By default, this is 5 failed and 3 successful runs. If
necessary, this can be customized:

```yaml
kind: Runnable
metadata:
  name: some-runnable
spec:
  retentionPolicy:
    maxFailedRuns: 3
    maxSuccessfulRuns: 1
  ...
```

### Wrapping Tekton in Runnable

We can now put together Tekton and Runnables. We'll submit to the cluster

- A Tekton Task
- A Tekton TaskRun wrapped in a Cartographer ClusterRunTemplate
- A Cartographer Runnable

The [Tekton Task](./00-setup/tekton-task.yml) is unchanged. The [ClusterRunTemplate](./00-setup/cluster-run-template.yml)
wraps a Tekton TaskRun and replaces the hardcoded values in the params with template fields. It also changes
the TaskRun `name` to a `generateName`. And it specifies the url and revision of the Runnable to be the
outputs of successful run. The [Runnable](./01-tests-pass/runnable.yml) defines params to pass into the
ClusterRunTemplate.

[buildpacks]: https://buildpacks.io/
[knative-serving]: https://knative.dev/docs/serving/
[kpack/Image]: https://github.com/pivotal/kpack/blob/main/docs/image.md
[tekton]: https://github.com/tektoncd/pipeline
[basic-sc]: ../basic-sc/README.md
[testing-sc]: ../testing-sc/README.md
