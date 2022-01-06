# Draft RFC 0017 Workload Conventions

## Summary

Cartographer supply chains choreograph multiple components on behalf of workloads. One of these components, the `ClusterConfigTemplate`, is able to enrich the resources that are ultimately delivered. Cartographer should provide hooks for a higher level of choreography for that enrichment out of the box, while still allowing users to skip or plug in their own solutions.

As part of this proposal, only the conventions APIs and controller is covered. A new RFC will be required for individual conventions to be added to the project, aside from demonstrative samples.

## Motivation

Kubernetes workloads often need to instruct the platform as to how the workload should run. Kubernetes is an application agnostic platform that will do what you ask of it, even if that is suboptimal. The `Workload` resource offered by Cartographer exposes the core essentials of a workload to be defined by an application developer, but there are many aspects of a well behaved workload at runtime that are not exposed. These settings are inherently specific to a class of workloads, and are not one-size fits all.

Workload conventions are means for operators to distill their knowledge of how to best run specific types of workloads on Kubernetes without forcing each application developer to become an expert in each knob that Kubernetes provides.

For example, liveness and readiness probes are how workloads signal to Kubernetes that it is alive and able to receive network traffic, respectively. By default Kubernetes will watch the command process in the container, so long as it is running, the pod is healthy and alive. If a workload is an http microservice, this is not a good measure of health, nor readiness. No web server is able to start immediately, meaning there could be user requests sent to the container before it is listening. It's also possible that the process is deadlocked, a zombie still "alive", but unable to receive requests. Defining HTTP based probes for liveness and readiness provide a more accurate indication of health. However, Kubernetes doesn't know what port to connect to, or which path to use, or if the workload is using HTTPS. These settings require deeper knowledge of both Kubernetes *and* the workload to be set correctly. This type of deep knowledge can be defined by conventions and selectively applied when appropriate.

## Possible Solutions

Tanzu Application Platform has a component named the Convention Service that we would like to open source as part of Cartographer. The Convention Service defines two CRDs: `PodIntent` and `ClusterPodConvention`.

At a high level:

The `PodIntent` resource is created by the supply chain's config template. It includes the intended `PodTemplateSpec` based on the definition of the supply chain, the `Workload` resource and outputs from other components in the supply chain (like a built image). OCI metadata and SBOMs for each image referenced in the `PodTemplateSpec` is resolved from the image registry. The controller then passes the `PodTemplateSpec` and image metadata to each `ClusterPodConvention` to be enriched and reflects the enriched PodTemplateSpec on the `PodIntent`'s status.

Each `ClusterPodConvention` has the opportunity to enrich the `PodTemplateSpec` defined by the `PodIntent`. In many cases, a convention will do nothing if the specific convention does not have an opinion about the `PodTemplateSpec` based on the content or image metadata. For example, a Spring Boot convention can use an SBOM detect that an image contains a Spring Boot working and then apply Spring best practices based on the content of the image and the existing setting on the `PodTemplateSpec`, all while ignoring a Rails workload it doesn't understand. (a separate Rail convention can apply knowledge of Rails workload)

Individual conventions are implemented as a webhook that is registered with the system by the `ClusterPodConvention` resource.

Operators can install conventions into the cluster to add new capabilities similar to how new buildpacks can be exposed within kpack. Unlike a Workload which is processed by exactly one `ClusterSupplyChain`, the `PodIntent` is advised by all `ClusterPodConvention`s installed on the cluster. This enables cross cutting behavior to be defined once by an operator and applied to all instances.

Other types of conventions beyond pod intents could be introduced in the future.

## Cross References and Prior Art

The Tanzu Application Platform's [Convention Service](https://docs.vmware.com/en/VMware-Tanzu-Application-Platform/0.4/tap/GUID-convention-service-about.html) is the inspiration for this RFC.
