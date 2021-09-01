#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly DIR=$(cd $(dirname $0) && pwd)
readonly HOST_ADDR=${HOST_ADDR:-$($DIR/ip.py)}
readonly REGISTRY_PORT=${REGISTRY_PORT:-5000}
readonly REGISTRY=${HOST_ADDR}:${REGISTRY_PORT}
readonly KIND_IMAGE=${KIND_IMAGE:-kindest/node:v1.21.1}

readonly REGISTRY_CONTAINER_NAME=cartographer-registry
readonly KUBERNETES_CONTAINER_NAME=cartographer-control-plane
readonly COMMAND=${1:-run}

main() {
        case $COMMAND in
        run)
                display_vars
                start_registry
                start_local_cluster

                generate_cartographer_release
                install_dependencies

                setup_example
                test_example
                ;;
        teardown)
                delete_containers
                ;;
        *)
                echo "error: unknown command '$COMMAND'."
                echo "usage: $0 (run|teardown)."
                exit 1
                ;;
        esac

}

display_vars() {
        echo "Variables:

	DIR:		$DIR
	HOST_ADDR:	$HOST_ADDR
	KIND_IMAGE:	$KIND_IMAGE
	REGISTRY:	$REGISTRY
	REGISTRY_PORT:	$REGISTRY_PORT
	"
}

setup_example() {
        ytt --ignore-unknown-comments \
                -f $DIR/../../examples/source-to-knative-service \
                --data-value registry.server=$REGISTRY \
                --data-value registry.username="" \
                --data-value registry.password="" \
                --data-value image_prefix="$REGISTRY/example-" |
                kapp deploy --yes -a example -f-
}

test_example() {
        log "testing"

        for i in {15..1}; do
                echo "- attempt $i"

                local deployed_pods=$(kubectl get pods \
                        -l 'serving.knative.dev/configuration=dev' \
                        -o name)

                if [[ ! -z "$deployed_pods" ]]; then
                        log "SUCCEEDED! sweet"
                        exit 0
                fi

                sleep $i
        done

        log "FAILED :("
        exit 1
}

start_registry() {
        docker run \
                --detach \
                --name $REGISTRY_CONTAINER_NAME \
                --publish "${REGISTRY_PORT}:5000" \
                registry:2
}

start_local_cluster() {
        log "starting local cluster"

        cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: cartographer
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry]
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
      [plugins."io.containerd.grpc.v1.cri".registry.mirrors."${REGISTRY}"]
        endpoint = ["http://${REGISTRY}"]
    [plugins."io.containerd.grpc.v1.cri".registry.configs]
      [plugins."io.containerd.grpc.v1.cri".registry.configs."${REGISTRY}".tls]
        insecure_skip_verify = true
nodes:
  - role: control-plane
    image: ${KIND_IMAGE}
EOF
}

advertise_local_cluster() {
        log "advertising local cluster"

        cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "${REGISTRY}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF
}

generate_cartographer_release() {
        log "generating release"

        KO_DOCKER_REPO=${REGISTRY}/controller \
                make release
}

install_dependencies() {
        log "installing example dependencies"

        install_cert_manager
        install_cartographer
        install_source_controller
        install_kpack
        install_kapp_controller
        install_knative_serving
}

install_cert_manager() {
        kapp deploy --yes -a cert-manager \
                -f https://github.com/jetstack/cert-manager/releases/download/v1.2.0/cert-manager.yaml
}

install_cartographer() {
        kubectl create namespace cartographer-system
        kapp deploy --yes -a cartographer -f $DIR/../../releases/release.yaml
}

install_source_controller() {
        kubectl create namespace gitops-toolkit

        kubectl create clusterrolebinding gitops-toolkit-admin \
                --clusterrole=cluster-admin \
                --serviceaccount=gitops-toolkit:default

        kapp deploy --yes -a gitops-toolkit \
                --into-ns gitops-toolkit \
                -f https://github.com/fluxcd/source-controller/releases/download/v0.15.3/source-controller.crds.yaml \
                -f https://github.com/fluxcd/source-controller/releases/download/v0.15.3/source-controller.deployment.yaml
}

install_kpack() {
        kapp deploy --yes -a kpack \
                -f https://github.com/pivotal/kpack/releases/download/v0.3.1/release-0.3.1.yaml
}

install_kapp_controller() {
        kubectl create clusterrolebinding default-admin \
                --clusterrole=cluster-admin \
                --serviceaccount=default:default

        kapp deploy --yes -a kapp-controller \
                -f https://github.com/vmware-tanzu/carvel-kapp-controller/releases/download/v0.22.0/release.yml
}

install_knative_serving() {
        ytt --ignore-unknown-comments \
                -f https://github.com/knative/serving/releases/download/v0.25.0/serving-core.yaml \
                -f https://github.com/knative/serving/releases/download/v0.25.0/serving-crds.yaml \
                -f $DIR/overlays/remove-resource-requests-from-deployments.yaml |
                kapp deploy --yes -a knative-serving -f-
}

delete_containers() {
        docker rm -f $REGISTRY_CONTAINER_NAME || true
        docker rm -f $KUBERNETES_CONTAINER_NAME || true
}

log() {
        printf '\n\t\033[1m%s\033[0m\n\n' "$1" 1>&2
}

main "$@"
