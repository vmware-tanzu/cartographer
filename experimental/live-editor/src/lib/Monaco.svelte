<script>

    import {onMount} from "svelte";
    import {editor, Uri} from 'monaco-editor'
    import {document} from '../store.js'

    let instance
    let editorContainer

    const modelUri = Uri.parse('https://cartographer.sh/file.yaml');
    const model = editor.getModel(modelUri) || editor.createModel("", 'yaml', modelUri)

    onMount(() => {
        model.setValue($document)
        instance = editor.create(editorContainer,
            {
                automaticLayout: true,
                model: model
            },
        )

        instance.onDidChangeModelContent(e => {
            $document = instance.getValue()
        })
    })

</script>

<div class="monaco-editor-container {$$props.class}" bind:this={editorContainer}></div>

<style>
    .monaco-editor-container {
        outline: lightgray solid 1px;
    }
</style>