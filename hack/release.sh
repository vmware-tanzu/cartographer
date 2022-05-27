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

ROOT=$(cd "$(dirname $0)"/.. && pwd)
readonly ROOT

readonly SCRATCH=${SCRATCH:-$(mktemp -d)}
readonly REGISTRY=${REGISTRY:-"$($ROOT/hack/ip.py):5000"}
readonly RELEASE_DATE=${RELEASE_DATE:-$(TZ=UTC date +"%Y-%m-%dT%H:%M:%SZ")}

readonly YTT_VERSION=0.39.0
readonly YTT_CHECKSUM=7a472b8c62bfec5c12586bb39065beda42c6fe43cf24271275e4dbc0a04acb8b

main() {
        readonly RELEASE_VERSION=${RELEASE_VERSION:-"v0.0.0-dev"}
        readonly PREVIOUS_VERSION=${PREVIOUS_VERSION:-$(git_previous_version $RELEASE_VERSION)}

        show_vars
        cd $ROOT

        download_ytt_to_kodata
        generate_release
        create_release_notes
}

show_vars() {
        echo "
        PREVIOUS_VERSION:       $PREVIOUS_VERSION
        REGISTRY:               $REGISTRY
        RELEASE_DATE:           $RELEASE_DATE
        RELEASE_VERSION:        $RELEASE_VERSION
        ROOT:                   $ROOT
        SCRATCH:                $SCRATCH
        YTT_VERSION:            $YTT_VERSION
        "
}

download_ytt_to_kodata() {
        local url=https://github.com/vmware-tanzu/carvel-ytt/releases/download/v${YTT_VERSION}/ytt-linux-amd64
        local fname=ytt-linux-amd64

        local dest
        dest=$(realpath ./cmd/cartographer/kodata/$fname)

        test -x $dest && echo "${YTT_CHECKSUM}  $dest" | sha256sum -c && {
                echo "ytt already found in kodata."
                return
        }

        pushd "$(mktemp -d)"
        curl -sSOL $url
        echo "${YTT_CHECKSUM}  $fname" | sha256sum -c
        install -m 0755 $fname $dest
        popd
}

generate_release() {
        mkdir -p ./release
        ytt --ignore-unknown-comments -f ./config \
                -f ./hack/overlays/webhook-configuration.yaml \
                -f ./hack/overlays/component-labels.yaml \
                --data-value version=$RELEASE_VERSION |
                KO_DOCKER_REPO=$REGISTRY ko resolve -B -f- > \
                        ./release/cartographer.yaml
}

create_release_notes() {
        local changeset
        changeset="$(git_changeset $RELEASE_VERSION $PREVIOUS_VERSION)"

        local assets_checksums
        assets_checksums=$(checksums ./release)

        release_body "$changeset" "$assets_checksums" "$PREVIOUS_VERSION" >./release/CHANGELOG.md
}

checksums() {
        local assets_directory=$1

        pushd $assets_directory &>/dev/null
        find . -name "*" -type f -exec sha256sum {} +
        popd &>/dev/null
}

git_changeset() {
        local current_version=$1
        local previous_version=$2

        [[ $current_version != v* ]] && current_version=v$current_version
        [[ $previous_version != v* ]] && previous_version=v$previous_version
        [[ $(git tag -l $current_version) == "" ]] && current_version=HEAD

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
        version_filter=$(printf '^%s$' $current_version)

        [[ $(git tag -l $current_version) == "" ]] && version_filter='.'

        git tag --sort=-v:refname -l |
                grep -A30 $version_filter |
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
        printf "$fmt" "$previous_version" "$changeset" "$checksums"
}

main "$@"
