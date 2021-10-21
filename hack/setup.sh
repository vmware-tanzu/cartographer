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
readonly GIT_WRITER_SSH_USER=${GIT_WRITER_SSH_USER:-"git"}
readonly GIT_WRITER_SERVER=${GIT_WRITER_SERVER:-"github.com"}
readonly GIT_WRITER_PROJECT=${GIT_WRITER_PROJECT:-"waciumawanjohi"}
readonly GIT_WRITER_REPOSITORY=${GIT_WRITER_REPOSITORY:-"test-tekton-git-cli"}
readonly GIT_WRITER_SSH_TOKEN=${GIT_WRITER_SSH_TOKEN:-"$(lpass show --notes waciumawanjohi-test-tekton-git-cli-ssh-key)"}
readonly GIT_WRITER_SERVER_PUBLIC_TOKEN=${GIT_WRITER_SERVER_PUBLIC_TOKEN:-"$(ssh-keyscan -H "$GIT_WRITER_SERVER")"}

# shellcheck disable=SC2034  # This _should_ be marked as an extern but I clearly don't understand how it operates in github actions
readonly DOCKER_CONFIG="/tmp/cartographer-docker"

readonly REGISTRY_CONTAINER_NAME=cartographer-registry
readonly KUBERNETES_CONTAINER_NAME=cartographer-control-plane

readonly CERT_MANAGER_VERSION=1.5.3
readonly KAPP_CONTROLLER_VERSION=0.27.0
readonly KNATIVE_SERVING_VERSION=0.25.0
readonly KPACK_VERSION=0.4.0
readonly SECRETGEN_CONTROLLER_VERSION=0.5.0
readonly SOURCE_CONTROLLER_VERSION=0.15.4
readonly TEKTON_VERSION=0.28.0

main() {
        test $# -eq 0 && show_usage_help
        display_vars "$@"

        for command in "$@"; do
                case $command in
                cluster)
                        start_registry
#                        start_repository
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
                        install_tekton_git_cli_task
                        ;;

                example)
                        setup_example
                        test_example
                        ;;

                teardown-example)
                        teardown_example
                        ;;

                teardown)
                        delete_containers
#                        delete_generated_repository_keys
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

