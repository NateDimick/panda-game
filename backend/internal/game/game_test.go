package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGame(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Fail()
		}
	}()
	g := NewGame()
	assert.NotNil(t, g)
	assert.Equal(t, 15, len(g.ObjectiveDecks[PlotObjectiveType]))
	assert.Equal(t, 15, len(g.ObjectiveDecks[GardenerObjectiveType]))
	assert.Equal(t, 15, len(g.ObjectiveDecks[PandaObjectiveType]))
}

func TestDrawPlots(t *testing.T) {
	g := new(GameState)
	g.PlotDeck = []DeckPlot{
		{Type: GreenBambooPlot},
		{Type: GreenBambooPlot},
		{Type: GreenBambooPlot},
		{Type: GreenBambooPlot},
		{Type: GreenBambooPlot},
	}
	plotOptions := g.DrawPlots()
	assert.Equal(t, 3, len(plotOptions))
	assert.Equal(t, 2, len(g.PlotDeck))

	plotOptions2 := g.DrawPlots()
	assert.Equal(t, 2, len(plotOptions2))
	assert.Equal(t, 0, len(g.PlotDeck))
}
