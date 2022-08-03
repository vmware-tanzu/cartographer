# blueprints proof of concept

It's all kubebuilder stuff except:


Run ciro's samples (as they're pulled over to `tests/kuttl/demonstration/ciros`):
```shell
make test-demonstration
```

deploy to a kind cluster easily (with restart):

```shell
make quick-deploy
```

Sometimes you need to run `quick-deploy` twice as it doesn't wait for the cert-manager install to complete.