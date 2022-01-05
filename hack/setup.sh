#!/usr/bin/env bash
# Copyright 2021 VMware
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

# shellcheck disable=SC2155
readonly DIR="$(cd "$(dirname "$0")" && pwd)"
readonly HOST_ADDR=${HOST_ADDR:-$("$DIR"/ip.py)}
readonly REGISTRY_PORT=${REGISTRY_PORT:-5000}
readonly REGISTRY=${REGISTRY:-"${HOST_ADDR}:${REGISTRY_PORT}"}
readonly KIND_IMAGE=${KIND_IMAGE:-kindest/node:v1.21.1}
readonly RELEASE_VERSION=${RELEASE_VERSION:-""}
# shellcheck disable=SC2034  # This _should_ be marked as an extern but I clearly don't understand how it operates in github actions
readonly DOCKER_CONFIG=${DOCKER_CONFIG:-"/tmp/cartographer-docker"}

readonly REGISTRY_CONTAINER_NAME=cartographer-registry
readonly KUBERNETES_CONTAINER_NAME=cartographer-control-plane

readonly CERT_MANAGER_VERSION=1.5.3
readonly KAPP_CONTROLLER_VERSION=0.30.0
readonly KNATIVE_SERVING_VERSION=0.26.0
readonly KPACK_VERSION=0.4.3
readonly SECRETGEN_CONTROLLER_VERSION=0.6.0
readonly SOURCE_CONTROLLER_VERSION=0.17.0
readonly TEKTON_VERSION=0.30.0

main() {
        test $# -eq 0 && show_usage_help
        display_vars "$@"

        for command in "$@"; do
                case $command in
                cluster)
                        start_registry
                        start_local_cluster
                        install_cert_manager
                        install_kapp_controller
                        install_secretgen_controller
                        ;;

                cartographer)
                        install_cartographer_package
                        ;;

                example-dependencies)
                        install_source_controller
                        install_kpack
                        install_knative_serving
                        install_tekton
                        ;;

                example)
                        test_runnable_example
                        teardown_runnable_example

                        setup_example_sc "basic-sc"
                        test_example_sc "basic-sc"
                        teardown_example_sc "basic-sc"

                        setup_example_sc "testing-sc"
                        test_example_sc "testing-sc"
                        teardown_example_sc "testing-sc"

                        log "all tests passed!!"
                        ;;

                teardown-example)
                        teardown_runnable_example
                        teardown_example_sc "basic-sc"
                        teardown_example_sc "testing-sc"
                        ;;

                teardown)
                        delete_containers
                        ;;

                *)
                        echo "error: unknown command '$command'."
                        show_usage_help
                        exit 1
                        ;;
                esac
        done
}

install_cartographer_package() {
        log "build cartographer release and installing it"
        env REGISTRY="$REGISTRY" RELEASE_VERSION="$RELEASE_VERSION" DOCKER_CONFIG="$DOCKER_CONFIG" ./hack/release.sh

        ytt --ignore-unknown-comments \
                --data-value registry="$REGISTRY" \
                -f ./hack/registry-auth |
                kapp deploy -a cartographer --yes \
                        -f ./release/package \
                        -f-
}

show_usage_help() {
        echo "usage: $0 <command...>"
        cat <<-COMMANDS
	commands:

	- cluster 		brings up a local cluster and a registry

	- cartographer 		build a release of cartographer and install it in the
			cluster

	- example-dependencies 	installs dependencies used throughout examples
	- example 		install the example and runs a minimal test on it

	- teardown		gets rid of the local cluster and registry created
	- teardown-example 	gets rid of just the example installed (workload, etc)
	COMMANDS
}

display_vars() {
        cat <<-DISPLAY
	Variables:

	COMMANDS: 	$*

	DIR:		$DIR
	HOST_ADDR:	$HOST_ADDR
	KIND_IMAGE:	$KIND_IMAGE
	REGISTRY:	$REGISTRY
	REGISTRY_PORT:	$REGISTRY_PORT
	DISPLAY
}

start_registry() {
        log "starting registry"

        echo -e "\n\nregistry credentials:\n
        username: admin
        password: admin
        "

        env DOCKER_USERNAME=admin \
                DOCKER_PASSWORD=admin \
                DOCKER_REGISTRY="$REGISTRY" \
                DOCKER_CONFIG="$DOCKER_CONFIG" \
                "$DIR/docker-login.sh"

        docker container inspect $REGISTRY_CONTAINER_NAME &>/dev/null && {
                echo "registry already exists"
                return
        }

        docker run \
                --detach \
                -v "$DIR/registry-auth:/auth" \
                -e "REGISTRY_AUTH=htpasswd" \
                -e "REGISTRY_AUTH_HTPASSWD_REALM=Registry Realm" \
                -e "REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd" \
                --name "$REGISTRY_CONTAINER_NAME" \
                --publish "${REGISTRY_PORT}":5000 \
                registry:2
}

