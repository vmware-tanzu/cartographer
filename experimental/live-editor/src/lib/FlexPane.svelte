<script>
    import {onMount} from "svelte";

    let resizer
    let leftPane
    let width = "50%"

    const resize = (e) => {
        const newWidth = e.pageX - leftPane.getBoundingClientRect().left;
        if (newWidth > 50) {
            width = `${newWidth}px`;
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

<div class="flex-1 flex overflow-hidden">
    <div class="hidden md:flex flex-col" bind:this={leftPane} style="width: {width};">
        <slot name="left">
            <div style="text-align: center;">
                <em>missing &lt;slot=&quot;left&quot;&gt;</em>
            </div>
        </slot>
    </div>
    <div id="resizeHandler" class="hover:bg-sky-400 bg-sky-700 cursor-col-resize pr-1" bind:this={resizer}></div>
    <div class="flex-1 flex flex-col overflow-hidden">
        <div style="text-align: center;">
            <slot name="right"><em>missing &lt;slot=&quot;right&quot;&gt;</em></slot>
        </div>
    </div>
</div>
