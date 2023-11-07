package events

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"pandagame/internal/auth"
	"pandagame/internal/game"
	"pandagame/internal/mongoconn"
	"pandagame/internal/pandaplex"
	"pandagame/internal/redisconn"
	"pandagame/internal/util"
	"slices"
	"strings"

	"github.com/redis/go-redis/v9"
)

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
	JoinGame        ClientEventType = "JoinGame"
	LeaveGame       ClientEventType = "LeaveGame"
	GameChat        ClientEventType = "GameChat"
	TakeAction      ClientEventType = "TakeAction"
	Reprompt        ClientEventType = "RePrompt"
	CreateGame      ClientEventType = "CreateGame"
	StartGame       ClientEventType = "StartGame"
	ChangeSettings  ClientEventType = "ChangeSettings"
	Matchmake       ClientEventType = "Matchmake"
	CancelMatchmake ClientEventType = "CancelMatchmake"
)

type ClientEvent struct {
	Type    ClientEventType `json:"messageType"`
	Payload json.RawMessage `json:"message"`
}

type ServerEvent struct {
	Type    ServerEventType `json:"messageType"`
	Payload interface{}     `json:"message"`
}

const (
	gamePfx  string = "g-" // game prefix for redis keys
	lobbyPfx string = "l-" // lobby prefix for redis keys
)

type GameServer struct {
	Redis redisconn.RedisConn
	Mongo mongoconn.CollectionConn
}

