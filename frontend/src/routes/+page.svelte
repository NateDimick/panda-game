<script lang="ts">
    import { env } from "$env/dynamic/public"
    import { UserInfo } from "$lib/user";

</script>

<h1>Panda Game Landing Page</h1>
<p>Backend server is at {env.PUBLIC_BACKEND_HOSTNAME}</p>

<div>
    {#if UserInfo.name === null}
    <a href="/login">Login</a>
    <a href="/register">Register</a>
    {:else}
    <p>Hello, {UserInfo.name}</p>
    <a href="/">Create a Game</a>
    <a href="/">Join a Game</a>
    {/if}
</div>

