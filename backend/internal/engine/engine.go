package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"pandagame/internal/config"
	"pandagame/internal/framework"
	"pandagame/internal/game"
	"pandagame/internal/pocketbase"
	"pandagame/internal/web"
	"strings"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type ClientEventShell struct {
	MessageType string          `json:"messageType"`
	Message     json.RawMessage `json:"message"`
}

type ServerEventShell struct {
	MessageType string `json:"messageType"`
	Message     any    `json:"message"`
}

type GameRecord struct {
	GID   string          `json:"gameId"`
	State *game.GameState `json:"state"`
	Lobby game.Lobby      `json:"lobby"`
}

func MessageDeserializer(raw string, req *http.Request) (string, any, error) {
	msg := new(ClientEventShell)
	if err := json.NewDecoder(strings.NewReader(raw)).Decode(msg); err != nil {
		return "", nil, err
	}
	var payload any
	decodeJson := false
	switch ClientEventType(msg.MessageType) {
	case CreateGame, Matchmake, CancelMatchmake:
		payload = ""
	case JoinGame, LeaveGame, StartGame, Reprompt:
		payload = string(msg.Message) // this is always the game id
	case GameChat:
		payload = new(game.ChatMessage)
		decodeJson = true
	case TakeAction:
		payload = new(game.PromptResponse)
		decodeJson = true
	case ChangeSettings:
		// TODO will be json
	default:
		return "", nil, fmt.Errorf("invalid message type: %s", msg.MessageType)
	}
	if decodeJson {
		if err := json.NewDecoder(bytes.NewReader(msg.Message)).Decode(payload); err != nil {
			return "", nil, err
		}
	}
	return msg.MessageType, payload, nil
}

func MessageSerializer(messageType string, payload any, req *http.Request) (string, error) {
	respType := chi.URLParam(req, "type")
	switch respType {
	case "json":
		shell := ServerEventShell{
			MessageType: messageType,
			Message:     payload,
		}
		if cs, ok := payload.(game.ClientSafe); ok {
			id := web.IDFromRequest(req)
			safePayload := cs.ClientSafe(id)
			shell.Message = safePayload
		}
		bb := bytes.NewBuffer(make([]byte, 0))
		if err := json.NewEncoder(bb).Encode(shell); err != nil {
			return "", err
		}
		return bb.String(), nil
	case "htmx":
		fallthrough
	default:
		return SerializeToHTML(messageType, payload)
	}
}

func ConnectionAuthValidator(w http.ResponseWriter, r *http.Request) error {
	token, err := web.GetToken(r)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusUnauthorized)
		return err
	}
	cfg := config.LoadAppConfig()
	resp, err := pocketbase.NewPocketBase(cfg.PB.Address, nil).WithToken(token).AsUser().Auth("players").RefreshAuth(nil)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusForbidden)
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:  web.PandaGameCookie,
		Value: resp.Token,
	})
	return nil
}

type PandaGameEngine struct {
	PB pocketbase.PBClient
}

