package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
)

type Board struct {
	PondID           string
	Plots            map[string]Plot
	Edges            map[string]Edge
	PandaLocation    string
	GardenerLocation string
	plotCount        int
	edgeCount        int
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

func (p *ImprovementType) UnmarshalText(b []byte) error {
	*p = ImprovementType(string(b))
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
	Irrigated bool
	ID        string
	Plots     [2]string // ids of the plots this edge is between
}

type Plot struct {
	Type        PlotType
	Edges       [6]string // ids of all the edges around this plot
	Improvement Improvement
	Bamboo      int
	ID          string
}

func (b *Board) NextPlotID() string {
	defer func() {
		b.plotCount++
	}()
	return fmt.Sprintf("p%d", b.plotCount)
}

func (b *Board) NextEdgeID() string {
	defer func() {
		b.edgeCount++
	}()
	return fmt.Sprintf("e%d", b.edgeCount)
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

func (b *Board) PlotAddImprovement(pid string, it ImprovementType) error {
	p := b.Plots[pid]
	if p.Improvement.Type != NoImprovement {
		return errors.New("Tile already has an improvement")
	}
	if p.Bamboo > 0 {
		return errors.New("Can't add an improvement where bamboo is already growing")
	}
	p.Improvement = Improvement{it, false}
	b.Plots[pid] = p
	return nil
}

func (b *Board) PlotGrowBamboo(pid string) error {
	p := b.Plots[pid]
	if !b.PlotIsIrrigated(pid) {
		return errors.New("Bamboo can only grow on irrigated tiles")
	}
	if p.Bamboo == 4 {
		return errors.New("Bamboo is already at max height")
	}
	if p.Improvement.Type == FertilizerImprovement {
		p.Bamboo = int(math.Min(4, float64(p.Bamboo)+2))
		return nil
	}
	p.Bamboo++
	b.Plots[pid] = p
	return nil
}

func (b *Board) PlotEatBamboo(pid string) error {
	p := b.Plots[pid]
	if p.Improvement.Type == EnclosureImprovement {
		return errors.New("Can't eat from this tile")
	}
	if p.Bamboo == 0 {
		return errors.New("Can't eat - no bamboo")
	}
	p.Bamboo--
	b.Plots[pid] = p
	return nil
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
	// if either of the plots this edge is between is irrigated already, and it's not because of a watershed, then this edge can be irrigated
	if p1.Improvement.Type != WatershedImprovement && b.PlotIsIrrigated(p1.ID) {
		return true
	}
	if p2.Improvement.Type != WatershedImprovement && b.PlotIsIrrigated(p2.ID) {
		return true
	}
	return false
}

func (b *Board) EdgeAddIrrigation(eid string) {
	e := b.Edges[eid]
	e.Irrigated = true
	b.Edges[eid] = e
}

func (b *Board) MovePanda(pid string) {
	b.PandaLocation = pid
}

func (b *Board) MoveGardener(pid string) {
	b.GardenerLocation = pid
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
		plotIds = append(plotIds, b.TileIDsInRow(pid, i)...)
	}
	return plotIds
}

// gets a list of all plots that can are irrigated and thus can be grown upon
// useful for Rain weather
func (b *Board) AllIrrigatedPlots() []string {
	plotIDs := make([]string, 0)
	for pid := range b.Plots {
		if b.PlotIsIrrigated(pid) {
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
		if plot.Improvement.Type == NoImprovement && plot.Improvement.Permanent == false && plot.Bamboo == 0 {
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
	}

}

// initialize a board with a pond tile and 6 future tiles around it
func NewBoard() *Board {
	b := new(Board)
	b.Plots = make(map[string]Plot)
	b.Edges = make(map[string]Edge)
	pondUUID := b.NextPlotID()
	b.PondID = pondUUID
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
		ID:    pondUUID,
		Edges: [6]string(pondEdgeIDs),
	}

	b.Plots[pondUUID] = pond

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
			Plots:     [2]string{pondUUID, futurePlotIDs[i]},
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

// Call this to deserialize a Board
// it re-populates private fields that json does not serialize
func UnmarshalBoard(s io.Reader) *Board {
	b := new(Board)
	json.NewDecoder(s).Decode(b)
	b.edgeCount = len(b.Edges)
	b.plotCount = len(b.Plots)
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

func DrawTiles() {

}

func PlaceTile() {

}

func PlaceIrrigation() {

}

func PlaceImprovement() {

}

func MoveGardener() {

}

func MovePanda() {

}

func GrowBamboo() {

}
