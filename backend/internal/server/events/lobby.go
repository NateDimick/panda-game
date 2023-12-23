package events

import "pandagame/internal/auth"

// stored in redis
type Lobby struct {
	Host       *auth.UserSession
	Players    []*auth.UserSession
	Spectators []*auth.UserSession
	Started    bool
	GameId     string
}

// shared with uis
type UILobby struct {
	Host       string   `json:"host"`
	Players    []string `json:"players"`
	Spectators []string `json:"spectators"`
	Started    bool     `json:"started"`
	GameId     string   `json:"gid"`
}

func (l *Lobby) RemoveIDs() *UILobby {
	u := new(UILobby)
	u.Host = l.Host.Name
	u.Started = l.Started
	u.GameId = l.GameId
	u.Players = make([]string, len(l.Players))

	for i, p := range l.Players {
		u.Players[i] = p.Name
	}
	u.Spectators = make([]string, len(l.Spectators))
	for i, s := range l.Spectators {
		u.Spectators[i] = s.Name
	}

	return u
}
