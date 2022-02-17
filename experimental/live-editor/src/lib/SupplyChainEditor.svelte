<script>
    import Monaco from "./Monaco.svelte";
    import {document} from "../store";
</script>

<Monaco bind:document={$document}/>