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
set -o nounset

readonly DOCKER_REGISTRY=${DOCKER_REGISTRY:-"https://index.docker.io/v1/"}
readonly DOCKER_CONFIG=${DOCKER_CONFIG:-"$(realpath ~/.docker/config.json)"}
readonly DOCKER_USERNAME=$DOCKER_USERNAME
readonly DOCKER_PASSWORD=$DOCKER_PASSWORD

main() {
        local auth=$(basic_auth $DOCKER_USERNAME $DOCKER_PASSWORD)

        update_or_create_docker_config $DOCKER_REGISTRY $auth $DOCKER_CONFIG
}

update_or_create_docker_config() {
        local server=$1
        local token=$2
        local config_file=$3

        readonly qry='.auths["%s"] = {"auth": "%s"}'

        local current_content=$(jq '.' $config_file)
        if [[ "$current_content" == "" ]]; then
                echo '{}' >$config_file
        fi

        config=$(jq "$(printf "$qry" $server $token)" < "$config_file")
        echo "$config" > $config_file
}

basic_auth() {
        local username=$1
        local password=$2

        printf '%s:%s' $username $password | base64
}

main "$@"
