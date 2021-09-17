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

lint: copyright
	go run github.com/golangci/golangci-lint/cmd/golangci-lint --config lint-config.yaml run

copyright:
	go run github.com/google/addlicense \
		-f ./hack/boilerplate.go.txt \
		-ignore site/static/\*\* \
		-ignore site/themes/\*\* \
		.
