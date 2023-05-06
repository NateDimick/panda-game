package events

import (
	"fmt"
	"pandagame/internal/game"
	"pandagame/internal/redisconn"
	"pandagame/internal/util"

	socketio "github.com/googollee/go-socket.io"
	"go.uber.org/zap"
)

// TODO this file needs error handling and logging

func getConnectionContext(s socketio.Conn) ConnectionContext {
	if s == nil {
		return ConnectionContext{
			PlayerID: "No Connection",
			UserName: "No Connection",
		}
	}
	if s.Context() == nil {
		return ConnectionContext{
			PlayerID: fmt.Sprintf("No Context for %s", s.ID()),
			UserName: fmt.Sprintf("No Context for %s", s.ID()),
		}
	}
	cc := s.Context().(ConnectionContext)
	return cc
}

func broadcastRoomLobbyUpdate(gameID string, lobby *UILobby, server *socketio.Server) {
	msg, _ := util.ToJSONString(lobby)
	server.BroadcastToRoom(GNS, gameID, string(LobbyUpdate), msg)
}

func broadcastRoomGameUpdate(gameID string, game *game.GameState, server *socketio.Server) {
	msg, _ := util.ToJSONString(game)
	server.BroadcastToRoom(GNS, gameID, string(GameUpdate), msg)
}

func broadcastRoomGameStart(gameID string, game *game.GameState, server *socketio.Server) {
	msg, _ := util.ToJSONString(game)
	server.BroadcastToRoom(GNS, gameID, string(GameStart), msg)
}

func broadcastRoomGameOver(gameID string, game *game.GameState, server *socketio.Server) {
	msg, _ := util.ToJSONString(game)
	server.BroadcastToRoom(GNS, gameID, string(GameOver), msg)
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

func deferRecover(conn socketio.Conn) {
	if err := recover(); err != nil {
		zap.L().Error(fmt.Sprintf("event handler panic: %+v", err))
		if conn != nil {
			handleError(fmt.Errorf("%+v", err), conn)
		}
	}
}
