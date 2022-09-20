# Source to Knative Service with Testing

**before you proceed**: the example in this directory illustrates the use of
the latest components and functionality of Cartographer (including some that
may not have been included in the latest release yet). Make sure to check out
the version of this document in a tag that matches the latest version (for
instance, https://github.com/vmware-tanzu/cartographer/tree/v0.0.7/examples).

---

The [basic-sc] example illustrates how an App Operator group could set up a software
supply chain such that source code gets continuously built using the best
practices from [buildpacks] via [kpack/Image] and deployed to the cluster using
[knative-serving]. And the [runnable-tekton] example demonstrated how to make tekton tasks
updatable. This example will bring these two together to insert a testing step into a supply chain.

```

  source --> tests --> image --> knative service

```

As with [basic-sc], the directories here are structured in a way to reflect which Kubernetes
objects would be set by the different personas in the system:


```
  '── shared | testing-sc
      ├── 00-cluster                         preconfigured cluster-wide objects
      │                                        to configured systems other than
      │                                              cartographer (like, kpack)
      │
      │
      ├── app-operator                      cartographer-specific configuration
      │   ├── supply-chain-templates.yaml            that an app-operator would
      │   ├── ...                                                        submit
      │   └── supply-chain.yaml
      │
      │
      └── developer                         cartographer-specific configuration
          ├── ...                                       that an app-dev submits
          └── workload.yaml
```

## Prerequisites

1. Install the prerequisites of the [basic-sc] example.

2. Install [Tekton](https://tekton.dev/docs/getting-started/#installation)

### Resource Requirements
Read [here](../README.md#resource-requirements)

## Running the example in this directory

### Location of files

As with [basic-sc], this example uses two directories with sub-directories of Kubernetes resources:
[../shared](../shared) and [testing-sc (.)](.).

The shared directory has a subdirectory for cluster-wide configuration: [../shared/cluster](../shared/cluster)

There is a subdirectory of cartographer-specific files that an App Operator would submit
in both this and in the shared directory:
[./app-operator](./app-operator) and [../shared/app-operator](../shared/app-operator)

Finally, there is a subdirectory of cartographer-specific files that an
App Developer would submit in both this and in the shared directory:
[./developer](./developer) and [../shared/developer](../shared/developer)

### Configuring the example

Follow the [example configuration steps in basic-sc](../basic-sc/README.md#configuring-the-example).

### Deploying the files

Similar to the [deploy instructions in basic-sc](../basic-sc/README.md#deploying-the-files):

```bash
kapp deploy --yes -a example -f <(ytt --ignore-unknown-comments -f .) -f <(ytt --ignore-unknown-comments -f ../shared/ -f ./values.yaml)
```

### Observing the example

The [observation steps in basic-sc still apply](../basic-sc/README.md#observing-the-example).
When using tree, we see an additional child of the workload, a Runnable. As seen
in [runnable-tekton], the Runnable will have new TaskRun children as the supply
chain propagates updates.

```console
$ kubectl tree workload dev
NAMESPACE  NAME
default    Workload/dev
default    ├─GitRepository/dev                    source fetching
default    ├─Runnable/dev                            test running
default    │  └─TaskRun/dev-gdss4
default    │    └─Pod/dev-gdss4-pod-xvrj5
default    ├─Image/dev                             image building
default    │ ├─Build/dev-build-1-5dxn9
default    │ │ └─Pod/dev-build-1-5dxn9-build-pod
default    │ ├─PersistentVolumeClaim/dev-cache
default    │ └─SourceResolver/dev-source
default    └─App/dev                               app deployment
```

## Tearing down the example and uninstalling the dependencies

1. Uninstall Tekton: replace `apply` with `delete` in the [installation instructions](https://tekton.dev/docs/getting-started/#installation)
2. Follow the [basic-sc teardown instructions](../basic-sc/README.md#tearing-down-the-example).

## Explanation

This example creates a software supply chain like mentioned above

```

  source --> test --> image --> knative service


```

In [basic-sc] the details of `source`, `image` and `knative service`
are described. In [runnable-tekton] we are introduced to Runnables and how they
can wrap Tekton objects. Now we will wrap a Runnable in a `ClusterSourceTemplate`
and add it to the supply chain.

### Need: building the right container images out of source code

As mentioned in [basic-sc] `kpack` takes some form of source
code, builds a container image from it, and pushes that to a
registry. The image then be made available for deployments further down the
line.

An ability that was elided was to specify a given branch of a repository as in
this [kpack/Image] definition:

```yaml
apiVersion: kpack.io/v1alpha2
kind: Image
metadata:
  name: hello-world
spec:
  tag: projectcartographer/hello-world
  serviceAccount: cartographer-example-registry-creds-sa
  builder:
    kind: ClusterBuilder
    name: go-builder
  source:
    git:
      url: https://github.com/carto-run/hello-world
      revision: main
```

And that works great: push a commit to that repository, and `kpack` will build
an image, exposing under the object's `.status.latestImage` the absolute
reference to the image that has been built and pushed to a registry.

While that's great, we might want to actually gate the set of commits for which
images are built (for instance, only build those that passed
tests). And at the same time, we want to retain the behavior of kpack to update
the most recent image if there is an update to the base image.

i.e., it'd be great if we could somehow express:


```yaml
apiVersion: kpack.io/v1alpha2
kind: Image
metadata:
  name: hello-world
spec:
  source:
    git:
      url: https://github.com/carto-run/hello-world
      revision: $(commit_that_passed_tests)$
  ...
```

In order to do that, we'll need a way for the supply chain to pass forward just those
commits that pass tests.

### Testing an output

In [runnable-tekton], we saw an example of an object whose status is always the most
recently submitted commit that passed tests. We can now wrap this object in a
`ClusterSourceTemplate`. This is the Cartographer template object meant to expose
`url` and `revision` information to the supply chain.

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSourceTemplate
metadata:
  name: test
spec:
  urlPath: .status.outputs.url
  revisionPath: .status.outputs.revision

  template:
    apiVersion: carto.run/v1alpha1
    kind: Runnable
    metadata:
      name: $(workload.metadata.name)$
    spec:
      serviceAccountName: $(workload.spec.serviceAccountName)$

      runTemplateRef:
        name: tekton-taskrun

      inputs:
        params:
          - name: blob-url
            value: $(source.url)$
          - name: blob-revision
            value: $(source.revision)$
```

Next we update the supply chain. We create a new step in the supply-chain, `source-tester` that
references the new ClusterSourceTemplate. The supply chain passes this resource the values exposed
by the `source` step (the `url` and `revision` from the GitRepository object). And the supply chain
will read the `url` and `revision` values exposed by this new step, the revisions that pass tests.

```yaml
kind: ClusterSupplyChain
spec:
  selector:
    app.tanzu.vmware.com/workload-type: web

  resources:
    #
    - name: source-provider
      # same as before
      #
      templateRef:
        name: source
        kind: ClusterSourceTemplate

    - name: source-tester
      # declare that for this supply chain, a source-tester resource is
      # defined by the `ClusterSourceTemplate/test` object, making the source
      # information it exposes available to further resources in the chain.
      #
      templateRef:
        kind: ClusterSourceTemplate
        name: test
      # express that `source-tester` requires source (`{url, revision}`)
      # information from the `source-provider` resource, effectively making
      # that available to the template via `$(sources[0].)$` interpolation.
      #
      sources:
        - resource: source-provider
          name: source

    - name: image-builder
      # similar to before
      #
      templateRef:
        name: image
        kind: ClusterImageTemplate
      # NOTE: the sources now reference source-tester
      #
      sources:
        - resource: source-tester
          name: source
```

In this way, the supply chain ensures that every image built is from source code that passes tests.

[buildpacks]: https://buildpacks.io/
[knative-serving]: https://knative.dev/docs/serving/
[kpack/Image]: https://github.com/pivotal/kpack/blob/main/docs/image.md
[Tekton]: https://github.com/tektoncd/pipeline
[basic-sc]: ../basic-sc/README.md
[runnable-tekton]: ../runnable-tekton/README.md
