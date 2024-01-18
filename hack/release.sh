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

ROOT=$(cd "$(dirname "$0")"/.. && pwd)
readonly ROOT

readonly SCRATCH=${SCRATCH:-$(mktemp -d)}
readonly REGISTRY=${REGISTRY:-"$("$ROOT"/hack/ip.py):5001"}
readonly RELEASE_DATE=${RELEASE_DATE:-$(TZ=UTC date +"%Y-%m-%dT%H:%M:%SZ")}

main() {
        readonly RELEASE_VERSION=${RELEASE_VERSION:-"v0.0.0-dev"}
        readonly RELEASE_IMAGE=${RELEASE_IMAGE:-$REGISTRY/cartographer:$RELEASE_VERSION}
        readonly PREVIOUS_VERSION=${PREVIOUS_VERSION:-$(git_previous_version "$RELEASE_VERSION")}

        readonly RELEASE_USING_LEVER=${RELEASE_USING_LEVER:-false}
        readonly LEVER_COMMIT_REF=${LEVER_COMMIT_REF:-"$(git rev-parse HEAD)"}
        readonly LEVER_KUBECONFIG=${LEVER_KUBECONFIG:-""}

        show_vars
        cd "$ROOT"

        if [[ "$RELEASE_USING_LEVER" == true ]]; then
        # Lever build flow
                if [[ "$REGISTRY" == "192.168."* ]]; then
                        echo "REGISTRY must be set to a registry accessible by lever when RELEASE_USING_LEVER is true"
                        exit 1
                fi
                if [[ "$LEVER_KUBECONFIG" == "" ]]; then
                        echo "LEVER_KUBECONFIG must be set when RELEASE_USING_LEVER is true"
                        exit 1
                fi
                echo "Building using lever"
                lever_build_request
        else
                echo "Building locally"
                build_image
        fi
        generate_release
        create_release_notes
}

show_vars() {
        echo "
        PREVIOUS_VERSION:       $PREVIOUS_VERSION
        REGISTRY:               $REGISTRY
        RELEASE_DATE:           $RELEASE_DATE
        RELEASE_VERSION:        $RELEASE_VERSION
        RELEASE_IMAGE:          $RELEASE_IMAGE
        ROOT:                   $ROOT
        SCRATCH:                $SCRATCH
        RELEASE_USING_LEVER:    $RELEASE_USING_LEVER
        LEVER_KUBECONFIG:       $LEVER_KUBECONFIG
        LEVER_COMMIT_REF:       $LEVER_COMMIT_REF
        "
}

lever_build_request() {
        # Lever build request expects LEVER_KUBECONFIG to be the kubeconfig yaml to access the lever cluster
        BUILD_SUFFIX="$(git rev-parse HEAD | head -c 6)-$(echo $RANDOM | shasum | head -c 6; echo)"
        ytt --ignore-unknown-comments -f ./hack/lever_build_request.yaml \
        --data-value build_suffix="$BUILD_SUFFIX" \
        --data-value commit_ref="$LEVER_COMMIT_REF" \
        --data-value release_image="$RELEASE_IMAGE" \
        | kubectl --kubeconfig <(printf '%s' "${LEVER_KUBECONFIG}") apply -f -
        wait_for_lever_build "cartographer-$BUILD_SUFFIX"
}

