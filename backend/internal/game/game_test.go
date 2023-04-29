package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
