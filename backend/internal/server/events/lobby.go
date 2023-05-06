package events

type Lobby struct {
	Host       ConnectionContext
	Players    []ConnectionContext
	Spectators []ConnectionContext
	Started    bool
}

type UILobby struct {
	Host       string
	Players    []string
	Spectators []string
	Started    bool
}

func (l *Lobby) RemoveIDs() *UILobby {
	u := new(UILobby)
	u.Host = l.Host.UserName
	u.Started = l.Started
	u.Players = make([]string, len(l.Players))

	for i, p := range l.Players {
		u.Players[i] = p.UserName
	}
	u.Spectators = make([]string, len(l.Spectators))
	for i, s := range l.Spectators {
		u.Spectators[i] = s.UserName
	}

	return u
}
