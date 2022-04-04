# Uninstalling Cartographer

Having installed all the objects using [kapp], which keeps track of all of them as a single unit (an app), we can
uninstall everything by just referencing that name:

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

[admission webhook]: https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/
[carvel packaging]: https://carvel.dev/kapp-controller/docs/latest/packaging/
[cert-manager]: https://github.com/jetstack/cert-manager
[kapp-controller]: https://carvel.dev/kapp-controller/
[kapp]: https://carvel.dev/kapp/
[kind]: https://github.com/kubernetes-sigs/kind
