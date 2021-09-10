# Source to Knative Service

This example illustrates how an App Operator group could set up a software
supply chain such that source code gets continuously built using the best
practices from [buildpacks] via [kpack/Image] and deployed to the cluster using
[knative-serving].


```

  source --> image --> knative service

```

The directories here are structured in a way to reflect which Kubernetes
objects would be set by the different personas in the system:


```
  '── source-to-knative-service
      ├── 00-cluster                         preconfigured cluster-wide objects
      │                                        to configured systems other than
      │                                              cartographer (like, kpack)
      │
      │       
      ├── app-operator                      cartographer-specific configuration
      │   ├── supply-chain-templates.yaml            that an app-operator would
      │   └── supply-chain.yaml                                          submit
      │       
      │       
      └── developer                         cartographer-specific configuration
          └── workload.yaml                             that an app-dev submits
```


## Prerequisites

1. Kubernetes v1.19+

```bash
kind create cluster --image kindest/node:v1.21.1
```

2. [carvel] tools for templating and groups of kubernetes objects to the api
   server

  - [ytt]: templating the credentials
  - [kapp]: submitting objects to kubernetes


3. Cartographer, and dependencies used in the example

To install `cartographer`, refer to [README.md](../../README.md).

All that `cartographer` does is choreograph the passing of results from a
kubernetes object to another, following the graph described in the
[ClusterSupplyChain] object.

This means that `cartographer` by itself is not very useful - its powers arise
from integrating other kubernetes resources that when tied together with a
supplychain, forms something powerful.

In this example, we leverage the following dependencies:

- [kpack], for providing an opinionated way of continuously building container
  images using buildpacks

```bash
kapp deploy --yes -a kpack \
	-f https://github.com/pivotal/kpack/releases/download/v0.3.1/release-0.3.1.yaml
```

- [source-controller], for providing the ability to find new commits to a git
  repository and making it internally available to other components

```bash
kubectl create namespace gitops-toolkit

# THIS CLUSTERROLEBINDING IS FOR DEMO PURPOSES ONLY - THIS WILL GRANT MORE PERMISSIONS THAN NECESSARY
#
kubectl create clusterrolebinding gitops-toolkit-admin \
  --clusterrole=cluster-admin \
  --serviceaccount=gitops-toolkit:default

kapp deploy --yes -a gitops-toolkit \
  --into-ns gitops-toolkit \
  -f https://github.com/fluxcd/source-controller/releases/download/v0.15.4/source-controller.crds.yaml \
  -f https://github.com/fluxcd/source-controller/releases/download/v0.15.4/source-controller.deployment.yaml
```

- [kapp-controller], for providing us with the ability of deploying multiple
  Kubernetes objects as a single unit

```bash
# THIS CLUSTERROLEBINDING IS FOR DEMO PURPOSES ONLY - THIS WILL GRANT MORE PERMISSIONS THAN NECESSARY
#
kubectl create clusterrolebinding default-admin \
  --clusterrole=cluster-admin \
  --serviceaccount=default:default

kapp deploy --yes -a kapp-controller \
	-f https://github.com/vmware-tanzu/carvel-kapp-controller/releases/download/v0.22.0/release.yml
```

- [knative-serving], for being the runtime of the application we want to deploy.

```bash
kapp deploy --yes -a knative-serving \
  -f https://github.com/knative/serving/releases/download/v0.25.0/serving-crds.yaml \
  -f https://github.com/knative/serving/releases/download/v0.25.0/serving-core.yaml
```

4. Authorization to push images to a container image registry

As mentioned before, in this example we make use of `kpack` to build container
images out of the source code and push to a container image registry,
so we must provide a Kubernetes secret that contains the credentials for pushing to a
registry (see [kpack secrets] to know more about how kpack makes use of it).


## Running the example in this directory

As mentioned before, there are three directories: [./00-cluster](./00-cluster),
for cluster-wide configuration, [./app-operator](./app-operator), with
cartographer-specific files that an App Operator would submit, and
[./developer](./developer), containing Kubernetes objects that a developer
would submit (yes, just a [Workload]!)

Before we get started with the details pertaining to Cartographer, first we
need to set up a few details related to where container images should be pushed
to - because [kpack/Image] is going to be used for building container images
based on the source code, we need to tell it where those images should go, and
which credentials it should use.

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

