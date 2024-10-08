package engine

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
	resp, err := pocketbase.NewPocketBase("todo", nil).WithToken(token).AsUser().Auth("players").RefreshAuth(pocketbase.RecordQuery{})
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
		l := &Lobby{
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
		// TODO: store lobby (in redis or pocket base?)
		return []framework.Event{response}, nil
	case Matchmake:
		//
	case CancelMatchmake:
		//
	case JoinGame:
		gameId := event.Payload.(string)
		// TODO: get lobby from storage
		// TODO: add user to lobby
		response := framework.Event{
			Source:   framework.TargetServer,
			SourceId: event.SourceId,
			Dest:     framework.TargetJoinGroup,
			DestId:   gameId,
		}
		// TODO add lobby update broadcast to list
		return []framework.Event{response}, nil
	case LeaveGame:
	case StartGame:
		// TODO get lobby from storage
		// TODO apply settings from lobby
		// g := game.NewGame()
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
		// TODO get game state
		action := event.Payload.(game.PromptResponse)
		nextPrompt := game.GameFlow(nil, action)
		// TODO: broadcast all
		response := framework.Event{
			Source:  framework.TargetServer,
			Dest:    framework.TargetClient,
			DestId:  nextPrompt.Pid,
			Type:    string(ActionPrompt),
			Payload: nextPrompt,
		}
		return []framework.Event{response}, nil
	case ChangeSettings:
		// TODO will be json, part of lobby
	default:
		return make([]framework.Event, 0), fmt.Errorf("invalid message type: %s", event.Type)
	}
	return []framework.Event{}, nil
}
