package events

import (
	"context"
	"fmt"
	"pandagame/internal/game"
	"pandagame/internal/mongoconn"
	"pandagame/internal/redisconn"
	"pandagame/internal/util"
	"strings"

	socketio "github.com/googollee/go-socket.io"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const NS string = "/"
const GNS string = "/game"

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

const (
	gameSfx  string = "-g"
	lobbySfx string = "-l"
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
	defer deferRecover(s)
	// a new user connects
	fmt.Println("WHAT THE FUCK IS HAPPENING ON CONNECT?!?!?!?!?!?!? ", s.ID())
	zap.L().Info("Player connection happening", zap.String("sid", s.ID()))
	headers := s.RemoteHeader()
	cookie := headers.Get("Cookie")
	zap.L().Info("Found connection cookie", zap.String("cookie", cookie))
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
	defer deferRecover(nil)
	cc := getConnectionContext(s)
	zap.L().Warn("Disconnection", zap.String("playerId", cc.PlayerID))
}

func (gs *GameServer) OnError(s socketio.Conn, e error) {
	defer deferRecover(s)
	cc := getConnectionContext(s)
	zap.L().Error("Error in connection, leaving all rooms and closing", zap.Error(e), zap.String("playerId", cc.PlayerID))
	if s != nil {
		s.LeaveAll()
		s.Emit(string(Goodbye))
		s.Close()
	}
}

func (gs *GameServer) OnSearchForGame(s socketio.Conn, msg string) {
	defer deferRecover(s)
	// matchmaking is a way future feature
}

func (gs *GameServer) OnCancelSearchForGame(s socketio.Conn, msg string) {
	defer deferRecover(s)
	// matchmaking is a way future feature
}

func (gs *GameServer) OnCreateGameLobby(s socketio.Conn, msg string) {
	defer deferRecover(s)
	cc := getConnectionContext(s)
	// generate room name
	unique := false
	var gid string
	for !unique {
		gid = NewGameID()
		if err := gs.Redis.Get(context.Background(), gid+lobbySfx).Err(); err == redis.Nil {
			unique = true
		}
	}
	// join that room
	s.Join(gid)

	l := &Lobby{
		Host:    cc,
		Players: []ConnectionContext{cc},
	}

	storeLobbyState(gid, l, gs.Redis)

	// emit lobby update to the room
	broadcastRoomLobbyUpdate(gid, l.RemoveIDs(), gs.Server)
}

func (gs *GameServer) OnJoinGame(s socketio.Conn, msg string) {
	defer deferRecover(s)
	cc := getConnectionContext(s)
	zap.L().Info("Player joining game room", zap.String("playerId", cc.PlayerID), zap.String("room", msg))
	s.Join(msg)
	gs.Server.BroadcastToRoom(GNS, msg, string(LobbyUpdate), "todo event payload")
}

func (gs *GameServer) OnLeaveGame(s socketio.Conn, msg string) {
	defer deferRecover(s)
	cc := getConnectionContext(s)
	zap.L().Info("Player leaving game room", zap.String("playerId", cc.PlayerID), zap.String("room", msg))
	s.Leave(msg)
	gs.Server.BroadcastToRoom(GNS, msg, string(LobbyUpdate), "todo event payload")
}

func (gs *GameServer) OnChatInRoom(s socketio.Conn, msg string) {
	defer deferRecover(s)
	cc := getConnectionContext(s)
	zap.L().Info("Player Chat Message", zap.String("playerId", cc.PlayerID), zap.String("chat", msg))
	// unmarshal chat struct from json
	cm, err := util.FromJSONString[game.ChatMessage](msg)
	if err != nil {

	}
	// get game state from redis
	g := getGameState(cm.Gid, gs.Redis)
	// add chat message
	g.ChatLog = append(g.ChatLog, *cm)
	// store game state
	storeGameState(cm.Gid, g, gs.Redis)
	// emit game state to room
	broadcastRoomGameUpdate(cm.Gid, g, gs.Server)
}

func (gs *GameServer) OnStartGame(s socketio.Conn, msg string) {
	defer deferRecover(s)
	cc := getConnectionContext(s)
	// check if user is empowered to start games
	l := getLobbyState(msg, gs.Redis)

	if cc.PlayerID != l.Host.PlayerID {
		s.Emit(string(Warning), "You're not the Host")
		return
	}

	if len(l.Players) < 2 {
		s.Emit(string(Warning), "Not enough Players")
		return
	}

	g := game.NewGame()
	players := make([]*game.Player, 0)
	for i, p := range l.Players {
		player := &game.Player{
			Name:               p.UserName,
			ID:                 p.PlayerID,
			Order:              i,
			Bamboo:             make(game.BambooReserve),
			Improvements:       make(game.ImprovementReserve),
			Objectives:         make([]game.Objective, 0),
			CompleteObjectives: make([]game.Objective, 0),
		}
		players = append(players, player)
	}
	g.AddPlayers(players)
	g.NextTurn()
	prompt := g.NextChooseActionPrompt()
	g.CurrentTurn.CurrentPrompt = prompt

	storeGameState(msg, g, gs.Redis)

	l.Started = true

	storeLobbyState(msg, l, gs.Redis)

	broadcastRoomGameStart(msg, g, gs.Server)

	// send prompt to player
	gs.Server.ForEach(GNS, msg, func(c socketio.Conn) {
		cc := c.Context().(ConnectionContext)
		if cc.PlayerID == g.CurrentTurn.PlayerID {
			emitMessage("ActionPrompt", &prompt, s)
		}
	})
}

func (gs *GameServer) OnTakeTurnAction(s socketio.Conn, msg string) {
	defer deferRecover(s)
	cc := getConnectionContext(s)
	zap.L().Info("Player is taking action", zap.String("playerId", cc.PlayerID), zap.String("action", msg))
	// convert msg to PromptResponse
	pr, err := util.FromJSONString[game.PromptResponse](msg)
	if err != nil {
		handleError(err, s)
		return
	}
	// get game state from cache
	g := getGameState(pr.Gid, gs.Redis)

	if cc.PlayerID != g.CurrentTurn.PlayerID {
		s.Emit(string(Warning), "It's not your turn")
		return
	}

	// do game thing
	prompt := game.GameFlow(g, *pr)

	// store game state
	storeGameState(pr.Gid, g, gs.Redis)

	if g.CurrentTurn.PlayerID == cc.PlayerID {
		// send prompt to current player
		emitMessage("ActionPrompt", &prompt, s)
		return
	}
	// else, the prompt is for the next player, then do something like this:
	gs.Server.ForEach(GNS, pr.Gid, func(c socketio.Conn) {
		cc := c.Context().(ConnectionContext)
		if cc.PlayerID == g.CurrentTurn.PlayerID {
			emitMessage("ActionPrompt", &prompt, c)
		}
	})

	// broadcast gamestate to room
	broadcastRoomGameUpdate(pr.Gid, g, gs.Server)
}
