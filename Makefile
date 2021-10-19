.PHONY: build
build: gen-objects gen-manifests
	go build -o build/cartographer ./cmd/cartographer

.PHONY: run
run: build
	build/cartographer

crd_non_sources := pkg/apis/v1alpha1/zz_generated.deepcopy.go $(wildcard pkg/apis/v1alpha1/*_test.go)
crd_sources := $(filter-out $(crd_non_sources),$(wildcard pkg/apis/v1alpha1/*.go))

pkg/apis/v1alpha1/zz_generated.deepcopy.go: $(crd_sources)
	go run sigs.k8s.io/controller-tools/cmd/controller-gen \
                object \
                paths=./pkg/apis/v1alpha1

config/crd/bases/*.yaml &: $(crd_sources)
	go run sigs.k8s.io/controller-tools/cmd/controller-gen \
		crd \
		paths=./pkg/apis/v1alpha1 \
		output:crd:artifacts:config=config/crd/bases
	go run github.com/google/addlicense \
		-f ./hack/boilerplate.go.txt \
		config/crd/bases

.PHONY: gen-objects
gen-objects: pkg/apis/v1alpha1/zz_generated.deepcopy.go

.PHONY: gen-manifests
gen-manifests: config/crd/bases/*.yaml

test_crd_sources := $(wildcard tests/resources/*.go)
test_object_sources := $(filter-out tests/resources/zz_generated.deepcopy.go,$(test_crd_sources))

tests/resources/zz_generated.deepcopy.go: $(test_object_sources)
	go run sigs.k8s.io/controller-tools/cmd/controller-gen \
                object \
                paths=./tests/resources

.PHONY: test-gen-objects
test-gen-objects: tests/resources/zz_generated.deepcopy.go

test_crds := tests/resources/test.go
tests/resources/crds/test.run_tests.yaml: $(test_crds) tests/resources/groupversion_info.go
	go run sigs.k8s.io/controller-tools/cmd/controller-gen \
		crd \
		paths=./tests/resources \
		output:crd:artifacts:config=./tests/resources/crds
	go run github.com/google/addlicense \
		-f ./hack/boilerplate.go.txt \
		./tests/resources/crds

.PHONY: test-gen-manifests
test-gen-manifests: tests/resources/crds/*

.PHONY: clean-fakes
clean-fakes:
	find . -type d -name  '*fakes' | xargs -n1 rm -r

.PHONY: generate
generate: clean-fakes
	go generate ./...

.PHONY: test-unit
test-unit: test-gen-objects
	go run github.com/onsi/ginkgo/ginkgo -r pkg

.PHONY: test-integration
test-integration: test-gen-manifests test-gen-objects
	go run github.com/onsi/ginkgo/ginkgo -r tests/integration

.PHONY: test-kuttl
test-kuttl: build test-gen-manifests
	if [ -n "$$focus" ]; then kubectl kuttl test --test $$(basename $(focus)); else kubectl kuttl test; fi

.PHONY: test-kuttl-pipeline-service
test-kuttl-pipeline-service: build test-gen-manifests
	if [ -n "$$focus" ]; then kubectl kuttl test ./tests/kuttl/pipeline-service --test $$(basename $(focus)); else kubectl kuttl test ./tests/kuttl/pipeline-service; fi

.PHONY: test-kuttl-supply-chain
test-kuttl-supply-chain: build test-gen-manifests
	if [ -n "$$focus" ]; then kubectl kuttl test ./tests/kuttl/supply-chain --test $$(basename $(focus)); else kubectl kuttl test ./tests/kuttl/supply-chain; fi

.PHONY: test-kuttl-deliverable
test-kuttl-deliverable: build test-gen-manifests
	if [ -n "$$focus" ]; then kubectl kuttl test ./tests/kuttl/deliverable --test $$(basename $(focus)); else kubectl kuttl test ./tests/kuttl/deliverable; fi


.PHONY: list-kuttl
list-kuttl:
	(cd tests/kuttl && find . -maxdepth 2 -type d)

.PHONY: test-kuttl-kind
test-kuttl-kind: build
	kubectl kuttl test --start-kind=true --start-control-plane=false --artifacts-dir=/dev/null

.PHONY: test
test: test-unit test-kuttl test-integration

.PHONY: install
install:
	kapp deploy --file ./config/crd --app cartographer-controller --yes --diff-changes

.PHONY: uninstall
uninstall:
	kapp delete --app cartographer-controller --yes

.PHONY: coverage
coverage:
	go test -coverprofile=coverage.out ./pkg/...
	go tool cover -func=./coverage.out
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html

.PHONY: lint
lint: copyright
	go run github.com/golangci/golangci-lint/cmd/golangci-lint --config lint-config.yaml run
	$(MAKE) -C hack lint

.PHONY: copyright
copyright:
	go run github.com/google/addlicense \
		-f ./hack/boilerplate.go.txt \
		-ignore site/static/\*\* \
		-ignore site/themes/\*\* \
		.
