# Installing Cartographer

## Pre-requisites

- Running kubernetes cluster (v1.18+)
- [cert-manager]: necessary for setting up certificates for the controller's [admission webhook]


## Steps

0. Clone the repository

```bash
kubectl apply -f https://github.com/vmware-tanzu/cartographer/releases/download/v0.0.3/release.yaml
```

1. Create the `cartographer-system` namespace

```bash
kubectl create namespace cartographer-system
```

2. Submit the Kubernetes objects that will extend Kubernetes and run the
   controller inside the cluster

you can do so either via plain `kubectl`

```bash
kubectl apply -f ./releases/release.yaml
```

or, better, with `kapp`:

```shell
$ kapp deploy -a cartographer -f ./releases/release.yaml

Target cluster 'https://127.0.0.1:53218' (nodes: kind-control-plane)

Changes

Namespace            Name                                 Kind                      Conds.  Age  Op      Op st.  Wait to    Rs  Ri
(cluster)            clusterimagetemplates.carto.run    CustomResourceDefinition  -       -    create  -       reconcile  -   -
^                    clustertemplates.carto.run         CustomResourceDefinition  -       -    create  -       reconcile  -   -
^                    clusterconfigtemplates.carto.run   CustomResourceDefinition  -       -    create  -       reconcile  -   -
^                    clustersourcetemplates.carto.run   CustomResourceDefinition  -       -    create  -       reconcile  -   -
^                    clustersupplychains.carto.run      CustomResourceDefinition  -       -    create  -       reconcile  -   -
^                    deliverables.carto.run             CustomResourceDefinition  -       -    create  -       reconcile  -   -
^                    cartographer-cluster-admin         ClusterRoleBinding        -       -    create  -       reconcile  -   -
^                    workloads.carto.run                CustomResourceDefinition  -       -    create  -       reconcile  -   -
cartographer-system  cartographer-controller            Deployment                -       -    create  -       reconcile  -   -
^                    cartographer-controller            ServiceAccount            -       -    create  -       reconcile  -   -

Op:      10 create, 0 delete, 0 update, 0 noop
Wait to: 10 reconcile, 0 delete, 0 noop

Continue? [yN]: y

11:01:49AM: ---- applying 9 changes [0/10 done] ----
...
11:01:58AM: ok: reconcile deployment/cartographer-controller (apps/v1) namespace: cartographer-system
11:01:58AM: ---- applying complete [10/10 done] ----
11:01:58AM: ---- waiting complete [10/10 done] ----

Succeeded
```

Remember: to properly install `cartographer` you must have already have
[cert-manager] installed. Below you'll find a suggestion of how you can do it,
but make sure you check their documentation:

```bash
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.2.0/cert-manager.yaml

# or, better with kapp
#
kapp deploy -a cert-manager \
    -f https://github.com/jetstack/cert-manager/releases/download/v1.2.0/cert-manager.yaml
```


## Uninstalling

With plain `kubectl`:

```bash
kubectl delete -f ./releases/release.yaml
```

with `kapp`:

```bash
kapp delete -a cartographer
```

[cert-manager]: https://github.com/jetstack/cert-manager
[admission webhook]: https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/
