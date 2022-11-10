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

# wokeignore:rule=disable no shellcheck alternative
# shellcheck disable=SC2155
readonly DIR="$(cd "$(dirname "$0")" && pwd)"
readonly HOST_ADDR=${HOST_ADDR:-$("$DIR"/ip.py)}
readonly REGISTRY_PORT=${REGISTRY_PORT:-5001}
readonly REGISTRY=${REGISTRY:-"${HOST_ADDR}:${REGISTRY_PORT}"}
# wokeignore:rule=disable no shellcheck alternative
# shellcheck disable=SC2034  # This _should_ be marked as an extern but I clearly don't understand how it operates in github actions
readonly DOCKER_CONFIG=${DOCKER_CONFIG:-"/tmp/cartographer-docker"}

readonly RELEASE_VERSION="v0.0.0-dev"
readonly TEST_NAME="gitwriter-sc"
readonly GIT_SERVER="git-server.default.svc.cluster.local:80"
readonly SOURCE_REPO="hello-world"
readonly SOURCE_BRANCH="main"
readonly CONFIG_REPO="gitops-test"
readonly CONFIG_BRANCH="main"
readonly CONFIG_COMMIT_MESSAGE="Update config"

main() {
  KIND_IMAGE='kindest/node:v1.24.6' "$DIR/setup.sh" cluster example-dependencies
  install_latest_released_cartographer

  port=$(available_port)
  setup_git_server "$port"

  source_dir="$(mktemp -d)"
  setup_source_repo "$source_dir" "$port"
  setup_source_to_gitops

  config_dir="$(mktemp -d)"
  test_source_to_gitops "$config_dir" "$port"

  setup_gitops_to_app
  test_gitops_to_app

  log "deploying cartographer from main"
  install_cartographer_from_current_commit
  verify_new_carto_deployed

  update_source "$source_dir" "$port"
  wait_for_new_config_commit "$config_dir"
  wait_for_knative_deployment_update
}

install_latest_released_cartographer() {
  log "installing latest released cartographer"

  ytt --ignore-unknown-comments \
    -f "$DIR/overlays/remove-resource-requests-from-deployments.yaml" \
    -f https://github.com/vmware-tanzu/cartographer/releases/latest/download/cartographer.yaml |
    kapp deploy --yes -a cartographer -f-
}

install_cartographer_from_current_commit() {
  log "building cartographer release and installing it"
  env \
    REGISTRY="$REGISTRY" \
    RELEASE_VERSION="$RELEASE_VERSION" \
    DOCKER_CONFIG="$DOCKER_CONFIG" \
    "$DIR/release.sh"

  ytt --ignore-unknown-comments \
    --data-value registry="$REGISTRY" \
    -f "$DIR/overlays/remove-resource-requests-from-deployments.yaml" \
    -f release/cartographer.yaml |
    kapp deploy --yes -a cartographer -f-
}

setup_git_server() {
  local port=$1

  log "setting up git server"

  kubectl apply -f "$DIR/git-server.yaml"

  until [[ $(kubectl get deployment git-server -o json | jq '.status.readyReplicas') == 1 ]]; do
    log "waiting for git-server deployment"
    sleep 10
  done

  kubectl port-forward service/git-server $port:80 &
  # wokeignore:rule=disable no shellcheck alternative
  # shellcheck disable=SC2064
  trap "kill $! || true" EXIT # wokeignore:rule=kill no *NIX alternative
  sleep 5
}

setup_source_repo() {
  local source_dir=$1
  local port=$2

  log "setting up source repo"

  pushd "$source_dir"
    git clone "http://localhost:$port/$SOURCE_REPO.git"
    pushd "$SOURCE_REPO"
      git pull https://github.com/carto-run/hello-world.git
      if [[ $(git branch --show-current) != "$SOURCE_BRANCH" ]]; then
        git checkout -b $SOURCE_BRANCH
      fi
      git push origin $SOURCE_BRANCH
    popd
  popd
}

setup_source_to_gitops() {
  log "setting up source to gitops"

  kapp deploy --yes -a "setup-example-$TEST_NAME" \
    -f <(ytt --ignore-unknown-comments \
      -f "$DIR/../examples/shared" \
      -f "$DIR/../examples/$TEST_NAME/values.yaml" \
      --data-value registry.server="$REGISTRY" \
      --data-value registry.username=admin \
      --data-value registry.password=admin \
      --data-value image_prefix="$REGISTRY/example-$TEST_NAME-")

  kapp deploy --yes -a "example-$TEST_NAME" \
    -f <(ytt --ignore-unknown-comments \
      -f "$DIR/../examples/$TEST_NAME" \
      --data-value registry.server="$REGISTRY" \
      --data-value registry.username=admin \
      --data-value registry.password=admin \
      --data-value workload_name="$TEST_NAME" \
      --data-value image_prefix="$REGISTRY/example-$TEST_NAME-" \
      --data-value source_repo.url="http://$GIT_SERVER/$SOURCE_REPO.git" \
      --data-value source_repo.branch="$SOURCE_BRANCH" \
      --data-value git_repository="http://$GIT_SERVER/$CONFIG_REPO.git" \
      --data-value git_branch="$CONFIG_BRANCH" \
      --data-value git_user_name="gitops-user" \
      --data-value git_user_email="gitops-user@example.com" \
      --data-value git_commit_message="$CONFIG_COMMIT_MESSAGE")
}

