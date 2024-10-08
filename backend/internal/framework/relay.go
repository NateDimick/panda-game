package framework

import "sync"

// A relayer sends massages from one connection to many other connections
// it can also receive messages to this connection
type Relayer interface {
	Broadcast(RelayMessage)
	ReceiveBroadcasts() []RelayMessage
}

type RelayMessage struct {
	Message      Event    `json:"message"`
	RecipientIds []string `json:"recipientIds"`
	All          bool     `json:"all"`
}

type inMemRelayer struct {
	messages []RelayMessage
	lock     sync.Mutex
}

func NewInMemRelayer() Relayer {
	imr := new(inMemRelayer)
	imr.messages = make([]RelayMessage, 0)
	return imr
}

func (i *inMemRelayer) Broadcast(m RelayMessage) {
	i.lock.Lock()
	defer i.lock.Unlock()
	i.messages = append(i.messages, m)
}

func (i *inMemRelayer) ReceiveBroadcasts() []RelayMessage {
	i.lock.Lock()
	defer func() {
		i.messages = make([]RelayMessage, 0)
		i.lock.Unlock()
	}()
	return i.messages
}
