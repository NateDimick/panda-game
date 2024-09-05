package framework

import (
	"fmt"
	"net/http"
)

type Engine interface {
	HandleEvent(Event) ([]Event, error)
}

type Event struct {
	Source   EventTarget
	SourceId string // a client id when the source
	Dest     EventTarget
	DestId   string // either client id or room id, depends on Dest
	Type     string
	Payload  any
	Metadata map[string]any
}

type EventTarget int

const (
	TargetServer          EventTarget = iota // Source: when from a server broadcast, Dest: when from a client
	TargetClient                             // Source: when from a client, Dest: when to send back to client
	TargetJoinRoom                           // Source: never, Dest: when the event should result in a user joining a room (produced by engine)
	TargetLeaveRoom                          // Source: never, Dest: when the event should result in a user leaving a room (produced by engine)
	TargetRoom                               // Source: never, Dest: when the event should be sent to a room
	TargetServerBroadcast                    // Source: never, Dest: when the message should be relayed to other servers
	TargetClientBroadcast                    // Source: never, Dest: when the message should be relayed to other clients (including clients on other servers)
	TargetNone                               // Source: never, Dest: when the event should be dropped
)

type EventPayloadDeserializer func(string, *http.Request) (eventType string, payload any, err error)

type EventPayloadSerializer func(eventType string, payload any) (string, error)

func defaultDeserializer(message string, _ *http.Request) (string, any, error) {
	return "", message, nil
}

func defaultSerializer(t string, a any) (string, error) {
	return fmt.Sprintf("%s: %+v", t, a), nil
}
