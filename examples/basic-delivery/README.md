# Source to App Config In Git

**before you proceed**: the example in this directory illustrates the use of
the latest components and functionality of Cartographer (including some that
may not have been included in the latest release yet). Make sure to check out
the version of this document in a tag that matches the latest version (for
instance, https://github.com/vmware-tanzu/cartographer/tree/v0.0.7/examples).

---

The [gitwriter-sc] example illustrates how a supply chain can write the configuration
for a new app into a git repository. This example picks up from that point, using 
Delivery to read the configuration from git and create the app in the cluster.

```

  git --> knative service

```

As with [basic-sc], the directories here are structured in a way to reflect which Kubernetes
objects would be set by the different personas in the system. But with Delivery, the
app developer persona need not be involved; the app operator submits all objects.


```
  '── shared | basic-delivery
      ├── 00-cluster                         preconfigured cluster-wide objects
      │                                        to configured systems other than
      │                                              cartographer (like, kpack)
      │
      │
      └── app-operator                      cartographer-specific configuration
          ├── supply-chain-templates.yaml            that an app-operator would
          ├── ...                                                        submit
          └── supply-chain.yaml
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

4. The controllers of the objects created in the templates referred to in the [ClusterDelivery] object:

- [source-controller](https://fluxcd.io/docs/gitops-toolkit/source-watcher/#install-flux),
  for providing the ability to find new commits to a git
  repository and making it internally available to other resources

- [kapp-controller](https://carvel.dev/kapp-controller/docs/latest/install/),
  for providing us with the ability of deploying multiple
  Kubernetes objects as a single unit

- [knative-serving](https://knative.dev/docs/install/serving/install-serving-with-yaml/),
  for being the runtime of the application we want to deploy.

5. [tree](https://github.com/ahmetb/kubectl-tree#installation) and
  [yq](https://github.com/kislyuk/yq#installation) for observing objects in the cluster and their fields.

### Resource Requirements
Read [here](../README.md#resource-requirements)

## Running the example in this directory

### Location of files

As with [basic-sc], this example uses two directories with sub-directories of kubernetes resources:
[../shared](../shared) and [.](.). Like [basic-sc], all cluster-wide configuration is in the
shared [../shared/cluster](../shared/cluster/README.md) directory. All other objects for this example
are in the local [./app-operator](./app-operator/README.md) directory.

### Configuring the example

Update [values.yaml](./values.yaml) with information about your git repository, which should already have
a kubernetes object written to it. (e.g. use the same values used in the [gitwriter-sc configuration](../gitwriter-sc/README.md#configuring-the-example))

```yaml
#@data/values
---
# configuration necessary for pushing the config to a git repository.
#
git_writer:
  # the git server, project, and repo name
  repository: github.com/example/example.git
  # the branch to which configuration was pushed
  branch: main
```

Further update [values.yaml](./values.yaml) with information about the registry that holds the image used in the
application. See the [configuration instructions of the basic-sc](../basic-sc/README.md#configuring-the-example) for
more information. These registry values should match the values specified in the supply chain used to write a service to into
your git repository (e.g. the values in [gitwriter-sc/values.yaml](../gitwriter-sc/values.yaml))

```yaml
registry:
  server: https://index.docker.io/v1/
  username: a-username-of-mine
  password: a-very-hard-to-guess-password
```

### Deploying the files

We will deploy in much the same way as other examples, except that we will not deploy the kpack resources in the
shared cluster configuration.

```bash
kapp deploy --yes -a example \
  -f <(ytt --ignore-unknown-comments -f .) \
  -f <(ytt --ignore-unknown-comments -f ../shared/cluster/rbac.yaml -f ./values.yaml) \
  -f <(ytt --ignore-unknown-comments -f ../shared/cluster/secret.yaml -f ./values.yaml) \
  -f <(ytt --ignore-unknown-comments -f ../shared/cluster/serviceaccount.yaml -f ./values.yaml)
```

### Observing the example

When using tree, we see the deliverable has two child objects: a gitrepository picking up an
object definition, and a kapp app deploying that object.

```console
$ kubectl tree deliverable gitops
NAMESPACE  NAME
default    Deliverable/gitops
default    ├─GitRepository/gitops       configuration fetching
default    └─App/gitops                 configuration application
```

The App can be inspected for objects created:

```bash
kubectl get app gitops -o yaml | yq -r .status.inspect.stdout
```

If your git repository contains the output of the [gitwriter-sc], you will find:

```console
Target cluster 'https://...' (nodes: cartographer-control-plane)
Resources in app 'gitops-ctrl'
Namespace  Name          Kind           Owner    Conds.  Rs  Ri                Age
default    gitwriter-sc  Configuration  cluster  1/1 t   ok  -                 2m
^          gitwriter-sc  Ingress        cluster  -       ok  -                 2m
^          gitwriter-sc  Route          cluster  2/4 t   ok  -                 2m
^          gitwriter-sc  Service        kapp     1/3 t   ok  -                 2m
^          gitwriter-sc  Service        cluster  -       ok  External service  2m
Rs: Reconcile state
Ri: Reconcile information
5 resources
Succeeded
```

Following the [basic-sc observation steps](../basic-sc/README.md#observing-the-example) you can port-forward and
curl the newly deployed application.

## Tearing down the example and uninstalling the dependencies

1. Follow the [basic-sc teardown instructions](../basic-sc/README.md#tearing-down-the-example).
2. Manually delete any commits in the git repository that are no longer desired.

## Step by step

This example creates a software supply chain like mentioned above

```

    git --> knative service

```

In [basic-sc] a supply chain creates a knative service and deploys it to the cluster.
Here we assume that a supply chain has already defined this service and the delivery
need only deploy it. We find this service definition in a git repository.

### Need: separating build and production environment

In [gitwriter-sc we discuss the need for splitting build and production environments](../gitwriter-sc/README.md#need-separating-build-and-production-environment)

### Reading the configuration from git

Delivery assumes that the configuration of a cluster, the definition of kubernetes objects
that will be deployed, is already written and stored (while this example uses git,
the object definitions could similarly be stored in an image registry).

The Deliverable is responsible for supplying the location of this repository.

```yaml
apiVersion: carto.run/v1alpha1
kind: Deliverable
metadata:
  name: gitops
  labels:
    app.tanzu.vmware.com/workload-type: deliver # <=== must match the selector on a Delivery
spec:
  source:
    git:
      url: https://github.com/example/repo-with-k8s-objects # <=== the gitops repo
      ref:
        branch: main # <=== the gitops branch
```

A GitRepository object watches this repository. Whenever there is an update to the repo,
the GitRepository exposes the new state to the Delivery. The GitRepository is wrapped in a
ClusterSourceTemplate which is almost exactly like the
[cluster source template in the shared supply chain examples](../shared/app-operator/source-git-repository.yaml).

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSourceTemplate
metadata:
  name: gitops-source
spec:
  urlPath: .status.artifact.url
  revisionPath: .status.artifact.revision

  template:
    apiVersion: source.toolkit.fluxcd.io/v1beta1
    kind: GitRepository
    metadata:
      name: $(deliverable.metadata.name)$
    spec:
      interval: 1m0s
      url: $(deliverable.spec.source.git.url)$
      ref: $(deliverable.spec.source.git.ref)$
      gitImplementation: go-git
      ignore: ""
```

The new state of the repo is then passed to the ClusterDeploymentTemplate. This object is responsible
for deploying kubernetes objects into the cluster. It wraps a kapp app which is able to deploy
objects defined at a given url. The ClusterDeploymentTemplate waits to see that the object has been
properly created.

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterDeploymentTemplate
metadata:
  name: app-deploy
spec:
  # values that will be read from the deployed object
  observedCompletion:
    
    # If this condition is met the template will expose to the Delivery the
    # `deployment` field that was passed in by the Delivery
    succeeded:
    
      # path on the object that is inspected
      key: '.status.conditions[?(@.type=="ReconcileSucceeded")].status'
    
      # value that must match what is found at the above path
      value: 'True'
    
    # If this condition is met, the Deliverable status will show as not Ready
    failed:
      key: '.status.conditions[?(@.type=="ReconcileSucceeded")].status'
      value: 'False'

  # the definition of the object to deploy
  template:
    apiVersion: kappctrl.k14s.io/v1alpha1
    kind: App
    metadata:
      name: $(deliverable.metadata.name)$
    spec:
      serviceAccountName: default
      fetch:
        - http:
    
            # an example of a field leveraging the Deployment passed in by the Delivery
            url: $(deployment.url)$
    
      template:
        - ytt: {}
      deploy:
        - kapp: {}
```

### Including the templates in the delivery

These templates are then referenced in the Delivery.

Note that while a ClusterSourceTemplate exposes a `source`, in a `ClusterDelivery` this value can
be cast as a `deployment` for consumption by a `ClusterDeploymentTemplate`.

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterDelivery
metadata:
  name: deliver
spec:
  # a selector to match with deliverables
  selector:
    # a list of labels expected on deliverables that match
    app.tanzu.vmware.com/workload-type: deliver
  resources:
    - name: source-provider
      templateRef:
        kind: ClusterSourceTemplate
        name: gitops-source

    - name: deployer
      templateRef:
        kind: ClusterDeploymentTemplate
        name: app-deploy
      # in a supply chain, the output of a source template is only available as a `source`
      # in a delivery, a source template's output can be consumed as a `deployment`
      # every ClusterDeploymentTemplate must be passed exactly one deployment. the deployment
      # is not named (as other inputs can be)
      deployment:
        resource: source-provider
```

In this way, the delivery ensures that every update to the kubernetes objects in the gitops repo
results in an updated deployment on the production cluster.

[ClusterDelivery]: ./app-operator/delivery.yaml
[carvel]: https://carvel.dev/
[ytt]: https://carvel.dev/ytt
[kapp]: https://carvel.dev/kapp
[source-controller]: https://github.com/fluxcd/source-controller
[kapp-controller]: https://carvel.dev/kapp-controller/
[knative-serving]: https://knative.dev/docs/serving/
[basic-sc]: ../basic-sc/README.md
[gitwriter-sc]: ../gitwriter-sc/README.md
