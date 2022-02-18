<script>
    import Mermaid from "./Mermaid.svelte";
    import {mermaidDoc} from "../store.js";

</script>
<!--<pre>-->
<!--    {$document}-->
<!--</pre>-->
<Mermaid doc={$mermaidDoc} class="{$$props.class}"/>
