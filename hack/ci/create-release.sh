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
echo "Hello"

set -euo pipefail

SCRIPT_PATH=$(dirname $0)

export DEBIAN_FRONTEND=noninteractive
apt-get -y update
apt-get install -y curl

curl -L https://github.com/google/ko/releases/download/v0.8.3/ko_0.8.3_Linux_arm64.tar.gz | tar xzf - ko
mv ko /usr/local/bin/ko
chmod +x /usr/local/bin/ko

mkdir -p ~/.docker
"${SCRIPT_PATH}/docker-login.sh" ~/.docker/config.json

ytt --ignore-unknown-comments -f ./config | ko resolve -f- > release.yaml

echo "# Change Set" > CHANGELOG.md
CURRENT_VERSION="${GITHUB_REF##*/}"
PREVIOUS_VERSION="$(git tag --sort=-v:refname -l 'v*' | grep -A1 "${CURRENT_VERSION}" | tail -n 1)"
git -c log.showSignature=false log --pretty=oneline --abbrev-commit --no-decorate --no-color "${PREVIOUS_VERSION}..${CURRENT_VERSION}" >> CHANGELOG.md

echo "## Checksums" >> CHANGELOG.md
shasum=$(sha256sum release.yaml)
echo '```'"${shasum}"'```' >> CHANGELOG.md
