# Cartographer

<img src="assets/cartographer-logo.png">
<a href="https://bestpractices.coreinfrastructure.org/en/projects/5329"> <img src="assets/passing.svg"></a> 

![Validations](https://github.com/vmware-tanzu/cartographer/actions/workflows/validation.yaml/badge.svg?branch=main)
![Upgrade Test](https://github.com/vmware-tanzu/cartographer/actions/workflows/upgrade-test.yaml/badge.svg?branch=main)

Cartographer is a Kubernetes-native [Choreographer] providing higher modularity and scalability for the software supply chain.

[Learn more about Cartographer](https://cartographer.sh/docs/latest/)

[Choreographer]: https://tanzu.vmware.com/developer/guides/supply-chain-choreography/

## Getting Started

Examples of using Cartographer can be found in the
[examples folder of this repository](examples/README.md).
The examples begin by demonstrating how to define a Supply Chain that pulls code from a repository,
builds an image for the code, and deploys in the same cluster. Enhancements of that example
(e.g. adding tests) are then demonstrated.
## Installation

### Pre requisites
1. Administrative capabilities in a Kubernetes cluster (1.19+)
2. [cert-manager](https://cartographer.sh/docs/v0.0.7/install/) 1.5.3 installed

The quickest method to install Cartographer leverages the `cartographer.yaml` file provided with each release:

1. Create the namespace where the controller will be installed:

```bash
kubectl create namespace cartographer-system
```
2. Submit the objects included in the release:

```bash
kubectl apply -f https://github.com/vmware-tanzu/cartographer/releases/latest/download/cartographer.yaml
```
And you're done!

<img src="site/themes/template/static/img/Carto-install-yaml-v2.gif">

## Documentation

Detailed documentation for Cartographer can be found in the `site` folder of this repository:

* [About Cartographer](https://cartographer.sh/docs/latest/): details the design and philosophy of Cartographer
* [Architecture](https://cartographer.sh/docs/latest/architecture/): covers the concepts and theory of operation
* [Examples](https://github.com/vmware-tanzu/cartographer/tree/main/examples): contains a collection of examples that demonstrate how common patterns can be implemented in Cartographer 
* Spec Reference: detailed descriptions of the CRD Specs for Cartographer
  * [GVK](https://cartographer.sh/docs/latest/reference/gvk/)
  * [Workload and Supply Chains](https://cartographer.sh/docs/latest/reference/workload/)
  * [Deliverable and Delivery](https://cartographer.sh/docs/latest/reference/deliverable/)
  * [Templates](https://cartographer.sh/docs/latest/reference/template/)
  * [Runnable](https://cartographer.sh/docs/latest/reference/runnable/)

## 🤗 Community, discussion, contribution, and support

Cartographer is developed in the open and is constantly improved by our users, contributors and maintainers. It is
because of you that we are able to configure Kubernetes resources into reusable Supply Chains.

Join us!

If you have questions or want to get the latest project news, you can connect with us in the following ways:

- Chat with us in the Kubernetes [Slack](https://slack.k8s.io) in
  the [#cartographer](https://kubernetes.slack.com/archives/C02HKPSEKV1) channel
- Subscribe to the [Cartographer](https://groups.google.com/g/cartographeross) Google Group for access to discussions
  and calendars
- Join our weekly community meetings where we share the latest project news, demos, answer questions, among other
  topics:
    - Every Wednesday @ 8:00 AM PT on [Zoom](https://VMware.zoom.us/j/93284305373?pwd=UnJKL0ZaN0pqeXVMczk1WThOSUp6QT09)
    - Previous
      meetings: [[notes](https://docs.google.com/document/d/1HwsjzxpsNI0l1sVAUia4A65lhrkfSF-_XfKoZUHI120/edit?usp=sharing) | [recordings](https://www.youtube.com/playlist?list=PL7bmigfV0EqSZA5OLwrqKsAYXA1GqPtu8)]
- Do you have ideas to improve the project? Explore the [RFC process](https://github.com/vmware-tanzu/cartographer/blob/main/rfc/README.md) we currently follow and join the weekly Office Hours meeting where the maintainers team discuss RFCs  with users and contributors:
    - Every Monday @ 2:00PM ET on [Zoom](https://VMware.zoom.us/j/94592229106?pwd=eEtpekxsSERoOVNlemJWZGJTK3hvdz09)
    - Previous meetings: [[notes](https://docs.google.com/document/d/1ImIh7qBrOLOvGMCzY6AURhE-a68IE9_EbCf0g5s18vc/edit?usp=sharing) | [recordings](https://youtube.com/playlist?list=PL7bmigfV0EqSkIcCBTr3nQq04hh_EFK2a)]  
## Contributing

Pull Requests and feedback on issues are very welcome! See
the [issue tracker](https://github.com/vmware-tanzu/cartographer/issues) if you're unsure where to start, especially
the [Good first issue](https://github.com/vmware-tanzu/cartographer/labels/good%20first%20issue) label, and also feel
free to reach out to discuss.

If you are ready to jump in and test, add code, or help with documentation, please follow the instructions on
our [Contribution Guidelines](CONTRIBUTING.md) to get started and - at all times- follow
our [Code of Conduct](CODE-OF-CONDUCT.md).

## License

Apache 2.0. Refer to [LICENSE](LICENSE) for details.

[admission webhook]: https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/
[carvel Packaging]: https://carvel.dev/kapp-controller/docs/latest/packaging/
[cert-manager]: https://github.com/jetstack/cert-manager
[kapp-controller]: https://carvel.dev/kapp-controller/
[kapp]: https://carvel.dev/kapp/
[kind]: https://github.com/kubernetes-sigs/kind
