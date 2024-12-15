package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"pandagame/internal/game"
	"pandagame/internal/htmx/websocket"
)

func SerializeToHTML(messageType string, payload any) (string, error) {
	mt := ServerEventType(messageType)
	switch mt {
	case LobbyUpdate:
		l, ok := payload.(*game.Lobby)
		if !ok {
			return "", errors.New("bad lobby payload")
		}
		return serializeLobbyUpdate(*l)
	case GameStart, GameUpdate, GameOver:
		g, ok := payload.(*game.GameState)
		if !ok {
			return "", errors.New("bad game state payload")
		}
		return serializeGameState(*g)
	case ActionPrompt:
		p, ok := payload.(*game.Prompt)
		if !ok {
			return "", errors.New("bad prompt payload")
		}
		return serializeActionPrompt(*p)
	case Goodbye:
		return serializeGoodbye()
	case Warning:
		s, ok := payload.(string)
		if !ok {
			return "", errors.New("warning payload not a string")
		}
		return serializeWarning(s)
	default:
		return "", fmt.Errorf("cannot serialize message type %s, not a server event type", messageType)
	}
}

func serializeLobbyUpdate(l game.Lobby) (string, error) {
	bb := bytes.NewBuffer(make([]byte, 0))
	err := websocket.RenderLobby(l).Render(context.Background(), bb)
	return bb.String(), err
}

func serializeGameState(g game.GameState) (string, error) {
	bb := bytes.NewBuffer(make([]byte, 0))
	err := websocket.RenderGameState(g).Render(context.Background(), bb)
	return bb.String(), err
}

func serializeActionPrompt(p game.Prompt) (string, error) {
	bb := bytes.NewBuffer(make([]byte, 0))
	err := websocket.RenderPrompt(p).Render(context.Background(), bb)
	return bb.String(), err
}

func serializeWarning(message string) (string, error) {
	bb := bytes.NewBuffer(make([]byte, 0))
	err := websocket.RenderWarning(message).Render(context.Background(), bb)
	return bb.String(), err
}

func serializeGoodbye() (string, error) {
	bb := bytes.NewBuffer(make([]byte, 0))
	err := websocket.RenderGoodbye().Render(context.Background(), bb)
	return bb.String(), err
}
