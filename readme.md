# Panda Game

A bamboo gardening game like *Takenoko*

## Technical overview

### Backend

* Golang
  * Socket.io server
  * Chi router (lightweight, simple, close to std lib)
  * redis for current games + matchmaking
  * mongodb for user records
  * zap logging

### Frontend

* Sveltkit
  * [socket.io](https://github.com/socketio/socket.io) client
  * [pixi.js](https://github.com/pixijs/pixijs) game renderer
