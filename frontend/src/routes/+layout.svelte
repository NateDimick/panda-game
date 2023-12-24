<script lang="ts">
    import { env } from "$env/dynamic/public"
    import { UserInfo } from "$lib/user";
    import { onMount } from "svelte";

    

    onMount(() => {
        let xrequest = new XMLHttpRequest()
        xrequest.open("GET", `${env.PUBLIC_BACKEND_HOSTNAME}/userinfo`, false)
        xrequest.withCredentials = true
        xrequest.send()
        if ( xrequest.status === 200 ) {
            console.log(xrequest.statusText)
            console.log(xrequest.responseText)
            UserInfo.update(JSON.parse(xrequest.responseText))
        }
    })

</script>


<slot></slot>