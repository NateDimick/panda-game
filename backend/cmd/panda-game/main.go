package main

import (
	"log/slog"
	"net/http"
	"os"
	"pandagame/internal/config"
	"pandagame/internal/server"
)

func main() {
	config.SetLogger()
	s := server.NewServer()
	go func() {
		if err := s.Socket.Serve(); err != nil {
			slog.Error("Socketio server error, exiting", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()
	slog.Info("panda game server is running")
	http.ListenAndServe(":3000", s.Router)
}
