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

import {CancellationToken, editor, languages, Position} from "monaco-editor";

import {isPair, isScalar, isSeq, LineCounter, Pair, parseDocument, Scalar, visit, YAMLMap, YAMLSeq} from 'yaml'
import {upperCaseFirst} from "upper-case-first";
import {CompletionItemKind} from "vscode-languageserver-types";
import ITextModel = editor.ITextModel;
import ProviderResult = languages.ProviderResult;
import CompletionList = languages.CompletionList;
import CompletionItem = languages.CompletionItem;
import Definition = languages.Definition;
import LocationLink = languages.LocationLink;
import Location = languages.Location;

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

const getSuggestions = (model: editor.ITextModel, kind: string, position: Position): CompletionItem[] => {
    let doc = model.getValue()
    let lineCounter = new LineCounter()
    try {
        let docNode = parseDocument(doc, {keepSourceTokens: true, lineCounter: lineCounter})

        let resourcesByType = docNode.getIn(["spec", "resources"]).items
            .filter((item: YAMLMap) => {
                let endOfItem = lineCounter.linePos(item.range[2])
                // normally you would use position.lineNumber+1 to make it 1-based
                // however if you're autocompleting on the last line of a resource, then the current line
                // is one higher than the end of the item (we want to exclude self-refs), so we need to
                // subtract 1. so: (endOfItem.line < position.lineNumber + 1 - 1)
                // becomes: (endOfItem.line < position.lineNumber)
                return (endOfItem.line < position.lineNumber) &&
                    item.getIn(["templateRef", "kind"]) === kind
            })

        let mappedResources = resourcesByType.map((resource: YAMLMap): CompletionItem => {
            let name: string = <string>resource.get("name")
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

function getReference(model: editor.ITextModel, position: Position) {
    let doc = model.getValue()
    let lineCounter = new LineCounter()

    // Not currently testing the path, we should make sure it's a `spec.resources.[configs|images|sources].resource`
    const isResourcePair = (pair: Pair) => isScalar(pair.key) &&
        isScalar(pair.value) &&
        (<Scalar>pair.key).value === "resource";

    const isUnderCaret = (range) => {
        let startPos = lineCounter.linePos(range[0])
        let endPos = lineCounter.linePos(range[1])
        return (endPos.line >= position.lineNumber && position.lineNumber >= startPos.line) &&
            (endPos.col >= position.column && position.column >= startPos.col)
    }

    let range

    try {
        let objNode = parseDocument(doc, {keepSourceTokens: true, lineCounter: lineCounter})
        visit(objNode, {
            Pair(id, pair, path) {
                if (isResourcePair(pair) && isUnderCaret((<Scalar>pair.value).range)) {
                    let collectionNode = path[path.length - 3]
                    if (isPair(collectionNode) && isSeq(collectionNode.value) && isScalar(collectionNode.key)) {
                        let collectionKind = {
                            "configs": "ClusterConfigTemplate",
                            "sources": "ClusterSourceTemplate",
                            "images": "ClusterImageTemplate",
                        }[<string>collectionNode.key.value];

                        let definitionNode:YAMLMap = (<YAMLMap[]>(<YAMLSeq>objNode.getIn(["spec", "resources"])).items).find(
                            (resourceNode: YAMLMap) => resourceNode.getIn(["templateRef", "kind"]) === collectionKind
                        );

                        if (definitionNode) {
                            let nameNode = definitionNode.items[0].value // this is wrong
                            let start = lineCounter.linePos(nameNode.range[0])
                            let end = lineCounter.linePos(nameNode.range[1])
                            range = {
                                startLineNumber: start.line,
                                endLineNumber: end.line,
                                startColumn: start.col,
                                endColumn: end.col
                            }
                            return visit.BREAK
                        }
                    }
                }
            },
        })
    } catch (e) {
        // no-op, don't care
    }
    return range
}

export const AddSupplyChainLang = () => {
    languages.registerDefinitionProvider(
        'yaml',
        {
            provideDefinition(model: ITextModel, position: Position, token: CancellationToken): ProviderResult<Definition | LocationLink[]> {
                let usage = getReference(model, position)
                if (usage) {
                    console.log(usage)
                    return <Location>{
                        uri: model.uri,
                        range: usage
                    }
                }
            }
        }
    )

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