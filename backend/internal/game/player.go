package game

import (
	"encoding/json"
)

type Player struct {
	Name string
	// a unique identifier for the player
	ID string
	// the player's turn order
	Order int
	// the number of irrigations in reserve
	Irrigations int
	// the player's eaten bamboo reserve (a count of each type)
	Bamboo BambooReserve
	// the player's improvement reserve (a count of each type)
	Improvements ImprovementReserve
	// the objective cards in the player's possession, complete or incomplete
	Objectives []Objective
}

type BambooReserve struct {
	Green  int
	Yellow int
	Pink   int
}

type ImprovementReserve struct {
	Watersheds   int
	Fertilizers  int
	Enclosements int
}

type ObjectiveType string

const (
	PandaObjectiveType   ObjectiveType = "PANDA"
	GardnerObjectiveType ObjectiveType = "GARDNER"
	PlotObjectiveType    ObjectiveType = "PLOT"
)

func (p *ObjectiveType) UnmarshalText(b []byte) error {
	*p = ObjectiveType(string(b))
	return nil
}

type ObjectiveChecker interface {
	// takes the player and the board (type TBD) and
	IsComplete(Player, Board) bool
	// the value of the objective, if complete
	Points() int
	// the type of the objective, either PANDA, GARDNER or PLOT
	Type() ObjectiveType
}

type Objective struct {
	ObjectiveChecker
}

func (o *Objective) UnmarshalJSON(b []byte) error {
	m := make(map[string]interface{})
	json.Unmarshal(b, &m)
	ob := m["ObjectiveChecker"].(map[string]interface{})

	bs, _ := json.Marshal(ob)

	switch ObjectiveType(ob["OT"].(string)) {
	case PandaObjectiveType:
		ob := new(PandaObjective)
		if err := json.Unmarshal(bs, ob); err != nil {
			return err
		}
		o.ObjectiveChecker = *ob
	case PlotObjectiveType:
		ob := new(PlotObjective)
		if err := json.Unmarshal(bs, ob); err != nil {
			return err
		}
		o.ObjectiveChecker = *ob
	case GardnerObjectiveType:
		ob := new(GardnerObjective)
		if err := json.Unmarshal(bs, ob); err != nil {
			return err
		}
		o.ObjectiveChecker = *ob
	}
	return nil
}

type PandaObjective struct {
	GreenCount  int
	YellowCount int
	PinkCount   int
	Value       int
	OT          ObjectiveType
}

func (o PandaObjective) Points() int {
	return o.Value
}

func (o PandaObjective) Type() ObjectiveType {
	return PandaObjectiveType
}

func (o PandaObjective) IsComplete(p Player, b Board) bool {
	greenOK := p.Bamboo.Green >= o.GreenCount
	yellowOK := p.Bamboo.Yellow >= o.YellowCount
	pinkOK := p.Bamboo.Pink >= o.PinkCount
	return greenOK && yellowOK && pinkOK
}

type GardnerObjective struct {
	Color       PlotType
	Height      int
	Count       int
	Improvement ImprovementType
	Value       int
	OT          ObjectiveType
}

func (o GardnerObjective) Points() int {
	return o.Value
}

func (o GardnerObjective) Type() ObjectiveType {
	return GardnerObjectiveType
}

func (o GardnerObjective) IsComplete(p Player, b Board) bool {
	okPlots := 0
	for _, plot := range b.Plots {
		improvementOK := ImprovementTypeEqual(o.Improvement, plot.Improvement.Type)
		colorOK := o.Color == plot.Type
		heightOK := plot.Bamboo >= o.Height
		plotOK := improvementOK && heightOK && colorOK
		if plotOK {
			okPlots++
		}
		if okPlots >= o.Count {
			return true
		}
	}
	return false
}

type PlotObjective struct {
	Value       int
	AnchorColor PlotType
	Neighbors   [6]PlotType
	OT          ObjectiveType
}

func (o PlotObjective) Points() int {
	return o.Value
}

func (o PlotObjective) Type() ObjectiveType {
	return PlotObjectiveType
}

func (o PlotObjective) IsComplete(p Player, b Board) bool {
	for _, plot := range b.Plots {
		if plot.Type == o.AnchorColor && b.PlotIsIrrigated(plot.ID) {
			var neighborPlots [6]*Plot
			// build array of neighboring plots
			for i := 0; i < 6; i++ {
				neighbor := b.PlotNeighbor(plot.ID, i)
				neighborPlots[i] = neighbor
			}
			// check neighbor pattern against each perspective
			for persp := 0; persp < 6; persp++ {
				match := true
				for i, neighborType := range o.Neighbors {
					if neighborType == AnyPlot {
						continue
					}
					neighborIdx := edgeIndex(persp + i)
					if neighborPlots[neighborIdx] == nil {
						match = false
						break
					}
					if neighborPlots[neighborIdx].Type != neighborType || !b.PlotIsIrrigated(neighborPlots[neighborIdx].ID) {
						match = false
						break
					}
				}
				if match {
					return true
				}
			}
		}
	}
	return false
}
