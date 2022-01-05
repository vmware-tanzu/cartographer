# Cartographer

<img src="site/themes/template/static/img/cartographer-logo.png">

Cartographer is a Kubernetes-native [Choreographer].

[Learn more about Cartographer](https://cartographer.sh/docs/latest/)

[Choreographer]: https://tanzu.vmware.com/developer/guides/supply-chain-choreography/

## Documentation

Detailed documentation for Cartographer can be found in the `site` folder of this repository:

* [About Cartographer](https://cartographer.sh/docs/latest/): Details the design and philosophy of Cartographer
* [Examples](examples/testing-sc/README.md): Contains an example of using Cartographer to create a
  Supply Chain that takes a repository, creates an image, and deploys it to a cluster
* Spec Reference: Detailed descriptions of the CRD Specs for Cartographer
  * [GVK](https://cartographer.sh/docs/latest/reference/gvk/)
  * [Workload and Supply Chains](https://cartographer.sh/docs/latest/reference/workload/)
  * [Deliverable and Delivery](https://cartographer.sh/docs/latest/reference/deliverable/)
  * [Templates](https://cartographer.sh/docs/latest/reference/template/)
  * [Runnable](https://cartographer.sh/docs/latest/reference/runnable/)

## Getting Started

Examples of using Cartographer can be found in the
[examples folder of this repository](examples/README.md).
The examples begin by demonstrating how to define a Supply Chain that pulls code from a repository,
builds an image for the code, and deploys in the same cluster. Enhancements of that example
(e.g. adding tests) are then demonstrated.

## Installation

Installation details are provided in the documentation
at [cartographer.sh/docs/install](http://cartographer.sh/docs/latest/install)

## Uninstall

Uninstallation details are provided in the documentation
at [cartographer.sh/docs/uninstall](http://cartographer.sh/docs/latest/uninstall)

### Running Tests

Refer to [CONTRIBUTING.md](CONTRIBUTING.md) for instructions on running tests.

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

## Contributing

Pull Requests and feedback on issues are very welcome! See
the [issue tracker](https://github.com/vmware-tanzu/cartographer/issues) if you're unsure where to start, especially
the [Good first issue](https://github.com/vmware-tanzu/cartographer/labels/good%20first%20issue) label, and also feel
free to reach out to discuss.

If you are ready to jump in and test, add code, or help with documentation, please follow the instructions on
our [Contribution Guidelines](CONTRIBUTING.md) to get started and - at all times- follow
our [Code of Conduct](CODE-OF-CONDUCT.md).

## License

Refer to [LICENSE](LICENSE) for details.

[admission webhook]: https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/
[carvel Packaging]: https://carvel.dev/kapp-controller/docs/latest/packaging/
[cert-manager]: https://github.com/jetstack/cert-manager
[kapp-controller]: https://carvel.dev/kapp-controller/
[kapp]: https://carvel.dev/kapp/
[kind]: https://github.com/kubernetes-sigs/kind