func (p *PandaGameEngine) HandleEvent(event framework.Event) ([]framework.Event, error) {
	switch ClientEventType(event.Type) {
	case CreateGame:
		gameId := uuid.NewString()
		l := game.Lobby{
			Host:       event.SourceId,
			Players:    []string{event.SourceId},
			Spectators: make([]string, 0),
			GameId:     gameId,
		}
		response := framework.Event{
			Source:  framework.TargetServer,
			Dest:    framework.TargetClient,
			DestId:  event.SourceId,
			Type:    string(LobbyUpdate),
			Payload: l,
		}
		record := GameRecord{
			GID:   gameId,
			Lobby: l,
			State: nil,
		}
		if err := StoreGame(&record, false, p.PB); err != nil {
			return make([]framework.Event, 0), err
		}
		return []framework.Event{response}, nil
	case Matchmake:
		//
	case CancelMatchmake:
		//
	case JoinGame:
		gameId := event.Payload.(string)
		gr, err := GetGame(gameId, p.PB)
		if err != nil {
			return make([]framework.Event, 0), err
		}
		if len(gr.Lobby.Players) < 4 {
			gr.Lobby.Players = append(gr.Lobby.Players, event.SourceId)
		} else {
			gr.Lobby.Spectators = append(gr.Lobby.Spectators, event.SourceId)
		}
		response := framework.Event{
			Source:   framework.TargetServer,
			SourceId: event.SourceId,
			Dest:     framework.TargetJoinGroup,
			DestId:   gameId,
		}
		broadcast := framework.Event{
			Source:  framework.TargetServer,
			Dest:    framework.TargetGroup,
			DestId:  gameId,
			Payload: gr.Lobby,
			Type:    string(LobbyUpdate),
		}
		if err := StoreGame(gr, true, p.PB); err != nil {
			return make([]framework.Event, 0), err
		}
		return []framework.Event{response, broadcast}, nil
	case LeaveGame:
		// TODO
	case StartGame:
		gameId := event.Payload.(string)
		gr, err := GetGame(gameId, p.PB)
		if err != nil {
			return []framework.Event{}, err
		}
		players := make([]game.Player, len(gr.Lobby.Players))
		for i, p := range gr.Lobby.Players {
			players[i] = game.Player{
				ID:    p,
				Name:  "TODO", // get user name from DB base on ID?
				Order: i + 1,
			}
		}
		// TODO apply settings from lobby
		g := game.StartGame(players)
		gr.State = g
		broadcast := framework.Event{
			Source:  framework.TargetServer,
			Dest:    framework.TargetGroup,
			DestId:  gameId,
			Payload: *g,
			Type:    string(GameStart),
		}
		firstPrompt := game.GameFlow(g, game.PromptResponse{Action: game.NextPlayerTurn})
		prompt := framework.Event{
			Source:  framework.TargetServer,
			Dest:    framework.TargetClient,
			DestId:  firstPrompt.Pid,
			Type:    string(ActionPrompt),
			Payload: firstPrompt,
		}
		if err := StoreGame(gr, true, p.PB); err != nil {
			return make([]framework.Event, 0), err
		}
		return []framework.Event{broadcast, prompt}, nil

	case Reprompt:
		//
	case GameChat:
		// TODO get game state
		g := game.GameState{} // placeholder
		msg := event.Payload.(game.ChatMessage)
		g.ChatLog = append(g.ChatLog, msg)
		// TODO store game state
		// TODO broadcast game state update
	case TakeAction:
		action := event.Payload.(game.PromptResponse)
		gr, err := GetGame(action.Gid, p.PB)
		if err != nil {
			return []framework.Event{}, err
		}

		nextPrompt := game.GameFlow(gr.State, action)
		response := framework.Event{
			Source:  framework.TargetServer,
			Dest:    framework.TargetClient,
			DestId:  nextPrompt.Pid,
			Type:    string(ActionPrompt),
			Payload: nextPrompt,
		}
		broadcast := framework.Event{
			Source:  framework.TargetServer,
			Dest:    framework.TargetGroup,
			DestId:  gr.GID,
			Payload: *gr.State,
			Type:    string(GameStart),
		}
		return []framework.Event{broadcast, response}, nil
	case ChangeSettings:
		// TODO will be json, part of lobby
	default:
		return make([]framework.Event, 0), fmt.Errorf("invalid message type: %s", event.Type)
	}
	return []framework.Event{}, nil
}

func GetGame(gameId string, pb pocketbase.PBClient) (*GameRecord, error) {
	c := pb.AsAdmin().Records("games")
	gr := new(GameRecord)
	if _, err := c.View(gameId, gr, nil); err != nil {
		return nil, err
	}
	return gr, nil
}

func StoreGame(gr *GameRecord, update bool, pb pocketbase.PBClient) error {
	c := pb.AsAdmin().Records("games")
	var err error
	if update {
		_, err = c.Update(gr.GID, gr, nil, nil)
	} else {
		_, err = c.Create(pocketbase.NewRecord{ID: gr.GID, CustomFields: gr}, nil, nil)
	}
	return err
}
