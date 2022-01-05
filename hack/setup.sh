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

readonly GIT_WRITER_SSH_USER=${GIT_WRITER_SSH_USER:-"ssh://git"}
readonly GIT_WRITER_SERVER=${GIT_WRITER_SERVER:-$HOST_ADDR}
readonly GIT_WRITER_SERVER_PORT=${GIT_WRITER_SERVER_PORT:-"222"}
readonly GIT_WRITER_USERNAME=${GIT_WRITER_USERNAME:-"example"}
readonly GIT_WRITER_SSH_USER_EMAIL=${GIT_WRITER_SSH_USER_EMAIL:-"example@example.com"}
readonly GIT_WRITER_PROJECT=${GIT_WRITER_PROJECT:-"example"}
readonly GIT_WRITER_REPOSITORY=${GIT_WRITER_REPOSITORY:-"test-tekton-git-cli"}

if [[ -z "$(which ssh)" ]]; then
      apt-get -yq update && apt-get -yq install openssh-client
fi

if [[ -n "$GIT_WRITER_SERVER_PORT" ]]; then
      readonly PORT_FLAG="-p $GIT_WRITER_SERVER_PORT"
fi

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
                        start_repository
                        install_source_controller
                        install_kpack
                        install_knative_serving
                        install_tekton
                        install_tekton_git_cli_task
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

                        test_gitops

                        log "all tests passed!!"
                        ;;

                teardown-example)
                        teardown_runnable_example
                        teardown_example_sc "basic-sc"
                        teardown_example_sc "testing-sc"
                        teardown_gitops_example
                        ;;

                teardown)
                        delete_containers
                        delete_repository
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

        # Check if repository exists
        add_ssh_token &&
          git ls-remote "$GIT_WRITER_SSH_USER@$GIT_WRITER_SERVER:$GIT_WRITER_SERVER_PORT/$GIT_WRITER_PROJECT/$GIT_WRITER_REPOSITORY.git" 1> /dev/null 2> /dev/null &&
          echo 'repository found' &&
          return || true

        # Check if gitea server exists
        if docker ps | grep gitea > /dev/null; then
          # Check if user exists
          CONTAINER_ID=$(docker ps --filter name=gitea | cut -d ' ' -f1 | head -2 | tail -1)
          if docker exec -u git "$CONTAINER_ID" gitea admin user list | grep "$GIT_WRITER_SSH_USER_EMAIL" 1> /dev/null 2> /dev/null; then
            echo 'bad state: user exists on gitea server but expected repository does not'
            echo 'teardown gitea server and rerun'
            exit 1
          fi
        fi

        # Bring up gitea server container
        docker-compose -f hack/docker-compose.yaml up -d

        # Apply gitea configuration
        CONTAINER_ID=$(docker ps --filter name=gitea | cut -d ' ' -f1 | head -2 | tail -1)
        docker cp hack/golden-app.ini "$CONTAINER_ID:/data/gitea/conf/app.ini"
        docker restart "$CONTAINER_ID"

        # Create user in server
        DOCKER_EXEC_SUCCESS=false
        until $DOCKER_EXEC_SUCCESS
        do
          GITEA_TOKEN="$(docker exec -u git "$CONTAINER_ID" gitea admin user create \
            --username "$GIT_WRITER_USERNAME" \
            --password hi \
            --email "$GIT_WRITER_SSH_USER_EMAIL" \
            --admin \
            --access-token \
            --must-change-password=false |
            tee >(cat 1>&2) |
            head -1 |
            rev |
            cut -d ' ' -f1 |
            rev)" && DOCKER_EXEC_SUCCESS=true || echo "attempting docker exec again" && sleep 5
        done

        # Generate public/private key
        ssh-keygen -t rsa -b 4096 -C "$GIT_WRITER_SSH_USER_EMAIL" -f hack/gitea-key -P ""
        GITEA_KEY_PUB="$(cat hack/gitea-key.pub)"

        # Add the public key to the user on the server
        CURL_SUCCEED=false
        until $CURL_SUCCEED
        do
          curl -X 'POST' \
            "http://localhost:3000/api/v1/admin/users/$GIT_WRITER_USERNAME/keys" \
            -H 'accept: application/json' \
            -H 'Content-Type: application/json' \
            -H "Authorization: token $GITEA_TOKEN" \
            -d "{\"key\": \"$GITEA_KEY_PUB\", \"read_only\": false, \"title\": \"string\"}" \
            --fail && CURL_SUCCEED=true || echo "retrying curl" && sleep 5
        done

        # Create a repo on the server
        curl -X 'POST'   'http://localhost:3000/api/v1/user/repos'   -H 'accept: application/json'   -H 'Content-Type: application/json'   -d '{"name": "'"$GIT_WRITER_REPOSITORY"'", "private": false}' -H "Authorization: token $GITEA_TOKEN"

        GIT_WRITER_SERVER_PUBLIC_TOKEN=${GIT_WRITER_SERVER_PUBLIC_TOKEN:-"$(ssh-keyscan "$PORT_FLAG" "$GIT_WRITER_SERVER")"}
        echo "$GIT_WRITER_SERVER_PUBLIC_TOKEN" > hack/gitea-server-public-key

        add_ssh_token

        pushd "$(mktemp -d)"
                touch README.md

                git init

                git config --local user.email "$GIT_WRITER_USERNAME"
                git config --local user.name "$GIT_WRITER_SSH_USER_EMAIL"
                git config --local init.defaultBranch main

                git add README.md
                git commit -m "first commit"
                git remote add origin "$GIT_WRITER_SSH_USER@$GIT_WRITER_SERVER:$GIT_WRITER_SERVER_PORT/$GIT_WRITER_PROJECT/$GIT_WRITER_REPOSITORY.git"
                git branch -m main
                git push -u origin main
        popd
}

