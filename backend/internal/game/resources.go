package game

import (
	_ "embed"
)

//go:embed resources/plotdeck.json
var InitialPlotDeck []byte

//go:embed resources/objectivedeck.json
var InitialObjectiveDeck []byte

//go:embed resources/improvements.json
var InitialImprovements []byte
