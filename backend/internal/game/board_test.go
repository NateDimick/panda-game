package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBoard(t *testing.T) {
	b := NewBoard()
	assert.Equal(t, 7, len(b.Plots))
	assert.Equal(t, 12, len(b.Edges))
	for i := 0; i < 7; i++ {
		pid := fmt.Sprintf("p%d", i)
		plot := b.Plots[pid]
		assert.NotEmpty(t, plot)
	}
	for i := 0; i < 12; i++ {
		eid := fmt.Sprintf("e%d", i)
		edge := b.Edges[eid]
		assert.NotEmpty(t, edge)
	}
}

func TestMarshalBoard(t *testing.T) {
	b := NewBoard()
	bb := new(bytes.Buffer)
	json.NewEncoder(bb).Encode(b)
	assert.NotEmpty(t, bb.String())
	b2 := UnmarshalBoard(bb)
	assert.Equal(t, b, b2)
}

func TestAddPlot(t *testing.T) {
	b := NewBoard()
	b.AddPlot("p1", GreenBambooPlot, NoImprovement)
	// adding the first plot did not add any plots or edges
	assert.Equal(t, 7, len(b.Plots))
	assert.Equal(t, 12, len(b.Edges))
	b.AddPlot("p2", YellowBambooPlot, NoImprovement)
	// the second plot adds a future plot + edges for future plot
	assert.Equal(t, 8, len(b.Plots))
	assert.Equal(t, 14, len(b.Edges))
	// add third plot in the future plot that was created
	b.AddPlot("p7", PinkBambooPlot, NoImprovement)
	assert.Equal(t, 10, len(b.Plots))
	assert.Equal(t, 18, len(b.Edges))
}

func TestAllLegalMovesFromPlot(t *testing.T) {
	b := NewBoard()
	// no moves from pond with no added plots
	moves := b.LegalMovesFromPlot("p0")
	assert.Equal(t, 0, len(moves))
	// adding p1 gives one plot to move to
	b.AddPlot("p1", GreenBambooPlot, NoImprovement)
	moves = b.LegalMovesFromPlot("p0")
	assert.Equal(t, 1, len(moves))
	// adding p2 gives a second option
	b.AddPlot("p2", GreenBambooPlot, NoImprovement)
	moves = b.LegalMovesFromPlot("p0")
	assert.Equal(t, 2, len(moves))
	// p7 is not in a legal move postion from p0 (nestled between p1 and p2)
	b.AddPlot("p7", GreenBambooPlot, NoImprovement)
	moves = b.LegalMovesFromPlot("p0")
	assert.Equal(t, 2, len(moves))
	// p8 will be a legal move from p0, via p1 or p2
	b.AddPlot("p8", GreenBambooPlot, NoImprovement)
	moves = b.LegalMovesFromPlot("p0")
	assert.Equal(t, 3, len(moves))
}

func TestAllPresentPlots(t *testing.T) {
	b := NewBoard()
	// only 1 preset plot at board inception
	plots := b.AllPresentPlots()
	assert.Equal(t, 1, len(plots))
	// adding p1 adds 1 present plot and no future plots
	b.AddPlot("p1", GreenBambooPlot, NoImprovement)
	plots = b.AllPresentPlots()
	assert.Equal(t, 2, len(plots))
	// adding p2 adds 1 present plot and 1 future plot
	b.AddPlot("p2", GreenBambooPlot, NoImprovement)
	plots = b.AllPresentPlots()
	assert.Equal(t, 3, len(plots))
}

