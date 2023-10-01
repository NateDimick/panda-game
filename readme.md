# Panda Game

A bamboo gardening game like *Takenoko*

## Technical overview

### Backend

* Golang
  * Socket.io server
  * Chi router (lightweight, simple, close to std lib)
  * redis for current games + matchmaking
  * mongodb for user records

### Frontend

* Sveltkit
  * [socket.io](https://github.com/socketio/socket.io) client
  * [pixi.js](https://github.com/pixijs/pixijs) game renderer

## Tools

* Taskfile `go install github.com/go-task/task/v3/cmd/task@latest`
* covreport `go install github.com/cancue/covreport@latest`
* Mockery `go install github.com/vektra/mockery/v2@latest`
