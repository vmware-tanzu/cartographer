import {derived, writable} from 'svelte/store';
import {parse} from "yaml";
import {toMermaid} from "./lib/viz-processor.js";

export const document = writable("---\n" +
    "apiVersion: carto.run/v1alpha1\n" +
    "kind: ClusterSupplyChain\n" +
    "metadata:\n" +
    "  name: supply-chain\n" +
    "spec:\n" +
    "  selector:\n" +
    "    app.tanzu.vmware.com/workload-type: web\n" +
    "\n" +
    "  resources:\n" +
    "    - name: source-provider\n" +
    "      templateRef:\n" +
    "        kind: ClusterSourceTemplate\n" +
    "        name: source\n" +
    "\n" +
    "    - name: image-builder\n" +
    "      templateRef:\n" +
    "        kind: ClusterImageTemplate\n" +
    "        name: image\n" +
    "      params:\n" +
    "        - name: image_prefix\n" +
    "          value: \"pref-\"\n" +
    "      sources:\n" +
    "        - resource: source-provider\n" +
    "          name: source\n" +
    "\n" +
    "    - name: config-provider\n" +
    "      templateRef:\n" +
    "        kind: ClusterConfigTemplate\n" +
    "        name: app-config\n" +
    "      images:\n" +
    "        - resource: image-builder\n" +
    "          name: image\n" +
    "\n" +
    "    - name: git-writer\n" +
    "      templateRef:\n" +
    "        kind: ClusterTemplate\n" +
    "        name: git-writer\n" +
    "      configs:\n" +
    "        - resource: config-provider\n" +
    "          name: data\n"
);

export const documentObject = derived(
    document,
    $document => {
        try {
            return parse($document)
        } catch (e) {
            console.log(`could not parse to yaml object: ${e}`)
        }
    }
)


export const mermaidDoc = derived(
    documentObject,
    ($docObj, set) => {
        try {
            set(toMermaid($docObj))
        } catch (e) {
            console.log(`could not parse to mermaid: ${e}`)
        }

    }
)


