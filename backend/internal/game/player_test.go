package game

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshalPlayer(t *testing.T) {
	p := &Player{
		Name: "test dude",
		ID:   "me",
		Objectives: []Objective{
			{PlotObjective{
				AnchorColor: GreenBambooPlot,
				Neighbors:   [6]PlotType{YellowBambooPlot, GreenBambooPlot, YellowBambooPlot, AnyPlot, AnyPlot, AnyPlot},
				Value:       5,
				OT:          PlotObjectiveType,
			}},
			{GardnerObjective{
				Color:       PinkBambooPlot,
				Improvement: AnyImprovement,
				Height:      4,
				Count:       4,
				Value:       8,
				OT:          GardnerObjectiveType,
			}},
			{PandaObjective{
				GreenCount:  1,
				YellowCount: 1,
				PinkCount:   1,
				Value:       6,
				OT:          PandaObjectiveType,
			}},
		},
	}
	assert.Equal(t, 5, p.Objectives[0].Points())
	bb := new(bytes.Buffer)
	json.NewEncoder(bb).Encode(p)
	assert.NotEmpty(t, bb.String())
	p2 := new(Player)
	json.NewDecoder(bb).Decode(p2)
	assert.Equal(t, p, p2)
}

func TestWTFFFFFF(t *testing.T) {
	o := &PandaObjective{1, 2, 3, 7, PandaObjectiveType}
	bs, _ := json.Marshal(o)
	o2 := new(PandaObjective)
	json.Unmarshal(bs, o2)
	assert.Equal(t, o, o2)
}

func TestPlotObjectiveComplete(t *testing.T) {
	b := NewBoard()
	b.AddPlot("p1", GreenBambooPlot, NoImprovement)
	b.AddPlot("p2", GreenBambooPlot, NoImprovement)
	b.AddPlot("p7", GreenBambooPlot, WatershedImprovement) // use watershed to get irrigation without placing irrigation on edges to plot

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
