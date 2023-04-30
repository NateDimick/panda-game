package game

import (
	"encoding/json"
	"errors"
	"math/rand"

	"golang.org/x/exp/slices"
)

type GameState struct {
	// what has been placed on the board and where
	Board *Board
	// who's playing and what they posess
	Players []*Player
	// info about the current player turn. can be used to re-issue a command to a re-joining player
	CurrentTurn *Turn
	// unused plot tiles
	PlotDeck []DeckPlot
	// unused improvements
	AvailableImprovements ImprovementReserve
	// number of irrigations available
	IrrigationReserve int
	// undrawn objectives
	ObjectiveDecks map[ObjectiveType][]Objective
	// messages sent by the server to players
	GameLog []string
	// messages sent by players to the room
	ChatLog []ChatMessage
}

type DeckPlot struct {
	Type        PlotType
	Improvement ImprovementType
}

type Turn struct {
	PlayerID          string
	ActionsUsed       []ActionType // the actions the player has already taken
	CurrentCommand    PromptType
	CurrentOptions    []interface{}
	CurrentOptionType SelectType
	ContextSelection  interface{} // for actions that require 2 choices, this is the first choice
	Weather           WeatherType
	Pid               string
}

func NewTurn() Turn {
	return Turn{Weather: NoWeather}
}

type ChatMessage struct {
	From    string
	Message string
}

func NewGame() *GameState {
	od := make(map[ObjectiveType][]Objective)
	ir := make(map[ImprovementType]int)
	pd := make([]DeckPlot, 0)
	err1 := json.Unmarshal(InitialObjectiveDeck, &od)
	err2 := json.Unmarshal(InitialImprovements, &ir)
	err3 := json.Unmarshal(InitialPlotDeck, &pd)
	err := errors.Join(err1, err2, err3)
	if err != nil {
		panic(err)
	}
	// shuffle plots and objectives, 3 times each to get them mixed up good
	for x := 0; x < 3; x++ {
		rand.Shuffle(len(pd), func(i, j int) {
			pd[i], pd[j] = pd[j], pd[i]
		})
		rand.Shuffle(len(od[PlotObjectiveType]), func(i, j int) {
			od[PlotObjectiveType][i], od[PlotObjectiveType][j] = od[PlotObjectiveType][j], od[PlotObjectiveType][i]
		})
		rand.Shuffle(len(od[PandaObjectiveType]), func(i, j int) {
			od[PandaObjectiveType][i], od[PandaObjectiveType][j] = od[PandaObjectiveType][j], od[PandaObjectiveType][i]
		})
		rand.Shuffle(len(od[GardenerObjectiveType]), func(i, j int) {
			od[GardenerObjectiveType][i], od[GardenerObjectiveType][j] = od[GardenerObjectiveType][j], od[GardenerObjectiveType][i]
		})
	}
	g := &GameState{
		Board:                 NewBoard(),
		IrrigationReserve:     20,
		ObjectiveDecks:        od,
		AvailableImprovements: ir,
		PlotDeck:              pd,
		GameLog:               make([]string, 0),
		ChatLog:               make([]ChatMessage, 0),
	}
	return g
}

func (g *GameState) DrawPlots() []DeckPlot {
	if len(g.PlotDeck) > 3 {
		plotOptions := g.PlotDeck[:3]
		g.PlotDeck = g.PlotDeck[3:]
		return plotOptions
	}
	plotOptions := g.PlotDeck
	g.PlotDeck = make([]DeckPlot, 0)
	return plotOptions
}

func (g *GameState) ReturnPlots(usedPlot map[string]interface{}, options []interface{}) {
	usedSkipped := false
	for _, i := range options {
		p := i.(map[string]interface{})
		if !usedSkipped {
			if p["Type"] == usedPlot["Type"] && p["Improvement"] == usedPlot["Improvement"] {
				usedSkipped = true
				continue
			}
		}
		dp := DeckPlot{
			Type:        PlotType(p["Type"].(string)),
			Improvement: ImprovementType(p["Improvement"].(string)),
		}
		g.PlotDeck = append(g.PlotDeck, dp)
	}

}

func (g *GameState) AvailableObjectiveTypes() []ObjectiveType {
	s := make([]ObjectiveType, 0)
	for k, v := range g.ObjectiveDecks {
		if len(v) > 0 {
			s = append(s, k)
		}
	}
	return s
}

func (g *GameState) DrawObjective(ot ObjectiveType) Objective {
	o := g.ObjectiveDecks[ot][0]
	g.ObjectiveDecks[ot] = g.ObjectiveDecks[ot][1:]
	return o
}

func (g GameState) GetPlayer(pid string) *Player {
	for _, p := range g.Players {
		if p.ID == pid {
			return p
		}
	}
	return nil
}

