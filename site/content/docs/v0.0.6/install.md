# Installing Cartographer

There are three recommended ways of installing Cartographer:

1. with a single YAML file (quick!)
2. using an [imgpkg bundle], making easier for relocation
3. using a [carvel Packaging] objects, for an opinionated package-based workflow

If all you want to do is get started quick and fetching images from DockerHub is not an issue,
[`1`](#1-single-file-quick-install) is the way to go. Otherwise, check out the other approaches (especially if you want
to have everything relocated to a registry of your own).

tl;dr _(having [cert-manager] already installed)_:

```bash
kubectl create namespace cartographer-system
kubectl apply -f https://github.com/vmware-tanzu/cartographer/releases/latest/download/cartographer.yaml
```

## Pre-requisites

### Admin capabilities in a recent-enough Kubernetes cluster (v1.19+)

Currently, Cartographer doesn't provide support for granular permissions (see [#51]), as such, to configure the supply
chain (which is cluster-wide), you must have admin capabilities in the cluster you're targetting.

### [cert-manager] for managing the controller's certificates

In order to have Cartographer's validation webhooks up and running in the cluster, [cert-manager] is utilized to
generate TLS certificates as well to keep them up to date.

First, verify if you have the dependency installed:

```bash
kubectl get crds certificates.cert-manager.io
```

```console
NAME                           CREATED AT
certificates.cert-manager.io   2021-08-27T18:41:40Z
```

In case you don't (i.e., you see _""certificates.cert-manager.io" not found"_), proceed with installing it.

```bash
CERT_MANAGER_VERSION=1.5.3

kapp deploy --yes -a cert-manager \
	-f https://github.com/jetstack/cert-manager/releases/download/v$CERT_MANAGER_VERSION/cert-manager.yaml
```

```console
Target cluster 'https://127.0.0.1:39495' (nodes: kind-control-plane)

Changes
Namespace     Name                     Kind               ...
(cluster)     cert-manager             Namespace          ...
^             cert-manager-cainjector  ClusterRole        ...
^             cert-manager-cainjector  ClusterRoleBinding ...
...
7:53:32AM: ok: reconcile customresourcedefinition/issuers.cert-manager.io (apiextensions.k8s.io/v1) cluster
7:53:32AM: ---- applying complete [6/6 done] ----
7:53:32AM: ---- waiting complete [6/6 done] ----

Succeeded
```

ps.: although we recommend using [kapp] as provided in the instructions you'll see here, its use can be replaced by
`kubectl apply`.

## 1. Single file, Quick install

Cartographer releases include a file (`cartographer.yaml`) that contains all necessary files for installing it in a
cluster

```console
Resources in file './bundle/cartographer.yaml'

Namespace            Name                              Kind
(cluster)            cartographer-cluster-admin        ClusterRoleBinding
^                    clusterconfigtemplates.carto.run  CustomResourceDefinition
^                    clusterimagetemplates.carto.run   CustomResourceDefinition
^                    clustersourcetemplates.carto.run  CustomResourceDefinition
^                    clustersupplychains.carto.run     CustomResourceDefinition
^                    clustersupplychainvalidator       ValidatingWebhookConfiguration
^                    clustertemplates.carto.run        CustomResourceDefinition
^                    pipelines.carto.run               CustomResourceDefinition
^                    runtemplates.carto.run            CustomResourceDefinition
^                    workloads.carto.run               CustomResourceDefinition

cartographer-system  cartographer-controller           Deployment
^                    cartographer-controller           ServiceAccount
^                    cartographer-webhook              Certificate
^                    cartographer-webhook              Secret
^                    cartographer-webhook              Service
^                    private-registry-credentials      Secret
^                    selfsigned-issuer                 Issuer
```

providing all the necessary objects that would entail a full installation of it.

To install it, first create the namespace where the controller will be installed:

```bash
kubectl create namespace cartographer-system
```

```console
namespace/cartographer-system created
```

and then, submit the objects included in the release:

```bash
kubectl apply -f https://github.com/vmware-tanzu/cartographer/releases/latest/download/cartographer.yaml
```

```console
customresourcedefinition.apiextensions.k8s.io/clusterconfigtemplates.carto.run configured
customresourcedefinition.apiextensions.k8s.io/clusterimagetemplates.carto.run configured
...
secret/cartographer-webhook created
```

