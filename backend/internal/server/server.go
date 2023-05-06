package server

import (
	"net/http"
	"pandagame/internal/mongoconn"
	"pandagame/internal/redisconn"
	"pandagame/internal/server/events"
	custommw "pandagame/internal/server/middleware"
	"pandagame/internal/server/routes"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
)

type Server struct {
	Router *chi.Mux
	Socket *socketio.Server
}

// setup a new server (chi mux) with all middlewares and routes
func NewServer() *Server {
	r := chi.NewRouter()
	// router middlewares
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(time.Minute))
	r.Use(custommw.RequestLogger)
	// socket router
	s := newSocketServer()
	r.Handle("/socket/", s)
	// auth api
	a := routes.NewAuthAPI(mongoconn.NewMongoConn())
	r.Post("/register", a.RegisterUser)
	r.Post("/login", a.LoginUser)
	r.Post("/empower", a.EmpowerUser)
	r.Get("/guest", a.LoginAsGuest)
	server := &Server{r, s}
	return server
}

func newSocketServer() *socketio.Server {
	s := socketio.NewServer(&engineio.Options{
		Transports: []transport.Transport{
			&polling.Transport{
				CheckOrigin: allowOriginFunc,
			},
			&websocket.Transport{
				CheckOrigin: allowOriginFunc,
			},
		},
		PingTimeout:  time.Second,
		PingInterval: time.Millisecond,
	})
	gs := events.GameServer{Server: s, Redis: redisconn.NewRedisConn(), Mongo: mongoconn.NewMongoConn()}
	// add callbacks
	s.OnConnect(events.NS, gs.OnConnect)
	s.OnDisconnect(events.NS, gs.OnDisconnect)
	s.OnError(events.NS, gs.OnError)
	s.OnEvent(events.GNS, string(events.TakeAction), gs.OnTakeTurnAction)
	// and so on...
	return s
}

// from the examples for go-socketio
var allowOriginFunc = func(r *http.Request) bool {
	return true
}
