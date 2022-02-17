import {editor, IPosition, languages, Range} from "monaco-editor";
import {CompletionItemKind, integer} from "vscode-languageserver-types";
import ITextModel = editor.ITextModel;

import {parse, stringify} from 'yaml'
import {upperCaseFirst} from "upper-case-first";

const resourceGroupRE = /(config|image|source)s:/

const inResourceKind = (model: ITextModel, lineNumber: integer) => {
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


const getSuggestions = (model: ITextModel, type: string) => {
    let doc = model.getValue()

    try {
        let obj = parse(doc)

        let typedResources = obj.spec.resources.filter(resource => resource.templateRef.kind === type)
        let mappedResources = typedResources.map(resource => ({
            insertText: resource.name,
            kind: CompletionItemKind.Reference,
            range: undefined,
            label: resource.name
        }))
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
            // triggerCharacters: [' ', ':'],
            triggerCharacters: [' '],
            provideCompletionItems(model, position) {
                let resourceKind = inResourceKind(model, position.lineNumber)
                if (resourceKind) {
                    return {
                        incomplete: false,
                        suggestions: getSuggestions(model, resourceKind),
                    };
                } else {
                    return null
                }
            },
        }
    )
}

export default AddSupplyChainLang