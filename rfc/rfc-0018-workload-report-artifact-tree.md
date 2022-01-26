# Draft RFC 0018 Workload Report Artifact Provenance

## Summary

Workload/Deliverable status should report the artifacts currently exposed by the resources of the Supply Chain/Delivery.
Those artifacts should the values exposed, as well as the resources from which that value has been read.

## Motivation

See RFC 14

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

See RFC 14 discussion

## Implementation

As cartographer iterates over the resources of the supply chain and reads values, it will save those values for writing
to the status of the workload. This can be accomplished through the use of an `artifact-manager` patterned from the
current implementation of the `condition-manager`.
