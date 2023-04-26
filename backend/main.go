package main

import (
	"net/http"
	"pandagame/server"

	"go.uber.org/zap"
)

func main() {
	s := server.NewServer()
	go func() {
		if err := s.Socket.Serve(); err != nil {
			zap.L().Fatal("Socketio server error, exiting", zap.Error(err))
		}
	}()

	http.ListenAndServe(":3000", s.Router)
}
