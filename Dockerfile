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
# docker build -t cartographer:dev --build-arg BASE_IMAGE="ubuntu:jammy" --build-arg GOLANG_IMAGE=golang:1.19 .

ARG BASE_IMAGE=harbor-repo.vmware.com/dockerhub-proxy-cache/library/ubuntu:jammy
ARG GOLANG_IMAGE=harbor-repo.vmware.com/dockerhub-proxy-cache/library/golang:1.20

FROM ${BASE_IMAGE} AS ytt

RUN set -x && \
	apt-get update && \
	apt-get install -y curl=7.81.0-1ubuntu1.15

ARG ytt_CHECKSUM=a6729fb8514f10ab58f9ed3b50cd90ef79bf16d1cb29173baa84e1af0bc5ad4f
ARG ytt_VERSION=0.45.3

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

FROM gcr.io/paketo-buildpacks/run-jammy-tiny@sha256:35702d19f93e06041db1573b1140742df2182494cc93f646fd57c6d8922dc7a7
COPY --from=ytt 	/usr/local/bin/ytt	/usr/local/bin/ytt
COPY --from=cartographer /build/cartographer	/usr/local/bin/cartographer
ENTRYPOINT [ "cartographer" ]