## 2. Bundle tarball

This installation method is great for those that want to host Cartographer's image and installation objects by
themselves - either so they live in a public container image registry of their own, or for air-gapped scenarios where
bits need to be moved in an offline fashion.

First, head to the [releases page] and download the `bundle.tar` file available for the release you want to install.

```bash
curl -SOL https://github.com/vmware-tanzu/cartographer/releases/latest/download/bundle.tar
```

This bundle contains everything we need to install Cartographer, from container images to Kubernetes manifests, it's all
in the bundle.

First, relocate it from the tarball you just downloaded to a container image registry that the cluster you're willing to
install Cartographer has access to:

```bash
# set the repository where images should be reloated to, for instance, to
# relocate to a project named "foo" in DockerHub: DOCKER_REPO=foo
#
DOCKER_REPO=10.188.0.3:5000


# relocate
#
imgpkg copy \
  --tar bundle.tar \
  --to-repo ${DOCKER_REPO?:Required}/cartographer-bundle \
  --lock-output cartographer-bundle.lock.yaml
```

```console
copy | importing 2 images...
copy | done uploading images
Succeeded
```

With the the bundle and all Cartographer-related images moved to the destination registry, we can move on to pulling the
YAML files that we can use to submit to Kubernetes for installing Cartographer:

```bash
# pull to the directory `cartographer-bundle` the contents of the imgpkg
# bundle as specified by the lock file.
#
imgpkg pull \
  --lock cartographer-bundle.lock.yaml \
  --output ./cartographer-bundle
```

```console
Pulling bundle '10.188.0.3:5000/cartographer-bundle@sha256:e296a316385a57048cb189a3a710cecf128a62e77600a495f32d46309c6e8113'
  Extracting layer 'sha256:0b93d1878c9be97b872f08da8d583796985919df345c39c874766142464d80e7' (1/1)

Locating image lock file images...
The bundle repo (10.188.0.3:5000/cartographer-bundle) is hosting every image specified in the bundle's Images Lock file (.imgpkg/images.yml)

Succeeded
```

Create the namespace where Cartographer's objects will be placed:

```bash
kubectl create namespace cartographer-system
```

```console
namespace/cartographer-system created
```

If the registry you pushed the bundle to is accessible to the cluster without any extra authentication needs, skip this
step. Otherwise, make sure you provide to the `Deployment` object that runs the controller the image pull secrets
necessary for fetching the image.

By default, the controller's ServiceAccount points at a placeholder secret called 'private-registry-credentials' to be
used as the image pull secret, so, by creating such secret in the `cartographer-system` namespace Kubernetes will then
be able to use that secret as the credenital provider for fetching the controller's image:

```bash
# create a secret that will have .dockerconfigjson populated with the
# credentials for the image registry.
#
kubectl create secret -n cartographer-system \
  docker-registry private-registry-credentials \
  --docker-server=$DOCKER_REPO \
  --docker-username=admin \
  --docker-password=admin
```

```console
secret/private-registry-credentials created
```

Now that we have the Kubernetes YAML under the `./cartographer-bundle` directory, we can make use of a combination of
`kbld` and `kapp` to submit the Kubernetes the final objects that will define the installation of Cartographer already
pointing all the image references to your registry:

```bash
# submit to kubernetes the kubernetes objects that describe the installation of
# Cartographer already pointing all images to the registry we configured
# ($DOCKER_REPO).
#
kapp deploy -a cartographer -f <(kbld -f ./cartographer-bundle)
```

```console
resolve | final: 10.188.0.3:5000/cartographer-27a7f49719016b1cfc534e74c3d36805@sha256:26c3dd5c8a38658218f22c03136c3b8adf45398a72ea7dde9524ec24bfa04783 -> 10.188.0.3:5000/cartographer-bundle@sha256:26c3dd5c8a38658218f22c03136c3b8adf45398a72ea7dde9524ec24bfa04783
Target cluster 'https://127.0.0.1:32907' (nodes: cartographer-control-plane)

Changes

Namespace            Name                              Kind                      Conds.  Age  Op      Op st.  Wait to    Rs  Ri
(cluster)            clusterconfigtemplates.carto.run  CustomResourceDefinition  0/0 t   49s  update  -       reconcile  ok  -
^                    clusterimagetemplates.carto.run   CustomResourceDefinition  0/0 t   49s  update  -       reconcile  ok  -
^                    clustersourcetemplates.carto.run  CustomResourceDefinition  0/0 t   49s  update  -       reconcile  ok  -
^                    clustersupplychains.carto.run     CustomResourceDefinition  0/0 t   48s  update  -       reconcile  ok  -
...


6:34:57PM: ok: reconcile customresourcedefinition/clustersupplychains.carto.run (apiextensions.k8s.io/v1) cluster
6:34:57PM: ---- applying complete [11/11 done] ----
6:34:57PM: ---- waiting complete [11/11 done] ----

Succeeded
```

