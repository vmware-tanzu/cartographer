# Website for [Template]

## Prerequisites

- [Hugo](https://github.com/gohugoio/hugo)
  - macOS: `brew install hugo`
  - Windows: `choco install hugo-extended -confirm`

## Serve

```bash
make serve
```

Visit (http://localhost:1313)[http://localhost:1313]

## Generate a Release

to create a release copy of `development` use

```bash
make release version=v1.2.3
```

The new version should appear in the site and be the default.

## Generating CRD Documentation

There is a tool, `hack/crd.rb` designed to autogenerate CRD documentation based off the content of our Go doc-comments
in `/cartographer/pkg/apis`.

To update CRD documentation:

1. Edit the doc-comments in the api, eg: "/cartographer/pkg/apis/v1alpha1/workload.go"
2. Generate the new CRDs `make gen-manifests`
3. generate the CRD documentation `cd site && make gen-crd-reference`
4. review the changes to files in `/cartographer/site/content/docs/development/crds/*.yaml`
   1. Custom edits will be removed, so look for delta's that represent developer edits and roll those line's back
