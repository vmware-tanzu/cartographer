.PHONY: serve
serve:
	open http://localhost:1313
	hugo server --disableFastRender

.PHONY: release
release:
	if [ -z "$$version" ]; then echo "\nERROR: must provide version=v#.#.#\n" && exit 1; fi
	./hack/new-release.sh "$$version"
