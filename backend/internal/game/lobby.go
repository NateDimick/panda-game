package game

type Lobby struct {
	Host       string
	Players    []string
	Spectators []string
	Started    bool
	GameId     string
}
