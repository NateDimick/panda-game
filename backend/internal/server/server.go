package server

import (
	"pandagame/internal/mongoconn"
	"pandagame/internal/server/events"
	"pandagame/internal/server/routes"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
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
	r.Use(middleware.RequestID) // hmmmm maybe delete? depends on how hostname performs and if it poses a security risk
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
	s := socketio.NewServer(&engineio.Options{})
	gs := events.GameServer{Server: s}
	// add callbacks
	s.OnConnect("/", gs.OnConnect)
	s.OnDisconnect("/", gs.OnDisconnect)
	s.OnError("/", gs.OnError)
	s.OnEvent("/", string(events.TakeAction), gs.OnTakeTurnAction)
	// and so on...
	return s
}
