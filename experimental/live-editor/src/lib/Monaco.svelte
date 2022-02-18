<script>

    import {onMount} from "svelte";
    import {editor, Uri} from 'monaco-editor'

    let instance
    let editorContainer
    export let document

    const modelUri = Uri.parse('https://cartographer.sh/file.yaml');

    onMount(() => {
        let model = editor.getModel(modelUri) || editor.createModel(document, 'yaml', modelUri)

        instance = editor.create(editorContainer,
            {
                automaticLayout: true,
                model: model
            },
        )
        instance.onDidChangeModelContent(e => {
            console.log("edit")
            document = instance.getValue()
        })
    })

</script>

<div class="monaco-editor-container {$$props.class}" bind:this={editorContainer} ></div>

<style>
    .monaco-editor-container {
        outline: lightgray solid 1px;
    }
</style>