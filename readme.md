# Panda Game

A bamboo gardening game like *Takenoko*, in your browser! That's the goal.

## Technical overview

### Backend

* Golang
  * custom websocket event protocol with Gorilla Websockets
  * Chi router (lightweight, simple, close to std lib)
  * nats for event pub/sub, and kv
  * surrealdb for auth, record storage

### Frontend

* htmx + templ + tailwind (probably) + [pixi.js](https://github.com/pixijs/pixijs) (probably)

## Tools

* Taskfile `go install github.com/go-task/task/v3/cmd/task@latest`
* covreport `go install github.com/cancue/covreport@latest`
* Mockery `go install github.com/vektra/mockery/v2@latest`
* Templ `go install github.com/a-h/templ/cmd/templ@latest`
