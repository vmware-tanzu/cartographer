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

# These base images are behind a proxy for rate-limit reasons
# If building locally, you can simply use:
# docker build -t cartographer:dev --build-arg BASE_IMAGE="ubuntu:jammy" --build-arg GOLANG_IMAGE=golang:1.21 .

ARG BASE_IMAGE=harbor-repo.vmware.com/dockerhub-proxy-cache/library/ubuntu:jammy
ARG GOLANG_IMAGE=harbor-repo.vmware.com/dockerhub-proxy-cache/library/golang:1.21

FROM ${BASE_IMAGE} AS ytt

RUN set -x && \
	apt-get update && \
	apt-get install -y curl=7.81.0-1ubuntu1.16

ARG ytt_CHECKSUM=357ec754446b1eda29dd529e088f617e85809726c686598ab03cfc1c79f43b56
ARG ytt_VERSION=0.49.0

RUN set -eux && \
	url=https://github.com/vmware-tanzu/carvel-ytt/releases/download/v${ytt_VERSION}/ytt-linux-amd64 ; \
	curl -sSL $url -o /usr/local/bin/ytt && \
	echo "${ytt_CHECKSUM}  /usr/local/bin/ytt" | sha256sum -c && \
	chmod +x /usr/local/bin/ytt

FROM ${GOLANG_IMAGE} AS cartographer
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY pkg/ ./pkg/
RUN GOOS=linux GOARCH=amd64 go build -o /build/ github.com/vmware-tanzu/cartographer/cmd/cartographer

FROM gcr.io/paketo-buildpacks/run-jammy-tiny@sha256:ab8cdab34ea0c71f408ab354b0234ab4fd6d6a7d0660ca66993784fc4daa3fb2
COPY --from=ytt 	/usr/local/bin/ytt	/usr/local/bin/ytt
COPY --from=cartographer /build/cartographer	/usr/local/bin/cartographer
ENTRYPOINT [ "cartographer" ]