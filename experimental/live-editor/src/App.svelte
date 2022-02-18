<script>
    import FlexPane from "./lib/FlexPane.svelte";
    import SupplyChainEditor from "./lib/SupplyChainEditor.svelte";
    import Viz from "./lib/Viz.svelte";


    const showHelp = () => {
        
    }
</script>
<main class="h-screen text-primary-content">
    <nav class="font-sans flex flex-col text-center sm:flex-row sm:text-left sm:justify-between py-2 px-6 bg-sky-700 text-sky-50 shadow sm:items-baseline w-full">
        <div class="mb-2 sm:mb-0">
            <h1 class="text-lg uppercase">Cartographer Live Editor</h1>
        </div>
        <div>
            <button class="text-lg no-underline text-grey-darkest hover:text-orange-300 ml-2" on:click={showHelp}>Help</button>
        </div>
    </nav>
    <div class="h-full flex flex-col overflow-hidden">
        <FlexPane>
            <SupplyChainEditor slot="left" class="h-full m-2"/>
            <Viz slot="right" class="content-center"/>
        </FlexPane>
    </div>

</main>

