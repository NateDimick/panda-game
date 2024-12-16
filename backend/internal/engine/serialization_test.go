package engine

import (
	"bytes"
	"encoding/json"
	"pandagame/internal/framework"
	"pandagame/internal/game"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSerializationLoop(t *testing.T) {
	// struct -> json -> map -> struct
	cases := []struct {
		MsgType string
		Payload any
	}{
		{"LobbyUpdate", game.Lobby{Host: "larry", Players: []string{"larry", "big bird"}}},
		{"GameUpdate", game.GameState{Board: &game.Board{Plots: map[string]game.Plot{"a": {Type: game.FuturePlot}}}}},
		{"ActionPrompt", game.Prompt{Action: game.ChooseGrowth, SelectType: game.PlotIDSelectType, SelectFrom: []any{"a", "b", "c"}}},
	}
	for _, tc := range cases {
		t.Run(tc.MsgType, func(tt *testing.T) {
			bb := bytes.NewBuffer(make([]byte, 0))
			json.NewEncoder(bb).Encode(tc.Payload)
			m := make(map[string]any)
			json.NewDecoder(bb).Decode(&m)
			event := framework.Event{
				Type:    tc.MsgType,
				Payload: m,
			}
			result, _ := StructMiddleware(event, nil)
			assert.Equal(tt, tc.Payload, result.Payload)
		})
	}
}

func TestStructConverter(t *testing.T) {
	l := game.Lobby{Host: "beavis", Players: []string{"beavis", "butthead"}}
	l2 := structConverter[game.Lobby](l)
	assert.Equal(t, l, l2)
	l3 := structConverter[game.Lobby](&l)
	assert.Equal(t, l, l3)
}
