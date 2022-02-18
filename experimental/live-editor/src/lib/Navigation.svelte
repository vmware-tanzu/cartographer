<script>
    import {getContext} from "svelte";
    import Help from "./Help.svelte";

    const { open } = getContext('simple-modal');

    const showHelp = () => {
        open(Help)
    }

</script>


<nav class="font-sans flex flex-col text-center sm:flex-row sm:text-left sm:justify-between py-2 px-6 bg-sky-700 text-sky-50 shadow sm:items-baseline w-full">
    <div class="mb-2 sm:mb-0">
        <h1 class="text-lg uppercase">Cartographer Live Editor</h1>
    </div>
    <div>
        <button class="text-lg no-underline text-grey-darkest hover:text-orange-300 ml-2" on:click={showHelp}>
            Help
        </button>
    </div>
</nav>