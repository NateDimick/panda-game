package game

import (
	"bytes"
	"encoding/json"
	"fmt"
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
			{GardenerObjective{
				Color:       PinkBambooPlot,
				Improvement: AnyImprovement,
				Height:      4,
				Count:       4,
				Value:       8,
				OT:          GardenerObjectiveType,
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

func TestImprovementReserve(t *testing.T) {
	var ir ImprovementReserve = map[ImprovementType]int{
		FertilizerImprovement: 1,
		WatershedImprovement:  1,
		EnclosureImprovement:  1,
	}
	assert.False(t, ir.IsEmpty())
	assert.Equal(t, 3, len(ir.AvailableImprovements()))
	ir[FertilizerImprovement] = 0
	assert.Equal(t, 2, len(ir.AvailableImprovements()))
	ir[WatershedImprovement] = 0
	ir[EnclosureImprovement] = 0
	assert.True(t, ir.IsEmpty())
	assert.Empty(t, ir.AvailableImprovements())
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

func TestPandaObjectiveComplete(t *testing.T) {
	b := NewBoard()
	o := PandaObjective{
		GreenCount:  1,
		YellowCount: 2,
	}
	p := Player{
		Bamboo: map[PlotType]int{
			GreenBambooPlot:  3,
			YellowBambooPlot: 2,
			PinkBambooPlot:   1,
		},
	}

	assert.True(t, o.IsComplete(p, *b))
}

func TestPandaObjectiveNotComplete(t *testing.T) {
	b := NewBoard()
	o := PandaObjective{
		PinkCount:   3,
		YellowCount: 3,
	}
	p := Player{
		Bamboo: map[PlotType]int{
			GreenBambooPlot:  3,
			YellowBambooPlot: 2,
			PinkBambooPlot:   1,
		},
	}

	assert.False(t, o.IsComplete(p, *b))
}

func TestGardenerObjectiveComplete(t *testing.T) {
	o := GardenerObjective{
		Color:       GreenBambooPlot,
		Height:      3,
		Count:       4,
		Improvement: AnyImprovement,
	}
	b := NewBoard()
	for i := 1; i < 5; i++ {
		pid := fmt.Sprintf("p%d", i)
		b.AddPlot(pid, GreenBambooPlot, NoImprovement)
		b.PlotGrowBamboo(pid)
		b.PlotGrowBamboo(pid)
		b.PlotGrowBamboo(pid)
	}
	p := Player{}

	assert.True(t, o.IsComplete(p, *b))
}

func TestGardenerObjectiveNotComplete(t *testing.T) {
	o := GardenerObjective{
		Color:       GreenBambooPlot,
		Height:      3,
		Count:       4,
		Improvement: AnyImprovement,
	}
	b := NewBoard()
	for i := 1; i < 5; i++ {
		pid := fmt.Sprintf("p%d", i)
		b.AddPlot(pid, GreenBambooPlot, NoImprovement)
		for j := 0; j < i; j++ {
			b.PlotGrowBamboo(pid)
		}
	}
	p := Player{}

	assert.False(t, o.IsComplete(p, *b))
}
