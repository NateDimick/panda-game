package events

import (
	"errors"
	"log/slog"
	"pandagame/internal/auth"
	"pandagame/internal/game"
	"pandagame/internal/mongoconn"
	"pandagame/internal/redisconn"
	"pandagame/internal/util"
	"slices"

	"github.com/njones/socketio"
	"github.com/njones/socketio/serialize"
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
	*socketio.ServerV4
	Redis redisconn.RedisConn
	Mongo mongoconn.CollectionConn
}

func (gs *GameServer) OnConnect(s *socketio.SocketV4) error {
	defer deferRecover(s)
	// a new user connects
	slog.Info("Player connection happening", slog.String("sid", string(s.ID())))
	us, err := getSession(s, gs.Redis)
	if err != nil {
		s.Emit(string(Warning), serialize.String("please log in to connect"))
		return err
	}

	// join a personal room to receive personal messages
	if err := s.Join(us.PlayerID); err != nil {
		slog.Error("Could not join personal room", slog.String("error", err.Error()))
	}

	// need to register all callbacks here
	s.OnDisconnect(func(reason string) {
		// slog the input string
		slog.Info("Disconnect", slog.String("user", us.Name), slog.String("id", us.PlayerID), slog.String("reason", reason))

		// leave personal room
		s.Leave(us.PlayerID)

		// leave game room
		// TODO
	})
	s.On(string(TakeAction), EventHandler{s: s, f: gs.OnTakeTurnAction})
	// chat
	// join game
	// start game

	slog.Info("New Player connected", slog.String("userName", us.Name), slog.String("playerId", us.PlayerID))
	// TODO - re-join rooms if coming back from disconnect
	return nil
}

// func (gs *GameServer) OnDisconnect(reason string) {
// 	defer deferRecover(nil)
// 	cc := getSession(s)
// 	slog.Warn("Disconnection", slog.String("playerId", cc.PlayerID))
// 	for _, gid := range s.Rooms() {
// 		l := getLobbyState(gid, gs.Redis)
// 		if l.Host.PlayerID == cc.PlayerID {
// 			// TODO need to pick a new host out of the players
// 		}
// 	}
// }

// new socketio lib requires this clunky interface for handling events
type EventHandler struct {
	s *socketio.SocketV4
	f func(*socketio.SocketV4, ...interface{}) error
}

func (eh EventHandler) Callback(data ...interface{}) error {
	slog.Info("event data", slog.Any("data", data))
	return eh.f(eh.s, data...)
}

// func (gs *GameServer) OnError(s socketio.Conn, e error) {
// 	defer deferRecover(s)
// 	cc := getSession(s)
// 	slog.Error("Error in connection, leaving all rooms and closing", slog.String("error", e.Error()), slog.String("playerId", cc.PlayerID))
// 	if s != nil {
// 		s.LeaveAll()
// 		s.Emit(string(Goodbye))
// 		s.Close()
// 	}
// }

func (gs *GameServer) OnSearchForGame(s *socketio.SocketV4, data ...interface{}) error {
	defer deferRecover(s)
	// matchmaking is a way future feature
	return nil
}

func (gs *GameServer) OnCancelSearchForGame(s *socketio.SocketV4, data ...interface{}) error {
	defer deferRecover(s)
	// matchmaking is a way future feature
	return nil
}

func (gs *GameServer) OnCreateGameLobby(s *socketio.SocketV4, data ...interface{}) error {
	defer deferRecover(s)
	us, err := getSession(s, gs.Redis)
	if err != nil {
		return err
	}
	// check if user is allowed to create lobbies
	if !us.Empowered {
		s.To(us.PlayerID).Emit(string(Warning), serialize.String("You are not allowed to create lobbies"))
		return errors.New("must be empowered to create lobbies")
	}
	// generate room name
	unique := false
	var gid string
	for !unique {
		gid = NewGameID()
		if _, err := redisconn.GetThing[Lobby](lobbyPfx+gid, gs.Redis); err == redis.Nil {
			unique = true
		}
	}
	// join that room
	s.Join(gid)

	l := &Lobby{
		Host:       us,
		Players:    []*auth.UserSession{us},
		Spectators: make([]*auth.UserSession, 0),
	}

	storeLobbyState(gid, l, gs.Redis)

	// emit lobby update to the room
	broadcastEventToGameRoom(gid, LobbyUpdate, l.RemoveIDs(), s)
	return nil
}

func (gs *GameServer) OnJoinGame(s *socketio.SocketV4, data ...interface{}) error {
	defer deferRecover(s)
	us, err := getSession(s, gs.Redis)
	if err != nil {
		return err
	}
	msg, ok := data[0].(string)
	if !ok {
		return errors.New("bad format join room message")
	}
	slog.Info("Player joining game room", slog.String("playerId", us.PlayerID), slog.String("room", msg))
	s.Join(msg)
	l := getLobbyState(msg, gs.Redis)
	if len(l.Players) < 4 && !l.Started {
		l.Players = append(l.Players, us)
	} else {
		l.Spectators = append(l.Spectators, us)
	}

	storeLobbyState(msg, l, gs.Redis)

	broadcastEventToGameRoom(msg, LobbyUpdate, l.RemoveIDs(), s)
	return nil
}

