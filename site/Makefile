.PHONY: serve
serve:
	echo "Open docs at: http://localhost:1313"
	hugo server --disableFastRender

.PHONY: release
release:
	if [ -z "$$version" ]; then echo "\nERROR: must provide version=v#.#.#\n" && exit 1; fi
	./hack/new-release.sh "$$version"

.PHONY: gen-crd-reference
gen-crd-reference:
	./hack/crds.rb

.PHONY: dev-dependencies
dev-dependencies:
	yarn install

.PHONY: lint
lint: dev-dependencies
	yarn lint
