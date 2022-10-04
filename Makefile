CONTROLLER_GEN ?= go run -modfile hack/tools/go.mod sigs.k8s.io/controller-tools/cmd/controller-gen
ADDLICENSE ?= go run -modfile hack/tools/go.mod github.com/google/addlicense
GOLANGCI_LINT ?= go run -modfile hack/tools/go.mod github.com/golangci/golangci-lint/cmd/golangci-lint
GINKGO ?= go run -modfile hack/tools/go.mod github.com/onsi/ginkgo/ginkgo
GCI_LINT ?= go run -modfile hack/tools/go.mod github.com/daixiang0/gci
CONTROL_PLANE_KUTTL ?= kubectl kuttl test --control-plane-config=kuttl-control-plane.config
WOKE ?= go run -modfile hack/tools/go.mod github.com/get-woke/woke
UNAME := $(shell uname)

ifndef ($LOG_LEVEL)
	# set a default LOG_LEVEL whenever we run the controller
	# and for our kuttl tests which require something to be set.
	export LOG_LEVEL = info
endif

.PHONY: build
build: gen-objects gen-manifests
	go build -o build/cartographer ./cmd/cartographer

.PHONY: build-cartotest
build-cartotest:
	GOOS=darwin GOARCH=arm64 go build -o build/cartotest_darwin_arm64 ./cmd/cartotest
	GOOS=darwin GOARCH=amd64 go build -o build/cartotest_darwin_amd64 ./cmd/cartotest
	GOOS=linux GOARCH=amd64 go build -o build/cartotest_linux_amd64 ./cmd/cartotest
	GOOS=windows GOARCH=amd64 go build -o build/cartotest_windows_amd64 ./cmd/cartotest

.PHONY: run
run: build
	build/cartographer --pprof-port 9999 --metrics-port 9998

