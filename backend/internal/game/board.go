package game

import (
	"fmt"
	"math"
	"slices"
)

type Board struct {
	PondID           string          `json:"pondLocation"`
	Plots            map[string]Plot `json:"plots"`
	Edges            map[string]Edge `json:"edges"`
	PandaLocation    string          `json:"pandaLocation"`
	GardenerLocation string          `json:"gardenerLocation"`
	PlotCount        int             `json:"plotCount"`
	EdgeCount        int             `json:"edgeCount"`
}

type PlotType string

const (
	PondPlot         PlotType = "POND"
	GreenBambooPlot  PlotType = "GREEN_BAMBOO"
	YellowBambooPlot PlotType = "YELLOW_BAMBOO"
	PinkBambooPlot   PlotType = "PINK_BAMBOO"
	FuturePlot       PlotType = "FUTURE"
	AnyPlot          PlotType = "ANY"
)

func (p *PlotType) UnmarshalText(b []byte) error {
	*p = PlotType(string(b))
	return nil
}

type ImprovementType string

const (
	WatershedImprovement  ImprovementType = "WATERSHED"
	EnclosureImprovement  ImprovementType = "ENCLOSURE"
	FertilizerImprovement ImprovementType = "FERTILIZER"
	NoImprovement         ImprovementType = "NONE"
	AnyImprovement        ImprovementType = "ANY"
)

func (i *ImprovementType) UnmarshalText(b []byte) error {
	*i = ImprovementType(string(b))
	return nil
}

func ImprovementTypeEqual(a, b ImprovementType) bool {
	if a == AnyImprovement || b == AnyImprovement {
		return true
	}
	return a == b
}

type Improvement struct {
	Type      ImprovementType
	Permanent bool
}

type Edge struct {
	Irrigated bool      `json:"irrigated"`
	ID        string    `json:"id"`
	Plots     [2]string `json:"plots"` // ids of the plots this edge is between
}

type Plot struct {
	Type        PlotType    `json:"type"`
	Edges       [6]string   `json:"edges"` // ids of all the edges around this plot
	Improvement Improvement `json:"improvement"`
	Bamboo      int         `json:"bambooHeight"`
	ID          string      `json:"id"`
}

func (b *Board) NextPlotID() string {
	defer func() {
		b.PlotCount++
	}()
	return fmt.Sprintf("p%d", b.PlotCount)
}

func (b *Board) NextEdgeID() string {
	defer func() {
		b.EdgeCount++
	}()
	return fmt.Sprintf("e%d", b.EdgeCount)
}

func (b *Board) PlotNeighbor(pid string, idx int) *Plot {
	p := b.Plots[pid]
	eid := p.Edges[edgeIndex(idx)]
	if eid == "" {
		return nil
	}
	e := b.Edges[eid]
	neighborPlotID := e.Plots[0]
	if neighborPlotID == pid {
		neighborPlotID = e.Plots[1]
	}
	neighbor := b.Plots[neighborPlotID]
	return &neighbor
}

func (b *Board) PlotIsIrrigated(pid string) bool {
	p := b.Plots[pid]
	if p.Type == PondPlot {
		return true
	}
	if p.Improvement.Type == WatershedImprovement {
		return true
	}
	for _, eid := range p.Edges {
		e := b.Edges[eid]
		if e.Irrigated {
			return true
		}
	}
	return false
}

func (b *Board) PlotAddImprovement(pid string, it ImprovementType) {
	p := b.Plots[pid]
	p.Improvement = Improvement{it, false}
	b.Plots[pid] = p
}

func (b *Board) PlotGrowBamboo(pid string) {
	p := b.Plots[pid]
	if !b.PlotIsIrrigated(pid) || p.Bamboo == 4 {
		return
	}
	if p.Bamboo < 3 && p.Improvement.Type == FertilizerImprovement {
		p.Bamboo++ // extra growth
	}
	p.Bamboo++
	b.Plots[pid] = p
}

func (b *Board) PlotEatBamboo(pid string) PlotType {
	p := b.Plots[pid]
	if p.Bamboo == 0 || p.Improvement.Type == EnclosureImprovement {
		return AnyPlot // anyplot == signal of no bamboo eaten
	}
	p.Bamboo--
	b.Plots[pid] = p
	return p.Type
}

func (b *Board) EdgeCouldBeIrrigated(eid string) bool {
	e := b.Edges[eid]
	if e.Irrigated {
		return false
	}
	p1 := b.Plots[e.Plots[0]]
	p2 := b.Plots[e.Plots[1]]
	if p1.Type == FuturePlot || p2.Type == FuturePlot {
		return false
	}
	for _, p := range []Plot{p1, p2} {

		i := slices.Index(p.Edges[:], eid)
		for _, dx := range []int{1, -1} {
			adjacentEdgeId := p.Edges[edgeIndex(i+dx)]
			adjacentEdge := b.Edges[adjacentEdgeId]
			if adjacentEdge.Irrigated {
				return true
			}
		}
	}
	return false
}

