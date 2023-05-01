package events

import (
	"pandagame/internal/mongoconn"
	"pandagame/internal/redisconn"
	"strings"

	socketio "github.com/googollee/go-socket.io"
	"go.uber.org/zap"
)

const NS string = "/"

type ServerEventType string

const (
	LobbyUpdate  ServerEventType = "LobbyUpdate"
	GameStart    ServerEventType = "GameStart"
	GameUpdate   ServerEventType = "GameUpdate"
	GameOver     ServerEventType = "GameOver"
	ActionPrompt ServerEventType = "ActionPrompt"
	Goodbye      ServerEventType = "Goodbye" // the server has forced the connection closed
	Warning      ServerEventType = "Warning" // the last message received was bad. Warn the client to do better
)

type ClientEventType string

const (
	JoinGame       ClientEventType = "JoinGame"
	LeaveGame      ClientEventType = "LeaveGame"
	GameChat       ClientEventType = "GameChat"
	TakeAction     ClientEventType = "TakeAction"
	CreateGame     ClientEventType = "CreateGame"
	StartGame      ClientEventType = "StartGame"
	ChangeSettings ClientEventType = "ChangeSettings"
)

type GameServer struct {
	*socketio.Server
	Redis redisconn.RedisConn
	Mongo mongoconn.CollectionConn
}

type ConnectionContext struct {
	PlayerID string
	UserName string
}

func (gs *GameServer) OnConnect(s socketio.Conn) error {
	// a new user connects
	headers := s.RemoteHeader()
	cookie := headers.Get("Cookie")
	kvPairs := strings.Split(cookie, ";")
	cookieMap := make(map[string]string)
	for _, mapping := range kvPairs {
		if k, v, found := strings.Cut(mapping, "="); found {
			cookieMap[k] = v
		}
	}
	cc := ConnectionContext{
		UserName: cookieMap["UserName"],
		PlayerID: cookieMap["PlayerId"],
	}
	s.SetContext(cc)
	zap.L().Info("New Player connected", zap.String("userName", cc.UserName), zap.String("playerId", cc.PlayerID), zap.String("cookie", cookie))
	return nil
}

func (gs *GameServer) OnDisconnect(s socketio.Conn, reason string) {
	cc := s.Context().(ConnectionContext)
	zap.L().Warn("Disconnection", zap.String("playerId", cc.PlayerID))
}

func (gs *GameServer) OnError(s socketio.Conn, e error) {
	cc := s.Context().(ConnectionContext)
	zap.L().Error("Error in connection, leaving all rooms and closing", zap.Error(e), zap.String("playerId", cc.PlayerID))
	s.LeaveAll()
	s.Emit(string(Goodbye))
	s.Close()
}

func (gs *GameServer) OnSearchForGame(s socketio.Conn, msg string) {
	// matchmaking is a way future feature
}

func (gs *GameServer) OnCancelSearchForGame(s socketio.Conn, msg string) {
	// matchmaking is a way future feature
}

func (gs *GameServer) OnCreateGameLobby(s socketio.Conn, msg string) {
	//
	// generate room name
	// join that room
	// emit lobby update to the room
}

func (gs *GameServer) OnJoinGame(s socketio.Conn, msg string) {
	cc := s.Context().(ConnectionContext)
	zap.L().Info("Player joining game room", zap.String("playerId", cc.PlayerID), zap.String("room", msg))
	s.Join(msg)
	gs.Server.BroadcastToRoom(NS, msg, string(LobbyUpdate), "todo event payload")
}

func (gs *GameServer) OnLeaveGame(s socketio.Conn, msg string) {
	cc := s.Context().(ConnectionContext)
	zap.L().Info("Player leaving game room", zap.String("playerId", cc.PlayerID), zap.String("room", msg))
	s.Leave(msg)
	gs.Server.BroadcastToRoom(NS, msg, string(LobbyUpdate), "todo event payload")
}

func (gs *GameServer) OnChatInRoom(s socketio.Conn, msg string) {
	cc := s.Context().(ConnectionContext)
	zap.L().Info("Player Chat Message", zap.String("playerId", cc.PlayerID), zap.String("chat", msg))
	// unmarshal chat struct from json
	// get game state from redis
	// add chat message
	// store game state
	// emit game state to room
}

func (gs *GameServer) OnStartGame(s socketio.Conn, msg string) {
	//cc := s.Context().(ConnectionContext)
	// check if user is empowered to start games
}

func (gs *GameServer) OnTakeTurnAction(s socketio.Conn, msg string) {
	cc := s.Context().(ConnectionContext)
	zap.L().Info("Player is taking action", zap.String("playerId", cc.PlayerID), zap.String("action", msg))
	// get game state from cache
	// convert msg to PromptResponse
	// do game thing
	// if the prompt is for the next player, then do something like this:
	gs.Server.ForEach(NS, "todo room id", func(c socketio.Conn) {
		cc := c.Context().(ConnectionContext)
		if cc.PlayerID == "todo - next player id" {
			c.Emit("ActionPrompt", "todo - the prompt the game engine produced, json marshalled")
		}
	})
}
