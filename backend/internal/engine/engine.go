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

type PandaGameEngine struct {
}

func (p *PandaGameEngine) HandleEvent(framework.Event) (framework.Event, error) {
	return framework.Event{}, nil
}
