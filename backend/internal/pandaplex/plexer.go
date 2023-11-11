package pandaplex

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

// MessageHandler is the entrypoint for handling messages that come off the socket
type MessageHandler func(PlexerConnection, string)

// handles a disconnect on a connection
type DisconnectHandler func(PlexerConnection)

func NoOpHandler(pc PlexerConnection, m string) {
	// this function does nothing
}

func NoOpDisconnector(pc PlexerConnection) {
	// this function does nothing
}

// ConnectionIdGenerator allows for connection ids to be generated
type ConnectionIdGenerator func(*http.Request) string

func UUIDGenerator(r *http.Request) string {
	id, _ := uuid.NewV4()
	return id.String()
}

type PlexerConfig struct {
	MaxConnections int
	Handler        MessageHandler
	Disconnector   DisconnectHandler
	IdGenerator    ConnectionIdGenerator
	Relayer        PlexerRelayer
	Storage        PlexerStorage
}

var defaultPlexerConfig = PlexerConfig{
	MaxConnections: 1000,
	Relayer:        NewInMemRelayer(),
	Storage:        NewInMemStorage(),
	Handler:        NoOpHandler,
	Disconnector:   NoOpDisconnector,
	IdGenerator:    UUIDGenerator,
}

// this interface is provided to MessageHandlers to communicate with other connections
type PlexerConnection interface {
	// get the id of the connection handling this message
	ID() string
	// request headers of the connection
	Headers() http.Header
	// request cookies
	Cookies() []*http.Cookie
	// send a message back down on the same connection
	Reply(message string)
	// send the message to each specified connection id
	SendTo(message string, recipientIds ...string)
	// send a message to a room and all members of that room
	SendToRoom(message string, roomId string)
	// send a message to all connections
	Broadcast(message string)
	// join a room
	JoinRoom(roomId string)
	// leave a room
	LeaveRoom(roomId string)
}

// The Plexer is a message infrastructure tool which facilitates socketio-like communication between many websocket connections.
type Plexer interface {
	http.Handler
	// start plexer background functionality
	Start()
	// send a message in to the plexer from outside the plexer message handling
	Send(message string, recipientIds ...string)
}

type plexerImpl struct {
	connections map[string]chan<- string // allows for send messages to active connection. maybe this value should have a mutex associated with it. keys should be unique though, so... not yet.
	config      *PlexerConfig
	started     bool
}

type plexerInternalImpl struct {
	conn    *websocket.Conn
	id      string
	config  *PlexerConfig
	headers http.Header
	cookies []*http.Cookie
}

// configure a new plexer
func NewPlexer(config ...func(*PlexerConfig)) Plexer {
	p := new(plexerImpl)
	p.connections = make(map[string]chan<- string)
	cfg := defaultPlexerConfig
	for _, fn := range config {
		fn(&cfg)
	}
	p.config = &cfg
	return p
}

// connect to server via websocket. Upgrade the connection. Listen on the connection.
func (p *plexerImpl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	go func() {
		defer func() {
			conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, ""), time.Now().Add(time.Second))
			conn.Close()
		}()
		connId := p.config.IdGenerator(r)
		writeChan := make(chan string)
		readChan := make(chan string)
		closeCtx, cncl := context.WithCancel(context.Background())
		p.connections[connId] = writeChan

		pInternal := &plexerInternalImpl{
			conn:    conn,
			config:  p.config,
			id:      connId,
			headers: r.Header,
			cookies: r.Cookies(),
		}

		go func() {
			for {
				select {
				case <-closeCtx.Done():
					slog.Info("Connection closed on read loop", slog.String("connId", connId))
					return
				default:
					mt, msg, err := conn.ReadMessage() // this blocks?
					if mt == websocket.TextMessage && err == nil {
						readChan <- string(msg)
					} else if err != nil {
						slog.Error("Read off connection returned error, must close connection", slog.String("error", err.Error()))
						slog.Info("Connection closed by client", slog.String("connId", connId))
						delete(p.connections, connId)
						cncl()
						p.panicSafeDisconnect(pInternal)
						return
					}
				}
			}
		}()

		for {
			select {
			case msg := <-readChan:
				// handle incoming message
				p.panicSafeHandler(pInternal, msg)
			case msg := <-writeChan:
				// send outgoing message
				err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					cncl()
					delete(p.connections, connId)
					slog.Error("Unexpected close error writing message, closing connection", slog.String("error", err.Error()))
					p.panicSafeDisconnect(pInternal)
					return
				}
			case <-closeCtx.Done():
				slog.Info("Connection closed on write loop", slog.String("connId", connId))
				return
			}
		}
	}()
}

func (p *plexerImpl) panicSafeHandler(internal *plexerInternalImpl, msg string) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("plexer handler panic", slog.Any("recoverValue", err))
		}
	}()
	p.config.Handler(internal, msg)
}

func (p *plexerImpl) panicSafeDisconnect(internal *plexerInternalImpl) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("plexer disconnector panic", slog.Any("recoverValue", err))
		}
	}()
	p.config.Disconnector(internal)
}

// starts additional plexer functionality in the background
func (p *plexerImpl) Start() {
	if p.started {
		return
	}
	p.started = true
	go func() {
		for {
			messages := p.config.Relayer.ReceiveBroadcasts()
			for _, m := range messages {
				for _, id := range m.RecipientIds {
					if c, ok := p.connections[id]; ok || m.All {
						c <- m.Message
					}
				}
			}
			time.Sleep(time.Millisecond * 100) // 10x/second
		}
	}()
}

func (p *plexerImpl) Send(message string, recipientIds ...string) {
	p.config.Relayer.Broadcast(RelayMessage{Message: message, RecipientIds: recipientIds})
}

// get the id of the connection handling this message
func (p *plexerInternalImpl) ID() string {
	return p.id
}

func (p *plexerInternalImpl) Headers() http.Header {
	return p.headers
}

func (p *plexerInternalImpl) Cookies() []*http.Cookie {
	return p.cookies
}

// send a message back down on the same connection
func (p *plexerInternalImpl) Reply(message string) {
	p.conn.WriteMessage(websocket.TextMessage, []byte(message))
}

// send the message to each specified connection id
func (p *plexerInternalImpl) SendTo(message string, recipientIds ...string) {
	p.config.Relayer.Broadcast(RelayMessage{Message: message, RecipientIds: recipientIds})
}

// send a message to a room and all members of that room
func (p *plexerInternalImpl) SendToRoom(message string, roomId string) {
	recipients := p.config.Storage.RoomMembers(roomId)
	p.config.Relayer.Broadcast(RelayMessage{Message: message, RecipientIds: recipients})
}

// send a message to all connections
func (p *plexerInternalImpl) Broadcast(message string) {
	p.config.Relayer.Broadcast(RelayMessage{Message: message, RecipientIds: []string{}, All: true})
}

func (p *plexerInternalImpl) JoinRoom(roomId string) {
	p.config.Storage.AddToRoom(p.id, roomId)
}

func (p *plexerInternalImpl) LeaveRoom(roomId string) {
	p.config.Storage.RemoveFromRoom(p.id, roomId)
}
