package pandaplex

import (
	"net"
	"net/http"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/gofrs/uuid"
)

// MessageHandler is the entrypoint for handling messages that come off the socket
type MessageHandler func(PlexerInternal, string)

func NoOpHandler(pi PlexerInternal, m string) {
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
	IdGenerator    ConnectionIdGenerator
	Relayer        PlexerRelayer
	Storage        PlexerStorage
}

var defaultPlexerConfig = PlexerConfig{
	MaxConnections: 1000,
	Relayer:        NewInMemRelayer(),
	Storage:        NewInMemStorage(),
	Handler:        NoOpHandler,
	IdGenerator:    UUIDGenerator,
}

// this interface is provided to MessageHandlers to communicate with other connections
type PlexerInternal interface {
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
	connections map[string]chan<- string // allows for send messages to active connection
	config      *PlexerConfig
	started     bool
}

type plexerInternalImpl struct {
	conn    net.Conn
	id      string
	config  *PlexerConfig
	headers http.Header
	cookies []*http.Cookie
}

// configure a new plexer
func NewPlexer(config ...func(*PlexerConfig)) Plexer {
	p := new(plexerImpl)
	p.connections = make(map[string]chan<- string)
	cfg := &defaultPlexerConfig
	for _, fn := range config {
		fn(cfg)
	}
	p.config = cfg
	return p
}

func (p *plexerImpl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	go func() {
		connId := p.config.IdGenerator(r)
		writeChan := make(chan string)
		readChan := make(chan string)
		p.connections[connId] = writeChan
		go func() {
			for {
				msg, op, err := wsutil.ReadClientData(conn)
				if op == ws.OpText && err == nil {
					readChan <- string(msg)
				}
				time.Sleep(time.Millisecond * 10)
			}
		}()

		pInternal := &plexerInternalImpl{
			conn:    conn,
			config:  p.config,
			id:      connId,
			headers: r.Header,
			cookies: r.Cookies(),
		}

		for {

			select {
			case msg := <-readChan:
				// handle incoming message
				p.config.Handler(pInternal, msg)
			case msg := <-writeChan:
				// send outgoing message
				wsutil.WriteServerText(conn, []byte(msg))
			}
			// TODO: handle disconnections
		}
	}()
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
	wsutil.WriteServerText(p.conn, []byte(message))
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
