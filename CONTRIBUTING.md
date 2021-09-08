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
- [`pack`]: for building the controller's container image using buildpacks.

[`ctlptl`]: https://github.com/tilt-dev/ctlptl
[`go`]: https://golang.org/dl/
[`kapp`]: https://github.com/vmware-tanzu/carvel-kapp
[`kbld`]: https://github.com/vmware-tanzu/carvel-kbld
[`kind`]: https://kind.sigs.k8s.io/docs/user/quick-start/
[`ko`]: https://github.com/google/ko
[`kuttl`]: https://github.com/kudobuilder/kuttl
[`pack`]: https://github.com/buildpacks/pack
[`ytt`]: https://github.com/vmware-tanzu/carvel-ytt

## Running a local cluster
A local kind cluster with Cartographer installed can be stood up with the command:
```yaml
make deploy-local
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

The `envTest` package runs a simplified control plane, using `kubernetes-apiserver` and `etcd`. You'll need `etcd` 
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

## Updating GitLab CI configuration

The configuration for GitLab CI that is picked up by the GitLab runner
([./.gitlab-ci.yml](./.gitlab-ci.yml)) is generated (through `make
gen-ci-manifest`) based of a template that you can find at
[./.gitlab/.gitlab-ci.yml](./.gitlab/.gitlab-ci.yml).

i.e., if you want to introduce extra commands for pipeline, make sure you
update `./.gitlab/.gitlab-ci.yml` and then run `make gen-ci-manifest`. If you
want to include or update dependencies in the base image, update
`./.gitlab/Dockerfile` and then run `make gen-ci-manifest`.

This allows us to declaratively express how the base image to be used when
running the tests should look like
([./.gitlab/Dockerfile](./.gitlab/Dockerfile)) and have such image reference
specified in the final CI manifest by leveraging [kbld].

See `gen-ci-manifest` on [./Makefile](./Makefile) to know more about how the
generation takes place.


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

At the moment, releasing Cartographer is manual.

Before getting started, make sure you have a running `docker` daemon and all
the dependencies mentioned above in the [Development Dependencies
section](#development-dependencies).


### What it consists of

Releasing Cartographer consists of producing a YAML file that contains all the
necessary Kubernetes objects that needs to be submitted to a cluster in order
to have the controller running there with the proper access to the Kubernetes
API in such environment.


```bash
# grab the credentials from lastpass for dockerhub
#
docker login

# point `make` at the `release` target, which takes care of generating any YAML
# files based of Go code, as well as building the container image with the
# controller that's then placed in the Deployment's pod template spec.
#
# when done, a `releases/release.yaml` file will have been populated with all
# the YAML necessary for bringing `cartographer` up in a Kubernetes cluster via
# `kubectl apply -f ./releases/release.yaml`.
#
KO_DOCKER_REPO=projectcartographer \
	make release
```

That final file (`releases/release.yaml`) consists of:

1. `CustomResourceDefinition` objects that tells `kube-apiserver` how to expand
   its API to now know about Cartographer primitives

1. RBAC objects that are bound to the controller ServiceAccount so that the
   controller can reach the Kubernetes API server to retrieve state and mutate
   it when necessary

1. Deployment-related objects that stand Cartographer up in a Pod so our code runs
   inside the cluster (using that ServiceAccount that grants it the permissions
   when it comes to reaching out to the API).

As the `Deployment` needs a container image for the pods to use to run our
controller, we must have a way of building that container image in the first
place. For that, we make use of `ko`, which given a YAML file where it can find
an `image: ko://<package>`, it then replaces that with the reference to the
image it built and pushed to a registry configured via `KO_DOCKER_REPO` (see
[deployment.yaml](./config/manager/deployment.yaml)). 

## Running the e2e tests

Cartographer has a  script that allows users to:
   - create a local cluster with a local repository
   - push the image of the controller to that cluster
   - run the controller
   - create the supply chain(s) and workload(s) the example directory
   - assure that the expected objects are created by the Cartographer controller

To run the tests:
```bash
./hack/ci/e2e.sh run
```
To teardown (necessary if users wish to rerun the tests):
```bash
./hack/ci/e2e.sh teardown
```