func TestAllIrrigatedPlots(t *testing.T) {
	b := NewBoard()
	// pond does not count as irrigated, neither do future plots
	plots := b.AllIrrigatedPlots()
	assert.Equal(t, 0, len(plots))
	// p1 is irrigated, next to pond
	b.AddPlot("p1", GreenBambooPlot, NoImprovement)
	plots = b.AllIrrigatedPlots()
	assert.Equal(t, 1, len(plots))
	// p2 is also irrigated, next to pond
	b.AddPlot("p2", GreenBambooPlot, NoImprovement)
	plots = b.AllIrrigatedPlots()
	assert.Equal(t, 2, len(plots))
	// p7 is not irrigated
	b.AddPlot("p7", GreenBambooPlot, NoImprovement)
	plots = b.AllIrrigatedPlots()
	assert.Equal(t, 2, len(plots))
	// p8 is irrigated thanks to improvement
	b.AddPlot("p8", GreenBambooPlot, WatershedImprovement)
	plots = b.AllIrrigatedPlots()
	assert.Equal(t, 3, len(plots))
	// make p7 irrigated by edges
	b.EdgeAddIrrigation("e6")
	b.EdgeAddIrrigation("e12")
	plots = b.AllIrrigatedPlots()
	assert.Equal(t, 4, len(plots))
	// p9 gets watershed improvement added to it
	b.AddPlot("p9", GreenBambooPlot, NoImprovement)
	b.PlotAddImprovement("p9", WatershedImprovement)
	plots = b.AllIrrigatedPlots()
	assert.Equal(t, 5, len(plots))
}

func TestAllFuturePlots(t *testing.T) {
	b := NewBoard()
	plots := b.AllFuturePlots()
	// all plots except pond are future
	assert.Equal(t, 6, len(plots))
	// p1 converts 1 future to present plot without adding a new future plot
	b.AddPlot("p1", GreenBambooPlot, NoImprovement)
	plots = b.AllFuturePlots()
	assert.Equal(t, 5, len(plots))
	// adding p2 creates 1 future plot, no change in total futures
	b.AddPlot("p2", GreenBambooPlot, NoImprovement)
	plots = b.AllFuturePlots()
	assert.Equal(t, 5, len(plots))
	// p7 creates 2 futures, which increases futures by 1
	b.AddPlot("p7", GreenBambooPlot, NoImprovement)
	plots = b.AllFuturePlots()
	assert.Equal(t, 6, len(plots))
}

func TestAllImprovablePlots(t *testing.T) {
	b := NewBoard()
	plots := b.AllImprovablePlots()
	// pond and future plots are not improvable
	assert.Equal(t, 0, len(plots))
	// p1 has no bamboo or improvement, so it is improvable
	b.AddPlot("p1", GreenBambooPlot, NoImprovement)
	plots = b.AllImprovablePlots()
	assert.Equal(t, 1, len(plots))
	// adding bamboo to p1 makes it ineligible for improvement
	b.PlotGrowBamboo("p1")
	plots = b.AllImprovablePlots()
	assert.Equal(t, 0, len(plots))
	// p2 has an improvement, so it is also ineligible
	b.AddPlot("p2", YellowBambooPlot, FertilizerImprovement)
	plots = b.AllImprovablePlots()
	assert.Equal(t, 0, len(plots))
}

func TestIrrigatableEdges(t *testing.T) {
	b := NewBoard()
	// no edges at inception are irrigatable
	irrigatableEdges := b.AllIrrigatableEdges()
	assert.Equal(t, 0, len(irrigatableEdges))
	// adding two adjacent plots to pond should create 1 irrigatable edge
	b.AddPlot("p1", GreenBambooPlot, NoImprovement)
	b.AddPlot("p2", YellowBambooPlot, NoImprovement)
	irrigatableEdges = b.AllIrrigatableEdges()
	assert.Equal(t, 1, len(irrigatableEdges))
	// because e6 between p1 and p2 is no irrigated, p7's edges can't be irrigated
	b.AddPlot("p7", PinkBambooPlot, WatershedImprovement)
	irrigatableEdges = b.AllIrrigatableEdges()
	assert.Equal(t, 1, len(irrigatableEdges))
	// irrigating the edge between p1 and p2 should allow p7's edges to be irrigated
	b.EdgeAddIrrigation("e6")
	irrigatableEdges = b.AllIrrigatableEdges()
	assert.Equal(t, 2, len(irrigatableEdges))
}