start_local_cluster() {
        log "starting local cluster"

        docker container inspect $KUBERNETES_CONTAINER_NAME &>/dev/null && {
                echo "cluster already exists"
                return
        }

        cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: cartographer
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry]
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
      [plugins."io.containerd.grpc.v1.cri".registry.mirrors."${REGISTRY}"]
        endpoint = ["http://${REGISTRY}"]
    [plugins."io.containerd.grpc.v1.cri".registry.configs]
      [plugins."io.containerd.grpc.v1.cri".registry.configs."${REGISTRY}".tls]
        insecure_skip_verify = true
nodes:
  - role: control-plane
    image: ${KIND_IMAGE}
EOF
}

install_cert_manager() {
        ytt --ignore-unknown-comments \
                -f "$DIR/overlays/remove-resource-requests-from-deployments.yaml" \
                -f https://github.com/jetstack/cert-manager/releases/download/v$CERT_MANAGER_VERSION/cert-manager.yaml |
                kapp deploy --yes -a cert-manager -f-
}

install_source_controller() {
        kubectl create namespace gitops-toolkit || true

        kubectl create clusterrolebinding gitops-toolkit-admin \
                --clusterrole=cluster-admin \
                --serviceaccount=gitops-toolkit:default || true

        ytt --ignore-unknown-comments \
                -f "$DIR/overlays/remove-resource-requests-from-deployments.yaml" \
                -f https://github.com/fluxcd/source-controller/releases/download/v$SOURCE_CONTROLLER_VERSION/source-controller.crds.yaml \
                -f https://github.com/fluxcd/source-controller/releases/download/v$SOURCE_CONTROLLER_VERSION/source-controller.deployment.yaml |
                kapp deploy --yes -a gitops-toolkit --into-ns gitops-toolkit -f-
}

install_kpack() {
        ytt --ignore-unknown-comments \
                -f "$DIR/overlays/remove-resource-requests-from-deployments.yaml" \
                -f https://github.com/pivotal/kpack/releases/download/v$KPACK_VERSION/release-$KPACK_VERSION.yaml |
                kapp deploy --yes -a kpack -f-

}

install_kapp_controller() {
        # Ensure script halts if kubectl is not installed
        kubectl version --client

        kubectl create clusterrolebinding default-admin \
                --clusterrole=cluster-admin \
                --serviceaccount=default:default || true

        ytt --ignore-unknown-comments \
                -f "$DIR/overlays/remove-resource-requests-from-deployments.yaml" \
                -f https://github.com/vmware-tanzu/carvel-kapp-controller/releases/download/v$KAPP_CONTROLLER_VERSION/release.yml |
                kapp deploy --yes -a kapp-controller -f-
}

install_secretgen_controller() {
        ytt --ignore-unknown-comments \
                -f "$DIR/overlays/remove-resource-requests-from-deployments.yaml" \
                -f https://github.com/vmware-tanzu/carvel-secretgen-controller/releases/download/v$SECRETGEN_CONTROLLER_VERSION/release.yml |
                kapp deploy --yes -a secretgen-controller -f-
}

install_knative_serving() {
        ytt --ignore-unknown-comments \
                -f https://github.com/knative/serving/releases/download/v$KNATIVE_SERVING_VERSION/serving-core.yaml \
                -f https://github.com/knative/serving/releases/download/v$KNATIVE_SERVING_VERSION/serving-crds.yaml \
                -f "$DIR/overlays/remove-resource-requests-from-deployments.yaml" |
                kapp deploy --yes -a knative-serving -f-
}

install_tekton() {
        ytt --ignore-unknown-comments \
                -f https://storage.googleapis.com/tekton-releases/pipeline/previous/v$TEKTON_VERSION/release.yaml \
                -f "$DIR/overlays/remove-resource-requests-from-deployments.yaml" |
                kapp deploy --yes -a tekton -f-
}

