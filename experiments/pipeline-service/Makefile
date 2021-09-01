build:
	cd cmd/controller && go build -v
debug:
	cd cmd/controller && dlv debug
run: build
	cmd/controller/controller
lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run
install-crds:
	kapp deploy -a pipeline-service -f ./config/crd


release:
	ytt --ignore-unknown-comments -f ./config | \
		ko resolve -f- > ./releases/release.yaml


generate-objects:
	go run sigs.k8s.io/controller-tools/cmd/controller-gen \
                object \
                paths=./pkg/apis/v1alpha1
generate-manifests:
	go run sigs.k8s.io/controller-tools/cmd/controller-gen \
		crd \
		paths=./pkg/apis/v1alpha1 \
		output:crd:artifacts:config=config/crd/bases
