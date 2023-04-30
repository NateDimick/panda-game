package game

import (
	"encoding/json"
	"errors"
	"math/rand"

	"golang.org/x/exp/slices"
)

type WeatherType string

const (
	SunWeather    WeatherType = "SUN"    // Sun allows the player to take a third action
	RainWeather   WeatherType = "RAIN"   // Rain allows the player to grow one bamboo shoot by 1 unit
	WindWeather   WeatherType = "WIND"   // Wind allows the player to take the same action twice (not required, just allowed)
	BoltWeather   WeatherType = "BOLT"   // Bolt allows the player to move the panda anywhere
	CloudWeather  WeatherType = "CLOUD"  // Cloud allows the player to take one available improvement. If no improvements are available, then this becomes a choice
	ChoiceWeather WeatherType = "CHOICE" // Choice allows the player to choose which of the 5 above weather conditions for their turn.
	NoWeather     WeatherType = "NONE"
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
	// playerID of player who won emperor
	EmperorWinner string
}

type DeckPlot struct {
	Type        PlotType
	Improvement ImprovementType
}

type Turn struct {
	PlayerID         string
	Order            int
	ActionsUsed      []ActionType // the actions the player has already taken
	CurrentPrompt    Prompt
	ContextSelection interface{} // for actions that require 2 choices, this is the first choice
	Weather          WeatherType
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
		CurrentTurn: &Turn{
			Order:       0,
			Weather:     NoWeather,
			ActionsUsed: make([]ActionType, 0),
		},
	}
	return g
}

func (g *GameState) AddPlayers(ps []*Player) {
	// shuffle player order
	rand.Shuffle(len(ps), func(i, j int) {
		ps[1], ps[j] = ps[j], ps[i]
	})
	g.Players = ps
	g.CurrentTurn.Order = len(ps) - 1 // when NextTurn is called the first time, it will increment up to player 0
	g.CurrentTurn.PlayerID = ps[0].ID
}