func (b *Board) EdgeAddIrrigation(eid string) {
	e := b.Edges[eid]
	e.Irrigated = true
	b.Edges[eid] = e
}

// returns AnyType plot if no bamboo is eaten
func (b *Board) MovePanda(pid string) PlotType {
	b.PandaLocation = pid
	return b.PlotEatBamboo(pid) // needs to place bamboo in player's inventory
}

func (b *Board) MoveGardener(pid string) {
	b.GardenerLocation = pid
	b.PlotGrowBamboo(pid)
}

// returns all the tiles is a row in the direction of eidx (edge index)
// useful for building a list of all legal moves for panda and garnder moves
func (b *Board) TileIDsInRow(pid string, eidx int) []string {
	p := b.Plots[pid]
	if p.Type == FuturePlot {
		return []string{}
	}
	next := b.PlotNeighbor(pid, eidx)
	if next == nil {
		return []string{pid}
	}
	return append(b.TileIDsInRow(next.ID, eidx), pid)
}

// gets all tiles in a straight line from the given plot that the gardener or panda could move to from the specified plot
func (b *Board) LegalMovesFromPlot(pid string) []string {
	plotIds := make([]string, 0)
	for i := 0; i < 6; i++ {
		nextPlot := b.PlotNeighbor(pid, i)
		plotIds = append(plotIds, b.TileIDsInRow(nextPlot.ID, i)...)
	}
	return plotIds
}

// returns the plotids of all non-future plots
func (b *Board) AllPresentPlots() []string {
	plotIDs := make([]string, 0)
	for pid, plot := range b.Plots {
		if plot.Type != FuturePlot {
			plotIDs = append(plotIDs, pid)
		}
	}
	return plotIDs
}

// gets a list of all plots that can are irrigated and thus can be grown upon
// useful for Rain weather
func (b *Board) AllIrrigatedPlots() []string {
	plotIDs := make([]string, 0)
	for pid, plot := range b.Plots {
		if b.PlotIsIrrigated(pid) && plot.Type != PondPlot && plot.Type != FuturePlot {
			plotIDs = append(plotIDs, pid)
		}
	}
	return plotIDs
}

// gets all the future plots, which are valid places for a new plot
func (b *Board) AllFuturePlots() []string {
	plotIDs := make([]string, 0)
	for pid, plot := range b.Plots {
		if plot.Type == FuturePlot {
			plotIDs = append(plotIDs, pid)
		}
	}
	return plotIDs
}

// gets all non-future plots where an improvement could be placed
func (b *Board) AllImprovablePlots() []string {
	plotIDs := make([]string, 0)
	for pid, plot := range b.Plots {
		if plot.Improvement.Type == NoImprovement && !plot.Improvement.Permanent && plot.Bamboo == 0 {
			plotIDs = append(plotIDs, pid)
		}
	}
	return plotIDs
}

// gets all ids of edges that are not irrigated but could be irrigated
func (b *Board) AllIrrigatableEdges() []string {
	edgeIDs := make([]string, 0)
	for eid := range b.Edges {
		if b.EdgeCouldBeIrrigated(eid) {
			edgeIDs = append(edgeIDs, eid)
		}
	}
	return edgeIDs
}

