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

import {editor, languages, Position} from "monaco-editor";
import ITextModel = editor.ITextModel;

import {parseDocument, YAMLMap, Scalar, YAMLSeq, Document, LineCounter} from 'yaml'
import {upperCaseFirst} from "upper-case-first";
import {CompletionItemKind} from "vscode-languageserver-types";
import ProviderResult = languages.ProviderResult;
import CompletionList = languages.CompletionList;
import CompletionItem = languages.CompletionItem;

const resourceGroupRE = /(config|image|source)s:/

const inResourceKind = (model: ITextModel, lineNumber: number) => {
    let line = model.getLineContent(lineNumber)
    let resourcePos = line.search(/\s+resource:/)
    if (resourcePos < 0) {
        return null
    }

    let searchLineNumber = lineNumber

    while (--searchLineNumber > 0) {
        let searchLine = model.getLineContent(searchLineNumber)
        let matches = searchLine.match(resourceGroupRE)
        if (matches) {
            return `Cluster${upperCaseFirst(matches[1])}Template`
        }
    }

    return null
}

const itemByKey = (map: YAMLMap, key: string) => map.items.find(item => (<Scalar>item.key).value === key)
const specNodeFromDocument = (docNode: Document) => itemByKey(<YAMLMap>docNode.contents, "spec")
const resourcesNodeFromDocument = (docNode: Document) => itemByKey(<YAMLMap>specNodeFromDocument(docNode).value, "resources")

const templateKindFromResourceNode = (resourceNode: YAMLMap) => (<Scalar>itemByKey(<YAMLMap>(itemByKey(resourceNode, "templateRef").value), "kind").value).value
// const filterResourcesByKind = (resourcesNode: YAMLMap, kind: string) => resourcesNode.items.filter((item: YAMLMap) => templateKindFromResourceNode(item) === kind))

const getSuggestions = (model: editor.ITextModel, kind: string, position: Position): CompletionItem[] => {
    let doc = model.getValue()
    let lineCounter = new LineCounter()
    try {
        let objNode = parseDocument(doc, {keepSourceTokens: true, lineCounter: lineCounter})

        let resourcesByType = (<YAMLSeq>resourcesNodeFromDocument(objNode).value).items
            .filter((item: YAMLMap) => {
                let endOfItem = lineCounter.linePos(item.range[2])
                // normally you would use position.lineNumber+1 to make it 1-based
                // however if you're autocompleting on the last line of a resource, then the current line
                // is one higher than the end of the item (we want to exclude self-refs), so we need to
                // subtract 1. so: (endOfItem.line < position.lineNumber + 1 - 1)
                // becomes: (endOfItem.line < position.lineNumber)
                return (endOfItem.line < position.lineNumber) &&
                    (templateKindFromResourceNode(item) === kind)
            })

        let mappedResources = resourcesByType.map((resource: YAMLMap): CompletionItem => {
            let name: string = <string>(<Scalar>itemByKey(resource, "name").value).value
            return {
                insertText: name,
                kind: CompletionItemKind.Reference,
                range: null,
                label: name
            }
        })
        console.log(mappedResources)
        return mappedResources
    } catch (e) {
        // no-op, don't care
    }
    return []
};

export const AddSupplyChainLang = () => {

    languages.registerCompletionItemProvider(
        'yaml',
        {
            triggerCharacters: [' '],
            provideCompletionItems(model, position): ProviderResult<CompletionList> {
                let resourceKind = inResourceKind(model, position.lineNumber)
                if (resourceKind) {
                    return {
                        incomplete: true,
                        suggestions: getSuggestions(model, resourceKind, position),
                    };
                } else {
                    return null
                }
            },
        }
    )
}

export default AddSupplyChainLang