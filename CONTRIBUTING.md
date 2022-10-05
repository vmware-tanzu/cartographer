# Contributing

## CLA

The Cartographer project team welcomes contributions from the community. If you wish to contribute code and you have not signed our [contributor license agreement](https://cla.vmware.com/cla/1/preview), our bot will update the issue when you open a Pull Request. For any questions about the CLA process, please refer to our [FAQ](https://cla.vmware.com/faq).

## Development Dependencies

- [`ctlptl`]: for deploying local changes to a local registry
- [`go`]: for compiling the controllers as well as other dependencies - 1.18+
- [`kapp`]: for managing groups of kubernetes objects in a cluster (like our CRDs etc)
- [`kbld`]: for resolving image references to absolute ones
- [`kind`]: to run a local cluster
- [`ko`]: for building and pushing the controller's container image
- [`kuttl`]: for integration testing. *MUST* use version >1.13
- [`shellcheck`]: for shell script linting
- [`ytt`]: for Cluster Templates with `ytt:` sections
- [`woke`]: for detecting non-inclusive terminology 

[`ctlptl`]: https://github.com/tilt-dev/ctlptl
[`go`]: https://golang.org/dl/
[`kapp`]: https://github.com/vmware-tanzu/carvel-kapp
[`kbld`]: https://github.com/vmware-tanzu/carvel-kbld
[`kind`]: https://kind.sigs.k8s.io/docs/user/quick-start/
[`ko`]: https://github.com/google/ko
[`kuttl`]: https://github.com/kudobuilder/kuttl
[`shellcheck`]: https://github.com/koalaman/shellcheck
[`ytt`]: https://github.com/vmware-tanzu/carvel-ytt
[`woke`]: https://docs.getwoke.tech/installation/#build-from-source

## Error handling and logging
### Best practices
Do not use `fmt.Sprintf` in log messages. Use key-value pairs.
See [Logging messages](https://github.com/kubernetes-sigs/controller-runtime/blob/master/TMP-LOGGING.md#logging-messages).

When logging kubernetes objects, you can log the object as a value. Only name, namespace, apiVersion and kind will be printed.
See [Logging Kubernetes Objects](https://github.com/kubernetes-sigs/controller-runtime/blob/master/TMP-LOGGING.md#logging-kubernetes-objects).

### Rules of thumb
- When an error occurs, think about what debug will improve our ability to debug issues related to the error.
- Do not blindly add logs at every call site.
- With values that will provide context in higher abstractions, add them as soon as you can, for example "supplychain"
  in the workload reconciler. Do not add them if they're only going to be useful locally (eg stampedObject in 
  our reconcilers will mean nothing in callee's).
- Place context on errors and describe the error in 'local' terms. Eg: for a get to the api server: "failed to get
supply-chain from api server" but for the same error in the reconciler "failed to get supply-chain". There will still 
be a lot of repetition, but the changing context gives the duplicate messages a reason to exist.

### Error types
There are many kinds of 'error':
- an exception: "This code should not be reached, but no one likes a panic". Use: Log.Error and return an unhandled error 
(retry-with-backoff)
- a recoverable error: async/external comms usually. Use: Log.Error and return an unhandled erorr (retry-with-backoff)
- a message that we consider recoverable by user action: "I tried to stamp but couldn't" (often these are deeper user validations). 
Use: Status.Conditions (primary form of user comms), Log.Info, and return nil or a handled error (will not retry-with-backoff)
- a message that we consider recoverable with a retry: "something is not ready yet". Use: Status.Conditions, Log.Info and return
nil or a handled error (will not retry-with-backoff)

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

Before a PR is accepted it will need to pass the validation checks.

Validations should pass if you can run `make test pre-push` without failing.

If you run `make pre-push` and it fails, it usually autocorrects lints and generable files, so `git add` the changes,
commit them and run `make pre-push` again.

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
if [[ "$(go env GOARCH)" == "arm64" ]] && [[ "$(go env GOOS)" == "darwin" ]]; then ENV_TEST_ARCH="amd64"; else ENV_TEST_ARCH="$(go env GOARCH)"; fi # use Rosetta on M1 macs
curl -sSLo envtest-bins.tar.gz "https://storage.googleapis.com/kubebuilder-tools/kubebuilder-tools-${K8S_VERSION}-$(go env GOOS)-${ENV_TEST_ARCH}.tar.gz"
```

**Note:** `envTest` cannot run pods, so is limited, but is typically all that's needed to test controllers and webhooks.

### Dealing with flaky integration tests

Tests in `test/integration/` use envTest, which is a real APIServer and ETCD instance,
so they are sometimes slower than Gomega's 1s `eventually` timeout. 

When run locally, without the env var `CI` set to `true`, Gomega uses the 1s timeout.

**Please do not** specify a timeout for `eventually`. If the server takes too long because of load
on your machine, you can use:
```shell
$ export GOMEGA_DEFAULT_EVENTUALLY_TIMEOUT=2s
$ make test-integration 
$ # or 
$ gingko ./test/integration/<path to specific spec>
```

When run in CI, we run `make test CI=true` which ensures integrations run with a 10s timeout to avoid flaky tests. 


### Running integration tests with a complete cluster

Declarative `kuttl` tests can be run against a real cluster, using `kind`. This approach is slower but can be useful
sometimes for easy debugging with the `kubectl` command line tool.
```
make test-kuttl-kind  # see below section about testing with a full cluster.
```


## Pull Request Flow

Often, code in a pull request will be worked on by a pair from the team maintaining Catographer.
In this case, a pull request should be made by one engineer. The second engineer should approve
the pull request. If the work is not blocking other stories, the pull request should stay open
overnight, to allow others on the team time to read. The following morning, the pull request should be
merged.


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

1. a tagged commit in the Git history
2. a GitHub release aiming at that tag making available a few assets that make
   Cartographer installable in a cluster

While the first step is manual (someone needs to `git tag` and then `git
push`), the second isn't: the [release.yaml](.github/workflows/release.yaml)
GitHub workflow takes care of preparing the assets and pushing to GitHub all
the necessary bits.

Although `2` is automated, it's still possible to do the procedure manually.

### Automatic release
1. Check the previous release - [https://github.com/vmware-tanzu/cartographer/releases](https://github.com/vmware-tanzu/cartographer/releases)
2. Create a tag for the new release
```bash
git tag v0.0.x # or v0.0.x-rc.n
git push origin <tag-name>
```
3. Ensure workflow has kicked off - [https://github.com/vmware-tanzu/cartographer/actions](https://github.com/vmware-tanzu/cartographer/actions)


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
RELEASE_VERSION=v0.0.6-rc ./hack/publish-github-release.sh
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


## Maintaining Documentation

### CRD definitions
CRD manifests must include clear documentation. Use go doc comments liberally in the CRD objects under `/pkg/api`

The doc comments are presented in both `kubectl explain` documentation as well in our documentation website.
It's important to realise there are limitations to formatting in both situations. In `kubectl explain` the entire comment
runs together, eg:

```
	// Sources is a list of references to other 'source' resources in this list.
	// A source resource has the kind ClusterSourceTemplate
	//
	// In a template, sources can be consumed as:
	//    $(sources.<name>.url)$ and $(sources.<name>.revision)$
	//
	// If there is only one source, it can be consumed as:
	//    $(source.url)$ and $(source.revision)$
```

becomes:
```text
  sources	<[]Object>
     Sources is a list of references to other 'source' resources in this list. A
     source resource has the kind ClusterSourceTemplate In a template, these can
     be consumed as: $(sources.<name>.url)$ and $(sources.<name>.revision)$ If there
     is only one source, it can be consumed as: $(source.url)$
     and $(sources.revision)$
```

#### Use conjunctions and comma seperated lists.
```
// aye
// bee
// cee
```
should be:
```
// aye, bee and cee
```

#### Stops at the end of every sentence, capitals at the start of every sentence.
```text
Talking about one thing

talking about another
```
presents as:
```
Talking about one thing talking about another
```

instead use:

```text
Talking about one thing.

Talking about another.
```
which is presented as:
```
Talking about one thing. Talking about another.
```

[carvel Packaging]: https://carvel.dev/kapp-controller/docs/latest/packaging/
[imgpkg bundle]: https://carvel.dev/imgpkg/docs/latest/
 