test_source_to_gitops() {
  local config_dir=$1
  local port=$2

  log "testing source to gitops"

  local success=false

  pushd "$config_dir"
    git clone "http://localhost:$port/$CONFIG_REPO.git"
    pushd "$CONFIG_REPO"
      for sleep_duration in {20..1}; do
        git fetch --all --prune

        git checkout "$CONFIG_BRANCH" >/dev/null 2>/dev/null || {
          echo "- waiting $sleep_duration seconds"
          sleep "$sleep_duration"
          continue
        }

        git pull >/dev/null 2>/dev/null
        MOST_RECENT_GIT_MESSAGE="$(git log -1 --pretty=%B)"

        if [[ "$CONFIG_COMMIT_MESSAGE" = "$MOST_RECENT_GIT_MESSAGE" ]]; then
          log 'gitops worked! sweet'
          success=true
          break
        fi

        echo "- waiting $sleep_duration seconds"
        sleep "$sleep_duration"
      done
    popd
  popd

  if [[ "$success" = true ]]; then
    return 0
  else
    log 'FAILED :('
    exit 1
  fi
}

setup_gitops_to_app() {
  log "setting up gitops-to-app"

  ytt --ignore-unknown-comments \
    -f "$DIR/../examples/basic-delivery" \
    --data-value git_writer.repository="http://$GIT_SERVER/$CONFIG_REPO.git" \
    --data-value git_writer.branch="$CONFIG_BRANCH" |
    kapp deploy --yes -a example-delivery -f-
}

test_gitops_to_app() {
  log "testing gitops-to-app"

  for _ in {1..5}; do
    for sleep_duration in {15..1}; do
      local deployed_pods
      deployed_pods=$(kubectl get pods \
        -l "serving.knative.dev/configuration=$TEST_NAME" \
        -o name)

      if [[ "$deployed_pods" == *"$TEST_NAME"* ]]; then
        log "testing '$TEST_NAME' SUCCEEDED! sweet"
        return 0
      fi

      echo "- waiting $sleep_duration seconds"
      sleep "$sleep_duration"
    done

    kubectl tree deliverable gitops
  done

  log "testing gitops-to-app FAILED :("
  exit 1
}

verify_new_carto_deployed() {
  until [[ $(kubectl -n cartographer-system get deployment cartographer-controller -o json | jq '.metadata.labels."app.kubernetes.io/version"') == \"$RELEASE_VERSION\" &&
    $(kubectl -n cartographer-system get deployment cartographer-controller -o json | jq '.status.availableReplicas') != 0 ]]; do
    log "waiting for new cartographer deployment"
    sleep 10
  done
}

update_source(){
  local source_dir=$1
  local port=$2

  log "updating source repo"

  pushd "$source_dir/$SOURCE_REPO"
    if [ "$(uname)" == "Darwin" ]; then
      sed -i '' 's/hello world/hello universe/g' main.go
    else
      sed -i 's/hello world/hello universe/g' main.go
    fi
    git config user.email "gitops-user@example.com"
    git config user.name "Gitops User"
    git add .
    git commit -m "Not a meaningless change"
    git push origin $SOURCE_BRANCH
  popd
}

wait_for_new_config_commit() {
  local config_dir=$1

  log "testing source to gitops part 2"
  log "waiting for new config commit"

  local success=false

  pushd "$config_dir/$CONFIG_REPO"
    for sleep_duration in {20..1}; do
      git pull
      if [[ $(git log --pretty=oneline | wc -l | tr -d ' ') == 2 ]]; then
        log "found new config commit! source to gitops part 2 worked!"
        success=true
        break
      fi

      echo "- waiting $sleep_duration seconds"
      sleep "$sleep_duration"
    done
  popd

  if [[ "$success" = true ]]; then
    return 0
  else
    log 'FAILED :('
    exit 1
  fi
}

wait_for_knative_deployment_update() {
  log "testing gitops to app part 2"

  for sleep_duration in {15..1}; do

    if [[ $(kubectl get deployment -l "serving.knative.dev/configuration=gitwriter-sc" -o name | wc -l | tr -d ' ') == 2 ]]; then
      log "gitops to app part 2 SUCCEEDED! sweet"
      return 0
    fi

    echo "- waiting $sleep_duration seconds"
    sleep "$sleep_duration"
  done

  log "testing gitops-to-app part 2 FAILED :("
  exit 1
}

available_port() {
  python - <<-EOF
import socket

s = socket.socket()
s.bind(('', 0))
print(s.getsockname()[1])
s.close()
EOF
}

log() {
  printf '\n\t\033[1m%s\033[0m\n\n' "$1" 1>&2
}

main "$@"
