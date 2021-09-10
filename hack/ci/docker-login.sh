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
set -o pipefail

readonly DOCKER_USER=${DOCKER_USER:-"USER MUST BE SUPPLIED"}
readonly DOCKER_PASSWORD=${DOCKER_PASSWORD:-"PASSWORD MUST BE SUPPLIED"}
readonly DOCKER_REGISTRY=${DOCKER_REGISTRY:-"https://index.docker.io/v1/"}

main() {
        if [ $# -ne 1 ]; then
                echo "usage: $0 dest"
                echo "env vars: DOCKER_USER, DOCKER_PASSWORD, DOCKER_REGISTRY"
                exit 1
        fi

        set -o nounset

        local destination=$1
        local basic_auth=$(printf '%s:%s' $DOCKER_USER $DOCKER_PASSWORD | base64)

        echo "{
        \"auths\": {
                \"$DOCKER_REGISTRY\": {
                        \"auth\": "\"$basic_auth\""
                }
        }
}" >$destination
}

main "$@"