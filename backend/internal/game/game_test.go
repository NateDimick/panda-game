package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGame(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Fail()
		}
	}()
	g := NewGame()
	assert.NotNil(t, g)
	assert.Equal(t, 15, len(g.ObjectiveDecks[PlotObjectiveType]))
	assert.Equal(t, 15, len(g.ObjectiveDecks[GardenerObjectiveType]))
	assert.Equal(t, 15, len(g.ObjectiveDecks[PandaObjectiveType]))
}

func TestMarshalGame(t *testing.T) {

}

func TestDrawPlots(t *testing.T) {
	g := new(GameState)
	g.PlotDeck = []DeckPlot{
		{Type: GreenBambooPlot},
		{Type: GreenBambooPlot},
		{Type: GreenBambooPlot},
		{Type: GreenBambooPlot},
		{Type: GreenBambooPlot},
	}
	plotOptions := g.DrawPlots()
	assert.Equal(t, 3, len(plotOptions))
	assert.Equal(t, 2, len(g.PlotDeck))

	plotOptions2 := g.DrawPlots()
	assert.Equal(t, 2, len(plotOptions2))
	assert.Equal(t, 0, len(g.PlotDeck))
}

func TestReturnPlots(t *testing.T) {
	g := NewGame()
	plotsBefore := len(g.PlotDeck)
	options := g.DrawPlots()
	selection := options[1]
	g.ReturnPlots(selection, options)
	assert.Equal(t, plotsBefore-1, len(g.PlotDeck))
}

func TestAvailableObjectiveTypes(t *testing.T) {
	g := NewGame()
	types := g.AvailableObjectiveTypes()
	assert.Equal(t, 3, len(types))
	g.ObjectiveDecks[PandaObjectiveType] = []Objective{}
	types = g.AvailableObjectiveTypes()
	assert.Equal(t, 2, len(types))
}

func TestDrawObjective(t *testing.T) {
	g := NewGame()
	o := g.DrawObjective(GardenerObjectiveType)
	assert.Equal(t, 14, len(g.ObjectiveDecks[GardenerObjectiveType]))
	// gardener objectives are unique and so are plot objectives
	// this test would not work with panda objectives, but it does prove that the functionality works for panda objectives
	assert.NotContains(t, g.ObjectiveDecks[GardenerObjectiveType], o)
}

func TestGetCurrentPlayer(t *testing.T) {
	g := NewGame()
	g.Players = []*Player{
		{ID: "Harvey", Order: 1},
		{ID: "Gwendolyn", Order: 2},
		{ID: "Oliver", Order: 3},
		{ID: "Mackenzie", Order: 4},
	}
	g.CurrentTurn = &Turn{
		PlayerID: "Oliver",
	}

	p := g.GetCurrentPlayer()
	assert.Equal(t, 3, p.Order)
}

func TestValidatePlayerAction(t *testing.T) {
	pid := NewPromptID()
	cases := []struct {
		Name       string
		PR         PromptResponse
		ExpectPass bool
	}{
		{"Valid", PromptResponse{Action: ChooseAction, Selection: "MovePanda", Pid: pid}, true},
		{"Invalid Action", PromptResponse{Action: ChooseGrowth, Selection: "MovePanda", Pid: pid}, false},
		{"Invalid PID", PromptResponse{Action: ChooseAction, Selection: "MovePanda", Pid: "not pid"}, false},
		{"Invalid Selection", PromptResponse{Action: ChooseAction, Selection: "DrawObjective", Pid: pid}, false},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(tt *testing.T) {
			g := NewGame()
			g.CurrentTurn.CurrentPrompt = Prompt{
				Action:     ChooseAction,
				SelectType: ActionSelectType,
				SelectFrom: []interface{}{
					string(PlacePlot), // when gamestate is deserialized, values like this will be raw strings
					string(MovePanda),
					string(MoveGardener),
				},
				Pid: pid,
			}

			result := g.ValidatePlayerAction(tc.PR)
			assert.Equal(tt, tc.ExpectPass, result)
		})
	}

}

