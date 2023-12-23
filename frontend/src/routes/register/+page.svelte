<script lang="ts">
    import { env } from "$env/dynamic/public"

    let username: string = ""
    let password: string = ""
    let errorMessage: string = ""

    function register() {
        let body: string = `username=${username}&password=${password}`
        fetch(`${env.PUBLIC_BACKEND_HOSTNAME}/register`, {method: "POST", headers: {"Content-Type": "application/x-www-form-urlencoded"}, body: body}).then(resp => {
            if ( resp.status === 201 ) {
                // redirect to /login
                window.location.href = "/"
            } else {
                // TODO: better error message handling
                errorMessage = resp.statusText
            }
        })
    }

</script>

<main>
    <label for="username">Username</label>
    <input type="text" name="username" id="" bind:value={username}>
    <label for="password">Password</label>
    <input type="password" name="password" id="" bind:value={password}>
    <button on:click={register}>Register</button>
    <p>{errorMessage}</p>
</main>