func (g *GameServer) HandleMessage(p pandaplex.PlexerInternal, message string) {
	slog.Info("game server got a message", slog.String("message", message))

	event := new(ClientEvent)
	json.NewDecoder(strings.NewReader(message)).Decode(event)

	switch event.Type {
	case JoinGame:
		//
		g.OnJoinGame(p, string(event.Payload))
	case LeaveGame:
		//
		g.OnLeaveGame(p, string(event.Payload))
	case GameChat:
		//
		cm := new(game.ChatMessage)
		json.NewDecoder(bytes.NewReader(event.Payload)).Decode(cm)
		g.OnChatInRoom(p, *cm)
	case TakeAction:
		//
		pr := new(game.PromptResponse)
		json.NewDecoder(bytes.NewReader(event.Payload)).Decode(pr)
		g.OnTakeTurnAction(p, *pr)
	case Reprompt:
		//
	case CreateGame:
		//
		g.OnCreateGameLobby(p)
	case StartGame:
		//
		g.OnStartGame(p, string(event.Payload))
	case ChangeSettings:
		//
	case Matchmake:
		//
	case CancelMatchmake:
		//
	default:
		resp := &ServerEvent{
			Type: Warning,
			Payload: map[string]error{
				"error": errors.New("unsupported event type"),
			},
		}
		warning, _ := util.ToJSONString(resp)
		p.Reply(warning)
	}
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

// func (gs *GameServer) OnSearchForGame(s *socketio.SocketV4, data ...interface{}) error {
// 	defer deferRecover(s)
// 	// matchmaking is a way future feature
// 	return nil
// }

// func (gs *GameServer) OnCancelSearchForGame(s *socketio.SocketV4, data ...interface{}) error {
// 	defer deferRecover(s)
// 	// matchmaking is a way future feature
// 	return nil
// }

func (gs *GameServer) OnCreateGameLobby(p pandaplex.PlexerInternal) error {
	defer deferRecover(p)
	us, err := getSession(p, gs.Redis)
	if err != nil {
		return err
	}
	// check if user is allowed to create lobbies
	if !us.Empowered {
		se := &ServerEvent{
			Type:    Warning,
			Payload: "You are not allowed to create lobbies",
		}
		resp, _ := util.ToJSONString(se)
		p.Reply(resp)
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
	p.JoinRoom(gid)

	l := &Lobby{
		Host:       us,
		Players:    []*auth.UserSession{us},
		Spectators: make([]*auth.UserSession, 0),
	}

	storeLobbyState(gid, l, gs.Redis)

	// emit lobby update to the room
	broadcastEventToGameRoom(gid, LobbyUpdate, l.RemoveIDs(), p)
	return nil
}

func (gs *GameServer) OnJoinGame(p pandaplex.PlexerInternal, roomId string) error {
	defer deferRecover(p)
	us, err := getSession(p, gs.Redis)
	if err != nil {
		return err
	}

	slog.Info("Player joining game room", slog.String("playerId", us.PlayerID), slog.String("room", roomId))
	p.JoinRoom(roomId)
	l := getLobbyState(roomId, gs.Redis)
	if len(l.Players) < 4 && !l.Started {
		l.Players = append(l.Players, us)
	} else {
		l.Spectators = append(l.Spectators, us)
	}

	storeLobbyState(roomId, l, gs.Redis)

	broadcastEventToGameRoom(roomId, LobbyUpdate, l.RemoveIDs(), p)
	return nil
}

func (gs *GameServer) OnLeaveGame(p pandaplex.PlexerInternal, roomId string) error {
	defer deferRecover(p)
	us, err := getSession(p, gs.Redis)
	if err != nil {
		return err
	}

	slog.Info("Player leaving game room", slog.String("playerId", us.PlayerID), slog.String("room", roomId))
	p.LeaveRoom(roomId)

	l := getLobbyState(roomId, gs.Redis)
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
	storeLobbyState(roomId, l, gs.Redis)
	broadcastEventToGameRoom(roomId, LobbyUpdate, l.RemoveIDs(), p)
	return nil
}

func (gs *GameServer) OnChatInRoom(p pandaplex.PlexerInternal, cm game.ChatMessage) error {
	defer deferRecover(p)
	us, err := getSession(p, gs.Redis)
	if err != nil {
		return err
	}

	slog.Info("Player Chat Message", slog.String("playerId", us.PlayerID), slog.String("chat", cm.Message))
	// get game state from redis
	cm.From = us.Name
	g := getGameState(cm.Gid, gs.Redis)
	// add chat message
	g.ChatLog = append(g.ChatLog, cm)
	// store game state
	storeGameState(cm.Gid, g, gs.Redis)
	// emit game state to room
	broadcastEventToGameRoom(cm.Gid, GameUpdate, g, p)
	return nil
}

func (gs *GameServer) OnStartGame(p pandaplex.PlexerInternal, roomId string) error {
	defer deferRecover(p)
	us, err := getSession(p, gs.Redis)
	if err != nil {
		return err
	}

	// check if user is empowered to start games
	l := getLobbyState(roomId, gs.Redis)

	if us.PlayerID != l.Host.PlayerID {
		se := &ServerEvent{
			Type:    Warning,
			Payload: "You're not the Host",
		}
		resp, _ := util.ToJSONString(se)
		p.Reply(resp)
		return errors.New("only the host can start the game")
	}

	if len(l.Players) < 2 {
		se := &ServerEvent{
			Type:    Warning,
			Payload: "Not enough Players",
		}
		resp, _ := util.ToJSONString(se)
		p.Reply(resp)
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

	storeGameState(roomId, g, gs.Redis)

	l.Started = true

	storeLobbyState(roomId, l, gs.Redis)

	broadcastEventToGameRoom(roomId, GameStart, g, p)

	// send prompt to player
	se := &ServerEvent{
		Type:    ActionPrompt,
		Payload: prompt,
	}
	pe, _ := util.ToJSONString(se)
	p.SendTo(pe, g.CurrentTurn.PlayerID)
	return nil
}

func (gs *GameServer) OnTakeTurnAction(p pandaplex.PlexerInternal, pr game.PromptResponse) error {
	defer deferRecover(p)
	us, err := getSession(p, gs.Redis)
	if err != nil {
		return err
	}

	slog.Info("Player is taking action", slog.String("playerId", us.PlayerID), slog.Any("action", pr.Action))
	// get game state from cache
	g := getGameState(pr.Gid, gs.Redis)

	if us.PlayerID != g.CurrentTurn.PlayerID {
		se := &ServerEvent{
			Type:    Warning,
			Payload: "It's not your turn",
		}
		resp, _ := util.ToJSONString(se)
		p.Reply(resp)
		return errors.New("it is not your turn")
	}

	// do game thing
	prompt := game.GameFlow(g, pr)

	// store game state
	storeGameState(pr.Gid, g, gs.Redis)

	se := &ServerEvent{
		Type:    ActionPrompt,
		Payload: prompt,
	}
	pe, _ := util.ToJSONString(se)
	if g.CurrentTurn.PlayerID == us.PlayerID {
		// send prompt to current player
		p.Reply(pe)
	} else {
		// else, the prompt is for the next player, then do something like this:
		p.SendTo(pe, g.CurrentTurn.PlayerID)
	}

	// broadcast gamestate to room
	broadcastEventToGameRoom(pr.Gid, GameUpdate, g, p)
	return nil
}
