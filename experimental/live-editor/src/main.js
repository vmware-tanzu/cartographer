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

import App from './App.svelte'
import {setDiagnosticsOptions} from "./monaco-yaml";
import YamlWorker from './monaco-yaml/yaml.worker?worker';
import TsWorker from 'monaco-editor/esm/vs/language/typescript/ts.worker?worker';
import EditorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker';
import {schema} from "./lib/monaco/schema";
import AddSupplyChainLang from "./lib/monaco/supply-chain-lang";
import "./app.css"

window.MonacoEnvironment = {
    getWorker(moduleId, label) {
        switch (label) {
            case 'yaml':
                return new YamlWorker();
            case 'javascript':
                return new TsWorker();
            default:
                return new EditorWorker();
        }
    },
};

setDiagnosticsOptions({
    enableSchemaRequest: true,
    hover: true,
    completion: true,
    validate: true,
    format: true,
    schemas: [
        {
            fileMatch: ["*.yaml"],
            schema: schema,
        },
    ],
});

AddSupplyChainLang()

const app = new App({
    target: document.body
})

export default app
