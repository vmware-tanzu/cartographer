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

// src/index.ts
import { Emitter, languages as languages4 } from "monaco-editor/esm/vs/editor/editor.api.js";

// src/constants.ts
var languageId = "yaml";

// src/yamlMode.ts
import { languages as languages3 } from "monaco-editor/esm/vs/editor/editor.api.js";

// src/languageFeatures.ts
import {
  editor,
  languages,
  MarkerSeverity,
  Range,
  Uri
} from "monaco-editor/esm/vs/editor/editor.api.js";
import {
  CompletionItemKind,
  DiagnosticSeverity,
  InsertTextFormat,
  SymbolKind
} from "vscode-languageserver-types";
function toSeverity(lsSeverity) {
  switch (lsSeverity) {
    case DiagnosticSeverity.Error:
      return MarkerSeverity.Error;
    case DiagnosticSeverity.Warning:
      return MarkerSeverity.Warning;
    case DiagnosticSeverity.Information:
      return MarkerSeverity.Info;
    case DiagnosticSeverity.Hint:
      return MarkerSeverity.Hint;
    default:
      return MarkerSeverity.Info;
  }
}
function toDiagnostics(diag) {
  return {
    severity: toSeverity(diag.severity),
    startLineNumber: diag.range.start.line + 1,
    startColumn: diag.range.start.character + 1,
    endLineNumber: diag.range.end.line + 1,
    endColumn: diag.range.end.character + 1,
    message: diag.message,
    code: String(diag.code),
    source: diag.source
  };
}
function createDiagnosticsAdapter(getWorker, defaults) {
  const listeners = new Map();
  const resetSchema = async (resource) => {
    const worker = await getWorker();
    worker.resetSchema(String(resource));
  };
  const doValidate = async (resource) => {
    const worker = await getWorker(resource);
    const diagnostics = await worker.doValidation(String(resource));
    const markers = diagnostics.map(toDiagnostics);
    const model = editor.getModel(resource);
    if (model && model.getLanguageId() === languageId) {
      editor.setModelMarkers(model, languageId, markers);
    }
  };
  const onModelAdd = (model) => {
    if (model.getLanguageId() !== languageId) {
      return;
    }
    let handle;
    listeners.set(String(model.uri), model.onDidChangeContent(() => {
      clearTimeout(handle);
      handle = setTimeout(() => doValidate(model.uri), 500);
    }));
    doValidate(model.uri);
  };
  const onModelRemoved = (model) => {
    editor.setModelMarkers(model, languageId, []);
    const uriStr = String(model.uri);
    const listener = listeners.get(uriStr);
    if (listener) {
      listener.dispose();
      listeners.delete(uriStr);
    }
  };
  editor.onDidCreateModel(onModelAdd);
  editor.onWillDisposeModel((model) => {
    onModelRemoved(model);
    resetSchema(model.uri);
  });
  editor.onDidChangeModelLanguage((event) => {
    onModelRemoved(event.model);
    onModelAdd(event.model);
    resetSchema(event.model.uri);
  });
  defaults.onDidChange(() => {
    for (const model of editor.getModels()) {
      if (model.getLanguageId() === languageId) {
        onModelRemoved(model);
        onModelAdd(model);
      }
    }
  });
  for (const model of editor.getModels()) {
    onModelAdd(model);
  }
}
function fromPosition(position) {
  if (!position) {
    return;
  }
  return { character: position.column - 1, line: position.lineNumber - 1 };
}
function toRange(range) {
  if (!range) {
    return;
  }
  return new Range(range.start.line + 1, range.start.character + 1, range.end.line + 1, range.end.character + 1);
}
function toCompletionItemKind(kind) {
  const mItemKind = languages.CompletionItemKind;
  switch (kind) {
    case CompletionItemKind.Text:
      return mItemKind.Text;
    case CompletionItemKind.Method:
      return mItemKind.Method;
    case CompletionItemKind.Function:
      return mItemKind.Function;
    case CompletionItemKind.Constructor:
      return mItemKind.Constructor;
    case CompletionItemKind.Field:
      return mItemKind.Field;
    case CompletionItemKind.Variable:
      return mItemKind.Variable;
    case CompletionItemKind.Class:
      return mItemKind.Class;
    case CompletionItemKind.Interface:
      return mItemKind.Interface;
    case CompletionItemKind.Module:
      return mItemKind.Module;
    case CompletionItemKind.Property:
      return mItemKind.Property;
    case CompletionItemKind.Unit:
      return mItemKind.Unit;
    case CompletionItemKind.Value:
      return mItemKind.Value;
    case CompletionItemKind.Enum:
      return mItemKind.Enum;
    case CompletionItemKind.Keyword:
      return mItemKind.Keyword;
    case CompletionItemKind.Snippet:
      return mItemKind.Snippet;
    case CompletionItemKind.Color:
      return mItemKind.Color;
    case CompletionItemKind.File:
      return mItemKind.File;
    case CompletionItemKind.Reference:
      return mItemKind.Reference;
    default:
      return mItemKind.Property;
  }
}
function toTextEdit(textEdit) {
  if (!textEdit) {
    return;
  }
  return {
    range: toRange(textEdit.range),
    text: textEdit.newText
  };
}
function createCompletionItemProvider(getWorker) {
  return {
    triggerCharacters: [" ", ":"],
    async provideCompletionItems(model, position) {
      const resource = model.uri;
      const worker = await getWorker(resource);
      const info = await worker.doComplete(String(resource), fromPosition(position));
      if (!info) {
        return;
      }
      const wordInfo = model.getWordUntilPosition(position);
      const wordRange = new Range(position.lineNumber, wordInfo.startColumn, position.lineNumber, wordInfo.endColumn);
      const items = info.items.map((entry) => {
        const item = {
          label: entry.label,
          insertText: entry.insertText || entry.label,
          sortText: entry.sortText,
          filterText: entry.filterText,
          documentation: entry.documentation,
          detail: entry.detail,
          kind: toCompletionItemKind(entry.kind),
          range: wordRange
        };
        if (entry.textEdit) {
          item.range = toRange("range" in entry.textEdit ? entry.textEdit.range : entry.textEdit.replace);
          item.insertText = entry.textEdit.newText;
        }
        if (entry.additionalTextEdits) {
          item.additionalTextEdits = entry.additionalTextEdits.map(toTextEdit);
        }
        if (entry.insertTextFormat === InsertTextFormat.Snippet) {
          item.insertTextRules = languages.CompletionItemInsertTextRule.InsertAsSnippet;
        }
        return item;
      });
      return {
        incomplete: info.isIncomplete,
        suggestions: items
      };
    }
  };
}
function createDefinitionProvider(getWorker) {
  return {
    async provideDefinition(model, position) {
      const resource = model.uri;
      const worker = await getWorker(resource);
      const definitions = await worker.doDefinition(String(resource), fromPosition(position));
      return definitions == null ? void 0 : definitions.map((definition) => ({
        originSelectionRange: definition.originSelectionRange,
        range: toRange(definition.targetRange),
        targetSelectionRange: definition.targetSelectionRange,
        uri: Uri.parse(definition.targetUri)
      }));
    }
  };
}
function createHoverProvider(getWorker) {
  return {
    async provideHover(model, position) {
      const resource = model.uri;
      const worker = await getWorker(resource);
      const info = await worker.doHover(String(resource), fromPosition(position));
      if (!info) {
        return;
      }
      return {
        range: toRange(info.range),
        contents: [{ value: info.contents.value }]
      };
    }
  };
}
function toSymbolKind(kind) {
  const mKind = languages.SymbolKind;
  switch (kind) {
    case SymbolKind.File:
      return mKind.Array;
    case SymbolKind.Module:
      return mKind.Module;
    case SymbolKind.Namespace:
      return mKind.Namespace;
    case SymbolKind.Package:
      return mKind.Package;
    case SymbolKind.Class:
      return mKind.Class;
    case SymbolKind.Method:
      return mKind.Method;
    case SymbolKind.Property:
      return mKind.Property;
    case SymbolKind.Field:
      return mKind.Field;
    case SymbolKind.Constructor:
      return mKind.Constructor;
    case SymbolKind.Enum:
      return mKind.Enum;
    case SymbolKind.Interface:
      return mKind.Interface;
    case SymbolKind.Function:
      return mKind.Function;
    case SymbolKind.Variable:
      return mKind.Variable;
    case SymbolKind.Constant:
      return mKind.Constant;
    case SymbolKind.String:
      return mKind.String;
    case SymbolKind.Number:
      return mKind.Number;
    case SymbolKind.Boolean:
      return mKind.Boolean;
    case SymbolKind.Array:
      return mKind.Array;
    default:
      return mKind.Function;
  }
}
function toDocumentSymbol(item) {
  return {
    detail: item.detail || "",
    range: toRange(item.range),
    name: item.name,
    kind: toSymbolKind(item.kind),
    selectionRange: toRange(item.selectionRange),
    children: item.children.map(toDocumentSymbol),
    tags: []
  };
}
function createDocumentSymbolProvider(getWorker) {
  return {
    async provideDocumentSymbols(model) {
      const resource = model.uri;
      const worker = await getWorker(resource);
      const items = await worker.findDocumentSymbols(String(resource));
      if (!items) {
        return;
      }
      return items.map(toDocumentSymbol);
    }
  };
}
function fromFormattingOptions(options) {
  return {
    tabSize: options.tabSize,
    insertSpaces: options.insertSpaces,
    ...options
  };
}
function createDocumentFormattingEditProvider(getWorker) {
  return {
    async provideDocumentFormattingEdits(model, options) {
      const resource = model.uri;
      const worker = await getWorker(resource);
      const edits = await worker.format(String(resource), fromFormattingOptions(options));
      if (!edits || edits.length === 0) {
        return;
      }
      return edits.map(toTextEdit);
    }
  };
}
function toLink(link) {
  return {
    range: toRange(link.range),
    tooltip: link.tooltip,
    url: link.target
  };
}
function createLinkProvider(getWorker) {
  return {
    async provideLinks(model) {
      const resource = model.uri;
      const worker = await getWorker(resource);
      const links = await worker.findLinks(String(resource));
      return {
        links: links.map(toLink)
      };
    }
  };
}

