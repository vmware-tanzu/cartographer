# Source to Knative Service

**before you proceed**: the example in this directory illustrates the use of
the latest components and functionality of Cartographer (including some that
may not have been included in the latest release yet). Make sure to check out
the version of this document in a tag that matches the latest version (for
instance, https://github.com/vmware-tanzu/cartographer/tree/v0.0.7/examples).

---

This example illustrates how an App Operator group could set up a software
supply chain such that source code gets continuously built using the best
practices from [buildpacks] via [kpack/Image] and deployed to the cluster using
[knative-serving].


```

  source --> image --> knative service

```

The example directories are structured in a way to reflect which Kubernetes
objects would be set by the different personas in the system:


```
  '── shared | basic-sc
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

1. Kubernetes v1.19+

```bash
kind create cluster --image kindest/node:v1.21.1
```

2. [Carvel] tools for templating and groups of Kubernetes objects to the api
   server

  - [ytt]: templating the credentials
  - [kapp]: submitting objects to Kubernetes

3. Cartographer, and dependencies used in the example

To install `cartographer`, refer to [README.md](../../README.md).

All that `cartographer` does is choreograph the passing of results from a
Kubernetes object to another, following the graph described in the
[ClusterSupplyChain] object.

This means that `cartographer` by itself is not very useful - its powers arise
from integrating other Kubernetes resources that when tied together with a
supplychain, forms something powerful.

- [kpack](https://github.com/pivotal/kpack/blob/main/docs/install.md),
  for providing an opinionated way of continuously building container
  images using buildpacks

- [source-controller](https://fluxcd.io/docs/gitops-toolkit/source-watcher/#install-flux),
  for providing the ability to find new commits to a git
  repository and making it internally available to other resources

- [kapp-controller](https://carvel.dev/kapp-controller/docs/latest/install/),
  for providing us with the ability of deploying multiple
  Kubernetes objects as a single unit

- [knative-serving](https://knative.dev/docs/install/serving/install-serving-with-yaml/),
  for being the runtime of the application we want to deploy.

4. A container image registry in which you have authorization to push images. The
  credentials will be passed to kpack in the configuration steps below.

5. [Tree](https://github.com/ahmetb/kubectl-tree), a tool that we will use to observe the objects created.

## Running the example in this directory

### Location of files

This example uses two directories with sub-directories of Kubernetes resources:
[../shared](../shared) and [basic-sc (.)](.).

The shared directory has a subdirectory for cluster-wide configuration: [../shared/cluster](../shared/cluster)

There is a subdirectory of cartographer-specific files that an App Operator would submit
in both this and in the shared directory:
[./app-operator](./app-operator) and [../shared/app-operator](../shared/app-operator)

Finally, there is a subdirectory of cartographer-specific files that an
App Developer would submit in both this and in the shared directory:
[./developer](./developer) and [../shared/developer](../shared/developer)

### Configuring the example

Before we get started with the details pertaining to Cartographer, first we
need to set up a few details related to where container images should be pushed
to - because [kpack/Image] is going to be used for building container images
based on the source code, we need to tell it where those images should go, and
which credentials it should use. Create a secret with the authentication details
of the container registry you plan to use for the example (see [kpack documentation
of secrets](https://github.com/pivotal/kpack/blob/main/docs/secrets.md) for reference)

To make the example easy to set up, here we make use of [ytt] for templating
those Kubernetes objects out.

Update [`values.yaml`](./values.yaml) with the details of where you want the
images to go to:

```yaml
#@data/values
---
# prefix to use for any images that are going to be built and pushed.
#
# for example, for dockerhub, replace `projectcartographer` with your repository
# given a workload named 'foo', we'd perform the equivalent of a
# `docker push ${prefix}foo`, (e.g. `docker push projectcartographer/demo-foo`).
#
# if using a registry other than docker, the registry must be included in the prefix.
# Example: gcr.io/project-name/registry/demo-
image_prefix: projectcartographer/demo-


# configuration necessary for pushing the images to a container image registry.
#
registry:
  # for instance, if you're using `gcr`, this would be `gcr.io`
  #
  server: https://index.docker.io/v1/
  username: projectcartographer
  password: a-very-hard-to-break-password
```

Next the developer needs to give Cartographer permission to create the
items that will be stamped out in the supply chain. That will be done
with a service account, whose name will be referenced in the workload.
You can update [`values.yaml`](./values.yaml) with a desired name:

```yaml
#@data/values
---
# a name that won't collide with other service accounts
service_account_name: cartographer-example-basic-sc-sa
```

### Deploying the files

That done, we can make use of `ytt` to interpolate those values into the
templates, and then pass the result to [kapp] for submitting to Kubernetes
(Ensure you are running this command from `examples/basic-sc`):

```bash
# submit to Kubernetes all of the Kubernetes objects defined under this #
# directory (recursively) making them all owned by a single "kapp App" named
# "example".
#
kapp deploy --yes -a example -f <(ytt --ignore-unknown-comments -f .) -f <(ytt --ignore-unknown-comments -f ../shared/ -f ./values.yaml)
```

### Observing the example

As Cartographer choreographs the resources that will drive the application from
source to a running App, you'll be able to see that the cartographer/Workload
incrementally owns more objects. For instance, we can see that using the plugin
[kubectl tree](https://github.com/ahmetb/kubectl-tree):

```console
$ kubectl tree workload dev
NAMESPACE  NAME
default    Workload/dev
default    ├─GitRepository/dev                    source fetching
default    ├─Image/dev                             image building
default    │ ├─Build/dev-build-1-5dxn9
default    │ │ └─Pod/dev-build-1-5dxn9-build-pod
default    │ ├─PersistentVolumeClaim/dev-cache
default    │ └─SourceResolver/dev-source
default    └─App/dev                               app deployment
```

ps.: [octant](https://github.com/vmware-tanzu/octant) is another tool that
could very well be used to visualize the object tree.

Once `App/dev` is ready ("Reconciliation Succeeded")

```bash
kubectl get app
```
```console
NAME   DESCRIPTION           SINCE-DEPLOY   AGE
dev    Reconcile succeeded   19s            7m13s
```

we can see that the service has been deployed:

```bash
kubectl get services.serving.knative
```
```console
kubectl get services.serving
NAME   URL                              LATESTCREATED   LATESTREADY   READY     REASON
dev    http://dev.default.example.com   dev-00001       dev-00001     Unknown   IngressNotConfigured
```

Because we haven't installed and configured an ingress controller, we can't
just hit that URL, but we can still verify that we have our application up and
running by making use of port-forwarding, first by finding the deployment
corresponding to the current revion (`dev-00001`)

```bash
kubectl get deployment
```
```console
NAME                   READY   UP-TO-DATE   AVAILABLE   AGE
dev-00001-deployment   1/1     1            1           4m24s
```

and the doing the port-forwarding:

```bash
kubectl port-forward deployment/dev-00001-deployment 8080:8080
```
```console
Forwarding from 127.0.0.1:8080 -> 8080
Forwarding from [::1]:8080 -> 8080
```

That done, we can hit `8080` and get the result we expect by making a request
from another terminal:

```bash
curl localhost:8080
```
```console
hello world
```


## Tearing down the example

Having used `kapp` to deploy the example, you can get rid of it by deleting the
`kapp` app:

```bash
kapp delete -a example
```

### Uninstalling the dependencies

- [knative serving](https://knative.dev/docs/install/uninstall/)
- kpack: replace `apply` with `delete` in the [installation instructions](https://github.com/pivotal/kpack/blob/main/docs/install.md)
- [source-controller](https://fluxcd.io/docs/installation/#uninstall)
- kapp-controller: replace `apply` with `delete` in the [installation instructions](https://carvel.dev/kapp-controller/docs/latest/install/)

## Explanation

With the goal of creating a software supply chain like mentioned above

```

  source --> image --> knative service

```

let's go step by step adding the resources to form that supply chain.


### Building container images out of source code

Tools like `kpack` are great at doing the job of taking some form of source
code, and then based on that, building a container image and pushing that to a
registry, which can then be made available for deployments further down the
line.

`kpack` allows you to be very specific in terms of what to build, for
instance, by specifying which revision to use:


```yaml
apiVersion: kpack.io/v1alpha2
kind: Image
metadata:
  name: hello-world
spec:
  source:
    git:
      url: https://github.com/kontinue/hello-world
      revision: 3d42c19a618bb8fc13f72178b8b5e214a2f989c4
  ...
```


Even better, if there's an update to the base image that should be used to
either build (for instance, `golang` bumped from 1.15 to 1.15.1 to address a
security vulnerability) or run our application (let's say, a CA certificate
should not be trusted anymore), even if there have been no changes to our
source code, `kpack` would take care of building a fresh new image for us.


```yaml
apiVersion: kpack.io/v1alpha2
kind: Image
metadata:
  name: hello-world
spec:
  source:
    git: {url: https://github.com/kontinue/hello-world, revision: main}
    ...
status:
	latestImage: index.docker.io/projectcartographer/hello-world@sha256:27452d42b
```


but, we need a way of updating that exact field
(`spec.source.git.revision`) with the commits we want.

i.e., it'd be great if we could somehow express:

```yaml
apiVersion: kpack.io/v1alpha2
kind: Image
metadata:
  name: hello-world
spec:
  source:
    git:
      url: https://github.com/kontinue/hello-world
      revision: $(most_recent_commit)$
  ...
```

Not only that, it would be great if we could also make this reusable so that
any developer wanting to have their code built, could _"just"_ get it done
without having to know the details of `kpack`, something like


```yaml
apiVersion: kpack.io/v1alpha2
kind: Image
metadata:
  name: $(name_of_the_project)
spec:
  source:
    git:
      url: $(developer's_repository)
      revision: $(most_recent_commit)$
  ...
```

But before we do that, we'd first need to have a way of obtaining those commits
in the first place.


### Keeping track of commits to a repository

In the same declarative manner as `kpack`, there's a Kubernetes resource called
GitRepository that, once pointed at a git repository, it updates its status
with pointers to the exact version of the source code that has been discovered:

```yaml
apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: GitRepository
metadata:
  name: git-repository
spec:
  interval: 1m
  url: https://github.com/kontinue/hello-world
  ref: {branch: main}
```

which then, as it finds new revisions, updates its status letting us know where
to get that very specific revision:

```yaml
apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: GitRepository
metadata:
  name: git-repository
spec:
  interval: 1m
  url: https://github.com/kontinue/hello-world
  ref: {branch: main}
status:
  artifact:
    checksum: b2d2af59c64189efe008ba20fcbbe58b0f1532d5
    lastUpdateTime: "2021-08-11T18:20:26Z"
    path: gitrepository...2f989c4.tar.gz
    revision: main/3d42c19a61
    url: http://source-cont...89c4.tar.gz
```

But again, just like with `kpack/Image`, it would be great if we could
templatize that definition such that any developer team could make use of this,
something like:

```yaml
apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: GitRepository
metadata:
  name: $(name_of_the_project)
spec:
  interval: 1m
  url: $(developers_repository)$
  ref: {branch: $(branch_developers_want_to_deploy_from)$}
```


### Passing the commits discovered from GitRepository to Image

So at this point, we have two Kubernetes resources that could very well work
together:

- `fluxcd/GitRepository`, providing that source information to other resources
- `kpack/Image`, consuming source information, and then making `image`
  information available to further resources


```
kind: GitRepository
apiVersion: source.toolkit.fluxcd.io/v1beta1
spec:
  url: <git_url>
status:                                                       outputting source
  artifact:                                                      information to
    url: http://source-controller./b871db69.tar.gz ---.                  others
    revision: b871db69 -------------------------------|-----.
                                                      |     |
                                                      |     |
---                                                   |     |
kind: Image                                           |     |  outputting image
apiVersion: kpack.io/v1alpha2                         |     |    information to
spec:                                                 |     |            others
  source:                                             |     |
    git:                                              |     |
      url: <url_from_gitrepository_obj>        <------'     |
      revision: <revision_from_gitrepo_obj>    <------------'
status:
  conditions: [...]
  latestImage: gcr.io/foo/bar@sha256:b4df00d --------.
                                                     |
                                                    ...
                                                     |
                                                     ∨
                                          possibly .. a Deployment?
```

To make that possible, with Cartographer one would provide the definition of
how a GitRepository object should be managed by writing a ClusterSourceTemplate
(as it outputs "source" information)

```yaml
#
#
# `source` instantiates a GitRepository object, responsible for keeping track
# of commits made to a git repository, making them available as blobs to
# further resources in the supply chain.
#
#
apiVersion: carto.run/v1alpha1
kind: ClusterSourceTemplate
metadata:
  name: source
spec:

  # because we're implementing a `Cluster___Source___Template`, we must specify
  # how to grab information about the source code that should be promoted to
  # further resources.
  #
  # `*Path` fields expect a `jsonpath` query expression to run over the object
  # that has been templatized and submitted to Kubernetes.
  #

  urlPath: .status.artifact.url       # in the object, where to pick up the url
  revisionPath: .status.artifact.revision       # where to pick up the revision

  template:                                      # what to stamp out and submit
    apiVersion: source.toolkit.fluxcd.io/v1beta1 # to Kubernetes.
    kind: GitRepository
    metadata:
      name: $(workload.name)$            #     `$(workload.*)$` provides access
    spec:                                #        to fields from the `Workload`
      interval: 1m                       #              object submitted by the
      url: $(workload.source.git.url)$   #                           developers
      ref: $(workload.source.git.ref)$
      gitImplementation: libgit2
      ignore: ""
```

and with a `ClusterImageTemplate`, how a `kpack/Image` should be managed as it
can expose image information to other resources in the supply chain:

```yaml
#
#
# `image` instantiates a `kpack/Image` object, responsible for ensuring that
# there's a container image built and pushed to a container image registry
# whenever there's either new source code, or its image builder gets na update.
#
#
apiVersion: carto.run/v1alpha1
kind: ClusterImageTemplate
metadata:
  name: image
spec:
  imagePath: .status.latestImage    # where in the `kpack/Image` object to find
                                    #  the reference to the image that it built
                                    #  and pushed to a container image registry
  template:
    apiVersion: kpack.io/v1alpha2
    kind: Image
    metadata:
      name: $(workload.name)$
    spec:
      tag: projectcartographer/$(workload.name)$
      serviceAccount: cartographer-example-registry-creds-sa
      builder:
        kind: ClusterBuilder
        name: go-builder
      source:
        blob:
          url: $(sources[0].url)$       # source information from that a source
                                        #    provider like fluxcd/GitRepository
                                        #          from a ClusterSourceTemplate
                                        #                              provides
```

Having the templates written, the next step is to then define the link between
those two, that is, the dependency that `kpack/Image`, as described by a
ClusterImageTemplate, has on a source, `fluxcd/GitRepository`, as describe by a
ClusterSourceTemplate.

This definition of the link between the resources (and developer Workload
objects) is described by a ClusterSupplyChain:


```yaml
kind: ClusterSupplyChain
spec:
  # describe which Workloads this supply chain is applicable to
  #
  selector:
    app.tanzu.vmware.com/workload-type: web


  # declare the set of resources that form the software supply chain that
  # we are building.
  #
  resources:
    #
    - name: source-provider
      # declare that for this supply chain, a source-provider resource is
      # defined by the `ClusterSourceTemplate/source` object, making the source
      # information it exposes available to further resources in the chain.
      #
      templateRef:
        name: source
        kind: ClusterSourceTemplate

    - name: image-builder
      # declare that for this supply chain, an image-builder resource is
      # defined by the `ClusterImageTemplate/image` object, making the image
      #information it exposes available to further resources in the chain.
      #
      templateRef:
        name: image
        kind: ClusterImageTemplate
      # express that `image-builder` requires source (`{url, revision}`)
      # information from the `source-provider` resource, effectively making
      # that available to the template via `$(sources[0].)$` interpolation.
      #
      sources:
        - resource: source-provider
          name: source
```

Having gotten the definition of how each one of the objects
(fluxcd/GitRepository and kpack/Image) should be maintained, and the link
between them, all we need is for a developer to submit the intention of having
their repository built matching that supplychain:

```yaml
kind: Workload
metadata:
  labels:
    app.tanzu.vmware.com/workload-type: web
spec:
  source:
    git:
      url: https://github.com/kontinue/hello-world
      ref: {branch: main}
```

### Continuously deploying the image built by kpack

Having wrapped `kpack/Image` as a `ClusterImageTemplate`, we can add another
resources to supply chain, one that would actually deploy that code that has
been built.

In the simplest form, we could do that with a Kubernetes Deployment object,
something like:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
        - name: main
          image: $(my-container-image)$
```

Similarly to the example before where it'd be great if `kpack/Image` got source
code information from another Kubernetes object, here it'd be great if we could
continuously pass the image that `kpack` built for us to `Deployment`, forming
that full supply chain that brings source code all the way to a deployment.


```
kind: GitRepository
apiVersion: source.toolkit.fluxcd.io/v1beta1
spec:
  url: <git_url>
status:                                                       outputting source
  artifact:                                                      information to
    url: http://source-controller./b871db69.tar.gz ---.                  others
    revision: b871db69 -------------------------------|-----.
                                                      |     |
                                                      |     |
---                                                   |     |
kind: Image                                           |     |  outputting image
apiVersion: kpack.io/v1alpha2                         |     |    information to
spec:                                                 |     |            others
  source:                                             |     |
    git:                                              |     |
      url: <url_from_gitrepository_obj>        <------'     |
      revision: <revision_from_gitrepo_obj>    <------------'
status:
  conditions: [...]
  latestImage: gcr.io/foo/bar@sha256:b4df00d --------.
                                                     |
---                                                  |
apiVersion: apps/v1                                  |
kind: Deployment                                     |
metadata:                                            |
  name: my-app                                       |
spec:                                                |
  selector:                                          |
    matchLabels:                                     |
      app: my-app                                    |
  template:                                          |
    metadata:                                        |
      labels:                                        |
        app: my-app                                  |
    spec:                                            |
      containers:                                    |
        - name: main                                 |
          image: $(my-container-image)$ <------------'
```

In order to make this happen, we'd then engage in the very same activity is
before:

- wrap the definition of `apps/Deployment` as a `ClusterTemplate` object,
- add a resource in the supplychain that make use of such `ClusterTemplate`
  taking image information as input

```
#
#
# `app-deploy` instantiates a deployment making use of a built container image.
#
#
apiVersion: carto.run/v1alpha1
kind: ClusterTemplate
metadata:
  name: app-deploy
spec:
  # definition of the object to instantiate and keep patching when
  # new changes occur.
  #
  # note that we don't specify anything like `urlPath` or `imagePath` - that's
  # because `ClusterTemplate` objects don't output any information to further
  # resources (unlike `ClusterSourceTemplate` which outputs `source`
  # information).
  #
  template:
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: $(workload.name)$
    spec:
      selector:
        matchLabels:
          app: $(workload.name)$
      template:
        metadata:
          labels:
            app: $(workload.name)$
        spec:
          containers:
            - name: main
              image: $(images[0].image)$     # consume the image that we depend
                                             #            on from `kpack/Image`
```

Having the template created, all it takes then is to update the supplychain to
have that extra resource:

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSupplyChain
metadata:
  name: supply-chain
spec:
  selector:
    app.tanzu.vmware.com/workload-type: web

  #
  #
  #   source-provider <--[src]-- image-builder <--[img]--  deployer
  #     GitRepository               Image                     App
  #
  #
  resources:
    - name: source-provider
      templateRef:
        kind: ClusterSourceTemplate
        name: source

    - name: image-builder
      templateRef:
        kind: ClusterImageTemplate
        name: image
      sources:
        - resource: source-provider
          name: source

    - name: deployer
      templateRef:
        kind: ClusterTemplate
        name: app-deploy
      images:
        - resource: image-builder
          name: image
```


### What next?

#### Swap out resources of the Supply Chain

As you can tell, Cartographer is not necessarily tied to any of the resources
utilized: if you want to switch [kpack] by any other image builder that does so
with a declarative interface, go for it! Indeed, the YAML in this directory replace
the `deployer` resource in the supply chain described here with a resource that
leverages [knative] and [kapp-controller].

#### Expand on the Supply Chain

The intention of this step-by-step guide was to illustrate the basic shape
of the supplychain in this repository. Next, check out the
[testing-sc] example to see how testing can be added to the
supply-chain.

[ClusterBuilder]: https://github.com/pivotal/kpack/blob/main/docs/builders.md#cluster-builders
[ClusterSupplyChain]: ../../site/content/docs/reference.md#ClusterSupplyChain
[Pipeline]: https://github.com/tektoncd/pipeline/blob/main/docs/pipelines.md
[Workload]: ../../site/content/docs/reference.md#Workload
[buildpacks]: https://buildpacks.io/
[Carvel]: https://carvel.dev/
[kapp-controller]: https://carvel.dev/kapp-controller/
[kapp]: https://carvel.dev/kapp
[knative-serving]: https://knative.dev/docs/serving/
[knative]: https://knative.dev/docs/
[kpack/Image]: https://github.com/pivotal/kpack/blob/main/docs/image.md
[kpack]: https://github.com/pivotal/kpack
[reference]: ../../site/content/docs/reference.md
[source-controller]: https://github.com/fluxcd/source-controller
[ytt]: https://carvel.dev/ytt
[tekton]: https://github.com/tektoncd/pipeline
[testing-sc]: ../testing-sc/README.md
