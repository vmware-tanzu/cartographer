# Draft RFC 0018 Workload Report Artifact Provenance

## Summary

Workload/Deliverable status should report the artifacts currently exposed by the resources of the Supply Chain/Delivery.
Those artifacts should report the values exposed, as well as the resources from which that value has been read.

## Motivation

From [RFC 14](https://github.com/paulcwarren/cartographer/blob/rfc-0014-change-tracking/rfc/rfc-0014-change-tracking.md):

### Context

At the core of the architecture of cartographer is the concept that a supply chain is the choreography of objects and their controllers.

A supply chain, as the name suggests, chains together a set of objects defining how the (status) fields of one object feeds into the spec (and sometimes data) fields of another. Thus creating an ordered chain of interacting objects. Because controllers continuously reconcile their objects towards a desired state. A supply chain is, therefore, able to choreograph an otherwise set of independent objects (and their controllers).

A workload acts as input into this supply chain seeding initial values into an object to make the first controller do work. The supply chain manages this workload input data and outputs from each object as it propagates through the supply chain.

It is important to note however, that a workload can have multiple inputs and input into several objects in the supply chain all at the same time. As a result when a workload changes, it may cause several controllers to do work simultaneously.

It is also important to note that a controller may also do work outside of the work that the supply chain is choreographing. Kpack images, for example, are often choreographed as part of a supply chain. But the kpack controller may also do work in response to a base OS image update. Nothing to do with the supply chain but impacting it none-the-less. As a result the supply chain may be triggered part way through by this input.

### Problem

Whilst, on the one hand, the behavior described above is virtuous and generally beneficial to automated outer loop workflows. On the other hand, this behavior can be problematic for more user-centric, imperative workflows, such as those often found in the inner loop. Those of debugging and live update.

### Providing a real world example.

Given a supply chain that has several inputs, let’s say source code image and a debug flag. When a developer applies a workload after changing their source code and turning on debug. Several services in the supply chain may trigger at the same time and these inputs will traverse through the supply chain at different times. But the developer, initiating the debug session will want to wait for that specific source code change to arrive “in cluster” in an image prepped for debugging before attaching their debugger. So, it is important to know when both inputs have fully traversed the supply chain.

## Detailed Explanation

Each workload/Deliverable will include in their status an artifacts field:
```yaml
status:
  artifacts:
    # oneOf(source,image,config)
    - source:
        # the sha256 of the ordered JSON of all other non-from fields
        id: <SHA:string>
        # exposed fields - in this case url and revision
        uri: <:string>
        revision: <:string>
        # an ordered list of resources from which this artifact has been exposed
        passed:
            # name of the resource in the supply chain
            resource-name: <:string>
            # GVK of the resource
            kind: <:string>
            apiVersion: <:string>
            # name of the resource on the cluster
            name: <:string>
            # namespace of the resource
            namespace: <:string>
            # resource version of the object
            resourceVersion: <:string>
        from:
          - # id of the previous resource that was transformed into this one
            id: <SHA:string>
```

example:
Assume a supply chain that stamps out a GitRepository, Runnable, Image and config in a configmap.
Assume that a new commit was just made, so that the GitRepository has just changed url/revision
but the Runnable has not yet updated.

Previous Source output:
uri: https://www.some-site.com/my-project/my-repo
revision: b974272e27c47a01e7a7da07cf8e4415bdb83dae

Current Source output:
uri: https://www.some-site.com/my-project/my-repo
revision: b31d09004503e52e84ff633e547f4d5b40503ab3

```yaml
status:
  artifacts:
    - source:
        id: e2212e77caf0cb64a25dfa1aca39599b69d72015dbb4ed2ad740f0666af35968
        uri: https://www.some-site.com/my-project/my-repo
        revision: b31d09004503e52e84ff633e547f4d5b40503ab3
        passed:
          - resource-name: source-provider
            kind: GitRepository
            apiVersion: source.toolkit.fluxcd.io/v1beta1
            name: my-app
            namespace: my-namespace
            resourceVersion: "11125094"
    - source:
        id: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
        uri: https://www.some-site.com/my-project/my-repo
        revision: b974272e27c47a01e7a7da07cf8e4415bdb83dae
        passed:
          - resource-name: source-provider
            kind: GitRepository
            apiVersion: source.toolkit.fluxcd.io/v1beta1
            name: my-app
            namespace: my-namespace
            resourceVersion: "11125090"
          - resource-name: source-tester
            kind: Runnable
            apiVersion: carto.run/v1alpha1
            name: my-app
            namespace: my-namespace
            resourceVersion: "11125545"
    - image:
        id: 373c0dc7d3cccd8ef31cd3dcf0f07b6bcc3a9ad1270db8fe43f84e26595af32a
        image: 10.138.0.2:5000/example-testing-sc-testing-sc@sha256:9aca70a5408b7d5615724bcb8e5eea3bf0765f95eac177433993cf6002311d9b
        passed:
          - resource-name: image-builder
            kind: Image
            apiVersion: kpack.io/v1alpha2
            name: my-app
            namespace: my-namespace
            resourceVersion: "11126348"
        from:
          - id: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
    - config:
        id: 22fe69c924b55327ce3e8ca079d7295a947a3641b5ab74bdf7541dc680258c81
        config: |
            apiVersion: serving.knative.dev/v1
            kind: Service
            metadata:
              name: gitwriter-sc
              labels:
                app.kubernetes.io/part-of: hello-world
                carto.run/workload-name: gitwriter-sc
                app.kubernetes.io/component: run
            spec:
              template:
                metadata:
                  annotations:
                    autoscaling.knative.dev/minScale: '1'
                spec:
                  containers:
                    - name: workload
                      image: 10.138.0.2:5000/example-gitwriter-sc-gitwriter-sc@sha256:5bd3ff4e08015350835371444fc370f85dce8ff296dfe3af801ad6ce9771bb02
                      securityContext:
                        runAsUser: 1000
                  imagePullSecrets:
                    - name: registry-credentials
        passed:
            resource-name: config-provider
            kind: ConfigMap
            apiVersion: v1
            name: my-app
            namespace: my-namespace
            resourceVersion: "11181186"
        from:
          - id: 373c0dc7d3cccd8ef31cd3dcf0f07b6bcc3a9ad1270db8fe43f84e26595af32a
```

## Rationale and Alternatives

See [RFC 14 discussion](https://github.com/vmware-tanzu/cartographer/pull/274)

## Implementation

As cartographer iterates over the resources of the supply chain and reads values, it will save those values for writing
to the status of the workload. This can be accomplished through the use of an `artifact-manager` patterned from the
current implementation of the `condition-manager`.
