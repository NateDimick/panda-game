package game

import "github.com/gofrs/uuid"

type ActionType string

const (
	PlacePlot         ActionType = "PlacePlot"
	CollectIrrigation ActionType = "CollectIrrigation"
	MovePanda         ActionType = "MovePanda"
	MoveGardener      ActionType = "MoveGardener"
	DrawObjective     ActionType = "DrawObjective"
	PlaceIrrigation   ActionType = "PlaceIrrigation"
	PlaceImprovement  ActionType = "PlaceImprovement"
	EndTurn           ActionType = "EndTurn"
)

func (p *ActionType) UnmarshalText(b []byte) error {
	*p = ActionType(string(b))
	return nil
}

type PromptType string

const (
	RollDie                      PromptType = "RollDie"
	ChooseWeather                PromptType = "ChooseWeather"
	ChooseImprovementToUse       PromptType = "ChooseImprovementToUse"
	ChooseImprovementToStash     PromptType = "ChooseImprovementToStash"
	ChooseGrowth                 PromptType = "ChooseGrowth"
	ChoosePandaDestination       PromptType = "ChoosePandaDestination"
	ChooseAction                 PromptType = "ChooseAction"
	ChoosePlot                   PromptType = "ChoosePlot"
	ChoosePlotDestination        PromptType = "ChoosePlotDestination"
	ChooseGardenerDestination    PromptType = "ChooseGardenerDestination"
	ChooseObjectiveType          PromptType = "ChooseObjectiveType"
	ChooseIrrigationDestination  PromptType = "ChooseIrrigationDestination"
	ChooseImprovementDestination PromptType = "ChooseImprovementDestination"
)

func (p *PromptType) UnmarshalText(b []byte) error {
	*p = PromptType(string(b))
	return nil
}

type SelectType string

const (
	RollSelectType        SelectType = "RollDie"
	ActionSelectType      SelectType = "ActionType"
	ImprovementSelectType SelectType = "ImprovementType"
	ObjectiveSelectType   SelectType = "ObjectiveType"
	PlotSelectType        SelectType = "PlotType"
	WeatherSelectType     SelectType = "WeatherType"
	EdgeIDSelectType      SelectType = "EdgeId"
	PlotIDSelectType      SelectType = "PlotId"
)

func (p *SelectType) UnmarshalText(b []byte) error {
	*p = SelectType(string(b))
	return nil
}

type Prompt struct {
	Action     PromptType
	SelectType SelectType
	SelectFrom []interface{}
	Time       int
	Pid        string
}

type PromptResponse struct {
	Action    PromptType
	Selection interface{}
	Pid       string
}

func NewPromptID() string {
	return uuid.Must(uuid.NewV4()).String()
}

func ConvertToInterfaceSlice(a ...any) []interface{} {
	s := make([]interface{}, len(a))
	for i, v := range a {
		s[i] = v
	}
	return s
}

// if a prompt times out, use this for the game system to make an action and advance the game
func AutoPlay(t Turn) PromptResponse {
	return PromptResponse{
		Action:    t.CurrentCommand,
		Pid:       t.Pid,
		Selection: t.CurrentOptions[0],
	}
}
