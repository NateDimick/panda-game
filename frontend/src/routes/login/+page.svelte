<script lang="ts">
    import { env } from "$env/dynamic/public"

    let username: string = ""
    let password: string = ""
    let errorMessage: string = ""

    function login() {
        sharedLogin("/login")
    }

    function guestLogin() {
        sharedLogin("/login/guest")
    }

    function sharedLogin(path: string) {
        fetch(`${env.PUBLIC_BACKEND_HOSTNAME}${path}`, {
            method: "POST",
            headers: {

            }
        }).then(resp => {
            if ( resp.status === 200 ) {
                window.location.href = "/"
            } else {
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
    <div>
        <button on:click={login}>Login</button>
        <button on:click={guestLogin}>Login as Guest</button>
    </div>
    <p>{errorMessage}</p>
</main>