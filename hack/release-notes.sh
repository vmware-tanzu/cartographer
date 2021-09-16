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

readonly ROOT=$(cd $(dirname $0)/.. && pwd)
readonly ASSETS_DIR=${ASSETS_DIR:-$(realpath $ROOT/release)}
readonly DESTINATION=${DESTINATION:-$ASSETS_DIR/CHANGELOG.md}

main() {
        readonly CURRENT_VERSION=${CURRENT_VERSION:-$(git_current_version)}
        readonly PREVIOUS_VERSION=${PREVIOUS_VERSION:-$(git_previous_version $CURRENT_VERSION)}

        display_vars

        local changeset="$(git_changeset $CURRENT_VERSION $PREVIOUS_VERSION)"
        local assets_checksums=$(checksums $ASSETS_DIR)

        release_body "$changeset" "$assets_checksums" >$DESTINATION
}

display_vars() {
        echo "Variables:

	ASSETS_DIR		$ASSETS_DIR
	CURRENT_VERSION		$CURRENT_VERSION
	DESTINATION		$DESTINATION
	PREVIOUS_VERSION	$PREVIOUS_VERSION
	ROOT			$ROOT
	"
}

checksums() {
        local assets_directory=$1

        pushd $assets_directory &>/dev/null
        find . -name "*.yaml" -type f | xargs sha256sum
        popd &>/dev/null
}

git_changeset() {
        local current_version=$1
        local previous_version=$2

        git -c log.showSignature=false \
                log \
                --pretty=oneline \
                --abbrev-commit \
                --no-decorate \
                --no-color \
                "${previous_version}..${current_version}"
}

git_current_version() {
        git tag --sort=-v:refname -l 'v*' |
                head -n1
}

git_previous_version() {
        local current_version=$1

        git tag --sort=-v:refname -l 'v*' |
                grep -A1 $current_version |
                tail -n1
}

release_body() {
        local changeset="$1"
        local checksums="$2"

        readonly fmt='
# Change Set

%s


# Installation

See https://github.com/vmware-tanzu/cartographer#installation.


# Checksums

```
%s
```
'
        printf "$fmt" "$changeset" "$checksums"
}

main "$@"
