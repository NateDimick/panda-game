package engine

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"pandagame/internal/config"
	"pandagame/internal/framework"
	"pandagame/internal/game"
	"pandagame/internal/web"

	"github.com/google/uuid"
	"github.com/surrealdb/surrealdb.go"
	"github.com/surrealdb/surrealdb.go/pkg/models"
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
	RID   *models.RecordID `json:"id"`
	GID   string           `json:"gameId"`
	State *game.GameState  `json:"state"`
	Lobby game.Lobby       `json:"lobby"`
}

func ConnectionAuthValidator(w http.ResponseWriter, r *http.Request) error {
	token, err := web.GetToken(r)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusUnauthorized)
		return err
	}
	db, _ := config.Surreal()
	if err := db.Authenticate(token); err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusForbidden)
		return err
	}
	userInfo, _ := db.Info()
	slog.Info("db user info", slog.Any("info", userInfo))
	// http.SetCookie(w, &http.Cookie{
	// 	Name:  web.PandaGameCookie,
	// 	Value: resp.Token,
	// })
	return nil
}

type PandaGameEngine struct{}

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
		join := framework.Event{
			Source:   framework.TargetServer,
			SourceId: event.SourceId,
			Dest:     framework.TargetJoinGroup,
			DestId:   gameId,
		}
		record := GameRecord{
			RID:   recordID(gameId),
			GID:   gameId,
			Lobby: l,
			State: nil,
		}
		if err := StoreGame(&record, false); err != nil {
			return make([]framework.Event, 0), err
		}
		return []framework.Event{join, response}, nil
	case Matchmake:
		//
	case CancelMatchmake:
		//
	case JoinGame:
		gameId := event.Payload.(string)
		gr, err := GetGame(gameId)
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
		if err := StoreGame(gr, true); err != nil {
			return make([]framework.Event, 0), err
		}
		return []framework.Event{response, broadcast}, nil
	case LeaveGame:
		// TODO
	case StartGame:
		gameId := event.Payload.(string)
		gr, err := GetGame(gameId)
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
		if err := StoreGame(gr, true); err != nil {
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
		gr, err := GetGame(action.Gid)
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

func recordID(gameId string) *models.RecordID {
	return &models.RecordID{
		ID:    gameId,
		Table: "game",
	}
}

func GetGame(gameId string) (*GameRecord, error) {
	db, _ := config.AdminSurreal()
	id := recordID(gameId)
	record, err := surrealdb.Select[GameRecord](db, *id)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func StoreGame(gr *GameRecord, update bool) error {
	db, _ := config.AdminSurreal()
	_, err := surrealdb.Upsert[[]GameRecord](db, models.Table("game"), gr)
	return err
}

func DeleteGame(gr *GameRecord) error {
	db, _ := config.AdminSurreal()
	_, err := surrealdb.Delete[GameRecord](db, *gr.RID)
	return err
}
