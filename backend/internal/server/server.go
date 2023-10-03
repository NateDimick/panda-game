package server

import (
	"pandagame/internal/mongoconn"
	"pandagame/internal/redisconn"
	"pandagame/internal/server/events"
	custommw "pandagame/internal/server/middleware"
	"pandagame/internal/server/routes"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/njones/socketio"
)

type Server struct {
	Router *chi.Mux
	Socket *socketio.ServerV4
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
	a := routes.NewAuthAPI(mongoconn.NewMongoConn(), redisconn.NewRedisConn())
	r.Post("/register", a.RegisterUser)
	r.Post("/login", a.LoginUser)
	r.Post("/empower", a.EmpowerUser)
	r.Post("/guest", a.LoginAsGuest)
	server := &Server{r, s}
	return server
}

func newSocketServer() *socketio.ServerV4 {
	s := socketio.NewServerV4()
	gs := events.GameServer{ServerV4: s, Redis: redisconn.NewRedisConn(), Mongo: mongoconn.NewMongoConn()}
	// only register onConnect. It registers all other events to the socket connection.
	s.OnConnect(gs.OnConnect)
	return s
}
