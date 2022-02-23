<script>
    import mermaid from 'mermaid';
    import {afterUpdate, onMount} from 'svelte';

    export let doc = ""
    let graph = null;

    mermaid.initialize({
        startOnLoad: false,
        theme: 'forest',
    });

    const rerender = () => {
        try {
            mermaid.mermaidAPI.render('graph-div', doc, (svgCode) => {
                graph.innerHTML = svgCode;
            });
        } catch (e) {
            // FIXME we should ensure a valid mermaid doc before we land here.
        }
    }

    onMount(() => {
        rerender()
    });

    afterUpdate(() => {
        rerender()
    })
</script>

<pre bind:this={graph} class="{$$props.class}">
</pre>
