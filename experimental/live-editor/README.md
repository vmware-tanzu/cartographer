
# Blueprint and Owner Live editor.

## Roadmap
[Original Spike Issue](https://github.com/vmware-tanzu/cartographer/issues/566)

### Proof Of Concept

* Supply chain editor (no delivery, workload or deliverable)
* Pako style compact URL with saving
* Visualiser for Supply Chain
* Autocomplete for resources
  * Don't show forward references (or self references)
* Host it somewhere. Hugo seems inordinately painful for this, but I could be co-erced
* document build and delivery process to update the editor wherever it lands


### Near future

* Finish the `scheming` tool that takes any CRD and produces a json schema
  * This will make it easier to keep the editor up to date
* Host the schema and point to it. Then other's can use it in their editors yaml/json schema validation
* 
* Type validation for references
* Add [Options](../../tests/kuttl/supplychain/options-with-values/01-supply-chain.yaml) 

### Further Future

* Workload with supply chain, including viz 
  * Missing params and resulting-set params
* Params
  * Param autocomplete across files 
  * Blueprints show default and required params
* Use a worker for the language extension
* Do something about transpiling/packing, Vite hates Monaco-Yaml afaict (monaco-yaml working on a fix)
* Automate build/deploy (perhaps at this point we get a dedicated repository?)
* Template tabs (showing params?) and autocompleting template names in the supply chain resources
* Gist reference support (and maybe create-new-gist)
* Show that a workload will select the given supply chain AND/OR supply-chains
* Runnable? Delivery? Deliverable?

# Howto

## Updating the Schema today

1. Grab the schema with:
   ```
   cat ../../config/crd/bases/carto.run_clustersupplychains.yaml | yq '.spec.versions[] | select(.name="v1alpha1") | .schema.openAPIV3Schema'
   ```
2. paste into [`./hack/schema.js`](./hack/schema.js)
3. then run 
   ```
   ./hack/schema.js  | pbcopy
   ```

4. and paste final schema into [`./src/lib/monaco/schema.js`](./src/lib/monaco/schema.js)


## Install/Update blog post

```
make build install
```