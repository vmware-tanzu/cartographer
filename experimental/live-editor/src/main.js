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
    target: document.getElementById('app')
})

export default app
