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
readonly KO_CHECKSUM=9f004fa1c2b55ac765ec0c287ad0311a517a86299b7a633bc542f2fbbb3a4ea4
readonly KUBERNETES_VERSION=1.19.2
readonly KUBERNETES_CHECKSUM=fb13a93a800389029b06fcc74ab6a3b969ff74178252709a040e4756251739d2
readonly KUTTL_VERSION=0.11.1
readonly KUTTL_CHECKSUM=0fb13f8fbb6109803a06847a8ad3fae4fedc8cd159e2b0fd6c1a1d8737191e5f
readonly GH_VERSION=2.0.0
readonly GH_CHECKSUM=20c2d1b1915a0ff154df453576d9e97aab709ad4b236ce8313435b8b96d31e5c

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
        local fname=ko_${KO_VERSION}_Linux_x86_64.tar.gz

        curl -sSOL $url
        echo "${KO_CHECKSUM} $fname" | sha256sum -c
        tar xzf $fname

        install -m 0755 ./ko /usr/local/bin
}

install_kubebuilder() {
        local url=https://storage.googleapis.com/kubebuilder-tools/kubebuilder-tools-${KUBERNETES_VERSION}-linux-amd64.tar.gz
        local fname=kubebuilder-tools-${KUBERNETES_VERSION}-linux-amd64.tar.gz

        curl -sSOL $url
        echo "${KUBERNETES_CHECKSUM} $fname" | sha256sum -c
        tar xvzf $fname

        mv ./kubebuilder /usr/local
}

install_kuttl() {
        local url=https://github.com/kudobuilder/kuttl/releases/download/v${KUTTL_VERSION}/kubectl-kuttl_${KUTTL_VERSION}_linux_x86_64
        local fname=kubectl-kuttl_${KUTTL_VERSION}_linux_x86_64

        curl -sSOL $url
        echo "${KUTTL_CHECKSUM} $fname" | sha256sum -c

        install -m 0755 $fname /usr/local/bin/kubectl-kuttl
}

install_gh() {
        local url=https://github.com/cli/cli/releases/download/v${GH_VERSION}/gh_${GH_VERSION}_linux_amd64.tar.gz
        local fname=gh_${GH_VERSION}_linux_amd64.tar.gz

        curl -sSOL $url
        echo "${GH_CHECKSUM} $fname" | sha256sum -c
        tar xzf $fname --strip-components=1

        mv ./bin/gh /usr/local/bin
}

main "$@"
