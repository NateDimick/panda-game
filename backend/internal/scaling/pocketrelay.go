package scaling

import (
	"pandagame/internal/config"
	"pandagame/internal/framework"
	"pandagame/internal/pocketbase"
	"sync"

	"github.com/google/uuid"
)

type PocketBaseRelayer struct {
	pb        pocketbase.PBClient
	events    []framework.RelayMessage
	eventLock sync.Mutex
}

// TODO needs a background job that listens to pocketbase events
// also TODO: figure out what pocketbase events data format looks like
// The first event is PB_CONNECT with the data {"clientId": "<value>"}
// pocketbase sse's also appear to include an id: <clientId> line with each event
// event names appear to be in the pattern of <collectionName/recordId>
// event data appears to be json like this: {"action": "", "record": ""}

func NewPocketBaseRelay(cfg config.PocketBaseConfig) framework.Relayer {
	pbr := &PocketBaseRelayer{
		pb:     *pocketbase.NewPocketBase(cfg.Address, nil),
		events: make([]framework.RelayMessage, 0),
	}

	pbr.pb.AsAdmin().Admins().PasswordAuth(pocketbase.AdminPasswordBody{
		Identity: cfg.AdminIdentity,
		Password: cfg.AdminPassword,
	})

	pbr.pb.AsAdmin().Realtime().Connect(func(re pocketbase.RealtimeEvent) {
		if re.Event == "PB_CONNECT" {
			pbr.pb.AsAdmin().Realtime().SetSubscriptions(pocketbase.Subscription{
				ClientID:      "todo - from data",
				Subscriptions: []string{"events"},
			})
		}
		// get record id from re.data.record
		// get the record
		// get the relay message from record.event
		// put relay message in the events structure
	})

	return pbr
}

func (p *PocketBaseRelayer) Broadcast(msg framework.RelayMessage) {
	record := pocketbase.NewRecord{
		ID: uuid.NewString(),
		Fields: map[string]any{
			"event": msg,
		},
	}
	p.pb.AsAdmin().Records("events").Create(record, pocketbase.RecordQuery{})
}

func (p *PocketBaseRelayer) ReceiveBroadcasts() []framework.RelayMessage {
	p.eventLock.Lock()
	defer p.eventLock.Unlock()
	defer func() {
		p.events = make([]framework.RelayMessage, 0)
	}()
	return p.events
}