func (g *GameState) NextTurn() Turn {
	order := (g.CurrentTurn.Order + 1) % len(g.Players)
	player := g.Players[order]
	return Turn{
		Order:       order,
		PlayerID:    player.ID,
		Weather:     NoWeather,
		ActionsUsed: make([]ActionType, 0),
	}
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

func (g *GameState) ReturnPlots(usedPlot DeckPlot, options []DeckPlot) {
	usedSkipped := false
	for _, opt := range options {
		if !usedSkipped {
			if opt.Type == usedPlot.Type && opt.Improvement == usedPlot.Improvement {
				usedSkipped = true
				continue
			}
		}
		g.PlotDeck = append(g.PlotDeck, opt)
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

func (g GameState) GetCurrentPlayer() *Player {
	pid := g.CurrentTurn.PlayerID
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

	currentPlayer := g.GetCurrentPlayer()

	options := make([]ActionType, 0) // values will be ActionType

	options = append(options, g.availableRegularActions()...)

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
	if len(options) == 0 {
		return Prompt{Action: NextPlayerTurn}
	}
	p.SelectFrom = ConvertToInterfaceSlice(options)
	return p
}

func (g *GameState) availableRegularActions() []ActionType {
	used := g.CurrentTurn.ActionsUsed
	weather := g.CurrentTurn.Weather
	if len(used) == 2 && weather != SunWeather {
		return []ActionType{}
	}
	if len(used) == 3 {
		return []ActionType{}
	}
	regularActions := []ActionType{PlacePlot, MovePanda, MoveGardener, CollectIrrigation, DrawObjective}
	if weather != WindWeather {
		for _, a := range used {
			i := slices.Index(regularActions, a)
			regularActions = slices.Delete(regularActions, i, i+1)
		}
	}
	if g.IrrigationReserve == 0 {
		i := slices.Index(regularActions, CollectIrrigation)
		regularActions = slices.Delete(regularActions, i, i+1)
	}
	if len(g.PlotDeck) == 0 {
		i := slices.Index(regularActions, PlacePlot)
		regularActions = slices.Delete(regularActions, i, i+1)
	}
	if len(g.AvailableObjectiveTypes()) == 0 {
		i := slices.Index(regularActions, DrawObjective)
		regularActions = slices.Delete(regularActions, i, i+1)
	}
	return regularActions
}

func (g *GameState) ValidatePlayerAction(action PromptResponse) bool {
	if g.CurrentTurn.CurrentPrompt.Action != action.Action {
		return false
	}
	if g.CurrentTurn.CurrentPrompt.Pid != action.Pid {
		return false
	}
	for _, opt := range g.CurrentTurn.CurrentPrompt.SelectFrom {
		if opt == action.Selection { // this is a toss up whether it will work or not (because custom types) but my gut says both of these will be freshly json parsed as generic json types (string or map[string]interface{}) so the comparison will be ok
			return true
		}
	}
	return false
}

// process a player's choice and return the next prompt
func (g *GameState) ProcessPlayerAction(action PromptResponse) Prompt {
	switch action.Action {
	case ChooseAction:
		// the player has chosen an action. They require a next prompt
		at := GetSelection(action.Action, action.Selection).(ActionType)
		return g.PromptForAction(at)
	case ChooseWeather:
		// set the weather and prompt next
		g.CurrentTurn.Weather = GetSelection(action.Action, action.Selection).(WeatherType)
		return g.NextChooseActionPrompt()
	case ChooseGrowth:
		// grow 1 bamboo on the selected plot, then prompt next
		g.Board.PlotGrowBamboo(action.Selection.(string))
		return g.NextChooseActionPrompt()
	case ChoosePandaDestination:
		// eat 1 bamboo on the selected plot, then prompt next
		bamboo := g.Board.MovePanda(action.Selection.(string))
		if bamboo != AnyPlot && bamboo != PondPlot {
			p := g.GetCurrentPlayer()
			p.Bamboo[bamboo]++
		}
		return g.NextChooseActionPrompt()
	case ChooseGardenerDestination:
		// move gardner
		g.Board.MoveGardener(action.Selection.(string))
		return g.NextChooseActionPrompt()
	case ChooseImprovementDestination:
		// place the improvement the player chose earlier on the plot they just chose
		it := GetSelection(ChooseImprovementToUse, g.CurrentTurn.ContextSelection).(ImprovementType)
		g.Board.PlotAddImprovement(action.Selection.(string), it)
		return g.NextChooseActionPrompt()
	case ChoosePlotDestination:
		//
		selectedPlot := GetSelection(ChoosePlot, g.CurrentTurn.ContextSelection).(DeckPlot)
		g.Board.AddPlot(action.Selection.(string), selectedPlot.Type, selectedPlot.Improvement)
		return g.NextChooseActionPrompt()
	case ChooseIrrigationDestination:
		//
		g.Board.EdgeAddIrrigation(action.Selection.(string))
		p := g.GetCurrentPlayer()
		p.Irrigations--
		return g.NextChooseActionPrompt()
	case ChoosePlot:
		//
		g.CurrentTurn.ContextSelection = action.Selection
		selectedPlot := GetSelection(ChoosePlot, action.Selection).(DeckPlot)
		drawnPlots := make([]DeckPlot, 0)
		for _, opt := range g.CurrentTurn.CurrentPrompt.SelectFrom {
			dp := GetSelection(ChoosePlot, opt).(DeckPlot)
			drawnPlots = append(drawnPlots, dp)
		}
		g.ReturnPlots(selectedPlot, drawnPlots)
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
		it := GetSelection(action.Action, action.Selection).(ImprovementType)
		p := g.GetCurrentPlayer()
		p.Improvements[it]--
		options := g.Board.AllImprovablePlots()
		return Prompt{
			Action:     ChooseImprovementDestination,
			SelectType: PlotIDSelectType,
			SelectFrom: ConvertToInterfaceSlice(options),
			Time:       45,
			Pid:        NewPromptID(),
		}
	case ChooseImprovementToStash:
		//
		p := g.GetCurrentPlayer()
		it := GetSelection(action.Action, action.Selection).(ImprovementType)
		p.Improvements[it]++
		g.AvailableImprovements[it]--
		return g.NextChooseActionPrompt()
	case ChooseObjectiveType:
		//
		ot := GetSelection(action.Action, action.Selection).(ObjectiveType)
		o := g.DrawObjective(ot)
		p := g.GetCurrentPlayer()
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
	return Prompt{Action: NextPlayerTurn} // this line *should* be unreachable with proper PromptResponse validation
}

func (g *GameState) PromptForAction(at ActionType) Prompt {
	switch at {
	case PlacePlot:
		//
		g.CurrentTurn.ActionsUsed = append(g.CurrentTurn.ActionsUsed, at)
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
		g.CurrentTurn.ActionsUsed = append(g.CurrentTurn.ActionsUsed, at)
		p := g.GetCurrentPlayer()
		p.Irrigations++
		return g.NextChooseActionPrompt()
	case MovePanda:
		//
		g.CurrentTurn.ActionsUsed = append(g.CurrentTurn.ActionsUsed, at)
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
		g.CurrentTurn.ActionsUsed = append(g.CurrentTurn.ActionsUsed, at)
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
		g.CurrentTurn.ActionsUsed = append(g.CurrentTurn.ActionsUsed, at)
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
		p := g.GetCurrentPlayer()
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
		return Prompt{Action: NextPlayerTurn}
	}
}

func (g *GameState) CompleteObjectives() {
	p := g.GetCurrentPlayer()
	incomplete := make([]Objective, 0)
	for _, o := range p.Objectives {
		if o.IsComplete(*p, *g.Board) {
			p.CompleteObjectives = append(p.CompleteObjectives, o)
		} else {
			incomplete = append(incomplete, o)
		}
	}
	if g.awardEmperorCard(p) {
		p.CompleteObjectives = append(p.CompleteObjectives, Objective{EmperorObjective{}})
	}
	p.Objectives = incomplete
}

func (g *GameState) awardEmperorCard(p *Player) bool {
	if g.EmperorWinner != "" {
		return false
	}
	players := len(g.Players)
	completed := len(p.CompleteObjectives)
	if completed >= (11 - players) { // 2 player = 9 objectives, 3 p = 8o, 4 p = 7o
		g.EmperorWinner = p.ID
		return true
	}
	return false
}

var roll func(int) int = rand.Intn

// roll the weather die. The outcome depends on how many improvements are available.
func RollWeatherDie(improvements bool) WeatherType {
	var w [6]WeatherType
	if improvements {
		w = [6]WeatherType{SunWeather, RainWeather, WindWeather, BoltWeather, CloudWeather, ChoiceWeather}
	} else {
		w = [6]WeatherType{SunWeather, RainWeather, WindWeather, BoltWeather, ChoiceWeather, ChoiceWeather}
	}
	r := roll(6)

	return w[r]
}