crd_non_sources := pkg/apis/v1alpha1/zz_generated.deepcopy.go $(wildcard pkg/apis/v1alpha1/*_test.go)
crd_sources := $(filter-out $(crd_non_sources),$(wildcard pkg/apis/v1alpha1/*.go))

pkg/apis/v1alpha1/zz_generated.deepcopy.go: $(crd_sources)
	$(CONTROLLER_GEN) \
                object \
                paths=./pkg/apis/v1alpha1

config/crd/bases/*.yaml &: $(crd_sources)
	$(CONTROLLER_GEN) \
		crd \
		paths=./pkg/apis/v1alpha1 \
		output:crd:artifacts:config=config/crd/bases
	$(ADDLICENSE) \
		-f ./hack/boilerplate.go.txt \
		config/crd/bases

config/webhook/manifests.yaml: $(crd_sources)
	$(CONTROLLER_GEN) \
		webhook \
		paths=./pkg/apis/v1alpha1 \
		output:webhook:dir=config/webhook
	$(ADDLICENSE) \
		-f ./hack/boilerplate.go.txt \
		config/webhook/manifests.yaml

.PHONY: gen-objects
gen-objects: pkg/apis/v1alpha1/zz_generated.deepcopy.go

.PHONY: gen-manifests
gen-manifests: config/crd/bases/*.yaml config/webhook/manifests.yaml

test_crd_sources := $(filter-out tests/resources/zz_generated.deepcopy.go,$(wildcard tests/resources/*.go))

tests/resources/zz_generated.deepcopy.go: $(test_crd_sources)
	$(CONTROLLER_GEN) \
                object \
                paths=./tests/resources

.PHONY: test-gen-objects
test-gen-objects: tests/resources/zz_generated.deepcopy.go

tests/resources/crds/*.yaml: $(test_crd_sources)
	$(CONTROLLER_GEN) \
		crd \
		paths=./tests/resources \
		output:crd:artifacts:config=tests/resources/crds
	$(ADDLICENSE) \
		-f ./hack/boilerplate.go.txt \
		tests/resources/crds

.PHONY: test-gen-manifests
test-gen-manifests: tests/resources/crds/*.yaml

.PHONY: clean-fakes
clean-fakes:
	find . -type d -name  '*fakes' | xargs -n1 rm -r

.PHONY: generate
generate: clean-fakes
	go generate ./...

.PHONY: test-cartotest
test-cartotest: test-cartotest-go test-cartotest-cli

.PHONY: test-cartotest-cli
test-cartotest-cli:
	go run ./cmd/cartotest/main.go --directory ./tests/templates/

.PHONY: test-cartotest-go
test-cartotest-go:
	go test ./tests/templates

.PHONY: test-unit
test-unit: test-gen-objects
	$(GINKGO) -r pkg

.PHONY: test-integration
ifeq ($(CI), true)
test-integration: export GOMEGA_DEFAULT_EVENTUALLY_TIMEOUT = 10s
test-integration: export GOMEGA_DEFAULT_CONSISTENTLY_DURATION = 2s
endif
test-integration: test-gen-manifests test-gen-objects
	$(GINKGO) -r tests/integration

.PHONY: test-kuttl
test-kuttl: build test-gen-manifests
	if [ -n "$$focus" ]; then $(CONTROL_PLANE_KUTTL) --test $$(basename $(focus)); else $(CONTROL_PLANE_KUTTL); fi

.PHONY: test-kuttl-runnable
test-kuttl-runnable: build test-gen-manifests
	if [ -n "$$focus" ]; then $(CONTROL_PLANE_KUTTL) ./tests/kuttl/runnable --test $$(basename $(focus)); else $(CONTROL_PLANE_KUTTL) ./tests/kuttl/runnable; fi

.PHONY: test-kuttl-supplychain
test-kuttl-supplychain: build test-gen-manifests
	if [ -n "$$focus" ]; then $(CONTROL_PLANE_KUTTL) ./tests/kuttl/supplychain --test $$(basename $(focus)); else $(CONTROL_PLANE_KUTTL) ./tests/kuttl/supplychain; fi

.PHONY: test-kuttl-delivery
test-kuttl-delivery: build test-gen-manifests
	if [ -n "$$focus" ]; then $(CONTROL_PLANE_KUTTL) ./tests/kuttl/delivery --test $$(basename $(focus)); else $(CONTROL_PLANE_KUTTL) ./tests/kuttl/delivery; fi

.PHONY: list-kuttl
list-kuttl:
	(cd tests/kuttl && find . -maxdepth 2 -type d)

.PHONY: test-kuttl-kind
test-kuttl-kind: build
	kubectl kuttl test --start-kind=true --start-control-plane=false --artifacts-dir=/dev/null

.PHONY: test
test: test-unit test-kuttl test-integration test-cartotest

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

.PHONY: woke
woke:
	$(WOKE) -c https://via.vmw.com/its-woke-rules

.PHONY: lint
lint: copyright woke
	$(GCI_LINT) write -s standard -s default -s "prefix(github.com/vmware-tanzu/cartographer)" $$(find ./pkg ! -name "fake_*" -type f)
	$(GOLANGCI_LINT) --config lint-config.yaml run
	$(MAKE) -C hack lint

.PHONY: copyright
copyright:
	$(ADDLICENSE) \
		-f ./hack/boilerplate.go.txt .

# Creates a GHA alike container image so we Mac user's can test flakies from CI
ifeq ($(UNAME), Darwin)
.PHONY: mac-compat
mac-compat:
	docker build . -f ./Dockerfile.compat -t compat:latest
endif

# Shortcut to run GHA alike container image with the repo mounted in /app
# eg: first: `make run-in-compat` then `make test-unit`
ifeq ($(UNAME), Darwin)
.PHONY: run-in-compat
run-in-compat:
	docker run -it -v $$(pwd):/app -w /app -it compat:latest /bin/bash
endif

.PHONY: pre-push .pre-push-check
.pre-push-check: copyright lint gen-manifests gen-objects test-gen-manifests test-gen-objects generate

# pre-push ensures that all generated content, copywrites and lints are
# run and ends with an error if a mutation is caused.
#
# usage:
#  1. with all your work added and committed (or stashed)
#  2. run `make pre-push && git push`
#  3. if any mutations occur, you can amend/rewrite or otherwise adjust your commits to include the changes
pre-push:
	[ -z "$$(git status --porcelain)" ] || (echo "not everything is committed, failing" && exit 1)
	$(MAKE) .pre-push-check
	[ -z "$$(git status --porcelain)" ] || (echo "changes occurred during pre-push check" && git diff HEAD --exit-code)
