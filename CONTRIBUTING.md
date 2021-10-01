# Contributing

## CLA

The Cartographer project team welcomes contributions from the community. If you wish to contribute code and you have not signed our [contributor license agreement](https://cla.vmware.com/cla/1/preview), our bot will update the issue when you open a Pull Request. For any questions about the CLA process, please refer to our [FAQ](https://cla.vmware.com/faq).


## Development Dependencies

- [`ctlptl`]: for deploying local changes to a local registry
- [`go`]: for compiling the controllers as well as other dependencies - 1.17+
- [`kapp`]: for managing groups of kubernetes objects in a cluster (like our CRDs etc)
- [`kbld`]: for resolving image references to absolute ones
- [`kind`]: to run a local cluster
- [`ko`]: for building and pushing the controller's container image
- [`kuttl`]: for integration testing
- [`shellcheck`]: for shell script linting
  [`ytt`]: for Cluster Templates with `ytt:` sections

[`ctlptl`]: https://github.com/tilt-dev/ctlptl
[`go`]: https://golang.org/dl/
[`kapp`]: https://github.com/vmware-tanzu/carvel-kapp
[`kbld`]: https://github.com/vmware-tanzu/carvel-kbld
[`kind`]: https://kind.sigs.k8s.io/docs/user/quick-start/
[`ko`]: https://github.com/google/ko
[`kuttl`]: https://github.com/kudobuilder/kuttl
[`shellcheck`]: https://github.com/koalaman/shellcheck
[`ytt`]: https://github.com/vmware-tanzu/carvel-ytt


## Contribution workflow

This section describes the process for contributing a bug fix or new feature.

### Before you submit a pull request

This project operates according to the _talk, then code_ rule.
If you plan to submit a pull request for anything more than a typo or obvious bug fix, first you _should_ [raise an issue][new-issue] to discuss your proposal, before submitting any code.

Depending on the size of the feature you may be expected to first write an RFC. Follow the [RFC Process](https://github.com/vmware-tanzu/cartographer/blob/main/rfc/README.md) documented in Cartographer's Governance.

### Commit message and PR guidelines

- Have a short subject on the first line and a body.
- Use the imperative mood (ie "If applied, this commit will (subject)" should make sense).
- Put a summary of the main area affected by the commit at the start,
  with a colon as delimiter. For example 'docs:', 'extensions/(extensionname):', 'design:' or something similar.
- Do not merge commits that don't relate to the affected issue (e.g. "Updating from PR comments", etc). Should
  the need to cherrypick a commit or rollback arise, it should be clear what a specific commit's purpose is.
- If the main branch has moved on, you'll need to rebase before we can merge,
  so merging upstream main or rebasing from upstream before opening your
  PR will probably save you some time.

Pull requests *must* include a `Fixes #NNNN` or `Updates #NNNN` comment.
Remember that `Fixes` will close the associated issue, and `Updates` will link the PR to it.

#### Sample commit message

```text
extensions/extenzi: Add quux functions

To implement the quux functions from #xxyyz, we need to
flottivate the crosp, then ensure that the orping is
appred.

Fixes #xxyyz

Signed-off-by: Your Name <you@youremail.com>
```

### Merging commits

Pull requests with multiple commits can be merged with the [Create a merge commit](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/about-pull-request-merges) option.
Merging pull requests with multiple commits can make sense in cases where a change involves code generation or mechanical changes that can be cleanly separated from semantic changes.

Maintainers should feel free to request authors to squash their branches. Maintainers should default to this if the branch contains WIP commits.

### PR CI Actions

Before a PR is accepted it will need to pass the validation checks. Authors can find a reasonable degree of of surety that they will pass by running `make test` though this is only a subset of the checks that validation undertakes (for example, validation does an e2e test).

---

## Running a local cluster

A local Kubernetes cluster with a local registry and Cartographer installed can
be stood up with the [hack/setup.sh](./hack/setup.sh):


```bash
# here we're performing a few actions, one after another:
#
# - bringing the cluster up with a local registry already trusted
# - installing cert-manager, a cartographer's dependency
# - installing cartographer from the single-file release
#
./hack/setup.sh cluster cert-manager cartographer
```

This cluster will use a local registry. The controller running in the cluster will
demonstrate the behavior of the codebase when the deploy command was run (Devs can
check expected behavior by rerunning the deploy command).


## Running the tests

### Unit tests

Nothing else required aside from doing the equivalent of `go test ./...`:

```
make test
```


### Integration tests

Integration tests involve a Kubernetes API server, and a persistence service (etcd).

There are two sets of Integration tests:
1. Those that are purely declarative, and run using [kuttl](https://github.com/kudobuilder/kuttl)
   ```
   make test-kuttl
   or
   make test-kuttl-kind  # see below section about testing with a full cluster.
   ```

2. Those that require asynchronous testing, run using [ginkgo](https://onsi.github.io/ginkgo/).
   ```
   make test-integration
   ```


### Running integration tests without a complete cluster

For speed, both `kuttl` and the ginkgo tests can (and usually do) use
[`envtest` from controller-runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest)

The `envtest` package runs a simplified control plane, using `kubernetes-apiserver` and `etcd`. You'll need `etcd`
and `kube-apiserver` installed in `/usr/local/kubebuilder/bin`
and can download them with:

```
K8S_VERSION=1.19.2
curl -sSLo envtest-bins.tar.gz "https://storage.googleapis.com/kubebuilder-tools/kubebuilder-tools-${K8S_VERSION}-$(go env GOOS)-$(go env GOARCH).tar.gz"
```

**Note:** `envTest` cannot run pods, so is limited, but is typically all that's needed to test controllers and webhooks.


### Running integration tests with a complete cluster

Declarative `kuttl` tests can be run against a real cluster, using `kind`. This approach is slower but can be useful
sometimes for easy debugging with the `kubectl` command line tool.
```
make test-kuttl-kind  # see below section about testing with a full cluster.
```


## Merge Request Flow

Often, code in a merge request will be worked on by a pair from the team maintaining Catographer.
In this case, a merge request should be made by one engineer. The second engineer should approve
the merge request. If the work is not blocking other stories, the merge request should stay open
overnight, to allow others on the team time to read. The following morning, a merge should be
completed.


## Maintaining a useful commit history

The commit history should be legible and (to our greatest ability) logical. In pursuit of this:

1. Use small commits. Keeping logical work in their own commit helps document code.

1. Remove WIP commits from a branch before merging it into another. If a WIP commit is made at the
end of a day, a soft reset the following morning can help ensure that only logical commits remain
in the branch.

1. When merging, do not fast forward. E.g. use `git merge --no-ff` This will make clear the small
commits that belong to a logical group.

1. When merging Branch A into Branch B, perform a rebase of Branch A on Branch B. This will
ensure that the commits of Branch A are easily readable when reading Branch B's history.


## Creating new releases

A release of Cartographer consists of:

1. a tagget commit in the Git history
2. a GitHub release aiming at that tag making available a few assets that make
   Cartographer installable in a cluster

While the second step is manual (someone needs to `git tag` and then `git
push`), the second isn't: the [release.yaml](.github/workflows/release.yaml)
GitHub workflow takes care of preparing the assets and pushing to GitHub all
the necessary bits.

Although `2` is automated, it's still possible to do the procedure manually.


### Manually building and pushing release assets to GitHub


0. a container image registry that we can temporarily push images to

```bash
# `cluster` brings up a local container image registry and a kubernetes
# cluster that we can use to push images to and try the installation
# out (this cluster already trusts the local registry).
#
./hack/setup.sh cluster
```


1. tag the current commit for release

```bash
git tag v0.0.6-rc
```


2. generate the release assets

```bash
RELEASE_VERSION=v0.0.6-rc ./hack/release.sh
```


3. submit to github

```bash
RELEASE_VERSION=v0.0.6-rc ./hack/release.sh
```


## Running the end-to-end test

Cartographer has a  script that allows users to:

- create a local cluster with a local repository
- push the image of the controller to that cluster
- run the controller
- create the supply chain(s) and workload(s) the example directory
- assure that the expected objects are created by the Cartographer controller

To run the tests, first make sure you have Docker running and no Kubernetes
clusters already there, then proceed with the use of the
[setup.sh](./hack/setup.sh) script:


```bash
# 1. create a local cluster using kind as well as a
#    local container registry that is already
#    trusted by the cluster, including
#    pre-requisite dependencies like kapp-ctrl and
#    cert-manager.
#
# 2. produce a new local release of cartographer and
#    have it installed in the local cluster
#
# 3. install all the dependencies that are used by
#    the examples
#
# 4. submit the example and wait for the final
#    deployment (knative service with our app) to
#    occur.
#
./hack/setup.sh cluster cartographer example-dependencies example
```

Once the execution finished, you can either play around with the environment or
tear it down.

```bash
# get rid of all of the containers created to
# support the containers
#
./hack/setup.sh teardown
```

ps.: those commands can be all specified at once, for instance:

[carvel Packaging]: https://carvel.dev/kapp-controller/docs/latest/packaging/
[imgpkg bundle]: https://carvel.dev/imgpkg/docs/latest/
 