# Project Cartographer

Cartographer is a Kubernetes native [Choreographer]. It allows users to configure K8s resources into re-usable [_Supply Chains_](site/content/docs/reference.md#ClusterSupplyChain) that can be used to define all of the stages that an [_Application Workload_](site/content/docs/reference.md#Workload) must go through to get to an environment.

[Choreographer]: https://solace.com/blog/microservices-choreography-vs-orchestration/#:~:text=Orchestration%20entails%20actively%20controlling%20all,without%20requiring%20supervision%20and%20instructions

Cartographer also allows for separation of controls between a user who is responsible for defining a Supply Chain (known as a App Operator) and a user who is focused on creating applications (Developer). These roles are not necessarily mutually exclusive, but provide the ability to create a separation concern.


## Documentation

Detailed documentation for Cartographer can be found in the `site` folder of this repository:

* [About Cartographer](site/content/docs/about.md): Details the design and philosophy of Cartographer
* [Examples](examples/source-to-knative-service/README.md): Contains an example of using Cartographer to create a supply chain that takes a repository, creates and image and deploys it to a cluster
* [Spec Reference](site/content/docs/reference.md): Detailed descriptions of the CRD Specs for Cartographer


## Getting Started

An example of using Cartographer to define a Supply Chain that pulls code from a repository, builds an image for the code and deploys it the the same cluster can be found in the [examples folder of this repository](examples/source-to-knative-service/README.md)


## Installation

### Pre-requisites

1. A running Kubernetes cluster (v1.19+) with admin permissions

In case you don't, [kind] is a great alternative for running Kubernetes
locally, specially if you already have Docker installed (that is all it takes).

2. [cert-manager], so that certificates can be created and maintained to
   support the controller's [admission webhook]


### Steps

#### 1. Ensure **[cert-manager]** is installed in the cluster

In order to have Cartographer's validation webhooks up and running in the
cluster, [cert-manager] is utilized to generate TLS certificates as well to keep
them up to date.

First, verify if you have the dependency installed:

```bash
kubectl get crds certificates.cert-manager.io
```
```console
NAME                           CREATED AT
certificates.cert-manager.io   2021-08-27T18:41:40Z
```

In case you don't (i.e., you see _""certificates.cert-manager.io" not found"_),
proceed with installing it.

```bash
kapp deploy --yes -a cert-manager \
	-f https://github.com/jetstack/cert-manager/releases/download/v1.5.3/cert-manager.yaml
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

ps.: although we recommend using [kapp] as provided in the instructions you'll
see here, its use can be replaced by `kubectl apply`.


#### 2. Submit Project Cartographer's Kubernetes objects to the cluster

With the prerequisites met, it's a matter of submitting to Kubernetes the
objects that extend its API and provide the foundation for the controller to
run inside the cluster.

```bash
kapp deploy --yes -a cartographer -f ./releases/release.yaml
```
```console
Target cluster 'https://127.0.0.1:34135' (nodes: kind-control-plane)

Changes

Namespace            Name                                Kind                            Conds.  Age  Op      Op st.  Wait to    Rs  Ri
(cluster)            cartographer-cluster-admin          ClusterRoleBinding              -       -    create  -       reconcile  -   -
^                    clusterconfigtemplates.carto.run    CustomResourceDefinition        -       -    create  -       reconcile  -   -
^                    clusterimagetemplates.carto.run     CustomResourceDefinition        -       -    create  -       reconcile  -   -
^                    clustersourcetemplates.carto.run    CustomResourceDefinition        -       -    create  -       reconcile  -   -
^                    clustersupplychains.carto.run       CustomResourceDefinition        -       -    create  -       reconcile  -   -
^                    clustersupplychainvalidator         ValidatingWebhookConfiguration  -       -    create  -       reconcile  -   -
^                    clustertemplates.carto.run          CustomResourceDefinition        -       -    create  -       reconcile  -   -
^                    deliverables.carto.run              CustomResourceDefinition        -       -    create  -       reconcile  -   -
^                    workloads.carto.run                 CustomResourceDefinition        -       -    create  -       reconcile  -   -
cartographer-system  cartographer-controller             Deployment                      -       -    create  -       reconcile  -   -
^                    cartographer-controller             ServiceAccount                  -       -    create  -       reconcile  -   -
^                    cartographer-webhook                Certificate                     -       -    create  -       reconcile  -   -
^                    cartographer-webhook                Secret                          -       -    create  -       reconcile  -   -
^                    cartographer-webhook                Service                         -       -    create  -       reconcile  -   -
^                    selfsigned-issuer                   Issuer                          -       -    create  -       reconcile  -   -

Op:      15 create, 0 delete, 0 update, 0 noop
Wait to: 15 reconcile, 0 delete, 0 noop

Continue? [yN]:

8:25:17AM: ---- applying 11 changes [0/15 done] ----
8:25:18AM: create secret/cartographer-webhook (v1) namespace: cartographer-system
8:25:18AM: create customresourcedefinition/clusterimagetemplates.carto.run (apiextensions.k8s.io/v1) cluster
...

Succeeded
```


ps.: if you didn't use `kapp`, but instead just `kubectl apply`, make sure you
wait for the deployment to finish before proceeding as `kubectl apply` doesn't
wait by default:

```bash
kubectl get deployment --namespace cartographer-system --watch
```
```console
NAME                      READY   UP-TO-DATE   AVAILABLE   AGE
cartographer-controller   1/1     1            1           3s
```

When "READY" reaches `1/1` (i.e., all the instances are up and running), hit
`CTRL+c` to interrupt the watch sessions, and you're good to go!

Once finished, Project Cartographer has been installed in the cluster -
navigate to the [examples directory](./examples) for a walkthrough.


### extra: installation using Carvel Packaging

Although Cartographer can be installed via plain `kubectl apply` or  `kapp
deploy` like mentioned above, this repository also provides [carvel Packaging]
objects.

To make use of them, first, make sure those pre-requisites above are satified

1. admin access to a Kubernetes Cluster and [cert-manager]

2. [kapp-controller] is already installed in the cluster

```bash
kubectl get crd packageinstalls.packaging.carvel.dev
```
```console
NAME                                   CREATED AT
packageinstalls.packaging.carvel.dev   2021-09-13T14:32:00Z
```

In case you don't (i.e., you see _"packageinstalls.packaging.carvel.dev" not
found_), proceed with installing it.

```bash
kapp deploy --yes -a kapp-controller \
  -f https://github.com/vmware-tanzu/carvel-kapp-controller/releases/download/v0.24.0/release.yml
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

3. the `default` service account has the capabilities necessary for installing
   submitting all those objects above to the cluster

```bash
kubectl create clusterrolebinding default-cluster-admin \
	--clusterrole=cluster-admin \
	--serviceaccount=default:default
```
```console
clusterrolebinding.rbac.authorization.k8s.io/default-cluster-admin created
```

That done, submit the packaging objects to Kubernetes so that `kapp-controller`
will materialize them into an installation of Cartographer:

```bash
kapp deploy --yes -a cartographer -f ./release/package
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

ps.: if you relocated the images here to a private registry that requires
authentication, make sure you create a Secret with the credentials to the
registry as well as a `SecretExport` object to make those credentials available
to other namespaces.


```yaml
apiVersion: v1
kind: Secret
metadata:
  name: registry-credentials
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
  name: registry-credentials
spec:
  toNamespaces:
    - "*"
```


## Uninstall

Having installed all the objects using [kapp], which keeps track of all of them
as a single unit (an app), we can uninstall everything by just referencing that
name:

```bash
kapp delete -a cartographer
kubectl delete namespace cartographer-system
```
```console
Target cluster 'https://127.0.0.1:34135' (nodes: kind-control-plane)

Changes

Namespace            Name                                     Kind                            Conds.  Age  Op      Op st.  Wait to  Rs  Ri
(cluster)            cartographer-cluster-admin               ClusterRoleBinding              -       11s  delete  -       delete   ok  -
^                    clusterconfigtemplates.carto.run         CustomResourceDefinition        2/2 t   12s  delete  -       delete   ok  -
...
^                    selfsigned-issuer                        Issuer                          1/1 t   10s  delete  -       delete   ok  -

Op:      0 create, 15 delete, 0 update, 5 noop
Wait to: 0 reconcile, 20 delete, 0 noop

Continue? [yN]: y
...
8:28:22AM: ok: delete pod/cartographer-controller-dbcf767b8-bw2nf (v1) namespace: cartographer-system
8:28:22AM: ---- applying complete [20/20 done] ----
8:28:22AM: ---- waiting complete [20/20 done] ----

Succeeded
```

In case you used `kubectl apply` instead of `kapp`, you can point at the same
file used to install (`./releases/release.yaml`) but then use the `delete`
command:

```bash
kubectl delete -f ./releases/release.yaml
kubectl delete namespace cartographer-system
```
```console
customresourcedefinition.apiextensions.k8s.io "clusterconfigtemplates.carto.run" deleted
customresourcedefinition.apiextensions.k8s.io "clusterimagetemplates.carto.run" deleted
customresourcedefinition.apiextensions.k8s.io "clustersourcetemplates.carto.run" deleted
customresourcedefinition.apiextensions.k8s.io "clustersupplychains.carto.run" deleted
customresourcedefinition.apiextensions.k8s.io "clustertemplates.carto.run" deleted
customresourcedefinition.apiextensions.k8s.io "deliverables.carto.run" deleted
customresourcedefinition.apiextensions.k8s.io "workloads.carto.run" deleted
deployment.apps "cartographer-controller" deleted
serviceaccount "cartographer-controller" deleted
clusterrolebinding.rbac.authorization.k8s.io "cartographer-cluster-admin" deleted
validatingwebhookconfiguration.admissionregistration.k8s.io "clustersupplychainvalidator" deleted
certificate.cert-manager.io "cartographer-webhook" deleted
issuer.cert-manager.io "selfsigned-issuer" deleted
service "cartographer-webhook" deleted
secret "cartographer-webhook" deleted
```


### Running Tests

Refer to [CONTRIBUTING.md](CONTRIBUTING.md) for instructions on running tests.


## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on the process for submitting pull requests to us.


## Code of Conduct

Refer to [CODE-OF-CONDUCT.md](CODE-OF-CONDUCT.md) for details on our code of conduct. This code of conduct applies to the Cartographer community at large (Slack, mailing lists, Twitter, etc...)


## License

Refer to [LICENSE](LICENSE) for details.

[admission webhook]: https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/
[carvel Packaging]: https://carvel.dev/kapp-controller/docs/latest/packaging/
[cert-manager]: https://github.com/jetstack/cert-manager
[kapp-controller]: https://carvel.dev/kapp-controller/
[kapp]: https://carvel.dev/kapp/
[kind]: https://github.com/kubernetes-sigs/kind
