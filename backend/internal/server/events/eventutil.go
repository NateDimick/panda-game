package events

import (
	"fmt"
	"log/slog"
	"pandagame/internal/auth"
	"pandagame/internal/game"
	"pandagame/internal/redisconn"
	"pandagame/internal/util"

	"github.com/njones/socketio"
	"github.com/njones/socketio/serialize"
)

// TODO this file needs error handling and logging

func getSession(s *socketio.SocketV4, r redisconn.RedisConn) (*auth.UserSession, error) {
	sessionCookie, err := s.Request().Cookie("pandaGameSession")
	if err != nil {
		return nil, err
	}
	slog.Info("Found connection cookie", slog.String("cookie", sessionCookie.Value))
	us, err := redisconn.GetThing[auth.UserSession](sessionCookie.Value+"-session", r)
	if err != nil {
		return nil, err
	}
	return us, nil
}

func broadcastEventToGameRoom[T any](gameID string, eventType ServerEventType, eventBody *T, socket *socketio.SocketV4) {
	msg, _ := util.ToJSONString(eventBody)
	socket.To(gameID).Emit(string(eventType), serialize.String(msg))
}

func getLobbyState(gameID string, conn redisconn.RedisConn) *Lobby {
	l, _ := redisconn.GetThing[Lobby](lobbyPfx+gameID, conn)
	return l
}

func storeLobbyState(gameID string, lobby *Lobby, conn redisconn.RedisConn) {
	redisconn.SetThing(lobbyPfx+gameID, lobby, conn)
}

func getGameState(gameID string, conn redisconn.RedisConn) *game.GameState {
	g, _ := redisconn.GetThing[game.GameState](gamePfx+gameID, conn)
	return g
}

func storeGameState(gameID string, game *game.GameState, conn redisconn.RedisConn) {
	redisconn.SetThing(gamePfx+gameID, game, conn)
}

func handleError(err error, conn *socketio.SocketV4) {
	conn.Emit(string(Warning), serialize.Error(err))
}

func deferRecover(conn *socketio.SocketV4) {
	if err := recover(); err != nil {
		slog.Error(fmt.Sprintf("event handler panic: %+v", err))
		// if conn != nil {
		// 	handleError(fmt.Errorf("%+v", err), conn)
		// }
	}
}
