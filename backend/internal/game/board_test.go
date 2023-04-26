package game

import (
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

func TestAddPlot(t *testing.T) {
	b := NewBoard()
	b.AddPlot("p1", GreenBambooPlot, NoImprovement)
	// adding the first plot did not add any plots or edges
	assert.Equal(t, 7, len(b.Plots))
	assert.Equal(t, 12, len(b.Edges))
	b.AddPlot("p2", YellowBambooPlot, NoImprovement)
	assert.Equal(t, 8, len(b.Plots))
	assert.Equal(t, 14, len(b.Edges))
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