func (gs *GameServer) OnLeaveGame(s *socketio.SocketV4, data ...interface{}) error {
	defer deferRecover(s)
	us, err := getSession(s, gs.Redis)
	if err != nil {
		return err
	}
	msg, ok := data[0].(string)
	if !ok {
		return errors.New("bad format leave room message")
	}
	slog.Info("Player leaving game room", slog.String("playerId", us.PlayerID), slog.String("room", msg))
	s.Leave(msg)

	l := getLobbyState(msg, gs.Redis)
	if l.Host.PlayerID == us.PlayerID {
		// find new host
		for _, occ := range l.Players {
			if occ.PlayerID != us.PlayerID {
				l.Host = occ
				break
			}
		}
	}
	// check if user is in player list, and remove (if game is not started, move up spectator to player slot)
	i := slices.IndexFunc(l.Players, func(us2 *auth.UserSession) bool { return us2.PlayerID == us.PlayerID })
	if i >= 0 {
		l.Players = slices.Delete(l.Players, i, i+1)
		if len(l.Spectators) > 0 {
			l.Players = append(l.Players, l.Spectators[0])
			l.Spectators = l.Spectators[1:]
		}
	}
	// check if user is in spectator list, and remove
	j := slices.IndexFunc(l.Spectators, func(us2 *auth.UserSession) bool { return us2.PlayerID == us.PlayerID })
	if i >= 0 {
		l.Players = slices.Delete(l.Spectators, j, j+1)
	}
	storeLobbyState(msg, l, gs.Redis)
	broadcastEventToGameRoom(msg, LobbyUpdate, l.RemoveIDs(), s)
	return nil
}

func (gs *GameServer) OnChatInRoom(s *socketio.SocketV4, data ...interface{}) error {
	defer deferRecover(s)
	us, err := getSession(s, gs.Redis)
	if err != nil {
		return err
	}
	msg, ok := data[0].(string)
	if !ok {
		return errors.New("bad format leave room message")
	}
	slog.Info("Player Chat Message", slog.String("playerId", us.PlayerID), slog.String("chat", msg))
	// unmarshal chat struct from json
	cm, err := util.FromJSONString[game.ChatMessage](msg)
	if err != nil {
		return err
	}
	// get game state from redis
	g := getGameState(cm.Gid, gs.Redis)
	// add chat message
	g.ChatLog = append(g.ChatLog, *cm)
	// store game state
	storeGameState(cm.Gid, g, gs.Redis)
	// emit game state to room
	broadcastEventToGameRoom(cm.Gid, GameUpdate, g, s)
	return nil
}

func (gs *GameServer) OnStartGame(s *socketio.SocketV4, data ...interface{}) error {
	defer deferRecover(s)
	us, err := getSession(s, gs.Redis)
	if err != nil {
		return err
	}
	msg, ok := data[0].(string)
	if !ok {
		return errors.New("bad format leave room message")
	}
	// check if user is empowered to start games
	l := getLobbyState(msg, gs.Redis)

	if us.PlayerID != l.Host.PlayerID {
		s.Emit(string(Warning), serialize.String("You're not the Host"))
		return errors.New("only the host can start the game")
	}

	if len(l.Players) < 2 {
		s.Emit(string(Warning), serialize.String("Not enough Players"))
		return errors.New("2 or more players required to play")
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

	broadcastEventToGameRoom(msg, GameStart, g, s)

	// send prompt to player
	p, _ := util.ToJSONString(&prompt)
	s.To(g.CurrentTurn.PlayerID).Emit("ActionPrompt", serialize.String(p))
	return nil
}

func (gs *GameServer) OnTakeTurnAction(s *socketio.SocketV4, data ...interface{}) error {
	defer deferRecover(s)
	us, err := getSession(s, gs.Redis)
	if err != nil {
		return err
	}
	msg, ok := data[0].(string)
	if !ok {
		return errors.New("bad format leave room message")
	}
	slog.Info("Player is taking action", slog.String("playerId", us.PlayerID), slog.Any("action", msg))
	// convert msg to PromptResponse
	pr, err := util.FromJSONString[game.PromptResponse](msg)
	if err != nil {
		handleError(err, s)
		return err
	}
	// get game state from cache
	g := getGameState(pr.Gid, gs.Redis)

	if us.PlayerID != g.CurrentTurn.PlayerID {
		s.Emit(string(Warning), serialize.String("It's not your turn"))
		return errors.New("it is not your turn")
	}

	// do game thing
	prompt := game.GameFlow(g, *pr)

	// store game state
	storeGameState(pr.Gid, g, gs.Redis)

	p, _ := util.ToJSONString(&prompt)
	if g.CurrentTurn.PlayerID == us.PlayerID {
		// send prompt to current player
		s.In(us.PlayerID).Emit("ActionPrompt", serialize.String(p))
	} else {
		// else, the prompt is for the next player, then do something like this:
		s.In(g.CurrentTurn.PlayerID).Emit("ActionPrompt", serialize.String(p))
	}

	// broadcast gamestate to room
	broadcastEventToGameRoom(pr.Gid, GameUpdate, g, s)
	return nil
}
