#!/usr/bin/env bash

# Copyright 2017 The Kubernetes Authors.
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

#CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ./cmd/code-generator)}
GENERATOR_VERSION=v0.22.4
(
  # To support running this script from anywhere, we have to first cd into this directory
  # so we can install the tools.
  cd "$(dirname "${0}")"
  go install k8s.io/code-generator/cmd/{defaulter-gen,client-gen,lister-gen,informer-gen,deepcopy-gen}@$GENERATOR_VERSION
)

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")
echo $SCRIPT_ROOT

rm -rf "${SCRIPT_ROOT}"/../pkg/generated
# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
bash hack/generate-groups.sh all \
  github.com/vmware-tanzu/cartographer/pkg/generated  github.com/vmware-tanzu/cartographer/pkg/apis \
  carto:v1alpha1 \
  --output-base "$(dirname "${BASH_SOURCE[0]}")/../../../.." \
  --go-header-file "${SCRIPT_ROOT}"/boilerplate.go.txt 