func (g *GameState) NextChooseActionPrompt() Prompt {
	p := Prompt{
		Action:     ChooseAction,
		SelectType: ActionSelectType,
		Time:       60,
		Pid:        NewPromptID(),
	}

	currentPlayer := g.GetPlayer(g.CurrentTurn.PlayerID)

	options := make([]interface{}, 0) // values will be ActionType

	options = append(options, availableRegularActions(g.CurrentTurn.ActionsUsed, g.CurrentTurn.Weather))

	couldEndTurn := false

	if len(options) == 0 {
		// if no regular action options, then manual end turn could be on the table if the player has resources in reserve
		couldEndTurn = true
	}

	if currentPlayer.Irrigations > 0 {
		options = append(options, PlaceIrrigation)
	}
	if !currentPlayer.Improvements.IsEmpty() {
		options = append(options, PlaceImprovement)
	}
	// add end turn if condition is met
	if len(options) > 0 && couldEndTurn {
		options = append(options, EndTurn)
	}
	p.SelectFrom = options
	return p
}

func availableRegularActions(used []ActionType, weather WeatherType) []ActionType {
	if len(used) == 2 && weather != SunWeather {
		return []ActionType{}
	}
	if len(used) == 3 {
		return []ActionType{}
	}
	regularActions := []ActionType{PlacePlot, MovePanda, MoveGardener, CollectIrrigation, DrawObjective}
	if weather == WindWeather {
		return regularActions
	}
	for _, a := range used {
		i := slices.Index(regularActions, a)
		regularActions = slices.Delete(regularActions, i, i+1)
	}
	return regularActions
}

func (g *GameState) ValidatePlayerAction(action PromptResponse) bool {
	if g.CurrentTurn.CurrentCommand != action.Action {
		return false
	}
	if g.CurrentTurn.Pid != action.Pid {
		return false
	}
	for _, opt := range g.CurrentTurn.CurrentOptions {
		if opt == action.Selection { // this is a toss up whether it will work or not (because custom types) but my gut says both of these will be freshly json parsed as generic json types (string or map[string]interface{}) so the comparison will be ok
			return true
		}
	}
	return false
}

func (g *GameState) ProcessPlayerAction(action PromptResponse) Prompt {
	switch action.Action {
	case ChooseAction:
		// the player has chosen an action. They require a next prompt
		at := ActionType(action.Selection.(string))
		g.CurrentTurn.ActionsUsed = append(g.CurrentTurn.ActionsUsed, at)
		return g.PromptForAction(at)
	case ChooseWeather:
		//
		g.CurrentTurn.Weather = WeatherType(action.Selection.(string))
		return g.NextChooseActionPrompt()
	case ChooseGrowth:
		// grow 1 bamboo on the selected plot, the prompt next
		g.Board.PlotGrowBamboo(action.Selection.(string))
		return g.NextChooseActionPrompt()
	case ChoosePandaDestination:
		// eat 1 bamboo on the selected plot, then prompt next
		bamboo := g.Board.MovePanda(action.Selection.(string))
		if bamboo != AnyPlot && bamboo != PondPlot {
			p := g.GetPlayer(g.CurrentTurn.PlayerID)
			p.Bamboo[bamboo]++
		}
		return g.NextChooseActionPrompt()
	case ChooseGardenerDestination:
		// move gardner
		g.Board.MoveGardener(action.Selection.(string))
		return g.NextChooseActionPrompt()
	case ChooseImprovementDestination:
		//
		g.Board.PlotAddImprovement(action.Selection.(string), ImprovementType(g.CurrentTurn.ContextSelection.(string)))
		g.NextChooseActionPrompt()
	case ChoosePlotDestination:
		//
		selectedPlot := g.CurrentTurn.ContextSelection.(map[string]interface{})
		g.Board.AddPlot(action.Selection.(string), PlotType(selectedPlot["Type"].(string)), ImprovementType(selectedPlot["Improvement"].(string)))
		g.ReturnPlots(selectedPlot, g.CurrentTurn.CurrentOptions)
	case ChooseIrrigationDestination:
		//
		g.Board.EdgeAddIrrigation(action.Selection.(string))
		// TODO: remove irrigation from player's inventory
		return g.NextChooseActionPrompt()
	case ChoosePlot:
		//
		g.CurrentTurn.ContextSelection = action.Selection
		options := g.Board.AllFuturePlots()
		return Prompt{
			Action:     ChoosePlotDestination,
			SelectType: PlotIDSelectType,
			SelectFrom: ConvertToInterfaceSlice(options),
			Time:       45,
			Pid:        NewPromptID(),
		}
	case ChooseImprovementToUse:
		//
		g.CurrentTurn.ContextSelection = action.Selection
		// TODO: remove improvement from player's inventory
		options := g.Board.AllImprovablePlots()
		return Prompt{
			Action:     ChooseImprovementDestination,
			SelectType: PlotIDSelectType,
			SelectFrom: ConvertToInterfaceSlice(options),
			Time:       45,
			Pid:        NewPromptID(),
		}
	case ChooseImprovementToStash:
		p := g.GetPlayer(g.CurrentTurn.PlayerID)
		p.Improvements[ImprovementType(action.Selection.(string))]++
		return g.NextChooseActionPrompt()
	case ChooseObjectiveType:
		//
		o := g.DrawObjective(ObjectiveType(action.Selection.(string)))
		p := g.GetPlayer(g.CurrentTurn.PlayerID)
		p.Objectives = append(p.Objectives, o)
		return g.NextChooseActionPrompt()
	case RollDie:
		//
		w := RollWeatherDie(!g.AvailableImprovements.IsEmpty())
		if w == ChoiceWeather {
			options := []WeatherType{
				SunWeather,
				WindWeather,
				RainWeather,
				BoltWeather,
			}
			if !g.AvailableImprovements.IsEmpty() {
				options = append(options, CloudWeather)
			}
			return Prompt{
				Action:     ChooseWeather,
				SelectType: WeatherSelectType,
				SelectFrom: ConvertToInterfaceSlice(options),
				Pid:        NewPromptID(),
				Time:       30,
			}
		}
		// else, prompt user with given weather
		g.CurrentTurn.Weather = w
		if w == RainWeather {
			// prompt for growth
			options := g.Board.AllIrrigatedPlots()
			return Prompt{
				Action:     ChooseGrowth,
				SelectType: PlotIDSelectType,
				SelectFrom: ConvertToInterfaceSlice(options),
				Time:       45,
				Pid:        NewPromptID(),
			}
		} else if w == BoltWeather {
			// prompt for panda move
			options := g.Board.AllPresentPlots()
			return Prompt{
				Action:     ChoosePandaDestination,
				SelectType: PlotIDSelectType,
				SelectFrom: ConvertToInterfaceSlice(options),
				Time:       45,
				Pid:        NewPromptID(),
			}
		} else if w == CloudWeather {
			// prompt for improvement selection
			options := g.AvailableImprovements.AvailableImprovements()
			return Prompt{
				Action:     ChooseImprovementToStash,
				SelectType: ImprovementSelectType,
				SelectFrom: ConvertToInterfaceSlice(options),
				Time:       30,
				Pid:        NewPromptID(),
			}
		} else {
			// sun and wind proceed like normal
			return g.NextChooseActionPrompt()
		}
	}
	return Prompt{}
}

