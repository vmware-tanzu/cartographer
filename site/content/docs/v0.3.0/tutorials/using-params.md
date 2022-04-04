# Using Params

## Overview

In this tutorial we’ll explore how to use params to pass information that isn’t anticipated by the standard workload
fields. We’ll see how params can either be set or delegated to other objects in the supply chain.

## Environment setup

For this tutorial you will need a kubernetes cluster with Cartographer installed. You may follow
[the installation instructions here](https://github.com/vmware-tanzu/cartographer#installation).

Alternatively, you may choose to use the
[./hack/setup.sh](https://github.com/vmware-tanzu/cartographer/blob/main/hack/setup.sh) script to install a kind cluster
with Cartographer. _This script is meant for our end-to-end testing and while we rely on it working in that role, no
user guarantees are made about the script._

Command to run from the Cartographer directory:

```shell
$ ./hack/setup.sh cluster cartographer-latest
```

If you later wish to tear down this generated cluster, run

```shell
$ ./hack/setup.sh teardown
```

## Scenario

As in the tutorial ["Build Your First Supply Chain"](first-supply-chain.md), we will act as both the app dev and app
operator in a company creating hello world web applications. Our applications will again have been built and stored in a
registry. But now each image will be in some private registry and each app dev will need to provide the appropriate
service account with imagePullCredentials for our deployment. We will assume that the app devs have already created the
necessary service account in their namespace. We will see how to write our templates to accept a parameter with the
service account name as well as how to write a workload to supply that value.

## Steps

### App Operator Steps

A new field must be added to our template of a deployment:

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
          serviceAccountName: # <=== NEW FIELD
          containers:
            - name: $(workload.metadata.name)$
              image: $(workload.spec.image)$
```

Let’s assume that the app operator has created an image registry in which they have expansive read credentials. The
operator can reasonably expect that many devs will use this registry. So the operator can be responsible for the
creation of a service account with the correct imagePullCredentials and can make sure this object is in the expected
namespaces with the expected name. Let’s set this expected name as "expected-service-account".

While this will be the default value, we will allow the developer to override this choice. We’ll do this by setting the
name as a param for the ClusterTemplate:

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterTemplate
metadata:
  ...
spec:
  template:
    ...
    spec:
      ...
      template:
        ...
        spec:
          serviceAccountName: $(params.image-pull-sa-name)$
          ...
  params:
    - name: image-pull-sa-name
      default: expected-service-account
```

Here we see two changes to the ClusterTemplate’s spec:

- The first is a field in the template. The field has been filled with a reference to a param:
  `$(params.image-pull-sa-name)$`
- Second, we see that a new field has been introduced to the spec: `params`. This is a list of params, each of which
  requires a name and a default value. Here, the app operator indicates that many devs will have a service account named
  `expected-service-account` in their namespace.

We can see here the full object we'll apply to the cluster:

{{< tutorial app-deploy-template.yaml >}}

We apply this template to the cluster along with the same supply chain from the
["Build Your First Supply Chain"](first-supply-chain.md) tutorial.

### App Dev Steps

As app devs, we’ve decided not to follow the app operator service account naming convention. We’ve created a service
account named "unconventionally-named-service-account" which has the imagePullSecrets to get our app image. In the
workload, we'll create a param. Params in workloads require two fields, a name and a value. We must give our param the
same name as the param in the template, `image-pull-sa-name`. And for the value, we'll provide our chosen service
account's name.

```yaml
spec:
  params:
    - name: image-pull-sa-name
      value: unconventionally-named-service-account
```

We can now apply this workload to the cluster:

{{< tutorial  workload.yaml >}}

## Observe

When we observe the created deployment, we can see that the value specified by the workload is present:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-again-deployment
  ...
spec:
  template:
    metadata: ...
    spec:
      Containers: ...
      dnsPolicy: ...
      restartPolicy: ...
      schedulerName: ...
      securityContext: {}
      serviceAccount: unconventionally-named-service-account     # <=== huzzah!
      serviceAccountName: unconventionally-named-service-account
      terminationGracePeriodSeconds: 30
  ...
status: ...
```

## Further information

Params are created to allow delegation between different personas. The individual writing a template may be different
from the person writing the supply chain. Perhaps the template author is unsure of a parameter’s value, but the supply
chain author knows exactly the desired value and does not want the workload author to be able to override their choice.
The supply chain can be altered to specify a `value`.

```yaml
---
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
      params: # <=== New field
        - name: image-pull-sa-name
          value: inevitable-service-account
```

If the supply chain is redeployed with this definition, we can observe that the deployment changes:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-deployment
  ...
spec:
  template:
    metadata: ...
    spec:
      Containers: ...
      dnsPolicy: ...
      restartPolicy: ...
      schedulerName: ...
      securityContext: {}
      serviceAccount: inevitable-service-account
      serviceAccountName: inevitable-service-account
      terminationGracePeriodSeconds: 30
  ...
status: ...
```

Further information about params and the order of precedence can be found in the
["Parameter Heirarchy"](../architecture/#parameter-hierarchy) architecture documentation.

In the Cartographer tests you can find an example of creating an object with numerous parameters which demonstrates the
precedence rules:

- [A template with many param fields, providing default values for some](https://github.com/vmware-tanzu/cartographer/tree/main/tests/kuttl/supplychain/params-supply-chain/00-proper-templates.yaml)
- [A supply chain providing some defaults and some values for params](https://github.com/vmware-tanzu/cartographer/tree/main/tests/kuttl/supplychain/params-supply-chain/01-supply-chain.yaml)
- [A workload providing some values for params](https://github.com/vmware-tanzu/cartographer/tree/main/tests/kuttl/supplychain/params-supply-chain/02-workload.yaml)
- [The expected object which will be created when that trio is submitted to the cluster](https://github.com/vmware-tanzu/cartographer/tree/main/tests/kuttl/supplychain/params-supply-chain/02-assert.yaml)

## Wrap Up

Congratulations, you’ve used parameters in your supply chain! You’ve learned:

- How a template requests information from a workload not available in the standard workload fields
- How to provide default values for params
- How the supply chain can provide mandatory values for params
- How to find more information on params