wait_for_lever_build() {
        local build_name=$1
        local conditions_json=""
        local components_status="-- "
        local build_status="-- "
        local srp_status="-- "
        local ready_status="-- "

        local counter=1

        echo "Waiting for lever build $build_name to complete..."
        # Lever build request has a few build statuses with the final status being the aggregate and the one we wait on
        while [[ $ready_status != 'False' && $ready_status != 'True' ]]; do
                conditions_json=$(kubectl --kubeconfig <(printf '%s' "${LEVER_KUBECONFIG}") get request/"$build_name" -o jsonpath='{.status.conditions}')
                components_status=$(echo "$conditions_json" | jq -r 'map(select(.type == "ComponentsReady"))[0].status')
                build_status=$(echo "$conditions_json" | jq -r 'map(select(.type == "BuildReady"))[0].status')
                srp_status=$(echo "$conditions_json" | jq -r 'map(select(.type == "SRPResourceSubmitted"))[0].status')
                ready_status=$(echo "$conditions_json" | jq -r 'map(select(.type == "Ready"))[0].status')
                loading_char=$(printf "%${counter}s")
                printf "ComponentsReady: %s; BuildReady: %s; SRPResourceSubmitted: %s; Ready: %s; ${loading_char// /.}\033[0K\r" "$components_status" "$build_status" "$srp_status" "$ready_status"
                counter=$((counter + 1))
                if [[ $counter -gt 3 ]]; then
                        counter=1
                fi
                sleep 2
        done

        if [[ $ready_status == 'False' ]]; then
                echo "Lever build $build_name failed"
                ready_message=$(echo "$conditions_json" | jq 'map(select(.type == "Ready"))[0].message')
                echo "Error: $ready_message"
                exit 1
        else
                # Output here is being parsed by the release pipeline to pass references to package-for-cartographer and catalog
                echo "Lever build $build_name succeeded. Image published:"
                kubectl --kubeconfig <(printf '%s' "${LEVER_KUBECONFIG}") get request/"$build_name" -o jsonpath='{.status.artifactStatus.images[0].name}'
                echo ""
                kubectl --kubeconfig <(printf '%s' "${LEVER_KUBECONFIG}") get request/"$build_name" -o jsonpath='{.status.artifactStatus.images[0].image.tag}'
        fi
}

build_image() {
        # Build the image locally using Docker instead of ko
        docker build "$ROOT" -t "$RELEASE_IMAGE" --build-arg BASE_IMAGE="ubuntu:jammy" --build-arg GOLANG_IMAGE="golang:1.20"
        docker push "$RELEASE_IMAGE"
}

generate_release() {
        mkdir -p ./release
        ytt --ignore-unknown-comments -f ./config \
                -f ./hack/overlays/webhook-configuration.yaml \
                -f ./hack/overlays/component-labels.yaml \
                --data-value version="$RELEASE_VERSION" \
                --data-value controller_image="$RELEASE_IMAGE" > ./release/cartographer.yaml
}

create_release_notes() {
        local changeset
        changeset="$(git_changeset "$RELEASE_VERSION" "$PREVIOUS_VERSION")"

        local assets_checksums
        assets_checksums=$(checksums ./release)

        release_body "$changeset" "$assets_checksums" "$PREVIOUS_VERSION" >./release/CHANGELOG.md
}

checksums() {
        local assets_directory=$1

        pushd "$assets_directory" &>/dev/null
        find . -name "*" -type f -exec sha256sum {} +
        popd &>/dev/null
}

git_changeset() {
        local current_version=$1
        local previous_version=$2

        [[ "$current_version" != v* ]] && current_version="v${current_version}"
        [[ "$previous_version" != v* ]] && previous_version="v${previous_version}"
        [[ $(git tag -l "$current_version") == "" ]] && current_version=HEAD

        git -c log.showSignature=false \
                log \
                --pretty=oneline \
                --abbrev-commit \
                --no-decorate \
                --no-color \
                "${previous_version}..${current_version}"
}

git_previous_version() {
        local current_version=$1

        local version_filter
        version_filter=$(printf '^%s$' "$current_version")

        [[ $(git tag -l "$current_version") == "" ]] && version_filter='.'

        git tag --sort=-v:refname -l |
                grep -A30 "$version_filter" |
                grep -E '^v[[:digit:]]+\.[[:digit:]]+\.[[:digit:]]+$' |
                head -n1
}

release_body() {
        local changeset="$1"
        local checksums="$2"
        local previous_version="$3"

        readonly fmt='
# 😎 Easy Installation

```
kubectl apply -f https://github.com/vmware-tanzu/cartographer/releases/download/<NEW_TAG>/cartographer.yaml
```

# 🚨 Breaking Changes

- <REPLACE_ME>

# 🚀 New Features

- <REPLACE_ME>

# 🐛 Bug Fixes

- <REPLACE_ME>

# ❤️ Thanks

Thanks to these contributors who contributed to <NEW_TAG>!
- <REPLACE_ME>

**Full Changelog**: https://github.com/vmware-tanzu/cartographer/compare/%s...<NEW_TAG>

# Change Set

%s


# Checksums

```
%s
```
  '
        # wokeignore:rule=disable
        # shellcheck disable=SC2059
        printf "$fmt" "$previous_version" "$changeset" "$checksums"
}

main "$@"
