# Cartographer
<img src="site/themes/template/static/img/cartographer-logo.png">

Cartographer is a Kubernetes native [Choreographer]. It allows users to configure K8s resources into re-usable [_Supply Chains_](site/content/docs/reference.md#ClusterSupplyChain) that can be used to define all of the stages that an [_Application Workload_](site/content/docs/reference.md#Workload) must go through to get to an environment.

[Choreographer]: https://tanzu.vmware.com/developer/guides/supply-chain-choreography/

Cartographer also allows for separation of controls between a user who is responsible for defining a Supply Chain (known as a App Operator) and a user who is focused on creating applications (Developer). These roles are not necessarily mutually exclusive, but provide the ability to create a separation concern.

## Known Issues
- **WARNING!!** At this time, the Supply Chain ClusterRoleBinding has more permissions than it needs. This will be fixed in an upcoming release.
The issue can be tracked [here](https://github.com/vmware-tanzu/cartographer/issues/51).

## Documentation

Detailed documentation for Cartographer can be found in the `site` folder of this repository:

* [About Cartographer](site/content/docs/about.md): Details the design and philosophy of Cartographer
* [Examples](examples/source-to-gitops/README.md): Contains an example of using Cartographer to create a supply chain that takes a repository, creates an image and deploys it to a cluster
* [Spec Reference](site/content/docs/reference.md): Detailed descriptions of the CRD Specs for Cartographer

## Getting Started

An example of using Cartographer to define a Supply Chain that pulls code from a repository, builds an image for the code and deploys it to the same cluster can be found in the [examples folder of this repository](examples/source-to-gitops/README.md)


## Installation

Installation details are provided in the documentation at [cartographer.sh/docs/install](http://cartographer.sh/docs/install)


## Uninstall

Uninstallation details are provided in the documentation at [cartographer.sh/docs/uninstall](http://cartographer.sh/docs/uninstall)


### Running Tests

Refer to [CONTRIBUTING.md](CONTRIBUTING.md) for instructions on running tests.


## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on the process for submitting pull requests to us.


## Code of Conduct

Refer to [CODE-OF-CONDUCT.md](CODE-OF-CONDUCT.md) for details on our code of conduct. This code of conduct applies to the Cartographer community at large (Slack, mailing lists, Twitter, etc...)


## License

Refer to [LICENSE](LICENSE) for details.

[admission webhook]: https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/
[carvel Packaging]: https://carvel.dev/kapp-controller/docs/latest/packaging/
[cert-manager]: https://github.com/jetstack/cert-manager
[kapp-controller]: https://carvel.dev/kapp-controller/
[kapp]: https://carvel.dev/kapp/
[kind]: https://github.com/kubernetes-sigs/kind