func TestNextChooseAction(t *testing.T) {
	g := NewGame()
	g.AddPlayers([]*Player{{ID: "dummy", Improvements: make(ImprovementReserve)}})
	g.NextTurn()
	// player has no resources and has used no actions, so should have 5 options
	prompt := g.NextChooseActionPrompt()
	assert.Equal(t, 5, len(prompt.SelectFrom))
	// give player resources, see that the options are added
	g.Players[0].Irrigations++
	g.Players[0].Improvements[WatershedImprovement]++
	prompt = g.NextChooseActionPrompt()
	assert.Equal(t, 7, len(prompt.SelectFrom))
	// set an action to used and see that it doesn't appear
	g.CurrentTurn.ActionsUsed = append(g.CurrentTurn.ActionsUsed, MovePanda)
	prompt = g.NextChooseActionPrompt()
	assert.Equal(t, 6, len(prompt.SelectFrom))
	assert.NotContains(t, prompt.SelectFrom, MovePanda)
	// set the weather to wind and see that move panda becomes an option again
	g.CurrentTurn.Weather = WindWeather
	prompt = g.NextChooseActionPrompt()
	assert.Equal(t, 7, len(prompt.SelectFrom))
	assert.Contains(t, prompt.SelectFrom, MovePanda)
	// set another action as used and see that only the free actions and end game are available
	g.CurrentTurn.ActionsUsed = append(g.CurrentTurn.ActionsUsed, DrawObjective)
	prompt = g.NextChooseActionPrompt()
	assert.Equal(t, 3, len(prompt.SelectFrom))
	assert.Contains(t, prompt.SelectFrom, EndTurn)
	// set the weather to sun and see that end turn is gone, and regular actions are back
	g.CurrentTurn.Weather = SunWeather
	prompt = g.NextChooseActionPrompt()
	assert.Equal(t, 5, len(prompt.SelectFrom))
	assert.NotContains(t, prompt.SelectFrom, EndTurn)
	// use one more action, see same result as 2 actions used + wind
	g.CurrentTurn.ActionsUsed = append(g.CurrentTurn.ActionsUsed, CollectIrrigation)
	prompt = g.NextChooseActionPrompt()
	assert.Equal(t, 3, len(prompt.SelectFrom))
	assert.Contains(t, prompt.SelectFrom, EndTurn)
	// remove resources, see next player turn
	g.Players[0].Irrigations--
	g.Players[0].Improvements[WatershedImprovement]--
	prompt = g.NextChooseActionPrompt()
	assert.Equal(t, NextPlayerTurn, prompt.Action)
	assert.Empty(t, prompt.SelectFrom)
	// reset used actions, start depleting board resources to remove options
	g.CurrentTurn.ActionsUsed = make([]ActionType, 0)
	// empty plot deck - PlacePlot not an option
	g.PlotDeck = make([]DeckPlot, 0)
	prompt = g.NextChooseActionPrompt()
	assert.Equal(t, 4, len(prompt.SelectFrom))
	assert.NotContains(t, prompt.SelectFrom, PlacePlot)
	// empty irrigation - collect irrigation not an option
	g.IrrigationReserve = 0
	prompt = g.NextChooseActionPrompt()
	assert.Equal(t, 3, len(prompt.SelectFrom))
	assert.NotContains(t, prompt.SelectFrom, CollectIrrigation)
	// empty objectives - draw objective not an option
	g.ObjectiveDecks = make(map[ObjectiveType][]Objective)
	prompt = g.NextChooseActionPrompt()
	assert.Equal(t, 2, len(prompt.SelectFrom))
	assert.NotContains(t, prompt.SelectFrom, DrawObjective)
}

