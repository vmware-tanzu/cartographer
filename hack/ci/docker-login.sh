#!/usr/bin/env bash

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