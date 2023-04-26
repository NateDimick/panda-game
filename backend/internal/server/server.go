package server

import (
	"github.com/go-chi/chi"
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
	s := NewSocketServer()
	r.Handle("/socket", s)
	server := &Server{r, s}
	return server
}

func NewSocketServer() *socketio.Server {
	s := socketio.NewServer(&engineio.Options{})
	// add callbacks
	return s
}