func TestProcessPlayerAction(t *testing.T) {
	// 24 inputs (11 + 8 for choose action + 5 weather die roll outcomes) = 20 tests cases that all require similar setup and logic to run
	// this test is massive, but it covers about 50% of game.go
	g := NewGame()
	g.AddPlayers([]*Player{{ID: "dummy", Improvements: make(ImprovementReserve), Bamboo: make(BambooReserve)}})
	g.NextTurn()

	g.Board.AddPlot("p1", GreenBambooPlot, NoImprovement)
	g.Board.AddPlot("p2", YellowBambooPlot, NoImprovement)
	g.Board.AddPlot("p3", GreenBambooPlot, FertilizerImprovement)
	g.Board.AddPlot("p4", PinkBambooPlot, NoImprovement)
	g.Board.AddPlot("p5", GreenBambooPlot, EnclosureImprovement)
	g.Board.AddPlot("p6", YellowBambooPlot, NoImprovement)
	// this this game state, e0-5 are irrigated, e6-11 are irrigatable, and e12-31 exist but cannot be irrigated
	// p7-18 also exist

	// for forcing certain die rolls
	rollResult := 0
	roll = func(int) int { return rollResult }

	cases := []struct {
		Name  string
		P     Prompt           // the prompt the player is responding to
		PC    interface{}      // prompt context - a previous choice made by the player
		PR    PromptResponse   // the player's choice
		NP    PromptType       // Next Prompt type expected after the choice
		Setup func()           // allow test cases to do additional setup
		Adhoc func(*testing.T) // additional assertions in this adhoc func
	}{
		{
			Name: "Choose Weather",
			PR: PromptResponse{
				Action:    ChooseWeather,
				Selection: "BOLT",
			},
			NP: ChooseAction,
			Adhoc: func(tt *testing.T) {
				assert.Equal(tt, g.CurrentTurn.Weather, BoltWeather)
			},
		},
		{
			Name: "Choose Growth",
			PR: PromptResponse{
				Action:    ChooseGrowth,
				Selection: "p1",
			},
			NP: ChooseAction,
			Adhoc: func(tt *testing.T) {
				assert.Equal(tt, g.Board.Plots["p1"].Bamboo, 1)
			},
		},
		{
			Name: "Choose Panda Destination",
			PR: PromptResponse{
				Action:    ChoosePandaDestination,
				Selection: "p1",
			},
			NP: ChooseAction,
			Setup: func() {
				g.Board.PlotGrowBamboo("p1")
			},
			Adhoc: func(tt *testing.T) {
				assert.Equal(tt, g.Board.PandaLocation, "p1")
				assert.Equal(tt, g.Players[0].Bamboo[GreenBambooPlot], 1)
			},
		},
		{
			Name: "Choose Gardener Destination",
			PR: PromptResponse{
				Action:    ChooseGardenerDestination,
				Selection: "p2",
			},
			NP: ChooseAction,
			Adhoc: func(tt *testing.T) {
				assert.Equal(tt, g.Board.GardenerLocation, "p2")
				assert.Equal(tt, g.Board.Plots["p2"].Bamboo, 1)
			},
		},
		{
			Name: "Choose Improvement Destination",
			PR: PromptResponse{
				Action:    ChooseImprovementDestination,
				Selection: "p4",
			},
			PC: "FERTILIZER",
			NP: ChooseAction,
			Adhoc: func(tt *testing.T) {
				assert.Equal(tt, g.Board.Plots["p4"].Improvement.Type, FertilizerImprovement)
			},
		},
		{
			Name: "Choose Plot Destination",
			PR: PromptResponse{
				Action:    ChoosePlotDestination,
				Selection: "p7",
			},
			PC: map[string]interface{}{"Type": "PINK_BAMBOO", "Improvement": "NONE"},
			NP: ChooseAction,
			Adhoc: func(tt *testing.T) {
				assert.Equal(tt, g.Board.Plots["p7"].Type, PinkBambooPlot)
			},
		},
		{
			Name: "Choose Irrigation Destination",
			PR: PromptResponse{
				Action:    ChooseIrrigationDestination,
				Selection: "e8",
			},
			NP: ChooseAction,
			Setup: func() {
				g.Players[0].Irrigations++
			},
			Adhoc: func(tt *testing.T) {
				assert.Equal(tt, g.Board.Edges["e8"].Irrigated, true)
				assert.Equal(tt, g.Players[0].Irrigations, 0)
			},
		},
		{
			Name: "Choose Plot",
			P: Prompt{
				Action:     ChoosePlot,
				SelectType: PlotSelectType,
				SelectFrom: []interface{}{
					map[string]interface{}{"Type": "GREEN_BAMBOO", "Improvement": "NONE"},
					map[string]interface{}{"Type": "PINK_BAMBOO", "Improvement": "NONE"},
					map[string]interface{}{"Type": "YELLOW_BAMBOO", "Improvement": "NONE"},
				},
			},
			PR: PromptResponse{
				Action:    ChoosePlot,
				Selection: map[string]interface{}{"Type": "PINK_BAMBOO", "Improvement": "NONE"},
			},
			NP: ChoosePlotDestination,
			Adhoc: func(tt *testing.T) {
				assert.NotNil(tt, g.CurrentTurn.ContextSelection)
				assert.Equal(tt, 29, len(g.PlotDeck)) // plot deck starts at 27, and because this test doesn't draw the size should increase by 2
			},
		},
		{
			Name: "Choose Improvement to use",
			PR: PromptResponse{
				Action:    ChooseImprovementToUse,
				Selection: "WATERSHED",
			},
			NP: ChooseImprovementDestination,
			Setup: func() {
				g.AvailableImprovements[WatershedImprovement]--
				g.Players[0].Improvements[WatershedImprovement]++
			},
			Adhoc: func(tt *testing.T) {
				assert.NotNil(tt, g.CurrentTurn.ContextSelection)
			},
		},
		{
			Name: "Choose Improvement to stash",
			PR: PromptResponse{
				Action:    ChooseImprovementToStash,
				Selection: "FERTILIZER",
			},
			NP: ChooseAction,
			Adhoc: func(tt *testing.T) {
				assert.Equal(tt, 2, g.AvailableImprovements[FertilizerImprovement])
				assert.Equal(tt, 1, g.Players[0].Improvements[FertilizerImprovement])
			},
		},
		{
			Name: "Choose Objective Type",
			PR: PromptResponse{
				Action:    ChooseObjectiveType,
				Selection: "PANDA",
			},
			NP: ChooseAction,
			Adhoc: func(tt *testing.T) {
				assert.Equal(tt, 14, len(g.ObjectiveDecks[PandaObjectiveType]))
				assert.Equal(tt, 1, len(g.Players[0].Objectives))
			},
		},
		// choose action cases
		{
			Name: "Chose Action PlacePlot",
			PR: PromptResponse{
				Action:    ChooseAction,
				Selection: "PlacePlot",
			},
			NP: ChoosePlot,
			Adhoc: func(tt *testing.T) {
				assert.Equal(tt, 26, len(g.PlotDeck)) // yuck, this test is impacted by a previous case, but we expect 3 fewer plots in the plot deck
			},
		},
		{
			Name: "Chose Action CollectIrrigation",
			PR: PromptResponse{
				Action:    ChooseAction,
				Selection: "CollectIrrigation",
			},
			NP: ChooseAction,
			Adhoc: func(tt *testing.T) {
				assert.Equal(tt, 1, g.Players[0].Irrigations)
			},
		},
		{
			Name: "Chose Action MovePanda",
			PR: PromptResponse{
				Action:    ChooseAction,
				Selection: "MovePanda",
			},
			NP: ChoosePandaDestination,
		},
		{
			Name: "Chose Action MoveGardener",
			PR: PromptResponse{
				Action:    ChooseAction,
				Selection: "MoveGardener",
			},
			NP: ChooseGardenerDestination,
		},
		{
			Name: "Chose Action DrawObjective",
			PR: PromptResponse{
				Action:    ChooseAction,
				Selection: "DrawObjective",
			},
			NP: ChooseObjectiveType,
		},
		{
			Name: "Chose Action PlaceIrrigation",
			PR: PromptResponse{
				Action:    ChooseAction,
				Selection: "PlaceIrrigation",
			},
			NP: ChooseIrrigationDestination,
		},
		{
			Name: "Chose Action PlaceImprovement",
			PR: PromptResponse{
				Action:    ChooseAction,
				Selection: "PlaceImprovement",
			},
			NP: ChooseImprovementToUse,
		},
		{
			Name: "Chose Action EndTurn",
			PR: PromptResponse{
				Action:    ChooseAction,
				Selection: "EndTurn",
			},
			NP: NextPlayerTurn,
		},
		// die roll cases
		{
			Name: "Die roll sun/wind",
			PR: PromptResponse{
				Action:    RollDie,
				Selection: "RollDie",
			},
			Setup: func() {
				rollResult = 0 // sun, 2 for wind would do the same
			},
			NP: ChooseAction,
			Adhoc: func(tt *testing.T) {
				assert.Equal(tt, g.CurrentTurn.Weather, SunWeather)
			},
		},
		{
			Name: "Die roll rain",
			PR: PromptResponse{
				Action:    RollDie,
				Selection: "RollDie",
			},
			Setup: func() {
				rollResult = 1
			},
			NP: ChooseGrowth,
			Adhoc: func(tt *testing.T) {
				assert.Equal(tt, g.CurrentTurn.Weather, RainWeather)
			},
		},
		{
			Name: "Die roll bolt",
			PR: PromptResponse{
				Action:    RollDie,
				Selection: "RollDie",
			},
			Setup: func() {
				rollResult = 3
			},
			NP: ChoosePandaDestination,
			Adhoc: func(tt *testing.T) {
				assert.Equal(tt, g.CurrentTurn.Weather, BoltWeather)
			},
		},
		{
			Name: "Die roll cloud",
			PR: PromptResponse{
				Action:    RollDie,
				Selection: "RollDie",
			},
			Setup: func() {
				rollResult = 4
			},
			NP: ChooseImprovementToStash,
			Adhoc: func(tt *testing.T) {
				assert.Equal(tt, g.CurrentTurn.Weather, CloudWeather)
			},
		},
		{
			Name: "Die roll choice",
			PR: PromptResponse{
				Action:    RollDie,
				Selection: "RollDie",
			},
			Setup: func() {
				rollResult = 5
			},
			NP: ChooseWeather,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(tt *testing.T) {
			g.CurrentTurn.CurrentPrompt = tc.P
			g.CurrentTurn.ContextSelection = tc.PC
			if tc.Setup != nil {
				tc.Setup()
			}
			result := g.ProcessPlayerAction(tc.PR)
			assert.Equal(tt, tc.NP, result.Action)
			if result.Action != NextPlayerTurn {
				assert.NotEmpty(tt, result.SelectFrom)
			}
			if tc.Adhoc != nil {
				tc.Adhoc(tt)
			}

		})
	}
}