func TestBambooMechanics(t *testing.T) {
	b := NewBoard()
	b.AddPlot("p1", GreenBambooPlot, NoImprovement)
	b.AddPlot("p2", GreenBambooPlot, EnclosureImprovement)
	b.AddPlot("p3", GreenBambooPlot, FertilizerImprovement)
	b.AddPlot("p7", GreenBambooPlot, NoImprovement)
	// manually grow some bamboo, see that it works
	b.PlotGrowBamboo("p3")
	assert.Equal(t, 2, b.Plots["p3"].Bamboo)
	b.PlotGrowBamboo("p2")
	b.PlotGrowBamboo("p1")
	assert.Equal(t, 1, b.Plots["p1"].Bamboo)
	assert.Equal(t, 1, b.Plots["p2"].Bamboo)
	// move the gardener and check that he works properly
	b.MoveGardener("p3")
	b.MoveGardener("p2")
	assert.Equal(t, 4, b.Plots["p3"].Bamboo)
	assert.Equal(t, 2, b.Plots["p2"].Bamboo)
	assert.Equal(t, "p2", b.GardenerLocation)
	// move the panda around and check that he works properly
	bb := b.MovePanda("p1")
	assert.Equal(t, GreenBambooPlot, bb)
	bb = b.MovePanda("p2")
	assert.Equal(t, AnyPlot, bb) // panda cannot eat from enclosed plot
	bb = b.MovePanda("p7")
	assert.Equal(t, AnyPlot, bb) // panda cannot eat when no bamboo is grown (but it is still a player option)
	bb = b.MovePanda("p3")
	assert.Equal(t, GreenBambooPlot, bb) // panda eats one off the fertilizer plot
	assert.Equal(t, "p3", b.PandaLocation)
	// finish up with some gardener limitations
	b.MoveGardener("p3")
	assert.Equal(t, 4, b.Plots["p3"].Bamboo) // when fertilizer tile height = 3, only 1 bamboo is grown
	b.PlotGrowBamboo("p2")
	b.PlotGrowBamboo("p2")
	b.MoveGardener("p2")
	assert.Equal(t, 4, b.Plots["p2"].Bamboo) // gardener cannot make bamboo height exceed 4 on regular tile (but movement to that tile is a player option)
}

func TestImprovementTypeEqual(t *testing.T) {
	cases := []struct {
		a      ImprovementType
		b      ImprovementType
		expect bool
	}{
		{AnyImprovement, NoImprovement, true},
		{FertilizerImprovement, AnyImprovement, true},
		{NoImprovement, EnclosureImprovement, false},
		{FertilizerImprovement, FertilizerImprovement, true},
	}

	for _, tc := range cases {
		t.Run(string(tc.a+tc.b), func(tt *testing.T) {
			result := ImprovementTypeEqual(tc.a, tc.b)
			assert.Equal(tt, tc.expect, result)
		})
	}
}

func TestEdgeIndex(t *testing.T) {
	cases := []struct {
		Name string
		In   int
		Out  int
	}{
		{"", -2, 4},
		{"", 7, 1},
		{"", 6, 0},
		{"", 0, 0},
		{"", 3, 3},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(tt *testing.T) {
			i := edgeIndex(tc.In)
			assert.Equal(t, tc.Out, i)
		})
	}
}

func TestInverseIndex(t *testing.T) {
	cases := []struct {
		Name string
		In   int
		Out  int
	}{
		{"", 0, 3},
		{"", 1, 4},
		{"", 2, 5},
		{"", 3, 0},
		{"", 4, 1},
		{"", 5, 2},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(tt *testing.T) {
			i := inverseEdgeIndex(tc.In)
			assert.Equal(t, tc.Out, i)
		})
	}
}
