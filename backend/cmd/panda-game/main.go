package main

import (
	"log/slog"
	"net/http"
	"pandagame/internal/config"
	"pandagame/internal/server"
)

func main() {
	config.SetLogger()
	s := server.NewServer()
	slog.Info("panda game server is running")
	http.ListenAndServe(":3000", s.Router)
}
