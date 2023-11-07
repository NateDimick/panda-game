package pandaplex

const broadcastAllRecipients string = "all"

// a relayer can relay across multiple instances running the plexer
// if one plexer communicates from client to server, then a relayer communicates server to server
type PlexerRelayer interface {
	Broadcast(RelayMessage)
	ReceiveBroadcasts() []RelayMessage
}

type RelayMessage struct {
	Message      string
	RecipientIds []string
}

type inMemRelayer struct {
	messages []RelayMessage
}

func NewInMemRelayer() PlexerRelayer {
	imr := new(inMemRelayer)
	imr.messages = make([]RelayMessage, 0)
	return imr
}

func (i *inMemRelayer) Broadcast(m RelayMessage) {
	i.messages = append(i.messages, m)
}

func (i *inMemRelayer) ReceiveBroadcasts() []RelayMessage {
	defer func() {
		i.messages = make([]RelayMessage, 0)
	}()
	return i.messages
}