// src/workerManager.ts
import { editor as editor2 } from "monaco-editor/esm/vs/editor/editor.api.js";
var STOP_WHEN_IDLE_FOR = 2 * 60 * 1e3;
function createWorkerManager(defaults) {
  let worker;
  let client;
  let lastUsedTime = 0;
  const stopWorker = () => {
    if (worker) {
      worker.dispose();
      worker = void 0;
    }
    client = void 0;
  };
  setInterval(() => {
    if (!worker) {
      return;
    }
    const timePassedSinceLastUsed = Date.now() - lastUsedTime;
    if (timePassedSinceLastUsed > STOP_WHEN_IDLE_FOR) {
      stopWorker();
    }
  }, 30 * 1e3);
  defaults.onDidChange(() => stopWorker());
  const getClient = () => {
    lastUsedTime = Date.now();
    if (!client) {
      worker = editor2.createWebWorker({
        moduleId: "vs/language/yaml/yamlWorker",
        label: defaults.languageId,
        createData: {
          languageSettings: defaults.diagnosticsOptions,
          enableSchemaRequest: defaults.diagnosticsOptions.enableSchemaRequest,
          isKubernetes: defaults.diagnosticsOptions.isKubernetes,
          customTags: defaults.diagnosticsOptions.customTags
        }
      });
      client = worker.getProxy();
    }
    return client;
  };
  return async (...resources) => {
    const client2 = await getClient();
    await worker.withSyncedResources(resources);
    return client2;
  };
}

