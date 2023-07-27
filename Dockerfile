ARG BASE_IMAGE=northamerica-northeast2-docker.pkg.dev/kontinue/library/ubuntu:jammy
ARG GOLANG_IMAGE=northamerica-northeast2-docker.pkg.dev/kontinue/library/golang:1.19

FROM ${BASE_IMAGE} AS base

RUN set -x && \
	apt-get update && \
	apt-get upgrade -y && \
	apt-get install  -y \
	jq curl git ca-certificates gnutls-bin && \
	rm -rf /var/lib/apt/lists/*

FROM base AS ytt

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