func TestCompleteObjectives(t *testing.T) {
	g := NewGame()
	g.AddPlayers([]*Player{
		{ID: "a", Objectives: make([]Objective, 0)},
		{ID: "b", Objectives: make([]Objective, 0)},
		{ID: "c", Objectives: make([]Objective, 0)},
		{ID: "d", Objectives: make([]Objective, 0)},
	})
	g.NextTurn()

	g.Players[0].CompleteObjectives = []Objective{
		{},
		{},
		{},
		{},
		{},
		{},
	}

	g.Players[0].Objectives = append(g.Players[0].Objectives, Objective{PlotObjective{AnchorColor: GreenBambooPlot, Neighbors: [6]PlotType{GreenBambooPlot, GreenBambooPlot, AnyPlot, AnyPlot, AnyPlot, AnyPlot}}})

	g.Board.AddPlot("p1", GreenBambooPlot, NoImprovement)
	g.Board.AddPlot("p2", GreenBambooPlot, NoImprovement)
	g.Board.AddPlot("p7", GreenBambooPlot, WatershedImprovement)

	g.CompleteObjectives()

	assert.Equal(t, g.Players[0].ID, g.EmperorWinner)
	assert.Equal(t, 8, len(g.Players[0].CompleteObjectives))
	assert.Equal(t, 0, len(g.Players[0].Objectives))
}