setup_example_sc() {
        export test_name="$1"
        kapp deploy --yes -a "setup-example-$test_name" \
            -f <(ytt --ignore-unknown-comments \
                -f "$DIR/../examples/shared" \
                -f "$DIR/../examples/$test_name/values.yaml" \
                --data-value registry.server="$REGISTRY" \
                --data-value registry.username=admin \
                --data-value registry.password=admin \
                --data-value image_prefix="$REGISTRY/example-$test_name-")
        kapp deploy --yes -a "example-$test_name" \
            -f <(ytt --ignore-unknown-comments \
                -f "$DIR/../examples/$test_name" \
                --data-value registry.server="$REGISTRY" \
                --data-value registry.username=admin \
                --data-value registry.password=admin \
                --data-value workload_name="$test_name" \
                --data-value image_prefix="$REGISTRY/example-$test_name-") \

}

teardown_example_sc() {
        export test_name="$1"
        kapp delete --yes -a "example-$test_name"
        kapp delete --yes -a "setup-example-$test_name"
#        until [[ -z $(kubectl get pods -l "serving.knative.dev/configuration=$test_name" -o name) ]]; do sleep 1; done
        log "teardown of '$test_name' complete"
}

test_example_sc() {
        export test_name="$1"
        log "testing '$test_name'"

        for _ in {1..5}; do
                for sleep_duration in {15..1}; do
                        local deployed_pods
                        deployed_pods=$(kubectl get pods \
                                -l "serving.knative.dev/configuration=$test_name" \
                                -o name)

                        if [[ "$deployed_pods" == *"$test_name"* ]]; then
                                log "testing '$test_name' SUCCEEDED! sweet"
                                return 0
                        fi

                        echo "- waiting $sleep_duration seconds"
                        sleep "$sleep_duration"
                done

                kubectl tree workload "$test_name"
        done

        log "testing '$test_name' FAILED :("
        exit 1
}

test_runnable_example() {
      log "test runnable"
      first_output_revision=""
      second_output_revision=""
      third_output_revision=""

      kubectl apply -f "$DIR/../examples/runnable-tekton/00-setup"
      kubectl apply -f "$DIR/../examples/runnable-tekton/01-tests-pass"

      counter=0
      until [[ -n $(kubectl get taskruns -o json | jq '.items[] | .status.conditions[0].status' | grep True) ]]; do
        sleep 5
        if [[ $counter -gt 12 ]]; then
          log "runnable test fails"
          exit 1
        else
          echo "waiting 5 seconds for expected passing test to succeed"
          (( counter+1 ))
        fi
      done

      sleep 5
      first_output_revision=$(kubectl get runnable test -o json | jq '.status.outputs.revision')

      kubectl patch runnable test --type merge --patch "$(cat "$DIR/../examples/runnable-tekton/02-tests-fail/runnable-patch.yml")"

      counter=0
      until [[ -n $(kubectl get taskruns -o json | jq '.items[] | .status.conditions[0].status' | grep False) ]]; do
        sleep 5
        if [[ $counter -gt 12 ]]; then
          log "runnable test fails"
          exit 1
        else
          echo "waiting 5 seconds for expected failing test to fail"
          (( counter+1 ))
        fi
      done

      sleep 5
      second_output_revision=$(kubectl get runnable test -o json | jq '.status.outputs.revision')
      if [[ "$first_output_revision" != "$second_output_revision" ]]; then
        log "runnable test fails"
        exit 1
      fi

      kubectl patch runnable test --type merge --patch "$(cat "$DIR/../examples/runnable-tekton/03-tests-pass/runnable-patch.yml")"

      counter=0
      until [[ $(kubectl get taskruns -o json | jq '.items[] | .status.conditions[0].status' | grep True | wc -l) -eq 2 ]]; do
        sleep 5
        if [[ $counter -gt 12 ]]; then
          log "runnable test fails"
          exit 1
        else
          echo "waiting 5 seconds for expected passing test to succeed"
          (( counter+1 ))
        fi
      done

      sleep 5
      third_output_revision=$(kubectl get runnable test -o json | jq '.status.outputs.revision')
      if [[ "$first_output_revision" == "$third_output_revision" ]]; then
        log "runnable test fails"
        exit 1
      fi

      log "runnable test passes"
      return 0
}

teardown_runnable_example() {
      kubectl delete -f "$DIR/../examples/runnable-tekton/00-setup" --ignore-not-found
      kubectl delete -f "$DIR/../examples/runnable-tekton/01-tests-pass" --ignore-not-found

      log "teardown of runnable example complete"
}

delete_containers() {
        docker rm -f $REGISTRY_CONTAINER_NAME || true
        docker rm -f $KUBERNETES_CONTAINER_NAME || true
}

log() {
        printf '\n\t\033[1m%s\033[0m\n\n' "$1" 1>&2
}

main "$@"
