package server

import (
	"log/slog"
	"net/http"
	"pandagame/internal/auth"
	"pandagame/internal/mongoconn"
	"pandagame/internal/pandaplex"
	"pandagame/internal/redisconn"
	"pandagame/internal/server/events"
	custommw "pandagame/internal/server/middleware"
	"pandagame/internal/server/routes"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Server struct {
	Router *chi.Mux
	Plexer pandaplex.Plexer
}

// setup a new server (chi mux) with all middlewares and routes
func NewServer() *Server {
	r := chi.NewRouter()
	// router middlewares
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(time.Minute))
	r.Use(custommw.RequestLogger)
	r.Use(custommw.AllowAllOrigins)
	// socket router
	gs := events.GameServer{Redis: redisconn.NewRedisConn(), Mongo: mongoconn.NewMongoConn()}
	plexerSettings := func(c *pandaplex.PlexerConfig) {
		c.Handler = gs.HandleMessage
		c.IdGenerator = ConnectionIdIsPlayerId
		// use default storage and relayer for development - in-memory solutions
	}
	socketPlexer := pandaplex.NewPlexer(plexerSettings)
	// auth api
	a := routes.NewAuthAPI(mongoconn.NewMongoConn(), redisconn.NewRedisConn())
	r.Post("/register", a.RegisterUser)
	r.Post("/login", a.LoginUser)
	r.Post("/guest", a.LoginAsGuest)
	r.Post("/logout", a.Logout)
	r.Get("/userinfo", a.UserInfo)
	r.Handle("/ws", socketPlexer)
	server := &Server{r, socketPlexer}
	socketPlexer.Start()
	return server
}

func ConnectionIdIsPlayerId(r *http.Request) string {
	slog.Info("getting connection id", slog.Any("headers", r.Header))
	rc := redisconn.NewRedisConn()
	sessionCookie, _ := r.Cookie("pandaGameSession")
	us, _ := redisconn.GetThing[auth.UserSession]("s-"+sessionCookie.Value, rc)
	return us.PlayerID
}
