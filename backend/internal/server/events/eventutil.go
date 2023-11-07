package events

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"pandagame/internal/auth"
	"pandagame/internal/game"
	"pandagame/internal/pandaplex"
	"pandagame/internal/redisconn"
	"pandagame/internal/util"
	"slices"
)

// TODO this file needs error handling and logging

func getSession(plexer pandaplex.PlexerInternal, r redisconn.RedisConn) (*auth.UserSession, error) {
	sessionIndex := slices.IndexFunc(plexer.Cookies(), func(c *http.Cookie) bool { return c.Name == "pandaGameSession" })
	if sessionIndex == -1 {
		return nil, errors.New("No pandaGameSession cookie")
	}
	sessionCookie := plexer.Cookies()[sessionIndex] //("pandaGameSession")

	slog.Info("Found connection cookie", slog.String("cookie", sessionCookie.Value))
	us, err := redisconn.GetThing[auth.UserSession]("s-"+sessionCookie.Value, r)
	if err != nil {
		return nil, err
	}
	return us, nil
}

func broadcastEventToGameRoom[T any](gameID string, eventType ServerEventType, eventBody *T, plexer pandaplex.PlexerInternal) {
	e := &ServerEvent{
		Type:    eventType,
		Payload: eventBody,
	}
	msg, _ := util.ToJSONString(e)
	plexer.SendToRoom(msg, gameID)
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

func deferRecover(conn pandaplex.PlexerInternal) {
	if err := recover(); err != nil {
		slog.Error(fmt.Sprintf("event handler panic: %+v", err))
		// if conn != nil {
		// 	handleError(fmt.Errorf("%+v", err), conn)
		// }
	}
}
