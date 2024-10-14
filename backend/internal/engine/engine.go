package engine

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"pandagame/internal/config"
	"pandagame/internal/framework"
	"pandagame/internal/game"
	"pandagame/internal/pocketbase"
	"strings"

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
	Lobby Lobby           `json:"lobby"`
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

func MessageSerializer(messageType string, payload any) (string, error) {
	shell := ServerEventShell{
		MessageType: messageType,
		Message:     payload,
	}
	// TODO: need some sort of way of checking if payload implements some CLientSafe interface, and format the message for just the specific recipient
	// ... would need to pass connection id here somehow (likely by passing the http.Request as a param)
	bb := bytes.NewBuffer(make([]byte, 0))
	if err := json.NewEncoder(bb).Encode(shell); err != nil {
		return "", err
	}
	return bb.String(), nil
}

func getToken(req *http.Request) (string, error) {
	tokenCookie, _ := req.Cookie("PGToken")
	tokenHeader := req.Header.Get("Authorization")
	token := ""
	if tokenCookie != nil {
		token = tokenCookie.Value
	}
	if tokenHeader != "" {
		token = tokenHeader
	}
	if token == "" {
		return "", errors.New("no token")
	}
	return token, nil
}

func IDFromToken(req *http.Request) string {
	token, _ := getToken(req)
	middle := strings.Split(token, ".")[1]
	raw, err := base64.URLEncoding.DecodeString(middle)
	if err != nil {
		return ""
	}
	claims := make(map[string]any)
	if err := json.NewDecoder(bytes.NewReader(raw)).Decode(&claims); err != nil {
		return ""
	}
	return claims["id"].(string)
}

func ConnectionAuthValidator(w http.ResponseWriter, r *http.Request) error {
	token, err := getToken(r)
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
	w.Header().Add("Set-Cookie", fmt.Sprintf("PGToken:%s", resp.Token))
	return nil
}

type Lobby struct {
	Host       string
	Players    []string
	Spectators []string
	Started    bool
	GameId     string
}

type PandaGameEngine struct {
	PB pocketbase.PBClient
}

func (p *PandaGameEngine) HandleEvent(event framework.Event) ([]framework.Event, error) {
	switch ClientEventType(event.Type) {
	case CreateGame:
		gameId := uuid.NewString()
		l := Lobby{
			Host:       event.SourceId,
			Players:    []string{event.SourceId},
			Spectators: make([]string, 0),
			GameId:     gameId}
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
			Payload: g,
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
			Payload: gr.State,
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
