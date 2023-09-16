# Panda Game Backend

## Technical Objectives

### Be one with the Zen of Go

Primarily, this means to use Go the Go way.

## Technical Overview

### Auth system

Panda Game uses basic auth for player authentication and sessions for player authorization. Player User Records are stored in MongoDB. Sessions are stored primarily in Redis. Sessions also get stored in websocket connections as a context object when a client connects to the websocket.