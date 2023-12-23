<script lang="ts">
    import Cookies from "js-cookie"
    import { onMount } from "svelte";

    // this page just removes the cookie and redirect back home
    console.log("logout")
    onMount(() => {
        console.log("logout - onMount")
        Cookies.remove("pandaGameSession")
        window.location.replace("/")
    })
</script>

<main>
    <p>If you see this text, then something probably went wrong. You may or may not be logged out. Check your browser cookies to find out.</p>
    <a href="/">Emergency return home link</a>
</main>