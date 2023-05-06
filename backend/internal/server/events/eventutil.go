package events

import (
	"pandagame/internal/game"
	"pandagame/internal/redisconn"
	"pandagame/internal/util"

	socketio "github.com/googollee/go-socket.io"
)

// TODO this file needs error handling and logging

func broadcastRoomLobbyUpdate(gameID string, lobby *UILobby, server *socketio.Server) {
	msg, _ := util.ToJSONString(lobby)
	server.BroadcastToRoom(NS, gameID, string(LobbyUpdate), msg)
}

func broadcastRoomGameUpdate(gameID string, game *game.GameState, server *socketio.Server) {
	msg, _ := util.ToJSONString(game)
	server.BroadcastToRoom(NS, gameID, string(GameUpdate), msg)
}

func broadcastRoomGameStart(gameID string, game *game.GameState, server *socketio.Server) {
	msg, _ := util.ToJSONString(game)
	server.BroadcastToRoom(NS, gameID, string(GameStart), msg)
}

func broadcastRoomGameOver(gameID string, game *game.GameState, server *socketio.Server) {
	msg, _ := util.ToJSONString(game)
	server.BroadcastToRoom(NS, gameID, string(GameOver), msg)
}

func getLobbyState(gameID string, conn redisconn.RedisConn) *Lobby {
	l, _ := redisconn.GetThing[Lobby](gameID+lobbySfx, conn)
	return l
}

func storeLobbyState(gameID string, lobby *Lobby, conn redisconn.RedisConn) {
	redisconn.SetThing(gameID+lobbySfx, lobby, conn)
}

func getGameState(gameID string, conn redisconn.RedisConn) *game.GameState {
	g, _ := redisconn.GetThing[game.GameState](gameID+gameSfx, conn)
	return g
}

func storeGameState(gameID string, game *game.GameState, conn redisconn.RedisConn) {
	redisconn.SetThing(gameID+gameSfx, game, conn)
}

func emitMessage[T any](eventName string, message *T, conn socketio.Conn) {
	msg, _ := util.ToJSONString(message)
	conn.Emit(eventName, msg)
}

func handleError(err error, conn socketio.Conn) {
	conn.Emit(string(Warning), err.Error())
}