func (b *Board) AddPlot(pid string, pt PlotType, it ImprovementType) {
	f := b.Plots[pid]
	p := Plot{
		Type: pt,
		Improvement: Improvement{
			Type:      it,
			Permanent: it != NoImprovement,
		},
		Edges: f.Edges,
		ID:    pid,
	}

	b.Plots[pid] = p // replace future plot

	// edges to/from existing plots to pid are already established
	// need to check if any future plots can be created now, and
	// add it to the board along with all edges to it
	for i := 0; i < 6; i++ {
		nextIdx := edgeIndex(i + 1)
		neighbor1 := b.PlotNeighbor(pid, i)
		neighbor2 := b.PlotNeighbor(pid, nextIdx)
		var adjacentEdgeIdx int
		var nextEdgeIdx int
		var edgeIdxStep int
		if neighbor1 != nil && neighbor2 != nil {
			// both edges exist -> no future edge to add
			continue
		}
		if neighbor1 == nil && neighbor2 == nil {
			// both empty -> no future edge to add
			continue
		}
		if neighbor1 != nil && neighbor2 == nil && neighbor1.Type != FuturePlot {
			// create a future plot at index i
			// ++ direction clockwise around new plot
			adjacentEdgeIdx = nextIdx
			nextEdgeIdx = i
			edgeIdxStep = 1
		}
		if neighbor1 == nil && neighbor2 != nil && neighbor2.Type != FuturePlot {
			// create a future plot at index i + 1
			// -- direction counter clockwise around new plot
			adjacentEdgeIdx = i
			nextEdgeIdx = nextIdx
			edgeIdxStep = -1
		}
		if neighbor1 != nil && neighbor1.Type == FuturePlot {
			// do not make a future plot next to a future plot
			continue
		}
		if neighbor2 != nil && neighbor2.Type == FuturePlot {
			// do not make a future plot next to a future plot
			continue
		}
		futurePid := b.NextPlotID()
		futurePlot := Plot{
			Type: FuturePlot,
			Improvement: Improvement{
				Type:      NoImprovement,
				Permanent: true,
			},
			ID: futurePid,
		}
		for j := 0; j < 3; j++ {
			newEdgeID := b.NextEdgeID()
			p.Edges[adjacentEdgeIdx] = newEdgeID
			futurePlot.Edges[inverseEdgeIndex(adjacentEdgeIdx)] = newEdgeID

			newEdge := Edge{
				Irrigated: false,
				ID:        newEdgeID,
				Plots:     [2]string{p.ID, futurePid},
			}
			b.Edges[newEdgeID] = newEdge
			// save the edge added to p
			b.Plots[p.ID] = p
			// step to next plot
			next := b.PlotNeighbor(p.ID, nextEdgeIdx)
			if next == nil {
				break
			}

			// make p next
			p = *next

			adjacentEdgeIdx = edgeIndex(adjacentEdgeIdx + edgeIdxStep)
			nextEdgeIdx = edgeIndex(nextEdgeIdx + edgeIdxStep)
		}
		b.Plots[futurePid] = futurePlot
		p = b.Plots[pid]
	}

}

// initialize a board with a pond tile and 6 future tiles around it
func NewBoard() *Board {
	b := new(Board)
	b.Plots = make(map[string]Plot)
	b.Edges = make(map[string]Edge)
	b.PondID = b.NextPlotID()
	b.PandaLocation = b.PondID
	b.GardenerLocation = b.PondID

	var pondEdgeIDs [6]string
	for i := 0; i < 6; i++ {
		pondEdgeIDs[i] = b.NextEdgeID()
	}
	pond := Plot{
		Type: PondPlot,
		Improvement: Improvement{
			Type:      NoImprovement,
			Permanent: true,
		},
		ID:    b.PondID,
		Edges: [6]string(pondEdgeIDs),
	}

	b.Plots[b.PondID] = pond

	var futurePlotIDs [6]string
	for i := 0; i < 6; i++ {
		futurePlotIDs[i] = b.NextPlotID()
	}
	var futureEdges [6]Edge
	for i := 0; i < 6; i++ {
		// future edge i is the edge between future plot i and future plot i + 1
		eid := b.NextEdgeID()
		futureEdges[i] = Edge{
			Irrigated: false,
			ID:        eid,
			Plots:     [2]string{futurePlotIDs[i], futurePlotIDs[edgeIndex(i+1)]},
		}
		b.Edges[eid] = futureEdges[i]
	}
	for i := 0; i < 6; i++ {
		// create the edge between the pond and future plot
		e := Edge{
			Irrigated: true,
			ID:        pondEdgeIDs[i],
			Plots:     [2]string{b.PondID, futurePlotIDs[i]},
		}
		// add that edge to the board
		b.Edges[pondEdgeIDs[i]] = e

		var edges [6]string
		edges[edgeIndex(i+3)] = pondEdgeIDs[i]
		edges[edgeIndex(i+2)] = futureEdges[i].ID
		edges[edgeIndex(i+4)] = futureEdges[edgeIndex(i-1)].ID
		future := Plot{
			Type: FuturePlot,
			Improvement: Improvement{
				Type:      NoImprovement,
				Permanent: true,
			},
			ID:    futurePlotIDs[i],
			Edges: edges,
		}

		b.Plots[futurePlotIDs[i]] = future
	}

	return b
}

// ensures i is always between i and 6
// example usage (where i = 2):
//   - edgeIndex(i-3) -> 5
//   - edgeIndex(i+4) -> 0
func edgeIndex(i int) int {
	if i < 0 {
		absi := int(math.Abs(float64(i)))
		return 6 - absi%6
	}
	return i % 6
}

// returns the inverse edge index of an edge index.
// If an edge between plots a and b, it will be at index i for a and j for b where i and j are inverse indices of each other
// (0 and 3, 1 and 4, 2 and 5)
func inverseEdgeIndex(i int) int {
	return edgeIndex(i + 3)
}
