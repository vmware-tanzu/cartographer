
# Blueprint and Owner Live editor.

## Roadmap

### Proof Of Concept

* Supply chain editor (no delivery, workload or deliverable)
* Pako style compact URL with saving
* Visualiser for Supply Chain
* Autocomplete for resources
  * Don't show forward references (or self references)
  * Type validation for references
* Host it somewhere. Hugo seems inordinately painful for this, but I could be co-erced
* document build and delivery process to update the editor wherever it lands


### Near future

* Finish the `scheming` tool that takes any CRD and produces a json schema
  * This will make it easier to keep the editor up to date

### Further Future
 
* Workload with supply chain, including viz and param matching
* Param autocomplete across files
* Use a worker for the language extension
* Do something about transpiling/packing, Vite hates Monaco-Yaml afaict
* Automate build/deploy (perhaps at this point we get a dedicated repository?)

# Howto

## Updating the Schema today

1. Grab the schema with:
   ```
   cat config/crd/bases/carto.run_clustersupplychains.yaml | yq '.spec.versions[] | select(.name="v1alpha1") | .schema.openAPIV3Schema'
   ```
2. paste into [`./hack/schema.js`](./hack/schema.js)
3. then run 
   ```
   ./hack/schema.js  | pbcopy
   ```

4. and paste final schema into [`./src/main.js`](./src/main.js)
