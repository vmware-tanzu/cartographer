# Draft RFC 0007 Support Service Definitions

## Summary

* Allow Operators to specify how service definitions become service configurations in the final configuration.
* Allow developers to specify service definitions in their workload

## Motivation

App Developers need to specify services for their applications. This RFC focuses on supporting [ServiceBindings][bindings]
It should work just as well with any 'Object Reference' that's satisfied by:
```yaml
apiVersion: <Group/Version>
kind: <Kind>
name: <name>
namespace: <namespace> #optional
```

## Possible Solutions

Using the existing `workload.spec.services` specification, rely on `kapp-controller`'s inbuilt capabilities to
apply ytt templates to provide an example for supply chains that support services.

### Example
Here is an example using the [vmware service binding implementation][vmware-bindings] with a `bindings.labs.vmware.com/ProvisionedService`

We've omitted the boilerplate, the full example is under [the rfc-0007 subdirectory](./rfc-0007)

#### provisioned-service.yaml
```yaml
apiVersion: bindings.labs.vmware.com/v1alpha1
kind: ProvisionedService
metadata:
  name: provisioned-service
spec:
  binding:
    name: production-db-secret

---
apiVersion: v1
kind: Secret
metadata:
  name: production-db-secret
type: service.binding/mysql
stringData:
  type: mysql
  provider: bitnami
  host: localhost
  port: "3306"
  username: root
  password: root
```

#### config-template.yaml
As shown here, we leverage `kappctrl.k14s.io/App`'s support for ytt templating.

We verbatim copy the  workload's `spec.services` object into `values.yml`, then
`ytt` enumerates these in the `service-bindings.yml` file. This binding is applied
alongside the app, connecting the app to the provisioned-service.

Please see [supply-chain-templates.yaml for the full file](./rfc-0007/app-operator/supply-chain-templates.yaml)

```yaml
---
apiVersion: kontinue.io/v1alpha1
kind: ClusterTemplate
metadata:
  name: app-deploy
spec:
  template:
    apiVersion: kappctrl.k14s.io/v1alpha1
    kind: App
    metadata:
      name: $(workload.metadata.name)$
    spec:
      serviceAccountName: default
      fetch:
        - inline:
            paths:
              values.yml: |
                #@data/values
                ---
                services: $(workload.spec.services)$

              service-bindings.yml: |
                #@ load("@ytt:data", "data")

                #@ for/end service in data.values.services:
                ---
                apiVersion: service.binding/v1alpha2
                kind: ServiceBinding
                metadata:
                  name: provisioned-service
                spec:
                  workload:
                    apiVersion: serving.knative.dev/v1
                    kind: Service
                    name: $(workload.metadata.name)$
                  service: #@ service

              manifest.yml: # App Manifest here
      template:
        - ytt: {}
      deploy:
        - kapp: {}
```

#### workload.yaml
```yaml
apiVersion: kontinue.io/v1alpha1
kind: Workload
metadata:
  name: dev
  labels:
    app.tanzu.vmware.com/workload-type: web
spec:
  source:
    git:
      url: https://github.com/kontinue/hello-world
      ref:
        branch: main
  services:
    - apiVersion: bindings.labs.vmware.com/v1alpha1
      kind: ProvisionedService
      name: provisioned-service
      namespace: default
```

## Cross References and Prior Art

* [Specification for ServiceBinding][bindings]

[bindings]: https://github.com/k8s-service-bindings/spec/blob/28537bac6da0a90512b806b9eded4cb690ef7dd6/README.md#terminology-definition
[vmware-bindings]: https://github.com/vmware-labs/service-bindings
