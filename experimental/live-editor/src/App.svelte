<script>
    import Viz from "./lib/Viz.svelte";
    import SupplyChainEditor from "./lib/SupplyChainEditor.svelte";
    import {onMount} from "svelte";

    let resizer
    let leftPane

    const resize = (e) => {
        const newWidth = e.pageX - leftPane.getBoundingClientRect().left;
        if (newWidth > 50) {
            leftPane.style.width = `${newWidth}px`;
        }
    };
    const stopResize = () => {
        window.removeEventListener('mousemove', resize);
    };

    onMount(() => {
        resizer.addEventListener('mousedown', (e) => {
            e.preventDefault();
            window.addEventListener('mousemove', resize);
            window.addEventListener('mouseup', stopResize);
        });
    })
</script>

<main class="h-screen text-primary-content">
    <div class="h-full flex flex-col overflow-hidden">
        <div class="navbar mb-2 shadow-lg bg-sky-700 text-sky-50 pl-2 pr-2">
            <h1 class="text-lg uppercase">Cartographer Live Editor</h1>
        </div>

        <div class="flex-1 flex overflow-hidden">
            <div class="hidden md:flex flex-col" bind:this={leftPane} style="width: 40%;">
                <SupplyChainEditor></SupplyChainEditor>
            </div>
            <div id="resizeHandler" class="md:block " bind:this={resizer}></div>
            <div class="flex-1 flex flex-col overflow-hidden">
                <Viz/>
            </div>
        </div>
    </div>

</main>

<style gobal>
    #resizeHandler {
        cursor: col-resize;
        padding: 0 2px;
    }
</style>