// src/yamlMode.ts
var richEditConfiguration = {
  comments: {
    lineComment: "#"
  },
  brackets: [
    ["{", "}"],
    ["[", "]"],
    ["(", ")"]
  ],
  autoClosingPairs: [
    { open: "{", close: "}" },
    { open: "[", close: "]" },
    { open: "(", close: ")" },
    { open: '"', close: '"' },
    { open: "'", close: "'" }
  ],
  surroundingPairs: [
    { open: "{", close: "}" },
    { open: "[", close: "]" },
    { open: "(", close: ")" },
    { open: '"', close: '"' },
    { open: "'", close: "'" }
  ],
  onEnterRules: [
    {
      beforeText: /:\s*$/,
      action: { indentAction: languages3.IndentAction.Indent }
    }
  ]
};
function setupMode(defaults) {
  const worker = createWorkerManager(defaults);
  languages3.registerCompletionItemProvider(languageId, createCompletionItemProvider(worker));
  languages3.registerHoverProvider(languageId, createHoverProvider(worker));
  languages3.registerDefinitionProvider(languageId, createDefinitionProvider(worker));
  languages3.registerDocumentSymbolProvider(languageId, createDocumentSymbolProvider(worker));
  languages3.registerDocumentFormattingEditProvider(languageId, createDocumentFormattingEditProvider(worker));
  languages3.registerLinkProvider(languageId, createLinkProvider(worker));
  createDiagnosticsAdapter(worker, defaults);
  languages3.setLanguageConfiguration(languageId, richEditConfiguration);
}

// src/index.ts
var diagnosticDefault = {
  completion: true,
  customTags: [],
  enableSchemaRequest: false,
  format: true,
  isKubernetes: false,
  hover: true,
  schemas: [],
  validate: true,
  yamlVersion: "1.2"
};
function createLanguageServiceDefaults(initialDiagnosticsOptions) {
  const onDidChange = new Emitter();
  let diagnosticsOptions = initialDiagnosticsOptions;
  const languageServiceDefaults = {
    get onDidChange() {
      return onDidChange.event;
    },
    get languageId() {
      return languageId;
    },
    get diagnosticsOptions() {
      return diagnosticsOptions;
    },
    setDiagnosticsOptions(options) {
      diagnosticsOptions = { ...diagnosticDefault, ...options };
      onDidChange.fire(languageServiceDefaults);
    }
  };
  return languageServiceDefaults;
}
var yamlDefaults = createLanguageServiceDefaults(diagnosticDefault);
function createAPI() {
  return {
    yamlDefaults
  };
}
languages4.yaml = createAPI();
languages4.register({
  id: languageId,
  extensions: [".yaml", ".yml"],
  aliases: ["YAML", "yaml", "YML", "yml"],
  mimetypes: ["application/x-yaml"]
});
languages4.onLanguage("yaml", () => {
  setupMode(yamlDefaults);
});
function setDiagnosticsOptions(options = {}) {
  languages4.yaml.yamlDefaults.setDiagnosticsOptions(options);
}
export {
  createLanguageServiceDefaults,
  setDiagnosticsOptions
};
//# sourceMappingURL=index.js.map