_(see the
[Kubernetes official documentation](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/#create-a-secret-by-providing-credentials-on-the-command-line)
on how to create a Secret to pull images from a private Docker registry or repository)._

## 3. Installation using Carvel Packaging

Another way that you can go about installing Cartographer is with the use of [carvel Packaging] provided by [kapp
controller]. These, when used alongside [secretgen controller], provide a great experience for, in a declarative way,
installing Cartographer.

To make use of them, first, make sure those pre-requisites above are satified

### Prerequisites

1. admin access to a Kubernetes Cluster and [cert-manager]

2. [kapp-controller] is already installed in the cluster

```bash
kubectl get crd packageinstalls.packaging.carvel.dev
```

```console
NAME                                   CREATED AT
packageinstalls.packaging.carvel.dev   2021-09-13T14:32:00Z
```

In case you don't (i.e., you see _"packageinstalls.packaging.carvel.dev" not found_), proceed with installing it.

```bash
KAPP_CONTROLLER_VERSION=0.30.0

kapp deploy --yes -a kapp-controller \
  -f https://github.com/vmware-tanzu/carvel-kapp-controller/releases/download/v$KAPP_CONTROLLER_VERSION/release.yml
```

```console
Target cluster 'https://127.0.0.1:39993' (nodes: cartographer-control-plane)

Changes

Namespace        Name                                                    Kind
(cluster)        apps.kappctrl.k14s.io                                   CustomResourceDefinition
^                internalpackagemetadatas.internal.packaging.carvel.dev  CustomResourceDefinition
^                internalpackages.internal.packaging.carvel.dev          CustomResourceDefinition
^                kapp-controller                                         Namespace


2:56:08PM: ---- waiting on 1 changes [14/15 done] ----
2:56:13PM: ok: reconcile apiservice/v1alpha1.data.packaging.carvel.dev (apiregistration.k8s.io/v1) cluster
2:56:13PM: ---- applying complete [15/15 done] ----
2:56:13PM: ---- waiting complete [15/15 done] ----

Succeeded
```

3. [secretgen-controller] installed

```bash
kubectl get crd secretexports.secretgen.carvel.dev
```

```console
NAME                                 CREATED AT
secretexports.secretgen.carvel.dev   2021-09-20T18:10:10Z
```

In case you don't (i.e., you see _"secretexports.secretgen.carvel.dev" not found_), proceed with installing it.

```bash
SECRETGEN_CONTROLLER_VERSION=0.6.0

kapp deploy --yes -a secretgen-controller \
  -f https://github.com/vmware-tanzu/carvel-secretgen-controller/releases/download/v$SECRETGEN_CONTROLLER_VERSION/release.yml
```

```console
Target cluster 'https://127.0.0.1:45829' (nodes: cartographer-control-plane)

Changes

Namespace             Name                                       Kind                      Conds.  Age  Op      Op st.  Wait to    Rs  Ri
(cluster)             certificates.secretgen.k14s.io             CustomResourceDefinition  -       -    create  -       reconcile  -   -
^                     passwords.secretgen.k14s.io                CustomResourceDefinition  -       -    create  -       reconcile  -   -
^                     rsakeys.secretgen.k14s.io                  CustomResourceDefinition  -       -    create  -       reconcile  -   -
...
6:13:25PM: ok: reconcile deployment/secretgen-controller (apps/v1) namespace: secretgen-controller
6:13:25PM: ---- applying complete [11/11 done] ----
6:13:25PM: ---- waiting complete [11/11 done] ----

Succeeded
```

3. the `default` service account has the capabilities necessary for installing submitting all those objects above to the
   cluster

```bash
kubectl create clusterrolebinding default-cluster-admin \
	--clusterrole=cluster-admin \
	--serviceaccount=default:default
```

```console
clusterrolebinding.rbac.authorization.k8s.io/default-cluster-admin created
```

### Package installation

With the prerequisites satisfied, go ahead and download the `package*` files, as well as the `imgpkg` bundle:

```bash
CARTOGRAPHER_VERSION=v0.0.6

curl -SOL https://github.com/vmware-tanzu/cartographer/releases/download/v$CARTOGRAPHER_VERSION/bundle.tar
curl -SOL https://github.com/vmware-tanzu/cartographer/releases/download/v$CARTOGRAPHER_VERSION/package.yaml
curl -SOL https://github.com/vmware-tanzu/cartographer/releases/download/v$CARTOGRAPHER_VERSION/package-metadata.yaml
curl -SOL https://github.com/vmware-tanzu/cartographer/releases/download/v$CARTOGRAPHER_VERSION/package-install.yaml
```

First, relocate the bundle:

```bash
# set the repository where images should be reloated to, for instance, to
# relocate to a project named "foo" in DockerHub: DOCKER_REPO=foo
#
DOCKER_REPO=10.188.0.3:5000


# relocate
#
imgpkg copy \
  --tar bundle.tar \
  --to-repo ${DOCKER_REPO?:Required}/cartographer-bundle \
  --lock-output cartographer-bundle.lock.yaml
```

Now that the bundle is in our repository, update the Package object to point at it (use the image from
`cartographer-bundle.lock.yaml`):

```diff
  apiVersion: data.packaging.carvel.dev/v1alpha1
  kind: Package
  ..
    template:
      spec:
        fetch:
        - imgpkgBundle:
-           image: IMAGE
+           image: 10.188.0.3:5000/cartographer-bundle@sha256:e296a3163
        template:
```

That done, submit the packaging objects to Kubernetes so that `kapp-controller` will materialize them into an
installation of Cartographer:

```bash
kapp deploy --yes -a cartographer \
  -f ./package-metadata.yaml \
  -f ./package.yaml \
  -f ./package-install.yaml
```

```console
Target cluster 'https://127.0.0.1:42483' (nodes: cartographer-control-plane)

Changes

Namespace  Name                              Kind             Conds.  Age  Op      Op st.  Wait to    Rs  Ri
default    cartographer.carto.run            PackageMetadata  -       -    create  -       reconcile  -   -
^          cartographer.carto.run.0.0.0-dev  Package          -       -    create  -       reconcile  -   -
^          cartographer.carto.run.0.0.0-dev  PackageInstall   -       -    create  -       reconcile  -   -

...

1:14:44PM: ---- applying 2 changes [0/3 done] ----
1:14:44PM: create packagemetadata/cartographer.carto.run (data.packaging.carvel.dev/v1alpha1) namespace: default
1:14:54PM: ok: reconcile packageinstall/cartographer.carto.run.0.0.0-dev (packaging.carvel.dev/v1alpha1) namespace: default
1:14:54PM: ---- applying complete [3/3 done] ----
1:14:54PM: ---- waiting complete [3/3 done] ----

Succeeded
```

if you relocated the images here to a private registry that requires authentication, make sure you create a `Secret`
with the credentials to the registry as well as a `SecretExport` object to make those credentials available to other
namespaces.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: shared-registry-credentials
type: kubernetes.io/dockerconfigjson # needs to be this type
stringData:
  .dockerconfigjson: |
    {
            "auths": {
                    "<registry>": {
                            "username": "<username>",
                            "password": "<password>"
                    }
            }
    }

---
apiVersion: secretgen.carvel.dev/v1alpha1
kind: SecretExport
metadata:
  name: shared-registry-credentials
spec:
  toNamespaces:
    - "*"
```

[#1]: https://github.com/vmware-tanzu/cartographer/issues/51
[admission webhook]: https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/
[carvel packaging]: https://carvel.dev/kapp-controller/docs/latest/packaging/
[cert-manager]: https://github.com/jetstack/cert-manager
[imgpkg bundle]: https://carvel.dev/imgpkg/docs/latest/
[kapp-controller]: https://carvel.dev/kapp-controller/
[kapp]: https://carvel.dev/kapp/
[kind]: https://github.com/kubernetes-sigs/kind
[releases page]: https://github.com/vmware-tanzu/cartographer/releases
