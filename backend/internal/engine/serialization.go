package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"pandagame/internal/game"
	"pandagame/internal/web"
	"strings"

	"github.com/go-chi/chi"
	"github.com/mitchellh/mapstructure"
)

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
	if p, ok := payload.(map[string]any); ok {
		payload = handleMapPayload(messageType, p)
	}
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

// convert map to type
// to be clear, I'd rather not use mapstructure, it's kinda a bandaid for bad deserialization upstream
func handleMapPayload(messageType string, payload map[string]any) any {
	mt := ServerEventType(messageType)
	dcfg := &mapstructure.DecoderConfig{TagName: "json", IgnoreUntaggedFields: true}
	switch mt {
	case LobbyUpdate:
		l := new(game.Lobby)
		dcfg.Result = l
		d, _ := mapstructure.NewDecoder(dcfg)
		if err := d.Decode(payload); err != nil {
			return nil
		}
		return l

	case GameStart, GameUpdate, GameOver:
		g := new(game.GameState)
		dcfg.Result = g
		d, _ := mapstructure.NewDecoder(dcfg)
		if err := d.Decode(payload); err != nil {
			return nil
		}
		return g

	case ActionPrompt:
		p := new(game.Prompt)
		dcfg.Result = p
		d, _ := mapstructure.NewDecoder(dcfg)
		if err := d.Decode(payload); err != nil {
			return nil
		}
		return p

	default:
		return payload
	}
}
