<script lang="ts">
    import { env } from "$env/dynamic/public";
    import { onMount } from "svelte";

    // this page just removes the cookie and redirect back home
    console.log("logout")
    onMount(() => {
        console.log("logout - onMount")
        let xrequest = new XMLHttpRequest()
        xrequest.open("POST", `${env.PUBLIC_BACKEND_HOSTNAME}/logout`, false)
        xrequest.withCredentials = true
        xrequest.send()
        window.location.replace("/")
    })
</script>

<main>
    <p>If you see this text, then something probably went wrong. You may or may not be logged out. Check your browser cookies to find out.</p>
    <a href="/">Emergency return home link</a>
</main>