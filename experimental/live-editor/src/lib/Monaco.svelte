<script>

    import {onMount} from "svelte";
    import {editor, Uri} from 'monaco-editor'

    let instance
    let editorContainer
    export let document

    const modelUri = Uri.parse('https://cartographer.sh/file.yaml');

    onMount(() => {
        instance = editor.create(editorContainer,
            {
                automaticLayout: true,
                model: editor.createModel(document, 'yaml', modelUri)
            },
        )
        instance.onDidChangeModelContent(e => {
            console.log("edit")
            document = instance.getValue()
        })
    })

</script>

<div class="monaco-editor-container" bind:this={editorContainer} style="height: 400px"></div>

<style>
    .monaco-editor-container {
        outline: lightgray solid 1px;
    }
</style>