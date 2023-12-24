<script lang="ts">
    import { env } from "$env/dynamic/public"

    let username: string = ""
    let password: string = ""
    let errorMessage: string = ""

    function login() {
        sharedLogin("login")
    }

    function guestLogin() {
        sharedLogin("login/guest")
    }

    function sharedLogin(path: string) {
        let rawBasicAuth = `${username}:${password}`
        let xrequest = new XMLHttpRequest()
        xrequest.open("POST", `${env.PUBLIC_BACKEND_HOSTNAME}/${path}`, false)
        xrequest.setRequestHeader("Authorization", `Basic ${btoa(rawBasicAuth)}`)
        xrequest.send()

        let status = xrequest.status
        if ( status === 200 ) {
            window.location.href = "/"
        } else {
            errorMessage = xrequest.statusText
        }
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