That done, we can make use of `ytt` to interpolate those values into the
templates, and then pass the result to [kapp] for submitting to Kubernetes
(Ensure you are running this command from `examples/source-to-knative-service`):

```bash
# submit to Kubernetes all of the Kubernetes objects defined under this #
# directory (recursively) making them all owned by a single "kapp App" named
# "example".
#
ytt --ignore-unknown-comments -f . | kapp deploy --yes -a example -f-
```

As Cartographer choreographs the resources that will drive the application from
source to a running App, you'll be able to see that the cartographer/Workload
incrementally owns more objects. For instance, we can see that using the plugin
[kubectl tree](https://github.com/ahmetb/kubectl-tree):

```console
$ kubectl tree workload dev
NAMESPACE  NAME                                   READY  REASON               AGE
default    Workload/dev                           True   Ready                6m4s
default    ├─GitRepository/dev                    True   GitOperationSucceed  6m4s
default    ├─Image/dev                            True                        6m2s
default    │ ├─Build/dev-build-1-5dxn9            -                           5m54s
default    │ │ └─Pod/dev-build-1-5dxn9-build-pod  False  PodCompleted         5m53s
default    │ ├─PersistentVolumeClaim/dev-cache    -                           6m2s
default    │ └─SourceResolver/dev-source          True                        6m2s
default    └─App/dev                              -                           52s
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

To uninstall the dependencies mentioned in the prerequisites section, delete
the corresponding `kapp` apps:

```console
$ kapp list
Apps in namespace 'default'

Name             Namespaces                 Lcs   Lca
gitops-toolkit   (cluster),gitops-toolkit   true  3m
knative-serving  (cluster),knative-serving  true  2m
kpack            (cluster),kpack            true  3m
```

```bash
kapp delete -a gitops-toolkit
kapp delete -a knative-serving
kapp delete -a kpack
```


## Step by step

With the goal of creating a software supply chain like mentioned above

```

  source --> image --> knative service
                      

```

let's go step by step adding the components to form that supply chain.


### Building container images out of source code

Tools like `kpack` are great at doing the job of taking some form of source
code, and then based on that, building a container image and pushing that to a
registry, which can then be made available for deployments further down the
line.

For instance, one can express the desire of continuously getting container
images built based on a branch with the following [kpack/Image] definition:


```yaml
apiVersion: kpack.io/v1alpha1
kind: Image
metadata:
  name: hello-world
spec:
  tag: projectcartographer/hello-world
  serviceAccount: service-account
  builder:
    kind: ClusterBuilder
    name: go-builder
  source:
    git:
      url: https://github.com/kontinue/hello-world
      revision: main
```

And that works great: push a commit to that repository, and `kpack` will build
an image, exposing under the object's `.status.latestImage` the absolute
reference to the image that has been built and pushed to a registry.


```yaml
apiVersion: kpack.io/v1alpha1
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


Even better, if there's an update to the base image that should be used to
either build (for instance, `golang` bumped from 1.15 to 1.15.1 to address a
security vulnerability) or run our application (let's say, a CA certificate
should not be trusted anymore), even if there have been no changes to our
source code, `kpack` would take care of building a fresh new image for us.

While that's all great, we might want to actually gate the set of commits that
should have images built for (for instance, only build those that passed
tests), similar to what you'd want to do in a traditional pipeline, but not
block those base image updates to occur.

`kpack` does allow you to be very specific in terms of what to build, for
instance, by specifying which revision to use:


```yaml
apiVersion: kpack.io/v1alpha1
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

but then in that case, we'd need a way of updating that exact field
(`spec.source.git.revision`) with the commits we want (i.e., those that passed
our tests).

i.e., it'd be great if we could somehow express:


```yaml
apiVersion: kpack.io/v1alpha1
kind: Image
metadata:
  name: hello-world
spec:
  source:
    git: 
      url: https://github.com/kontinue/hello-world
      revision: $(commit_that_passed_tests)$
  ...
```

Not only that, it would be great if we could also make this reusable so that
any developer wanting to have their code built, could _"just"_ get it done
without having to know the details of `kpack`, something like

```yaml
apiVersion: kpack.io/v1alpha1
kind: Image
metadata:
  name: $(name_of_the_project)
spec:
  source:
    git: 
      url: $(developer's_repository)
      revision: $(commit_that_passed_tests)$
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
  ref: {branch: $(branch_developers_want_to_deploy_from)}
```


### Passing the commits discovered from GitRepository to Image

So at this point, we have two Kubernetes components that could very well work
together:

- `fluxcd/GitRepository`, providing that source information to other components
- `kpack/Image`, consuming source information, and then making `image`
  information available to further components


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
apiVersion: kpack.io/v1alpha1                         |     |    information to
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
# further components in the supply chain.
#
#
apiVersion: carto.run/v1alpha1
kind: ClusterSourceTemplate
metadata:
  name: source
spec:

  # because we're implementing a `Cluster___Source___Template`, we must specify
  # how to grab information about the source code that should be promoted to
  # further components.
  #
  # `*Path` fields expect a `jsonpath` query expression to run over the object
  # that has been templatized and submitted to kubernetes.
  #

  urlPath: .status.artifact.url       # in the object, where to pick up the url
  revisionPath: .status.artifact.revision       # where to pick up the revision

  template:                                      # what to stamp out and submit
    apiVersion: source.toolkit.fluxcd.io/v1beta1 # to kubernetes.
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
can expose image information to other components in the supply chain:

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
    apiVersion: kpack.io/v1alpha1
    kind: Image
    metadata:
      name: $(workload.name)$
    spec:
      tag: projectcartographer/$(workload.name)$
      serviceAccount: service-account
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

This definition of the link between the components (and developer Workload
objects) is described by a ClusterSupplyChain:


```yaml
kind: ClusterSupplyChain
spec:
  # describe which Workloads this supply chain is applicable to
  #
  selector:
    app.tanzu.vmware.com/workload-type: web

  
  # declare the set of components that form the software supply chain that
  # we are building.
  #
  components:
    #
    - name: source-provider
      # declare that for this supply chain, a source-provider component is
      # defined by the `ClusterSourceTemplate/source` object, making the source
      # information it exposes available to further components in the chain.
      #
      templateRef:
        name: source
        kind: ClusterSourceTemplate

    - name: image-builder
      # declare that for this supply chain, an image-builder component is
      # defined by the `ClusterImageTemplate/image` object, making the image
      #information it exposes available to further components in the chain.
      #
      templateRef:
        name: image
        kind: ClusterImageTemplate
      # express that `image-builder` requires source (`{url, revision}`)
      # information from the `source-provider` component, effectively making
      # that available to the template via `$(sources[0].)$` interpolation.
      #
      sources:
        - component: source-provider
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
component to supply chain, one that would actually deploy that code that has
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
apiVersion: kpack.io/v1alpha1                         |     |    information to
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
- add a component in the supplychain that make use of such `ClusterTemplate`
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
  # components (unlike `ClusterSourceTemplate` which outputs `source`
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
have that extra component:

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
  components:
    - name: source-provider
      templateRef:
        kind: ClusterSourceTemplate
        name: source

    - name: image-builder
      templateRef:
        kind: ClusterImageTemplate
        name: image
      sources:
        - component: source-provider
          name: source

    - name: deployer
      templateRef:
        kind: ClusterTemplate
        name: app-deploy
      images:
        - component: image-builder
          name: image
```


### What next?

The intention of this step-by-step guide was to demonstrate how the base shape
of the supplychain in this repository is set up.

As you can tell, Cartographer is not necessarily tied to any of the resources
utilized: if you want to switch [kpack] by any other image builder that does so
with a declarative interface, go for it!

Make sure to check out the [reference] to know more about Cartographer's custom
resources, and the YAML in this directory for an example that leverages the
same ideas here, but adds [knative] and [kapp-controller] to the mix.

[ClusterBuilder]: https://github.com/pivotal/kpack/blob/main/docs/builders.md#cluster-builders
[ClusterSupplyChain]: ../../site/content/docs/reference.md#ClusterSupplyChain
[Workload]: ../../site/content/docs/reference.md#Workload
[buildpacks]: https://buildpacks.io/
[carvel]: https://carvel.dev/
[kapp-controller]: https://carvel.dev/kapp-controller/
[kapp]: https://carvel.dev/kapp
[knative-serving]: https://knative.dev/docs/serving/
[knative]: https://knative.dev/docs/
[kpack secrets]: https://github.com/pivotal/kpack/blob/main/docs/secrets.md
[kpack/Image]: https://github.com/pivotal/kpack/blob/main/docs/image.md
[kpack]: https://github.com/pivotal/kpack
[reference]: ../../site/content/docs/reference.md
[source-controller]: https://github.com/fluxcd/source-controller
[ytt]: https://carvel.dev/ytt
