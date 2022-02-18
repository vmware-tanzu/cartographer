<script>
    import FlexPane from "./lib/FlexPane.svelte";
    import SupplyChainEditor from "./lib/SupplyChainEditor.svelte";
    import Viz from "./lib/Viz.svelte";
</script>
<main class="h-screen text-primary-content">
    <div class="h-full flex flex-col overflow-hidden">
        <div class="navbar bg-sky-700 text-sky-50 pl-2 pr-2 shadow-lg">
            <h1 class="text-lg uppercase">Cartographer Live Editor</h1>
        </div>

        <FlexPane>
            <SupplyChainEditor slot="left" class="h-full m-2"/>
            <Viz slot="right" class="content-center"/>
        </FlexPane>
    </div>

</main>

