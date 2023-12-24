import { env } from "$env/dynamic/public"

let ws: WebSocket = $state(new WebSocket(`${env.PUBLIC_BACKEND_HOSTNAME}/ws`))
let handler: ((ev: MessageEvent) => any) | undefined = undefined

export let Socket = {
    send: (message: string) => {ws.send(message)},
    close: () => {ws.close},
    reset: () => {
        ws.close()
        ws = new WebSocket(`${env.PUBLIC_BACKEND_HOSTNAME}/ws`)
    },
    subscribe: (messageHandler: (ev: MessageEvent) => any) => {
        if ( handler !== undefined ) {
            ws.removeEventListener("message", handler)
        }
        ws.addEventListener("message", messageHandler)
        handler = messageHandler
    },
    unsubscribe: () => {
        if ( handler === undefined ) {
            return
        }
        ws.removeEventListener("message", handler)
        handler = undefined
    }
}