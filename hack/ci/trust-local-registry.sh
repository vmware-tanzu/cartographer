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

readonly local_registry=$(ip route get 8.8.8.8 | grep src | awk '{print $7}'):5001
readonly config_file=/etc/docker/daemon.json

echo "$(jq ". + {\"insecure-registries\": [\"$local_registry\"]}" $config_file)" >$config_file

# wokeignore:rule=kill no alternatives in *NIX
kill -s SIGHUP $(pidof dockerd)