start_repository() {
        log "start repository"

        # Bring up gitea server container
        docker-compose -f hack/docker-compose.yaml up -d

        # Apply gitea configuration
        CONTAINER_ID=$(docker ps | cut -d ' ' -f1 | head -2 | tail -1)
        docker cp hack/golden-app.ini "$CONTAINER_ID:/data/gitea/conf/app.ini"
        docker restart "$CONTAINER_ID"

        # Create user in server
        DOCKER_EXEC_SUCCESS=false
        until $DOCKER_EXEC_SUCCESS
        do
          GITEA_TOKEN="$(docker exec -u git "$CONTAINER_ID" gitea admin user create --username hi --password hi --email hi@example.com --admin --access-token --must-change-password=false | head -1 | rev | cut -d ' ' -f1 | rev)" && DOCKER_EXEC_SUCCESS=true || echo "attempting docker exec again" && sleep 5
        done

        # Generate public/private key
        ssh-keygen -t rsa -b 4096 -C "hi@example.com" -f hack/gitea-key -P ""
        GITEA_KEY_PUB="$(cat hack/gitea-key.pub)"

        # Add the public key to the user on the server
        CURL_SUCCEED=false
        until $CURL_SUCCEED
        do
          curl -X 'POST' \
            'http://localhost:3000/api/v1/admin/users/hi/keys' \
            -H 'accept: application/json' \
            -H 'Content-Type: application/json' \
            -H "Authorization: token $GITEA_TOKEN" \
            -d "{\"key\": \"$GITEA_KEY_PUB\", \"read_only\": false, \"title\": \"string\"}" \
            --fail && CURL_SUCCEED=true || echo "retrying curl" && sleep 5
        done

        # Create a repo on the server
        curl -X 'POST'   'http://localhost:3000/api/v1/user/repos'   -H 'accept: application/json'   -H 'Content-Type: application/json'   -d '{"name": "my-repo", "private": false}' -H "Authorization: token $GITEA_TOKEN"
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

install_tekton_git_cli_task() {
  kapp deploy --yes -a tekton-git-cli -f https://raw.githubusercontent.com/tektoncd/catalog/main/task/git-cli/0.2/git-cli.yaml
}

setup_example() {
        touch hack/git_entropy
        echo $RANDOM | base64 > hack/git_entropy

        ytt --ignore-unknown-comments \
                -f "$DIR/../examples/source-to-gitops" \
                --data-value registry.server="$REGISTRY" \
                --data-value registry.username=admin \
                --data-value registry.password=admin \
                --data-value image_prefix="$REGISTRY/example-" \
                --data-value git_writer.message="Some peturbation: $(cat hack/git_entropy)" \
                --data-value git_writer.ssh_user="$GIT_WRITER_SSH_USER" \
                --data-value git_writer.server="$GIT_WRITER_SERVER" \
                --data-value git_writer.repository="$GIT_WRITER_PROJECT/$GIT_WRITER_REPOSITORY.git" \
                --data-value git_writer.branch="$(cat hack/git_entropy)" \
                --data-value git_writer.base64_encoded_ssh_key="$(echo "$GIT_WRITER_SSH_TOKEN" | base64)" \
                --data-value git_writer.base64_encoded_known_hosts="$(echo "$GIT_WRITER_SERVER_PUBLIC_TOKEN" | base64)" |
                kapp deploy --yes -a example-supply -f-

        ytt --ignore-unknown-comments \
                -f "$DIR/../examples/gitops-to-app" \
                --data-value git_writer.server="$GIT_WRITER_SERVER" \
                --data-value git_writer.repository="$GIT_WRITER_PROJECT/$GIT_WRITER_REPOSITORY" \
                --data-value git_writer.branch="$(cat hack/git_entropy)" \
                --data-value git_writer.base64_encoded_ssh_key="$(echo "$GIT_WRITER_SSH_TOKEN" | base64)" \
                --data-value git_writer.base64_encoded_known_hosts="$(echo "$GIT_WRITER_SERVER_PUBLIC_TOKEN" | base64)" |
                kapp deploy --yes -a example-deliver -f-
}

teardown_example() {
        kapp delete --yes -a example-supply
        kapp delete --yes -a example-deliver
        rm hack/git_entropy
}

test_example() {
        log "testing"

        GIT_ENTROPY="$(cat hack/git_entropy)"
        BRANCH="$GIT_ENTROPY"

        EXPECTED_GIT_MESSAGE="Some peturbation: $GIT_ENTROPY"

        pushd "$(mktemp -d)"
              ssh-add -t 1000 - <<< "$GIT_WRITER_SSH_TOKEN" 2> /dev/null || {
                mkdir -p ~/.ssh
                echo "$GIT_WRITER_SSH_TOKEN" >> ~/.ssh/id_rsa
                echo "$GIT_WRITER_SERVER_PUBLIC_TOKEN" >> ~/.ssh/known_hosts
                chmod 600 ~/.ssh/id_rsa
              }
              git clone "$GIT_WRITER_SSH_USER@$GIT_WRITER_SERVER:$GIT_WRITER_PROJECT/$GIT_WRITER_REPOSITORY.git"
              echo "looking for branch $BRANCH"
              pushd "$GIT_WRITER_REPOSITORY"
                    for i in {20..1}; do
                            echo "- attempt $i"

                            local deployed_pods
                            deployed_pods=$(kubectl get pods \
                                    -l 'serving.knative.dev/configuration=dev' \
                                    -o name)

                            git pull > /dev/null 2> /dev/null
                            git checkout "$BRANCH" > /dev/null 2> /dev/null || continue
                            MOST_RECENT_GIT_MESSAGE="$(git log -1 --pretty=%B)"

                            if [[ -n "$deployed_pods" && "$EXPECTED_GIT_MESSAGE" = "$MOST_RECENT_GIT_MESSAGE" ]]; then
                                    echo "looks good; cleaning up git repo..."
                                    git push -d origin "$BRANCH"
                                    log 'SUCCEEDED! sweet'
                                    exit 0
                            fi

                            sleep "$i"
                    done
              popd
        popd

        log 'FAILED :('
        exit 1
}

delete_containers() {
        docker rm -f $REGISTRY_CONTAINER_NAME || true
        docker rm -f $KUBERNETES_CONTAINER_NAME || true
        docker-compose -f hack/docker-compose.yaml down -v
}

delete_generated_repository_keys() {
        rm hack/gitea-key.pub hack/gitea-key
}

log() {
        printf '\n\t\033[1m%s\033[0m\n\n' "$1" 1>&2
}

main "$@"
