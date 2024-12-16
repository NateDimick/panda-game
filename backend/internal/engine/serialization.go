package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"pandagame/internal/framework"
	"pandagame/internal/game"
	"pandagame/internal/web"
	"reflect"
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

// ensure payloads contain specific structs for various event types
func StructMiddleware(e framework.Event, r *http.Request) (framework.Event, error) {
	mt := ServerEventType(e.Type)
	switch mt {
	case LobbyUpdate:
		e.Payload = structConverter[game.Lobby](e.Payload)
	case GameStart, GameUpdate, GameOver:
		e.Payload = structConverter[game.GameState](e.Payload)
	case ActionPrompt:
		e.Payload = structConverter[game.Prompt](e.Payload)
	default:

	}
	return e, nil
}

func structConverter[T any](payload any) T {
	t := reflect.TypeOf(payload)
	switch t.Kind() {
	case reflect.Pointer:
		return *payload.(*T)
	case reflect.Struct:
		return payload.(T)
	case reflect.Map:
		out := new(T)
		dcfg := &mapstructure.DecoderConfig{TagName: "json", IgnoreUntaggedFields: true, Result: out}
		d, _ := mapstructure.NewDecoder(dcfg)
		if err := d.Decode(payload); err != nil {
			slog.Warn("failed to convert map to struct")
		}
		return *out
	default:
		slog.Warn("unsupported type")
		panic("this shouldn't happen")
	}
}
