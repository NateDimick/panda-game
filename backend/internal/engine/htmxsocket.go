package engine

import (
	"errors"
	"fmt"
	"pandagame/internal/game"
)

func SerializeToHTML(messageType string, payload any) (string, error) {
	mt := ServerEventType(messageType)
	switch mt {
	case LobbyUpdate:
		l, ok := payload.(Lobby)
		if !ok {
			return "", errors.New("bad lobby payload")
		}
		return serializeLobbyUpdate(l)
	case GameStart, GameUpdate, GameOver:
		g, ok := payload.(game.GameState)
		if !ok {
			return "", errors.New("bad game state payload")
		}
		return serializeGameState(g)
	case ActionPrompt:
		p, ok := payload.(game.Prompt)
		if !ok {
			return "", errors.New("bad prompt payload")
		}
		return serializeActionPrompt(p)
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

func serializeLobbyUpdate(l Lobby) (string, error) {
	return "todo", nil
}

func serializeGameState(g game.GameState) (string, error) {
	return "todo", nil
}

func serializeActionPrompt(p game.Prompt) (string, error) {
	return "todo", nil
}

func serializeWarning(message string) (string, error) {
	// target some normally empty warning div with the content text, maybe a popup?
	return "todo", nil
}

func serializeGoodbye() (string, error) {
	// target full ws div and display DISCONNECTED
	return "todo", nil
}