add_ssh_token() {
  if [[ -z ${GIT_WRITER_SSH_TOKEN+nullword} ]]; then
    if [[ -f hack/gitea-key ]]; then
      GIT_WRITER_SSH_TOKEN="$(cat hack/gitea-key)"
    else
      return 1
    fi
  fi

  ssh-add -t 1000 - <<< "$GIT_WRITER_SSH_TOKEN" 2> /dev/null || {
        mkdir -p ~/.ssh
        cp hack/gitea-key ~/.ssh/id_rsa
        chmod 600 ~/.ssh/id_rsa
  }

  GIT_WRITER_SERVER_PUBLIC_TOKEN=$(cat hack/gitea-server-public-key)
  echo "$GIT_WRITER_SERVER_PUBLIC_TOKEN" >> ~/.ssh/known_hosts
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

install_tekton_git_cli_task() {
  kapp deploy --yes -a tekton-git-cli -f https://raw.githubusercontent.com/tektoncd/catalog/main/task/git-cli/0.2/git-cli.yaml
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
                --data-value image_prefix="$REGISTRY/example-$test_name-")
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

teardown_gitops_example() {
      clean_up_git_repo
      rm hack/git_entropy || true
      test_name="gitwriter-sc"
      kapp delete --yes -a "example-$test_name"
      kapp delete --yes -a "setup-example-$test_name"
      kapp delete --yes -a example-deliver
#      until [[ -z $(kubectl get pods -l "serving.knative.dev/configuration=$test_name" -o name) ]]; do sleep 1; done
      log "teardown of '$test_name' complete"
}

clean_up_git_repo() {
      if [[ ! -f hack/git_entropy ]]; then
        return 0
      fi

      log "cleaning up git repo"

      BRANCH="$(cat hack/git_entropy)"

      add_ssh_token

      pushd "$(mktemp -d)"
            git clone "$GIT_WRITER_SSH_USER@$GIT_WRITER_SERVER:$GIT_WRITER_SERVER_PORT/$GIT_WRITER_PROJECT/$GIT_WRITER_REPOSITORY.git"
            pushd "$GIT_WRITER_REPOSITORY"
                  git ls-remote --heads origin "$BRANCH"
                  readonly BRANCH_EXISTS="$(git ls-remote --heads origin "$BRANCH")"

                  if [[ -n "$BRANCH_EXISTS" ]]; then
                        git push -d origin "$BRANCH"
                  fi
            popd
      popd

      log "done"
      return 0
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
        if [[ $counter -gt 20 ]]; then
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
        if [[ $counter -gt 20 ]]; then
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
        if [[ $counter -gt 20 ]]; then
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

test_gitops() {
      setup_source_to_gitops
      test_source_to_gitops
      setup_gitops_to_app
      test_example_sc "gitwriter-sc"
      teardown_gitops_example
}

setup_source_to_gitops() {
      log "setting up source-to-gitops"

      export test_name="gitwriter-sc"

      touch hack/git_entropy
      echo $RANDOM | base64 > hack/git_entropy

      GIT_WRITER_SERVER_PUBLIC_TOKEN=${GIT_WRITER_SERVER_PUBLIC_TOKEN:-"$(cat hack/gitea-server-public-key)"}

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
              --data-value image_prefix="$REGISTRY/example-$test_name-" \
              --data-value git_writer.message="Some perturbation: $(cat hack/git_entropy)" \
              --data-value git_writer.ssh_user="$GIT_WRITER_SSH_USER" \
              --data-value git_writer.username="$GIT_WRITER_USERNAME" \
              --data-value git_writer.user_email="$GIT_WRITER_SSH_USER_EMAIL" \
              --data-value git_writer.server="$GIT_WRITER_SERVER" \
              --data-value git_writer.port="$GIT_WRITER_SERVER_PORT" \
              --data-value git_writer.repository="$GIT_WRITER_PROJECT/$GIT_WRITER_REPOSITORY.git" \
              --data-value git_writer.branch="$(cat hack/git_entropy)" \
              --data-value git_writer.base64_encoded_ssh_key="$(echo "$(cat hack/gitea-key)" | base64)" \
              --data-value git_writer.base64_encoded_known_hosts="$(echo "$GIT_WRITER_SERVER_PUBLIC_TOKEN" | base64)")
}

test_source_to_gitops() {
        log "testing source-to-gitops"

        GIT_ENTROPY="$(cat hack/git_entropy)"
        BRANCH="$GIT_ENTROPY"

        EXPECTED_GIT_MESSAGE="Some perturbation: $GIT_ENTROPY"

        SUCCESS=false

        add_ssh_token

        pushd "$(mktemp -d)"
              git clone "$GIT_WRITER_SSH_USER@$GIT_WRITER_SERVER:$GIT_WRITER_SERVER_PORT/$GIT_WRITER_PROJECT/$GIT_WRITER_REPOSITORY.git"
              echo "looking for branch $BRANCH"
              pushd "$GIT_WRITER_REPOSITORY"
                    for i in {20..1}; do
                            echo "- attempt $i"

                            git fetch --all --prune
                            if [[ ! "$(git branch --show-current)" = "$BRANCH" ]]; then
                                  git checkout "$BRANCH" > /dev/null 2> /dev/null || sleep "$i" && continue
                            fi

                            git pull #> /dev/null 2> /dev/null
                            MOST_RECENT_GIT_MESSAGE="$(git log -1 --pretty=%B)"

                            if [[ "$EXPECTED_GIT_MESSAGE" = "$MOST_RECENT_GIT_MESSAGE" ]]; then
                                    log 'gitops worked! sweet'
                                    SUCCESS=true
                                    break
                            fi

                            sleep "$i"
                    done
              popd
        popd

        if [[ "$SUCCESS" = true ]]; then
              return 0
        else
              log 'FAILED :('
              exit 1
        fi
}

setup_gitops_to_app() {
        log "setting up gitops-to-app"

        GIT_WRITER_SERVER_PUBLIC_TOKEN=${GIT_WRITER_SERVER_PUBLIC_TOKEN:-"$(cat hack/gitea-server-public-key)"}

        ytt --ignore-unknown-comments \
                -f "$DIR/../examples/basic-delivery" \
                --data-value git_writer.server="$GIT_WRITER_SERVER:$GIT_WRITER_SERVER_PORT" \
                --data-value git_writer.repository="$GIT_WRITER_PROJECT/$GIT_WRITER_REPOSITORY" \
                --data-value git_writer.branch="$(cat hack/git_entropy)" \
                --data-value git_writer.base64_encoded_ssh_key="$(echo "$(cat hack/gitea-key)" | base64)" \
                --data-value git_writer.base64_encoded_known_hosts="$(echo "$GIT_WRITER_SERVER_PUBLIC_TOKEN" | base64)" |
                kapp deploy --yes -a example-deliver -f-
}

delete_containers() {
      docker rm -f $REGISTRY_CONTAINER_NAME || true
      docker rm -f $KUBERNETES_CONTAINER_NAME || true
}

delete_repository() {
      docker-compose -f hack/docker-compose.yaml down -v
      rm -rf "$DIR/gitea" || true
      rm hack/gitea-key.pub hack/gitea-key hack/gitea-server-public-key || true
}

log() {
        printf '\n\t\033[1m%s\033[0m\n\n' "$1" 1>&2
}

main "$@"
