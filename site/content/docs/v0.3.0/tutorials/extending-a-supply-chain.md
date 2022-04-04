# Extending a Supply Chain (_or_ Multiple Templates in a Supply Chain)

## Overview

So far our supply chains have been a bit anemic, single step affairs. In this tutorial we’ll explore two topics:

- Adding new templates to an existing supply chain
- Passing information from one object created by a template to the template about to create another object

## Environment setup

For this tutorial you will need a kubernetes cluster with Cartographer and
[kpack](https://buildpacks.io/docs/tools/kpack/) installed. You can find
[Cartographer's installation instructions here](https://github.com/vmware-tanzu/cartographer#installation) and
[kpack's installation instructions can be found here](https://github.com/pivotal/kpack/blob/main/docs/install.md).

You will also need an image registry for which you have read and write permission.

Alternatively, you may choose to use the
[./hack/setup.sh](https://github.com/vmware-tanzu/cartographer/blob/main/hack/setup.sh) script to install a kind cluster
with Cartographer, kpack and a local registry. _This script is meant for our end-to-end testing and while we rely on it
working in that role, no user guarantees are made about the script._

Command to run from the Cartographer directory:

```shell
$ ./hack/setup.sh cluster cartographer-latest example-dependencies
```

If you later wish to tear down this generated cluster, run

```shell
$ ./hack/setup.sh teardown
```

## Scenario

### App Operator Steps

#### Supply Chain

We’ll start by considering the supply chain from our [“Build Your First Supply Chain"](first-supply-chain.md) tutorial:

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSupplyChain
metadata:
  name: supply-chain
spec:
  selector:
    workload-type: pre-built

  resources:
    - name: deploy
      templateRef:
        kind: ClusterTemplate
        name: app-deploy
```

Let’s think through what we want to change here. We know that we’re going to create a new step, so we’ll need a new
resource. This step is going to build an image, so it will come before the existing step that creates a deployment of
the image. For readability we’ll list the new resource before the current `deploy` resource. We’ll give the resource a
reasonable name. So far our `.spec.resources` will look like this:

<!-- prettier-ignore-start -->
```yaml
  resources:
    - name: build-image
      templateRef:
        kind: ???
        Name: ???
    - name: deploy
      templateRef:
        kind: ClusterTemplate
        name: app-deploy
```
<!-- prettier-ignore-end -->

So far we’ve only seen the template kind ClusterTemplate. But now we want to template an object that will pass
information to further templates in the supply chain. Cartographer expects three different types of information to be
passed through a supply chain. There are subsequently three template types that expose information to later resources in
a supply chain:

- ClusterSourceTemplates expose location of source code
- ClusterImageTemplates expose location of images
- ClusterConfigTemplates expose yaml specification of k8s objects

(Documentation of these custom resources [can be found here](../reference/template/))

In our scenario, we know that we’re going to template some object that will take the location of source code from the
workload, build an image and then we’ll want to share the location of that image with the next object in the supply
chain. As we want to expose the location of an image to the supply chain, we’ll use the ClusterImageTemplate and we’ll
give it a reasonable name. Our supply chain .spec.resources now looks like this:

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
        name: app-deploy
```
<!-- prettier-ignore-end -->

There’s one more addition we must make to the resources. While the build-image step will make information available for
consumption, we need to explicitly indicate that the deploy step will consume that information. We’ll do so by adding an
images field to that step. We’ll refer to the resource providing an image and give that value a name by which the
app-deploy template can refer to that value:

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
        name: app-deploy
      images:
        - resource: build-image
          name: built-image
```
<!-- prettier-ignore-end -->

Our resources section is looking good. Before we move on to writing the templates, let’s take a moment to think about
our app platform. We previously had just one supply chain that worked for all of our devs that provided prebuilt images.
We’re in the process of adding a supply chain that accepts apps defined in source code. This new supply chain doesn't
also support the prebuilt images; the final deploy step of this supply chain has a dependency on the build-image step.
We need to make three changes:

1. Give the supply chain a new name.
2. Give the supply chain different selector(s).
3. Change the template reference for the deploy step.

The name change is straightforward:

```yaml
metadata:
  name: source-code-supply-chain
```

The selector change in similarly straightforward:

<!-- prettier-ignore-start -->
```yaml
  selector:
    workload-type: source-code
```
<!-- prettier-ignore-end -->

Before changing the template reference, let’s take a moment to think about why the deploy step needs a new reference. In
the general case, it is completely fine for 2 supply chains to refer to common templates; that reusability is a feature
of Cartographer. But in this case, we know that the deploy step of our two supply chains have different dependencies. In
our original supply chain the deploy step depended only on the workload values. In our new supply chain we’ve declared
that the deploy step depends on values from the build-image step. This indicates to us that the templates will need to
differ. So we’ll need to write a new deploy template. We’ll refer to it in the supply chain:

<!-- prettier-ignore-start -->
```yaml
    - name: deploy
      templateRef:
        kind: ClusterTemplate
        name: app-deploy-from-sc-image
      ...
```
<!-- prettier-ignore-end -->

Finally, we'll need a new service account for this supply chain, one that has permission to create the objects in both
templates. We'll specify a name for that service account now and create it below (after completing our templates).

<!-- prettier-ignore-start -->
```yaml
  serviceAccountRef:
    name: cartographer-from-source-sa
    namespace: default
```
<!-- prettier-ignore-end -->

We can see our final supply chain defined here:

{{< tutorial supply-chain.yaml >}}

#### Templates

Now we’re ready to define our templates. Let’s begin with the template for the deploy step, as we’re familiar with it
already. There’s only one field that will change; previously the image location was defined by the workload.

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterTemplate
metadata:
  name: app-deploy
spec:
  template:
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: $(workload.metadata.name)$-deployment
      labels:
        app: $(workload.metadata.name)$
    spec:
      replicas: 3
      selector:
        matchLabels:
          app: $(workload.metadata.name)$
      template:
        metadata:
          labels:
            app: $(workload.metadata.name)$
        spec:
          serviceAccountName: $(params.image-pull-sa-name)$
          containers:
            - name: $(workload.metadata.name)$
              image: $(workload.spec.image)$ # <=== No longer the proper source
  params:
    - name: image-pull-sa-name
      default: expected-service-account
```

Now the template expects that value to come from a previous step in the supply chain. So we’ll simply replace

<!-- prettier-ignore-start -->
```yaml
              image: $(workload.spec.image)$
```
<!-- prettier-ignore-end -->

with

<!-- prettier-ignore-start -->
```yaml
              image: $(images.built-image.image)$
```
<!-- prettier-ignore-end -->

Let’s break down that syntax. In the supply chain we specified that we were providing an array of images to this
template. So we start with `images`. In the supply chain we further declared that the one image in the array of images
would have the name "built-image", so we continue `images.built-image`. Finally, images provide a single value, an image
location (if we were providing sources, each source would provide both a url and a revision). So we complete the
reference: `images.built-image.image`.

Finally, we'll give this ClusterTemplate a new name (already specified in the supply chain):

```yaml
metadata:
  name: app-deploy-from-sc-image
```

The template for our deploy step is complete:

{{< tutorial app-deploy-template.yaml >}}

Now we’re ready to create our new template, `image-builder`. We’re going to rely on kpack, a kubernetes native container
build service. We’ll need to template out a kpack Image object. Much of this will be familiar from the
[Build Your First Supply Chain](../first-supply-chain) tutorial. We’ll have a value (the location of the source code)
that is only known by the application developer. We’ll also have values that must remain unique among templated objects
in the name space, for which we’ll use the name of the workload. And similar to the tutorial on Using Params, we’ll
leverage params to provide a default image registry but allow devs to specify another image registry if they so desire.
Let’s look at the .spec.template field of our ClusterImageTemplate:

<!-- prettier-ignore-start -->
```yaml
  template:
    apiVersion: kpack.io/v1alpha2
    kind: Image
    metadata:
      name: $(workload.metadata.name)$
    spec:
      tag: $(params.image_prefix)$$(workload.metadata.name)$
      serviceAccountName: $(params.image-pull-sa-name)$
      builder:
        kind: ClusterBuilder
        name: my-builder
      source:
        git:
          url: $(workload.spec.source.git.url)$
          revision: $(workload.spec.source.git.ref.branch)$
```
<!-- prettier-ignore-end -->

Though learners may not be as familiar with kpack Images as they are with the Deployment resource, there’s nothing novel
here from a Cartographer perspective.

Since our template leverages 2 params, we’ll provide default values for them:

<!-- prettier-ignore-start -->
```yaml
  params:
    - name: image-pull-sa-name
      default: expected-service-account
    - name: image_prefix
      default: 0.0.0.0:5000/example-basic-sc-
```
<!-- prettier-ignore-end -->

_For those using using dockerhub, gcr or other registry, substitute the appropriate default value for image_prefix. For
those using the hack script to create a local registry for this tutorial, run the following command to get the ip
address to use in the place of 0.0.0.0:_

```shell
$ ./hack/ip.py
```

So far, creating our ClusterImageTemplate has been similar to what we’ve done in previous tutorials using the
ClusterTemplate. There’s only one novel step we must do. We must specify what value will be exposed from this template.
Again, we’re using a ClusterImageTemplate and this custom resource requires the specification of an imagePath. This is
the path on the templated object where we will find our desired value.
[The documentation for the kpack Image resource](https://github.com/pivotal/kpack/blob/main/docs/image.md) lets us know
“When an image resource has successfully built with its current configuration, its status will report the up to date
fully qualified built OCI image reference.” We can see this value is in the `.status.latestImage` field. So that is the
path that we put in the ClusterImageTemplate’s `.spec.imagePath`:

<!-- prettier-ignore-start -->
```yaml
  imagePath: .status.latestImage
```
<!-- prettier-ignore-end -->

We now have our full object:

{{< tutorial image-builder-template.yaml >}}

#### Service Account

We now have our two templates. It's time to create the service account that will give Cartographer permission to create
the objects in the templates (a deployment and a kpack image). In the default namespace (because that's where we
declared it would be in the supply chain) we create:

{{< tutorial cartographer-service-account.yaml >}}

#### kpack Dependencies

From Cartographer's perspective, we've completed our work as app operators. We've created our templates, our supply
chain and our service account. When app devs create workloads, Cartographer will happily begin creating objects. But
before we switch over to the app dev role, we have to consider the objects that we're creating and whether they have any
dependencies. We've already seen how the deployment will rely on a service account existing with imagePullCredentials.
Similarly, our kpack image relies on both a service account and on other kpack resources having been installed in the
cluster.

This isn’t a tutorial on kpack, so we’ll quickly specify objects below. Learners interested in exploring kpack should
[read more here](https://github.com/pivotal/kpack/blob/main/docs/tutorial.md).

{{< tutorial kpack-boilerplate.yaml >}}

{{< tutorial registry-service-account.yaml >}}

- Note that the ClusterBuilder object has a placeholder value in the `.spec.tag` field. Either a local registry ip
  address or another image registry specification should be put here.

- Note that the registry-credentials secret has placeholder values for the `.stringData` field. If learners are using
  the local registry from the hack script, use the ./hack/ip.py ip address in place of the 0.0.0.0 address (the username
  and password will then be correct). Otherwise learners should put the appropriate credentials for their image
  registry.

Now we've completed our work as app operators. Let’s step into our role as app devs.

### App Dev Steps

As is appropriate for an app platform, the complication undertaken by the app operators above is hidden from the app
devs. As devs, all we need to know is that our request has been answered: we can now submit a workload that specifies
the location of our source code and it will be built and deployed. Let’s do that!

For our app we’ve used a copy of one of the many paketo buildpack sample apps.
[That copy resides here](https://github.com/waciumawanjohi/demo-hello-world)

_Note: Copying a paketo sample app ensures that our app will be built. Troubleshooting kpack builds of an arbitrary
application is far outside the scope of this tutorial. The sample apps
[can be found here](https://github.com/paketo-buildpacks/samples). Note that for expediency we only installed kpack with
the ability to build golang and java applications. Users should feel free to install additional paketo buildpacks if
desired._

Let’s specify the location of our source code in the workload's spec:

```yaml
spec:
  source:
    git:
      ref:
        branch: main
      url: https://github.com/waciumawanjohi/demo-hello-world
```

We have a new type of app now; it is no longer pre-built. Let’s change our workload type label and match it to the new
supply chain's selector.

<!-- prettier-ignore-start -->
```yaml
    workload-type: source-code
```
<!-- prettier-ignore-end -->

And as app devs, we're done! Let’s look at the complete workload:

{{< tutorial workload.yaml >}}

## Observe

### Workload

Looking at the workload, we can see that it resolves to a healthy state:

```shell
$ kubectl get -o yaml workload hello-again
```

```yaml
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  ...
  name: hello-again
status:
  conditions:
  ...
  - lastTransitionTime: ...
    message: ""
    reason: Ready
    status: "True"
    type: Ready
```

### Stamped Objects

All the objects that we templated out exist.

```shell
$ kubectl get -o yaml image.kpack.io hello-again
```

```yaml
apiVersion: kpack.io/v1alpha2
kind: Image
metadata:
  ...
  name: hello-again
spec: ...
status:
  buildCacheName: hello-again-cache
  buildCounter: 1
  conditions:
  - lastTransitionTime: ...
    status: "True"
    type: Ready
  - lastTransitionTime: ...
    status: "True"
    type: BuilderReady
  latestBuildImageGeneration: 1
  latestBuildReason: CONFIG
  latestBuildRef: hello-again-build-1
  latestImage: 0.0.0.0:5000/example-basic-sc-hello-again@sha256:abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz12
  latestStack: ...
  observedGeneration: 1
```

```shell
$ kubectl get -o yaml deployment hello-again-deployment
```

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  ...
  name: hello-again-deployment
spec:
  ...
  template:
    ...
    spec:
      containers:
      - image: 0.0.0.0:5000/example-basic-sc-hello-again@sha256:abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz12
        ...
      ...
status:
  availableReplicas: 3
  conditions:
  - status: "True"
    type: Available
    ...
  - status: "True"
    type: Progressing
    ...
  ...
```

The conditions of all of these objects are healthy. In addition, we can see where the kpack image's
`.status.latestImage` field has been used by the deployment's `spec.template.spec.containers[0].image` field.

## Wrap Up

Another tutorial under your belt; you’re well on your way to building a robust app platform for your organization! In
this tutorial you learned:

- How to expose a value from a template to other steps in a supply chain
- How to consume an earlier exposed value in a template
- How to add to a supply chain
- How to create supply chains with different behavior/templates
