package engine

import (
	"bytes"
	"encoding/json"
	"pandagame/internal/game"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSerializationLoop(t *testing.T) {
	// struct -> json -> map -> struct
	cases := []struct {
		MsgType string
		payload any
	}{
		{"LobbyUpdate", &game.Lobby{Host: "larry", Players: []string{"larry", "big bird"}}},
		{"GameUpdate", &game.GameState{Board: &game.Board{Plots: map[string]game.Plot{"a": {Type: game.FuturePlot}}}}},
		{"ActionPrompt", &game.Prompt{Action: game.ChooseGrowth, SelectType: game.PlotIDSelectType, SelectFrom: []any{"a", "b", "c"}}},
	}
	for _, tc := range cases {
		t.Run(tc.MsgType, func(tt *testing.T) {
			bb := bytes.NewBuffer(make([]byte, 0))
			json.NewEncoder(bb).Encode(tc.payload)
			m := make(map[string]any)
			json.NewDecoder(bb).Decode(&m)
			result := handleMapPayload(tc.MsgType, m)
			assert.Equal(tt, tc.payload, result)
		})
	}
}

func TestHandleMapPayload(t *testing.T) {
	m := map[string]any{"GameId": "15c425c2-4f8b-48c7-b5f8-f57a4b4af806", "Host": "player:nate", "Players": []string{"player:nate"}, "Spectators": []string{}, "Started": false}
	lu := handleMapPayload("LobbyUpdate", m)
	update := lu.(*game.Lobby)
	assert.Equal(t, update.GameId, m["GameId"])
}
