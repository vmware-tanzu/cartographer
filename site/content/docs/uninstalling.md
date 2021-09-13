## Uninstalling

With plain `kubectl`:

```bash
kubectl delete -f ./releases/release.yaml
```

with `kapp`:

```bash
kapp delete -a kontinue
```

[cert-manager]: https://github.com/jetstack/cert-manager
[admission webhook]: https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/