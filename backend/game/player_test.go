package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlotObjectiveComplete(t *testing.T) {
	b := NewBoard()
	b.AddPlot("p1", GreenBambooPlot, NoImprovement)
	b.AddPlot("p2", GreenBambooPlot, NoImprovement)
	b.AddPlot("p7", GreenBambooPlot, WatershedImprovement)

	o := PlotObjective{
		AnchorColor: GreenBambooPlot,
		Neighbors:   [6]PlotType{GreenBambooPlot, GreenBambooPlot, AnyPlot, AnyPlot, AnyPlot, AnyPlot},
	}

	complete := o.IsComplete(Player{}, *b)
	assert.True(t, complete)
}

func TestPlotObjectiveNotComplete(t *testing.T) {
	b := NewBoard()
	b.AddPlot("p1", GreenBambooPlot, NoImprovement)
	b.AddPlot("p2", GreenBambooPlot, NoImprovement)
	b.AddPlot("p7", GreenBambooPlot, NoImprovement) // no irrigation == not complete

	o := PlotObjective{
		AnchorColor: GreenBambooPlot,
		Neighbors:   [6]PlotType{GreenBambooPlot, GreenBambooPlot, AnyPlot, AnyPlot, AnyPlot, AnyPlot},
	}

	complete := o.IsComplete(Player{}, *b)
	assert.False(t, complete)
}
