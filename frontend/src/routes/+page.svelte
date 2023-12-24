<script lang="ts">
    import { env } from "$env/dynamic/public"
    import { onMount } from "svelte";

    onMount(() => {
        let xrequest = new XMLHttpRequest()
        xrequest.open("GET", `${env.PUBLIC_BACKEND_HOSTNAME}/userinfo`, false)
        xrequest.send()
        console.log(xrequest.statusText)
        console.log(xrequest.responseText)
    })

</script>
<h1>Panda Game Landing Page</h1>
<p>Backend server is at {env.PUBLIC_BACKEND_HOSTNAME}</p>
