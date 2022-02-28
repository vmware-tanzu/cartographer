/**
 * Copyright 2021 VMware
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import {derived, writable} from 'svelte/store';
import {parse, parseDocument} from "yaml";
import {toMermaid} from "./lib/viz-processor.js";
import {deflate, deflateRaw, inflate, inflateRaw} from "pako";
import {fromUint8Array} from "js-base64";

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
    "        options:\n" +
    "          - name: from-git\n" +
    "            selector:\n" +
    "              matchFields:\n" +
    "                - key: spec.source.git\n" +
    "                  operator: Exists\n" +
    "          - name: from-repo\n" +
    "            selector:\n" +
    "              matchFields:\n" +
    "                - key: spec.source.image\n" +
    "                  operator: Exists\n" +
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
    "          name: data\n" +
    "\n"
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

export const compressedState = derived(
    document,
    $document => {
        try {
            let data = new TextEncoder().encode($document)
            let compressed = deflate(data, {options: 9})
            return fromUint8Array(compressed, true)
        } catch (e) {
            console.log(`could not compress document: ${e}`)
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


