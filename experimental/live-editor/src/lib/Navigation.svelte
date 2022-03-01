<script>
    import {getContext} from "svelte";
    import Help from "./Help.svelte";
    import CopyToClipboard from "svelte-copy-to-clipboard";
    import {document} from "../store.js";
    import Shared from "./Shared.svelte";

    const { open } = getContext('simple-modal');
    let url

    const getUrl = () => window.location.href
    $: url = getUrl($document)

</script>


<nav class="font-sans flex flex-col text-center sm:flex-row sm:text-left sm:justify-between py-2 px-6 bg-sky-700 text-sky-50 shadow sm:items-baseline w-full">
    <div class="mb-2 sm:mb-0 flex sm:items-baseline">
        <h1 class="text-lg uppercase">Cartographer Live Editor</h1>
        &nbsp;&nbsp;
        <p class="text-sm">v0.0.1</p>
    </div>
    <div>

        <CopyToClipboard text={url} on:copy={() => open(Shared)} on:fail={() => alert("Something went wrong with the copy to clipboard, sorry!")} let:copy>
            <button class="text-lg no-underline text-grey-darkest hover:text-orange-300 ml-2" on:click={copy}>
                Share
            </button>
        </CopyToClipboard>

        <button class="text-lg no-underline text-grey-darkest hover:text-orange-300 ml-2" on:click={() => open(Help)}>
            Help
        </button>
    </div>
</nav>