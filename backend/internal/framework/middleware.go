package framework

import (
	"net/http"

	"github.com/google/uuid"
)

type Middleware func(Event, *http.Request) (Event, error)

type ErrorHandler func(Event, error) Event

func defaultErrorHandler(e Event, err error) Event {
	return Event{
		Source:   TargetServer,
		Dest:     TargetClient,
		Type:     "Error",
		Payload:  err,
		Metadata: e.Metadata,
	}
}

type DisconnectHandler func(*http.Request)

func defaultDisconnectHandler(_ *http.Request) {

}

type IdGenerator func(*http.Request) string

func defaultIdGenerator(_ *http.Request) string {
	return uuid.NewString()
}

// return an error to refuse the websocket connection
type ConnectionHandler func(http.ResponseWriter, *http.Request) error

func defaultConnectionHandler(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}
