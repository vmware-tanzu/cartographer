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

readonly KO_VERSION=0.8.1
readonly KUBERNETES_VERSION=1.19.2
readonly KUTTL_VERSION=0.11.1
readonly GH_VERSION=2.0.0

main() {
        cd $(mktemp -d)

        for binary in $@; do
                case $binary in
                ko)
                        install_ko
                        ;;
                kubebuilder)
                        install_kubebuilder
                        ;;
                kuttl)
                        install_kuttl
                        ;;
                gh)
                        install_gh
                        ;;
                *) ;;
                esac
        done

}

install_ko() {
        local url=https://github.com/google/ko/releases/download/v${KO_VERSION}/ko_${KO_VERSION}_Linux_x86_64.tar.gz

        curl -sSL $url | tar -xzf -
        sudo install -m 0755 ./ko /usr/local/bin
}

install_kubebuilder() {
        local url=https://storage.googleapis.com/kubebuilder-tools/kubebuilder-tools-${KUBERNETES_VERSION}-linux-amd64.tar.gz

        curl -sSL $url | tar xzf -
        sudo mv ./kubebuilder /usr/local
        sudo chown -R $(whoami) /usr/local/kubebuilder
}

install_kuttl() {
        local url=https://github.com/kudobuilder/kuttl/releases/download/v${KUTTL_VERSION}/kubectl-kuttl_${KUTTL_VERSION}_linux_x86_64

        curl -sSL -o kubectl-kuttl $url
        sudo install -m 0755 ./kubectl-kuttl /usr/local/bin
}

install_gh() {
	local url=https://github.com/cli/cli/releases/download/v${GH_VERSION}/gh_${GH_VERSION}_linux_amd64.tar.gz

	curl -sSL $url | tar xzf - --strip-components=1
	sudo mv ./bin/gh /usr/local/bin
}

main "$@"
