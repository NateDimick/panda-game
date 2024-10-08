package framework

import (
	"fmt"
	"net/http"
	"strconv"
)

type Engine interface {
	HandleEvent(Event) ([]Event, error)
}

type Event struct {
	Source   EventTarget    `json:"source"`
	SourceId string         `json:"sourceId"` // a client id when the source
	Dest     EventTarget    `json:"eventTarget"`
	DestId   string         `json:"destinationId"` // either client id or Groupid, depends on Dest
	Type     string         `json:"type"`
	Payload  any            `json:"payload"`
	Metadata map[string]any `json:"metadata"`
}

type EventTarget int

func (e *EventTarget) MarshalJSON() ([]byte, error) {
	i := int(*e)
	s := strconv.Itoa(i)
	return []byte(s), nil
}

func (e *EventTarget) UnmarshalJSON(b []byte) error {
	s := string(b)
	i, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*e = EventTarget(i)
	return nil
}

const (
	TargetServer          EventTarget = iota // Source: when from a server broadcast, Dest: when from a client
	TargetClient                             // Source: when from a client, Dest: when to send back to client
	TargetJoinGroup                          // Source: never, Dest: when the event should result in a user joining a Group(produced by engine)
	TargetLeaveGroup                         // Source: never, Dest: when the event should result in a user leaving a Group(produced by engine)
	TargetGroup                              // Source: never, Dest: when the event should be sent to a room
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
