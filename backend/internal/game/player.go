package game

import (
	"encoding/json"
)

type Player struct {
	Name string `json:"name"`
	// a unique identifier for the player
	ID string `json:"id"`
	// the player's turn order
	Order int `json:"order"`
	// the number of irrigations in reserve
	Irrigations int `json:"irrigationReserve"`
	// the player's eaten bamboo reserve (a count of each type)
	Bamboo BambooReserve `json:"bambooReserve"`
	// the player's improvement reserve (a count of each type)
	Improvements ImprovementReserve `json:"improvementReserve"`
	// the objective cards in the player's possession, incomplete
	Objectives []Objective `json:"objectives"` // TODO: when sending to UI, share number of objectives and types, but not secret info (value, goal)
	// Objectives in the player's possession that have been completed
	CompleteObjectives []Objective `json:"completeObjectives"`
}

type ClientPlayer struct {
	Name               string                `json:"name"`
	Position           int                   `json:"position"`
	Irrigations        int                   `json:"irrigationReserve"`
	Bamboo             BambooReserve         `json:"bambooReserve"`
	Improvements       ImprovementReserve    `json:"improvementReserve"`
	Objectives         []Objective           `json:"objectives,omitempty"`
	HiddenObjectives   map[ObjectiveType]int `json:"hiddenObjectives,omitempty"`
	CompleteObjectives []Objective           `json:"completeObjectives"`
}

func (p Player) ClientSafe(recipient string) ClientPlayer {
	c := ClientPlayer{
		Name:               p.Name,
		Position:           p.Order,
		Irrigations:        p.Irrigations,
		Bamboo:             p.Bamboo,
		Improvements:       p.Improvements,
		CompleteObjectives: p.CompleteObjectives,
	}

	if p.ID == recipient {
		c.Objectives = p.Objectives
	} else {
		h := make(map[ObjectiveType]int)
		for _, o := range p.Objectives {
			h[o.Type()]++
		}
		c.HiddenObjectives = h
	}

	return c
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

func (o *ObjectiveType) UnmarshalText(b []byte) error {
	*o = ObjectiveType(string(b))
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

func (o *Objective) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.ObjectiveChecker)
}

func (o *Objective) UnmarshalJSON(b []byte) error {
	m := make(map[string]any)
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	switch ObjectiveType(m["type"].(string)) {
	case PandaObjectiveType:
		ob := new(PandaObjective)
		if err := json.Unmarshal(b, ob); err != nil {
			return err
		}
		o.ObjectiveChecker = *ob
	case PlotObjectiveType:
		ob := new(PlotObjective)
		if err := json.Unmarshal(b, ob); err != nil {
			return err
		}
		o.ObjectiveChecker = *ob
	case GardenerObjectiveType:
		ob := new(GardenerObjective)
		if err := json.Unmarshal(b, ob); err != nil {
			return err
		}
		o.ObjectiveChecker = *ob
	}
	return nil
}

type PandaObjective struct {
	GreenCount  int           `json:"greenRequired"`
	YellowCount int           `json:"yellowRequired"`
	PinkCount   int           `json:"pinkRequired"`
	Value       int           `json:"points"`
	OT          ObjectiveType `json:"type"`
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
	Color       PlotType        `json:"plotType"`
	Height      int             `json:"requiredHeight"`
	Count       int             `json:"requiredShoots"`
	Improvement ImprovementType `json:"improvementCondition"`
	Value       int             `json:"points"`
	OT          ObjectiveType   `json:"type"`
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
	Value       int           `json:"points"`
	AnchorColor PlotType      `json:"anchorPlot"`
	Neighbors   [6]PlotType   `json:"neighborPlots"`
	OT          ObjectiveType `json:"type"`
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

type EmperorObjective struct {
	Value int           `json:"points"`
	OT    ObjectiveType `json:"type"`
}

func (o EmperorObjective) Points() int {
	return 2
}

func (o EmperorObjective) Type() ObjectiveType {
	return EmperorObjectiveType
}

func (o EmperorObjective) IsComplete(Player, Board) bool {
	return true
}
