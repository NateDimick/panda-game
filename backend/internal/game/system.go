package game

import (
	"github.com/gofrs/uuid"
)

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
	NextPlayerTurn               PromptType = "NextPlayerTurn" // internal only.
	EndGame                      PromptType = "EndGame"        // internal only
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
	Gid        string
}

type PromptResponse struct {
	Action    PromptType
	Selection interface{}
	Pid       string
	Gid       string
}

// given the type of prompt, perform some type conversion so the returned value can be directly asserted to the desired type
func GetSelection(pt PromptType, s interface{}) interface{} {
	switch pt {
	case ChooseAction:
		// return ActionType
		return ActionType(s.(string))
	case ChooseObjectiveType:
		// convert to objectivetype
		return ObjectiveType(s.(string))
	case ChooseWeather:
		//convert to weathertype
		return WeatherType(s.(string))
	case ChooseImprovementToUse, ChooseImprovementToStash:
		// convert to ImprovementType
		return ImprovementType(s.(string))
	case ChoosePlot:
		// convert to DeckPlot
		m := s.(map[string]interface{})
		t := PlotType(m["Type"].(string))
		i := ImprovementType(m["Improvement"].(string))
		dp := DeckPlot{
			Type:        t,
			Improvement: i,
		}
		return dp
	default: // includes: ChooseGardenerDestination, ChooseImprovementDestination, ChooseIrrigationDestination, ChoosePandaDestination, ChoosePlotDestination, ChooseGrowth, RollDie (plotIds and edgeIds)
		return s
	}
}

func NewPromptID() string {
	return uuid.Must(uuid.NewV4()).String()
}

func ConvertToInterfaceSlice[T any](a []T) []interface{} {
	s := make([]interface{}, len(a))
	for i, v := range a {
		s[i] = v
	}
	return s
}

// if a prompt times out, use this for the game system to make an action and advance the game
func AutoPlay(t Turn) PromptResponse {
	return PromptResponse{
		Action:    t.CurrentPrompt.Action,
		Pid:       t.CurrentPrompt.Pid,
		Selection: t.CurrentPrompt.SelectFrom[0], // could also be random...
	}
}

func StartGame(players []*Player) *GameState {
	g := NewGame()
	g.AddPlayers(players)
	return g
}

func GameFlow(g *GameState, p PromptResponse) Prompt {
	if !g.ValidatePlayerAction(p) {
		// re-send prompt
		// TODO reduce prompt.Time based on how much time has passed since prompt issued
		return g.CurrentTurn.CurrentPrompt
	}
	prompt := g.ProcessPlayerAction(p)
	// complete objectives based on what the player just did
	g.CompleteObjectives()
	if prompt.Action == NextPlayerTurn {
		g.NextTurn()
		if g.GetCurrentPlayer().ID == g.EmperorWinner {
			// when turn circles back to emperor winner, the game ends
			return Prompt{
				Action: EndGame,
			}
		}
		// complete objectives at the beginning of a player's turn if other player's actions completed for them
		g.CompleteObjectives()
		prompt = Prompt{
			Action:     RollDie,
			SelectType: RollSelectType,
			SelectFrom: []interface{}{
				RollDie,
			},
			Time: 10,
			Pid:  NewPromptID(),
		}
	}
	g.CurrentTurn.CurrentPrompt = prompt
	return prompt
}
