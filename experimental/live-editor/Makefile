.PHONY: deps
deps:
	yarn

.PHONY: build
build: deps
	yarn build

.PHONY: install
install:
	mkdir -p ../../site/themes/template/static/live-editor
	cp -r dist/* ../../site/themes/template/static/live-editor/

