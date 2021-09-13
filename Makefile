build: gen-objects gen-manifests
	go build -o build/cartographer ./cmd/cartographer
run: build
	build/cartographer

gen-objects:
	go run sigs.k8s.io/controller-tools/cmd/controller-gen \
                object \
                paths=./pkg/apis/v1alpha1
gen-manifests:
	go run sigs.k8s.io/controller-tools/cmd/controller-gen \
		crd \
		paths=./pkg/apis/v1alpha1 \
		output:crd:artifacts:config=config/crd/bases
	go run github.com/google/addlicense \
		-f ./hack/boilerplate.go.txt \
		config/crd/bases

test-gen-objects:
	go run sigs.k8s.io/controller-tools/cmd/controller-gen \
                object \
                paths=./tests/integration/pipeline_service/testapi

test-gen-manifests:
	go run sigs.k8s.io/controller-tools/cmd/controller-gen \
		crd \
		paths=./tests/integration/pipeline_service/testapi \
		output:crd:artifacts:config=./tests/integration/pipeline_service/testapi/crds
	go run github.com/google/addlicense \
		-f ./hack/boilerplate.go.txt \
		./tests/integration/pipeline_service/testapi/crds

clean-fakes:
	find . -type d -name  '*fakes' | xargs -n1 rm -r

generate: clean-fakes
	go generate ./...

test-unit:
	go run github.com/onsi/ginkgo/ginkgo -r pkg

test-integration:
	go run github.com/onsi/ginkgo/ginkgo -r tests/integration

test-kuttl: build
	if [ -n "$$focus" ]; then kubectl kuttl test --test $$(basename $(focus)); else kubectl kuttl test; fi

list-kuttl:
	(cd tests/kuttl && find . -maxdepth 1 -type d)

test-kuttl-kind: build
	kubectl kuttl test --start-kind=true --start-control-plane=false --artifacts-dir=/dev/null

test: test-unit test-kuttl test-integration

install:
	kapp deploy --file ./config/crd --app cartographer-controller --yes --diff-changes
uninstall:
	kapp delete --app cartographer-controller --yes

coverage:
	go test -coverprofile=coverage.out ./pkg/...
	go tool cover -func=./coverage.out
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint --config lint-config.yaml run

release: gen-manifests
	mkdir -p releases
	ytt --ignore-unknown-comments -f ./config | ko resolve -f- > ./releases/release.yaml
	kbld -f releases/release.yaml --imgpkg-lock-output releases/.imgpkg/images.yml
	imgpkg push -b projectcartographer/cartographer-bundle -f releases --file-exclusion releases/kbld.lock.yml --lock-output releases/kbld.lock.yml
	go run github.com/google/addlicense \
		-f ./hack/boilerplate.go.txt \
		releases

install-cert-manager:
	kapp deploy --yes -a cert-manager \
		-f https://github.com/jetstack/cert-manager/releases/download/v1.2.0/cert-manager.yaml

prep-deploy: install-cert-manager
	kubectl create ns cartographer-system || true

deploy: prep-deploy
	kapp deploy --yes -a cartographer -f ./releases/release.yaml

tear-down-local:
	./hack/local-dev/delete-local-cluster-and-registry.sh

create-local:
	./hack/local-dev/create-local-cluster-and-registry.sh

deploy-local: create-local prep-deploy gen-manifests
	ytt -f ./config --ignore-unknown-comments | \
		KO_DOCKER_REPO=localhost:65432 ko resolve -f- | \
		kapp deploy --yes -a cartographer -f -

copyright:
	go run github.com/google/addlicense \
		-f ./hack/boilerplate.go.txt \
		-ignore site/static/\*\* \
		-ignore site/themes/\*\* \
		.