func (g *GameState) PromptForAction(at ActionType) Prompt {
	switch at {
	case PlacePlot:
		//
		options := g.DrawPlots()
		return Prompt{
			Action:     ChoosePlot,
			SelectType: PlotSelectType,
			SelectFrom: ConvertToInterfaceSlice(options),
			Time:       45,
			Pid:        NewPromptID(),
		}
	case CollectIrrigation:
		//
		p := g.GetPlayer(g.CurrentTurn.PlayerID)
		p.Irrigations++
		return g.NextChooseActionPrompt()
	case MovePanda:
		//
		options := g.Board.LegalMovesFromPlot(g.Board.PandaLocation)
		return Prompt{
			Action:     ChoosePandaDestination,
			SelectType: PlotIDSelectType,
			SelectFrom: ConvertToInterfaceSlice(options),
			Time:       45,
			Pid:        NewPromptID(),
		}
	case MoveGardener:
		//
		options := g.Board.LegalMovesFromPlot(g.Board.GardenerLocation)
		return Prompt{
			Action:     ChooseGardenerDestination,
			SelectType: PlotIDSelectType,
			SelectFrom: ConvertToInterfaceSlice(options),
			Time:       45,
			Pid:        NewPromptID(),
		}
	case DrawObjective:
		//
		options := g.AvailableObjectiveTypes()
		return Prompt{
			Action:     ChooseObjectiveType,
			SelectType: ObjectiveSelectType,
			SelectFrom: ConvertToInterfaceSlice(options),
			Time:       45,
			Pid:        NewPromptID(),
		}
	case PlaceIrrigation:
		//
		options := g.Board.AllIrrigatableEdges()
		return Prompt{
			Action:     ChooseIrrigationDestination,
			SelectType: EdgeIDSelectType,
			SelectFrom: ConvertToInterfaceSlice(options),
			Time:       45,
			Pid:        NewPromptID(),
		}
	case PlaceImprovement:
		//
		p := g.GetPlayer(g.CurrentTurn.PlayerID)
		options := p.Improvements.AvailableImprovements()
		return Prompt{
			Action:     ChooseImprovementToUse,
			SelectType: ImprovementSelectType,
			SelectFrom: ConvertToInterfaceSlice(options),
			Time:       30,
			Pid:        NewPromptID(),
		}
	case EndTurn:
		fallthrough
	default:
		return Prompt{}
	}
}
