package scaling

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"pandagame/internal/framework"
	"sync"

	"github.com/nats-io/nats.go"
)

type NatsRelay struct {
	nc           *nats.Conn
	subject      string
	buffer       []json.RawMessage
	lock         sync.Mutex
	subscription *nats.Subscription
}

func NewNatsRelay(addr, subject string) *NatsRelay {
	conn, _ := nats.Connect(addr)
	nr := &NatsRelay{
		nc:      conn,
		subject: subject,
		buffer:  make([]json.RawMessage, 0),
	}
	sub, _ := conn.Subscribe(subject, func(msg *nats.Msg) {
		nr.lock.Lock()
		defer nr.lock.Unlock()
		slog.Info("incoming message from subject", slog.Any("message", *msg))
		nr.buffer = append(nr.buffer, msg.Data)
	})
	nr.subscription = sub
	return nr
}

func (n *NatsRelay) Broadcast(m framework.RelayMessage) {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(m)
	if err := n.nc.Publish(n.subject, buf.Bytes()); err != nil {
		slog.Error("failed to publish to nats", slog.String("error", err.Error()))
	}
}

func (n *NatsRelay) ReceiveBroadcasts() []framework.RelayMessage {
	n.lock.Lock()
	defer n.lock.Unlock()
	result := make([]framework.RelayMessage, 0)
	for _, r := range n.buffer {
		m := new(framework.RelayMessage)
		json.Unmarshal(r, m)
		result = append(result, *m)
	}
	n.buffer = make([]json.RawMessage, 0)
	if len(result) > 0 {
		slog.Info("broadcasts consumed", slog.Any("messages", result))
	}
	return result
}
