package main

import (
	"net/http"
	"pandagame/internal/config"
	"pandagame/internal/server"

	"go.uber.org/zap"
)

func main() {
	logger, _ := config.LoggerConfig().Build()
	zap.ReplaceGlobals(logger)
	s := server.NewServer()
	go func() {
		if err := s.Socket.Serve(); err != nil {
			zap.L().Fatal("Socketio server error, exiting", zap.Error(err))
		}
	}()
	zap.L().Info("panda game server is running")
	http.ListenAndServe(":3000", s.Router)
}
