package game

type GameState struct {
	Board       Board
	Players     []Player
	CurrentTurn Turn
}

type Turn struct {
	PlayerID       string
	ActionsUsed    int
	CurrentCommand string
	CurrentOptions []string
	Weather        Weather
}
