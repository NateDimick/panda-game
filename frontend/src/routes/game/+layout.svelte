<script lang="ts">
    import { Socket } from '$lib/socket';
    import { onDestroy, onMount } from 'svelte';

    onMount(() => {
        Socket.reset()
    })

    onDestroy(() => {
        Socket.close()
    })
</script>

<main>
    This is the game layout
</main>

<slot></slot>