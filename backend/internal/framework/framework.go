package framework

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type FrameworkConfig struct {
	Serializer        EventPayloadSerializer
	Deserializer      EventPayloadDeserializer
	ErrorHandler      ErrorHandler
	ConnectHandler    ConnectionHandler
	DisconnectHandler DisconnectHandler
	IdGenerator       IdGenerator
	Relayer           Relayer
	Rooms             Rooms
	RelayInterval     time.Duration
}

func NewFramework(engine Engine) *Framework {
	f := &Framework{
		engine:             engine,
		receiveMiddlewares: make([]Middleware, 0),
		sendMiddlewares:    make([]Middleware, 0),
		config: &FrameworkConfig{
			Deserializer:      defaultDeserializer,
			Serializer:        defaultSerializer,
			ErrorHandler:      defaultErrorHandler,
			ConnectHandler:    defaultConnectionHandler,
			DisconnectHandler: defaultDisconnectHandler,
			IdGenerator:       defaultIdGenerator,
			Relayer:           NewInMemRelayer(),
			Rooms:             NewInMemStorage(),
			RelayInterval:     time.Millisecond * 100,
		},
		connections: make(map[string]chan Event),
		upgrader:    websocket.Upgrader{},
	}
	return f
}

// Framework implements http.Handler
type Framework struct {
	engine             Engine
	receiveMiddlewares []Middleware
	sendMiddlewares    []Middleware
	config             *FrameworkConfig
	connections        map[string]chan Event
	started            bool
	upgrader           websocket.Upgrader
}

func (f *Framework) Configure(cfgs ...func(*FrameworkConfig)) {
	for _, fn := range cfgs {
		fn(f.config)
	}
}

func (f *Framework) AddReceiveMiddleware(m Middleware) {
	f.receiveMiddlewares = append(f.receiveMiddlewares, m)
}

func (f *Framework) InsertReceiveMiddleware(m Middleware) {
	f.receiveMiddlewares = append([]Middleware{m}, f.receiveMiddlewares...)
}

func (f *Framework) AddSendMiddleware(m Middleware) {
	f.sendMiddlewares = append(f.sendMiddlewares, m)
}

func (f *Framework) InsertSendMiddleware(m Middleware) {
	f.sendMiddlewares = append([]Middleware{m}, f.sendMiddlewares...)
}

func (f *Framework) Start() {
	if f.started {
		return
	}
	f.started = true
	go func() {
		for {
			messages := f.config.Relayer.ReceiveBroadcasts()
			for _, m := range messages {
				if m.All {
					for _, c := range f.connections {
						c <- m.Message
					}
				} else {
					for _, id := range m.RecipientIds {
						if c, ok := f.connections[id]; ok {
							c <- m.Message
						}
					}
				}
			}
			time.Sleep(f.config.RelayInterval)
		}
	}()
}

func (f *Framework) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := f.config.ConnectHandler(w, r); err != nil {
		return
	}
	conn, err := f.upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	go func() {
		defer func() {
			conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, ""), time.Now().Add(time.Second))
			conn.Close()
		}()
		connId := f.config.IdGenerator(r)
		writeChan := make(chan Event)
		readChan := make(chan Event)
		closeCtx, cncl := context.WithCancel(context.Background())
		f.connections[connId] = writeChan

		go func() {
			for {
				select {
				case <-closeCtx.Done():
					slog.Info("Connection closed on read loop", slog.String("connId", connId))
					return
				default:
					mt, msg, err := conn.ReadMessage() // this blocks?
					if mt == websocket.TextMessage && err == nil {
						event := Event{}
						et, payload, err := f.config.Deserializer(string(msg), r)
						if err != nil {
							writeChan <- f.config.ErrorHandler(Event{}, err)
							continue
						}
						event.Payload = payload
						event.Type = et
						event.Dest = TargetServer
						event.Source = TargetClient
						event.SourceId = connId
						event, err = executeMiddlewares(event, r, f.receiveMiddlewares)
						if err != nil {
							writeChan <- f.config.ErrorHandler(event, err)
							continue
						}
						readChan <- event
					} else if err != nil {
						slog.Error("Read off connection returned error, must close connection", slog.String("error", err.Error()))
						slog.Info("Connection closed by client", slog.String("connId", connId))
						delete(f.connections, connId)
						cncl()
						f.config.DisconnectHandler(r)
						return
					}
				}
			}
		}()

		for {
			select {
			case event := <-readChan:
				// handle incoming message
				if err := f.handleEvent(event); err != nil {
					writeChan <- f.config.ErrorHandler(event, err)
				}
			case event := <-writeChan:
				event, err = executeMiddlewares(event, r, f.sendMiddlewares)
				if err != nil {
					writeChan <- f.config.ErrorHandler(event, err)
				}
				if event.Dest == TargetNone {
					continue
				}
				msg, err := f.config.Serializer(event.Type, event.Payload)
				if err != nil {
					msg = fmt.Sprintf("Failed to serialize event: %s", err.Error())
				}
				// send outgoing message
				err = conn.WriteMessage(websocket.TextMessage, []byte(msg))
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					cncl()
					delete(f.connections, connId)
					slog.Error("Unexpected close error writing message, closing connection", slog.String("error", err.Error()))
					f.config.DisconnectHandler(r)
					return
				}
			case <-closeCtx.Done():
				slog.Info("Connection closed on write loop", slog.String("connId", connId))
				close(writeChan)
				close(readChan)
				return
			}
		}
	}()
}

func (f *Framework) handleEvent(e Event) error {
	responseEvent, err := f.engine.HandleEvent(e)
	if err != nil {
		return err
	}
	var msg RelayMessage
	switch responseEvent.Dest {
	case TargetClient:
		msg = RelayMessage{
			Message:      responseEvent,
			RecipientIds: []string{responseEvent.DestId},
		}
	case TargetJoinRoom:
		f.config.Rooms.AddToRoom(responseEvent.SourceId, responseEvent.DestId)
	case TargetLeaveRoom:
		f.config.Rooms.RemoveFromRoom(responseEvent.SourceId, responseEvent.DestId)
	case TargetRoom:
		msg = RelayMessage{
			Message:      responseEvent,
			RecipientIds: f.config.Rooms.RoomMembers(responseEvent.DestId),
		}
	case TargetClientBroadcast:
		msg = RelayMessage{
			Message: responseEvent,
			All:     true,
		}
	case TargetNone:
		fallthrough
	default:
		return nil
	}
	if len(msg.RecipientIds) > 0 || msg.All {
		f.config.Relayer.Broadcast(msg)
	}

	return nil
}

func executeMiddlewares(e Event, r *http.Request, mws []Middleware) (Event, error) {
	var resultEvent Event = e
	var originalEvent Event = e
	var err error
	for _, mw := range mws {
		resultEvent, err = mw(resultEvent, r)
		if err != nil {
			return originalEvent, err
		}
	}
	return resultEvent, nil
}
