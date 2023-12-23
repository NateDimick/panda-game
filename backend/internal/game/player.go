package game

import (
	"encoding/json"
)

type Player struct {
	Name string `json:"name"`
	// a unique identifier for the player
	ID string `json:"-"`
	// the player's turn order
	Order int `json:"-"`
	// the number of irrigations in reserve
	Irrigations int `json:"irrigationReserve"`
	// the player's eaten bamboo reserve (a count of each type)
	Bamboo BambooReserve `json:"bambooReserve"`
	// the player's improvement reserve (a count of each type)
	Improvements ImprovementReserve `json:"improvementReserve"`
	// the objective cards in the player's possession, incomplete
	Objectives []Objective `json:"-"` // TODO: when sending to UI, share number of objectives and types, but not secret info (value, goal)
	// Objectives in the player's possession that have been completed
	CompleteObjectives []Objective `json:"completeObjectives"`
}

type BambooReserve map[PlotType]int

type ImprovementReserve map[ImprovementType]int

func (i ImprovementReserve) IsEmpty() bool {
	isum := 0
	for _, v := range i {
		isum += v
	}
	return isum == 0
}

func (i ImprovementReserve) AvailableImprovements() []ImprovementType {
	s := make([]ImprovementType, 0)
	for k, v := range i {
		if v > 0 {
			s = append(s, k)
		}
	}
	return s
}

type ObjectiveType string

const (
	PandaObjectiveType    ObjectiveType = "PANDA"
	GardenerObjectiveType ObjectiveType = "GARDENER"
	PlotObjectiveType     ObjectiveType = "PLOT"
	EmperorObjectiveType  ObjectiveType = "EMPEROR"
)

func (p *ObjectiveType) UnmarshalText(b []byte) error {
	*p = ObjectiveType(string(b))
	return nil
}

type ObjectiveChecker interface {
	// takes the player and the board and checks if the objective is satisfied
	IsComplete(Player, Board) bool
	// the value of the objective, if complete
	Points() int
	// the type of the objective, either PANDA, GARDENER or PLOT
	Type() ObjectiveType
}

type Objective struct {
	ObjectiveChecker
}

func (o *Objective) UnmarshalJSON(b []byte) error {
	m := make(map[string]interface{})
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	ob := m["ObjectiveChecker"].(map[string]interface{})

	bs, err := json.Marshal(ob)
	if err != nil {
		return err
	}

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
	case GardenerObjectiveType:
		ob := new(GardenerObjective)
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
	greenOK := p.Bamboo[GreenBambooPlot] >= o.GreenCount
	yellowOK := p.Bamboo[YellowBambooPlot] >= o.YellowCount
	pinkOK := p.Bamboo[PinkBambooPlot] >= o.PinkCount
	return greenOK && yellowOK && pinkOK
}

type GardenerObjective struct {
	Color       PlotType
	Height      int
	Count       int
	Improvement ImprovementType
	Value       int
	OT          ObjectiveType
}

func (o GardenerObjective) Points() int {
	return o.Value
}

func (o GardenerObjective) Type() ObjectiveType {
	return GardenerObjectiveType
}

func (o GardenerObjective) IsComplete(p Player, b Board) bool {
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

type EmperorObjective struct{}

func (o EmperorObjective) Points() int {
	return 2
}

func (o EmperorObjective) Type() ObjectiveType {
	return EmperorObjectiveType
}

func (o EmperorObjective) IsComplete(Player, Board) bool {
	return true
}
