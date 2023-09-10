package events

import (
	"context"
	"log/slog"
	"pandagame/internal/auth"
	"pandagame/internal/game"
	"pandagame/internal/mongoconn"
	"pandagame/internal/redisconn"
	"pandagame/internal/util"
	"slices"
	"strings"

	socketio "github.com/googollee/go-socket.io"
	"github.com/redis/go-redis/v9"
)

const NS string = "/" // namespace - there is only one namespace in this application

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
	Reprompt       ClientEventType = "RePrompt"
	CreateGame     ClientEventType = "CreateGame"
	StartGame      ClientEventType = "StartGame"
	ChangeSettings ClientEventType = "ChangeSettings"
)

const (
	gamePfx  string = "g-" // game prefix for redis keys
	lobbyPfx string = "l-" // lobby prefix for redis keys
)

type GameServer struct {
	*socketio.Server
	Redis redisconn.RedisConn
	Mongo mongoconn.CollectionConn
}

func (gs *GameServer) OnConnect(s socketio.Conn) error {
	defer deferRecover(s)
	// a new user connects
	slog.Info("Player connection happening", slog.String("sid", s.ID()))
	headers := s.RemoteHeader()
	cookie := headers.Get("Cookie")
	slog.Info("Found connection cookie", slog.String("cookie", cookie))
	kvPairs := strings.Split(cookie, ";")
	cookieMap := make(map[string]string)
	for _, mapping := range kvPairs {
		if k, v, found := strings.Cut(mapping, "="); found {
			cookieMap[k] = v
		}
	}

	us, err := redisconn.GetThing[auth.UserSession](cookieMap["pandaGameSession"]+"-session", gs.Redis)
	if err != nil {
		s.Emit(string(Warning), "please log in to connect")
		return err
	}
	s.SetContext(us)
	slog.Info("New Player connected", slog.String("userName", us.Name), slog.String("playerId", us.PlayerID), slog.String("cookie", cookie))
	// TODO - re-join rooms if coming back from disconnect
	return nil
}

func (gs *GameServer) OnDisconnect(s socketio.Conn, reason string) {
	defer deferRecover(nil)
	cc := getConnectionContext(s)
	slog.Warn("Disconnection", slog.String("playerId", cc.PlayerID))
	for _, gid := range s.Rooms() {
		l := getLobbyState(gid, gs.Redis)
		if l.Host.PlayerID == cc.PlayerID {
			// TODO need to pick a new host out of the players
		}
	}
}

func (gs *GameServer) OnError(s socketio.Conn, e error) {
	defer deferRecover(s)
	cc := getConnectionContext(s)
	slog.Error("Error in connection, leaving all rooms and closing", slog.String("error", e.Error()), slog.String("playerId", cc.PlayerID))
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
	// check if user is allowed to create lobbies
	if !cc.Empowered {
		s.Emit(string(Warning), "You are not allowed to create lobbies")
		return
	}
	// generate room name
	unique := false
	var gid string
	for !unique {
		gid = NewGameID()
		if err := gs.Redis.Get(context.Background(), lobbyPfx+gid).Err(); err == redis.Nil {
			unique = true
		}
	}
	// join that room
	s.Join(gid)

	l := &Lobby{
		Host:       cc,
		Players:    []auth.UserSession{cc},
		Spectators: make([]auth.UserSession, 0),
	}

	storeLobbyState(gid, l, gs.Redis)

	// emit lobby update to the room
	broadcastRoomLobbyUpdate(gid, l.RemoveIDs(), gs.Server)
}

func (gs *GameServer) OnJoinGame(s socketio.Conn, msg string) {
	defer deferRecover(s)
	cc := getConnectionContext(s)
	slog.Info("Player joining game room", slog.String("playerId", cc.PlayerID), slog.String("room", msg))
	s.Join(msg)
	l := getLobbyState(msg, gs.Redis)
	if len(l.Players) < 4 && !l.Started {
		l.Players = append(l.Players, cc)
	} else {
		l.Spectators = append(l.Spectators, cc)
	}

	storeLobbyState(msg, l, gs.Redis)

	broadcastRoomLobbyUpdate(msg, l.RemoveIDs(), gs.Server)
}

func (gs *GameServer) OnLeaveGame(s socketio.Conn, msg string) {
	defer deferRecover(s)
	cc := getConnectionContext(s)
	slog.Info("Player leaving game room", slog.String("playerId", cc.PlayerID), slog.String("room", msg))
	s.Leave(msg)

	l := getLobbyState(msg, gs.Redis)
	if l.Host.PlayerID == cc.PlayerID {
		// find new host
		for _, occ := range l.Players {
			if occ.PlayerID != cc.PlayerID {
				l.Host = occ
				break
			}
		}
	}
	// check if user is in player list, and remove (if game is not started, move up spectator to player slot)
	i := slices.IndexFunc(l.Players, func(us auth.UserSession) bool { return us.PlayerID == cc.PlayerID })
	if i >= 0 {
		l.Players = slices.Delete(l.Players, i, i+1)
		if len(l.Spectators) > 0 {
			l.Players = append(l.Players, l.Spectators[0])
			l.Spectators = l.Spectators[1:]
		}
	}
	// check if user is in spectator list, and remove
	j := slices.IndexFunc(l.Spectators, func(us auth.UserSession) bool { return us.PlayerID == cc.PlayerID })
	if i >= 0 {
		l.Players = slices.Delete(l.Spectators, j, j+1)
	}
	storeLobbyState(msg, l, gs.Redis)
	broadcastRoomLobbyUpdate(msg, l.RemoveIDs(), gs.Server)
}

func (gs *GameServer) OnChatInRoom(s socketio.Conn, msg string) {
	defer deferRecover(s)
	cc := getConnectionContext(s)
	slog.Info("Player Chat Message", slog.String("playerId", cc.PlayerID), slog.String("chat", msg))
	// unmarshal chat struct from json
	cm, err := util.FromJSONString[game.ChatMessage](msg)
	if err != nil {
		return
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
			Name:               p.Name,
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
	gs.Server.ForEach(NS, msg, func(c socketio.Conn) {
		cc := c.Context().(auth.UserSession)
		if cc.PlayerID == g.CurrentTurn.PlayerID {
			emitMessage("ActionPrompt", &prompt, s)
		}
	})
}

func (gs *GameServer) OnTakeTurnAction(s socketio.Conn, msg string) {
	defer deferRecover(s)
	cc := getConnectionContext(s)
	slog.Info("Player is taking action", slog.String("playerId", cc.PlayerID), slog.String("action", msg))
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
	gs.Server.ForEach(NS, pr.Gid, func(c socketio.Conn) {
		cc := c.Context().(auth.UserSession)
		if cc.PlayerID == g.CurrentTurn.PlayerID {
			emitMessage("ActionPrompt", &prompt, c)
		}
	})

	// broadcast gamestate to room
	broadcastRoomGameUpdate(pr.Gid, g, gs.Server)
}
