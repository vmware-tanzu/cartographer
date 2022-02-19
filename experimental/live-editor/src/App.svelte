<script>
    import FlexPane from "./lib/FlexPane.svelte";
    import Viz from "./lib/Viz.svelte";
    import Modal from 'svelte-simple-modal';
    import Navigation from "./lib/Navigation.svelte";
    import {compressedState, document} from "./store.js";
    import {inflate} from "pako";
    import {toUint8Array} from "js-base64";
    import Monaco from "./lib/Monaco.svelte";

    let loaded = false

    const pushURL = (isLoaded, state) => {
        if (!isLoaded || !state) {
            return
        }
        let pageUrl = new URL(window.location.href)
        pageUrl.searchParams.set("pako", state)
        window.history.pushState('', '', pageUrl);
    }

    $: pushURL(loaded, $compressedState)


    const onPageLoad = () => {
        let pageUrl = new URL(window.location.href)
        let pako = pageUrl.searchParams.get("pako")
        if (pako) {
            $document = inflate(toUint8Array(pako), {to: 'string'})
        }
        loaded = true
    }
</script>

<svelte:window on:load={onPageLoad}/>

<main class="h-screen">
    <Modal
            unstyled={true}
            closeButton={false}
            classBg="fixed top-0 left-0 w-screen h-screen flex flex-col justify-center bg-gray-400/[.6] z-50"
            classWindowWrap="relative m-2 max-h-full"
            classWindow="relative w-1/3 max-w-full max-h-full my-2 mx-auto text-black border border-sky-600 shadow-lg bg-white"
            classContent="relative p-2 overflow-auto"
    >

        <div class="h-full flex flex-col overflow-hidden">
            <Navigation/>
            {#if loaded}
                <FlexPane>
                    <Monaco slot="left" class="h-full m-2"/>
                    <Viz slot="right" class="content-center"/>
                </FlexPane>
            {/if}
        </div>
    </Modal>
</main>

