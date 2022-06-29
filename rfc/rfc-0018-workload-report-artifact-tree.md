# Draft RFC 0018 Workload Report Artifact Provenance

## Summary

Workload/Deliverable status should report the artifacts currently exposed by the resources of the Supply Chain/Delivery.
Those artifacts should report the values exposed, as well as the resources from which that value has been read.

## Motivation

From [RFC 14](https://github.com/paulcwarren/cartographer/blob/rfc-0014-change-tracking/rfc/rfc-0014-change-tracking.md):

### Context

At the core of the architecture of cartographer is the concept that a supply chain is the choreography of objects and
their controllers.

A supply chain, as the name suggests, chains together a set of objects defining how the (status) fields of one object
feeds into the spec (and sometimes data) fields of another. Thus creating an ordered chain of interacting objects.
Because controllers continuously reconcile their objects towards a desired state. A supply chain is, therefore, able to
choreograph an otherwise set of independent objects (and their controllers).

A workload acts as input into this supply chain seeding initial values into an object to make the first controller do
work. The supply chain manages this workload input data and outputs from each object as it propagates through the
supply chain.

It is important to note however, that a workload can have multiple inputs and input into several objects in the supply
chain all at the same time. As a result when a workload changes, it may cause several controllers to do work
simultaneously.

It is also important to note that a controller may also do work outside of the work that the supply chain is
choreographing. Kpack images, for example, are often choreographed as part of a supply chain. But the kpack controller
may also do work in response to a base OS image update. Nothing to do with the supply chain but impacting it
none-the-less. As a result the supply chain may be triggered part way through by this input.

### Problem

Whilst, on the one hand, the behavior described above is virtuous and generally beneficial to automated outer loop
workflows. On the other hand, this behavior can be problematic for more user-centric, imperative workflows, such as
those often found in the inner loop. Those of debugging and live update.

### Providing a real world example.

Given a supply chain that has several inputs, let’s say source code image and a debug flag. When a developer applies a
workload after changing their source code and turning on debug. Several services in the supply chain may trigger at the
same time and these inputs will traverse through the supply chain at different times. But the developer, initiating the
debug session will want to wait for that specific source code change to arrive “in cluster” in an image prepped for
debugging before attaching their debugger. So, it is important to know when both inputs have fully traversed the supply
chain.

## Detailed Explanation

Each workload/Deliverable will include in their status an artifacts field:
```yaml
status:
  artifacts:
    # oneOf(source,image,config, object)
    - source:
        # the sha256 of the ordered JSON of all other non-from fields
        id: <SHA:string>
        # exposed fields - in this case url and revision
        uri: <:string>
        revision: <:string>
        # the object which produced this artifact
        resource:
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
          - # id of any artifact(s) that were inputs to the template
            id: <SHA:string>
```

example:
We'll consider a supply chain that stamps out a GitRepository, Runnable (to test the source code), Image, config in a
configmap, and finally a Kapp App. We'll assume that a new commit was just made, so that the GitRepository has
just changed url/revision but the Runnable has not yet updated.

Previous Source output:
uri: https://www.some-site.com/my-project/my-repo
revision: b974272e27c47a01e7a7da07cf8e4415bdb83dae

Current Source output:
uri: https://www.some-site.com/my-project/my-repo
revision: b31d09004503e52e84ff633e547f4d5b40503ab3

```yaml
status:
  artifacts:
    - source:   # <--- new revision output by the GitRepository object
        id: 146c7d74eb956191487236b579e8e4e68462fc7d97c4f1a4677b0ded39e2a3ca
        uri: https://www.some-site.com/my-project/my-repo
        revision: b31d09004503e52e84ff633e547f4d5b40503ab3
        resource:
          resource-name: source-provider
          kind: GitRepository
          apiVersion: source.toolkit.fluxcd.io/v1beta1
          name: my-app
          namespace: my-namespace
          resourceVersion: "11125094"
    - source:   # <--- old revision from GitRepository. This source is still reported as another artifact still relies on it
        id: 23156c7ac2170fe95f85a1ad42522c408e67038f8393eea6cd8551e07457c5d7 # <--- the sha256sum of the passed, revision and uri fields
        uri: https://www.some-site.com/my-project/my-repo
        revision: b974272e27c47a01e7a7da07cf8e4415bdb83dae
        resource:
          resource-name: source-provider
          kind: GitRepository
          apiVersion: source.toolkit.fluxcd.io/v1beta1
          name: my-app
          namespace: my-namespace
          resourceVersion: "11125090"
    - source:   # <--- source artifact from the Runnable
        id: 74cb6607d64da1e0324196517340ac7a668042f0e9dfbdd58f8a00a5d0ee9580
        uri: https://www.some-site.com/my-project/my-repo
        revision: b974272e27c47a01e7a7da07cf8e4415bdb83dae
        resource:
          resource-name: source-tester
          kind: Runnable
          apiVersion: carto.run/v1alpha1
          name: my-app
          namespace: my-namespace
          resourceVersion: "11125545"
        from:
          - id: 23156c7ac2170fe95f85a1ad42522c408e67038f8393eea6cd8551e07457c5d7 # <--- The id noted above, as the source-tester consumes the source artifact from source-provider
    - image:
        id: bde9bc48c3d7a9dd20e94e138422264efca9770c9beb32ac18227eb204315a4a
        image: 10.138.0.2:5000/example-testing-sc-testing-sc@sha256:9aca70a5408b7d5615724bcb8e5eea3bf0765f95eac177433993cf6002311d9b
        resource:
          resource-name: image-builder
          kind: Image
          apiVersion: kpack.io/v1alpha2
          name: my-app
          namespace: my-namespace
          resourceVersion: "11126348"
        from:
          - id: 74cb6607d64da1e0324196517340ac7a668042f0e9dfbdd58f8a00a5d0ee9580
    - config:
        id: 6e0c84490ecb71a4c5a970383c05a520f4c94546033f48a81c06472fe61f1aad
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
        resource:
          resource-name: config-provider
          kind: ConfigMap
          apiVersion: v1
          name: my-app
          namespace: my-namespace
          resourceVersion: "11181186"
        from:
          - id: bde9bc48c3d7a9dd20e94e138422264efca9770c9beb32ac18227eb204315a4a
    - object:
        id: a03aed19284140c8093fe65a43cb1df5d16ecc12874d76aee63e4da4d7855436
        resource:
          resource-name: app-deploy
          kind: App
          apiVersion: kappctrl.k14s.io/v1alpha1
          name: my-app
          namespace: my-namespace
          resourceVersion: "21212378"
        from:
          - id: 6e0c84490ecb71a4c5a970383c05a520f4c94546033f48a81c06472fe61f1aad
```

## Rationale and Alternatives

See [RFC 14 discussion](https://github.com/vmware-tanzu/cartographer/pull/274)

## Implementation

As cartographer iterates over the resources of the supply chain and reads values, it will save those values for writing
to the status of the workload. This can be accomplished through the use of an `artifact-manager` patterned from the
current implementation of the `condition-manager`. An example spike on this can be found
[here](https://github.com/vmware-tanzu/cartographer/tree/waciuma/spike-rfc-18).

[RFC 20](https://github.com/vmware-tanzu/cartographer/pull/556) is a prerequisite to achieve the `from` field behavior
specified here.
