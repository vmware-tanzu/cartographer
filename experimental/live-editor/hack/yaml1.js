import {parseDocument} from "yaml"

let doc = [
    "---",
    "foo: bar",
    "boom: bang"
].join("\n")


let out = parseDocument(doc, )
